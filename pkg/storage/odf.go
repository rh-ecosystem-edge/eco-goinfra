package storage

import (
	"context"
	"fmt"

	odfoperatorv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var systemODFGVK = odfoperatorv1alpha1.GroupVersion.WithKind("StorageSystem")

// SystemODFBuilder provides struct for SystemODF object containing connection
// to the cluster and the SystemODF definitions.
type SystemODFBuilder struct {
	common.EmbeddableBuilder[odfoperatorv1alpha1.StorageSystem, *odfoperatorv1alpha1.StorageSystem]
	common.EmbeddableCreator[
		odfoperatorv1alpha1.StorageSystem, SystemODFBuilder,
		*odfoperatorv1alpha1.StorageSystem, *SystemODFBuilder]
	common.EmbeddableDeleter[odfoperatorv1alpha1.StorageSystem, *odfoperatorv1alpha1.StorageSystem]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *SystemODFBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the StorageSystem GVK for this builder.
func (builder *SystemODFBuilder) GetGVK() schema.GroupVersionKind {
	return systemODFGVK
}

// NewSystemODFBuilder creates a new instance of Builder.
func NewSystemODFBuilder(apiClient *clients.Settings, name, nsname string) *SystemODFBuilder {
	return common.NewNamespacedBuilder[odfoperatorv1alpha1.StorageSystem, SystemODFBuilder](
		apiClient, odfoperatorv1alpha1.AddToScheme, name, nsname)
}

// PullSystemODF gets an existing SystemODF object from the cluster.
func PullSystemODF(apiClient *clients.Settings, name, namespace string) (*SystemODFBuilder, error) {
	return common.PullNamespacedBuilder[odfoperatorv1alpha1.StorageSystem, SystemODFBuilder](
		context.TODO(), apiClient, odfoperatorv1alpha1.AddToScheme, name, namespace)
}

// WithSpec sets the SystemODF with storageCluster spec values.
func (builder *SystemODFBuilder) WithSpec(
	kind odfoperatorv1alpha1.StorageKind, name, nsname string) *SystemODFBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting SystemODF %s in namespace %s with storageCluster spec; \n"+
			"kind: %v, name: %s, namespace %s",
		builder.Definition.Name, builder.Definition.Namespace, kind, name, nsname)

	if kind == "" {
		klog.V(100).Info("The kind of the SystemODF spec is empty")

		builder.SetError(fmt.Errorf("SystemODF spec 'kind' cannot be empty"))

		return builder
	}

	if name == "" {
		klog.V(100).Info("The name of the storageCluster spec is empty")

		builder.SetError(fmt.Errorf("SystemODF spec 'name' cannot be empty"))

		return builder
	}

	if nsname == "" {
		klog.V(100).Info("The namespace of the SystemODF spec is empty")

		builder.SetError(fmt.Errorf("SystemODF spec 'nsname' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Kind = kind
	builder.Definition.Spec.Name = name
	builder.Definition.Spec.Namespace = nsname

	return builder
}
