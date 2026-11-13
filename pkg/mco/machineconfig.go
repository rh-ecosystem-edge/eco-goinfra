package mco

import (
	"context"
	"fmt"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var machineConfigGVK = mcv1.GroupVersion.WithKind("MachineConfig")

// MCBuilder provides struct for MachineConfig Object which contains connection to cluster
// and MachineConfig definitions.
type MCBuilder struct {
	common.EmbeddableBuilder[mcv1.MachineConfig, *mcv1.MachineConfig]
	common.EmbeddableCreator[mcv1.MachineConfig, MCBuilder, *mcv1.MachineConfig, *MCBuilder]
	common.EmbeddableDeleter[mcv1.MachineConfig, *mcv1.MachineConfig]
	common.EmbeddableUpdater[mcv1.MachineConfig, MCBuilder, *mcv1.MachineConfig, *MCBuilder]
	common.EmbeddableWithOptions[mcv1.MachineConfig, MCBuilder, *mcv1.MachineConfig, *MCBuilder, MCAdditionalOptions]
}

// MCAdditionalOptions for machineconfig object.
type MCAdditionalOptions func(builder *MCBuilder) (*MCBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *MCBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the MachineConfig GVK for this builder.
func (builder *MCBuilder) GetGVK() schema.GroupVersionKind {
	return machineConfigGVK
}

// NewMCBuilder provides struct for MachineConfig object which contains connection to cluster
// and MachineConfig definition.
func NewMCBuilder(apiClient *clients.Settings, name string) *MCBuilder {
	klog.V(100).Infof("Initializing new MCBuilder structure with following params: %s", name)

	return common.NewClusterScopedBuilder[mcv1.MachineConfig, MCBuilder](apiClient, mcv1.Install, name)
}

// PullMachineConfig fetches existing machineconfig from cluster.
func PullMachineConfig(apiClient *clients.Settings, name string) (*MCBuilder, error) {
	klog.V(100).Infof("Pulling existing machineconfig name %s from cluster", name)

	return common.PullClusterScopedBuilder[mcv1.MachineConfig, MCBuilder](
		context.TODO(), apiClient, mcv1.Install, name)
}

// WithLabel redefines machineconfig definition with the given label.
func (builder *MCBuilder) WithLabel(key, value string) *MCBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Labeling the machineconfig %s with %s=%s", builder.Definition.Name, key, value)

	if key == "" {
		klog.V(100).Info("The key cannot be empty")

		builder.SetError(fmt.Errorf(errEmptyKey))

		return builder
	}

	if builder.Definition.Labels == nil {
		builder.Definition.Labels = map[string]string{}
	}

	builder.Definition.Labels[key] = value

	return builder
}

// WithKernelArguments sets the specified KernelArguments to the MachineConfig.
func (builder *MCBuilder) WithKernelArguments(kernelArgs []string) *MCBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if len(kernelArgs) == 0 {
		klog.V(100).Info("The kernelArgs cannot be empty")

		builder.SetError(fmt.Errorf("'kernelArgs' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting KernelArguments: %v", kernelArgs)

	builder.Definition.Spec.KernelArguments = kernelArgs

	return builder
}

// WithExtensions sets the specified Extensions to the MachineConfig.
func (builder *MCBuilder) WithExtensions(extensions []string) *MCBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if len(extensions) == 0 {
		klog.V(100).Info("The extensions cannot be empty")

		builder.SetError(fmt.Errorf("'extensions' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting Extensions: %v", extensions)

	builder.Definition.Spec.Extensions = extensions

	return builder
}

// WithFIPS sets the specified FIPS value to the MachineConfig.
func (builder *MCBuilder) WithFIPS(fips bool) *MCBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting FIPS: %v", fips)

	builder.Definition.Spec.FIPS = fips

	return builder
}

// WithKernelType sets the specified kernelType to the MachineConfig.
func (builder *MCBuilder) WithKernelType(kernelType string) *MCBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if kernelType == "" {
		klog.V(100).Info("The kernelType cannot be empty")

		builder.SetError(fmt.Errorf("'kernelType' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting KernelType: %v", kernelType)

	builder.Definition.Spec.KernelType = kernelType

	return builder
}

// WithRawConfig sets the specified Config to the MachineConfig.
func (builder *MCBuilder) WithRawConfig(config []byte) *MCBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if len(config) == 0 {
		klog.V(100).Info("The Config.Raw cannot be empty")

		builder.SetError(fmt.Errorf("'Config.Raw' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting Config.Raw: %s", string(config))

	builder.Definition.Spec.Config.Raw = config

	return builder
}
