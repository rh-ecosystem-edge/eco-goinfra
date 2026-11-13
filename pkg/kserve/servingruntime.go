package kserve

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	kservev1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kserve/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var servingRuntimeGVK = kservev1alpha1.GroupVersion.WithKind("ServingRuntime")

// ServingRuntimeBuilder provides struct for the ServingRuntime object
// containing connection to the cluster and the ServingRuntime definition.
type ServingRuntimeBuilder struct {
	common.EmbeddableBuilder[kservev1alpha1.ServingRuntime, *kservev1alpha1.ServingRuntime]
	common.EmbeddableCreator[
		kservev1alpha1.ServingRuntime,
		ServingRuntimeBuilder,
		*kservev1alpha1.ServingRuntime,
		*ServingRuntimeBuilder,
	]
	common.EmbeddableDeleteReturner[
		kservev1alpha1.ServingRuntime,
		ServingRuntimeBuilder,
		*kservev1alpha1.ServingRuntime,
		*ServingRuntimeBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ServingRuntimeBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
}

// GetGVK returns the ServingRuntime GVK for this builder.
func (builder *ServingRuntimeBuilder) GetGVK() schema.GroupVersionKind {
	return servingRuntimeGVK
}

// NewServingRuntimeBuilder creates a new instance of ServingRuntimeBuilder.
func NewServingRuntimeBuilder(
	apiClient *clients.Settings,
	name, namespace string) *ServingRuntimeBuilder {
	builder := common.NewNamespacedBuilder[kservev1alpha1.ServingRuntime, ServingRuntimeBuilder](
		apiClient, kservev1alpha1.AddToScheme, name, namespace)
	multiModel := false
	builder.Definition.Spec.MultiModel = &multiModel

	return builder
}

// WithModelFormat adds a supported model format to the ServingRuntime.
func (builder *ServingRuntimeBuilder) WithModelFormat(
	name string, autoSelect bool) *ServingRuntimeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding model format %s to ServingRuntime %s",
		name, builder.Definition.Name)

	if name == "" {
		builder.SetError(fmt.Errorf("ServingRuntime model format 'name' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.SupportedModelFormats = append(
		builder.Definition.Spec.SupportedModelFormats,
		kservev1alpha1.SupportedModelFormat{
			Name:       name,
			AutoSelect: autoSelect,
		})

	return builder
}

// WithContainer adds a container to the ServingRuntime.
func (builder *ServingRuntimeBuilder) WithContainer(
	container corev1.Container) *ServingRuntimeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding container %s to ServingRuntime %s",
		container.Name, builder.Definition.Name)

	builder.Definition.Spec.Containers = append(
		builder.Definition.Spec.Containers, container)

	return builder
}

// WithVolume adds a volume to the ServingRuntime.
func (builder *ServingRuntimeBuilder) WithVolume(
	volume corev1.Volume) *ServingRuntimeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding volume %s to ServingRuntime %s",
		volume.Name, builder.Definition.Name)

	builder.Definition.Spec.Volumes = append(
		builder.Definition.Spec.Volumes, volume)

	return builder
}

// WithAnnotation adds an annotation to the ServingRuntime.
func (builder *ServingRuntimeBuilder) WithAnnotation(
	key, value string) *ServingRuntimeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if builder.Definition.Annotations == nil {
		builder.Definition.Annotations = make(map[string]string)
	}

	builder.Definition.Annotations[key] = value

	return builder
}

// PullServingRuntime retrieves an existing ServingRuntime from the cluster.
func PullServingRuntime(
	apiClient *clients.Settings, name, namespace string) (*ServingRuntimeBuilder, error) {
	return common.PullNamespacedBuilder[kservev1alpha1.ServingRuntime, ServingRuntimeBuilder](
		context.TODO(), apiClient, kservev1alpha1.AddToScheme, name, namespace)
}
