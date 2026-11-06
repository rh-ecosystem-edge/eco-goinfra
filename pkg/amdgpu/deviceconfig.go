package amdgpu

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	amdgpuv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/amd/gpu-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
)

var deviceConfigGVK = amdgpuv1.GroupVersion.WithKind("DeviceConfig")

// Builder provides a struct for DeviceConfig object
// from the cluster and a DeviceConfig definition.
type Builder struct {
	common.EmbeddableBuilder[amdgpuv1.DeviceConfig, *amdgpuv1.DeviceConfig]
	common.EmbeddableCreator[
		amdgpuv1.DeviceConfig, Builder, *amdgpuv1.DeviceConfig, *Builder]
	common.EmbeddableDeleteReturner[
		amdgpuv1.DeviceConfig, Builder, *amdgpuv1.DeviceConfig, *Builder]
	common.EmbeddableForceUpdater[
		amdgpuv1.DeviceConfig, Builder, *amdgpuv1.DeviceConfig, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the DeviceConfig GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return deviceConfigGVK
}

// NewBuilderFromObjectString creates a Builder object from CSV alm-examples.
func NewBuilderFromObjectString(apiClient *clients.Settings, almExample string) *Builder {
	klog.V(100).Infof("Initializing new Builder structure from almExample string")

	if apiClient == nil {
		klog.V(100).Info("The apiClient of the DeviceConfig is nil")

		return nil
	}

	err := apiClient.AttachScheme(amdgpuv1.AddToScheme)
	if err != nil {
		klog.V(100).Info("Failed to add amdgpu v1 scheme to client schemes")

		return nil
	}

	builder := &Builder{}
	builder.AttachMixins()
	builder.SetGVK(builder.GetGVK())
	builder.SetClient(apiClient)

	deviceConfig, err := getDeviceConfigFromAlmExample(almExample)
	if err != nil {
		klog.V(100).Infof(
			"Error initializing DeviceConfig from alm-examples: %s", err.Error())

		builder.SetError(fmt.Errorf("error initializing DeviceConfig from alm-examples: %w", err))

		return builder
	}

	builder.SetDefinition(deviceConfig)

	klog.V(100).Infof(
		"Initializing Builder definition to DeviceConfig object")

	if builder.Definition == nil {
		klog.V(100).Info("The DeviceConfig object definition is nil")

		builder.SetError(fmt.Errorf("deviceConfig definition is nil"))

		return builder
	}

	return builder
}

// Pull loads an existing DeviceConfig into Builder struct.
func Pull(apiClient *clients.Settings, name, namespace string) (*Builder, error) {
	return common.PullNamespacedBuilder[amdgpuv1.DeviceConfig, Builder](
		context.TODO(), apiClient, amdgpuv1.AddToScheme, name, namespace)
}

// getDeviceConfigFromAlmExample extracts the DeviceConfig from the alm-examples block.
func getDeviceConfigFromAlmExample(almExample string) (*amdgpuv1.DeviceConfig, error) {
	deviceConfigList := &amdgpuv1.DeviceConfigList{}

	if almExample == "" {
		return nil, fmt.Errorf("almExample is an empty string")
	}

	err := json.Unmarshal([]byte(almExample), &deviceConfigList.Items)
	if err != nil {
		return nil, err
	}

	if len(deviceConfigList.Items) == 0 {
		return nil, fmt.Errorf("failed to get alm examples")
	}

	return &deviceConfigList.Items[0], nil
}
