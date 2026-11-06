package webhook

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	admregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var validatingConfigurationGVK = admregv1.SchemeGroupVersion.WithKind("ValidatingWebhookConfiguration")

// ValidatingConfigurationBuilder provides struct for ValidatingWebhookConfiguration object
// which contains connection to cluster and ValidatingWebhookConfiguration definition.
type ValidatingConfigurationBuilder struct {
	common.EmbeddableBuilder[admregv1.ValidatingWebhookConfiguration, *admregv1.ValidatingWebhookConfiguration]
	common.EmbeddableDeleteReturner[
		admregv1.ValidatingWebhookConfiguration, ValidatingConfigurationBuilder,
		*admregv1.ValidatingWebhookConfiguration, *ValidatingConfigurationBuilder,
	]
	common.EmbeddableUpdater[
		admregv1.ValidatingWebhookConfiguration, ValidatingConfigurationBuilder,
		*admregv1.ValidatingWebhookConfiguration, *ValidatingConfigurationBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ValidatingConfigurationBuilder) AttachMixins() {
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the ValidatingWebhookConfiguration GVK for this builder.
func (builder *ValidatingConfigurationBuilder) GetGVK() schema.GroupVersionKind {
	return validatingConfigurationGVK
}

// PullValidatingConfiguration pulls existing ValidatingWebhookConfiguration from cluster.
func PullValidatingConfiguration(apiClient *clients.Settings, name string) (*ValidatingConfigurationBuilder, error) {
	klog.V(100).Infof("Pulling existing ValidatingWebhookConfiguration name %s from cluster", name)

	return common.PullClusterScopedBuilder[admregv1.ValidatingWebhookConfiguration, ValidatingConfigurationBuilder](
		context.TODO(), apiClient, admregv1.AddToScheme, name)
}
