package nad

import (
	"context"
	"encoding/json"
	"fmt"

	nadV1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// Builder provides struct for NAD object which contains connection to cluster and the NAD object itself.
type Builder struct {
	common.EmbeddableBuilder[nadV1.NetworkAttachmentDefinition, *nadV1.NetworkAttachmentDefinition]
	common.EmbeddableCreator[nadV1.NetworkAttachmentDefinition, Builder, *nadV1.NetworkAttachmentDefinition, *Builder]
	common.EmbeddableDeleter[nadV1.NetworkAttachmentDefinition, *nadV1.NetworkAttachmentDefinition]
	common.EmbeddableUpdater[nadV1.NetworkAttachmentDefinition, Builder, *nadV1.NetworkAttachmentDefinition, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the NetworkAttachmentDefinition GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return nadV1.SchemeGroupVersion.WithKind("NetworkAttachmentDefinition")
}

// NewBuilder creates a new instance of NetworkAttachmentDefinition Builder.
// arguments:       "apiClient" -       the nad network client.
//
//	"name"      -       the name of the nad network.
//	"nsname"    -       the nad network namespace.
//
// return value:    the created Builder.
func NewBuilder(apiClient *clients.Settings, name, nsname string) *Builder {
	return common.NewNamespacedBuilder[nadV1.NetworkAttachmentDefinition, Builder](
		apiClient, nadV1.AddToScheme, name, nsname)
}

// Pull pulls existing networkattachmentdefinition from cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	return common.PullNamespacedBuilder[nadV1.NetworkAttachmentDefinition, Builder](
		context.TODO(), apiClient, nadV1.AddToScheme, name, nsname)
}

// GetString prints NetworkAttachmentDefinition resource.
// return value:    the builder details in json string format, and an error if any occurred.
func (builder *Builder) GetString() (string, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Info("Returning NetworkAttachmentDefinition resource in json format")

	nadByte, err := json.MarshalIndent(builder.Definition, "", "    ")
	if err != nil {
		return "", err
	}

	return string(nadByte), err
}

// WithMasterPlugin defines master plugin configuration in the NetworkAttachmentDefinition spec.
func (builder *Builder) WithMasterPlugin(masterPlugin *MasterPlugin) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if masterPlugin == nil {
		builder.SetError(fmt.Errorf("error 'masterPlugin' is empty"))
	}

	klog.V(100).Infof("Adding masterPlugin %v to NAD %s", masterPlugin, builder.Definition.Name)

	emptyNadConfig := nadV1.NetworkAttachmentDefinitionSpec{}

	if builder.Definition.Spec != emptyNadConfig {
		builder.SetError(fmt.Errorf("error to redefine predefine NAD"))

		return builder
	}

	masterPluginString, err := json.Marshal(masterPlugin)
	if err != nil {
		builder.SetError(err)

		return builder
	}

	builder.Definition.Spec.Config = string(masterPluginString)

	return builder
}

// WithPlugins defines nad with group of plugins.
func (builder *Builder) WithPlugins(name string, plugins *[]Plugin) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding plugins to NAD %s", builder.Definition.Name)

	pluginsConfig := MasterPlugin{
		CniVersion: "0.4.0",
		Name:       name,
		Plugins:    plugins,
	}

	pluginsConfigString, err := json.Marshal(pluginsConfig)
	if err != nil {
		builder.SetError(err)

		return builder
	}

	builder.Definition.Spec.Config = string(pluginsConfigString)

	return builder
}

// GetGVR returns nad's GroupVersionResource which could be used for Clean function.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: "k8s.cni.cncf.io", Version: "v1", Resource: "network-attachment-definitions",
	}
}
