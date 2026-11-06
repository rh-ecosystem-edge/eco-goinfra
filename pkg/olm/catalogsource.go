package olm

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	oplmV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CatalogSourceBuilder provides a struct for catalogsource object
// from the cluster and a catalogsource definition.
type CatalogSourceBuilder struct {
	common.EmbeddableBuilder[oplmV1alpha1.CatalogSource, *oplmV1alpha1.CatalogSource]
	common.EmbeddableCreator[oplmV1alpha1.CatalogSource, CatalogSourceBuilder, *oplmV1alpha1.CatalogSource, *CatalogSourceBuilder]
	common.EmbeddableDeleter[oplmV1alpha1.CatalogSource, *oplmV1alpha1.CatalogSource]
	common.EmbeddableForceUpdater[
		oplmV1alpha1.CatalogSource, CatalogSourceBuilder, *oplmV1alpha1.CatalogSource, *CatalogSourceBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *CatalogSourceBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the CatalogSource GVK for this builder.
func (builder *CatalogSourceBuilder) GetGVK() schema.GroupVersionKind {
	return oplmV1alpha1.SchemeGroupVersion.WithKind("CatalogSource")
}

// NewCatalogSourceBuilder creates new instance of CatalogSourceBuilder.
func NewCatalogSourceBuilder(apiClient *clients.Settings, name, nsname string) *CatalogSourceBuilder {
	return common.NewNamespacedBuilder[oplmV1alpha1.CatalogSource, CatalogSourceBuilder](
		apiClient, oplmV1alpha1.AddToScheme, name, nsname)
}

// PullCatalogSource loads an existing catalogsource into Builder struct.
func PullCatalogSource(apiClient *clients.Settings, name, nsname string) (*CatalogSourceBuilder, error) {
	return common.PullNamespacedBuilder[oplmV1alpha1.CatalogSource, CatalogSourceBuilder](
		context.TODO(), apiClient, oplmV1alpha1.AddToScheme, name, nsname)
}
