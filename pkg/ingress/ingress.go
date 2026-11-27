package ingress

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
)

// DefaultIngressClassName is the default ingress class name for the ingress. It is added in [NewIngressBuilder] to all
// ingresses.
const DefaultIngressClassName = "openshift-default"

var ingressGVK = networkingv1.SchemeGroupVersion.WithKind("Ingress")

// IngressBuilder provides a struct for an ingress object from the cluster and an ingress definition.
type IngressBuilder struct {
	common.EmbeddableBuilder[networkingv1.Ingress, *networkingv1.Ingress]
	common.EmbeddableCreator[networkingv1.Ingress, IngressBuilder, *networkingv1.Ingress, *IngressBuilder]
	common.EmbeddableDeleter[networkingv1.Ingress, *networkingv1.Ingress]
	common.EmbeddableUpdater[networkingv1.Ingress, IngressBuilder, *networkingv1.Ingress, *IngressBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *IngressBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the Ingress GVK for this builder.
func (builder *IngressBuilder) GetGVK() schema.GroupVersionKind {
	return ingressGVK
}

// NewIngressBuilder creates a new instance of IngressBuilder.
func NewIngressBuilder(apiClient *clients.Settings, name, nsname string) *IngressBuilder {
	builder := common.NewNamespacedBuilder[networkingv1.Ingress, IngressBuilder](
		apiClient, networkingv1.AddToScheme, name, nsname)
	builder.Definition.Spec.IngressClassName = ptr.To(DefaultIngressClassName)

	return builder
}

// PullIngress loads an existing ingress into IngressBuilder struct.
func PullIngress(apiClient *clients.Settings, name, nsname string) (*IngressBuilder, error) {
	return common.PullNamespacedBuilder[networkingv1.Ingress, IngressBuilder](
		context.TODO(), apiClient, networkingv1.AddToScheme, name, nsname)
}
