package hive

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	hiveV1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/hive/api/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// ConfigBuilder provides struct for the HiveConfig object containing connection to
// the cluster and the HiveConfig definitions.
type ConfigBuilder struct {
	common.EmbeddableBuilder[hiveV1.HiveConfig, *hiveV1.HiveConfig]
	common.EmbeddableDeleter[hiveV1.HiveConfig, *hiveV1.HiveConfig]
	common.EmbeddableUpdater[hiveV1.HiveConfig, ConfigBuilder, *hiveV1.HiveConfig, *ConfigBuilder]
	common.EmbeddableWithOptions[hiveV1.HiveConfig, ConfigBuilder, *hiveV1.HiveConfig, *ConfigBuilder, ConfigAdditionalOptions]
}

// ConfigAdditionalOptions additional options for HiveConfig object.
type ConfigAdditionalOptions func(builder *ConfigBuilder) (*ConfigBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ConfigBuilder) AttachMixins() {
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the HiveConfig GVK for this builder.
func (builder *ConfigBuilder) GetGVK() schema.GroupVersionKind {
	return hiveV1.SchemeGroupVersion.WithKind("HiveConfig")
}

// NewConfigBuilder creates a new instance of ConfigBuilder.
func NewConfigBuilder(apiClient *clients.Settings, name string) *ConfigBuilder {
	klog.V(100).Infof(
		`Initializing new HiveConfig structure with the following params: name: %s`, name)

	return common.NewClusterScopedBuilder[hiveV1.HiveConfig, ConfigBuilder](
		apiClient, hiveV1.AddToScheme, name)
}

// PullConfig loads an existing HiveConfig into ConfigBuilder struct.
func PullConfig(apiClient *clients.Settings, name string) (*ConfigBuilder, error) {
	klog.V(100).Infof("Pulling existing HiveConfig name: %s", name)

	return common.PullClusterScopedBuilder[hiveV1.HiveConfig, ConfigBuilder](
		context.TODO(), apiClient, hiveV1.AddToScheme, name)
}
