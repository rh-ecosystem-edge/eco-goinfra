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

// PreflightValidationOCPBuilder provides struct for the preflightvalidationocp object
// containing connection to the cluster and the preflightvalidationocp definitions.
type PreflightValidationOCPBuilder struct {
	common.EmbeddableBuilder[kmmv1beta2.PreflightValidationOCP, *kmmv1beta2.PreflightValidationOCP]
	common.EmbeddableCreator[kmmv1beta2.PreflightValidationOCP, PreflightValidationOCPBuilder,
		*kmmv1beta2.PreflightValidationOCP, *PreflightValidationOCPBuilder]
	common.EmbeddableDeleteReturner[kmmv1beta2.PreflightValidationOCP, PreflightValidationOCPBuilder,
		*kmmv1beta2.PreflightValidationOCP, *PreflightValidationOCPBuilder]
	common.EmbeddableUpdater[kmmv1beta2.PreflightValidationOCP, PreflightValidationOCPBuilder,
		*kmmv1beta2.PreflightValidationOCP, *PreflightValidationOCPBuilder]
	common.EmbeddableWithOptions[kmmv1beta2.PreflightValidationOCP, PreflightValidationOCPBuilder,
		*kmmv1beta2.PreflightValidationOCP, *PreflightValidationOCPBuilder, PreflightValidationOCPAdditionalOptions]
}

// PreflightValidationOCPAdditionalOptions additional options for preflightvalidationocp object.
type PreflightValidationOCPAdditionalOptions func(
	builder *PreflightValidationOCPBuilder) (*PreflightValidationOCPBuilder, error)

// AttachMixins wires the embedded CRUD and option mixins to this builder instance.
func (builder *PreflightValidationOCPBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the PreflightValidationOCP GVK for this builder.
func (builder *PreflightValidationOCPBuilder) GetGVK() schema.GroupVersionKind {
	return kmmv1beta2.GroupVersion.WithKind("PreflightValidationOCP")
}

// NewPreflightValidationOCPBuilder creates a new instance of PreflightValidationOCPBuilder.
func NewPreflightValidationOCPBuilder(
	apiClient *clients.Settings, name, nsname string) *PreflightValidationOCPBuilder {
	return common.NewNamespacedBuilder[kmmv1beta2.PreflightValidationOCP, PreflightValidationOCPBuilder](
		apiClient, kmmv1beta2.AddToScheme, name, nsname)
}

// WithKernelVersion sets the image for which the preflightvalidationocp checks the module.
func (builder *PreflightValidationOCPBuilder) WithKernelVersion(kernelVersion string) *PreflightValidationOCPBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if kernelVersion == "" {
		builder.SetError(fmt.Errorf("%s", errNilKernelVersion))

		return builder
	}

	klog.V(100).Infof("Creating new PreflightValidationOCP with kernelVersion: %s", kernelVersion)

	builder.Definition.Spec.KernelVersion = kernelVersion

	return builder
}

// WithDtkImage specifies if the kernel is realtime.
func (builder *PreflightValidationOCPBuilder) WithDtkImage(dtkImage string) *PreflightValidationOCPBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if dtkImage == "" {
		builder.SetError(fmt.Errorf("invalid 'dtkImage' argument can not be nil"))

		return builder
	}

	klog.V(100).Infof("Creating new PreflightValidationOCP with dtkImage set to: %s", dtkImage)

	builder.Definition.Spec.DTKImage = dtkImage

	return builder
}

// WithPushBuiltImage configures the build to be pushed to the registry.
func (builder *PreflightValidationOCPBuilder) WithPushBuiltImage(push bool) *PreflightValidationOCPBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating new PreflightValidationOCP with PushBuiltImage set to: %v", push)

	builder.Definition.Spec.PushBuiltImage = push

	return builder
}

// PullPreflightValidationOCP fetches existing PreflightValidationOCP from the cluster.
func PullPreflightValidationOCP(apiClient *clients.Settings,
	name, nsname string) (*PreflightValidationOCPBuilder, error) {
	return common.PullNamespacedBuilder[kmmv1beta2.PreflightValidationOCP, PreflightValidationOCPBuilder](
		context.TODO(), apiClient, kmmv1beta2.AddToScheme, name, nsname)
}
