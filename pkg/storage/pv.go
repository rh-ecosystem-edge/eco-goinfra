package storage

import (
	"context"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var persistentVolumeGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolume")

// PVBuilder provides struct for persistentvolume object containing connection
// to the cluster and the persistentvolume definitions.
type PVBuilder struct {
	common.EmbeddableBuilder[corev1.PersistentVolume, *corev1.PersistentVolume]
	common.EmbeddableDeleter[corev1.PersistentVolume, *corev1.PersistentVolume]
}

// AttachMixins wires the embedded deletion mixin to this builder instance.
func (builder *PVBuilder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the PersistentVolume GVK for this builder.
func (builder *PVBuilder) GetGVK() schema.GroupVersionKind {
	return persistentVolumeGVK
}

// PullPersistentVolume gets an existing PersistentVolume from the cluster.
func PullPersistentVolume(apiClient *clients.Settings, name string) (*PVBuilder, error) {
	return common.PullClusterScopedBuilder[corev1.PersistentVolume, PVBuilder](
		context.TODO(), apiClient, corev1.AddToScheme, name)
}

// DeleteAndWait deletes the PersistentVolume and waits up to timeout until it has been removed.
func (builder *PVBuilder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof(
		"Deleting PersistentVolume %s and waiting up to %s until it is removed", timeout, builder.Definition.Name)

	err := builder.Delete()
	if err != nil {
		return err
	}

	return builder.WaitUntilDeleted(timeout)
}

// WaitUntilDeleted waits for the duration of timeout or until the PersistentVolume has been deleted.
func (builder *PVBuilder) WaitUntilDeleted(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Waiting up to %s until PersistentVolume %s is deleted", timeout, builder.Definition.Name)

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			_, err := builder.Get()
			if err == nil {
				klog.V(100).Infof("PersistentVolume %s still present", builder.Definition.Name)

				return false, nil
			}

			if k8serrors.IsNotFound(err) {
				klog.V(100).Infof("PersistentVolume %s is gone", builder.Definition.Name)

				return true, nil
			}

			klog.V(100).Infof("failed to get PersistentVolume %s", builder.Definition.Name)

			return false, err
		})
}
