package metallb

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListServiceBGPStatus returns service bgp status inventory in the given cluster.
func ListServiceBGPStatus(
	apiClient *clients.Settings, options ...client.ListOptions) ([]*ServiceBGPStatusBuilder, error) {
	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[mlbtypes.ServiceBGPStatus, mlbtypes.ServiceBGPStatusList, ServiceBGPStatusBuilder](
		context.TODO(), apiClient, mlbtypes.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
