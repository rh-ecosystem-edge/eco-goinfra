package keda

import (
	"context"
	"fmt"

	kedav2v1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var triggerAuthenticationGVK = kedav2v1alpha1.GroupVersion.WithKind("TriggerAuthentication")

// TriggerAuthenticationBuilder provides a struct for TriggerAuthentication object from the cluster
// and a TriggerAuthentication definition.
type TriggerAuthenticationBuilder struct {
	common.EmbeddableBuilder[kedav2v1alpha1.TriggerAuthentication, *kedav2v1alpha1.TriggerAuthentication]
	common.EmbeddableCreator[
		kedav2v1alpha1.TriggerAuthentication,
		TriggerAuthenticationBuilder,
		*kedav2v1alpha1.TriggerAuthentication,
		*TriggerAuthenticationBuilder,
	]
	common.EmbeddableDeleteReturner[
		kedav2v1alpha1.TriggerAuthentication,
		TriggerAuthenticationBuilder,
		*kedav2v1alpha1.TriggerAuthentication,
		*TriggerAuthenticationBuilder,
	]
	common.EmbeddableUpdater[
		kedav2v1alpha1.TriggerAuthentication,
		TriggerAuthenticationBuilder,
		*kedav2v1alpha1.TriggerAuthentication,
		*TriggerAuthenticationBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *TriggerAuthenticationBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the TriggerAuthentication GVK for this builder.
func (builder *TriggerAuthenticationBuilder) GetGVK() schema.GroupVersionKind {
	return triggerAuthenticationGVK
}

// NewTriggerAuthenticationBuilder creates a new instance of TriggerAuthenticationBuilder.
func NewTriggerAuthenticationBuilder(
	apiClient *clients.Settings, name, nsname string) *TriggerAuthenticationBuilder {
	return common.NewNamespacedBuilder[kedav2v1alpha1.TriggerAuthentication, TriggerAuthenticationBuilder](
		apiClient, kedav2v1alpha1.AddToScheme, name, nsname)
}

// PullTriggerAuthentication pulls existing triggerAuthentication from cluster.
func PullTriggerAuthentication(apiClient *clients.Settings,
	name, nsname string) (*TriggerAuthenticationBuilder, error) {
	return common.PullNamespacedBuilder[kedav2v1alpha1.TriggerAuthentication, TriggerAuthenticationBuilder](
		context.TODO(), apiClient, kedav2v1alpha1.AddToScheme, name, nsname)
}

// WithSecretTargetRef sets the triggerAuthentication operator's secretTargetRef.
func (builder *TriggerAuthenticationBuilder) WithSecretTargetRef(
	secretTargetRef []kedav2v1alpha1.AuthSecretTargetRef) *TriggerAuthenticationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding secretTargetRef to triggerAuthentication %s in namespace %s; secretTargetRef %v",
		builder.Definition.Name, builder.Definition.Namespace, secretTargetRef)

	if len(secretTargetRef) == 0 {
		klog.V(100).Info("'secretTargetRef' argument cannot be empty")

		builder.SetError(fmt.Errorf("'secretTargetRef' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.SecretTargetRef = secretTargetRef

	return builder
}
