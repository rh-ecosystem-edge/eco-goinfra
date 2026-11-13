package kmm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	kmmv1beta2 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta2"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errNilKernelVersion = "invalid 'kernelVersion' argument can not be nil"
)

var preflightValidationGVK = kmmv1beta2.GroupVersion.WithKind("PreflightValidation")

// PreflightValidationBuilder provides struct for the preflightvalidation object
// containing connection to the cluster and the preflightvalidation definitions.
type PreflightValidationBuilder struct {
	common.EmbeddableBuilder[kmmv1beta2.PreflightValidation, *kmmv1beta2.PreflightValidation]
	common.EmbeddableCreator[kmmv1beta2.PreflightValidation, PreflightValidationBuilder,
		*kmmv1beta2.PreflightValidation, *PreflightValidationBuilder]
	common.EmbeddableDeleteReturner[kmmv1beta2.PreflightValidation, PreflightValidationBuilder,
		*kmmv1beta2.PreflightValidation, *PreflightValidationBuilder]
	common.EmbeddableUpdater[kmmv1beta2.PreflightValidation, PreflightValidationBuilder,
		*kmmv1beta2.PreflightValidation, *PreflightValidationBuilder]
	common.EmbeddableWithOptions[kmmv1beta2.PreflightValidation, PreflightValidationBuilder,
		*kmmv1beta2.PreflightValidation, *PreflightValidationBuilder, PreflightValidationAdditionalOptions]
}

// PreflightValidationAdditionalOptions additional options for preflightvalidation object.
type PreflightValidationAdditionalOptions func(
	builder *PreflightValidationBuilder) (*PreflightValidationBuilder, error)

// AttachMixins wires the embedded CRUD and option mixins to this builder instance.
func (builder *PreflightValidationBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the PreflightValidation GVK for this builder.
func (builder *PreflightValidationBuilder) GetGVK() schema.GroupVersionKind {
	return preflightValidationGVK
}

// NewPreflightValidationBuilder creates a new instance of PreflightValidationBuilder.
func NewPreflightValidationBuilder(
	apiClient *clients.Settings, name, nsname string) *PreflightValidationBuilder {
	return common.NewNamespacedBuilder[kmmv1beta2.PreflightValidation, PreflightValidationBuilder](
		apiClient, kmmv1beta2.AddToScheme, name, nsname)
}

// WithKernelVersion sets the kernel for which the preflightvalidation checks the module.
func (builder *PreflightValidationBuilder) WithKernelVersion(kernelVersion string) *PreflightValidationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if kernelVersion == "" {
		builder.SetError(fmt.Errorf("%s", errNilKernelVersion))

		return builder
	}

	klog.V(100).Infof("Creating new PreflightValidation with kernelVersion: %s", kernelVersion)

	builder.Definition.Spec.KernelVersion = kernelVersion

	return builder
}

// WithPushBuiltImage configures the build to be pushed to the registry.
func (builder *PreflightValidationBuilder) WithPushBuiltImage(push bool) *PreflightValidationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating new PreflightValidation with PushBuiltImage set to: %v", push)

	builder.Definition.Spec.PushBuiltImage = push

	return builder
}

// PullPreflightValidation fetches existing PreflightValidation from the cluster.
func PullPreflightValidation(apiClient *clients.Settings,
	name, nsname string) (*PreflightValidationBuilder, error) {
	return common.PullNamespacedBuilder[kmmv1beta2.PreflightValidation, PreflightValidationBuilder](
		context.TODO(), apiClient, kmmv1beta2.AddToScheme, name, nsname)
}
