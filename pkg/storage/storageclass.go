package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	corev1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const errEmptyProvisioner = "storageclass 'provisioner' cannot be empty"

var storageClassGVK = storageV1.SchemeGroupVersion.WithKind("StorageClass")

// ClassBuilder provides struct for storageclass object containing
// connection to the cluster and the storageclass definitions.
type ClassBuilder struct {
	common.EmbeddableBuilder[storageV1.StorageClass, *storageV1.StorageClass]
	common.EmbeddableCreator[storageV1.StorageClass, ClassBuilder, *storageV1.StorageClass, *ClassBuilder]
	common.EmbeddableDeleter[storageV1.StorageClass, *storageV1.StorageClass]
	common.EmbeddableForceUpdater[storageV1.StorageClass, ClassBuilder, *storageV1.StorageClass, *ClassBuilder]
	common.EmbeddableWithOptions[
		storageV1.StorageClass, ClassBuilder, *storageV1.StorageClass, *ClassBuilder, AdditionalOptions]
}

// AdditionalOptions additional options for storageclass object.
type AdditionalOptions func(builder *ClassBuilder) (*ClassBuilder, error)

// AttachMixins wires the embedded CRUD and options mixins to this builder instance.
func (builder *ClassBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the StorageClass GVK for this builder.
func (builder *ClassBuilder) GetGVK() schema.GroupVersionKind {
	return storageClassGVK
}

// NewClassBuilder creates a new instance of ClassBuilder.
func NewClassBuilder(apiClient *clients.Settings, name, provisioner string) *ClassBuilder {
	builder := common.NewClusterScopedBuilder[storageV1.StorageClass, ClassBuilder](
		apiClient, storageV1.AddToScheme, name)
	builder.Definition.Provisioner = provisioner

	if builder.GetError() != nil {
		return builder
	}

	if provisioner == "" {
		klog.V(100).Info("The provisioner of the storageclass is empty")

		builder.SetError(fmt.Errorf(errEmptyProvisioner))

		return builder
	}

	return builder
}

// WithReclaimPolicy adds a reclaimPolicy to the storageclass definition.
func (builder *ClassBuilder) WithReclaimPolicy(
	reclaimPolicy corev1.PersistentVolumeReclaimPolicy) *ClassBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if reclaimPolicy == "" {
		klog.V(100).Info("The reclaimPolicy of the storageclass is empty")

		builder.SetError(fmt.Errorf("storageclass 'reclaimPolicy' cannot be empty"))

		return builder
	}

	builder.Definition.ReclaimPolicy = &reclaimPolicy

	return builder
}

// WithVolumeBindingMode adds a volumeBindingMode to the storage class definition.
func (builder *ClassBuilder) WithVolumeBindingMode(
	bindingMode storageV1.VolumeBindingMode) *ClassBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if bindingMode == "" {
		klog.V(100).Info("The bindingMode of the storageclass is empty")

		builder.SetError(fmt.Errorf("storageclass 'bindingMode' cannot be empty"))

		return builder
	}

	builder.Definition.VolumeBindingMode = &bindingMode

	return builder
}

// WithParameter adds a parameter to the storage class definition.
func (builder *ClassBuilder) WithParameter(parameterKey, parameterValue string) *ClassBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if parameterKey == "" {
		klog.V(100).Info("The parameter key of the storageclass is empty")

		builder.SetError(fmt.Errorf("storageclass parameter key cannot be empty"))

		return builder
	}

	if parameterValue == "" {
		klog.V(100).Info("The parameter value of the storageclass is empty")

		builder.SetError(fmt.Errorf("storageclass parameter value cannot be empty"))

		return builder
	}

	if builder.Definition.Parameters == nil {
		builder.Definition.Parameters = make(map[string]string)
	}

	builder.Definition.Parameters[parameterKey] = parameterValue

	return builder
}

// PullClass pulls an existing storage class into a ClassBuilder struct.
func PullClass(apiClient *clients.Settings, name string) (*ClassBuilder, error) {
	return common.PullClusterScopedBuilder[storageV1.StorageClass, ClassBuilder](
		context.TODO(), apiClient, storageV1.AddToScheme, name)
}

// DeleteAndWait deletes the StorageClass and waits up to timeout until it has been removed.
func (builder *ClassBuilder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof(
		"Deleting StorageClass %s and waiting up to %s until it is removed", builder.Definition.Name, timeout)

	err := builder.Delete()
	if err != nil {
		return err
	}

	return builder.WaitUntilDeleted(timeout)
}

// WaitUntilDeleted waits for the duration of timeout or until the StorageClass has been deleted.
func (builder *ClassBuilder) WaitUntilDeleted(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Waiting up to %s until StorageClass %s is deleted", timeout, builder.Definition.Name)

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			err := builder.GetClient().Get(
				logging.WithDiscardLogger(ctx),
				runtimeclient.ObjectKeyFromObject(builder.Definition),
				new(storageV1.StorageClass),
			)
			if err == nil {
				klog.V(100).Infof("StorageClass %s still present", builder.Definition.Name)

				return false, nil
			}

			if k8serrors.IsNotFound(err) {
				klog.V(100).Infof("StorageClass %s is gone", builder.Definition.Name)

				return true, nil
			}

			klog.V(100).Infof("failed to get StorageClass %s", builder.Definition.Name)

			return false, err
		})
}
