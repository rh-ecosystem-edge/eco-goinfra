package oran

import (
	"context"
	"fmt"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListNodeAllocationRequests returns a list of NodeAllocationRequests in all namespaces, using the provided options.
func ListNodeAllocationRequests(
	apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*NARBuilder, error) {
	if len(options) > 1 {
		klog.V(100).Info("NodeAllocationRequests 'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[
		hardwaremanagementv1alpha1.NodeAllocationRequest,
		hardwaremanagementv1alpha1.NodeAllocationRequestList,
		NARBuilder,
	](
		context.TODO(), apiClient, hardwaremanagementv1alpha1.AddToScheme,
		common.ConvertListOptionsToOptions(options)...)
}
