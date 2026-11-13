package oran

import (
	"context"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NARBuilder provides a struct to inferface with NodeAllocationRequest resources on a specific cluster.
type NARBuilder struct {
	common.EmbeddableBuilder[
		hardwaremanagementv1alpha1.NodeAllocationRequest,
		*hardwaremanagementv1alpha1.NodeAllocationRequest,
	]
}

// GetGVK returns the NodeAllocationRequest GVK for this builder.
func (builder *NARBuilder) GetGVK() schema.GroupVersionKind {
	return hardwaremanagementv1alpha1.GroupVersion.WithKind("NodeAllocationRequest")
}

// PullNodeAllocationRequest pulls an existing NodeAllocationRequest into a NARBuilder struct.
func PullNodeAllocationRequest(apiClient *clients.Settings, name, nsname string) (*NARBuilder, error) {
	return common.PullNamespacedBuilder[hardwaremanagementv1alpha1.NodeAllocationRequest, NARBuilder](
		context.TODO(), apiClient, hardwaremanagementv1alpha1.AddToScheme, name, nsname)
}
