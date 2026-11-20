package ptp

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	ptpv1 "github.com/k8snetworkplumbingwg/ptp-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var ptpConfigGVK = ptpv1.GroupVersion.WithKind("PtpConfig")

// PtpConfigBuilder provides a struct for the PtpConfig resource containing a connection to the cluster and the
// PtpConfig definition.
type PtpConfigBuilder struct {
	common.EmbeddableBuilder[ptpv1.PtpConfig, *ptpv1.PtpConfig]
	common.EmbeddableCreator[ptpv1.PtpConfig, PtpConfigBuilder, *ptpv1.PtpConfig, *PtpConfigBuilder]
	common.EmbeddableDeleter[ptpv1.PtpConfig, *ptpv1.PtpConfig]
	common.EmbeddableUpdater[ptpv1.PtpConfig, PtpConfigBuilder, *ptpv1.PtpConfig, *PtpConfigBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *PtpConfigBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the PtpConfig GVK for this builder.
func (builder *PtpConfigBuilder) GetGVK() schema.GroupVersionKind {
	return ptpConfigGVK
}

// NewPtpConfigBuilder creates a new instance of a PtpConfig builder.
func NewPtpConfigBuilder(apiClient *clients.Settings, name, nsname string) *PtpConfigBuilder {
	klog.V(100).Infof("Initializing new PtpConfig structure with the following params: name: %s, nsname: %s", name, nsname)

	return common.NewNamespacedBuilder[ptpv1.PtpConfig, PtpConfigBuilder](
		apiClient, ptpv1.AddToScheme, name, nsname)
}

// intelPluginTypes is the list of Intel plugin types to check for in order.
var intelPluginTypes = []PluginType{PluginTypeE810, PluginTypeE825, PluginTypeE830}

// GetIntelPlugin retrieves an Intel plugin (E810, E825, or E830) from the specified profile in the PtpConfig,
// attempting to unmarshal the raw JSON. The plugin's Type field is set based on which plugin key was found. If the
// profile is not found or no Intel plugin exists, it returns an error.
func (builder *PtpConfigBuilder) GetIntelPlugin(profileName string) (*IntelPlugin, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Unmarshalling Intel plugin from PtpConfig %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if profileName == "" {
		klog.V(100).Info("The profileName is empty")

		return nil, fmt.Errorf("profileName cannot be empty")
	}

	for _, profile := range builder.Definition.Spec.Profile {
		if profile.Name == nil || *profile.Name != profileName {
			continue
		}

		// Check for each Intel plugin type in order
		for _, pluginType := range intelPluginTypes {
			if profile.Plugins == nil {
				continue
			}

			pluginJSON, ok := profile.Plugins[string(pluginType)]
			if !ok || pluginJSON == nil {
				continue
			}

			intelPlugin := &IntelPlugin{}

			err := json.Unmarshal(pluginJSON.Raw, intelPlugin)
			if err != nil {
				klog.V(100).Infof("Failed to unmarshal %s plugin: %v", pluginType, err)

				return nil, err
			}

			intelPlugin.Type = pluginType

			return intelPlugin, nil
		}

		klog.V(100).Infof("No Intel plugin found for profile %s", profileName)

		return nil, fmt.Errorf("ptpProfile %s does not have an Intel plugin", profileName)
	}

	klog.V(100).Infof("Profile %s not found in PtpConfig %s in namespace %s",
		profileName, builder.Definition.Name, builder.Definition.Namespace)

	return nil, fmt.Errorf("ptpProfile %s not found", profileName)
}

// WithIntelPlugin sets an Intel plugin in the specified profile of the PtpConfig, attempting to marshal the plugin
// struct into JSON. The plugin's Type field determines which key (e810, e825, or e830) is used in the Plugins map. If
// the plugin's Type is not set, an error is returned.
func (builder *PtpConfigBuilder) WithIntelPlugin(profileName string, plugin *IntelPlugin) *PtpConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting Intel plugin for PtpConfig %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if profileName == "" {
		klog.V(100).Info("The profileName is empty")

		builder.SetError(fmt.Errorf("cannot set Intel plugin: profileName cannot be empty"))

		return builder
	}

	if plugin == nil {
		klog.V(100).Info("Intel plugin is nil")

		builder.SetError(fmt.Errorf("cannot set Intel plugin: plugin is nil"))

		return builder
	}

	if plugin.Type == "" {
		klog.V(100).Info("Intel plugin Type is not set")

		builder.SetError(fmt.Errorf("cannot set Intel plugin: plugin Type is not set"))

		return builder
	}

	if !slices.Contains(intelPluginTypes, plugin.Type) {
		klog.V(100).Infof("Intel plugin type %s is not supported", plugin.Type)

		builder.SetError(fmt.Errorf("cannot set Intel plugin: plugin type %s is not supported", plugin.Type))

		return builder
	}

	for profileIndex, profile := range builder.Definition.Spec.Profile {
		if profile.Name == nil || *profile.Name != profileName {
			continue
		}

		pluginRaw, err := json.Marshal(plugin)
		if err != nil {
			klog.V(100).Infof("Failed to marshal %s plugin: %v", plugin.Type, err)

			builder.SetError(fmt.Errorf("cannot set Intel plugin: failed to marshal plugin struct: %w", err))

			return builder
		}

		if profile.Plugins == nil {
			profile.Plugins = make(map[string]*apiextensionsv1.JSON)
		}

		// Ensure only one Intel plugin key exists per profile.
		for _, pluginType := range intelPluginTypes {
			delete(profile.Plugins, string(pluginType))
		}

		profile.Plugins[string(plugin.Type)] = &apiextensionsv1.JSON{Raw: pluginRaw}
		builder.Definition.Spec.Profile[profileIndex] = profile

		return builder
	}

	builder.SetError(fmt.Errorf("cannot set Intel plugin: ptpProfile %s does not exist", profileName))

	return builder
}

// GetPluginType returns the Intel plugin type (e810, e825, or e830) for the specified profile, if one exists. This is a
// lightweight check that does not unmarshal the plugin data.
func (builder *PtpConfigBuilder) GetPluginType(profileName string) (PluginType, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Infof("Getting Intel plugin type from PtpConfig %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if profileName == "" {
		klog.V(100).Info("The profileName is empty")

		return "", fmt.Errorf("profileName cannot be empty")
	}

	for _, profile := range builder.Definition.Spec.Profile {
		if profile.Name == nil || *profile.Name != profileName {
			continue
		}

		for _, pluginType := range intelPluginTypes {
			if profile.Plugins == nil {
				continue
			}

			pluginJSON, ok := profile.Plugins[string(pluginType)]
			if !ok || pluginJSON == nil {
				continue
			}

			return pluginType, nil
		}

		klog.V(100).Infof("No Intel plugin found for profile %s", profileName)

		return "", fmt.Errorf("ptpProfile %s does not have an Intel plugin", profileName)
	}

	klog.V(100).Infof("Profile %s not found in PtpConfig %s in namespace %s",
		profileName, builder.Definition.Name, builder.Definition.Namespace)

	return "", fmt.Errorf("ptpProfile %s not found", profileName)
}

// PullPtpConfig pulls an existing PtpConfig into a Builder struct.
func PullPtpConfig(apiClient *clients.Settings, name, nsname string) (*PtpConfigBuilder, error) {
	return common.PullNamespacedBuilder[ptpv1.PtpConfig, PtpConfigBuilder](
		context.TODO(), apiClient, ptpv1.AddToScheme, name, nsname)
}
