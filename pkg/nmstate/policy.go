package nmstate

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"slices"
	"time"

	"gopkg.in/yaml.v2"

	"k8s.io/klog/v2"

	nmstateShared "github.com/nmstate/kubernetes-nmstate/api/shared"
	nmstateV1 "github.com/nmstate/kubernetes-nmstate/api/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	nodeNetConfPolIntError      = "nodenetworkconfigurationpolicy 'interfaceName' cannot be empty"
	nodeNetConfPolSelectorEmpty = "nodeNetworkConfigurationPolicy 'nodeSelector' cannot be empty map"
	nodeNetConfPolAltnamesEmpty = "altnames cannot be empty array"
	bondModeActiveBackup        = "active-backup"
	interfaceTypeEthernet       = "ethernet"
	errEmptyBaseInterface       = "nodenetworkconfigurationpolicy 'baseInterface' cannot be empty"
	errInvalidVLANID            = "invalid vlanID, allowed vlanID values are between 0-4094"
	interfaceStateAbsent        = "absent"
)

var (
	// allowedBondModes represents all allowed modes for Bond interface.
	allowedBondModes = []string{"balance-rr", bondModeActiveBackup, "balance-xor", "broadcast", "802.3ad"}
	pciAddressRegexp = regexp.MustCompile(`^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-7]$`)
	policyGVK        = nmstateV1.GroupVersion.WithKind("NodeNetworkConfigurationPolicy")
)

// AdditionalOptions additional options for NodeNetworkConfigurationPolicy objects.
type AdditionalOptions func(builder *PolicyBuilder) (*PolicyBuilder, error)

// PolicyBuilder provides struct for the NodeNetworkConfigurationPolicy object containing connection to
// the cluster and the NodeNetworkConfigurationPolicy definition.
type PolicyBuilder struct {
	common.EmbeddableBuilder[nmstateV1.NodeNetworkConfigurationPolicy, *nmstateV1.NodeNetworkConfigurationPolicy]
	common.EmbeddableCreator[
		nmstateV1.NodeNetworkConfigurationPolicy,
		PolicyBuilder,
		*nmstateV1.NodeNetworkConfigurationPolicy,
		*PolicyBuilder,
	]
	common.EmbeddableDeleteReturner[
		nmstateV1.NodeNetworkConfigurationPolicy,
		PolicyBuilder,
		*nmstateV1.NodeNetworkConfigurationPolicy,
		*PolicyBuilder,
	]
	common.EmbeddableForceUpdater[
		nmstateV1.NodeNetworkConfigurationPolicy,
		PolicyBuilder,
		*nmstateV1.NodeNetworkConfigurationPolicy,
		*PolicyBuilder,
	]
	common.EmbeddableWithOptions[
		nmstateV1.NodeNetworkConfigurationPolicy,
		PolicyBuilder,
		*nmstateV1.NodeNetworkConfigurationPolicy,
		*PolicyBuilder,
		AdditionalOptions,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *PolicyBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the NodeNetworkConfigurationPolicy GVK for this builder.
func (builder *PolicyBuilder) GetGVK() schema.GroupVersionKind {
	return policyGVK
}

// NewPolicyBuilder creates a new instance of PolicyBuilder.
func NewPolicyBuilder(apiClient *clients.Settings, name string, nodeSelector map[string]string) *PolicyBuilder {
	klog.V(100).Infof(
		"Initializing new NodeNetworkConfigurationPolicy structure with the following params: %s", name)

	builder := common.NewClusterScopedBuilder[nmstateV1.NodeNetworkConfigurationPolicy, PolicyBuilder](
		apiClient, nmstateV1.AddToScheme, name)
	if builder.GetError() != nil {
		return builder
	}

	if len(nodeSelector) == 0 {
		klog.V(100).Info("The nodeSelector of the NodeNetworkConfigurationPolicy is empty")

		builder.SetError(fmt.Errorf(nodeNetConfPolSelectorEmpty))

		return builder
	}

	builder.Definition.Spec.NodeSelector = nodeSelector

	return builder
}

// WithInterfaceAndVFs adds SR-IOV VF configuration to the NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) WithInterfaceAndVFs(sriovInterface string, numberOfVF uint8) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating NodeNetworkConfigurationPolicy %s with SR-IOV VF configuration: %d",
		builder.Definition.Name, numberOfVF)

	if sriovInterface == "" {
		klog.V(100).Info("The sriovInterface  can not be empty string")

		builder.SetError(fmt.Errorf("%s", "The sriovInterface is empty string"))

		return builder
	}

	intNumberOfVF := int(numberOfVF)
	newInterface := NetworkInterface{
		Name:  sriovInterface,
		Type:  interfaceTypeEthernet,
		State: "up",
		Ethernet: Ethernet{
			Sriov: Sriov{TotalVfs: &intNumberOfVF},
		},
	}

	return builder.withInterface(newInterface)
}

// WithBondInterface adds Bond interface configuration to the NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) WithBondInterface(slavePorts []string, bondName, mode string,
	options ...OptionsLinkAggregation) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with Bond interface configuration:"+
		" BondName %s, Mode %s, SlavePorts %v", builder.Definition.Name, bondName, mode, slavePorts)

	if !slices.Contains(allowedBondModes, mode) {
		klog.V(100).Infof("error to add Bond mode %s, allowed modes are %v", mode, allowedBondModes)

		builder.SetError(fmt.Errorf("invalid Bond mode parameter"))

		return builder
	}

	if bondName == "" {
		klog.V(100).Info("The bondName can not be empty string")

		builder.SetError(fmt.Errorf("%s", "The bondName is empty sting"))

		return builder
	}

	newInterface := NetworkInterface{
		Name:  bondName,
		Type:  interfaceTypeBond,
		State: "up",
		LinkAggregation: LinkAggregation{
			Mode: mode,
			Port: slavePorts,
		},
	}

	if len(options) > 0 {
		newInterface.LinkAggregation.Options = options[0]
	}

	return builder.withInterface(newInterface)
}

// WithVlanInterface adds VLAN interface configuration to the NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) WithVlanInterface(baseInterface string, vlanID uint16) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with VLAN interface %s and vlanID %d",
		builder.Definition.Name, baseInterface, vlanID)

	if baseInterface == "" {
		klog.V(100).Info("The baseInterface can not be empty string")

		builder.SetError(fmt.Errorf(errEmptyBaseInterface))

		return builder
	}

	if vlanID > 4094 {
		builder.SetError(fmt.Errorf(errInvalidVLANID))

		return builder
	}

	newInterface := NetworkInterface{
		Name:  fmt.Sprintf("%s.%d", baseInterface, vlanID),
		Type:  interfaceTypeVlan,
		State: "up",
		Vlan: Vlan{
			BaseIface: baseInterface,
			ID:        int(vlanID),
		},
	}

	return builder.withInterface(newInterface)
}

// WithVlanInterfaceIP adds VLAN interface with IP configuration to the NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) WithVlanInterfaceIP(baseInterface, ipv4Addresses, ipv6Addresses string,
	vlanID uint16) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with VLAN interface %s and vlanID %d",
		builder.Definition.Name, baseInterface, vlanID)

	if baseInterface == "" {
		klog.V(100).Info("The baseInterface can not be empty string")

		builder.SetError(fmt.Errorf(errEmptyBaseInterface))

		return builder
	}

	if vlanID > 4094 {
		klog.V(100).Info("the vlanID is out of range, allowed vlanID values are between 0-4094")

		builder.SetError(fmt.Errorf(errInvalidVLANID))

		return builder
	}

	if net.ParseIP(ipv4Addresses) == nil {
		klog.V(100).Info("the vlanInterface contains an invalid ipv4 address")

		builder.SetError(fmt.Errorf("vlanInterfaceIP 'ipv4Addresses' is an invalid ipv4 address"))

		return builder
	}

	if net.ParseIP(ipv6Addresses) == nil {
		klog.V(100).Info("the vlanInterface contains an invalid ipv6 address")

		builder.SetError(fmt.Errorf("vlanInterfaceIP 'ipv6Addresses' is an invalid ipv6 address"))

		return builder
	}

	newInterface := NetworkInterface{
		Name:  fmt.Sprintf("%s.%d", baseInterface, vlanID),
		Type:  interfaceTypeVlan,
		State: "up",
		Vlan: Vlan{
			BaseIface: baseInterface,
			ID:        int(vlanID),
		},
		Ipv4: InterfaceIpv4{
			Enabled: true,
			Dhcp:    false,
			Address: []InterfaceIPAddress{{
				PrefixLen: 24,
				IP:        net.ParseIP(ipv4Addresses),
			}},
		},
		Ipv6: InterfaceIpv6{
			Enabled:  true,
			Dhcp:     false,
			Autoconf: false,
			Address: []InterfaceIPAddress{{
				PrefixLen: 64,
				IP:        net.ParseIP(ipv6Addresses),
			}},
		},
	}

	return builder.withInterface(newInterface)
}

// WithInterfaceAltnames adds ethernet interface configuration with alternative names to the policy.
func (builder *PolicyBuilder) WithInterfaceAltnames(interfaceName string, altnames []string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with an interface altnames %v",
		builder.Definition.Name, altnames)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		builder.SetError(fmt.Errorf(nodeNetConfPolIntError))

		return builder
	}

	if len(altnames) == 0 {
		klog.V(100).Info("The altnames can not be empty array")

		builder.SetError(fmt.Errorf(nodeNetConfPolAltnamesEmpty))

		return builder
	}

	altnamesList := make([]InterfaceAltName, len(altnames))
	for i, altname := range altnames {
		altnamesList[i] = InterfaceAltName{
			Name: altname,
		}
	}

	newInterface := NetworkInterface{
		Name:     interfaceName,
		Type:     interfaceTypeEthernet,
		State:    "up",
		AltNames: altnamesList,
	}

	return builder.withInterface(newInterface)
}

// WithMACAddressAltnames adds ethernet interface configuration identified by MAC address with alternative names.
func (builder *PolicyBuilder) WithMACAddressAltnames(interfaceName string, macaddress string, altnames []string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with an interface macaddress %s and altnames %v",
		builder.Definition.Name, macaddress, altnames)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		builder.SetError(fmt.Errorf(nodeNetConfPolIntError))

		return builder
	}

	if macaddress == "" {
		klog.V(100).Info("The macaddress can not be empty string")

		builder.SetError(fmt.Errorf("macaddress cannot be empty string"))

		return builder
	}

	if _, err := net.ParseMAC(macaddress); err != nil {
		klog.V(100).Infof("The macaddress %s is invalid", macaddress)

		builder.SetError(fmt.Errorf("macaddress is invalid"))

		return builder
	}

	if len(altnames) == 0 {
		klog.V(100).Info("The altnames can not be empty array")

		builder.SetError(fmt.Errorf(nodeNetConfPolAltnamesEmpty))

		return builder
	}

	altnamesList := make([]InterfaceAltName, len(altnames))
	for i, altname := range altnames {
		altnamesList[i] = InterfaceAltName{
			Name: altname,
		}
	}

	newInterface := NetworkInterface{
		Name:       interfaceName,
		Type:       interfaceTypeEthernet,
		State:      "up",
		MacAddress: macaddress,
		Identifier: "mac-address",
		AltNames:   altnamesList,
	}

	return builder.withInterface(newInterface)
}

// WithPCIAddressAltnames adds ethernet interface configuration identified by PCI address with alternative names.
func (builder *PolicyBuilder) WithPCIAddressAltnames(interfaceName string, pciAddress string, altnames []string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with an interface pci-address %s and altnames %v",
		builder.Definition.Name, pciAddress, altnames)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		builder.SetError(fmt.Errorf(nodeNetConfPolIntError))

		return builder
	}

	if pciAddress == "" {
		klog.V(100).Info("The pciAddress can not be empty string")

		builder.SetError(fmt.Errorf("pciAddress cannot be empty string"))

		return builder
	}

	if !pciAddressRegexp.MatchString(pciAddress) {
		klog.V(100).Infof("The pciAddress %s is invalid", pciAddress)

		builder.SetError(fmt.Errorf("pciAddress is invalid"))

		return builder
	}

	if len(altnames) == 0 {
		klog.V(100).Info("The altnames can not be empty array")

		builder.SetError(fmt.Errorf(nodeNetConfPolAltnamesEmpty))

		return builder
	}

	altnamesList := make([]InterfaceAltName, len(altnames))
	for i, altname := range altnames {
		altnamesList[i] = InterfaceAltName{
			Name: altname,
		}
	}

	newInterface := NetworkInterface{
		Name:       interfaceName,
		Type:       interfaceTypeEthernet,
		State:      "up",
		PciAddress: pciAddress,
		Identifier: "pci-address",
		AltNames:   altnamesList,
	}

	return builder.withInterface(newInterface)
}

// WithEthernetInterface adds type ethernet interface and IPs configuration to the NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) WithEthernetInterface(interfaceName, ipv4Address, ipv6Address string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with an ethernet interface %s",
		builder.Definition.Name, interfaceName)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		builder.SetError(fmt.Errorf(nodeNetConfPolIntError))

		return builder
	}

	if net.ParseIP(ipv4Address) == nil {
		klog.V(100).Info("the ethernet interface contains an invalid ipv4 address")

		builder.SetError(fmt.Errorf("ethernet interface 'ipv4Addresses' is an invalid ipv4 address"))

		return builder
	}

	if net.ParseIP(ipv6Address) == nil {
		klog.V(100).Info("the ethernet interface contains an invalid ipv6 address")

		builder.SetError(fmt.Errorf("ethernet interface 'ipv6Addresses' is an invalid ipv6 address"))

		return builder
	}

	newInterface := NetworkInterface{
		Name:  interfaceName,
		Type:  interfaceTypeEthernet,
		State: "up",
		Ipv4: InterfaceIpv4{
			Enabled: true,
			Dhcp:    false,
			Address: []InterfaceIPAddress{{
				PrefixLen: 24,
				IP:        net.ParseIP(ipv4Address),
			}},
		},
		Ipv6: InterfaceIpv6{
			Enabled:  true,
			Dhcp:     false,
			Autoconf: false,
			Address: []InterfaceIPAddress{{
				PrefixLen: 64,
				IP:        net.ParseIP(ipv6Address),
			}},
		},
	}

	return builder.withInterface(newInterface)
}

// WithEthernetIPv6LinkLocalInterface enables IPv6 and adds link-local address to interface configuration in the
// NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) WithEthernetIPv6LinkLocalInterface(interfaceName string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with an ethernet interface %s",
		builder.Definition.Name, interfaceName)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		builder.SetError(fmt.Errorf(nodeNetConfPolIntError))

		return builder
	}

	newInterface := NetworkInterface{
		Name:  interfaceName,
		Type:  interfaceTypeEthernet,
		State: "up",
		Ipv4: InterfaceIpv4{
			Enabled: false,
		},
		Ipv6: InterfaceIpv6{
			Enabled: true,
		},
	}

	return builder.withInterface(newInterface)
}

// WithEthernetDualStackInterface appends the configuration for an ethernet interface with state "up",
// IPv4 enabled, and IPv6 enabled to the NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) WithEthernetDualStackInterface(interfaceName string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with a dual-stack ethernet interface %s",
		builder.Definition.Name, interfaceName)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		builder.SetError(fmt.Errorf(nodeNetConfPolIntError))

		return builder
	}

	newInterface := NetworkInterface{
		Name:  interfaceName,
		Type:  interfaceTypeEthernet,
		State: "up",
		Ipv4: InterfaceIpv4{
			Enabled: true,
		},
		Ipv6: InterfaceIpv6{
			Enabled: true,
		},
	}

	return builder.withInterface(newInterface)
}

// WithInterfaceUp appends the configuration for an interface with state "up" to the NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) WithInterfaceUp(interfaceName string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s to set interface %s to state up",
		builder.Definition.Name, interfaceName)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		builder.SetError(fmt.Errorf(nodeNetConfPolIntError))

		return builder
	}

	newInterface := NetworkInterface{
		Name:  interfaceName,
		State: "up",
		Type:  interfaceTypeEthernet,
	}

	return builder.withInterface(newInterface)
}

// WithAbsentInterface appends the configuration for an absent interface to the NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) WithAbsentInterface(interfaceName string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with absent interface configuration:"+
		" interface %s", builder.Definition.Name, interfaceName)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		builder.SetError(fmt.Errorf(nodeNetConfPolIntError))

		return builder
	}

	newInterface := NetworkInterface{
		Name:  interfaceName,
		State: interfaceStateAbsent,
	}

	return builder.withInterface(newInterface)
}

// WithStaticRoute adds a static route to the NodeNetworkConfigurationPolicy desired state.
func (builder *PolicyBuilder) WithStaticRoute(
	destination, nextHopAddress, nextHopInterface string, metric, tableID int) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding static route %s via %s dev %s to NodeNetworkConfigurationPolicy %s",
		destination, nextHopAddress, nextHopInterface, builder.Definition.Name)

	if destination == "" {
		builder.SetError(fmt.Errorf("route 'destination' cannot be empty"))

		return builder
	}

	if nextHopAddress == "" && nextHopInterface == "" {
		builder.SetError(fmt.Errorf("route must have either 'nextHopAddress' or 'nextHopInterface'"))

		return builder
	}

	var currentState DesiredState

	err := yaml.Unmarshal(builder.Definition.Spec.DesiredState.Raw, &currentState)
	if err != nil {
		klog.V(100).Info("Failed Unmarshal DesiredState")

		builder.SetError(fmt.Errorf("%s", "Failed Unmarshal DesiredState"))

		return builder
	}

	if currentState.Routes == nil {
		currentState.Routes = &DesiredRoutes{}
	}

	currentState.Routes.Config = append(currentState.Routes.Config, RouteConfig{
		Destination:      destination,
		NextHopAddress:   nextHopAddress,
		NextHopInterface: nextHopInterface,
		Metric:           metric,
		TableID:          tableID,
	})

	desiredStateYaml, err := yaml.Marshal(currentState)
	if err != nil {
		klog.V(100).Info("Failed Marshal DesiredState")

		builder.SetError(fmt.Errorf("failed to Marshal a new Desired state"))

		return builder
	}

	builder.Definition.Spec.DesiredState = nmstateShared.NewState(string(desiredStateYaml))

	return builder
}

// RemoveInterfaceAltname appends configuration to mark the given alternative names absent on the interface.
func (builder *PolicyBuilder) RemoveInterfaceAltname(interfaceName string, altnames []string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Removing interface altnames for interface %s from NodeNetworkConfigurationPolicy %s",
		interfaceName, builder.Definition.Name)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		builder.SetError(fmt.Errorf(nodeNetConfPolIntError))

		return builder
	}

	altnamesList := []InterfaceAltName{}
	for _, altname := range altnames {
		altnamesList = append(altnamesList, InterfaceAltName{Name: altname, State: interfaceStateAbsent})
	}

	newInterface := NetworkInterface{
		Name:     interfaceName,
		Type:     interfaceTypeEthernet,
		State:    "up",
		AltNames: altnamesList,
	}

	return builder.withInterface(newInterface)
}

// WaitUntilCondition waits for the duration of the defined timeout or until the
// NodeNetworkConfigurationPolicy gets to a specific condition.
func (builder *PolicyBuilder) WaitUntilCondition(condition nmstateShared.ConditionType, timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Waiting for the defined period until NodeNetworkConfigurationPolicy %s has condition %v",
		builder.Definition.Name, condition)

	if !builder.Exists() {
		return fmt.Errorf("cannot wait for NodeNetworkConfigurationPolicy condition because it does not exist")
	}

	// Polls every retryInterval to determine if NodeNetworkConfigurationPolicy is in desired condition.
	var err error

	return wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			for _, cond := range builder.Object.Status.Conditions {
				if cond.Type == condition && cond.Status == corev1.ConditionTrue {
					return true, nil
				}
			}

			return false, nil
		})
}

// withInterface adds given network interface to the NodeNetworkConfigurationPolicy.
func (builder *PolicyBuilder) withInterface(networkInterface NetworkInterface) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating NodeNetworkConfigurationPolicy %s with network interface %s",
		builder.Definition.Name, networkInterface.Name)

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Definition.Spec.DesiredState.Raw, &CurrentState)
	if err != nil {
		klog.V(100).Info("Failed Unmarshal DesiredState")

		builder.SetError(fmt.Errorf("%s", "Failed Unmarshal DesiredState"))

		return builder
	}

	CurrentState.Interfaces = append(CurrentState.Interfaces, networkInterface)

	desiredStateYaml, err := yaml.Marshal(CurrentState)
	if err != nil {
		klog.V(100).Info("Failed Marshal DesiredState")

		builder.SetError(fmt.Errorf("failed to Marshal a new Desired state"))

		return builder
	}

	builder.Definition.Spec.DesiredState = nmstateShared.NewState(string(desiredStateYaml))

	return builder
}
