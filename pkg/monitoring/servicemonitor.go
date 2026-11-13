package monitoring

import (
	"context"
	"fmt"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const errEmptyLabels = "labels can not be empty"

var serviceMonitorGVK = monv1.SchemeGroupVersion.WithKind("ServiceMonitor")

// Builder provides a struct for serviceMonitor object from the cluster
// and a serviceMonitor definition.
type Builder struct {
	common.EmbeddableBuilder[monv1.ServiceMonitor, *monv1.ServiceMonitor]
	common.EmbeddableCreator[monv1.ServiceMonitor, Builder, *monv1.ServiceMonitor, *Builder]
	common.EmbeddableDeleteReturner[monv1.ServiceMonitor, Builder, *monv1.ServiceMonitor, *Builder]
	common.EmbeddableUpdater[monv1.ServiceMonitor, Builder, *monv1.ServiceMonitor, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the ServiceMonitor GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return serviceMonitorGVK
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(
	apiClient *clients.Settings, name, nsname string) *Builder {
	return common.NewNamespacedBuilder[monv1.ServiceMonitor, Builder](
		apiClient, monv1.AddToScheme, name, nsname)
}

// Pull pulls existing serviceMonitor from cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	return common.PullNamespacedBuilder[monv1.ServiceMonitor, Builder](
		context.TODO(), apiClient, monv1.AddToScheme, name, nsname)
}

// WithEndpoints sets the serviceMonitor operator's endpoints.
func (builder *Builder) WithEndpoints(
	endpoints []monv1.Endpoint) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding endpoints to serviceMonitor %s in namespace %s; endpoints %v",
		builder.Definition.Name, builder.Definition.Namespace, endpoints)

	if len(endpoints) == 0 {
		klog.V(100).Info("'endpoints' argument cannot be empty")

		builder.SetError(fmt.Errorf("'endpoints' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Endpoints = endpoints

	return builder
}

// WithLabels redefines the serviceMonitor with labels.
func (builder *Builder) WithLabels(labels map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Defining serviceMonitor with labels: %v", labels)

	if len(labels) == 0 {
		klog.V(100).Info(errEmptyLabels)

		builder.SetError(fmt.Errorf(errEmptyLabels))

		return builder
	}

	for key := range labels {
		if key == "" {
			klog.V(100).Info("The 'labels' key cannot be empty")

			builder.SetError(fmt.Errorf("can not apply a labels with an empty key"))

			return builder
		}
	}

	builder.Definition.Labels = labels

	return builder
}

// WithSelector redefines the serviceMonitor with selector.
func (builder *Builder) WithSelector(selector map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Defining serviceMonitor with selector: %v", selector)

	if len(selector) == 0 {
		klog.V(100).Info("selector can not be empty")

		builder.SetError(fmt.Errorf("selector can not be empty"))

		return builder
	}

	for key := range selector {
		if key == "" {
			klog.V(100).Info("The 'selector' key cannot be empty")

			builder.SetError(fmt.Errorf("can not apply a selector with an empty key"))

			return builder
		}
	}

	builder.Definition.Spec.Selector.MatchLabels = selector

	return builder
}

// WithNamespaceSelector redefines the serviceMonitor with namespaceSelector.
func (builder *Builder) WithNamespaceSelector(namespaceSelector []string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Defining serviceMonitor with namespaceSelector: %v", namespaceSelector)

	if len(namespaceSelector) == 0 {
		klog.V(100).Info("namespaceSelector can not be empty")

		builder.SetError(fmt.Errorf("namespaceSelector can not be empty"))

		return builder
	}

	builder.Definition.Spec.NamespaceSelector = monv1.NamespaceSelector{
		MatchNames: namespaceSelector,
	}

	return builder
}
