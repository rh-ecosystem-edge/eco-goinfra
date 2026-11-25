package olm

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	operatorsV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const resourceInstallPlan = "installplan"

// InstallPlanBuilder provides a struct for installplan object from the cluster and an installplan definition.
type InstallPlanBuilder struct {
	common.EmbeddableBuilder[operatorsV1alpha1.InstallPlan, *operatorsV1alpha1.InstallPlan]
	common.EmbeddableCreator[
		operatorsV1alpha1.InstallPlan, InstallPlanBuilder, *operatorsV1alpha1.InstallPlan, *InstallPlanBuilder]
	common.EmbeddableDeleter[operatorsV1alpha1.InstallPlan, *operatorsV1alpha1.InstallPlan]
	common.EmbeddableUpdater[
		operatorsV1alpha1.InstallPlan, InstallPlanBuilder, *operatorsV1alpha1.InstallPlan, *InstallPlanBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *InstallPlanBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the InstallPlan GVK for this builder.
func (builder *InstallPlanBuilder) GetGVK() schema.GroupVersionKind {
	return operatorsV1alpha1.SchemeGroupVersion.WithKind("InstallPlan")
}

// NewInstallPlanBuilder creates new instance of InstallPlanBuilder.
func NewInstallPlanBuilder(apiClient *clients.Settings, name, nsname string) *InstallPlanBuilder {
	return common.NewNamespacedBuilder[operatorsV1alpha1.InstallPlan, InstallPlanBuilder](
		apiClient, operatorsV1alpha1.AddToScheme, name, nsname)
}

// PullInstallPlan loads existing InstallPlan from cluster into the InstallPlanBuilder struct.
func PullInstallPlan(apiClient *clients.Settings, name, nsName string) (*InstallPlanBuilder, error) {
	return common.PullNamespacedBuilder[operatorsV1alpha1.InstallPlan, InstallPlanBuilder](
		context.TODO(), apiClient, operatorsV1alpha1.AddToScheme, name, nsName)
}
