package daemonset

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var daemonSetGVK = appsv1.SchemeGroupVersion.WithKind("DaemonSet")

// Builder provides struct for daemonset object containing connection to the cluster and the daemonset definitions.
type Builder struct {
	common.EmbeddableBuilder[appsv1.DaemonSet, *appsv1.DaemonSet]
	common.EmbeddableCreator[appsv1.DaemonSet, Builder, *appsv1.DaemonSet, *Builder]
	common.EmbeddableUpdater[appsv1.DaemonSet, Builder, *appsv1.DaemonSet, *Builder]
	common.EmbeddableDeleter[appsv1.DaemonSet, *appsv1.DaemonSet]
	common.EmbeddableWithOptions[appsv1.DaemonSet, Builder, *appsv1.DaemonSet, *Builder, AdditionalOptions]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the DaemonSet GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return daemonSetGVK
}

// AdditionalOptions additional options for daemonset object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

var retryInterval = time.Second * 3

// NewBuilder creates a new instance of Builder.
func NewBuilder(
	apiClient *clients.Settings, name, nsname string, labels map[string]string, containerSpec corev1.Container) *Builder {
	klog.V(100).Infof(
		"Initializing new daemonset structure with the following params: "+
			"name: %s, namespace: %s, labels: %s, containerSpec %v",
		name, nsname, labels, containerSpec)

	builder := common.NewNamespacedBuilder[appsv1.DaemonSet, Builder](
		apiClient, appsv1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	builder.Definition.Spec.Template.Labels = labels
	builder.WithAdditionalContainerSpecs([]corev1.Container{containerSpec})

	if len(labels) == 0 {
		klog.V(100).Info("There are no labels for the daemonset")

		builder.SetError(fmt.Errorf("daemonset 'labels' cannot be empty"))

		return builder
	}

	return builder
}

// Pull loads an existing daemonSet into the Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	klog.V(100).Infof("Pulling existing daemonset name:%s under namespace:%s", name, nsname)

	return common.PullNamespacedBuilder[appsv1.DaemonSet, Builder](
		context.TODO(), apiClient, appsv1.AddToScheme, name, nsname)
}

// WithNodeSelector applies nodeSelector to the daemonset definition.
func (builder *Builder) WithNodeSelector(selector map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying nodeSelector %s to daemonset %s in namespace %s",
		selector, builder.Definition.Name, builder.Definition.Namespace)

	if len(selector) == 0 {
		klog.V(100).Info("The nodeselector is empty")

		builder.SetError(fmt.Errorf("cannot accept empty map as nodeselector"))

		return builder
	}

	builder.Definition.Spec.Template.Spec.NodeSelector = selector

	return builder
}

// WithHostNetwork applies HostNetwork to daemonset definition.
func (builder *Builder) WithHostNetwork() *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Enabling hostnetwork flag to daemonset %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.HostNetwork = true

	return builder
}

// WithPodAffinity applies pod's Affinity to daemonset definition.
func (builder *Builder) WithPodAffinity(podAffinity *corev1.Affinity) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if podAffinity == nil {
		klog.V(100).Info("The Affinity parameter is empty")

		builder.SetError(fmt.Errorf("affinity parameter is empty"))

		return builder
	}

	klog.V(100).Infof("Adding pod affinity to daemonset %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.Affinity = podAffinity

	return builder
}

// WithVolume defines Volume of daemonset under PodTemplateSpec.
func (builder *Builder) WithVolume(dsVolume corev1.Volume) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if dsVolume.Name == "" {
		klog.V(100).Info("The Volume name parameter is empty")

		builder.SetError(fmt.Errorf("volume name parameter is empty"))

		return builder
	}

	klog.V(100).Infof("Adding volume %s for daemonset %s pod template in namespace %s",
		dsVolume.Name, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.Volumes = append(
		builder.Definition.Spec.Template.Spec.Volumes,
		dsVolume)

	return builder
}

// WithAdditionalContainerSpecs appends a list of container specs to the daemonset definition.
func (builder *Builder) WithAdditionalContainerSpecs(specs []corev1.Container) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Appending a list of container specs %v to daemonset %s in namespace %s",
		specs, builder.Definition.Name, builder.Definition.Namespace)

	if len(specs) == 0 {
		klog.V(100).Info("The container specs are empty")

		builder.SetError(fmt.Errorf("cannot accept empty list as container specs"))

		return builder
	}

	builder.Definition.Spec.Template.Spec.Containers = append(builder.Definition.Spec.Template.Spec.Containers, specs...)

	return builder
}

// CreateAndWaitUntilReady creates a daemonset in the cluster and waits until the daemonset is available.
func (builder *Builder) CreateAndWaitUntilReady(timeout time.Duration) (*Builder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Creating daemonset %s in namespace %s and waiting for the defined period until it is ready",
		builder.Definition.Name, builder.Definition.Namespace)

	_, err := builder.Create()
	if err != nil {
		klog.V(100).Infof("Failed to create daemonset. Error is: '%s'", err.Error())

		return nil, err
	}

	// Polls every retryInterval to determine if daemonset is available.
	err = wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			// Check if the daemonset is ready.
			//nolint:gocritic
			if builder.Object.Status.DesiredNumberScheduled == builder.Object.Status.CurrentNumberScheduled &&
				builder.Object.Status.DesiredNumberScheduled == builder.Object.Status.NumberReady &&
				builder.Object.Status.DesiredNumberScheduled == builder.Object.Status.NumberAvailable &&
				builder.Object.Status.NumberUnavailable == 0 &&
				builder.Object.Status.NumberReady > 0 {
				return true, nil
			}

			// If none of the conditions above are met, evaluate the daemonset conditions.
			for _, condition := range builder.Object.Status.Conditions {
				if condition.Type == "Available" {
					return condition.Status == "True", nil
				}
			}

			return false, err
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}

// DeleteAndWait deletes a daemonset and waits until it is removed from the cluster.
func (builder *Builder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Deleting daemonset %s in namespace %s and waiting for the defined period until it is removed",
		builder.Definition.Name, builder.Definition.Namespace)

	if err := builder.Delete(); err != nil {
		return err
	}

	// Polls the daemonset every retryInterval until it is removed.
	return wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(context.Context) (bool, error) {
			_, err := builder.Get()
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			return false, nil
		})
}

// IsReady waits for the daemonset to reach expected number of pods in Ready state.
func (builder *Builder) IsReady(timeout time.Duration) bool {
	if err := common.Validate(builder); err != nil {
		return false
	}

	klog.V(100).Infof("Running periodic check until daemonset %s in namespace %s is ready or "+
		"timeout %s exceeded", builder.Definition.Name, builder.Definition.Namespace, timeout.String())

	// Polls every retryInterval to determine if daemonset is available.
	err := wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("Failed to get daemonset from cluster. Error is: '%s'", err.Error())

				return false, nil
			}

			if builder.Object.Status.NumberReady == builder.Object.Status.DesiredNumberScheduled {
				return true, nil
			}

			if builder.Object.Status.NumberReady == builder.Object.Status.UpdatedNumberScheduled &&
				builder.Object.Status.UpdatedNumberScheduled != 0 {
				return true, nil
			}

			return false, err
		})

	return err == nil
}

// GetGVR returns the GroupVersionResource for the daemonset.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "daemonsets",
	}
}
