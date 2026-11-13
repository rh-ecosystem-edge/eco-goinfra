package metallb

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var frrNodeStateGVK = frrtypes.GroupVersion.WithKind("FRRNodeState")

// FrrNodeStateBuilder provides struct for FrrNodeState object which contains connection to cluster and
// frrconfiguration definitions.
type FrrNodeStateBuilder struct {
	common.EmbeddableBuilder[frrtypes.FRRNodeState, *frrtypes.FRRNodeState]
}

// GetGVK returns the FRRNodeState GVK for this builder.
func (builder *FrrNodeStateBuilder) GetGVK() schema.GroupVersionKind {
	return frrNodeStateGVK
}

// PullFrrNodeState retrieves an existing FrrNodeState object from the cluster.
func PullFrrNodeState(apiClient *clients.Settings, name string) (*FrrNodeStateBuilder, error) {
	return common.PullClusterScopedBuilder[frrtypes.FRRNodeState, FrrNodeStateBuilder](
		context.TODO(), apiClient, frrtypes.AddToScheme, name)
}
