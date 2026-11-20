package oran

import (
	"context"
	"fmt"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListAllocatedNodes returns a list of AllocatedNodes in all namespaces, using the provided options.
func ListAllocatedNodes(
	apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*AllocatedNodeBuilder, error) {
	if len(options) > 1 {
		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[hardwaremanagementv1alpha1.AllocatedNode, hardwaremanagementv1alpha1.AllocatedNodeList, AllocatedNodeBuilder](
		context.TODO(), apiClient, hardwaremanagementv1alpha1.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
