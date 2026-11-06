package webhook

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	admregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var mutatingConfigurationGVK = admregv1.SchemeGroupVersion.WithKind("MutatingWebhookConfiguration")

// MutatingConfigurationBuilder provides struct for MutatingWebhookConfiguration object which contains connection
// to cluster and MutatingWebhookConfiguration definition.
type MutatingConfigurationBuilder struct {
	common.EmbeddableBuilder[admregv1.MutatingWebhookConfiguration, *admregv1.MutatingWebhookConfiguration]
	common.EmbeddableDeleteReturner[
		admregv1.MutatingWebhookConfiguration, MutatingConfigurationBuilder,
		*admregv1.MutatingWebhookConfiguration, *MutatingConfigurationBuilder,
	]
	common.EmbeddableUpdater[
		admregv1.MutatingWebhookConfiguration, MutatingConfigurationBuilder,
		*admregv1.MutatingWebhookConfiguration, *MutatingConfigurationBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *MutatingConfigurationBuilder) AttachMixins() {
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the MutatingWebhookConfiguration GVK for this builder.
func (builder *MutatingConfigurationBuilder) GetGVK() schema.GroupVersionKind {
	return mutatingConfigurationGVK
}

// PullMutatingConfiguration pulls existing MutatingWebhookConfiguration from cluster.
func PullMutatingConfiguration(apiClient *clients.Settings, name string) (*MutatingConfigurationBuilder, error) {
	klog.V(100).Infof("Pulling existing MutatingWebhookConfiguration name %s from cluster", name)

	return common.PullClusterScopedBuilder[admregv1.MutatingWebhookConfiguration, MutatingConfigurationBuilder](
		context.TODO(), apiClient, admregv1.AddToScheme, name)
}
