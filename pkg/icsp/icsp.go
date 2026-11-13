package icsp

import (
	"context"
	"fmt"

	v1alpha1 "github.com/openshift/api/operator/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errEmptySource  = "imageContentSourcePolicy 'source' cannot be empty"
	errEmptyMirrors = "imageContentSourcePolicy 'mirrors' cannot be empty"
)

var icspGVK = v1alpha1.GroupVersion.WithKind("ImageContentSourcePolicy")

// ICSPBuilder provides struct for the ImageContentSourcePolicy object with connection to the cluster.
type ICSPBuilder struct {
	common.EmbeddableBuilder[v1alpha1.ImageContentSourcePolicy, *v1alpha1.ImageContentSourcePolicy]
	common.EmbeddableCreator[
		v1alpha1.ImageContentSourcePolicy, ICSPBuilder, *v1alpha1.ImageContentSourcePolicy, *ICSPBuilder]
	common.EmbeddableDeleter[v1alpha1.ImageContentSourcePolicy, *v1alpha1.ImageContentSourcePolicy]
	common.EmbeddableUpdater[
		v1alpha1.ImageContentSourcePolicy, ICSPBuilder, *v1alpha1.ImageContentSourcePolicy, *ICSPBuilder]
	common.EmbeddableWithOptions[
		v1alpha1.ImageContentSourcePolicy, ICSPBuilder, *v1alpha1.ImageContentSourcePolicy, *ICSPBuilder,
		AdditionalOptions]
}

// AdditionalOptions additional options for ImageContentSourcePolicy object.
type AdditionalOptions func(builder *ICSPBuilder) (*ICSPBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ICSPBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the ImageContentSourcePolicy GVK for this builder.
func (builder *ICSPBuilder) GetGVK() schema.GroupVersionKind {
	return icspGVK
}

// NewICSPBuilder creates a new instance of ICSPBuilder.
func NewICSPBuilder(apiClient *clients.Settings, name, source string, mirrors []string) *ICSPBuilder {
	klog.V(100).Infof(
		"Initializing new ICSPBuilder structure with the following params: "+
			"name: %s, source: %s, mirrors: %v",
		name, source, mirrors)

	builder := common.NewClusterScopedBuilder[v1alpha1.ImageContentSourcePolicy, ICSPBuilder](
		apiClient, v1alpha1.Install, name)
	if builder.GetError() != nil {
		return builder
	}

	if source == "" {
		klog.V(100).Info("The Source of the ImageContentSourcePolicy is empty")

		builder.SetError(fmt.Errorf(errEmptySource))

		return builder
	}

	if len(mirrors) == 0 {
		klog.V(100).Info("The mirrors of the ImageContentSourcePolicy are empty")

		builder.SetError(fmt.Errorf(errEmptyMirrors))

		return builder
	}

	builder.Definition.Spec.RepositoryDigestMirrors = []v1alpha1.RepositoryDigestMirrors{{
		Source:  source,
		Mirrors: mirrors,
	}}

	return builder
}

// Pull pulls object definition from cluster to ICSPBuilder struct.
func Pull(apiClient *clients.Settings, name string) (*ICSPBuilder, error) {
	return common.PullClusterScopedBuilder[v1alpha1.ImageContentSourcePolicy, ICSPBuilder](
		context.TODO(), apiClient, v1alpha1.Install, name)
}

// WithRepositoryDigestMirror adds new RepositoryDigestMirror.
func (builder *ICSPBuilder) WithRepositoryDigestMirror(source string, mirrors []string) *ICSPBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding RepositoryDigestMirror with source %s and mirrors %v to ImageContentSourcePolicy %s",
		source, mirrors, builder.Definition.Name)

	if source == "" {
		klog.V(100).Info("The source is empty")

		builder.SetError(fmt.Errorf(errEmptySource))

		return builder
	}

	if len(mirrors) == 0 {
		klog.V(100).Info("Mirrors is empty")

		builder.SetError(fmt.Errorf(errEmptyMirrors))

		return builder
	}

	builder.Definition.Spec.RepositoryDigestMirrors = append(
		builder.Definition.Spec.RepositoryDigestMirrors, v1alpha1.RepositoryDigestMirrors{Source: source, Mirrors: mirrors})

	return builder
}
