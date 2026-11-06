package metallb

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const resourceIPAddressPools = "ipaddresspools"

// IPAddressPoolBuilder provides struct for the IPAddressPool object containing connection to
// the cluster and the IPAddressPool definitions.
type IPAddressPoolBuilder struct {
	common.EmbeddableBuilder[mlbtypes.IPAddressPool, *mlbtypes.IPAddressPool]
	common.EmbeddableWithOptions[
		mlbtypes.IPAddressPool, IPAddressPoolBuilder, *mlbtypes.IPAddressPool, *IPAddressPoolBuilder,
		IPAddressPoolAdditionalOptions]
	common.EmbeddableCreator[mlbtypes.IPAddressPool, IPAddressPoolBuilder, *mlbtypes.IPAddressPool, *IPAddressPoolBuilder]
	common.EmbeddableDeleteReturner[mlbtypes.IPAddressPool, IPAddressPoolBuilder, *mlbtypes.IPAddressPool, *IPAddressPoolBuilder]
	common.EmbeddableForceUpdater[mlbtypes.IPAddressPool, IPAddressPoolBuilder, *mlbtypes.IPAddressPool, *IPAddressPoolBuilder]
}

// IPAddressPoolAdditionalOptions additional options for IPAddressPool object.
type IPAddressPoolAdditionalOptions func(builder *IPAddressPoolBuilder) (*IPAddressPoolBuilder, error)

// AttachMixins wires the embedded CRUD and option mixins to this builder instance.
func (builder *IPAddressPoolBuilder) AttachMixins() {
	builder.EmbeddableWithOptions.SetBase(builder)
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the IPAddressPool GVK for this builder.
func (builder *IPAddressPoolBuilder) GetGVK() schema.GroupVersionKind {
	return mlbtypes.GroupVersion.WithKind("IPAddressPool")
}

// NewIPAddressPoolBuilder creates a new instance of IPAddressPoolBuilder.
func NewIPAddressPoolBuilder(
	apiClient *clients.Settings, name, nsname string, addrPool []string) *IPAddressPoolBuilder {
	klog.V(100).Infof(
		"Initializing new IPAddressPool structure with the following params: %s, %s %s",
		name, nsname, addrPool)

	builder := common.NewNamespacedBuilder[mlbtypes.IPAddressPool, IPAddressPoolBuilder](
		apiClient, mlbtypes.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	if len(addrPool) < 1 {
		klog.V(100).Info("The addrPool of the IPAddressPool is empty list")

		builder.SetError(fmt.Errorf("IPAddressPool 'addrPool' cannot be empty list"))

		return builder
	}

	builder.Definition.Spec.Addresses = addrPool

	return builder
}

// PullAddressPool pulls existing addresspool from cluster.
func PullAddressPool(apiClient *clients.Settings, name, nsname string) (*IPAddressPoolBuilder, error) {
	return common.PullNamespacedBuilder[mlbtypes.IPAddressPool, IPAddressPoolBuilder](
		context.TODO(), apiClient, mlbtypes.AddToScheme, name, nsname)
}

// WithAutoAssign defines the AutoAssign bool flag placed in the IPAddressPool spec.
func (builder *IPAddressPoolBuilder) WithAutoAssign(autoAssign bool) *IPAddressPoolBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating IPAddressPool %s in namespace %s with this autoAssign flag: %t",
		builder.Definition.Name, builder.Definition.Namespace, autoAssign)

	builder.Definition.Spec.AutoAssign = &autoAssign

	return builder
}

// WithAvoidBuggyIPs defines the AvoidBuggyIPs bool flag placed in the IPAddressPool spec.
func (builder *IPAddressPoolBuilder) WithAvoidBuggyIPs(avoidBuggyIPs bool) *IPAddressPoolBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating IPAddressPool %s in namespace %s with this avoidBuggyIPs flag: %t",
		builder.Definition.Name, builder.Definition.Namespace, avoidBuggyIPs)

	builder.Definition.Spec.AvoidBuggyIPs = avoidBuggyIPs

	return builder
}

// GetIPAddressPoolGVR returns ipaddresspool's GroupVersionResource, which could be used for Clean function.
func GetIPAddressPoolGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: resourceIPAddressPools,
	}
}
