package sriov

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	errEmptyTargetNsname = "SrIovNetwork 'targetNsname' cannot be empty"
	spoofChkOff          = "off"
	errInvalidLinkState  = "invalid 'linkState' parameters"
	errInvalidVlanQoS    = "Invalid QoS class. Supported vlan QoS class values are between 0...7"
)

var sriovNetworkGVK = srIovV1.GroupVersion.WithKind("SriovNetwork")

// NetworkBuilder provides struct for srIovNetwork object which contains connection to cluster and
// srIovNetwork definition.
type NetworkBuilder struct {
	common.EmbeddableBuilder[srIovV1.SriovNetwork, *srIovV1.SriovNetwork]
	common.EmbeddableCreator[srIovV1.SriovNetwork, NetworkBuilder, *srIovV1.SriovNetwork, *NetworkBuilder]
	common.EmbeddableDeleter[srIovV1.SriovNetwork, *srIovV1.SriovNetwork]
	common.EmbeddableForceUpdater[srIovV1.SriovNetwork, NetworkBuilder, *srIovV1.SriovNetwork, *NetworkBuilder]
	common.EmbeddableWithOptions[
		srIovV1.SriovNetwork, NetworkBuilder, *srIovV1.SriovNetwork, *NetworkBuilder, NetworkAdditionalOptions]
}

// NetworkAdditionalOptions additional options for SriovNetwork object.
type NetworkAdditionalOptions func(builder *NetworkBuilder) (*NetworkBuilder, error)

// AttachMixins wires the embedded CRUD and option mixins to this builder instance.
func (builder *NetworkBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the SriovNetwork GVK for this builder.
func (builder *NetworkBuilder) GetGVK() schema.GroupVersionKind {
	return sriovNetworkGVK
}

// NewNetworkBuilder creates new instance of Builder.
func NewNetworkBuilder(
	apiClient *clients.Settings, name, nsname, targetNsname, resName string) *NetworkBuilder {
	builder := common.NewNamespacedBuilder[srIovV1.SriovNetwork, NetworkBuilder](
		apiClient, srIovV1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	if resName == "" {
		builder.SetError(fmt.Errorf("SrIovNetwork 'resName' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.ResourceName = resName
	builder.Definition.Spec.NetworkNamespace = targetNsname

	return builder
}

// WithTargetNamespace sets the target namespace for the SriovNetwork.
func (builder *NetworkBuilder) WithTargetNamespace(targetNsname string) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding targetNsname %s to SriovNetwork %s in namespace %s",
		targetNsname, builder.Definition.Name, builder.Definition.Namespace)

	if targetNsname == "" {
		klog.V(100).Infof(errEmptyTargetNsname)

		builder.SetError(fmt.Errorf("%s", errEmptyTargetNsname))

		return builder
	}

	builder.Definition.Spec.NetworkNamespace = targetNsname

	return builder
}

// WithVLAN sets vlan id in the SrIovNetwork definition. Allowed vlanId range is between 0-4094.
func (builder *NetworkBuilder) WithVLAN(vlanID uint16) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if vlanID > 4094 {
		builder.SetError(fmt.Errorf("invalid vlanID, allowed vlanID values are between 0-4094"))

		return builder
	}

	builder.Definition.Spec.Vlan = int(vlanID)

	return builder
}

// WithVlanProto sets the VLAN protocol for qinq tunneling protocol in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithVlanProto(vlanProtocol string) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Info("Setting SriovNetwork vlanProtocol support for qinq")

	allowedVlanProto := []string{"802.1q", "802.1Q", "802.1ad", "802.1AD"}
	if !slices.Contains(allowedVlanProto, vlanProtocol) {
		builder.SetError(fmt.Errorf("invalid 'vlanProtocol' parameters"))
	}

	builder.Definition.Spec.VlanProto = vlanProtocol

	return builder
}

// WithSpoof sets spoof flag based on the given argument in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithSpoof(enabled bool) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if enabled {
		builder.Definition.Spec.SpoofChk = "on"
	} else {
		builder.Definition.Spec.SpoofChk = spoofChkOff
	}

	return builder
}

// WithMetaPluginAllMultiFlag metaplugin activates allmulti multicast mode on a SriovNetwork configuration.
func (builder *NetworkBuilder) WithMetaPluginAllMultiFlag(allMultiFlag bool) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	builder.Definition.Spec.MetaPluginsConfig = fmt.Sprintf(`{ "type": "tuning", "allmulti": %t }`, allMultiFlag)

	return builder
}

// WithMetaPluginRdma activates rdma plugin on a SriovNetwork configuration.
func (builder *NetworkBuilder) WithMetaPluginRdma() *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	builder.Definition.Spec.MetaPluginsConfig = `{ "type": "rdma" }`

	return builder
}

// WithLinkState sets linkState parameters in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithLinkState(linkState string) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	allowedLinkStates := []string{"enable", "disable", "auto"}

	if !slices.Contains(allowedLinkStates, linkState) {
		builder.SetError(fmt.Errorf("%s", errInvalidLinkState))

		return builder
	}

	builder.Definition.Spec.LinkState = linkState

	return builder
}

// WithMaxTxRate sets maxTxRate parameters in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithMaxTxRate(maxTxRate uint16) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	maxTxRateInt := int(maxTxRate)

	builder.Definition.Spec.MaxTxRate = &maxTxRateInt

	return builder
}

// WithMinTxRate sets minTxRate parameters in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithMinTxRate(minTxRate uint16) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	maxTxRateInt := int(minTxRate)
	builder.Definition.Spec.MaxTxRate = &maxTxRateInt

	return builder
}

// WithTrustFlag sets trust flag based on the given argument in the SrIoVNetwork definition spec.
func (builder *NetworkBuilder) WithTrustFlag(enabled bool) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if enabled {
		builder.Definition.Spec.Trust = "on"
	} else {
		builder.Definition.Spec.Trust = spoofChkOff
	}

	return builder
}

// WithVlanQoS sets qoSClass parameters in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithVlanQoS(qoSClass uint16) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if qoSClass > 7 {
		builder.SetError(fmt.Errorf("%s", errInvalidVlanQoS))

		return builder
	}

	qoSClassInt := int(qoSClass)

	builder.Definition.Spec.VlanQoS = qoSClassInt

	return builder
}

// WithLogLevel sets logLevel parameter in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithLogLevel(logLevel string) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	allowedLogLevels := []string{"panic", "error", "warning", "info", "debug", ""}

	if !slices.Contains(allowedLogLevels, logLevel) {
		builder.SetError(fmt.Errorf("invalid logLevel value, allowed logLevel values are:" +
			" panic, error, warning, info, debug or empty"))

		return builder
	}

	builder.Definition.Spec.LogLevel = logLevel

	return builder
}

// WithIPAddressSupport sets ips capabilities in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithIPAddressSupport() *NetworkBuilder {
	return builder.withCapabilities("ips")
}

// WithMacAddressSupport sets mac capabilities in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithMacAddressSupport() *NetworkBuilder {
	return builder.withCapabilities("mac")
}

// WithStaticIpam sets static IPAM in the SrIovNetwork definition spec.
func (builder *NetworkBuilder) WithStaticIpam() *NetworkBuilder {
	return builder.withIpam("static")
}

// WithWhereaboutsIPAM sets whereabouts IPAM with range/gateway and optional exclude/network name.
func (builder *NetworkBuilder) WithWhereaboutsIPAM(ipRange, gateway, exclude, networkName string) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting SriovNetwork whereabouts IPAM with range %s and gateway %s", ipRange, gateway)

	if ipRange == "" {
		builder.SetError(fmt.Errorf("failed to configure IPAM, 'ipRange' parameter is empty"))

		return builder
	}

	if gateway == "" {
		builder.SetError(fmt.Errorf("failed to configure IPAM, 'gateway' parameter is empty"))

		return builder
	}

	ipamConfig := map[string]any{
		"type":    "whereabouts",
		"range":   ipRange,
		"gateway": gateway,
	}

	if exclude != "" {
		ipamConfig["exclude"] = []string{exclude}
	}

	if networkName != "" {
		ipamConfig["network_name"] = networkName
	}

	ipamJSON, err := json.Marshal(ipamConfig)
	if err != nil {
		builder.SetError(fmt.Errorf("failed to marshal IPAM config: %w", err))

		return builder
	}

	builder.Definition.Spec.IPAM = string(ipamJSON)

	return builder
}

// PullNetwork pulls existing sriovnetwork from cluster.
func PullNetwork(apiClient *clients.Settings, name, nsname string) (*NetworkBuilder, error) {
	klog.V(100).Infof("Pulling existing sriovnetwork name %s under namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[srIovV1.SriovNetwork, NetworkBuilder](
		context.TODO(), apiClient, srIovV1.AddToScheme, name, nsname)
}

// DeleteAndWait deletes the SrIovNetwork resource and waits until it is deleted.
func (builder *NetworkBuilder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Deleting SrIovNetwork %s in namespace %s and waiting for the defined period until it is removed",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.Delete()
	if err != nil {
		return err
	}

	return builder.WaitUntilDeleted(timeout)
}

// WaitUntilDeleted waits for the duration of the defined timeout or until the SrIovNetwork is deleted.
func (builder *NetworkBuilder) WaitUntilDeleted(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof(
		"Waiting for the defined period until SrIovNetwork %s in namespace %s is deleted",
		builder.Definition.Name, builder.Definition.Namespace)

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			_, err := builder.Get()
			if err == nil {
				klog.V(100).Infof("SrIovNetwork %s/%s still present", builder.Definition.Name, builder.Definition.Namespace)

				return false, nil
			}

			if k8serrors.IsNotFound(err) {
				klog.V(100).Infof("SrIovNetwork %s/%s is gone", builder.Definition.Name, builder.Definition.Namespace)

				return true, nil
			}

			klog.V(100).Infof("Failed to get SrIovNetwork %s/%s: %v", builder.Definition.Name, builder.Definition.Namespace, err)

			return false, err
		})
}

// GetSriovNetworksGVR returns SriovNetwork's GroupVersionResource which could be used for Clean function.
func GetSriovNetworksGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: "sriovnetwork.openshift.io", Version: "v1", Resource: "sriovnetworks",
	}
}

func (builder *NetworkBuilder) withCapabilities(capability string) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if len(builder.Definition.Spec.Capabilities) == 0 {
		builder.Definition.Spec.Capabilities = fmt.Sprintf(`{ "%s": true }`, capability)
	} else {
		builder.Definition.Spec.Capabilities =
			"{" + builder.Definition.Spec.Capabilities[1:len(builder.Definition.Spec.Capabilities)-2] +
				fmt.Sprintf(`, "%s": true `, capability) + "}"
	}

	return builder
}

func (builder *NetworkBuilder) withIpam(ipamType string) *NetworkBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if ipamType == "" {
		klog.V(100).Info("sriov network 'ipamType' parameter can not be empty")

		builder.SetError(fmt.Errorf("failed to configure IPAM, 'ipamType' parameter is empty"))

		return builder
	}

	builder.Definition.Spec.IPAM = fmt.Sprintf(`{ "type": "%s" }`, ipamType)

	return builder
}
