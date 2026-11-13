package hive

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	hiveV1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/hive/api/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var clusterImageSetGVK = hiveV1.SchemeGroupVersion.WithKind("ClusterImageSet")

// ClusterImageSetBuilder provides struct for the clusterimageset object containing connection to
// the cluster and the clusterimageset definitions.
type ClusterImageSetBuilder struct {
	common.EmbeddableBuilder[hiveV1.ClusterImageSet, *hiveV1.ClusterImageSet]
	common.EmbeddableCreator[hiveV1.ClusterImageSet, ClusterImageSetBuilder, *hiveV1.ClusterImageSet, *ClusterImageSetBuilder]
	common.EmbeddableDeleter[hiveV1.ClusterImageSet, *hiveV1.ClusterImageSet]
	common.EmbeddableForceUpdater[
		hiveV1.ClusterImageSet, ClusterImageSetBuilder, *hiveV1.ClusterImageSet, *ClusterImageSetBuilder]
	common.EmbeddableWithOptions[
		hiveV1.ClusterImageSet,
		ClusterImageSetBuilder,
		*hiveV1.ClusterImageSet,
		*ClusterImageSetBuilder,
		ClusterImageSetAdditionalOptions,
	]
}

// ClusterImageSetAdditionalOptions additional options for ClusterImageSet object.
type ClusterImageSetAdditionalOptions func(builder *ClusterImageSetBuilder) (*ClusterImageSetBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ClusterImageSetBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the ClusterImageSet GVK for this builder.
func (builder *ClusterImageSetBuilder) GetGVK() schema.GroupVersionKind {
	return clusterImageSetGVK
}

// NewClusterImageSetBuilder creates a new instance of ClusterImageSetBuilder.
func NewClusterImageSetBuilder(apiClient *clients.Settings, name, releaseImage string) *ClusterImageSetBuilder {
	builder := common.NewClusterScopedBuilder[hiveV1.ClusterImageSet, ClusterImageSetBuilder](
		apiClient, hiveV1.AddToScheme, name)
	if builder.GetError() != nil {
		return builder
	}

	if releaseImage == "" {
		klog.V(100).Info("The releaseImage of the clusterimageset is empty")

		builder.SetError(fmt.Errorf("clusterimageset 'releaseImage' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.ReleaseImage = releaseImage

	return builder
}

// PullClusterImageSet loads an existing clusterimageset into ClusterImageSetBuilder struct.
func PullClusterImageSet(apiClient *clients.Settings, name string) (*ClusterImageSetBuilder, error) {
	return common.PullClusterScopedBuilder[hiveV1.ClusterImageSet, ClusterImageSetBuilder](
		context.TODO(), apiClient, hiveV1.AddToScheme, name)
}

// WithReleaseImage sets the releaseImage for the clusterimageset.
func (builder *ClusterImageSetBuilder) WithReleaseImage(image string) *ClusterImageSetBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting clusterimageset %s releaseImage to %s", builder.Definition.Name, image)

	if image == "" {
		klog.V(100).Info("The clusterimageset releaseImage is empty")

		builder.SetError(fmt.Errorf("cannot set releaseImage to empty string"))

		return builder
	}

	builder.Definition.Spec.ReleaseImage = image

	return builder
}
