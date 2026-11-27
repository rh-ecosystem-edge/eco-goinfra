package ovn

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// RouteAdvertisementBuilder provides a wrapper around RouteAdvertisements objects for the Kubernetes API.
type RouteAdvertisementBuilder struct {
	common.EmbeddableBuilder[ovnv1.RouteAdvertisements, *ovnv1.RouteAdvertisements]
	common.EmbeddableCreator[
		ovnv1.RouteAdvertisements, RouteAdvertisementBuilder,
		*ovnv1.RouteAdvertisements, *RouteAdvertisementBuilder,
	]
	common.EmbeddableDeleter[ovnv1.RouteAdvertisements, *ovnv1.RouteAdvertisements]
	common.EmbeddableUpdater[
		ovnv1.RouteAdvertisements, RouteAdvertisementBuilder,
		*ovnv1.RouteAdvertisements, *RouteAdvertisementBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *RouteAdvertisementBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the RouteAdvertisements GVK for this builder.
func (builder *RouteAdvertisementBuilder) GetGVK() schema.GroupVersionKind {
	return ovnv1.SchemeGroupVersion.WithKind("RouteAdvertisements")
}

// NewRouteAdvertisementBuilder creates a new instance of RouteAdvertisementBuilder.
func NewRouteAdvertisementBuilder(apiClient *clients.Settings, name string) *RouteAdvertisementBuilder {
	return common.NewClusterScopedBuilder[ovnv1.RouteAdvertisements, RouteAdvertisementBuilder](
		apiClient, ovnv1.AddToScheme, name)
}

// PullRouteAdvertisement pulls existing RouteAdvertisement from cluster.
func PullRouteAdvertisement(apiClient *clients.Settings, name string) (*RouteAdvertisementBuilder, error) {
	return common.PullClusterScopedBuilder[ovnv1.RouteAdvertisements, RouteAdvertisementBuilder](
		context.TODO(), apiClient, ovnv1.AddToScheme, name)
}

// WithTargetVRF sets the targetVRF field in the RouteAdvertisement definition.
func (builder *RouteAdvertisementBuilder) WithTargetVRF(targetVRF string) *RouteAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting RouteAdvertisement %s targetVRF to %s", builder.Definition.Name, targetVRF)

	builder.Definition.Spec.TargetVRF = targetVRF

	return builder
}

// WithAdvertisements sets the advertisements field in the RouteAdvertisement definition.
func (builder *RouteAdvertisementBuilder) WithAdvertisements(
	advertisements []ovnv1.AdvertisementType) *RouteAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting RouteAdvertisement %s advertisements to %v", builder.Definition.Name, advertisements)

	if len(advertisements) == 0 {
		klog.V(100).Infof("RouteAdvertisement 'advertisements' cannot be empty")

		builder.SetError(fmt.Errorf("RouteAdvertisement 'advertisements' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Advertisements = advertisements

	return builder
}

// WithNodeSelector sets the nodeSelector field in the RouteAdvertisement definition.
func (builder *RouteAdvertisementBuilder) WithNodeSelector(
	nodeSelector metav1.LabelSelector) *RouteAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting RouteAdvertisement %s nodeSelector to %v", builder.Definition.Name, nodeSelector)

	builder.Definition.Spec.NodeSelector = nodeSelector

	return builder
}

// WithFRRConfigurationSelector sets the frrConfigurationSelector field in the RouteAdvertisement definition.
func (builder *RouteAdvertisementBuilder) WithFRRConfigurationSelector(
	frrConfigurationSelector metav1.LabelSelector) *RouteAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting RouteAdvertisement %s frrConfigurationSelector to %v",
		builder.Definition.Name, frrConfigurationSelector)

	builder.Definition.Spec.FRRConfigurationSelector = frrConfigurationSelector

	return builder
}

// WithNetworkSelectors sets the networkSelectors field in the RouteAdvertisement definition.
func (builder *RouteAdvertisementBuilder) WithNetworkSelectors(
	networkSelectors types.NetworkSelectors) *RouteAdvertisementBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting RouteAdvertisement %s networkSelectors to %v", builder.Definition.Name, networkSelectors)

	if len(networkSelectors) == 0 {
		klog.V(100).Infof("RouteAdvertisement 'networkSelectors' cannot be empty")

		builder.SetError(fmt.Errorf("RouteAdvertisement 'networkSelectors' cannot be empty"))

		return builder
	}

	// Validate each network selector has a valid NetworkSelectionType.
	for index, selector := range networkSelectors {
		if selector.NetworkSelectionType == "" {
			klog.V(100).Infof("RouteAdvertisement networkSelector[%d] has empty NetworkSelectionType", index)

			builder.SetError(fmt.Errorf("RouteAdvertisement 'networkSelectors' must have valid NetworkSelectionType"))

			return builder
		}

		switch selector.NetworkSelectionType {
		case types.DefaultNetwork, types.ClusterUserDefinedNetworks,
			types.PrimaryUserDefinedNetworks, types.SecondaryUserDefinedNetworks,
			types.NetworkAttachmentDefinitions:
			// Valid types.
		default:
			klog.V(100).Infof("RouteAdvertisement networkSelector[%d] has invalid NetworkSelectionType: %s",
				index, selector.NetworkSelectionType)

			builder.SetError(fmt.Errorf(
				"RouteAdvertisement 'networkSelectors[%d]' has invalid NetworkSelectionType: %s",
				index, selector.NetworkSelectionType))

			return builder
		}
	}

	builder.Definition.Spec.NetworkSelectors = networkSelectors

	return builder
}

// GetRouteAdvertisementGVR returns RouteAdvertisement's GroupVersionResource which could be used for Clean function.
func GetRouteAdvertisementGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: ovnv1.GroupName, Version: ovnv1.SchemeGroupVersion.Version, Resource: "routeadvertisements",
	}
}
