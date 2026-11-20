package metallb

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errEmptyIPAddressPools = "error: IPAddressPools setting is empty list, the list should contain at least one element"
	errEmptyNodeSelectors  = "error: nodeSelectors setting is empty list, the list should contain at least one element"
)

var bgpAdvertisementGVK = mlbtypes.GroupVersion.WithKind("BGPAdvertisement")

// BGPAdvertisementBuilder provides struct for the BGPAdvertisement object containing connection to
// the cluster and the BGPAdvertisement definitions.
type BGPAdvertisementBuilder struct {
	common.EmbeddableBuilder[mlbtypes.BGPAdvertisement, *mlbtypes.BGPAdvertisement]
	common.EmbeddableCreator[
		mlbtypes.BGPAdvertisement,
		BGPAdvertisementBuilder,
		*mlbtypes.BGPAdvertisement,
		*BGPAdvertisementBuilder,
	]
	common.EmbeddableDeleteReturner[
		mlbtypes.BGPAdvertisement,
		BGPAdvertisementBuilder,
		*mlbtypes.BGPAdvertisement,
		*BGPAdvertisementBuilder,
	]
	common.EmbeddableForceUpdater[
		mlbtypes.BGPAdvertisement,
		BGPAdvertisementBuilder,
		*mlbtypes.BGPAdvertisement,
		*BGPAdvertisementBuilder,
	]
	common.EmbeddableWithOptions[
		mlbtypes.BGPAdvertisement,
		BGPAdvertisementBuilder,
		*mlbtypes.BGPAdvertisement,
		*BGPAdvertisementBuilder,
		BGPAdvertisementAdditionalOptions,
	]
}

// BGPAdvertisementAdditionalOptions additional options for BGPAdvertisement object.
type BGPAdvertisementAdditionalOptions func(builder *BGPAdvertisementBuilder) (*BGPAdvertisementBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *BGPAdvertisementBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the BGPAdvertisement GVK for this builder.
func (builder *BGPAdvertisementBuilder) GetGVK() schema.GroupVersionKind {
	return bgpAdvertisementGVK
}

// NewBGPAdvertisementBuilder creates a new instance of BGPAdvertisementBuilder.
func NewBGPAdvertisementBuilder(apiClient *clients.Settings, name, nsname string) *BGPAdvertisementBuilder {
	klog.V(100).Infof(
		"Initializing new BGPAdvertisement structure with the following params: %s, %s",
		name, nsname)

	return common.NewNamespacedBuilder[mlbtypes.BGPAdvertisement, BGPAdvertisementBuilder](
		apiClient, mlbtypes.AddToScheme, name, nsname)
}

// PullBGPAdvertisement pulls existing bgpadvertisement from cluster.
func PullBGPAdvertisement(apiClient *clients.Settings, name, nsname string) (*BGPAdvertisementBuilder, error) {
	klog.V(100).Infof("Pulling existing bgpadvertisement name %s under namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[mlbtypes.BGPAdvertisement, BGPAdvertisementBuilder](
		context.TODO(), apiClient, mlbtypes.AddToScheme, name, nsname)
}

// WithAggregationLength4 adds the specified AggregationLength to the BGPAdvertisement.
func (builder *BGPAdvertisementBuilder) WithAggregationLength4(aggregationLength int32) *BGPAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPAdvertisement %s in namespace %s with aggregationLength: %d",
		builder.Definition.Name, builder.Definition.Namespace, aggregationLength)

	if aggregationLength < 0 || aggregationLength > 32 {
		builder.SetError(fmt.Errorf(
			"AggregationLength %d is invalid, the value shoud be in range 0...32", aggregationLength))

		return builder
	}

	builder.Definition.Spec.AggregationLength = &aggregationLength

	return builder
}

// WithAggregationLength6 adds the specified AggregationLengthV6 to the BGPAdvertisement.
func (builder *BGPAdvertisementBuilder) WithAggregationLength6(aggregationLength int32) *BGPAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPAdvertisement %s in namespace %s with aggregationLength6: %d",
		builder.Definition.Name, builder.Definition.Namespace, aggregationLength)

	if aggregationLength < 0 || aggregationLength > 128 {
		builder.SetError(fmt.Errorf(
			"AggregationLength %d is invalid, the value shoud be in range 0...128", aggregationLength))

		return builder
	}

	builder.Definition.Spec.AggregationLengthV6 = &aggregationLength

	return builder
}

// WithLocalPref adds the specified LocalPref to the BGPAdvertisement.
func (builder *BGPAdvertisementBuilder) WithLocalPref(localPreference uint32) *BGPAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPAdvertisement %s in namespace %s with LocalPref: %d",
		builder.Definition.Name, builder.Definition.Namespace, localPreference)

	builder.Definition.Spec.LocalPref = localPreference

	return builder
}

// WithCommunities adds the specified Communities to the BGPAdvertisement.
func (builder *BGPAdvertisementBuilder) WithCommunities(communities []string) *BGPAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPAdvertisement %s in namespace %s with Communities: %s",
		builder.Definition.Name, builder.Definition.Namespace, communities)

	if len(communities) < 1 {
		builder.SetError(fmt.Errorf(
			"error: community setting is empty list, the list should contain at least one element"))

		return builder
	}

	builder.Definition.Spec.Communities = communities

	return builder
}

// WithIPAddressPools adds the specified IPAddressPools to the BGPAdvertisement.
func (builder *BGPAdvertisementBuilder) WithIPAddressPools(ipAddressPools []string) *BGPAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPAdvertisement %s in namespace %s with IPAddressPools: %s",
		builder.Definition.Name, builder.Definition.Namespace, ipAddressPools)

	if len(ipAddressPools) < 1 {
		builder.SetError(fmt.Errorf("%s", errEmptyIPAddressPools))

		return builder
	}

	builder.Definition.Spec.IPAddressPools = ipAddressPools

	return builder
}

// WithIPAddressPoolsSelectors adds the specified IPAddressPoolSelectors to the BGPAdvertisement.
func (builder *BGPAdvertisementBuilder) WithIPAddressPoolsSelectors(
	poolSelector []metav1.LabelSelector) *BGPAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPAdvertisement %s in namespace %s with IPAddressPoolSelectors: %s",
		builder.Definition.Name, builder.Definition.Namespace, poolSelector)

	if len(poolSelector) < 1 {
		builder.SetError(fmt.Errorf(
			"error: IPAddressPoolSelectors setting is empty list, the list should contain at least one element"))

		return builder
	}

	builder.Definition.Spec.IPAddressPoolSelectors = poolSelector

	return builder
}

// WithNodeSelector adds the specified NodeSelectors to the BGPAdvertisement.
func (builder *BGPAdvertisementBuilder) WithNodeSelector(
	nodeSelectors []metav1.LabelSelector) *BGPAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPAdvertisement %s in namespace %s with nodeSelectors: %v",
		builder.Definition.Name, builder.Definition.Namespace, nodeSelectors)

	if len(nodeSelectors) < 1 {
		builder.SetError(fmt.Errorf("%s", errEmptyNodeSelectors))

		return builder
	}

	builder.Definition.Spec.NodeSelectors = nodeSelectors

	return builder
}

// WithPeers adds the specified Peers to the BGPAdvertisement.
func (builder *BGPAdvertisementBuilder) WithPeers(peers []string) *BGPAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BGPAdvertisement %s in namespace %s with Peers: %v",
		builder.Definition.Name, builder.Definition.Namespace, peers)

	if len(peers) < 1 {
		builder.SetError(fmt.Errorf(
			"error: peers setting is empty list, the list should contain at least one element"))

		return builder
	}

	builder.Definition.Spec.Peers = peers

	return builder
}

// GetBGPAdvertisementGVR returns bgpadvertisement's GroupVersionResource, which could be used for Clean function.
func GetBGPAdvertisementGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "bgpadvertisements",
	}
}
