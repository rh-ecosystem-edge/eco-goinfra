package metallb

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypesv1beta2"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	bgppeerPeeripOfTheBgppeerContains    = "BGPPeer 'peerIP' of the BGPPeer contains invalid ip address"
	bgppeerBgppeeripOfTheBgppeerContains = "BGPPeer 'bgpPeerIP' of the BGPPeer contains invalid ip address"
	bgppeerConnecttimeValueIsNotValid    = "bgppeer 'connectTime' value is not valid"
)

var bgpPeerGVK = mlbtypesv1beta2.GroupVersion.WithKind("BGPPeer")

// BGPPeerBuilder provides struct for the BGPPeer object containing connection to
// the cluster and the BGPPeer definitions.
type BGPPeerBuilder struct {
	common.EmbeddableBuilder[mlbtypesv1beta2.BGPPeer, *mlbtypesv1beta2.BGPPeer]
	common.EmbeddableCreator[
		mlbtypesv1beta2.BGPPeer, BGPPeerBuilder, *mlbtypesv1beta2.BGPPeer, *BGPPeerBuilder]
	common.EmbeddableDeleteReturner[
		mlbtypesv1beta2.BGPPeer, BGPPeerBuilder, *mlbtypesv1beta2.BGPPeer, *BGPPeerBuilder]
	common.EmbeddableForceUpdater[
		mlbtypesv1beta2.BGPPeer, BGPPeerBuilder, *mlbtypesv1beta2.BGPPeer, *BGPPeerBuilder]
	common.EmbeddableWithOptions[
		mlbtypesv1beta2.BGPPeer, BGPPeerBuilder, *mlbtypesv1beta2.BGPPeer, *BGPPeerBuilder,
		BGPPeerAdditionalOptions]
}

// BGPPeerAdditionalOptions additional options for BGPPeer object.
type BGPPeerAdditionalOptions func(builder *BGPPeerBuilder) (*BGPPeerBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *BGPPeerBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the BGPPeer GVK for this builder.
func (builder *BGPPeerBuilder) GetGVK() schema.GroupVersionKind {
	return bgpPeerGVK
}

// NewBPGPeerBuilder creates a new instance of BGPPeer.
func NewBPGPeerBuilder(
	apiClient *clients.Settings, name, nsname, peerIP string, asn, remoteASN uint32) *BGPPeerBuilder {
	klog.V(100).Infof(
		"Initializing new BGPPeer structure with the following params: %s, %s %s %d %d",
		name, nsname, peerIP, asn, remoteASN)

	builder := common.NewNamespacedBuilder[mlbtypesv1beta2.BGPPeer, BGPPeerBuilder](
		apiClient, mlbtypesv1beta2.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Spec.MyASN = asn
	builder.Definition.Spec.ASN = remoteASN
	builder.Definition.Spec.Address = peerIP

	if net.ParseIP(peerIP) == nil {
		klog.V(100).Infof("The peerIP of the BGPPeer contains invalid ip address %s", peerIP)

		builder.SetError(fmt.Errorf(bgppeerPeeripOfTheBgppeerContains))

		return builder
	}

	return builder
}

// NewBGPPeerBuilder creates a new instance of BGPPeer.
func NewBGPPeerBuilder(
	apiClient *clients.Settings, name, nsname string, asn, remoteASN uint32) *BGPPeerBuilder {
	klog.V(100).Infof(
		"Initializing new BGPPeer structure with the following params: %s, %s %d %d",
		name, nsname, asn, remoteASN)

	builder := common.NewNamespacedBuilder[mlbtypesv1beta2.BGPPeer, BGPPeerBuilder](
		apiClient, mlbtypesv1beta2.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Spec.MyASN = asn
	builder.Definition.Spec.ASN = remoteASN

	return builder
}

// PullBGPPeer pulls existing bgppeer from cluster.
func PullBGPPeer(apiClient *clients.Settings, name, nsname string) (*BGPPeerBuilder, error) {
	klog.V(100).Infof("Pulling existing bgppeer name %s under namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[mlbtypesv1beta2.BGPPeer, BGPPeerBuilder](
		context.TODO(), apiClient, mlbtypesv1beta2.AddToScheme, name, nsname)
}

// WithBGPPeerIP defines the peer IP address.
func (builder *BGPPeerBuilder) WithBGPPeerIP(bgpPeerIP string) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating an BGPPeer %s in namespace %s with IP address: %s",
		builder.Definition.Name, builder.Definition.Namespace, bgpPeerIP)

	if net.ParseIP(bgpPeerIP) == nil {
		klog.V(100).Infof("The peerIP of the BGPPeer contains invalid ip address %s", bgpPeerIP)

		builder.SetError(fmt.Errorf(bgppeerBgppeeripOfTheBgppeerContains))

		return builder
	}

	builder.Definition.Spec.Address = bgpPeerIP

	return builder
}

// WithIPUnnumbered defines the interface to be used with the interface BGPPeer spec.
func (builder *BGPPeerBuilder) WithIPUnnumbered(interfaceName string) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating an unnumbered BGPPeer %s in namespace %s with interface: %s",
		builder.Definition.Name, builder.Definition.Namespace, interfaceName)

	if interfaceName == "" {
		klog.V(100).Info("Can not redefine BGPPeer with empty interface string")

		builder.SetError(fmt.Errorf("interface can not be empty string"))

		return builder
	}

	builder.Definition.Spec.Interface = interfaceName

	return builder
}

// WithDynamicASN defines the dynamicASN as either internal (iBGP) or external (eBGP). Both remoteAS and dynamicASN
// configure the remote ASN. They are mutually exclusive and only one can be used per remote peer.
func (builder *BGPPeerBuilder) WithDynamicASN(dynamicASN mlbtypesv1beta2.DynamicASNMode) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s using a dynamicASN: %s",
		builder.Definition.Name, builder.Definition.Namespace, dynamicASN)

	if dynamicASN != "internal" && dynamicASN != "external" {
		klog.V(100).Info("The dynamicASN of the BGPPeer is incorrect")

		builder.SetError(fmt.Errorf("bgpPeer 'dynamicASN' must be either internal or external"))

		return builder
	}

	builder.Definition.Spec.ASN = 0
	builder.Definition.Spec.DynamicASN = dynamicASN

	return builder
}

// WithRouterID defines the routerID placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithRouterID(routerID string) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this routerID: %s",
		builder.Definition.Name, builder.Definition.Namespace, routerID)

	if net.ParseIP(routerID) == nil {
		klog.V(100).Infof("The routerID of the BGPPeer contains invalid ip address %s, "+
			"routerID should be present in ip address format", routerID)

		builder.SetError(fmt.Errorf("the routerID of the BGPPeer contains invalid ip address %s", routerID))

		return builder
	}

	builder.Definition.Spec.RouterID = routerID

	return builder
}

// WithBFDProfile defines the bfdProfile placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithBFDProfile(bfdProfile string) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this bfdProfile: %s",
		builder.Definition.Name, builder.Definition.Namespace, bfdProfile)

	if bfdProfile == "" {
		klog.V(100).Info("The bfdProfile of the BGPPeer can not be empty string")

		//nolint:staticcheck // Preserve the existing public error text.
		builder.SetError(fmt.Errorf("The bfdProfile is empty string"))

		return builder
	}

	builder.Definition.Spec.BFDProfile = bfdProfile

	return builder
}

// WithSRCAddress defines the SRCAddress placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithSRCAddress(srcAddress string) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this srcAddress: %s",
		builder.Definition.Name, builder.Definition.Namespace, srcAddress)

	if net.ParseIP(srcAddress) == nil {
		klog.V(100).Infof("The srcAddress of the BGPPeer contains invalid ip address %s, "+
			"srcAddress should be present in ip address format", srcAddress)

		builder.SetError(fmt.Errorf("the srcAddress of the BGPPeer contains invalid ip address %s", srcAddress))

		return builder
	}

	builder.Definition.Spec.SrcAddress = srcAddress

	return builder
}

// WithPort defines the port placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithPort(port uint16) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this port: %d",
		builder.Definition.Name, builder.Definition.Namespace, port)

	builder.Definition.Spec.Port = port

	return builder
}

// WithHoldTime defines the holdTime placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithHoldTime(holdTime metav1.Duration) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this holdTime: %s",
		builder.Definition.Name, builder.Definition.Namespace, holdTime)

	builder.Definition.Spec.HoldTime = &holdTime

	return builder
}

// WithKeepalive defines the keepAliveTime placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithKeepalive(keepalive metav1.Duration) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this keepalive: %s",
		builder.Definition.Name, builder.Definition.Namespace, keepalive)

	builder.Definition.Spec.KeepaliveTime = &keepalive

	return builder
}

// WithConnectTime defines the reconnect timer between BGP neighbors.
func (builder *BGPPeerBuilder) WithConnectTime(connectTime metav1.Duration) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this connectTime: %s",
		builder.Definition.Name, builder.Definition.Namespace, connectTime)

	duration := connectTime.Duration

	if duration < time.Second || duration > 65535*time.Second {
		klog.V(100).Info("A valid connect time is between 1-65535")

		builder.SetError(fmt.Errorf(bgppeerConnecttimeValueIsNotValid))

		return builder
	}

	builder.Definition.Spec.ConnectTime = &connectTime

	return builder
}

// WithNodeSelector defines the nodeSelector placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithNodeSelector(nodeSelector map[string]string) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this nodeSelector: %s",
		builder.Definition.Name, builder.Definition.Namespace, nodeSelector)

	if len(nodeSelector) == 0 {
		klog.V(100).Info("Can not redefine BGPPeer with empty nodeSelector map")

		builder.SetError(fmt.Errorf("BGPPeer 'nodeSelector' cannot be empty map"))

		return builder
	}

	builder.Definition.Spec.NodeSelectors = []metav1.LabelSelector{{
		MatchLabels: nodeSelector,
	}}

	return builder
}

// WithPassword defines the password placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithPassword(password string) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this password: %s",
		builder.Definition.Name, builder.Definition.Namespace, password)

	if password == "" {
		klog.V(100).Info("Can not redefine BGPPeer with empty password")

		builder.SetError(fmt.Errorf("password can not be empty string"))

		return builder
	}

	builder.Definition.Spec.Password = password

	return builder
}

// WithEBGPMultiHop defines the EBGPMultiHop bool flag placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithEBGPMultiHop(eBGPMultiHop bool) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this eBGPMultiHop flag: %t",
		builder.Definition.Name, builder.Definition.Namespace, eBGPMultiHop)

	builder.Definition.Spec.EBGPMultiHop = eBGPMultiHop

	return builder
}

// WithGracefulRestart defines the EnableGracefulRestart bool flag placed in the BGPPeer spec.
func (builder *BGPPeerBuilder) WithGracefulRestart(gracefulRestart bool) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPPeer %s in namespace %s with this EnableGracefulRestart flag: %t",
		builder.Definition.Name, builder.Definition.Namespace, gracefulRestart)

	builder.Definition.Spec.EnableGracefulRestart = gracefulRestart

	return builder
}

// WithDisableMP disables Multiprotocol BGP for this peer.
func (builder *BGPPeerBuilder) WithDisableMP(disableMP bool) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting BGPPeer %s disableMP to %t", builder.Definition.Name, disableMP)

	//nolint:staticcheck // DisableMP is deprecated, keep for backward compatibility.
	builder.Definition.Spec.DisableMP = disableMP

	return builder
}

// WithDualStackAddressFamily enables dual stack address family for this BGP peer.
// This allows to advertise/receive IPv4 prefixes over IPv6 sessions and vice versa.
func (builder *BGPPeerBuilder) WithDualStackAddressFamily(dualStackAddressFamily bool) *BGPPeerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting BGPPeer %s dualStackAddressFamily to %t", builder.Definition.Name, dualStackAddressFamily)

	builder.Definition.Spec.DualStackAddressFamily = dualStackAddressFamily

	return builder
}

// GetBGPPeerGVR returns bgppeer's GroupVersionResource which could be used for Clean function.
func GetBGPPeerGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "bgppeers",
	}
}
