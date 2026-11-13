package bmh

import (
	"context"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

var hfsGVK = bmhv1alpha1.GroupVersion.WithKind("HostFirmwareSettings")

// HFSBuilder provides a struct to interface with HostFirmwareSettings resources on a specific cluster.
type HFSBuilder struct {
	common.EmbeddableBuilder[bmhv1alpha1.HostFirmwareSettings, *bmhv1alpha1.HostFirmwareSettings]
	common.EmbeddableCreator[bmhv1alpha1.HostFirmwareSettings, HFSBuilder,
		*bmhv1alpha1.HostFirmwareSettings, *HFSBuilder]
	common.EmbeddableDeleter[bmhv1alpha1.HostFirmwareSettings, *bmhv1alpha1.HostFirmwareSettings]
	common.EmbeddableUpdater[bmhv1alpha1.HostFirmwareSettings, HFSBuilder,
		*bmhv1alpha1.HostFirmwareSettings, *HFSBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *HFSBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the HostFirmwareSettings GVK for this builder.
func (builder *HFSBuilder) GetGVK() schema.GroupVersionKind {
	return hfsGVK
}

// PullHFS pulls an existing HostFirmwareSettings from the cluster.
func PullHFS(apiClient *clients.Settings, name, nsname string) (*HFSBuilder, error) {
	klog.V(100).Infof("Pulling existing HostFirmwareSettings name %s under namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[bmhv1alpha1.HostFirmwareSettings, HFSBuilder](
		context.TODO(), apiClient, bmhv1alpha1.AddToScheme, name, nsname)
}

// WithSettings sets the desired firmware settings on the HostFirmwareSettings definition.
func (builder *HFSBuilder) WithSettings(settings map[string]string) *HFSBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting firmware settings on HostFirmwareSettings %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if builder.Definition.Spec.Settings == nil {
		builder.Definition.Spec.Settings = make(bmhv1alpha1.DesiredSettingsMap)
	}

	for key, value := range settings {
		builder.Definition.Spec.Settings[key] = intstr.FromString(value)
	}

	return builder
}
