package keda

import (
	"context"
	"fmt"

	kedav1alpha1 "github.com/kedacore/keda-olm-operator/api/keda/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var controllerGVK = kedav1alpha1.GroupVersion.WithKind("KedaController")

// ControllerBuilder provides a struct for KedaController object from the cluster and a KedaController definition.
type ControllerBuilder struct {
	common.EmbeddableBuilder[kedav1alpha1.KedaController, *kedav1alpha1.KedaController]
	common.EmbeddableCreator[kedav1alpha1.KedaController, ControllerBuilder,
		*kedav1alpha1.KedaController, *ControllerBuilder]
	common.EmbeddableDeleteReturner[kedav1alpha1.KedaController, ControllerBuilder,
		*kedav1alpha1.KedaController, *ControllerBuilder]
	common.EmbeddableUpdater[kedav1alpha1.KedaController, ControllerBuilder,
		*kedav1alpha1.KedaController, *ControllerBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ControllerBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the KedaController GVK for this builder.
func (builder *ControllerBuilder) GetGVK() schema.GroupVersionKind {
	return controllerGVK
}

// NewControllerBuilder creates a new instance of ControllerBuilder.
func NewControllerBuilder(
	apiClient *clients.Settings, name, nsname string) *ControllerBuilder {
	klog.V(100).Infof(
		"Initializing new kedaController structure with the following params: "+
			"name: %s, namespace: %s", name, nsname)

	return common.NewNamespacedBuilder[kedav1alpha1.KedaController, ControllerBuilder](
		apiClient, kedav1alpha1.AddToScheme, name, nsname)
}

// PullController pulls existing kedaController from cluster.
func PullController(apiClient *clients.Settings, name, nsname string) (*ControllerBuilder, error) {
	klog.V(100).Infof("Pulling existing kedaController name %s in namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[kedav1alpha1.KedaController, ControllerBuilder](
		context.TODO(), apiClient, kedav1alpha1.AddToScheme, name, nsname)
}

// WithAdmissionWebhooks sets the kedaController operator's profile.
func (builder *ControllerBuilder) WithAdmissionWebhooks(
	admissionWebhooks kedav1alpha1.KedaAdmissionWebhooksSpec) *ControllerBuilder {
	klog.V(100).Infof(
		"Adding admissionWebhooks to kedaController %s in namespace %s; admissionWebhooks %v",
		builder.Definition.Name, builder.Definition.Namespace, admissionWebhooks)

	if err := common.Validate(builder); err != nil {
		return builder
	}

	builder.Definition.Spec.AdmissionWebhooks = admissionWebhooks

	return builder
}

// WithOperator sets the kedaController operator's profile.
func (builder *ControllerBuilder) WithOperator(
	operator kedav1alpha1.KedaOperatorSpec) *ControllerBuilder {
	klog.V(100).Infof(
		"Adding operator to kedaController %s in namespace %s; operator %v",
		builder.Definition.Name, builder.Definition.Namespace, operator)

	if err := common.Validate(builder); err != nil {
		return builder
	}

	builder.Definition.Spec.Operator = operator

	return builder
}

// WithMetricsServer sets the kedaController operator's metricsServer.
func (builder *ControllerBuilder) WithMetricsServer(
	metricsServer kedav1alpha1.KedaMetricsServerSpec) *ControllerBuilder {
	klog.V(100).Infof(
		"Adding metricsServer to kedaController %s in namespace %s; metricsServer %v",
		builder.Definition.Name, builder.Definition.Namespace, metricsServer)

	if err := common.Validate(builder); err != nil {
		return builder
	}

	builder.Definition.Spec.MetricsServer = metricsServer

	return builder
}

// WithWatchNamespace sets the kedaController operator's watchNamespace.
func (builder *ControllerBuilder) WithWatchNamespace(
	watchNamespace string) *ControllerBuilder {
	klog.V(100).Infof(
		"Adding watchNamespace to kedaController %s in namespace %s; watchNamespace %v",
		builder.Definition.Name, builder.Definition.Namespace, watchNamespace)

	if err := common.Validate(builder); err != nil {
		return builder
	}

	if watchNamespace == "" {
		klog.V(100).Info("The watchNamespace is empty")

		builder.SetError(fmt.Errorf("'watchNamespace' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.WatchNamespace = watchNamespace

	return builder
}
