package olm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	operatorsv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// OperatorGroupBuilder provides a struct for OperatorGroup object containing connection to the
// cluster and the OperatorGroup definition.
type OperatorGroupBuilder struct {
	common.EmbeddableBuilder[operatorsv1.OperatorGroup, *operatorsv1.OperatorGroup]
	common.EmbeddableCreator[
		operatorsv1.OperatorGroup, OperatorGroupBuilder, *operatorsv1.OperatorGroup, *OperatorGroupBuilder]
	common.EmbeddableDeleter[operatorsv1.OperatorGroup, *operatorsv1.OperatorGroup]
	common.EmbeddableUpdater[
		operatorsv1.OperatorGroup, OperatorGroupBuilder, *operatorsv1.OperatorGroup, *OperatorGroupBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *OperatorGroupBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the OperatorGroup GVK for this builder.
func (builder *OperatorGroupBuilder) GetGVK() schema.GroupVersionKind {
	return operatorsv1.SchemeGroupVersion.WithKind("OperatorGroup")
}

// NewOperatorGroupBuilder returns an OperatorGroupBuilder struct.
func NewOperatorGroupBuilder(apiClient *clients.Settings, groupName, nsName string) *OperatorGroupBuilder {
	builder := common.NewNamespacedBuilder[operatorsv1.OperatorGroup, OperatorGroupBuilder](
		apiClient, operatorsv1.AddToScheme, groupName, nsName)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.GenerateName = fmt.Sprintf("%v-", groupName)
	builder.Definition.Spec.TargetNamespaces = []string{nsName}

	return builder
}

// PullOperatorGroup loads existing OperatorGroup from cluster into the OperatorGroupBuilder struct.
func PullOperatorGroup(apiClient *clients.Settings, groupName, nsName string) (*OperatorGroupBuilder, error) {
	return common.PullNamespacedBuilder[operatorsv1.OperatorGroup, OperatorGroupBuilder](
		context.TODO(), apiClient, operatorsv1.AddToScheme, groupName, nsName)
}
