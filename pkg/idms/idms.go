package idms

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var idmsGVK = configv1.GroupVersion.WithKind("ImageDigestMirrorSet")

// Builder provides struct for an ImageDigestMirrorSet containing its cluster connection and definition.
type Builder struct {
	common.EmbeddableBuilder[configv1.ImageDigestMirrorSet, *configv1.ImageDigestMirrorSet]
	common.EmbeddableCreator[
		configv1.ImageDigestMirrorSet, Builder, *configv1.ImageDigestMirrorSet, *Builder]
	common.EmbeddableDeleter[configv1.ImageDigestMirrorSet, *configv1.ImageDigestMirrorSet]
	common.EmbeddableForceUpdater[
		configv1.ImageDigestMirrorSet, Builder, *configv1.ImageDigestMirrorSet, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the ImageDigestMirrorSet GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return idmsGVK
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(apiClient *clients.Settings, name string, mirror configv1.ImageDigestMirrors) *Builder {
	klog.V(100).Infof(
		"Initializing new imagedigestmirrorset structure with the following params: "+
			"name: %s, mirror: %v", name, mirror)

	builder := common.NewClusterScopedBuilder[configv1.ImageDigestMirrorSet, Builder](
		apiClient, configv1.AddToScheme, name)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Spec.ImageDigestMirrors = []configv1.ImageDigestMirrors{mirror}

	return builder
}

// Pull retrieves an existing imagedigestmirrorset from the cluster.
func Pull(apiClient *clients.Settings, name string) (*Builder, error) {
	return common.PullClusterScopedBuilder[configv1.ImageDigestMirrorSet, Builder](
		context.TODO(), apiClient, configv1.AddToScheme, name)
}

// WithMirror adds an imagedigestmirror for mirroring images.
func (builder *Builder) WithMirror(mirror configv1.ImageDigestMirrors) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding imagedigestmirror to imagedigestmirrorset %s: %v",
		builder.Definition.Name, mirror)

	builder.Definition.Spec.ImageDigestMirrors = append(builder.Definition.Spec.ImageDigestMirrors, mirror)

	return builder
}
