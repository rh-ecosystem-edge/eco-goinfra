package metallb

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// ServiceBGPStatusBuilder provides struct for ServiceBGPStatus object which contains connection to cluster and
// ServiceBGPStatus definitions.
type ServiceBGPStatusBuilder struct {
	common.EmbeddableBuilder[mlbtypes.ServiceBGPStatus, *mlbtypes.ServiceBGPStatus]
}

// GetGVK returns the ServiceBGPStatus GVK for this builder.
func (builder *ServiceBGPStatusBuilder) GetGVK() schema.GroupVersionKind {
	return mlbtypes.GroupVersion.WithKind("ServiceBGPStatus")
}

// PullServiceBGPStatus retrieves an existing ServiceBGPStatus object from the cluster.
func PullServiceBGPStatus(apiClient *clients.Settings, name string) (*ServiceBGPStatusBuilder, error) {
	klog.V(100).Infof("Pulling ServiceBGPStatus object name:%s", name)

	return common.PullClusterScopedBuilder[mlbtypes.ServiceBGPStatus, ServiceBGPStatusBuilder](
		context.TODO(), apiClient, mlbtypes.AddToScheme, name)
}
