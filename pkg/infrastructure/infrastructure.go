package infrastructure

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	infrastructureName = "cluster"
)

// Builder provides a struct for infrastructure object from the cluster and a infrastructure definition.
type Builder struct {
	common.EmbeddableBuilder[configv1.Infrastructure, *configv1.Infrastructure]
}

// GetGVK returns the Infrastructure GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return configv1.GroupVersion.WithKind("Infrastructure")
}

// Pull loads an existing infrastructure into Builder struct.
func Pull(apiClient *clients.Settings) (*Builder, error) {
	return common.PullClusterScopedBuilder[configv1.Infrastructure, Builder](
		context.TODO(), apiClient, configv1.Install, infrastructureName)
}
