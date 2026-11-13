package sriov

import (
	"context"
	"fmt"
	"slices"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errInvalidDeviceType = "invalid device type, allowed devType values are: vfio-pci or netdevice"
)

var policyGVK = srIovV1.GroupVersion.WithKind("SriovNetworkNodePolicy")

// PolicyBuilder provides struct for srIovPolicy object containing connection to the cluster and the srIovPolicy
// definitions.
type PolicyBuilder struct {
	common.EmbeddableBuilder[srIovV1.SriovNetworkNodePolicy, *srIovV1.SriovNetworkNodePolicy]
	common.EmbeddableCreator[srIovV1.SriovNetworkNodePolicy, PolicyBuilder,
		*srIovV1.SriovNetworkNodePolicy, *PolicyBuilder]
	common.EmbeddableDeleter[srIovV1.SriovNetworkNodePolicy, *srIovV1.SriovNetworkNodePolicy]
	common.EmbeddableForceUpdater[srIovV1.SriovNetworkNodePolicy, PolicyBuilder,
		*srIovV1.SriovNetworkNodePolicy, *PolicyBuilder]
	common.EmbeddableWithOptions[srIovV1.SriovNetworkNodePolicy, PolicyBuilder,
		*srIovV1.SriovNetworkNodePolicy, *PolicyBuilder, PolicyAdditionalOptions]
}

// PolicyAdditionalOptions additional options for SriovNetworkNodePolicy object.
type PolicyAdditionalOptions func(builder *PolicyBuilder) (*PolicyBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *PolicyBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the SriovNetworkNodePolicy GVK for this builder.
func (builder *PolicyBuilder) GetGVK() schema.GroupVersionKind {
	return policyGVK
}

// NewPolicyBuilder creates a new instance of PolicyBuilder.
func NewPolicyBuilder(
	apiClient *clients.Settings,
	name string,
	nsname string,
	resName string,
	vfsNumber int,
	nicNames []string,
	nodeSelector map[string]string) *PolicyBuilder {
	builder := common.NewNamespacedBuilder[srIovV1.SriovNetworkNodePolicy, PolicyBuilder](
		apiClient, srIovV1.AddToScheme, name, nsname)

	builder.Definition.Spec = srIovV1.SriovNetworkNodePolicySpec{
		NodeSelector: nodeSelector,
		NumVfs:       vfsNumber,
		ResourceName: resName,
		Priority:     1,
		NicSelector: srIovV1.SriovNetworkNicSelector{
			PfNames: nicNames,
		},
	}

	if builder.GetError() != nil {
		return builder
	}

	if resName == "" {
		builder.SetError(fmt.Errorf("SriovNetworkNodePolicy 'resName' cannot be empty"))

		return builder
	}

	if len(nicNames) == 0 {
		builder.SetError(fmt.Errorf("SriovNetworkNodePolicy 'nicNames' cannot be empty list"))

		return builder
	}

	if len(nodeSelector) == 0 {
		builder.SetError(fmt.Errorf("SriovNetworkNodePolicy 'nodeSelector' cannot be empty map"))

		return builder
	}

	if vfsNumber <= 0 {
		builder.SetError(fmt.Errorf("SriovNetworkNodePolicy 'vfsNumber' cannot be zero of negative"))

		return builder
	}

	return builder
}

// WithDevType sets device type in the SriovNetworkNodePolicy definition. Allowed devTypes are vfio-pci and netdevice.
func (builder *PolicyBuilder) WithDevType(devType string) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	allowedDevTypes := []string{"vfio-pci", "netdevice"}

	if !slices.Contains(allowedDevTypes, devType) {
		builder.SetError(fmt.Errorf(errInvalidDeviceType))

		return builder
	}

	builder.Definition.Spec.DeviceType = devType

	return builder
}

// WithVFRange sets specific VF range for each configured PF.
func (builder *PolicyBuilder) WithVFRange(firstVF, lastVF int) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if firstVF < 0 || lastVF < 0 {
		builder.SetError(fmt.Errorf("firstPF or lastVF can not be less than 0"))

		return builder
	}

	if firstVF > lastVF {
		builder.SetError(fmt.Errorf("firstPF argument can not be greater than lastPF"))

		return builder
	}

	if lastVF > 63 {
		builder.SetError(fmt.Errorf("lastVF can not be greater than 63"))

		return builder
	}

	var partitionedPFs []string
	for _, pf := range builder.Definition.Spec.NicSelector.PfNames {
		partitionedPFs = append(partitionedPFs, fmt.Sprintf("%s#%d-%d", pf, firstVF, lastVF))
	}

	builder.Definition.Spec.NicSelector.PfNames = partitionedPFs

	return builder
}

// WithMTU sets required MTU in the given SriovNetworkNodePolicy.
func (builder *PolicyBuilder) WithMTU(mtu int) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if 1 > mtu || mtu > 9192 {
		builder.SetError(fmt.Errorf("invalid mtu size %d allowed mtu should be in range 1...9192", mtu))

		return builder
	}

	builder.Definition.Spec.Mtu = mtu

	return builder
}

// WithRDMA sets RDMA mode in SriovNetworkNodePolicy object.
func (builder *PolicyBuilder) WithRDMA(rdma bool) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	builder.Definition.Spec.IsRdma = rdma

	return builder
}

// WithVhostNet sets Vhost mode in SriovNetworkNodePolicy object.
func (builder *PolicyBuilder) WithVhostNet(vhost bool) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Redefining SriovNetworkNodePolicy %s with NeedVhostNet: %t",
		builder.Definition.Name, vhost)

	builder.Definition.Spec.NeedVhostNet = vhost

	return builder
}

// WithExternallyManaged sets ExternallyManaged option in SriovNetworkNodePolicy object.
func (builder *PolicyBuilder) WithExternallyManaged(externallyManaged bool) *PolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Redefining SriovNetworkNodePolicy %s with externallyManaged: %t",
		builder.Definition.Name, externallyManaged)

	builder.Definition.Spec.ExternallyManaged = externallyManaged

	return builder
}

// PullPolicy pulls existing sriovnetworknodepolicy from cluster.
func PullPolicy(apiClient *clients.Settings, name, nsname string) (*PolicyBuilder, error) {
	return common.PullNamespacedBuilder[srIovV1.SriovNetworkNodePolicy, PolicyBuilder](
		context.TODO(), apiClient, srIovV1.AddToScheme, name, nsname)
}
