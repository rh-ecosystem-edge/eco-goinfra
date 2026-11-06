package neuron

import (
	"context"
	"errors"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	neuronv1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/neuron/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// AdditionalOptions additional options for DeviceConfig object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

const (
	errDriversImageEmpty      = "DeviceConfig 'driversImage' cannot be empty"
	errDriverVersionEmpty     = "DeviceConfig 'driverVersion' cannot be empty"
	errDevicePluginImageEmpty = "DeviceConfig 'devicePluginImage' cannot be empty"
	errDRADriverImageEmpty    = "DeviceConfig 'draDriverImage' cannot be empty"
	errDeviceClassesEmpty     = "DeviceConfig 'deviceClasses' cannot be empty"
	errSelectorEmpty          = "DeviceConfig 'selector' cannot be empty map"
	errSchedulerImageEmpty    = "DeviceConfig 'schedulerImage' cannot be empty"
	errExtensionImageEmpty    = "DeviceConfig 'extensionImage' cannot be empty"
	errImageRepoSecretEmpty   = "DeviceConfig 'imageRepoSecret' name cannot be empty"
	errNodeMetricsImageEmpty  = "DeviceConfig 'nodeMetricsImage' cannot be empty"
)

var deviceConfigGVK = neuronv1beta1.GroupVersion.WithKind("DeviceConfig")

// Builder provides struct for DeviceConfig object
// containing connection to the cluster and the DeviceConfig definition.
type Builder struct {
	common.EmbeddableBuilder[neuronv1beta1.DeviceConfig, *neuronv1beta1.DeviceConfig]
	common.EmbeddableCreator[neuronv1beta1.DeviceConfig, Builder, *neuronv1beta1.DeviceConfig, *Builder]
	common.EmbeddableDeleteReturner[neuronv1beta1.DeviceConfig, Builder, *neuronv1beta1.DeviceConfig, *Builder]
	common.EmbeddableForceUpdater[neuronv1beta1.DeviceConfig, Builder, *neuronv1beta1.DeviceConfig, *Builder]
	common.EmbeddableWithOptions[
		neuronv1beta1.DeviceConfig, Builder, *neuronv1beta1.DeviceConfig, *Builder, AdditionalOptions]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the DeviceConfig GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return deviceConfigGVK
}

// newDeviceConfigBuilder creates a namespaced DeviceConfig builder using the common builder initialization.
func newDeviceConfigBuilder(apiClient *clients.Settings, name, namespace string) *Builder {
	return common.NewNamespacedBuilder[neuronv1beta1.DeviceConfig, Builder](
		apiClient, neuronv1beta1.AddToScheme, name, namespace)
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(
	apiClient *clients.Settings,
	name, namespace string,
	driversImage, driverVersion, devicePluginImage string) *Builder {
	klog.V(100).Infof(
		"Initializing new DeviceConfig structure with name: %s, namespace: %s",
		name, namespace)

	builder := newDeviceConfigBuilder(apiClient, name, namespace)
	builder.Definition.Spec.DriversImage = driversImage
	builder.Definition.Spec.DriverVersion = driverVersion
	builder.Definition.Spec.DevicePluginImage = devicePluginImage

	if builder.GetError() != nil {
		return builder
	}

	if driversImage == "" {
		klog.V(100).Infof("The driversImage of the DeviceConfig is empty")

		builder.SetError(errors.New(errDriversImageEmpty))

		return builder
	}

	if driverVersion == "" {
		klog.V(100).Infof("The driverVersion of the DeviceConfig is empty")

		builder.SetError(errors.New(errDriverVersionEmpty))

		return builder
	}

	if devicePluginImage == "" {
		klog.V(100).Infof("The devicePluginImage of the DeviceConfig is empty")

		builder.SetError(errors.New(errDevicePluginImageEmpty))

		return builder
	}

	return builder
}

// NewBuilderWithInClusterBuild creates a new instance of Builder for in-cluster build mode.
// In this mode, DriversImage is left empty and the operator triggers KMM in-cluster builds
// using a Dockerfile ConfigMap. Only driverVersion and devicePluginImage are required.
func NewBuilderWithInClusterBuild(
	apiClient *clients.Settings,
	name, namespace string,
	driverVersion, devicePluginImage string) *Builder {
	klog.V(100).Infof(
		"Initializing new DeviceConfig (in-cluster build) with name: %s, namespace: %s",
		name, namespace)

	builder := newDeviceConfigBuilder(apiClient, name, namespace)
	builder.Definition.Spec.DriverVersion = driverVersion
	builder.Definition.Spec.DevicePluginImage = devicePluginImage

	if builder.GetError() != nil {
		return builder
	}

	if driverVersion == "" {
		klog.V(100).Infof("The driverVersion of the DeviceConfig is empty")

		builder.SetError(errors.New(errDriverVersionEmpty))

		return builder
	}

	if devicePluginImage == "" {
		klog.V(100).Infof("The devicePluginImage of the DeviceConfig is empty")

		builder.SetError(errors.New(errDevicePluginImageEmpty))

		return builder
	}

	return builder
}

// NewBuilderWithDRA creates a new instance of Builder for DRA mode.
// In DRA mode, draDriverImage replaces devicePluginImage and custom scheduler.
func NewBuilderWithDRA(
	apiClient *clients.Settings,
	name, namespace string,
	driversImage, driverVersion, draDriverImage string) *Builder {
	klog.V(100).Infof(
		"Initializing new DeviceConfig (DRA mode) with name: %s, namespace: %s",
		name, namespace)

	builder := newDeviceConfigBuilder(apiClient, name, namespace)
	builder.Definition.Spec.DriversImage = driversImage
	builder.Definition.Spec.DriverVersion = driverVersion
	builder.Definition.Spec.DRADriverImage = draDriverImage

	if builder.GetError() != nil {
		return builder
	}

	if driversImage == "" {
		klog.V(100).Infof("The driversImage of the DeviceConfig is empty")

		builder.SetError(errors.New(errDriversImageEmpty))

		return builder
	}

	if driverVersion == "" {
		klog.V(100).Infof("The driverVersion of the DeviceConfig is empty")

		builder.SetError(errors.New(errDriverVersionEmpty))

		return builder
	}

	if draDriverImage == "" {
		klog.V(100).Infof("The draDriverImage of the DeviceConfig is empty")

		builder.SetError(errors.New(errDRADriverImageEmpty))

		return builder
	}

	return builder
}

// NewBuilderWithInClusterBuildDRA creates a new instance of Builder for DRA mode with
// in-cluster build. DriversImage is left empty so KMM triggers in-cluster builds.
func NewBuilderWithInClusterBuildDRA(
	apiClient *clients.Settings,
	name, namespace string,
	driverVersion, draDriverImage string) *Builder {
	klog.V(100).Infof(
		"Initializing new DeviceConfig (DRA + in-cluster build) with name: %s, namespace: %s",
		name, namespace)

	builder := newDeviceConfigBuilder(apiClient, name, namespace)
	builder.Definition.Spec.DriverVersion = driverVersion
	builder.Definition.Spec.DRADriverImage = draDriverImage

	if builder.GetError() != nil {
		return builder
	}

	if driverVersion == "" {
		klog.V(100).Infof("The driverVersion of the DeviceConfig is empty")

		builder.SetError(errors.New(errDriverVersionEmpty))

		return builder
	}

	if draDriverImage == "" {
		klog.V(100).Infof("The draDriverImage of the DeviceConfig is empty")

		builder.SetError(errors.New(errDRADriverImageEmpty))

		return builder
	}

	return builder
}

// Pull retrieves an existing DeviceConfig from the cluster.
func Pull(apiClient *clients.Settings, name, namespace string) (*Builder, error) {
	return common.PullNamespacedBuilder[neuronv1beta1.DeviceConfig, Builder](
		context.TODO(), apiClient, neuronv1beta1.AddToScheme, name, namespace)
}

// WithDRADriverImage sets the DRA driver image for the DeviceConfig.
func (builder *Builder) WithDRADriverImage(image string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting DeviceConfig %s in namespace %s with draDriverImage: %s",
		builder.Definition.Name, builder.Definition.Namespace, image)

	if image == "" {
		klog.V(100).Infof(errDRADriverImageEmpty)

		builder.SetError(errors.New(errDRADriverImageEmpty))

		return builder
	}

	builder.Definition.Spec.DRADriverImage = image

	return builder
}

// WithDeviceClasses sets custom DeviceClasses on the DeviceConfig.
func (builder *Builder) WithDeviceClasses(
	deviceClasses []neuronv1beta1.DeviceClassSpec) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting DeviceConfig %s in namespace %s with %d deviceClasses",
		builder.Definition.Name, builder.Definition.Namespace, len(deviceClasses))

	if len(deviceClasses) == 0 {
		klog.V(100).Infof(errDeviceClassesEmpty)

		builder.SetError(errors.New(errDeviceClassesEmpty))

		return builder
	}

	builder.Definition.Spec.DeviceClasses = deviceClasses

	return builder
}

// WithSelector sets the node selector for the DeviceConfig.
func (builder *Builder) WithSelector(selector map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting DeviceConfig %s in namespace %s with selector: %v",
		builder.Definition.Name, builder.Definition.Namespace, selector)

	if len(selector) == 0 {
		klog.V(100).Infof(errSelectorEmpty)

		builder.SetError(errors.New(errSelectorEmpty))

		return builder
	}

	builder.Definition.Spec.Selector = selector

	return builder
}

// WithScheduler adds custom scheduler configuration to the DeviceConfig.
func (builder *Builder) WithScheduler(
	schedulerImage, extensionImage string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting DeviceConfig %s in namespace %s with scheduler images",
		builder.Definition.Name, builder.Definition.Namespace)

	if schedulerImage == "" {
		klog.V(100).Infof(errSchedulerImageEmpty)

		builder.SetError(errors.New(errSchedulerImageEmpty))

		return builder
	}

	if extensionImage == "" {
		klog.V(100).Infof(errExtensionImageEmpty)

		builder.SetError(errors.New(errExtensionImageEmpty))

		return builder
	}

	builder.Definition.Spec.CustomSchedulerImage = schedulerImage
	builder.Definition.Spec.SchedulerExtensionImage = extensionImage

	return builder
}

// WithImageRepoSecret sets the image repository secret.
func (builder *Builder) WithImageRepoSecret(secretName string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting DeviceConfig %s in namespace %s with imageRepoSecret: %s",
		builder.Definition.Name, builder.Definition.Namespace, secretName)

	if secretName == "" {
		klog.V(100).Infof(errImageRepoSecretEmpty)

		builder.SetError(errors.New(errImageRepoSecretEmpty))

		return builder
	}

	builder.Definition.Spec.ImageRepoSecret = &corev1.LocalObjectReference{
		Name: secretName,
	}

	return builder
}

// WithDriverVersion updates the driver version (triggers rolling upgrade).
func (builder *Builder) WithDriverVersion(version string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting DeviceConfig %s in namespace %s with driverVersion: %s",
		builder.Definition.Name, builder.Definition.Namespace, version)

	if version == "" {
		klog.V(100).Infof(errDriverVersionEmpty)

		builder.SetError(errors.New(errDriverVersionEmpty))

		return builder
	}

	builder.Definition.Spec.DriverVersion = version

	return builder
}

// WithNodeMetricsImage sets the node metrics image for the DeviceConfig.
func (builder *Builder) WithNodeMetricsImage(image string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting DeviceConfig %s in namespace %s with nodeMetricsImage: %s",
		builder.Definition.Name, builder.Definition.Namespace, image)

	if image == "" {
		klog.V(100).Infof(errNodeMetricsImageEmpty)

		builder.SetError(errors.New(errNodeMetricsImageEmpty))

		return builder
	}

	builder.Definition.Spec.NodeMetricsImage = image

	return builder
}
