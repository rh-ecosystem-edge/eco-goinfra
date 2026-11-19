package nto //nolint:misspell

import (
	"context"
	"fmt"
	"slices"

	performanceprofilev2 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/performanceprofile/v2"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var performanceProfileGVK = performanceprofilev2.GroupVersion.WithKind("PerformanceProfile")

// Builder provides a struct for PerformanceProfile object from the cluster and a PerformanceProfile definition.
type Builder struct {
	common.EmbeddableBuilder[performanceprofilev2.PerformanceProfile, *performanceprofilev2.PerformanceProfile]
	common.EmbeddableCreator[
		performanceprofilev2.PerformanceProfile, Builder, *performanceprofilev2.PerformanceProfile, *Builder]
	common.EmbeddableDeleteReturner[
		performanceprofilev2.PerformanceProfile, Builder, *performanceprofilev2.PerformanceProfile, *Builder]
	common.EmbeddableForceUpdater[
		performanceprofilev2.PerformanceProfile, Builder, *performanceprofilev2.PerformanceProfile, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the PerformanceProfile GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return performanceProfileGVK
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(
	apiClient *clients.Settings, name, cpuIsolated, cpuReserved string, nodeSelector map[string]string) *Builder {
	klog.V(100).Infof(
		"Initializing new PerformanceProfile structure with the following params: "+
			"name: %s, cpu isolated: %s, cpu reserved %s, nodeSelector %v", name, cpuIsolated, cpuReserved, nodeSelector)

	builder := common.NewClusterScopedBuilder[performanceprofilev2.PerformanceProfile, Builder](
		apiClient, performanceprofilev2.AddToScheme, name)

	isolatedCPUSet := performanceprofilev2.CPUSet(cpuIsolated)
	reservedCPUSet := performanceprofilev2.CPUSet(cpuReserved)
	builder.Definition.Spec.CPU = &performanceprofilev2.CPU{
		Isolated: &isolatedCPUSet,
		Reserved: &reservedCPUSet,
	}
	builder.Definition.Spec.NodeSelector = nodeSelector

	if builder.GetError() != nil {
		return builder
	}

	if cpuIsolated == "" {
		klog.V(100).Info("Isolated CPU of the PerformanceProfile is empty")

		builder.SetError(fmt.Errorf("PerformanceProfile's 'cpuIsolated' is empty"))

		return builder
	}

	if cpuReserved == "" {
		klog.V(100).Info("Reserved CPU of the PerformanceProfile is empty")

		builder.SetError(fmt.Errorf("PerformanceProfile's 'cpuReserved' is empty"))

		return builder
	}

	if len(nodeSelector) == 0 {
		klog.V(100).Info("NodeSelector of the PerformanceProfile is empty")

		builder.SetError(fmt.Errorf("PerformanceProfile's 'nodeSelector' is empty"))

		return builder
	}

	return builder
}

// Pull pulls existing PerformanceProfile from cluster.
func Pull(apiClient *clients.Settings, name string) (*Builder, error) {
	return common.PullClusterScopedBuilder[performanceprofilev2.PerformanceProfile, Builder](
		context.TODO(), apiClient, performanceprofilev2.AddToScheme, name)
}

// WithHugePages defines the HugePages in the PerformanceProfile. hugePageSize allowed values are 2M, 1G.
func (builder *Builder) WithHugePages(hugePageSize string, hugePages []performanceprofilev2.HugePage) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding hugePages to PerformanceProfile %s, size %s, hugePages %v",
		builder.Definition.Name, hugePageSize, hugePages)

	if hugePageSize == "" {
		klog.V(100).Info("The hugePageSize is empty")

		builder.SetError(fmt.Errorf("'hugePageSize' argument cannot be empty"))

		return builder
	}

	allowedHugePageSize := []string{"2M", "1G"}
	if !slices.Contains(allowedHugePageSize, hugePageSize) {
		klog.V(100).Infof("'hugePageSize' has invalid parameter %s. Allowed parameters %v",
			hugePageSize, allowedHugePageSize)

		builder.SetError(fmt.Errorf("'hugePageSize' argument is not in allowed list: %v", allowedHugePageSize))

		return builder
	}

	if len(hugePages) == 0 {
		klog.V(100).Info("'hugePages' argument cannot be empty")

		builder.SetError(fmt.Errorf("'hugePages' argument cannot be empty"))

		return builder
	}

	pageSize := performanceprofilev2.HugePageSize(hugePageSize)

	if builder.Definition.Spec.HugePages != nil {
		builder.Definition.Spec.HugePages.DefaultHugePagesSize = &pageSize
		builder.Definition.Spec.HugePages.Pages = hugePages

		return builder
	}

	builder.Definition.Spec.HugePages = &performanceprofilev2.HugePages{
		DefaultHugePagesSize: &pageSize,
		Pages:                hugePages,
	}

	return builder
}

// WithMachineConfigPoolSelector defines the MachineConfigPoolSelector in the PerformanceProfile.
func (builder *Builder) WithMachineConfigPoolSelector(machineConfigPoolSelector map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding MachineConfigPoolSelector %v to PerformanceProfile %s",
		machineConfigPoolSelector, builder.Definition.Name)

	if len(machineConfigPoolSelector) == 0 {
		klog.V(100).Info("'machineConfigPoolSelector' argument cannot be empty")

		builder.SetError(fmt.Errorf("'machineConfigPoolSelector' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.MachineConfigPoolSelector = machineConfigPoolSelector

	return builder
}

// WithNodeSelector defines the nodeSelector in the PerformanceProfile.
func (builder *Builder) WithNodeSelector(nodeSelector map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding nodeSelector %v to PerformanceProfile %s",
		nodeSelector, builder.Definition.Name)

	if len(nodeSelector) == 0 {
		klog.V(100).Info("'nodeSelector' argument cannot be empty")

		builder.SetError(fmt.Errorf("'nodeSelector' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.NodeSelector = nodeSelector

	return builder
}

// WithNumaTopology defines the NumaTopologyPolicy in the PerformanceProfile.
func (builder *Builder) WithNumaTopology(topologyPolicy string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding NumaTopologyPolicy %s to PerformanceProfile %s",
		topologyPolicy, builder.Definition.Name)

	if topologyPolicy == "" {
		klog.V(100).Info("The topologyPolicy is empty")

		builder.SetError(fmt.Errorf("'topologyPolicy' argument cannot be empty"))

		return builder
	}

	allowedTopologyPolicies := []string{"best-effort", "restricted", "single-numa-node"}
	if !slices.Contains(allowedTopologyPolicies, topologyPolicy) {
		klog.V(100).Infof("'allowedTopologyPolicies' has invalid parameter %s. Allowed parameters %v",
			topologyPolicy, allowedTopologyPolicies)

		builder.SetError(fmt.Errorf("'allowedTopologyPolicies' argument is not in allowed list %v",
			allowedTopologyPolicies))

		return builder
	}

	builder.Definition.Spec.NUMA = &performanceprofilev2.NUMA{TopologyPolicy: &topologyPolicy}

	return builder
}

// WithRTKernel defines the Real Time Kernel in the PerformanceProfile.
func (builder *Builder) WithRTKernel() *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding RTKernel flag to PerformanceProfile %s", builder.Definition.Name)

	trueFlag := true
	builder.Definition.Spec.RealTimeKernel = &performanceprofilev2.RealTimeKernel{Enabled: &trueFlag}

	return builder
}

// WithGloballyDisableIrqLoadBalancing defines the globallyDisableIrqLoadBalancing in the PerformanceProfile.
func (builder *Builder) WithGloballyDisableIrqLoadBalancing() *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding globallyDisableIrqLoadBalancing flag to PerformanceProfile %s",
		builder.Definition.Name)

	trueFlag := true
	builder.Definition.Spec.GloballyDisableIrqLoadBalancing = &trueFlag

	return builder
}

// WithWorkloadHints defines the Workload Hints in the PerformanceProfile.
func (builder *Builder) WithWorkloadHints(rtHint, perPodPowerMgmtHint, highPowerHint bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding WorkloadHints flags: RealTime=%t, PerPodPowerManagement=%t, HighPowerConsumption=%t to PerformanceProfile %s",
		rtHint, perPodPowerMgmtHint, highPowerHint, builder.Definition.Name)

	if builder.Definition.Spec.WorkloadHints == nil {
		builder.Definition.Spec.WorkloadHints = &performanceprofilev2.WorkloadHints{
			RealTime:              &rtHint,
			PerPodPowerManagement: &perPodPowerMgmtHint,
			HighPowerConsumption:  &highPowerHint,
		}

		return builder
	}

	builder.Definition.Spec.WorkloadHints.RealTime = &rtHint
	builder.Definition.Spec.WorkloadHints.PerPodPowerManagement = &perPodPowerMgmtHint
	builder.Definition.Spec.WorkloadHints.HighPowerConsumption = &highPowerHint

	return builder
}

// WithAnnotations defines the annotations in the PerformanceProfile.
func (builder *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding annotations %v to the PerformanceProfile %s",
		annotations, builder.Definition.Name)

	if len(annotations) == 0 {
		klog.V(100).Info("'annotations' argument cannot be empty")

		builder.SetError(fmt.Errorf("'annotations' argument cannot be empty"))

		return builder
	}

	builder.Definition.Annotations = annotations

	return builder
}

// WithNet defines the net in the PerformanceProfile.
func (builder *Builder) WithNet(userLevelNetworking bool, devices []performanceprofilev2.Device) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding net field to the PerformanceProfile %s", builder.Definition.Name)

	if len(devices) == 0 {
		klog.V(100).Info("'net' argument cannot be empty")

		builder.SetError(fmt.Errorf("'net' argument cannot be empty"))

		return builder
	}

	netField := performanceprofilev2.Net{
		UserLevelNetworking: &userLevelNetworking,
		Devices:             devices,
	}

	builder.Definition.Spec.Net = &netField

	return builder
}

// WithAdditionalKernelArgs defines the additionalKernelArgs in the PerformanceProfile.
func (builder *Builder) WithAdditionalKernelArgs(additionalKernelArgs []string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding additionalKernelArgs field to the PerformanceProfile %s",
		builder.Definition.Name)

	if len(additionalKernelArgs) == 0 {
		klog.V(100).Info("'additionalKernelArgs' argument cannot be empty")

		builder.SetError(fmt.Errorf("'additionalKernelArgs' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.AdditionalKernelArgs = additionalKernelArgs

	return builder
}
