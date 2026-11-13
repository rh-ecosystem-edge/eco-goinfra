package statefulset

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

// Builder provides struct for statefulset object containing connection to the cluster and the statefulset definitions.
type Builder struct {
	common.EmbeddableBuilder[appsv1.StatefulSet, *appsv1.StatefulSet]
	common.EmbeddableWithOptions[appsv1.StatefulSet, Builder, *appsv1.StatefulSet, *Builder, AdditionalOptions]
	common.EmbeddableCreator[appsv1.StatefulSet, Builder, *appsv1.StatefulSet, *Builder]
	common.EmbeddableDeleter[appsv1.StatefulSet, *appsv1.StatefulSet]
}

// AdditionalOptions additional options for StatefulSet object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableWithOptions.SetBase(builder)
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the StatefulSet GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return appsv1.SchemeGroupVersion.WithKind("StatefulSet")
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(
	apiClient *clients.Settings,
	name string,
	nsname string,
	labels map[string]string,
	containerSpec *corev1.Container) *Builder {
	klog.V(100).Infof(
		"Initializing new statefulset structure with the following params: "+
			"name: %s, namespace: %s, labels: %s, containerSpec %v",
		name, nsname, labels, containerSpec)

	builder := common.NewNamespacedBuilder[appsv1.StatefulSet, Builder](
		apiClient, appsv1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	builder.Definition.Spec.Template.Labels = labels

	if labels == nil {
		klog.V(100).Info("There are no labels for the statefulset")

		builder.SetError(fmt.Errorf("statefulset 'labels' cannot be empty"))

		return builder
	}

	if containerSpec == nil {
		klog.V(100).Info("The container spec for the statefulset is nil")

		builder.SetError(fmt.Errorf("statefulset 'containerSpec' cannot be nil"))

		return builder
	}

	return builder.WithAdditionalContainerSpecs([]corev1.Container{*containerSpec})
}

// WithAdditionalContainerSpecs appends a list of container specs to the statefulset definition.
func (builder *Builder) WithAdditionalContainerSpecs(specs []corev1.Container) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Appending a list of container specs %v to statefulset %s in namespace %s",
		specs, builder.Definition.Name, builder.Definition.Namespace)

	if specs == nil {
		klog.V(100).Info("The container specs are empty")

		builder.SetError(fmt.Errorf("cannot accept nil or empty list as container specs"))

		return builder
	}

	builder.Definition.Spec.Template.Spec.Containers = append(builder.Definition.Spec.Template.Spec.Containers, specs...)

	return builder
}

// WithPodAnnotations sets annotations on the pod template for pods managed by the statefulset.
func (builder *Builder) WithPodAnnotations(annotations map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		klog.V(100).Info("Failed to validate StatefulSet builder")

		return builder
	}

	klog.V(100).Infof("Setting pod annotations %v for statefulset %s in namespace %s",
		annotations, builder.Definition.Name, builder.Definition.Namespace)

	if len(annotations) == 0 {
		klog.V(100).Info("The pod annotations are empty")

		builder.SetError(fmt.Errorf("cannot accept nil or empty annotations"))

		return builder
	}

	if builder.Definition.Spec.Template.Annotations == nil {
		builder.Definition.Spec.Template.Annotations = make(map[string]string)
	}

	for key, value := range annotations {
		builder.Definition.Spec.Template.Annotations[key] = value
	}

	return builder
}

// Pull loads an existing statefulset into Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	return common.PullNamespacedBuilder[appsv1.StatefulSet, Builder](
		context.TODO(), apiClient, appsv1.AddToScheme, name, nsname)
}

// IsReady periodically checks if statefulset is in ready status.
func (builder *Builder) IsReady(timeout time.Duration) bool {
	if err := common.Validate(builder); err != nil {
		return false
	}

	klog.V(100).Infof("Running periodic check until statefulset %s in namespace %s is ready",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return false
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			object, err := common.Get(ctx, builder)
			if err != nil {
				return false, err
			}

			builder.Object = object

			if builder.Object.Status.ReadyReplicas > 0 &&
				builder.Object.Status.Replicas == builder.Object.Status.ReadyReplicas {
				return true, nil
			}

			return false, nil
		})

	return err == nil
}

// GetGVR returns pod's GroupVersionResource which could be used for Clean function.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
}
