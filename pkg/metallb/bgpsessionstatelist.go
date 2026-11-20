package metallb

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListBGPSessionState returns BGP session state inventory in the given cluster.
func ListBGPSessionState(
	apiClient *clients.Settings, options ...client.ListOptions) ([]*BGPSessionStateBuilder, error) {
	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[frrtypes.BGPSessionState, frrtypes.BGPSessionStateList, BGPSessionStateBuilder](
		context.TODO(), apiClient, frrtypes.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
