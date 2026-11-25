package mco

import (
	"context"
	"fmt"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

const errEmptyKey = "'key' cannot be empty"

var kubeletConfigGVK = mcv1.GroupVersion.WithKind("KubeletConfig")

// KubeletConfigBuilder provides struct for KubeletConfig Object which contains connection to cluster
// and KubeletConfig definitions.
type KubeletConfigBuilder struct {
	common.EmbeddableBuilder[mcv1.KubeletConfig, *mcv1.KubeletConfig]
	common.EmbeddableWithOptions[
		mcv1.KubeletConfig, KubeletConfigBuilder, *mcv1.KubeletConfig, *KubeletConfigBuilder, AdditionalOptions]
	common.EmbeddableCreator[mcv1.KubeletConfig, KubeletConfigBuilder, *mcv1.KubeletConfig, *KubeletConfigBuilder]
	common.EmbeddableDeleter[mcv1.KubeletConfig, *mcv1.KubeletConfig]
}

// AdditionalOptions for kubeletconfig object.
type AdditionalOptions func(builder *KubeletConfigBuilder) (*KubeletConfigBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *KubeletConfigBuilder) AttachMixins() {
	builder.EmbeddableWithOptions.SetBase(builder)
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the KubeletConfig GVK for this builder.
func (builder *KubeletConfigBuilder) GetGVK() schema.GroupVersionKind {
	return kubeletConfigGVK
}

// NewKubeletConfigBuilder provides struct for KubeletConfig object which contains connection to cluster
// and KubeletConfig definition.
func NewKubeletConfigBuilder(apiClient *clients.Settings, name string) *KubeletConfigBuilder {
	klog.V(100).Infof("Initializing new KubeletConfigBuilder structure with the name: %s", name)

	return common.NewClusterScopedBuilder[mcv1.KubeletConfig, KubeletConfigBuilder](apiClient, mcv1.Install, name)
}

// PullKubeletConfig fetches existing kubeletconfig from cluster.
func PullKubeletConfig(apiClient *clients.Settings, name string) (*KubeletConfigBuilder, error) {
	klog.V(100).Infof("Pulling existing kubeletconfig name %s from cluster", name)

	return common.PullClusterScopedBuilder[mcv1.KubeletConfig, KubeletConfigBuilder](
		context.TODO(), apiClient, mcv1.Install, name)
}

// WithMCPoolSelector redefines kubeletconfig definition with the given machineConfigPoolSelector field.
func (builder *KubeletConfigBuilder) WithMCPoolSelector(key, value string) *KubeletConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Labeling the kubeletconfig %s with %s=%s", builder.Definition.Name, key, value)

	if key == "" {
		klog.V(100).Info("The key cannot be empty")

		builder.SetError(fmt.Errorf(errEmptyKey))

		return builder
	}

	if builder.Definition.Spec.MachineConfigPoolSelector == nil {
		builder.Definition.Spec.MachineConfigPoolSelector = &metav1.LabelSelector{}
	}

	if builder.Definition.Spec.MachineConfigPoolSelector.MatchLabels == nil {
		builder.Definition.Spec.MachineConfigPoolSelector.MatchLabels = map[string]string{}
	}

	builder.Definition.Spec.MachineConfigPoolSelector.MatchLabels[key] = value

	return builder
}

// WithSystemReserved redefines kubeletconfig definition with the given systemreserved fields.
func (builder *KubeletConfigBuilder) WithSystemReserved(cpu, memory string) *KubeletConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting cpu=%s and memory=%s in the %s kubeletconfig definition",
		cpu, memory, builder.Definition.Name)

	if cpu == "" {
		klog.V(100).Info("The cpu cannot be empty")

		builder.SetError(fmt.Errorf("'cpu' cannot be empty"))

		return builder
	}

	if memory == "" {
		klog.V(100).Info("The memory cannot be empty")

		builder.SetError(fmt.Errorf("'memory' cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.KubeletConfig == nil {
		builder.Definition.Spec.KubeletConfig = &runtime.RawExtension{}
	}

	systemReservedKubeletConfiguration := &kubeletconfigv1beta1.KubeletConfiguration{
		SystemReserved: map[string]string{
			"cpu":    cpu,
			"memory": memory,
		},
	}

	builder.Definition.Spec.KubeletConfig.Object = systemReservedKubeletConfiguration

	return builder
}
