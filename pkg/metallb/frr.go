package metallb

import (
	"fmt"
	"net"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	frrtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const frrconfigurationConnecttimeValueIsNotValid = "frrConfiguration 'connectTime' value is not valid"

var frrConfigurationGVK = frrtypes.GroupVersion.WithKind("FRRConfiguration")

// FrrConfigurationBuilder provides struct for the FrrConfiguration object containing connection to
// the cluster and the FrrConfiguration definitions.
type FrrConfigurationBuilder struct {
	common.EmbeddableBuilder[frrtypes.FRRConfiguration, *frrtypes.FRRConfiguration]
	common.EmbeddableCreator[
		frrtypes.FRRConfiguration, FrrConfigurationBuilder, *frrtypes.FRRConfiguration, *FrrConfigurationBuilder]
	common.EmbeddableDeleter[frrtypes.FRRConfiguration, *frrtypes.FRRConfiguration]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *FrrConfigurationBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the FRRConfiguration GVK for this builder.
func (builder *FrrConfigurationBuilder) GetGVK() schema.GroupVersionKind {
	return frrConfigurationGVK
}

// NewFrrConfigurationBuilder creates a new instance of FRRConfiguration.
func NewFrrConfigurationBuilder(
	apiClient *clients.Settings, name, nsname string) *FrrConfigurationBuilder {
	return common.NewNamespacedBuilder[frrtypes.FRRConfiguration, FrrConfigurationBuilder](
		apiClient, frrtypes.AddToScheme, name, nsname)
}

// GetFrrConfigurationGVR returns FrrConfiguration GroupVersionResource which could be used for Clean function.
func GetFrrConfigurationGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: FRRAPIGroup, Version: APIVersion, Resource: "frrconfigurations",
	}
}

// WithBGPRouter defines a BGP router as localASN.
func (builder *FrrConfigurationBuilder) WithBGPRouter(localASN uint32) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Define BGP router with an ASN: %v", localASN)

	// If there are routers, append a new one
	builder.Definition.Spec.BGP.Routers = append(builder.Definition.Spec.BGP.Routers, frrtypes.Router{
		ASN: localASN,
	})

	return builder
}

// WithToReceiveModeAll allows all incoming routes.
func (builder *FrrConfigurationBuilder) WithToReceiveModeAll(routerIndex, neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Allows the FRR nodes to receive all routes in namespace %s with mode type: all",
		builder.Definition.Namespace)

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	// Update the ToReceive mode to AllowAll for the specified neighbor
	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].ToReceive.Allowed.Mode = frrtypes.AllowAll

	return builder
}

// WithToReceiveModeFiltered allows only specified prefixes to be received.
func (builder *FrrConfigurationBuilder) WithToReceiveModeFiltered(prefixes []string,
	routerIndex, neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Allows the FRR nodes to receive specific routes in namespace %s in mode type: filtered",
		builder.Definition.Namespace)

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	// Validate CIDR prefixes
	for _, prefix := range prefixes {
		if _, _, err := net.ParseCIDR(prefix); err != nil {
			klog.V(100).Infof("the frrConfiguration prefix %s is not a valid CIDR", prefix)
			builder.SetError(fmt.Errorf("the prefix %s is not a valid CIDR", prefix))

			return builder
		}
	}

	// Prepare the list of allowed prefixes
	allowedPrefixes := []frrtypes.PrefixSelector{}

	for _, prefix := range prefixes {
		klog.V(100).Infof("AllowMode 'filtered' receives only routes within the prefix %v", prefix)
		allowedPrefixes = append(allowedPrefixes, frrtypes.PrefixSelector{Prefix: prefix})
	}

	// Update the specific neighbor's ToReceive field
	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].ToReceive.Allowed.Prefixes = allowedPrefixes

	return builder
}

// WithBGPNeighbor defines a single neighbor IP and ASN number.
func (builder *FrrConfigurationBuilder) WithBGPNeighbor(bgpPeerIP string,
	remoteAS uint32, routerIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Define BGP Neighbor %s with peer IP %s and ASN %v",
		builder.Definition.Name, bgpPeerIP, remoteAS)

	if net.ParseIP(bgpPeerIP) == nil {
		klog.V(100).Infof("The peerIP for the frrConfiguration bgp peer contains invalid ip address %s",
			bgpPeerIP)

		builder.SetError(fmt.Errorf("frrConfiguration 'peerIP' of the BGPPeer contains invalid ip address"))

		return builder
	}

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors =
		append(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors,
			frrtypes.Neighbor{
				Address: bgpPeerIP,
				ASN:     remoteAS,
			})

	return builder
}

// WithBGPPassword defines the password used between BGP peers to form adjacency and attaches the password using the
// neighbor index from the neighbor list.
func (builder *FrrConfigurationBuilder) WithBGPPassword(bgpPassword string,
	routerIndex, neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Define FRRConfiguration %s in namespace %s with the password: %s",
		builder.Definition.Name, builder.Definition.Namespace, bgpPassword)

	if bgpPassword == "" {
		klog.V(100).Info("The bgpPassword should not be blank")

		builder.SetError(fmt.Errorf("the bgpPassword %s is an empty string", bgpPassword))

		return builder
	}

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	// Set the password for the specified neighbor
	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].Password = bgpPassword

	return builder
}

// WithEBGPMultiHop sets the EBGP multihop setting on the remote BGP peer.
func (builder *FrrConfigurationBuilder) WithEBGPMultiHop(routerIndex, neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Sets the EBGP multihop setting to true in the FrrConfiguration %s",
		builder.Definition.Name)

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].EBGPMultiHop = true

	return builder
}

// WithBGPNeighborDisableMP sets the DisableMP flag for the specified neighbor.
func (builder *FrrConfigurationBuilder) WithBGPNeighborDisableMP(disableMP bool,
	routerIndex, neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting FRRConfiguration %s neighbor disableMP to %t",
		builder.Definition.Name, disableMP)

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	// Set the DisableMP for the specified neighbor
	//nolint:staticcheck // DisableMP is deprecated, keep for backwards compatibility.
	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].DisableMP = disableMP

	return builder
}

// WithBGPNeighborDualStackAddressFamily enables dual stack address family for the specified BGP neighbor.
// This allows to advertise/receive IPv4 prefixes over IPv6 sessions and vice versa.
func (builder *FrrConfigurationBuilder) WithBGPNeighborDualStackAddressFamily(dualStackAddressFamily bool,
	routerIndex, neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting FRRConfiguration %s neighbor dualStackAddressFamily to %t",
		builder.Definition.Name, dualStackAddressFamily)

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	// Set the DualStackAddressFamily for the specified neighbor
	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].
		DualStackAddressFamily = dualStackAddressFamily

	return builder
}

// WithHoldTime defines the holdTime placed in the FrrConfiguration spec.
func (builder *FrrConfigurationBuilder) WithHoldTime(holdTime metav1.Duration, routerIndex,
	neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding a neighbor specific holdTime in the FrrConfiguration %s with holdTime: %v",
		builder.Definition.Name, holdTime)

	duration := holdTime.Duration

	if duration < time.Second || duration > 65535*time.Second {
		klog.V(100).Info("A valid holdtime is between 1-65535")

		builder.SetError(fmt.Errorf("frrConfiguration 'holdtime' value is not valid"))

		return builder
	}

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].HoldTime = &holdTime

	return builder
}

// WithKeepalive defines the keepAliveTime placed in the FrrConfiguration spec.
func (builder *FrrConfigurationBuilder) WithKeepalive(keepAlive metav1.Duration, routerIndex,
	neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Add a keepAlive timer to a specific neighbor in FrrConfiguration %s with keepalive: %v",
		builder.Definition.Name, keepAlive)

	duration := keepAlive.Duration

	if duration < time.Second || duration > 65535*time.Second {
		klog.V(100).Info("A valid keepAlive time is between 1-65535")

		builder.SetError(fmt.Errorf("frrConfiguration 'keepAlive' value is not valid"))

		return builder
	}

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].KeepaliveTime = &keepAlive

	return builder
}

// WithConnectTime defines the time to wait until trying to connect to a BGP peer.
func (builder *FrrConfigurationBuilder) WithConnectTime(connectTime metav1.Duration, routerIndex,
	neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Defines a connect retry hold time for a specific neighbor in FrrConfiguration %s with holdTime %s ",
		builder.Definition.Name, connectTime)

	duration := connectTime.Duration

	if duration < time.Second || duration > 65535*time.Second {
		klog.V(100).Info("A valid connect time is between 1-65535")

		builder.SetError(fmt.Errorf(frrconfigurationConnecttimeValueIsNotValid))

		return builder
	}

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].ConnectTime = &connectTime

	return builder
}

// WithPort defines the tcp port used to make BGP adjacency.
func (builder *FrrConfigurationBuilder) WithPort(port uint16, routerIndex,
	neighborIndex uint) *FrrConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Defines the tcp port to establish adjaceny to a neighbor in FrrConfiguration %s port number %d",
		builder.Definition.Name, port)

	if port > 16384 {
		klog.V(100).Infof("Invalid port number can not be greater then 16384: %d", port)

		builder.SetError(fmt.Errorf("invalid port number: %d", port))

		return builder
	}

	// Check if the routerIndex is within bounds
	if routerIndex >= uint(len(builder.Definition.Spec.BGP.Routers)) {
		klog.V(100).Infof("Invalid routerIndex: %d", routerIndex)

		builder.SetError(fmt.Errorf("invalid routerIndex: %d", routerIndex))

		return builder
	}

	// Check if the neighborIndex is within bounds
	if neighborIndex >= uint(len(builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors)) {
		klog.V(100).Infof("Invalid neighborIndex: %d", neighborIndex)

		builder.SetError(fmt.Errorf("invalid neighborIndex: %d", neighborIndex))

		return builder
	}

	builder.Definition.Spec.BGP.Routers[routerIndex].Neighbors[neighborIndex].Port = &port

	return builder
}
