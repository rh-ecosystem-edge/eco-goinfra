package resource

import (
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/klog/v2"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListResourceSlices returns ResourceSlice objects matching the given options.
func ListResourceSlices(
	apiClient *clients.Settings,
	options ...goclient.ListOption) ([]resourcev1.ResourceSlice, error) {
	klog.V(100).Info("Listing ResourceSlices")

	if apiClient == nil {
		klog.V(100).Info("The apiClient is empty")

		return nil, fmt.Errorf("resourceSlice 'apiClient' cannot be nil")
	}

	err := apiClient.AttachScheme(resourcev1.AddToScheme)
	if err != nil {
		klog.V(100).Info("Failed to add resource v1 scheme to client schemes")

		return nil, err
	}

	sliceList := &resourcev1.ResourceSliceList{}

	err = apiClient.List(logging.DiscardContext(), sliceList, options...)
	if err != nil {
		return nil, err
	}

	return sliceList.Items, nil
}

// ListResourceSlicesByDriver returns ResourceSlice objects filtered by driver name.
func ListResourceSlicesByDriver(
	apiClient *clients.Settings,
	driverName string) ([]resourcev1.ResourceSlice, error) {
	klog.V(100).Infof("Listing ResourceSlices for driver %s", driverName)

	if driverName == "" {
		return nil, fmt.Errorf("resourceSlice 'driverName' cannot be empty")
	}

	allSlices, err := ListResourceSlices(apiClient)
	if err != nil {
		return nil, err
	}

	var filtered []resourcev1.ResourceSlice

	for idx := range allSlices {
		if allSlices[idx].Spec.Driver == driverName {
			filtered = append(filtered, allSlices[idx])
		}
	}

	return filtered, nil
}
