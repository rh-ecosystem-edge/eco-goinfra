package metallb

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var l2AdvertisementGVK = mlbtypes.GroupVersion.WithKind("L2Advertisement")

// L2AdvertisementBuilder provides struct for the L2Advertisement object containing connection to
// the cluster and the L2Advertisement definitions.
type L2AdvertisementBuilder struct {
	common.EmbeddableBuilder[mlbtypes.L2Advertisement, *mlbtypes.L2Advertisement]
	common.EmbeddableCreator[
		mlbtypes.L2Advertisement,
		L2AdvertisementBuilder,
		*mlbtypes.L2Advertisement,
		*L2AdvertisementBuilder,
	]
	common.EmbeddableDeleteReturner[
		mlbtypes.L2Advertisement,
		L2AdvertisementBuilder,
		*mlbtypes.L2Advertisement,
		*L2AdvertisementBuilder,
	]
	common.EmbeddableForceUpdater[
		mlbtypes.L2Advertisement,
		L2AdvertisementBuilder,
		*mlbtypes.L2Advertisement,
		*L2AdvertisementBuilder,
	]
	common.EmbeddableWithOptions[
		mlbtypes.L2Advertisement,
		L2AdvertisementBuilder,
		*mlbtypes.L2Advertisement,
		*L2AdvertisementBuilder,
		L2AdvertisementAdditionalOptions,
	]
}

// L2AdvertisementAdditionalOptions additional options for L2Advertisement object.
type L2AdvertisementAdditionalOptions func(builder *L2AdvertisementBuilder) (*L2AdvertisementBuilder, error)

// AttachMixins wires the embedded CRUD and option mixins to this builder instance.
func (builder *L2AdvertisementBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the L2Advertisement GVK for this builder.
func (builder *L2AdvertisementBuilder) GetGVK() schema.GroupVersionKind {
	return l2AdvertisementGVK
}

// NewL2AdvertisementBuilder creates a new instance of L2AdvertisementBuilder.
func NewL2AdvertisementBuilder(apiClient *clients.Settings, name, nsname string) *L2AdvertisementBuilder {
	klog.V(100).Infof(
		"Initializing new L2Advertisement structure with the following params: %s, %s",
		name, nsname)

	return common.NewNamespacedBuilder[mlbtypes.L2Advertisement, L2AdvertisementBuilder](
		apiClient, mlbtypes.AddToScheme, name, nsname)
}

// PullL2Advertisement pulls existing L2Advertisement from cluster.
func PullL2Advertisement(apiClient *clients.Settings, name, nsname string) (*L2AdvertisementBuilder, error) {
	klog.V(100).Infof("Pulling existing L2Advertisement name %s under namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[mlbtypes.L2Advertisement, L2AdvertisementBuilder](
		context.TODO(), apiClient, mlbtypes.AddToScheme, name, nsname)
}

// WithNodeSelector adds the specified NodeSelectors to the L2Advertisement.
func (builder *L2AdvertisementBuilder) WithNodeSelector(nodeSelectors []metaV1.LabelSelector) *L2AdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Appending L2Advertisement %s in namespace %s with nodeSelectors: %v",
		builder.Definition.Name, builder.Definition.Namespace, nodeSelectors)

	if len(nodeSelectors) < 1 {
		builder.SetError(fmt.Errorf("%s", errEmptyNodeSelectors))

		return builder
	}

	builder.Definition.Spec.NodeSelectors = nodeSelectors

	return builder
}

// WithIPAddressPools adds the specified IPAddressPools to the L2Advertisement.
func (builder *L2AdvertisementBuilder) WithIPAddressPools(ipAddressPools []string) *L2AdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Appending L2Advertisement %s in namespace %s with IPAddressPools: %v",
		builder.Definition.Name, builder.Definition.Namespace, ipAddressPools)

	if len(ipAddressPools) < 1 {
		builder.SetError(fmt.Errorf("%s", errEmptyIPAddressPools))

		return builder
	}

	builder.Definition.Spec.IPAddressPools = ipAddressPools

	return builder
}

// WithIPAddressPoolsSelectors adds the specified IPAddressPoolSelectors to the L2Advertisement.
func (builder *L2AdvertisementBuilder) WithIPAddressPoolsSelectors(
	poolSelector []metaV1.LabelSelector) *L2AdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Appending L2Advertisement %s in namespace %s with IPAddressPoolSelectors: %v",
		builder.Definition.Name, builder.Definition.Namespace, poolSelector)

	if len(poolSelector) < 1 {
		builder.SetError(fmt.Errorf(
			"error: IPAddressPoolSelectors setting is empty list, the list should contain at least one element"))

		return builder
	}

	builder.Definition.Spec.IPAddressPoolSelectors = poolSelector

	return builder
}

// WithInterfaces adds the specified interfaces to the L2Advertisement.
func (builder *L2AdvertisementBuilder) WithInterfaces(interfaces []string) *L2AdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Appending L2Advertisement %s in namespace %s with Interfaces: %v",
		builder.Definition.Name, builder.Definition.Namespace, interfaces)

	if len(interfaces) < 1 {
		builder.SetError(fmt.Errorf(
			"error: Interfaces setting is empty list, the list should contain at least one element"))

		return builder
	}

	builder.Definition.Spec.Interfaces = interfaces

	return builder
}

// GetL2AdvertisementGVR returns l2advertisement's GroupVersionResource, which could be used for Clean function.
func GetL2AdvertisementGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "l2advertisements",
	}
}
