package kmm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	mcmV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm-hub/v1beta1"
	moduleV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var managedClusterModuleGVK = mcmV1Beta1.GroupVersion.WithKind("ManagedClusterModule")

// ManagedClusterModuleBuilder provides struct for the managedclustermodule object containing connection
// to cluster and the managedclustermodule definitions.
type ManagedClusterModuleBuilder struct {
	common.EmbeddableBuilder[mcmV1Beta1.ManagedClusterModule, *mcmV1Beta1.ManagedClusterModule]
	common.EmbeddableCreator[
		mcmV1Beta1.ManagedClusterModule,
		ManagedClusterModuleBuilder,
		*mcmV1Beta1.ManagedClusterModule,
		*ManagedClusterModuleBuilder,
	]
	common.EmbeddableDeleteReturner[
		mcmV1Beta1.ManagedClusterModule,
		ManagedClusterModuleBuilder,
		*mcmV1Beta1.ManagedClusterModule,
		*ManagedClusterModuleBuilder,
	]
	common.EmbeddableUpdater[
		mcmV1Beta1.ManagedClusterModule,
		ManagedClusterModuleBuilder,
		*mcmV1Beta1.ManagedClusterModule,
		*ManagedClusterModuleBuilder,
	]
	common.EmbeddableWithOptions[
		mcmV1Beta1.ManagedClusterModule,
		ManagedClusterModuleBuilder,
		*mcmV1Beta1.ManagedClusterModule,
		*ManagedClusterModuleBuilder,
		ManagedClusterModuleAdditionalOptions,
	]
}

// ManagedClusterModuleAdditionalOptions additional options for managedclustermodule object.
type ManagedClusterModuleAdditionalOptions func(builder *ManagedClusterModuleBuilder) (
	*ManagedClusterModuleBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ManagedClusterModuleBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the ManagedClusterModule GVK for this builder.
func (builder *ManagedClusterModuleBuilder) GetGVK() schema.GroupVersionKind {
	return managedClusterModuleGVK
}

// NewManagedClusterModuleBuilder creates a new instance of ManagedClusterModuleBuilder.
func NewManagedClusterModuleBuilder(apiClient *clients.Settings, name, nsname string) *ManagedClusterModuleBuilder {
	return common.NewNamespacedBuilder[mcmV1Beta1.ManagedClusterModule, ManagedClusterModuleBuilder](
		apiClient, mcmV1Beta1.AddToScheme, name, nsname)
}

// WithModuleSpec sets the ModuleSpec.
func (builder *ManagedClusterModuleBuilder) WithModuleSpec(
	moduleSpec moduleV1Beta1.ModuleSpec) *ManagedClusterModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	builder.Definition.Spec.ModuleSpec = moduleSpec

	return builder
}

// WithSpokeNamespace sets the namespace where the module will be deployed on the spoke.
func (builder *ManagedClusterModuleBuilder) WithSpokeNamespace(
	spokeNamespace string) *ManagedClusterModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if spokeNamespace == "" {
		builder.SetError(fmt.Errorf("invalid 'spokeNamespace' argument cannot be nil"))

		return builder
	}

	builder.Definition.Spec.SpokeNamespace = spokeNamespace

	return builder
}

// WithSelector sets the selector for the managedclustermodule object.
func (builder *ManagedClusterModuleBuilder) WithSelector(
	selector map[string]string) *ManagedClusterModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if len(selector) == 0 {
		builder.SetError(fmt.Errorf("invalid 'selector' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Selector = selector

	return builder
}

// WithImageRebuildTriggerGeneration adds the specified image rebuild trigger generation to the ManagedClusterModule.
// Setting this to a different value than the current status will trigger re-verification of the module image
// on spoke clusters. If the image is missing on spoke, a rebuild is triggered. If the image exists, no rebuild occurs.
func (builder *ManagedClusterModuleBuilder) WithImageRebuildTriggerGeneration(
	generation int) *ManagedClusterModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating ManagedClusterModule %s in namespace %s with ImageRebuildTriggerGeneration: %d",
		builder.Definition.Name, builder.Definition.Namespace, generation)

	builder.Definition.Spec.ModuleSpec.ImageRebuildTriggerGeneration = &generation

	return builder
}

// PullManagedClusterModule pulls existing module from cluster.
func PullManagedClusterModule(apiClient *clients.Settings, name, nsname string) (*ManagedClusterModuleBuilder, error) {
	return common.PullNamespacedBuilder[mcmV1Beta1.ManagedClusterModule, ManagedClusterModuleBuilder](
		context.TODO(), apiClient, mcmV1Beta1.AddToScheme, name, nsname)
}
