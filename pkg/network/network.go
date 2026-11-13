package network

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	clusterNetworkName = "cluster"
)

// ConfigBuilder provides a struct for network object from the cluster and a network definition.
type ConfigBuilder struct {
	common.EmbeddableBuilder[configv1.Network, *configv1.Network]
}

// GetGVK returns the network.config GVK for this builder.
func (builder *ConfigBuilder) GetGVK() schema.GroupVersionKind {
	return configv1.GroupVersion.WithKind("Network")
}

// PullConfig loads an existing network into ConfigBuilder struct.
func PullConfig(apiClient *clients.Settings) (*ConfigBuilder, error) {
	return common.PullClusterScopedBuilder[configv1.Network, ConfigBuilder](
		context.TODO(), apiClient, configv1.Install, clusterNetworkName)
}
