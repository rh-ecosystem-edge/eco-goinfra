package ovn

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListRouteAdvertisements returns RouteAdvertisement inventory in the given namespace.
func ListRouteAdvertisements(apiClient *clients.Settings, nsname string, options ...runtimeClient.ListOptions) ([]*RouteAdvertisementBuilder, error) {
	if apiClient == nil {
		glog.V(100).Infof("RouteAdvertisements 'apiClient' parameter can not be empty")

		return nil, fmt.Errorf("failed to list RouteAdvertisements, 'apiClient' parameter is empty")
	}

	err := apiClient.AttachScheme(ovnv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add ovn scheme to client schemes")

		return nil, err
	}

	logMessage := fmt.Sprintf("Listing RouteAdvertisements in the namespace %s", nsname)
	passedOptions := runtimeClient.ListOptions{}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	} else if len(options) > 1 {
		glog.V(100).Infof("error: more than one ListOptions was passed")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	glog.V(100).Infof(logMessage)

	routeAdvertisementList := &ovnv1.RouteAdvertisementList{}
	passedOptions.Namespace = nsname

	err = apiClient.Client.List(context.TODO(), routeAdvertisementList, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list RouteAdvertisements in the namespace %s due to %s", nsname, err.Error())

		return nil, err
	}

	var routeAdvertisementObjects []*RouteAdvertisementBuilder

	for _, routeAdvertisement := range routeAdvertisementList.Items {
		copiedRouteAdvertisement := routeAdvertisement
		routeAdvertisementBuilder := &RouteAdvertisementBuilder{
			apiClient:  apiClient.Client,
			Object:     &copiedRouteAdvertisement,
			Definition: &copiedRouteAdvertisement,
		}

		routeAdvertisementObjects = append(routeAdvertisementObjects, routeAdvertisementBuilder)
	}

	return routeAdvertisementObjects, nil
}
