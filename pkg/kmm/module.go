package kmm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	moduleV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errNilContainer               = "invalid 'container' argument can not be nil"
	canNotRedefineModuleWithEmpty = "can not redefine module with empty ServiceAccount"
)

var moduleGVK = moduleV1Beta1.GroupVersion.WithKind("Module")

// ModuleBuilder provides struct for the module object containing connection to
// the cluster and the module definitions.
type ModuleBuilder struct {
	common.EmbeddableBuilder[moduleV1Beta1.Module, *moduleV1Beta1.Module]
	common.EmbeddableCreator[moduleV1Beta1.Module, ModuleBuilder, *moduleV1Beta1.Module, *ModuleBuilder]
	common.EmbeddableDeleteReturner[moduleV1Beta1.Module, ModuleBuilder, *moduleV1Beta1.Module, *ModuleBuilder]
	common.EmbeddableUpdater[moduleV1Beta1.Module, ModuleBuilder, *moduleV1Beta1.Module, *ModuleBuilder]
	common.EmbeddableWithOptions[
		moduleV1Beta1.Module, ModuleBuilder, *moduleV1Beta1.Module, *ModuleBuilder, ModuleAdditionalOptions]
}

// ModuleAdditionalOptions additional options for module object.
type ModuleAdditionalOptions func(builder *ModuleBuilder) (*ModuleBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ModuleBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the Module GVK for this builder.
func (builder *ModuleBuilder) GetGVK() schema.GroupVersionKind {
	return moduleGVK
}

// NewModuleBuilder creates a new instance of ModuleBuilder.
func NewModuleBuilder(
	apiClient *clients.Settings, name, nsname string) *ModuleBuilder {
	klog.V(100).Infof(
		"Initializing new Module structure with following params: %s, %s", name, nsname)

	return common.NewNamespacedBuilder[moduleV1Beta1.Module, ModuleBuilder](
		apiClient, moduleV1Beta1.AddToScheme, name, nsname)
}

// WithNodeSelector adds the specified NodeSelector to the Module.
func (builder *ModuleBuilder) WithNodeSelector(nodeSelector map[string]string) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating Module %s in namespace %s with this nodeSelector: %s",
		builder.Definition.Name, builder.Definition.Namespace, nodeSelector)

	if len(nodeSelector) == 0 {
		klog.V(100).Info("Can not redefine Module with empty nodeSelector map")

		//nolint:staticcheck // Preserve the package-specific error string.
		builder.SetError(fmt.Errorf("Module 'nodeSelector' cannot be empty map"))

		return builder
	}

	builder.Definition.Spec.Selector = nodeSelector

	return builder
}

// WithLoadServiceAccount adds the specified Load ServiceAccount to the Module.
func (builder *ModuleBuilder) WithLoadServiceAccount(srvAccountName string) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating Module %s in namespace %s with ModuleLoad ServiceAccount: %s",
		builder.Definition.Name, builder.Definition.Namespace, srvAccountName)

	return builder.withServiceAccount(srvAccountName, "module")
}

// WithDevicePluginServiceAccount adds the specified Device Plugin ServiceAccount to the Module.
func (builder *ModuleBuilder) WithDevicePluginServiceAccount(srvAccountName string) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating Module %s in namespace %s with DevicePlugin ServiceAccount: %s",
		builder.Definition.Name, builder.Definition.Namespace, srvAccountName)

	return builder.withServiceAccount(srvAccountName, "device")
}

// WithImageRepoSecret adds the specific ImageRepoSecret to the Module.
func (builder *ModuleBuilder) WithImageRepoSecret(imageRepoSecret string) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if imageRepoSecret == "" {
		builder.SetError(fmt.Errorf("can not redefine module with empty imageRepoSecret"))

		return builder
	}

	builder.Definition.Spec.ImageRepoSecret = &corev1.LocalObjectReference{Name: imageRepoSecret}

	return builder
}

// WithDevicePluginAutomountServiceAccountToken sets the AutomountServiceAccountToken field
// on the DevicePlugin spec. When set to false, the projected service account volume
// (SA token, CA files) will not be auto-mounted into the device plugin pod.
func (builder *ModuleBuilder) WithDevicePluginAutomountServiceAccountToken(
	automount bool) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting Module %s in namespace %s DevicePlugin AutomountServiceAccountToken to %v",
		builder.Definition.Name, builder.Definition.Namespace, automount)

	if builder.Definition.Spec.DevicePlugin == nil {
		builder.Definition.Spec.DevicePlugin = &moduleV1Beta1.DevicePluginSpec{}
	}

	builder.Definition.Spec.DevicePlugin.AutomountServiceAccountToken = &automount

	return builder
}

// WithDevicePluginVolume adds the specified DevicePlugin volume to the Module.
func (builder *ModuleBuilder) WithDevicePluginVolume(name string, configMapName string) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if name == "" {
		builder.SetError(fmt.Errorf("cannot redefine with empty volume 'name'"))

		return builder
	}

	if configMapName == "" {
		builder.SetError(fmt.Errorf("cannot redefine with empty 'configMapName'"))

		return builder
	}

	if builder.Definition.Spec.DevicePlugin == nil {
		builder.Definition.Spec.DevicePlugin = &moduleV1Beta1.DevicePluginSpec{}
	}

	builder.Definition.Spec.DevicePlugin.Volumes = append(builder.Definition.Spec.DevicePlugin.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName},
			},
		},
	})

	return builder
}

// WithToleration adds the specified toleration to the Module.
func (builder *ModuleBuilder) WithToleration(key, operator, value, effect string,
	seconds *int64) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if key == "" {
		builder.SetError(fmt.Errorf("cannot redefine with empty 'key' value"))

		return builder
	}

	if operator == "" {
		builder.SetError(fmt.Errorf("cannot redefine with empty 'operator' value"))

		return builder
	}

	if effect == "" {
		builder.SetError(fmt.Errorf("cannot redefine with empty 'effect' value"))

		return builder
	}

	klog.V(100).Infof(
		"Appending toleration with key: %s, operator: %s, value: %s, effect: %s and seconds: %v",
		key, operator, value, effect, seconds)

	builder.Definition.Spec.Tolerations = append(builder.Definition.Spec.Tolerations, []corev1.Toleration{{
		Key:               key,
		Effect:            corev1.TaintEffect(effect),
		Operator:          corev1.TolerationOperator(operator),
		Value:             value,
		TolerationSeconds: seconds,
	}}...)

	return builder
}

// WithModuleLoaderContainer adds the specified ModuleLoader container to the Module.
func (builder *ModuleBuilder) WithModuleLoaderContainer(
	container *moduleV1Beta1.ModuleLoaderContainerSpec) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if container == nil {
		builder.SetError(fmt.Errorf(errNilContainer))

		return builder
	}

	if builder.Definition.Spec.ModuleLoader == nil {
		builder.Definition.Spec.ModuleLoader = &moduleV1Beta1.ModuleLoaderSpec{}
	}

	builder.Definition.Spec.ModuleLoader.Container = *container

	return builder
}

// WithDevicePluginContainer adds the specified DevicePlugin container to the Module.
func (builder *ModuleBuilder) WithDevicePluginContainer(
	container *moduleV1Beta1.DevicePluginContainerSpec) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if container == nil {
		builder.SetError(fmt.Errorf(errNilContainer))

		return builder
	}

	if builder.Definition.Spec.DevicePlugin == nil {
		builder.Definition.Spec.DevicePlugin = &moduleV1Beta1.DevicePluginSpec{}
	}

	builder.Definition.Spec.DevicePlugin.Container = *container

	return builder
}

// WithImageRebuildTriggerGeneration adds the specified image rebuild trigger generation to the Module.
// Setting this to a different value than the current status will trigger re-verification of the module image.
// If the image is missing, a rebuild is triggered. If the image exists, no rebuild occurs.
func (builder *ModuleBuilder) WithImageRebuildTriggerGeneration(generation int) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating Module %s in namespace %s with ImageRebuildTriggerGeneration: %d",
		builder.Definition.Name, builder.Definition.Namespace, generation)

	builder.Definition.Spec.ImageRebuildTriggerGeneration = &generation

	return builder
}

// WithDRA sets the entire DRA spec on the Module.
func (builder *ModuleBuilder) WithDRA(dra *moduleV1Beta1.DRASpec) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting Module %s in namespace %s DRA spec",
		builder.Definition.Name, builder.Definition.Namespace)

	if dra == nil {
		builder.SetError(fmt.Errorf("invalid 'dra' argument cannot be nil"))

		return builder
	}

	builder.Definition.Spec.DRA = dra

	return builder
}

// WithDRAContainer sets the DRA container on the Module.
func (builder *ModuleBuilder) WithDRAContainer(
	container *moduleV1Beta1.CommonContainerSpec) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if container == nil {
		builder.SetError(fmt.Errorf(errNilContainer))

		return builder
	}

	if builder.Definition.Spec.DRA == nil {
		builder.Definition.Spec.DRA = &moduleV1Beta1.DRASpec{}
	}

	builder.Definition.Spec.DRA.Container = *container

	return builder
}

// WithDRAInitContainer sets the DRA init container on the Module.
func (builder *ModuleBuilder) WithDRAInitContainer(
	container *moduleV1Beta1.CommonContainerSpec) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if container == nil {
		builder.SetError(fmt.Errorf(errNilContainer))

		return builder
	}

	if builder.Definition.Spec.DRA == nil {
		builder.Definition.Spec.DRA = &moduleV1Beta1.DRASpec{}
	}

	builder.Definition.Spec.DRA.InitContainer = container

	return builder
}

// WithDRAServiceAccount sets the DRA ServiceAccount on the Module.
func (builder *ModuleBuilder) WithDRAServiceAccount(srvAccountName string) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting Module %s in namespace %s DRA ServiceAccount: %s",
		builder.Definition.Name, builder.Definition.Namespace, srvAccountName)

	return builder.withServiceAccount(srvAccountName, "dra")
}

// WithDRADriverName sets the DRA driver name on the Module.
func (builder *ModuleBuilder) WithDRADriverName(driverName string) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting Module %s in namespace %s DRA DriverName: %s",
		builder.Definition.Name, builder.Definition.Namespace, driverName)

	if driverName == "" {
		builder.SetError(fmt.Errorf("cannot set empty DRA driverName"))

		return builder
	}

	if builder.Definition.Spec.DRA == nil {
		builder.Definition.Spec.DRA = &moduleV1Beta1.DRASpec{}
	}

	builder.Definition.Spec.DRA.DriverName = driverName

	return builder
}

// WithDRADeviceClass appends a DeviceClass to the Module's DRA spec.
func (builder *ModuleBuilder) WithDRADeviceClass(
	deviceClass moduleV1Beta1.DeviceClassSpec) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Appending DRA DeviceClass %s to Module %s in namespace %s",
		deviceClass.Name, builder.Definition.Name, builder.Definition.Namespace)

	if deviceClass.Name == "" {
		builder.SetError(fmt.Errorf("cannot append DeviceClass with empty name"))

		return builder
	}

	if builder.Definition.Spec.DRA == nil {
		builder.Definition.Spec.DRA = &moduleV1Beta1.DRASpec{}
	}

	builder.Definition.Spec.DRA.DeviceClasses = append(
		builder.Definition.Spec.DRA.DeviceClasses, deviceClass)

	return builder
}

// WithDRAVolume adds a ConfigMap-backed volume to the Module's DRA spec.
func (builder *ModuleBuilder) WithDRAVolume(name string, configMapName string) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if name == "" {
		builder.SetError(fmt.Errorf("cannot add DRA volume with empty 'name'"))

		return builder
	}

	if configMapName == "" {
		builder.SetError(fmt.Errorf("cannot add DRA volume with empty 'configMapName'"))

		return builder
	}

	if builder.Definition.Spec.DRA == nil {
		builder.Definition.Spec.DRA = &moduleV1Beta1.DRASpec{}
	}

	builder.Definition.Spec.DRA.Volumes = append(builder.Definition.Spec.DRA.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName},
			},
		},
	})

	return builder
}

// WithDRAHostPathVolume adds a hostPath volume to the Module's DRA spec.
func (builder *ModuleBuilder) WithDRAHostPathVolume(
	name, path string, pathType *corev1.HostPathType) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if name == "" {
		builder.SetError(fmt.Errorf("cannot add DRA hostPath volume with empty 'name'"))

		return builder
	}

	if path == "" {
		builder.SetError(fmt.Errorf("cannot add DRA hostPath volume with empty 'path'"))

		return builder
	}

	if builder.Definition.Spec.DRA == nil {
		builder.Definition.Spec.DRA = &moduleV1Beta1.DRASpec{}
	}

	builder.Definition.Spec.DRA.Volumes = append(builder.Definition.Spec.DRA.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: path,
				Type: pathType,
			},
		},
	})

	return builder
}

// WithDRAAutomountServiceAccountToken sets the AutomountServiceAccountToken on the DRA spec.
func (builder *ModuleBuilder) WithDRAAutomountServiceAccountToken(
	automount bool) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting Module %s in namespace %s DRA AutomountServiceAccountToken to %v",
		builder.Definition.Name, builder.Definition.Namespace, automount)

	if builder.Definition.Spec.DRA == nil {
		builder.Definition.Spec.DRA = &moduleV1Beta1.DRASpec{}
	}

	builder.Definition.Spec.DRA.AutomountServiceAccountToken = &automount

	return builder
}

// BuildModuleSpec returns module spec.
func (builder *ModuleBuilder) BuildModuleSpec() (moduleV1Beta1.ModuleSpec, error) {
	if err := common.Validate(builder); err != nil {
		return moduleV1Beta1.ModuleSpec{}, err
	}

	klog.V(100).Infof(
		"Returning the ModuleSpec structure %v", builder.Definition.Spec)

	return builder.Definition.Spec, nil
}

// Pull pulls existing module from cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*ModuleBuilder, error) {
	klog.V(100).Infof("Pulling existing module name %s under namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[moduleV1Beta1.Module, ModuleBuilder](
		context.TODO(), apiClient, moduleV1Beta1.AddToScheme, name, nsname)
}

func (builder *ModuleBuilder) withServiceAccount(srvAccountName string, accountType string) *ModuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if srvAccountName == "" {
		builder.SetError(fmt.Errorf(canNotRedefineModuleWithEmpty))

		return builder
	}

	switch accountType {
	case "module":
		if builder.Definition.Spec.ModuleLoader == nil {
			builder.Definition.Spec.ModuleLoader = &moduleV1Beta1.ModuleLoaderSpec{}
		}

		builder.Definition.Spec.ModuleLoader.ServiceAccountName = srvAccountName
	case "device":
		if builder.Definition.Spec.DevicePlugin == nil {
			builder.Definition.Spec.DevicePlugin = &moduleV1Beta1.DevicePluginSpec{}
		}

		builder.Definition.Spec.DevicePlugin.ServiceAccountName = srvAccountName
	case "dra":
		if builder.Definition.Spec.DRA == nil {
			builder.Definition.Spec.DRA = &moduleV1Beta1.DRASpec{}
		}

		builder.Definition.Spec.DRA.ServiceAccountName = srvAccountName
	default:
		builder.SetError(fmt.Errorf(
			"invalid account type parameter. Supported parameters are: 'module', 'device', 'dra'"))

		return builder
	}

	return builder
}
