package namespace

import (
	"context"
	"fmt"
	"time"

	"slices"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides struct for namespace object containing connection to the cluster and the namespace definitions.
type Builder struct {
	common.EmbeddableBuilder[corev1.Namespace, *corev1.Namespace]
	common.EmbeddableCreator[corev1.Namespace, Builder, *corev1.Namespace, *Builder]
	common.EmbeddableDeleter[corev1.Namespace, *corev1.Namespace]
	common.EmbeddableUpdater[corev1.Namespace, Builder, *corev1.Namespace, *Builder]
	common.EmbeddableWithOptions[corev1.Namespace, Builder, *corev1.Namespace, *Builder, AdditionalOptions]
}

// AdditionalOptions additional options for namespace object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the Namespace GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("Namespace")
}

// NewBuilder creates new instance of Builder.
func NewBuilder(apiClient *clients.Settings, name string) *Builder {
	return common.NewClusterScopedBuilder[corev1.Namespace, Builder](apiClient, corev1.AddToScheme, name)
}

// Pull loads existing namespace in to Builder struct.
func Pull(apiClient *clients.Settings, nsname string) (*Builder, error) {
	return common.PullClusterScopedBuilder[corev1.Namespace, Builder](
		context.TODO(), apiClient, corev1.AddToScheme, nsname)
}

// WithLabel redefines namespace definition with the given label.
func (builder *Builder) WithLabel(key string, value string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Labeling the namespace %s with %s=%s", builder.Definition.Name, key, value)

	if key == "" {
		klog.V(100).Info("The key cannot be empty")

		builder.SetError(fmt.Errorf("'key' cannot be empty"))

		return builder
	}

	if builder.Definition.Labels == nil {
		builder.Definition.Labels = map[string]string{}
	}

	builder.Definition.Labels[key] = value

	return builder
}

// WithMultipleLabels redefines namespace definition with the given labels.
func (builder *Builder) WithMultipleLabels(labels map[string]string) *Builder {
	for k, v := range labels {
		builder.WithLabel(k, v)
	}

	return builder
}

// RemoveLabels removes given label from Node metadata.
func (builder *Builder) RemoveLabels(labels map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Removing labels %v from namespace %s", labels, builder.Definition.Name)

	if len(labels) == 0 {
		klog.V(100).Info("labels to be removed cannot be empty")

		builder.SetError(fmt.Errorf("labels to be removed cannot be empty"))

		return builder
	}

	for key := range labels {
		delete(builder.Definition.Labels, key)
	}

	return builder
}

// DeleteAndWait deletes a namespace and waits until it is removed from the cluster.
func (builder *Builder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Deleting namespace %s and waiting for the removal to complete", builder.Definition.Name)

	if err := builder.Delete(); err != nil {
		return err
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var ns corev1.Namespace

			err := builder.GetClient().Get(
				logging.WithDiscardLogger(ctx),
				runtimeclient.ObjectKey{Name: builder.Definition.Name},
				&ns)
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			if err != nil {
				klog.V(100).Infof("Failed to get namespace %s: %v", builder.Definition.Name, err)
			}

			return false, nil
		})
}

// CleanObjects removes given objects from the namespace.
func (builder *Builder) CleanObjects(cleanTimeout time.Duration, objects ...schema.GroupVersionResource) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Clean namespace: %s", builder.Definition.Name)

	if len(objects) == 0 {
		return fmt.Errorf("failed to remove empty list of object from namespace %s",
			builder.Definition.Name)
	}

	if !builder.Exists() {
		return fmt.Errorf("failed to remove resources from non-existent namespace %s",
			builder.Definition.Name)
	}

	dynamicClient, ok := builder.GetClient().(dynamic.Interface)
	if !ok {
		return fmt.Errorf("client does not support dynamic resource operations")
	}

	for _, resource := range objects {
		klog.V(100).Infof("Clean all resources: %s in namespace: %s",
			resource.Resource, builder.Definition.Name)

		err := dynamicClient.Resource(resource).Namespace(builder.Definition.Name).DeleteCollection(
			context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{})
		if err != nil {
			klog.V(100).Infof("Failed to remove resources: %s in namespace: %s",
				resource.Resource, builder.Definition.Name)

			return err
		}

		err = wait.PollUntilContextTimeout(
			context.TODO(), 3*time.Second, cleanTimeout, true, func(ctx context.Context) (bool, error) {
				objList, err := dynamicClient.Resource(resource).Namespace(builder.Definition.Name).List(
					logging.DiscardContext(), metav1.ListOptions{})

				if err != nil || len(objList.Items) > 0 {
					// avoid timeout due to default automatically created openshift
					// configmaps: kube-root-ca.crt openshift-service-ca.crt
					if resource.Resource == "configmaps" {
						return builder.hasOnlyDefaultConfigMaps(objList, err)
					}

					return false, err
				}

				return true, err
			})
		if err != nil {
			klog.V(100).Infof("Failed to remove resources: %s in namespace: %s",
				resource.Resource, builder.Definition.Name)

			return err
		}
	}

	return nil
}

// hasOnlyDefaultConfigMaps returns true if only default configMaps are present in a namespace.
func (builder *Builder) hasOnlyDefaultConfigMaps(objList *unstructured.UnstructuredList, err error) (bool, error) {
	if err := common.Validate(builder); err != nil {
		return false, err
	}

	if err != nil {
		return false, err
	}

	if len(objList.Items) != 2 {
		return false, err
	}

	var existingConfigMaps []string
	for _, configMap := range objList.Items {
		existingConfigMaps = append(existingConfigMaps, configMap.GetName())
	}

	// return false if existing configmaps are NOT default pre-deployed openshift configmaps
	if !slices.Contains(existingConfigMaps, "kube-root-ca.crt") ||
		!slices.Contains(existingConfigMaps, "openshift-service-ca.crt") {
		return false, err
	}

	return true, nil
}
