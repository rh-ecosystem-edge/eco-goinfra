package lca

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	lcaipcv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ipchange/api/ipconfig/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	ipConfigName                 = "ipconfig"
	conditionTypeConfigCompleted = "ConfigCompleted"
)

var ipConfigGVK = lcaipcv1.GroupVersion.WithKind("IPConfig")

// IPConfigBuilder provides struct for the ipconfig object containing connection to
// the cluster and the ipconfig definitions.
type IPConfigBuilder struct {
	common.EmbeddableBuilder[lcaipcv1.IPConfig, *lcaipcv1.IPConfig]
	common.EmbeddableDeleteReturner[lcaipcv1.IPConfig, IPConfigBuilder, *lcaipcv1.IPConfig, *IPConfigBuilder]
	common.EmbeddableUpdater[lcaipcv1.IPConfig, IPConfigBuilder, *lcaipcv1.IPConfig, *IPConfigBuilder]
	common.EmbeddableWithOptions[
		lcaipcv1.IPConfig,
		IPConfigBuilder,
		*lcaipcv1.IPConfig,
		*IPConfigBuilder,
		IPConfigAdditionalOptions,
	]
}

// IPConfigAdditionalOptions additional options for ipconfig object.
type IPConfigAdditionalOptions func(builder *IPConfigBuilder) (*IPConfigBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *IPConfigBuilder) AttachMixins() {
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the IPConfig GVK for this builder.
func (builder *IPConfigBuilder) GetGVK() schema.GroupVersionKind {
	return ipConfigGVK
}

// PullIPConfig pulls existing ipconfig from cluster.
func PullIPConfig(apiClient *clients.Settings) (*IPConfigBuilder, error) {
	return common.PullClusterScopedBuilder[lcaipcv1.IPConfig, IPConfigBuilder](
		context.TODO(), apiClient, lcaipcv1.AddToScheme, ipConfigName)
}

// WaitUntilComplete waits the specified timeout for the ipconfig to complete
// actions.
func (builder *IPConfigBuilder) WaitUntilComplete(timeout time.Duration) (*IPConfigBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Waiting for ipconfig %s to complete actions",
		builder.Definition.Name)

	if !builder.Exists() {
		klog.V(100).Info("The ipconfig does not exist on the cluster")

		return builder, fmt.Errorf("ipconfig %s does not exist", ipConfigName)
	}

	// Polls periodically to determine if ipconfig is in desired state.
	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second*3, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Type == conditionTypeConfigCompleted && condition.Status == isTrue &&
					condition.Reason == isComplete {
					return true, nil
				}
			}

			return false, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}

// WaitUntilFailed waits the specified timeout for the ipconfig to fail.
func (builder *IPConfigBuilder) WaitUntilFailed(timeout time.Duration) (*IPConfigBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Waiting for ipconfig %s to fail",
		builder.Definition.Name)

	if !builder.Exists() {
		klog.V(100).Info("The ipconfig does not exist on the cluster")

		return builder, fmt.Errorf("ipconfig %s does not exist", ipConfigName)
	}

	// Polls periodically to determine if ipconfig is in desired state.
	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second*3, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Type == conditionTypeConfigCompleted && condition.Status == isFalse &&
					condition.Reason == "Failed" {
					return true, nil
				}
			}

			return false, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}

// WaitUntilIdle waits the specified timeout for the ipconfig to become Idle.
func (builder *IPConfigBuilder) WaitUntilIdle(timeout time.Duration) (*IPConfigBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Waiting for ipconfig %s to become Idle",
		builder.Definition.Name)

	if !builder.Exists() {
		klog.V(100).Info("The ipconfig does not exist on the cluster")

		return builder, fmt.Errorf("ipconfig %s does not exist", ipConfigName)
	}

	// Polls periodically to determine if ipconfig is in desired state.
	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second*3, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Type == idle && condition.Status == isTrue &&
					condition.Reason == idle {
					return true, nil
				}
			}

			return false, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}

// WithIPv4Address sets the IPv4 address used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv4Address(
	ipv4Address string) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	ipAddress := net.ParseIP(ipv4Address)

	if ipAddress.To4() == nil {
		klog.V(100).Infof("Invalid IPv4 address %s", ipv4Address)

		builder.SetError(fmt.Errorf("invalid IPv4 address argument %s", ipv4Address))

		return builder
	}

	klog.V(100).Infof("Setting IPv4 %s in ipconfig", ipv4Address)

	if builder.Definition.Spec.IPv4 == nil {
		builder.Definition.Spec.IPv4 = &lcaipcv1.IPv4Config{}
	}

	builder.Definition.Spec.IPv4.Address = ipv4Address

	return builder
}

// WithIPv6Address sets the IPv6 address used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv6Address(
	ipv6Address string) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	ipAddress := net.ParseIP(ipv6Address)

	if ipAddress == nil || ipAddress.To4() != nil {
		klog.V(100).Infof("Invalid IPv6 address %s", ipv6Address)

		builder.SetError(fmt.Errorf("invalid IPv6 argument %s", ipv6Address))

		return builder
	}

	klog.V(100).Infof("Setting IPv6 %s in ipconfig", ipv6Address)

	if builder.Definition.Spec.IPv6 == nil {
		builder.Definition.Spec.IPv6 = &lcaipcv1.IPv6Config{}
	}

	builder.Definition.Spec.IPv6.Address = ipv6Address

	return builder
}

// WithIPv6Gateway sets the IPv6 gateway address used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv6Gateway(
	ipv6Address string) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	gwIPAddress := net.ParseIP(ipv6Address)

	if gwIPAddress == nil || gwIPAddress.To4() != nil {
		klog.V(100).Infof("Invalid IPv6 address %s", ipv6Address)

		builder.SetError(fmt.Errorf("invalid IPv6 argument %s", ipv6Address))

		return builder
	}

	klog.V(100).Infof("Setting IPv6 gateway %s in ipconfig", ipv6Address)

	if builder.Definition.Spec.IPv6 == nil {
		builder.Definition.Spec.IPv6 = &lcaipcv1.IPv6Config{}
	}

	builder.Definition.Spec.IPv6.Gateway = ipv6Address

	return builder
}

// WithIPv4Gateway sets the IPv4 gateway address used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv4Gateway(
	ipv4Address string) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	gwIPAddress := net.ParseIP(ipv4Address)

	if gwIPAddress.To4() == nil {
		klog.V(100).Infof("Invalid IPv4 address %s", ipv4Address)

		builder.SetError(fmt.Errorf("invalid IPv4 address argument %s", ipv4Address))

		return builder
	}

	klog.V(100).Infof("Setting IPv4 gateway %s in ipconfig", ipv4Address)

	if builder.Definition.Spec.IPv4 == nil {
		builder.Definition.Spec.IPv4 = &lcaipcv1.IPv4Config{}
	}

	builder.Definition.Spec.IPv4.Gateway = ipv4Address

	return builder
}

// WithIPv4MachineNetwork sets the IPv4 machine network used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv4MachineNetwork(
	ipv4MachineNetwork string) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	ipv4, _, err := net.ParseCIDR(ipv4MachineNetwork)
	if err != nil || ipv4.To4() == nil {
		klog.V(100).Infof("Invalid CIDR %s", ipv4MachineNetwork)

		builder.SetError(fmt.Errorf("invalid CIDR argument %s", ipv4MachineNetwork))

		return builder
	}

	klog.V(100).Infof("Setting IPv4 machine network %s in ipconfig", ipv4MachineNetwork)

	if builder.Definition.Spec.IPv4 == nil {
		builder.Definition.Spec.IPv4 = &lcaipcv1.IPv4Config{}
	}

	builder.Definition.Spec.IPv4.MachineNetwork = ipv4MachineNetwork

	return builder
}

// WithIPv6MachineNetwork sets the IPv6 machine network used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv6MachineNetwork(
	ipv6MachineNetwork string) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	ipv6, _, err := net.ParseCIDR(ipv6MachineNetwork)
	if err != nil || ipv6.To4() != nil {
		klog.V(100).Infof("Invalid CIDR %s", ipv6MachineNetwork)

		builder.SetError(fmt.Errorf("invalid CIDR argument %s", ipv6MachineNetwork))

		return builder
	}

	klog.V(100).Infof("Setting IPv6 machine network %s in ipconfig", ipv6MachineNetwork)

	if builder.Definition.Spec.IPv6 == nil {
		builder.Definition.Spec.IPv6 = &lcaipcv1.IPv6Config{}
	}

	builder.Definition.Spec.IPv6.MachineNetwork = ipv6MachineNetwork

	return builder
}

// WithStage sets the stage used by the ipconfig.
func (builder *IPConfigBuilder) WithStage(
	stage string) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting stage %s in ipconfig", stage)
	builder.Definition.Spec.Stage = lcaipcv1.IPConfigStage(stage)

	return builder
}

// WithVlanID sets the VLAN ID used by the ipconfig.
func (builder *IPConfigBuilder) WithVlanID(
	vlanID int) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting VLAN ID %d in ipconfig", vlanID)
	builder.Definition.Spec.VLANID = vlanID

	return builder
}

// WithAutoRollbackOnFailure sets the AutoRollbackOnFailure used by the ipconfig.
func (builder *IPConfigBuilder) WithAutoRollbackOnFailure(
	seconds int) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting AutoRollbackOnFailure to %d seconds in ipconfig", seconds)
	builder.Definition.Spec.AutoRollbackOnFailure = &lcaipcv1.AutoRollbackOnFailure{
		InitMonitorTimeoutSeconds: seconds,
	}

	return builder
}

// WithDNS sets the list of DNS servers used by the ipconfig.
func (builder *IPConfigBuilder) WithDNS(
	dnsServers []string) *IPConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting DNS servers %s in ipconfig", dnsServers)

	if len(dnsServers) == 0 {
		klog.V(100).Infof("The DNS servers list is empty")

		builder.SetError(fmt.Errorf("dns servers list cannot be empty"))

		return builder
	}

	for _, dnsServer := range dnsServers {
		dnsServerTrimmed := strings.TrimSpace(dnsServer)

		if dnsServerTrimmed == "" {
			klog.V(100).Infof("DNS server entry is empty")

			builder.SetError(fmt.Errorf("dns server cannot be empty"))

			return builder
		}

		builder.Definition.Spec.DNSServers =
			append(builder.Definition.Spec.DNSServers, lcaipcv1.IPAddress(dnsServerTrimmed))
	}

	return builder
}
