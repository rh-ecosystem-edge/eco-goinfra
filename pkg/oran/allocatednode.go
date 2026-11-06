package oran

import (
	"context"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// AllocatedNodeBuilder provides a struct to interface with AllocatedNode resources on a specific cluster.
type AllocatedNodeBuilder struct {
	common.EmbeddableBuilder[hardwaremanagementv1alpha1.AllocatedNode, *hardwaremanagementv1alpha1.AllocatedNode]
}

// GetGVK returns the AllocatedNode GVK for this builder.
func (builder *AllocatedNodeBuilder) GetGVK() schema.GroupVersionKind {
	return hardwaremanagementv1alpha1.GroupVersion.WithKind("AllocatedNode")
}

// PullAllocatedNode pulls an existing AllocatedNode into a AllocatedNodeBuilder struct.
func PullAllocatedNode(apiClient *clients.Settings, name, nsname string) (*AllocatedNodeBuilder, error) {
	return common.PullNamespacedBuilder[hardwaremanagementv1alpha1.AllocatedNode, AllocatedNodeBuilder](
		context.TODO(), apiClient, hardwaremanagementv1alpha1.AddToScheme, name, nsname)
}
