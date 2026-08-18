package resource

import (
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/klog/v2"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListResourceClaims returns ResourceClaim objects matching the given options.
func ListResourceClaims(
	apiClient *clients.Settings,
	namespace string,
	options ...goclient.ListOption) ([]resourcev1.ResourceClaim, error) {
	klog.V(100).Infof("Listing ResourceClaims in namespace %s", namespace)

	if apiClient == nil {
		klog.V(100).Info("The apiClient is empty")

		return nil, fmt.Errorf("resourceClaim 'apiClient' cannot be nil")
	}

	if namespace == "" {
		klog.V(100).Info("The namespace is empty")

		return nil, fmt.Errorf("resourceClaim 'namespace' cannot be empty")
	}

	err := apiClient.AttachScheme(resourcev1.AddToScheme)
	if err != nil {
		klog.V(100).Info("Failed to add resource v1 scheme to client schemes")

		return nil, err
	}

	claimList := &resourcev1.ResourceClaimList{}

	allOptions := append(append([]goclient.ListOption{}, options...), goclient.InNamespace(namespace))

	err = apiClient.List(logging.DiscardContext(), claimList, allOptions...)
	if err != nil {
		return nil, err
	}

	return claimList.Items, nil
}

// GetResourceClaim fetches a single ResourceClaim by name and namespace.
func GetResourceClaim(
	apiClient *clients.Settings,
	name, namespace string) (*resourcev1.ResourceClaim, error) {
	klog.V(100).Infof("Getting ResourceClaim %s in namespace %s", name, namespace)

	if apiClient == nil {
		return nil, fmt.Errorf("resourceClaim 'apiClient' cannot be nil")
	}

	if name == "" {
		return nil, fmt.Errorf("resourceClaim 'name' cannot be empty")
	}

	if namespace == "" {
		return nil, fmt.Errorf("resourceClaim 'namespace' cannot be empty")
	}

	err := apiClient.AttachScheme(resourcev1.AddToScheme)
	if err != nil {
		return nil, err
	}

	claim := &resourcev1.ResourceClaim{}

	err = apiClient.Get(logging.DiscardContext(), goclient.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, claim)
	if err != nil {
		return nil, err
	}

	return claim, nil
}
