package nmstate

import (
	"context"

	nmstateV1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var nmStateGVK = nmstateV1.GroupVersion.WithKind("NMState")

// Builder provides struct for the NMState object containing connection to
// the cluster and the NMState definitions.
type Builder struct {
	common.EmbeddableBuilder[nmstateV1.NMState, *nmstateV1.NMState]
	common.EmbeddableCreator[nmstateV1.NMState, Builder, *nmstateV1.NMState, *Builder]
	common.EmbeddableDeleteReturner[nmstateV1.NMState, Builder, *nmstateV1.NMState, *Builder]
	common.EmbeddableForceUpdater[nmstateV1.NMState, Builder, *nmstateV1.NMState, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the NMState GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return nmStateGVK
}

// NewBuilder creates a new instance of nmstate Builder.
func NewBuilder(apiClient *clients.Settings, name string) *Builder {
	return common.NewClusterScopedBuilder[nmstateV1.NMState, Builder](
		apiClient, nmstateV1.AddToScheme, name)
}

// PullNMstate retrieves an existing NMState object from the cluster.
func PullNMstate(apiClient *clients.Settings, name string) (*Builder, error) {
	return common.PullClusterScopedBuilder[nmstateV1.NMState, Builder](
		context.TODO(), apiClient, nmstateV1.AddToScheme, name)
}
