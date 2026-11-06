package sriov

import (
	"context"
	"fmt"
	"time"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var networkNodeStateGVK = srIovV1.GroupVersion.WithKind("SriovNetworkNodeState")

// NetworkNodeStateBuilder provides struct for SriovNetworkNodeState object which contains connection to cluster and
// SriovNetworkNodeState definitions.
type NetworkNodeStateBuilder struct {
	common.EmbeddableBuilder[srIovV1.SriovNetworkNodeState, *srIovV1.SriovNetworkNodeState]
	// Objects retains the legacy public field and is kept in sync with the embedded builder's Object field.
	Objects *srIovV1.SriovNetworkNodeState
}

// GetGVK returns the SriovNetworkNodeState GVK for this builder.
func (builder *NetworkNodeStateBuilder) GetGVK() schema.GroupVersionKind {
	return networkNodeStateGVK
}

// GetObject returns the last object stored in the builder.
func (builder *NetworkNodeStateBuilder) GetObject() *srIovV1.SriovNetworkNodeState {
	return builder.Objects
}

// SetObject updates both the legacy Objects field and the embedded builder's Object field.
func (builder *NetworkNodeStateBuilder) SetObject(object *srIovV1.SriovNetworkNodeState) {
	builder.Objects = object
	builder.EmbeddableBuilder.SetObject(object)
}

// Exists checks whether the SriovNetworkNodeState exists and keeps the legacy Objects field synchronized.
func (builder *NetworkNodeStateBuilder) Exists() bool {
	if !builder.EmbeddableBuilder.Exists() {
		return false
	}

	builder.SetObject(builder.EmbeddableBuilder.GetObject())

	return true
}

// NewNetworkNodeStateBuilder creates new instance of NetworkNodeStateBuilder.
func NewNetworkNodeStateBuilder(apiClient *clients.Settings, nodeName, nsname string) *NetworkNodeStateBuilder {
	return common.NewNamespacedBuilder[srIovV1.SriovNetworkNodeState, NetworkNodeStateBuilder](
		apiClient, srIovV1.AddToScheme, nodeName, nsname)
}

// Discover method gets the SriovNetworkNodeState items and stores them in the NetworkNodeStateBuilder struct.
func (builder *NetworkNodeStateBuilder) Discover() error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Getting the SriovNetworkNodeState object in namespace %s for node %s",
		builder.Definition.Namespace, builder.Definition.Name)

	nodeNetworkState, err := builder.Get()
	if err != nil {
		return err
	}

	builder.SetObject(nodeNetworkState)

	return nil
}

// GetUpNICs returns a list of SrIov interfaces in UP state.
func (builder *NetworkNodeStateBuilder) GetUpNICs() (srIovV1.InterfaceExts, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Collection of sriov interfaces in UP state for node %s", builder.Definition.Name)

	sriovNics, err := builder.GetNICs()
	if err != nil {
		klog.V(100).Infof("Error to discover sriov interfaces for node %s", builder.Definition.Name)

		return nil, err
	}

	var sriovNicsUp srIovV1.InterfaceExts

	for _, nic := range sriovNics {
		if nic.LinkSpeed != "" && nic.LinkSpeed != "-1 Mb/s" {
			klog.V(100).Infof("Interface %s is UP on node %s. Append to list", nic.Name, builder.Definition.Name)
			sriovNicsUp = append(sriovNicsUp, nic)
		}
	}

	klog.V(100).Infof("Collected sriov UP interfaces list %v for node %s",
		builder.Objects.Status.Interfaces, builder.Definition.Name)

	return sriovNicsUp, nil
}

// GetNICs returns a list of SrIov interfaces.
func (builder *NetworkNodeStateBuilder) GetNICs() (srIovV1.InterfaceExts, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	if err := builder.Discover(); err != nil {
		klog.V(100).Infof("Error to discover sriov interfaces for node %s", builder.Definition.Name)

		return nil, err
	}

	klog.V(100).Infof("Collected sriov interfaces list %v for node %s",
		builder.Objects.Status.Interfaces, builder.Definition.Name)

	return builder.Objects.Status.Interfaces, nil
}

// WaitUntilSyncStatus waits for the duration of the defined timeout or until the
// SriovNetworkNodeState gets to a specific syncStatus.
func (builder *NetworkNodeStateBuilder) WaitUntilSyncStatus(syncStatus string, timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Waiting for the defined period until SriovNetworkNodeState %s has syncStatus %s",
		builder.Definition.Name, syncStatus)

	if syncStatus == "" {
		klog.V(100).Info("The syncStatus parameter is empty")

		return fmt.Errorf("syncStatus cannot be empty")
	}

	// Polls every retryInterval to determine if SriovNetworkNodeState is in desired syncStatus.
	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			err := builder.Discover()
			if err != nil {
				return false, nil
			}

			return builder.Objects.Status.SyncStatus == syncStatus, nil
		})
}

// GetNumVFs returns num-vfs under the given interface.
func (builder *NetworkNodeStateBuilder) GetNumVFs(sriovInterfaceName string) (int, error) {
	klog.V(100).Infof("Getting num-vfs under interface %s from SriovNetworkNodeState %s",
		sriovInterfaceName, builder.Definition.Name)

	sriovInterface, err := builder.findInterfaceByName(sriovInterfaceName)
	if err != nil {
		return 0, err
	}

	return sriovInterface.NumVfs, nil
}

// GetDriverName returns driver name under the given interface.
func (builder *NetworkNodeStateBuilder) GetDriverName(sriovInterfaceName string) (string, error) {
	klog.V(100).Infof("Getting driver name for interface %s from SriovNetworkNodeState %s",
		sriovInterfaceName, builder.Definition.Name)

	sriovInterface, err := builder.findInterfaceByName(sriovInterfaceName)
	if err != nil {
		return "", err
	}

	return sriovInterface.Driver, nil
}

// GetPciAddress returns PciAddress under the given interface.
func (builder *NetworkNodeStateBuilder) GetPciAddress(sriovInterfaceName string) (string, error) {
	klog.V(100).Infof("Getting PCI address for interface %s from SriovNetworkNodeState %s",
		sriovInterfaceName, builder.Definition.Name)

	sriovInterface, err := builder.findInterfaceByName(sriovInterfaceName)
	if err != nil {
		return "", err
	}

	return sriovInterface.PciAddress, nil
}

// GetTotalVFs returns total VFs under the given interface.
func (builder *NetworkNodeStateBuilder) GetTotalVFs(sriovInterfaceName string) (int, error) {
	klog.V(100).Infof("Getting totalvfs under interface %s from SriovNetworkNodeState %s",
		sriovInterfaceName, builder.Definition.Name)

	sriovInterface, err := builder.findInterfaceByName(sriovInterfaceName)
	if err != nil {
		return 0, err
	}

	return sriovInterface.TotalVfs, nil
}

func (builder *NetworkNodeStateBuilder) findInterfaceByName(sriovInterfaceName string) (*srIovV1.InterfaceExt, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	if err := builder.Discover(); err != nil {
		klog.V(100).Infof("Error to discover sriov network node state for node %s", builder.Definition.Name)

		return nil, err
	}

	if sriovInterfaceName == "" {
		klog.V(100).Info("The sriovInterface can not be empty string")

		return nil, fmt.Errorf("sriovInterface can not be empty string")
	}

	for _, sriovInterface := range builder.Objects.Status.Interfaces {
		if sriovInterface.Name == sriovInterfaceName {
			return &sriovInterface, nil
		}
	}

	return nil, fmt.Errorf("interface %s was not found", sriovInterfaceName)
}
