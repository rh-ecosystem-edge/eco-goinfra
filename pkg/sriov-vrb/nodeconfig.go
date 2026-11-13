package sriovvrb

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	sriovvrbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/vrbtypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NodeConfigBuilder provides struct for the SriovVrbNodeConfig object containing connection to
// the cluster and the SriovVrbNodeConfig definitions.
type NodeConfigBuilder struct {
	common.EmbeddableBuilder[sriovvrbtypes.SriovVrbNodeConfig, *sriovvrbtypes.SriovVrbNodeConfig]
	common.EmbeddableCreator[
		sriovvrbtypes.SriovVrbNodeConfig, NodeConfigBuilder, *sriovvrbtypes.SriovVrbNodeConfig, *NodeConfigBuilder]
	common.EmbeddableDeleteReturner[
		sriovvrbtypes.SriovVrbNodeConfig, NodeConfigBuilder, *sriovvrbtypes.SriovVrbNodeConfig, *NodeConfigBuilder]
	common.EmbeddableForceUpdater[
		sriovvrbtypes.SriovVrbNodeConfig, NodeConfigBuilder, *sriovvrbtypes.SriovVrbNodeConfig, *NodeConfigBuilder]
	common.EmbeddableWithOptions[
		sriovvrbtypes.SriovVrbNodeConfig, NodeConfigBuilder, *sriovvrbtypes.SriovVrbNodeConfig, *NodeConfigBuilder,
		NodeAdditionalOptions]
}

// NodeAdditionalOptions additional options for SriovVrbnodeconfig object.
type NodeAdditionalOptions func(builder *NodeConfigBuilder) (*NodeConfigBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *NodeConfigBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the SriovVrbNodeConfig GVK for this builder.
func (builder *NodeConfigBuilder) GetGVK() schema.GroupVersionKind {
	return sriovvrbtypes.GroupVersion.WithKind("SriovVrbNodeConfig")
}

// NewNodeConfigBuilder creates a new instance of NodeConfigBuilder.
func NewNodeConfigBuilder(apiClient *clients.Settings, name, nsname string) *NodeConfigBuilder {
	return common.NewNamespacedBuilder[sriovvrbtypes.SriovVrbNodeConfig, NodeConfigBuilder](
		apiClient, sriovvrbtypes.AddToScheme, name, nsname)
}

// PullNodeConfig retrieves an existing SriovVrbNodeConfig.io object from the cluster.
func PullNodeConfig(apiClient *clients.Settings, name, nsname string) (*NodeConfigBuilder, error) {
	return common.PullNamespacedBuilder[sriovvrbtypes.SriovVrbNodeConfig, NodeConfigBuilder](
		context.TODO(), apiClient, sriovvrbtypes.AddToScheme, name, nsname)
}

// GetSriovVrbNodeConfigIoGVR returns SriovVrbNodeConfig's GroupVersionResource which could be used for Clean function.
func GetSriovVrbNodeConfigIoGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: NodeConfigsResource,
	}
}
