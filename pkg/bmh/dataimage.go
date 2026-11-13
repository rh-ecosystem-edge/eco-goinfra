package bmh

import (
	"context"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DataImageBuilder provides struct for the dataimage object containing connection to
// the cluster and the dataimage definitions.
type DataImageBuilder struct {
	common.EmbeddableBuilder[bmhv1alpha1.DataImage, *bmhv1alpha1.DataImage]
	common.EmbeddableDeleteReturner[bmhv1alpha1.DataImage, DataImageBuilder, *bmhv1alpha1.DataImage, *DataImageBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *DataImageBuilder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the DataImage GVK for this builder.
func (builder *DataImageBuilder) GetGVK() schema.GroupVersionKind {
	return bmhv1alpha1.GroupVersion.WithKind("DataImage")
}

// PullDataImage retrieves an existing DataImage resource from the cluster.
func PullDataImage(apiClient *clients.Settings, name, nsname string) (*DataImageBuilder, error) {
	return common.PullNamespacedBuilder[bmhv1alpha1.DataImage, DataImageBuilder](
		context.TODO(), apiClient, bmhv1alpha1.AddToScheme, name, nsname)
}
