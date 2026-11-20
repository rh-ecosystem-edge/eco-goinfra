package poddisruptionbudget

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	policyv1 "k8s.io/api/policy/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var pdbGVK = policyv1.SchemeGroupVersion.WithKind("PodDisruptionBudget")

// Builder provides a struct for the PodDisruptionBudget object and definition.
type Builder struct {
	common.EmbeddableBuilder[policyv1.PodDisruptionBudget, *policyv1.PodDisruptionBudget]
	common.EmbeddableCreator[policyv1.PodDisruptionBudget, Builder, *policyv1.PodDisruptionBudget, *Builder]
	common.EmbeddableDeleter[policyv1.PodDisruptionBudget, *policyv1.PodDisruptionBudget]
	common.EmbeddableForceUpdater[policyv1.PodDisruptionBudget, Builder, *policyv1.PodDisruptionBudget, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the PodDisruptionBudget GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return pdbGVK
}

// NewBuilder creates a new PodDisruptionBudget builder.
func NewBuilder(apiClient *clients.Settings, name, nsname string) *Builder {
	return common.NewNamespacedBuilder[policyv1.PodDisruptionBudget, Builder](
		apiClient, policyv1.AddToScheme, name, nsname)
}

// Pull retrieves the PodDisruptionBudget from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	return common.PullNamespacedBuilder[policyv1.PodDisruptionBudget, Builder](
		context.TODO(), apiClient, policyv1.AddToScheme, name, nsname)
}

// WithPDBSpec sets the PodDisruptionBudgetSpec for the PodDisruptionBudget.
func (builder *Builder) WithPDBSpec(spec policyv1.PodDisruptionBudgetSpec) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting PodDisruptionBudgetSpec for PodDisruptionBudget %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec = spec

	return builder
}

// GetGVR returns the GroupVersionResource for the Pod Disruption Budget.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "policy",
		Version:  "v1",
		Resource: "poddisruptionbudgets",
	}
}
