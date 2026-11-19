package nrop

import (
	"context"
	"fmt"

	nropv1 "github.com/openshift-kni/numaresources-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var nropGVK = nropv1.GroupVersion.WithKind("NUMAResourcesOperator")

// Builder provides a struct for NUMAResourcesOperator object from the cluster and a NUMAResourcesOperator definition.
type Builder struct {
	common.EmbeddableBuilder[nropv1.NUMAResourcesOperator, *nropv1.NUMAResourcesOperator]
	common.EmbeddableCreator[nropv1.NUMAResourcesOperator, Builder, *nropv1.NUMAResourcesOperator, *Builder]
	common.EmbeddableDeleteReturner[nropv1.NUMAResourcesOperator, Builder, *nropv1.NUMAResourcesOperator, *Builder]
	common.EmbeddableUpdater[nropv1.NUMAResourcesOperator, Builder, *nropv1.NUMAResourcesOperator, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the NUMAResourcesOperator GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return nropGVK
}

// NewBuilder creates a new instance of NUMAResourcesOperator.
func NewBuilder(apiClient *clients.Settings, name string) *Builder {
	return common.NewClusterScopedBuilder[nropv1.NUMAResourcesOperator, Builder](
		apiClient, nropv1.AddToScheme, name)
}

// Pull pulls existing NUMAResourcesOperator from cluster.
func Pull(apiClient *clients.Settings, name string) (*Builder, error) {
	return common.PullClusterScopedBuilder[nropv1.NUMAResourcesOperator, Builder](
		context.TODO(), apiClient, nropv1.AddToScheme, name)
}

// WithMCPSelector sets the NUMAResourcesOperator operator's mcpSelector.
func (builder *Builder) WithMCPSelector(config nropv1.NodeGroupConfig, mcpSelector metav1.LabelSelector) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding machineConfigPoolSelector to the NUMAResourcesOperator %s; machineConfigPoolSelector %v",
		builder.Definition.Name, mcpSelector)

	nodeGroup := nropv1.NodeGroup{
		Config:                    &config,
		MachineConfigPoolSelector: &mcpSelector,
	}

	if len(mcpSelector.MatchLabels) == 0 && len(mcpSelector.MatchExpressions) == 0 {
		klog.V(100).Info("There are no labels for the machineConfigPoolSelector")

		builder.SetError(fmt.Errorf("NUMAResourcesOperator 'machineConfigPoolSelector' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.NodeGroups = append(builder.Definition.Spec.NodeGroups, nodeGroup)

	return builder
}
