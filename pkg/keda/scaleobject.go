package keda

import (
	"context"
	"fmt"

	kedav2v1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var scaledObjectGVK = kedav2v1alpha1.GroupVersion.WithKind("ScaledObject")

// ScaledObjectBuilder provides a struct for ScaledObject object from the cluster
// and a ScaledObject definition.
type ScaledObjectBuilder struct {
	common.EmbeddableBuilder[kedav2v1alpha1.ScaledObject, *kedav2v1alpha1.ScaledObject]
	common.EmbeddableCreator[
		kedav2v1alpha1.ScaledObject, ScaledObjectBuilder, *kedav2v1alpha1.ScaledObject, *ScaledObjectBuilder]
	common.EmbeddableDeleteReturner[
		kedav2v1alpha1.ScaledObject, ScaledObjectBuilder, *kedav2v1alpha1.ScaledObject, *ScaledObjectBuilder]
	common.EmbeddableUpdater[
		kedav2v1alpha1.ScaledObject, ScaledObjectBuilder, *kedav2v1alpha1.ScaledObject, *ScaledObjectBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ScaledObjectBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the ScaledObject GVK for this builder.
func (builder *ScaledObjectBuilder) GetGVK() schema.GroupVersionKind {
	return scaledObjectGVK
}

// NewScaledObjectBuilder creates a new instance of ScaledObjectBuilder.
func NewScaledObjectBuilder(apiClient *clients.Settings, name, nsname string) *ScaledObjectBuilder {
	return common.NewNamespacedBuilder[kedav2v1alpha1.ScaledObject, ScaledObjectBuilder](
		apiClient, kedav2v1alpha1.AddToScheme, name, nsname)
}

// PullScaledObject pulls existing scaledObject from cluster.
func PullScaledObject(apiClient *clients.Settings, name, nsname string) (*ScaledObjectBuilder, error) {
	return common.PullNamespacedBuilder[kedav2v1alpha1.ScaledObject, ScaledObjectBuilder](
		context.TODO(), apiClient, kedav2v1alpha1.AddToScheme, name, nsname)
}

// WithTriggers sets the scaledObject operator's maxReplicaCount.
func (builder *ScaledObjectBuilder) WithTriggers(
	triggers []kedav2v1alpha1.ScaleTriggers) *ScaledObjectBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding triggers to scaledObject %s in namespace %s; triggers %v",
		builder.Definition.Name, builder.Definition.Namespace, triggers)

	if len(triggers) == 0 {
		klog.V(100).Info("'triggers' argument cannot be empty")

		builder.SetError(fmt.Errorf("'triggers' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Triggers = triggers

	return builder
}

// WithMaxReplicaCount sets the scaledObject operator's maxReplicaCount.
func (builder *ScaledObjectBuilder) WithMaxReplicaCount(maxReplicaCount int32) *ScaledObjectBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding maxReplicaCount to scaledObject %s in namespace %s; maxReplicaCount %v",
		builder.Definition.Name, builder.Definition.Namespace, maxReplicaCount)

	builder.Definition.Spec.MaxReplicaCount = &maxReplicaCount

	return builder
}

// WithMinReplicaCount sets the scaledObject operator's minReplicaCount.
func (builder *ScaledObjectBuilder) WithMinReplicaCount(minReplicaCount int32) *ScaledObjectBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding minReplicaCount to scaledObject %s in namespace %s; minReplicaCount %v",
		builder.Definition.Name, builder.Definition.Namespace, minReplicaCount)

	builder.Definition.Spec.MinReplicaCount = &minReplicaCount

	return builder
}

// WithCooldownPeriod sets the scaledObject operator's cooldownPeriod.
func (builder *ScaledObjectBuilder) WithCooldownPeriod(cooldownPeriod int32) *ScaledObjectBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding cooldownPeriod to scaledObject %s in namespace %s; cooldownPeriod %v",
		builder.Definition.Name, builder.Definition.Namespace, cooldownPeriod)

	builder.Definition.Spec.CooldownPeriod = &cooldownPeriod

	return builder
}

// WithPollingInterval sets the scaledObject operator's pollingInterval.
func (builder *ScaledObjectBuilder) WithPollingInterval(pollingInterval int32) *ScaledObjectBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding pollingInterval to scaledObject %s in namespace %s; pollingInterval %v",
		builder.Definition.Name, builder.Definition.Namespace, pollingInterval)

	builder.Definition.Spec.PollingInterval = &pollingInterval

	return builder
}

// WithScaleTargetRef sets the scaledObject operator's scaleTargetRef.
func (builder *ScaledObjectBuilder) WithScaleTargetRef(
	scaleTargetRef kedav2v1alpha1.ScaleTarget) *ScaledObjectBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding scaleTargetRef to scaledObject %s in namespace %s; scaleTargetRef %v",
		builder.Definition.Name, builder.Definition.Namespace, scaleTargetRef)

	builder.Definition.Spec.ScaleTargetRef = &scaleTargetRef

	return builder
}
