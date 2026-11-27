package servicemesh

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	istiov2 "maistra.io/api/core/v2"
)

var controlPlaneGVK = istiov2.SchemeGroupVersion.WithKind("ServiceMeshControlPlane")

// ControlPlaneBuilder provides a struct for serviceMeshControlPlane object from the cluster and
// a serviceMeshControlPlane definition.
type ControlPlaneBuilder struct {
	common.EmbeddableBuilder[istiov2.ServiceMeshControlPlane, *istiov2.ServiceMeshControlPlane]
	common.EmbeddableCreator[
		istiov2.ServiceMeshControlPlane,
		ControlPlaneBuilder,
		*istiov2.ServiceMeshControlPlane,
		*ControlPlaneBuilder,
	]
	common.EmbeddableDeleter[istiov2.ServiceMeshControlPlane, *istiov2.ServiceMeshControlPlane]
	common.EmbeddableForceUpdater[
		istiov2.ServiceMeshControlPlane,
		ControlPlaneBuilder,
		*istiov2.ServiceMeshControlPlane,
		*ControlPlaneBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ControlPlaneBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the ServiceMeshControlPlane GVK for this builder.
func (builder *ControlPlaneBuilder) GetGVK() schema.GroupVersionKind {
	return controlPlaneGVK
}

// NewControlPlaneBuilder method creates new instance of builder.
func NewControlPlaneBuilder(apiClient *clients.Settings, name, nsname string) *ControlPlaneBuilder {
	klog.V(100).Infof("Initializing new ControlPlaneBuilder structure with the following "+
		"params: name: %s, namespace: %s", name, nsname)

	return common.NewNamespacedBuilder[istiov2.ServiceMeshControlPlane, ControlPlaneBuilder](
		apiClient, istiov2.AddToScheme, name, nsname)
}

// WithAllAddonsDisabled disables all addons to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithAllAddonsDisabled() *ControlPlaneBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	enablement := false

	klog.V(100).Infof(
		"Creating serviceMeshControlPlane %s in namespace %s with the all addons disabled",
		builder.Definition.Name, builder.Definition.Namespace)

	disableAllAddons := istiov2.AddonsConfig{
		Prometheus: &istiov2.PrometheusAddonConfig{
			Enablement: istiov2.Enablement{Enabled: &enablement},
		},
		Grafana: &istiov2.GrafanaAddonConfig{
			Enablement: istiov2.Enablement{Enabled: &enablement},
		},
		Kiali: &istiov2.KialiAddonConfig{
			Enablement: istiov2.Enablement{Enabled: &enablement},
		},
		ThreeScale: &istiov2.ThreeScaleAddonConfig{
			Enablement: istiov2.Enablement{Enabled: &enablement},
		},
	}

	builder.Definition.Spec.Addons = &disableAllAddons

	return builder
}

// WithGrafanaAddon adds grafana addon to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithGrafanaAddon(
	enablement bool,
	grafanaInstallConfig *istiov2.GrafanaInstallConfig,
	address string) *ControlPlaneBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating serviceMeshControlPlane %s in namespace %s with the Grafana addons defined: enablement %v, "+
			"grafanaInstallConfig %v, address %s", builder.Definition.Name, builder.Definition.Namespace,
		enablement, grafanaInstallConfig, address)

	if enablement {
		if grafanaInstallConfig == nil {
			klog.V(100).Info("The grafanaInstallConfig of the Grafana addon is empty")

			builder.SetError(fmt.Errorf(
				"the Grafana addon 'grafanaInstallConfig' cannot be empty when Grafana addon is enabled"))

			return builder
		}

		if address == "" {
			klog.V(100).Info("The address of the Grafana addon is empty")

			builder.SetError(fmt.Errorf(
				"the Grafana addon 'address' cannot be empty when Grafana addon is enabled"))

			return builder
		}
	}

	addonConfig := &istiov2.GrafanaAddonConfig{
		Enablement: istiov2.Enablement{
			Enabled: &enablement,
		},
		Install: grafanaInstallConfig,
		Address: &address,
	}

	if builder.Definition.Spec.Addons == nil {
		builder.Definition.Spec.Addons = new(istiov2.AddonsConfig)
	}

	builder.Definition.Spec.Addons.Grafana = addonConfig

	return builder
}

// WithJaegerAddon adds jaeger addon to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithJaegerAddon(
	name string,
	jaegerInstallConfig *istiov2.JaegerInstallConfig) *ControlPlaneBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating serviceMeshControlPlane %s in namespace %s with the JaegerAddonConfig defined",
		builder.Definition.Name, builder.Definition.Namespace)

	if name == "" {
		klog.V(100).Info("The name of the Jaeger addon is empty")

		builder.SetError(fmt.Errorf("the Jaeger addon 'name' cannot be empty"))

		return builder
	}

	if jaegerInstallConfig == nil {
		klog.V(100).Info("The jaegerInstallConfig of the Jaeger addon is empty")

		builder.SetError(fmt.Errorf("the Jaeger addon 'jaegerInstallConfig' cannot be empty"))

		return builder
	}

	addonConfig := &istiov2.JaegerAddonConfig{
		Name:    name,
		Install: jaegerInstallConfig,
	}

	if builder.Definition.Spec.Addons == nil {
		builder.Definition.Spec.Addons = new(istiov2.AddonsConfig)
	}

	builder.Definition.Spec.Addons.Jaeger = addonConfig

	return builder
}

// WithKialiAddon adds kiali addons to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithKialiAddon(
	enablement bool,
	name string,
	kialiInstallConfig *istiov2.KialiInstallConfig) *ControlPlaneBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating serviceMeshControlPlane %s in namespace %s with the KialiAddonConfig defined",
		builder.Definition.Name, builder.Definition.Namespace)

	if enablement {
		if kialiInstallConfig == nil {
			klog.V(100).Info("The kialiInstallConfig of the Kiali addon is empty")

			builder.SetError(fmt.Errorf(
				"the Kiali addon 'kialiInstallConfig' cannot be empty when Kiali addon is enabled"))

			return builder
		}

		if name == "" {
			klog.V(100).Info("The name of the Kiali addon is empty")

			builder.SetError(fmt.Errorf("the Kiali addon 'name' cannot be empty when Kiali addon is enabled"))

			return builder
		}
	}

	addonConfig := &istiov2.KialiAddonConfig{
		Enablement: istiov2.Enablement{
			Enabled: &enablement,
		},
		Name:    name,
		Install: kialiInstallConfig,
	}

	if builder.Definition.Spec.Addons == nil {
		builder.Definition.Spec.Addons = new(istiov2.AddonsConfig)
	}

	builder.Definition.Spec.Addons.Kiali = addonConfig

	return builder
}

// WithPrometheusAddon adds prometheus addons to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithPrometheusAddon(
	enablement bool,
	scrape bool,
	metricsExpiryDuration string,
	address string,
	prometheusInstallConfig *istiov2.PrometheusInstallConfig) *ControlPlaneBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating serviceMeshControlPlane %s in namespace %s with the PrometheusAddonConfig defined",
		builder.Definition.Name, builder.Definition.Namespace)

	if enablement {
		if prometheusInstallConfig == nil {
			klog.V(100).Info("The prometheusInstallConfig of the Prometheus addon is empty")

			builder.SetError(fmt.Errorf(
				"the Prometheus addon 'prometheusInstallConfig' cannot be empty when Prometheus addon is enabled"))

			return builder
		}

		if metricsExpiryDuration == "" {
			klog.V(100).Info("The metricsExpiryDuration of the Prometheus addon is empty")

			builder.SetError(fmt.Errorf(
				"the Prometheus addon 'metricsExpiryDuration' cannot be empty when Prometheus addon is enabled"))

			return builder
		}

		if address == "" {
			klog.V(100).Info("The address of the Prometheus addon is empty")

			builder.SetError(fmt.Errorf(
				"the Prometheus addon 'address' cannot be empty when Prometheus addon is enabled"))

			return builder
		}
	}

	addonConfig := &istiov2.PrometheusAddonConfig{
		Enablement: istiov2.Enablement{
			Enabled: &enablement,
		},
		MetricsExpiryDuration: metricsExpiryDuration,
		Scrape:                &scrape,
		Install:               prometheusInstallConfig,
		Address:               &address,
	}

	if builder.Definition.Spec.Addons == nil {
		builder.Definition.Spec.Addons = new(istiov2.AddonsConfig)
	}

	builder.Definition.Spec.Addons.Prometheus = addonConfig

	return builder
}

// WithGatewaysEnablement adds gateway enablement to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithGatewaysEnablement(enablement bool) *ControlPlaneBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating serviceMeshControlPlane %s in namespace %s with enabled Gateways",
		builder.Definition.Name, builder.Definition.Namespace)

	gatewaysConfig := istiov2.Enablement{Enabled: &enablement}

	if builder.Definition.Spec.Gateways == nil {
		builder.Definition.Spec.Gateways = &istiov2.GatewaysConfig{}
	}

	builder.Definition.Spec.Gateways.Enablement = gatewaysConfig

	return builder
}

// PullControlPlane retrieves an existing serviceMeshControlPlane object from the cluster.
func PullControlPlane(apiClient *clients.Settings, name, nsname string) (*ControlPlaneBuilder, error) {
	klog.V(100).Infof(
		"Pulling serviceMeshControlPlane object name %s in namespace: %s", name, nsname)

	return common.PullNamespacedBuilder[istiov2.ServiceMeshControlPlane, ControlPlaneBuilder](
		context.TODO(), apiClient, istiov2.AddToScheme, name, nsname)
}
