package nmstate

import (
	"context"
	"fmt"
	"slices"

	"gopkg.in/yaml.v2"

	"k8s.io/klog/v2"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"

	nmstateV1beta1 "github.com/nmstate/kubernetes-nmstate/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	interfaceTypeBond = "bond"
	interfaceTypeVlan = "vlan"
	unknown           = "unknown"
)

var (
	nodeNetworkStateGVK = nmstateV1beta1.GroupVersion.WithKind("NodeNetworkState")

	// allowedInterfaceTypes represents all allowed types for interface.
	allowedInterfaceTypes = []string{interfaceTypeEthernet, interfaceTypeBond, "ovs-bridge", unknown,
		interfaceTypeVlan, "vxlan", "linux-bridge", "team", "veth"}
)

// StateBuilder provides struct for the NodeNetworkState object containing connection to the cluster.
type StateBuilder struct {
	common.EmbeddableBuilder[nmstateV1beta1.NodeNetworkState, *nmstateV1beta1.NodeNetworkState]
}

// GetGVK returns the NodeNetworkState GVK for this builder.
func (builder *StateBuilder) GetGVK() schema.GroupVersionKind {
	return nodeNetworkStateGVK
}

// GetTotalVFs returns total-vfs under the given interface.
func (builder *StateBuilder) GetTotalVFs(sriovInterfaceName string) (int, error) {
	if err := builder.validateCurrentState(); err != nil {
		return 0, err
	}

	klog.V(100).Infof(
		"Getting total-vfs under interface %s from NodeNetworkState %s",
		sriovInterfaceName, builder.Object.Name)

	if sriovInterfaceName == "" {
		klog.V(100).Info("The sriovInterfaceName can not be empty string")

		return 0, fmt.Errorf("the sriovInterfaceName is empty sting")
	}

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &CurrentState)
	if err != nil {
		return 0, fmt.Errorf("failed to Unmarshal NMState state")
	}

	for _, interfaceFromCurrentState := range CurrentState.Interfaces {
		if interfaceFromCurrentState.Name == sriovInterfaceName {
			return *interfaceFromCurrentState.Ethernet.Sriov.TotalVfs, nil
		}
	}

	return 0, fmt.Errorf("failed to find interface %s", sriovInterfaceName)
}

// GetInterfaceAltnames returns non-empty alternative interface names for the given interface from current state.
func (builder *StateBuilder) GetInterfaceAltnames(interfaceName string) ([]string, error) {
	if err := builder.validateCurrentState(); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting interface altnames for interface %s from NodeNetworkState %s",
		interfaceName, builder.Object.Name)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		return nil, fmt.Errorf("the interfaceName is empty sting")
	}

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &CurrentState)
	if err != nil {
		return nil, fmt.Errorf("failed to Unmarshal NMState state")
	}

	for _, interfaceFromCurrentState := range CurrentState.Interfaces {
		if interfaceFromCurrentState.Name == interfaceName {
			names := make([]string, 0, len(interfaceFromCurrentState.AltNames))
			for _, alt := range interfaceFromCurrentState.AltNames {
				if alt.Name != "" {
					names = append(names, alt.Name)
				}
			}

			return names, nil
		}
	}

	return nil, fmt.Errorf("failed to find interface %s", interfaceName)
}

// GetInterfaceMACAddress returns the mac-address field for the interface from NodeNetworkState current state.
func (builder *StateBuilder) GetInterfaceMACAddress(interfaceName string) (string, error) {
	if err := builder.validateCurrentState(); err != nil {
		return "", err
	}

	klog.V(100).Infof("Getting interface MAC address for interface %s from NodeNetworkState %s",
		interfaceName, builder.Object.Name)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		return "", fmt.Errorf("the interfaceName is empty sting")
	}

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &CurrentState)
	if err != nil {
		return "", fmt.Errorf("failed to Unmarshal NMState state")
	}

	for _, interfaceFromCurrentState := range CurrentState.Interfaces {
		if interfaceFromCurrentState.Name == interfaceName {
			if interfaceFromCurrentState.MacAddress == "" {
				return "", fmt.Errorf("failed to find mac-address for interface %s", interfaceName)
			}

			return interfaceFromCurrentState.MacAddress, nil
		}
	}

	return "", fmt.Errorf("failed to find interface %s", interfaceName)
}

// GetInterfacePCIAddress returns the pci-address field for the interface from NodeNetworkState current state.
func (builder *StateBuilder) GetInterfacePCIAddress(interfaceName string) (string, error) {
	if err := builder.validateCurrentState(); err != nil {
		return "", err
	}

	klog.V(100).Infof("Getting interface PCI address for interface %s from NodeNetworkState %s",
		interfaceName, builder.Object.Name)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		return "", fmt.Errorf("the interfaceName is empty sting")
	}

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &CurrentState)
	if err != nil {
		return "", fmt.Errorf("failed to Unmarshal NMState state")
	}

	for _, interfaceFromCurrentState := range CurrentState.Interfaces {
		if interfaceFromCurrentState.Name == interfaceName {
			if interfaceFromCurrentState.PciAddress == "" {
				return "", fmt.Errorf("failed to find pci-address for interface %s", interfaceName)
			}

			return interfaceFromCurrentState.PciAddress, nil
		}
	}

	return "", fmt.Errorf("failed to find interface %s", interfaceName)
}

// GetInterfaceType returns Interface with the given interface name and given type.
func (builder *StateBuilder) GetInterfaceType(interfaceName, interfaceType string) (NetworkInterface, error) {
	if err := builder.validateCurrentState(); err != nil {
		return NetworkInterface{}, err
	}

	klog.V(100).Infof(
		"Getting interface %s with type %s from NodeNetworkState %s",
		interfaceName, interfaceType, builder.Object.Name)

	if interfaceName == "" {
		klog.V(100).Info("The interfaceName can not be empty string")

		return NetworkInterface{}, fmt.Errorf("the interfaceName is empty sting")
	}

	if !slices.Contains(allowedInterfaceTypes, interfaceType) {
		klog.V(100).Infof("error to add type %s, allowed types are %v", interfaceType, allowedInterfaceTypes)

		return NetworkInterface{}, fmt.Errorf("invalid interfaceType parameter")
	}

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &CurrentState)
	if err != nil {
		return NetworkInterface{}, fmt.Errorf("failed to Unmarshal NMState state")
	}

	for _, interf := range CurrentState.Interfaces {
		if interf.Name == interfaceName && interf.Type == interfaceType {
			return interf, nil
		}
	}

	return NetworkInterface{}, fmt.Errorf("failed to find interface %s or it is not a %s type",
		interfaceName, interfaceType)
}

// GetSriovVfs returns all configured VFs  under the given SR-IOV interface.
func (builder *StateBuilder) GetSriovVfs(sriovInterfaceName string) ([]Vf, error) {
	if err := builder.validateCurrentState(); err != nil {
		return nil, err
	}

	klog.V(100).Infof(
		"Getting all configured VFs under interface %s from NodeNetworkState %s",
		sriovInterfaceName, builder.Object.Name)

	if sriovInterfaceName == "" {
		klog.V(100).Info("The sriovInterfaceName can not be empty string")

		return nil, fmt.Errorf("the sriovInterfaceName is empty sting")
	}

	var CurrentState DesiredState

	err := yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &CurrentState)
	if err != nil {
		return nil, fmt.Errorf("failed to Unmarshal NMState state")
	}

	for _, networkInterface := range CurrentState.Interfaces {
		if networkInterface.Name == sriovInterfaceName && networkInterface.Ethernet.Sriov.Vfs != nil {
			return networkInterface.Ethernet.Sriov.Vfs, nil
		}
	}

	return nil, fmt.Errorf("failed to find interface %s "+
		"or SR-IOV VFs are not configured on it", sriovInterfaceName)
}

// PullNodeNetworkState retrieves an existing NodeNetworkState object from the cluster.
func PullNodeNetworkState(apiClient *clients.Settings, name string) (*StateBuilder, error) {
	return common.PullClusterScopedBuilder[nmstateV1beta1.NodeNetworkState, StateBuilder](
		context.TODO(), apiClient, nmstateV1beta1.AddToScheme, name)
}

// validateCurrentState checks that the builder is valid and has a pulled NodeNetworkState object.
func (builder *StateBuilder) validateCurrentState() error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	if builder.Object == nil {
		return fmt.Errorf("NodeNetworkState object is undefined")
	}

	return nil
}
