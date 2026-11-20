package ovn

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListRouteAdvertisements returns RouteAdvertisement inventory (cluster-scoped).
func ListRouteAdvertisements(apiClient *clients.Settings,
	options ...runtimeClient.ListOption) ([]*RouteAdvertisementBuilder, error) {
	return common.List[ovnv1.RouteAdvertisements, ovnv1.RouteAdvertisementsList, RouteAdvertisementBuilder](
		context.TODO(), apiClient, ovnv1.AddToScheme, options...)
}
