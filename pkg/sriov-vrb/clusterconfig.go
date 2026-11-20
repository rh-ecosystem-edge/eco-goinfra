package sriovvrb

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	sriovvrbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/vrbtypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ClusterConfigBuilder provides struct for the SriovVrbClusterConfig object containing connection to
// the cluster and the SriovVrbClusterConfig definitions.
type ClusterConfigBuilder struct {
	common.EmbeddableBuilder[sriovvrbtypes.SriovVrbClusterConfig, *sriovvrbtypes.SriovVrbClusterConfig]
	common.EmbeddableCreator[
		sriovvrbtypes.SriovVrbClusterConfig, ClusterConfigBuilder, *sriovvrbtypes.SriovVrbClusterConfig,
		*ClusterConfigBuilder]
	common.EmbeddableDeleteReturner[
		sriovvrbtypes.SriovVrbClusterConfig, ClusterConfigBuilder, *sriovvrbtypes.SriovVrbClusterConfig,
		*ClusterConfigBuilder]
	common.EmbeddableForceUpdater[
		sriovvrbtypes.SriovVrbClusterConfig, ClusterConfigBuilder, *sriovvrbtypes.SriovVrbClusterConfig,
		*ClusterConfigBuilder]
	common.EmbeddableWithOptions[
		sriovvrbtypes.SriovVrbClusterConfig, ClusterConfigBuilder, *sriovvrbtypes.SriovVrbClusterConfig,
		*ClusterConfigBuilder, ClusterAdditionalOptions]
}

// ClusterAdditionalOptions additional options for SriovVrbclusterconfig object.
type ClusterAdditionalOptions func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ClusterConfigBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the SriovVrbClusterConfig GVK for this builder.
func (builder *ClusterConfigBuilder) GetGVK() schema.GroupVersionKind {
	return sriovvrbtypes.GroupVersion.WithKind("SriovVrbClusterConfig")
}

// NewClusterConfigBuilder creates a new instance of ClusterConfigBuilder.
func NewClusterConfigBuilder(apiClient *clients.Settings, name, nsname string) *ClusterConfigBuilder {
	return common.NewNamespacedBuilder[sriovvrbtypes.SriovVrbClusterConfig, ClusterConfigBuilder](
		apiClient, sriovvrbtypes.AddToScheme, name, nsname)
}

// PullClusterConfig retrieves an existing SriovVrbClusterConfig.io object from the cluster.
func PullClusterConfig(apiClient *clients.Settings, name, nsname string) (*ClusterConfigBuilder, error) {
	return common.PullNamespacedBuilder[sriovvrbtypes.SriovVrbClusterConfig, ClusterConfigBuilder](
		context.TODO(), apiClient, sriovvrbtypes.AddToScheme, name, nsname)
}

// GetSriovVrbClusterConfigIoGVR returns SriovVrbClusterConfig's GroupVersionResource.
func GetSriovVrbClusterConfigIoGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: ClusterConfigsResource,
	}
}
