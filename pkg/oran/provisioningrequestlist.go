package oran

import (
	"context"
	"fmt"

	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListProvisioningRequests returns a list of ProvisioningRequests, using the provided options.
func ListProvisioningRequests(
	apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*ProvisioningRequestBuilder, error) {
	if len(options) > 1 {
		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[
		provisioningv1alpha1.ProvisioningRequest, provisioningv1alpha1.ProvisioningRequestList,
		ProvisioningRequestBuilder](
		context.TODO(), apiClient, provisioningv1alpha1.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
