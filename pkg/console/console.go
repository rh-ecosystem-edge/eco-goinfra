package console

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Builder provides a struct for console object from the cluster and a console definition.
type Builder struct {
	common.EmbeddableBuilder[configv1.Console, *configv1.Console]
	common.EmbeddableCreator[configv1.Console, Builder, *configv1.Console, *Builder]
	common.EmbeddableDeleter[configv1.Console, *configv1.Console]
	common.EmbeddableUpdater[configv1.Console, Builder, *configv1.Console, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the Console GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return configv1.GroupVersion.WithKind("Console")
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(apiClient *clients.Settings, name string) *Builder {
	return common.NewClusterScopedBuilder[configv1.Console, Builder](apiClient, configv1.Install, name)
}

// Pull loads an existing console into the Builder struct.
func Pull(apiClient *clients.Settings, name string) (*Builder, error) {
	return common.PullClusterScopedBuilder[configv1.Console, Builder](
		context.TODO(), apiClient, configv1.Install, name)
}
