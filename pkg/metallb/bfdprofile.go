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

var bfdProfileGVK = mlbtypes.GroupVersion.WithKind("BFDProfile")

// BFDBuilder provides struct for the BFDProfile object containing connection to
// the cluster and the BFDProfile definitions.
type BFDBuilder struct {
	common.EmbeddableBuilder[mlbtypes.BFDProfile, *mlbtypes.BFDProfile]
	common.EmbeddableCreator[mlbtypes.BFDProfile, BFDBuilder, *mlbtypes.BFDProfile, *BFDBuilder]
	common.EmbeddableDeleteReturner[mlbtypes.BFDProfile, BFDBuilder, *mlbtypes.BFDProfile, *BFDBuilder]
	common.EmbeddableForceUpdater[mlbtypes.BFDProfile, BFDBuilder, *mlbtypes.BFDProfile, *BFDBuilder]
	common.EmbeddableWithOptions[
		mlbtypes.BFDProfile, BFDBuilder, *mlbtypes.BFDProfile, *BFDBuilder, BFDAdditionalOptions]
}

// BFDAdditionalOptions additional options for BFDProfile object.
type BFDAdditionalOptions func(builder *BFDBuilder) (*BFDBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *BFDBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the BFDProfile GVK for this builder.
func (builder *BFDBuilder) GetGVK() schema.GroupVersionKind {
	return bfdProfileGVK
}

// NewBFDBuilder creates a new instance of BFDBuilder.
func NewBFDBuilder(apiClient *clients.Settings, name, nsname string) *BFDBuilder {
	return common.NewNamespacedBuilder[mlbtypes.BFDProfile, BFDBuilder](
		apiClient, mlbtypes.AddToScheme, name, nsname)
}

// PullBFDProfile pulls existing bfdprofile from cluster.
func PullBFDProfile(apiClient *clients.Settings, name, nsname string) (*BFDBuilder, error) {
	return common.PullNamespacedBuilder[mlbtypes.BFDProfile, BFDBuilder](
		context.TODO(), apiClient, mlbtypes.AddToScheme, name, nsname)
}

// WithRcvInterval defines the receiveInterval placed in the BFDProfile.
func (builder *BFDBuilder) WithRcvInterval(rcvInterval uint32) *BFDBuilder {
	return builder.withInterval("receiveInterval", rcvInterval)
}

// WithTransmitInterval defines the transmitInterval placed in the BFDProfile.
func (builder *BFDBuilder) WithTransmitInterval(transmitInterval uint32) *BFDBuilder {
	return builder.withInterval("transmitInterval", transmitInterval)
}

// WithEchoInterval defines the ecoInterval placed in the BFDProfile.
func (builder *BFDBuilder) WithEchoInterval(ecoInterval uint32) *BFDBuilder {
	return builder.withInterval("ecoInterval", ecoInterval)
}

// WithMultiplier defines the detectMultiplier placed in the BFDProfile.
func (builder *BFDBuilder) WithMultiplier(multiplier uint32) *BFDBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BFDProfile %s in namespace %s with this detectMultiplier: %d",
		builder.Definition.Name, builder.Definition.Namespace, multiplier)

	builder.Definition.Spec.DetectMultiplier = &multiplier

	return builder
}

// WithEchoMode defines the echoMode placed in the BFDProfile.
func (builder *BFDBuilder) WithEchoMode(echoMode bool) *BFDBuilder {
	return builder.withBoolFlagFor("echoMode", echoMode)
}

// WithPassiveMode defines the passiveMode placed in the BFDProfile.
func (builder *BFDBuilder) WithPassiveMode(passiveMode bool) *BFDBuilder {
	return builder.withBoolFlagFor("passiveMode", passiveMode)
}

// WithMinimumTTL defines the minimumTTTL placed in the BFDProfile.
func (builder *BFDBuilder) WithMinimumTTL(minimumTTL uint32) *BFDBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BFDProfile %s in namespace %s with this minimumTTL: %d",
		builder.Definition.Name, builder.Definition.Namespace, minimumTTL)

	builder.Definition.Spec.MinimumTTL = &minimumTTL

	return builder
}

// GetBFDProfileGVR returns bfdprofile's GroupVersionResource which could be used for Clean function.
func GetBFDProfileGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "bfdprofiles",
	}
}

// withBoolFlagFor sets a boolean BFDProfile spec field selected by name.
func (builder *BFDBuilder) withBoolFlagFor(flagName string, flagValue bool) *BFDBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BFDProfile %s in namespace %s with flag %s: %t",
		builder.Definition.Name, builder.Definition.Namespace, flagName, flagValue)

	switch flagName {
	case "echoMode":
		builder.Definition.Spec.EchoMode = &flagValue
	case "passiveMode":
		builder.Definition.Spec.PassiveMode = &flagValue
	default:
		builder.SetError(fmt.Errorf("invalid bool flag name parameter"))
	}

	return builder
}

// withInterval sets an interval BFDProfile spec field selected by name.
func (builder *BFDBuilder) withInterval(intervalName string, interval uint32) *BFDBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating BFDProfile %s in namespace %s with interval %s: %d",
		builder.Definition.Name, builder.Definition.Namespace, intervalName, interval)

	switch intervalName {
	case "transmitInterval":
		builder.Definition.Spec.TransmitInterval = &interval
	case "receiveInterval":
		builder.Definition.Spec.ReceiveInterval = &interval
	case "ecoInterval":
		builder.Definition.Spec.EchoInterval = &interval
	default:
		builder.SetError(fmt.Errorf("invalid interval parameters"))
	}

	return builder
}
