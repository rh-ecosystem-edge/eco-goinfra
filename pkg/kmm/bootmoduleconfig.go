package kmm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	bmcV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// BootModuleConfigBuilder provides struct for the bootmoduleconfig object containing connection to
// the cluster and the bootmoduleconfig definitions.
type BootModuleConfigBuilder struct {
	common.EmbeddableBuilder[bmcV1Beta1.BootModuleConfig, *bmcV1Beta1.BootModuleConfig]
	common.EmbeddableCreator[bmcV1Beta1.BootModuleConfig, BootModuleConfigBuilder,
		*bmcV1Beta1.BootModuleConfig, *BootModuleConfigBuilder]
	common.EmbeddableDeleteReturner[bmcV1Beta1.BootModuleConfig, BootModuleConfigBuilder,
		*bmcV1Beta1.BootModuleConfig, *BootModuleConfigBuilder]
	common.EmbeddableUpdater[bmcV1Beta1.BootModuleConfig, BootModuleConfigBuilder,
		*bmcV1Beta1.BootModuleConfig, *BootModuleConfigBuilder]
	common.EmbeddableWithOptions[bmcV1Beta1.BootModuleConfig, BootModuleConfigBuilder,
		*bmcV1Beta1.BootModuleConfig, *BootModuleConfigBuilder, BootModuleConfigAdditionalOptions]
}

// BootModuleConfigAdditionalOptions additional options for bootmoduleconfig object.
type BootModuleConfigAdditionalOptions func(builder *BootModuleConfigBuilder) (*BootModuleConfigBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *BootModuleConfigBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the BootModuleConfig GVK for this builder.
func (builder *BootModuleConfigBuilder) GetGVK() schema.GroupVersionKind {
	return bmcV1Beta1.GroupVersion.WithKind("BootModuleConfig")
}

// NewBootModuleConfigBuilder creates a new instance of BootModuleConfigBuilder.
func NewBootModuleConfigBuilder(
	apiClient *clients.Settings, name, nsname string) *BootModuleConfigBuilder {
	return common.NewNamespacedBuilder[bmcV1Beta1.BootModuleConfig, BootModuleConfigBuilder](
		apiClient, bmcV1Beta1.AddToScheme, name, nsname)
}

// WithMachineConfigName sets the MachineConfig name that is targeted by the BMC.
func (builder *BootModuleConfigBuilder) WithMachineConfigName(mcName string) *BootModuleConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if mcName == "" {
		builder.SetError(fmt.Errorf("bootmoduleconfig 'machineConfigName' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting BootModuleConfig MachineConfigName to: %s", mcName)

	builder.Definition.Spec.MachineConfigName = mcName

	return builder
}

// WithMachineConfigPoolName sets the MachineConfigPool name linked to the targeted MachineConfig.
func (builder *BootModuleConfigBuilder) WithMachineConfigPoolName(mcpName string) *BootModuleConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if mcpName == "" {
		builder.SetError(fmt.Errorf("bootmoduleconfig 'machineConfigPoolName' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting BootModuleConfig MachineConfigPoolName to: %s", mcpName)

	builder.Definition.Spec.MachineConfigPoolName = mcpName

	return builder
}

// WithKernelModuleImage sets the container image that contains the kernel module .ko file.
func (builder *BootModuleConfigBuilder) WithKernelModuleImage(image string) *BootModuleConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if image == "" {
		builder.SetError(fmt.Errorf("bootmoduleconfig 'kernelModuleImage' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting BootModuleConfig KernelModuleImage to: %s", image)

	builder.Definition.Spec.KernelModuleImage = image

	return builder
}

// WithKernelModuleName sets the name of the kernel module to be loaded.
func (builder *BootModuleConfigBuilder) WithKernelModuleName(moduleName string) *BootModuleConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if moduleName == "" {
		builder.SetError(fmt.Errorf("bootmoduleconfig 'kernelModuleName' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting BootModuleConfig KernelModuleName to: %s", moduleName)

	builder.Definition.Spec.KernelModuleName = moduleName

	return builder
}

// WithInTreeModulesToRemove sets the in-tree kernel module list to remove prior to loading the OOT kernel module.
func (builder *BootModuleConfigBuilder) WithInTreeModulesToRemove(modules []string) *BootModuleConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting BootModuleConfig InTreeModulesToRemove to: %v", modules)

	builder.Definition.Spec.InTreeModulesToRemove = modules

	return builder
}

// WithFirmwareFilesPath sets the path of the firmware files in the kernel module container image.
func (builder *BootModuleConfigBuilder) WithFirmwareFilesPath(path string) *BootModuleConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting BootModuleConfig FirmwareFilesPath to: %s", path)

	builder.Definition.Spec.FirmwareFilesPath = path

	return builder
}

// WithWorkerImage sets the KMM worker image.
func (builder *BootModuleConfigBuilder) WithWorkerImage(image string) *BootModuleConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting BootModuleConfig WorkerImage to: %s", image)

	builder.Definition.Spec.WorkerImage = image

	return builder
}

// PullBootModuleConfig pulls existing bootmoduleconfig from cluster.
func PullBootModuleConfig(apiClient *clients.Settings, name, nsname string) (*BootModuleConfigBuilder, error) {
	return common.PullNamespacedBuilder[bmcV1Beta1.BootModuleConfig, BootModuleConfigBuilder](
		context.TODO(), apiClient, bmcV1Beta1.AddToScheme, name, nsname)
}
