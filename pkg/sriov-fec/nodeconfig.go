package sriovfec

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	sriovfectypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/fectypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NodeConfigBuilder provides struct for the SriovFecNodeConfig object containing connection to
// the cluster and the SriovFecNodeConfig definitions.
type NodeConfigBuilder struct {
	common.EmbeddableBuilder[sriovfectypes.SriovFecNodeConfig, *sriovfectypes.SriovFecNodeConfig]
	common.EmbeddableWithOptions[
		sriovfectypes.SriovFecNodeConfig,
		NodeConfigBuilder,
		*sriovfectypes.SriovFecNodeConfig,
		*NodeConfigBuilder,
		AdditionalOptions,
	]
	common.EmbeddableCreator[
		sriovfectypes.SriovFecNodeConfig,
		NodeConfigBuilder,
		*sriovfectypes.SriovFecNodeConfig,
		*NodeConfigBuilder,
	]
	common.EmbeddableDeleteReturner[
		sriovfectypes.SriovFecNodeConfig,
		NodeConfigBuilder,
		*sriovfectypes.SriovFecNodeConfig,
		*NodeConfigBuilder,
	]
	common.EmbeddableForceUpdater[
		sriovfectypes.SriovFecNodeConfig,
		NodeConfigBuilder,
		*sriovfectypes.SriovFecNodeConfig,
		*NodeConfigBuilder,
	]
}

// AdditionalOptions additional options for sriovfecnodeconfig object.
type AdditionalOptions func(builder *NodeConfigBuilder) (*NodeConfigBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *NodeConfigBuilder) AttachMixins() {
	builder.EmbeddableWithOptions.SetBase(builder)
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the SriovFecNodeConfig GVK for this builder.
func (builder *NodeConfigBuilder) GetGVK() schema.GroupVersionKind {
	return sriovfectypes.GroupVersion.WithKind("SriovFecNodeConfig")
}

// NewNodeConfigBuilder creates a new instance of NodeConfigBuilder. The label argument is retained for API
// compatibility; it does not modify the resource.
func NewNodeConfigBuilder(
	apiClient *clients.Settings,
	name, nsname string,
	_ map[string]string,
) *NodeConfigBuilder {
	return common.NewNamespacedBuilder[sriovfectypes.SriovFecNodeConfig, NodeConfigBuilder](
		apiClient, sriovfectypes.AddToScheme, name, nsname)
}

// Pull retrieves an existing SriovFecNodeConfig.io object from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*NodeConfigBuilder, error) {
	return common.PullNamespacedBuilder[sriovfectypes.SriovFecNodeConfig, NodeConfigBuilder](
		context.TODO(), apiClient, sriovfectypes.AddToScheme, name, nsname)
}

// GetSriovFecNodeConfigIoGVR returns SriovFecNodeConfig's GroupVersionResource which could be used for Clean function.
func GetSriovFecNodeConfigIoGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: NodeConfigsResource,
	}
}
