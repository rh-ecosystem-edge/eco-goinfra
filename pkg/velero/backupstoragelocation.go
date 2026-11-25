package velero

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var backupStorageLocationGVK = velerov1.SchemeGroupVersion.WithKind("BackupStorageLocation")

// BackupStorageLocationBuilder provides a struct for backupstoragelocation
// object from the cluster and a backupstoragelocation definition.
type BackupStorageLocationBuilder struct {
	common.EmbeddableBuilder[velerov1.BackupStorageLocation, *velerov1.BackupStorageLocation]
	common.EmbeddableCreator[
		velerov1.BackupStorageLocation,
		BackupStorageLocationBuilder,
		*velerov1.BackupStorageLocation,
		*BackupStorageLocationBuilder,
	]
	common.EmbeddableDeleter[velerov1.BackupStorageLocation, *velerov1.BackupStorageLocation]
	common.EmbeddableUpdater[
		velerov1.BackupStorageLocation,
		BackupStorageLocationBuilder,
		*velerov1.BackupStorageLocation,
		*BackupStorageLocationBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *BackupStorageLocationBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the BackupStorageLocation GVK for this builder.
func (builder *BackupStorageLocationBuilder) GetGVK() schema.GroupVersionKind {
	return backupStorageLocationGVK
}

// NewBackupStorageLocationBuilder creates a new instance of BackupStorageLocationBuilder.
func NewBackupStorageLocationBuilder(
	apiClient *clients.Settings,
	name string,
	namespace string,
	provider string,
	objectStorage velerov1.ObjectStorageLocation) *BackupStorageLocationBuilder {
	builder := common.NewNamespacedBuilder[velerov1.BackupStorageLocation, BackupStorageLocationBuilder](
		apiClient, velerov1.AddToScheme, name, namespace)
	if builder.GetError() != nil {
		return builder
	}

	if provider == "" {
		klog.V(100).Info("The provider of the backupstoragelocation is empty")

		builder.SetError(fmt.Errorf("backupstoragelocation provider cannot be empty"))

		return builder
	}

	if objectStorage.Bucket == "" {
		klog.V(100).Info("The objectstorage bucket of the backupstoragelocation is empty")

		builder.SetError(fmt.Errorf("backupstoragelocation objectstorage bucket cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Provider = provider
	builder.Definition.Spec.ObjectStorage = &objectStorage

	return builder
}

// PullBackupStorageLocationBuilder pulls existing backupstoragelocation from cluster.
func PullBackupStorageLocationBuilder(
	apiClient *clients.Settings, name, namespace string) (*BackupStorageLocationBuilder, error) {
	return common.PullNamespacedBuilder[velerov1.BackupStorageLocation, BackupStorageLocationBuilder](
		context.TODO(), apiClient, velerov1.AddToScheme, name, namespace)
}

// WithConfig includes the provided configuration to the backupstoragelocation.
func (builder *BackupStorageLocationBuilder) WithConfig(config map[string]string) *BackupStorageLocationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if len(config) == 0 {
		klog.V(100).Info("The config of the backupstoragelocation is empty")

		builder.SetError(fmt.Errorf("backupstoragelocation cannot have empty config"))

		return builder
	}

	builder.Definition.Spec.Config = config

	return builder
}

// WaitUntilAvailable waits the specified timeout for the backupstoragelocation to become available.
func (builder *BackupStorageLocationBuilder) WaitUntilAvailable(
	timeout time.Duration) (*BackupStorageLocationBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	if !builder.Exists() {
		return builder, fmt.Errorf("cannot wait for backupstoragelocation that does not exist")
	}

	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			klog.V(100).Infof("Waiting for the backupstoragelocation %s in %s to become available",
				builder.Definition.Name, builder.Definition.Namespace)

			builder.Object, err = builder.Get()
			if err != nil {
				return false, err
			}

			return builder.Object.Status.Phase == velerov1.BackupStorageLocationPhaseAvailable, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, fmt.Errorf("error waiting for backupstoragelocation to become available: %w", err)
}

// WaitUntilUnavailable waits the specified timeout for the backupstoragelocation to become unavailable.
func (builder *BackupStorageLocationBuilder) WaitUntilUnavailable(
	timeout time.Duration) (*BackupStorageLocationBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	if !builder.Exists() {
		return builder, fmt.Errorf("cannot wait for backupstoragelocation that does not exist")
	}

	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			klog.V(100).Infof("Waiting for the backupstoragelocation %s in %s to become unavailable",
				builder.Definition.Name, builder.Definition.Namespace)

			builder.Object, err = builder.Get()
			if err != nil {
				return false, err
			}

			return builder.Object.Status.Phase == velerov1.BackupStorageLocationPhaseUnavailable, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, fmt.Errorf("error waiting for backupstoragelocation to become unavailable: %w", err)
}
