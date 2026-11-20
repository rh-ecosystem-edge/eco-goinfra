package assisted

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	agentServiceConfigName       = "agent"
	defaultDatabaseStorageSize   = "10Gi"
	defaultFilesystemStorageSize = "20Gi"
	defaultImageStoreStorageSize = "10Gi"
	ipxeEnabled                  = "enabled"
	accessModeReadWriteOnce      = "ReadWriteOnce"
)

var (
	agentServiceConfigGVK = agentInstallV1Beta1.GroupVersion.WithKind("AgentServiceConfig")
	validIPXEOptions      = []string{ipxeEnabled, "disabled"}
)

// AgentServiceConfigBuilder provides struct for the agentserviceconfig object containing connection to
// the cluster and the agentserviceconfig definition.
type AgentServiceConfigBuilder struct {
	common.EmbeddableBuilder[agentInstallV1Beta1.AgentServiceConfig, *agentInstallV1Beta1.AgentServiceConfig]
	common.EmbeddableCreator[
		agentInstallV1Beta1.AgentServiceConfig,
		AgentServiceConfigBuilder,
		*agentInstallV1Beta1.AgentServiceConfig,
		*AgentServiceConfigBuilder,
	]
	common.EmbeddableDeleter[agentInstallV1Beta1.AgentServiceConfig, *agentInstallV1Beta1.AgentServiceConfig]
	common.EmbeddableForceUpdater[
		agentInstallV1Beta1.AgentServiceConfig,
		AgentServiceConfigBuilder,
		*agentInstallV1Beta1.AgentServiceConfig,
		*AgentServiceConfigBuilder,
	]
	common.EmbeddableWithOptions[
		agentInstallV1Beta1.AgentServiceConfig,
		AgentServiceConfigBuilder,
		*agentInstallV1Beta1.AgentServiceConfig,
		*AgentServiceConfigBuilder,
		AgentServiceConfigAdditionalOptions,
	]
}

// AgentServiceConfigAdditionalOptions additional options for AgentServiceConfig object.
type AgentServiceConfigAdditionalOptions func(builder *AgentServiceConfigBuilder) (*AgentServiceConfigBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *AgentServiceConfigBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the AgentServiceConfig GVK for this builder.
func (builder *AgentServiceConfigBuilder) GetGVK() schema.GroupVersionKind {
	return agentServiceConfigGVK
}

// NewAgentServiceConfigBuilder creates a new instance of AgentServiceConfigBuilder.
func NewAgentServiceConfigBuilder(
	apiClient *clients.Settings,
	databaseStorageSpec,
	filesystemStorageSpec corev1.PersistentVolumeClaimSpec) *AgentServiceConfigBuilder {
	builder := common.NewClusterScopedBuilder[agentInstallV1Beta1.AgentServiceConfig, AgentServiceConfigBuilder](
		apiClient, agentInstallV1Beta1.AddToScheme, agentServiceConfigName)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Spec.DatabaseStorage = databaseStorageSpec
	builder.Definition.Spec.FileSystemStorage = filesystemStorageSpec

	return builder
}

// NewDefaultAgentServiceConfigBuilder creates a new instance of AgentServiceConfigBuilder
// with default storage specs already set.
func NewDefaultAgentServiceConfigBuilder(apiClient *clients.Settings) *AgentServiceConfigBuilder {
	builder := common.NewClusterScopedBuilder[agentInstallV1Beta1.AgentServiceConfig, AgentServiceConfigBuilder](
		apiClient, agentInstallV1Beta1.AddToScheme, agentServiceConfigName)
	if builder.GetError() != nil {
		return builder
	}

	imageStorageSpec, err := GetDefaultStorageSpec(defaultImageStoreStorageSize)
	if err != nil {
		builder.SetError(fmt.Errorf("error retrieving the storage size: %w", err))

		return builder
	}

	builder.Definition.Spec.ImageStorage = &imageStorageSpec

	databaseStorageSpec, err := GetDefaultStorageSpec(defaultDatabaseStorageSize)
	if err != nil {
		builder.SetError(fmt.Errorf("error retrieving the storage size: %w", err))

		return builder
	}

	builder.Definition.Spec.DatabaseStorage = databaseStorageSpec

	fileSystemStorageSpec, err := GetDefaultStorageSpec(defaultFilesystemStorageSize)
	if err != nil {
		builder.SetError(fmt.Errorf("error retrieving the storage size: %w", err))

		return builder
	}

	builder.Definition.Spec.FileSystemStorage = fileSystemStorageSpec

	return builder
}

// WithImageStorage sets the imageStorageSpec used by the agentserviceconfig.
func (builder *AgentServiceConfigBuilder) WithImageStorage(
	imageStorageSpec corev1.PersistentVolumeClaimSpec) *AgentServiceConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting imageStorage %v in agentserviceconfig", imageStorageSpec)

	builder.Definition.Spec.ImageStorage = &imageStorageSpec

	return builder
}

// WithMirrorRegistryRef adds a configmap ref to the agentserviceconfig containing mirroring information.
func (builder *AgentServiceConfigBuilder) WithMirrorRegistryRef(configMapName string) *AgentServiceConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding mirrorRegistryRef %s to agentserviceconfig %s", configMapName, builder.Definition.Name)

	if configMapName == "" {
		builder.SetError(errors.New("cannot add agentserviceconfig mirrorRegistryRef with empty configmap name"))

		return builder
	}

	builder.Definition.Spec.MirrorRegistryRef = &corev1.LocalObjectReference{
		Name: configMapName,
	}

	return builder
}

// WithOSImage appends an OSImage to the OSImages list used by the agentserviceconfig.
func (builder *AgentServiceConfigBuilder) WithOSImage(osImage agentInstallV1Beta1.OSImage) *AgentServiceConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding OSImage %v to agentserviceconfig %s", osImage, builder.Definition.Name)

	builder.Definition.Spec.OSImages = append(builder.Definition.Spec.OSImages, osImage)

	return builder
}

// WithUnauthenticatedRegistry appends an unauthenticated registry to the agentserviceconfig.
func (builder *AgentServiceConfigBuilder) WithUnauthenticatedRegistry(registry string) *AgentServiceConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding unauthenticatedRegistry %s to agentserviceconfig %s", registry, builder.Definition.Name)

	if registry == "" {
		builder.SetError(errors.New("agentserviceconfig cannot have empty unauthenticated registry"))

		return builder
	}

	builder.Definition.Spec.UnauthenticatedRegistries = append(builder.Definition.Spec.UnauthenticatedRegistries, registry)

	return builder
}

// WithIPXEHTTPRoute sets the IPXEHTTPRoute type to be used by the agentserviceconfig.
func (builder *AgentServiceConfigBuilder) WithIPXEHTTPRoute(route string) *AgentServiceConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding IPXEHTTPRout %s to agentserviceconfig %s", route, builder.Definition.Name)

	if !slices.Contains(validIPXEOptions, route) {
		builder.SetError(fmt.Errorf(
			"agentserviceconfig passed invalid ipxeroute: %s, valid options: %v", route, validIPXEOptions))

		return builder
	}

	builder.Definition.Spec.IPXEHTTPRoute = route

	return builder
}

// WaitUntilDeployed waits the specified timeout for the agentserviceconfig to deploy.
func (builder *AgentServiceConfigBuilder) WaitUntilDeployed(timeout time.Duration) (*AgentServiceConfigBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Waiting for agentserviceconfig %s to be deployed", builder.Definition.Name)

	if !builder.Exists() {
		klog.V(100).Info("The agentserviceconfig does not exist on the cluster")

		return builder, errors.New("cannot wait for non-existent agentserviceconfig to be deployed")
	}

	// Polls every retryInterval to determine if agentserviceconfig is in desired state.
	conditionIndex := -1

	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			if conditionIndex < 0 {
				for index, condition := range builder.Object.Status.Conditions {
					if condition.Type == agentInstallV1Beta1.ConditionDeploymentsHealthy {
						conditionIndex = index
					}
				}
			}

			if conditionIndex < 0 {
				return false, nil
			}

			return builder.Object.Status.Conditions[conditionIndex].Status == corev1.ConditionTrue, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}

// PullAgentServiceConfig loads the existing agentserviceconfig into AgentServiceConfigBuilder struct.
func PullAgentServiceConfig(apiClient *clients.Settings) (*AgentServiceConfigBuilder, error) {
	return common.PullClusterScopedBuilder[agentInstallV1Beta1.AgentServiceConfig, AgentServiceConfigBuilder](
		context.TODO(), apiClient, agentInstallV1Beta1.AddToScheme, agentServiceConfigName)
}

// DeleteAndWait deletes an agentserviceconfig and waits until it is removed from the cluster.
func (builder *AgentServiceConfigBuilder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof(`Deleting agentserviceconfig %s and
		waiting for the defined period until it is removed`,
		builder.Definition.Name)

	if err := builder.Delete(); err != nil {
		return err
	}

	// Polls the agentserviceconfig every second until it is removed.
	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			_, err := builder.Get()
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			return false, nil
		})
}

// GetDefaultStorageSpec returns a default PVC spec for the respective
// agentserviceconfig component's storage and a possible error.
func GetDefaultStorageSpec(defaultStorageSize string) (corev1.PersistentVolumeClaimSpec, error) {
	checkedDefaultStorageSize, err := resource.ParseQuantity(defaultStorageSize)
	if err != nil {
		return corev1.PersistentVolumeClaimSpec{}, errors.New("the storage size is in wrong format")
	}

	defaultSpec := corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			accessModeReadWriteOnce,
		},
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: checkedDefaultStorageSize,
			},
		},
	}

	klog.V(100).Infof("Getting default PVC spec: %v", defaultSpec)

	return defaultSpec, nil
}
