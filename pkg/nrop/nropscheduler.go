package nrop

import (
	"context"
	"fmt"

	nropv1 "github.com/openshift-kni/numaresources-operator/api/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var schedulerGVK = nropv1.GroupVersion.WithKind("NUMAResourcesScheduler")

// SchedulerBuilder provides a struct for NUMAResourcesScheduler object from the cluster
// and a NUMAResourcesScheduler definition.
type SchedulerBuilder struct {
	common.EmbeddableBuilder[nropv1.NUMAResourcesScheduler, *nropv1.NUMAResourcesScheduler]
	common.EmbeddableCreator[nropv1.NUMAResourcesScheduler, SchedulerBuilder, *nropv1.NUMAResourcesScheduler, *SchedulerBuilder]
	common.EmbeddableDeleteReturner[
		nropv1.NUMAResourcesScheduler,
		SchedulerBuilder,
		*nropv1.NUMAResourcesScheduler,
		*SchedulerBuilder,
	]
	common.EmbeddableUpdater[nropv1.NUMAResourcesScheduler, SchedulerBuilder, *nropv1.NUMAResourcesScheduler, *SchedulerBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *SchedulerBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the NUMAResourcesScheduler GVK for this builder.
func (builder *SchedulerBuilder) GetGVK() schema.GroupVersionKind {
	return schedulerGVK
}

// NewSchedulerBuilder creates a new instance of NUMAResourcesScheduler.
func NewSchedulerBuilder(apiClient *clients.Settings, name, nsname string) *SchedulerBuilder {
	return common.NewNamespacedBuilder[nropv1.NUMAResourcesScheduler, SchedulerBuilder](
		apiClient, nropv1.AddToScheme, name, nsname)
}

// PullScheduler pulls existing NUMAResourcesScheduler from cluster.
func PullScheduler(apiClient *clients.Settings, name, nsname string) (*SchedulerBuilder, error) {
	return common.PullNamespacedBuilder[nropv1.NUMAResourcesScheduler, SchedulerBuilder](
		context.TODO(), apiClient, nropv1.AddToScheme, name, nsname)
}

// WithImageSpec sets the NUMAResourcesScheduler operator's imageSpec.
func (builder *SchedulerBuilder) WithImageSpec(imageSpec string) *SchedulerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding imageSpec to the NUMAResourcesScheduler %s in namespace %s; imageSpec: %s",
		builder.Definition.Name, builder.Definition.Namespace, imageSpec)

	if imageSpec == "" {
		klog.V(100).Info("The 'NUMAResourcesScheduler' imageSpec cannot be empty")

		builder.SetError(fmt.Errorf("can not apply a NUMAResourcesScheduler with an empty imageSpec"))

		return builder
	}

	builder.Definition.Spec.SchedulerImage = imageSpec

	return builder
}

// WithSchedulerName sets the NUMAResourcesScheduler operator's schedulerName.
func (builder *SchedulerBuilder) WithSchedulerName(schedulerName string) *SchedulerBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding schedulerName to the NUMAResourcesScheduler %s in namespace %s; schedulerName: %s",
		builder.Definition.Name, builder.Definition.Namespace, schedulerName)

	if schedulerName == "" {
		klog.V(100).Info("The 'NUMAResourcesScheduler' schedulerName cannot be empty")

		builder.SetError(fmt.Errorf("can not apply a NUMAResourcesScheduler with an empty schedulerName"))

		return builder
	}

	builder.Definition.Spec.SchedulerName = schedulerName

	return builder
}
