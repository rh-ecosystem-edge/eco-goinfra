package proxy

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	clusterProxyName = "cluster"
)

// Builder provides a struct for proxy object from the cluster and a proxy definition.
type Builder struct {
	common.EmbeddableBuilder[configv1.Proxy, *configv1.Proxy]
}

// GetGVK returns the Proxy GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return configv1.GroupVersion.WithKind("Proxy")
}

// Pull loads an existing proxy into Builder struct.
func Pull(apiClient *clients.Settings) (*Builder, error) {
	return common.PullClusterScopedBuilder[configv1.Proxy, Builder](
		context.TODO(), apiClient, configv1.Install, clusterProxyName)
}
