package sriovfec

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	sriovfectypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/fectypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ClusterConfigBuilder provides struct for the SriovFecClusterConfig object containing connection to
// the cluster and the SriovFecClusterConfig definitions.
type ClusterConfigBuilder struct {
	common.EmbeddableBuilder[sriovfectypes.SriovFecClusterConfig, *sriovfectypes.SriovFecClusterConfig]
	common.EmbeddableWithOptions[
		sriovfectypes.SriovFecClusterConfig,
		ClusterConfigBuilder,
		*sriovfectypes.SriovFecClusterConfig,
		*ClusterConfigBuilder,
		ClusterAdditionalOptions,
	]
	common.EmbeddableCreator[
		sriovfectypes.SriovFecClusterConfig,
		ClusterConfigBuilder,
		*sriovfectypes.SriovFecClusterConfig,
		*ClusterConfigBuilder,
	]
	common.EmbeddableDeleteReturner[
		sriovfectypes.SriovFecClusterConfig,
		ClusterConfigBuilder,
		*sriovfectypes.SriovFecClusterConfig,
		*ClusterConfigBuilder,
	]
	common.EmbeddableForceUpdater[
		sriovfectypes.SriovFecClusterConfig,
		ClusterConfigBuilder,
		*sriovfectypes.SriovFecClusterConfig,
		*ClusterConfigBuilder,
	]
}

// ClusterAdditionalOptions additional options for SriovFecClusterConfig object.
type ClusterAdditionalOptions func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ClusterConfigBuilder) AttachMixins() {
	builder.EmbeddableWithOptions.SetBase(builder)
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the SriovFecClusterConfig GVK for this builder.
func (builder *ClusterConfigBuilder) GetGVK() schema.GroupVersionKind {
	return sriovfectypes.GroupVersion.WithKind("SriovFecClusterConfig")
}

// NewClusterConfigBuilder creates a new instance of ClusterConfigBuilder.
func NewClusterConfigBuilder(apiClient *clients.Settings, name, nsname string) *ClusterConfigBuilder {
	return common.NewNamespacedBuilder[sriovfectypes.SriovFecClusterConfig, ClusterConfigBuilder](
		apiClient, sriovfectypes.AddToScheme, name, nsname)
}

// PullClusterConfig retrieves an existing SriovFecClusterConfig.io object from the cluster.
func PullClusterConfig(apiClient *clients.Settings, name, nsname string) (*ClusterConfigBuilder, error) {
	return common.PullNamespacedBuilder[sriovfectypes.SriovFecClusterConfig, ClusterConfigBuilder](
		context.TODO(), apiClient, sriovfectypes.AddToScheme, name, nsname)
}

// GetSriovFecClusterConfigIoGVR returns SriovFecClusterConfig's GroupVersionResource.
func GetSriovFecClusterConfigIoGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: ClusterConfigsResource,
	}
}
