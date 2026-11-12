package sriov

import (
	"context"
	"fmt"
	"slices"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	sriovOperatorConfigName = "default"
	pluginMellanox          = "mellanox"
)

var (
	// allowedDisablePlugins represents all allowed plugins that could be disabled.
	allowedDisablePlugins = []string{pluginMellanox}
)

// OperatorConfigBuilder provides a struct for SriovOperatorConfig object from the cluster and
// a SriovOperatorConfig definition.
type OperatorConfigBuilder struct {
	common.EmbeddableBuilder[srIovV1.SriovOperatorConfig, *srIovV1.SriovOperatorConfig]
	common.EmbeddableCreator[srIovV1.SriovOperatorConfig, OperatorConfigBuilder,
		*srIovV1.SriovOperatorConfig, *OperatorConfigBuilder]
	common.EmbeddableDeleteReturner[srIovV1.SriovOperatorConfig, OperatorConfigBuilder,
		*srIovV1.SriovOperatorConfig, *OperatorConfigBuilder]
	common.EmbeddableUpdater[srIovV1.SriovOperatorConfig, OperatorConfigBuilder,
		*srIovV1.SriovOperatorConfig, *OperatorConfigBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *OperatorConfigBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the SriovOperatorConfig GVK for this builder.
func (builder *OperatorConfigBuilder) GetGVK() schema.GroupVersionKind {
	return srIovV1.GroupVersion.WithKind("SriovOperatorConfig")
}

// NewOperatorConfigBuilder creates new instance of OperatorConfigBuilder.
func NewOperatorConfigBuilder(apiClient *clients.Settings, nsname string) *OperatorConfigBuilder {
	return common.NewNamespacedBuilder[srIovV1.SriovOperatorConfig, OperatorConfigBuilder](
		apiClient, srIovV1.AddToScheme, sriovOperatorConfigName, nsname)
}

// PullOperatorConfig loads an existing SriovOperatorConfig into OperatorConfigBuilder struct.
func PullOperatorConfig(apiClient *clients.Settings, nsname string) (*OperatorConfigBuilder, error) {
	return common.PullNamespacedBuilder[srIovV1.SriovOperatorConfig, OperatorConfigBuilder](
		context.TODO(), apiClient, srIovV1.AddToScheme, sriovOperatorConfigName, nsname)
}

// WithInjector configures enableInjector in the SriovOperatorConfig.
func (builder *OperatorConfigBuilder) WithInjector(enable bool) *OperatorConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Configuring enableInjector %t to SriovOperatorConfig object %s",
		enable, builder.Definition.Name,
	)

	builder.Definition.Spec.EnableInjector = enable

	return builder
}

// WithOperatorWebhook configures enableOperatorWebhook in the SriovOperatorConfig.
func (builder *OperatorConfigBuilder) WithOperatorWebhook(enable bool) *OperatorConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Configuring WithOperatorWebhook %t to SriovOperatorConfig object %s",
		enable, builder.Definition.Name,
	)

	builder.Definition.Spec.EnableOperatorWebhook = enable

	return builder
}

// WithConfigDaemonNodeSelector configures configDaemonNodeSelector in the SriovOperatorConfig.
func (builder *OperatorConfigBuilder) WithConfigDaemonNodeSelector(
	configDaemonNodeSelector map[string]string) *OperatorConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Configuring configDaemonNodeSelector %s in SriovOperatorConfig object %s in namespace %s",
		configDaemonNodeSelector, builder.Definition.Name, builder.Definition.Namespace,
	)

	if len(configDaemonNodeSelector) == 0 {
		klog.V(100).Info("The 'configDaemonNodeSelector' of the SriovOperatorConfig is empty")

		builder.SetError(fmt.Errorf("can not apply empty configDaemonNodeSelector"))

		return builder
	}

	for selectorKey := range configDaemonNodeSelector {
		if selectorKey == "" {
			klog.V(100).Info("The 'configDaemonNodeSelector' selectorKey cannot be empty")

			builder.SetError(fmt.Errorf("can not apply configDaemonNodeSelector with an empty selectorKey value"))

			return builder
		}
	}

	builder.Definition.Spec.ConfigDaemonNodeSelector = configDaemonNodeSelector

	return builder
}

// WithDisablePlugins configures disablePlugins in the SriovOperatorConfig.
func (builder *OperatorConfigBuilder) WithDisablePlugins(plugins []string) *OperatorConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Configuring disablePlugins %v in SriovOperatorConfig object %s",
		plugins, builder.Definition.Name,
	)

	var pluginSlice srIovV1.PluginNameSlice

	for _, plugin := range plugins {
		if !slices.Contains(allowedDisablePlugins, plugin) {
			klog.V(100).Infof("error to add plugin %s, allowed modes are %v", plugin, allowedDisablePlugins)

			builder.SetError(fmt.Errorf("invalid plugin parameter"))

			return builder
		}

		pluginSlice = append(pluginSlice, srIovV1.PluginNameValue(plugin))
	}

	builder.Definition.Spec.DisablePlugins = pluginSlice

	return builder
}

// RemoveDisablePlugins deletes disablePlugins in the SriovOperatorConfig.
func (builder *OperatorConfigBuilder) RemoveDisablePlugins() *OperatorConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Removing disablePlugins in SriovOperatorConfig object %s", builder.Definition.Name)
	builder.Definition.Spec.DisablePlugins = nil

	return builder
}
