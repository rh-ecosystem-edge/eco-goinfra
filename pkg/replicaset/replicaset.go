package replicaset

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	canNotApplyEmptyLabels     = "can not apply empty labels"
	canNotApplyLabelsWithAn    = "can not apply labels with an empty labelKey value"
	volumeNameParameterIsEmpty = "volume name parameter is empty"
)

var replicaSetGVK = appsv1.SchemeGroupVersion.WithKind("ReplicaSet")

// Builder provides struct for replicaset object containing connection to the cluster and the replicaset definitions.
type Builder struct {
	common.EmbeddableBuilder[appsv1.ReplicaSet, *appsv1.ReplicaSet]
	common.EmbeddableCreator[appsv1.ReplicaSet, Builder, *appsv1.ReplicaSet, *Builder]
	common.EmbeddableDeleter[appsv1.ReplicaSet, *appsv1.ReplicaSet]
	common.EmbeddableUpdater[appsv1.ReplicaSet, Builder, *appsv1.ReplicaSet, *Builder]
}

// AdditionalOptions additional options for replicaset object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

var retryInterval = time.Second * 3

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the ReplicaSet GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return replicaSetGVK
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(
	apiClient *clients.Settings,
	name, nsname string,
	labels map[string]string,
	containerSpec []corev1.Container) *Builder {
	klog.V(100).Infof(
		"Initializing new replicaset structure with the following params: "+
			"name: %s, namespace: %s, containerSpec %v",
		name, nsname, containerSpec)

	builder := common.NewNamespacedBuilder[appsv1.ReplicaSet, Builder](
		apiClient, appsv1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Labels = labels
	builder.Definition.Spec.Template.Spec.Containers = containerSpec

	if len(labels) == 0 {
		klog.V(100).Info("The labels of the replicaset is empty")

		builder.SetError(fmt.Errorf("replicaset 'labels' cannot be empty"))

		return builder
	}

	if len(containerSpec) == 0 {
		klog.V(100).Info("The containerSpec of the replicaset is empty")

		builder.SetError(fmt.Errorf("replicaset 'containerSpec' cannot be empty"))

		return builder
	}

	return builder
}

// Pull loads an existing replicaset into the Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	klog.V(100).Infof("Pulling existing replicaset name:%s under namespace:%s", name, nsname)

	return common.PullNamespacedBuilder[appsv1.ReplicaSet, Builder](
		context.TODO(), apiClient, appsv1.AddToScheme, name, nsname)
}

// WithLabel applies label to replicaset's definition.
func (builder *Builder) WithLabel(labels map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%v", fmt.Sprintf("Defining replicaset's labels to %v", labels))

	if len(labels) == 0 {
		klog.V(100).Info("The 'labels' of the replicaset is empty")

		builder.SetError(fmt.Errorf(canNotApplyEmptyLabels))

		return builder
	}

	for labelKey := range labels {
		if labelKey == "" {
			klog.V(100).Info("The 'labels' labelKey cannot be empty")

			builder.SetError(fmt.Errorf(canNotApplyLabelsWithAn))

			return builder
		}
	}

	builder.Definition.Labels = labels

	return builder
}

// WithNodeSelector applies nodeSelector to the replicaset definition.
func (builder *Builder) WithNodeSelector(nodeSelector map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying nodeSelector %s to replicaset %s in namespace %s",
		nodeSelector, builder.Definition.Name, builder.Definition.Namespace)

	if len(nodeSelector) == 0 {
		klog.V(100).Info("The 'nodeSelector' of the replicaset is empty")

		builder.SetError(fmt.Errorf("can not apply empty nodeSelector"))

		return builder
	}

	for key := range nodeSelector {
		if key == "" {
			klog.V(100).Info("The 'nodeSelector' key value cannot be empty")

			builder.SetError(fmt.Errorf("can not apply a nodeSelector with an empty key value"))

			return builder
		}
	}

	builder.Definition.Spec.Template.Spec.NodeSelector = nodeSelector

	return builder
}

// WithVolume defines Volume of replicaset under ContainerTemplateSpec.
func (builder *Builder) WithVolume(rsVolume corev1.Volume) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if rsVolume.Name == "" {
		klog.V(100).Info("The Volume name parameter is empty")

		builder.SetError(fmt.Errorf(volumeNameParameterIsEmpty))

		return builder
	}

	klog.V(100).Infof("Adding volume %s for replicaset %s container template in namespace %s",
		rsVolume.Name, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.Volumes = append(
		builder.Definition.Spec.Template.Spec.Volumes,
		rsVolume)

	return builder
}

// WithAdditionalContainerSpecs appends a list of container specs to the replicaset definition.
func (builder *Builder) WithAdditionalContainerSpecs(specs []corev1.Container) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Appending a list of container specs %v to replicaset %s in namespace %s",
		specs, builder.Definition.Name, builder.Definition.Namespace)

	if len(specs) == 0 {
		klog.V(100).Info("The container specs are empty")

		builder.SetError(fmt.Errorf("cannot accept empty list as container specs"))

		return builder
	}

	builder.Definition.Spec.Template.Spec.Containers = append(builder.Definition.Spec.Template.Spec.Containers, specs...)

	return builder
}

// CreateAndWaitUntilReady creates a replicaset in the cluster and waits until the replicaset is available.
func (builder *Builder) CreateAndWaitUntilReady(timeout time.Duration) (*Builder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Creating replicaset %s in namespace %s and waiting for the defined period "+
		"until it is ready",
		builder.Definition.Name, builder.Definition.Namespace)

	_, err := builder.Create()
	if err != nil {
		klog.V(100).Infof("Failed to create replicaset. Error is: '%s'", err.Error())

		return nil, err
	}

	// Polls every retryInterval to determine if replicaset is available.
	err = wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			object, err := common.Get(ctx, builder)
			if err != nil {
				return false, nil
			}

			builder.SetObject(object)

			if builder.Object.Status.ReadyReplicas == builder.Object.Status.Replicas {
				return true, nil
			}

			return false, err
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}

// DeleteAndWait deletes a replicaset and waits until it is removed from the cluster.
func (builder *Builder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Deleting replicaset %s in namespace %s and waiting for the defined period "+
		"until it is removed",
		builder.Definition.Name, builder.Definition.Namespace)

	if err := builder.Delete(); err != nil {
		return err
	}

	// Polls the replicaset every retryInterval until it is removed.
	return wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			_, err := common.Get(ctx, builder)
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			return false, nil
		})
}

// IsReady waits for the replicaset to reach expected number of pods in Ready state.
func (builder *Builder) IsReady(timeout time.Duration) bool {
	if err := common.Validate(builder); err != nil {
		return false
	}

	klog.V(100).Infof("Running periodic check until replicaset %s in namespace %s is ready or "+
		"timeout %s exceeded", builder.Definition.Name, builder.Definition.Namespace, timeout.String())

	// Polls every retryInterval to determine if replicaset is available.
	err := wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			if !common.Exists(ctx, builder) {
				return false, fmt.Errorf("replicaset %s is not present on cluster", builder.Definition.Name)
			}

			object, err := common.Get(ctx, builder)
			if err != nil {
				klog.V(100).Infof("Failed to get replicaset from cluster. Error is: '%s'", err.Error())

				return false, nil
			}

			builder.SetObject(object)

			if builder.Object.Status.ReadyReplicas == builder.Object.Status.Replicas {
				return true, nil
			}

			return false, err
		})

	return err == nil
}

// GetGVR returns the GroupVersionResource for replicaset.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"}
}
