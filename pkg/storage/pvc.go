package storage

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	accessModeReadWriteOnce = "ReadWriteOnce"
)

var (
	persistentVolumeClaimGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim")

	validPVCModesMap = map[string]string{
		accessModeReadWriteOnce: accessModeReadWriteOnce,
		"ReadOnlyMany":          "ReadOnlyMany",
		"ReadWriteMany":         "ReadWriteMany",
		"ReadWriteOncePod":      "ReadWriteOncePod",
	}
)

// PVCBuilder provides struct for persistentvolumeclaim object containing connection
// to the cluster and the persistentvolumeclaim definitions.
type PVCBuilder struct {
	common.EmbeddableBuilder[corev1.PersistentVolumeClaim, *corev1.PersistentVolumeClaim]
	common.EmbeddableCreator[
		corev1.PersistentVolumeClaim,
		PVCBuilder,
		*corev1.PersistentVolumeClaim,
		*PVCBuilder,
	]
	common.EmbeddableDeleter[corev1.PersistentVolumeClaim, *corev1.PersistentVolumeClaim]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *PVCBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the PersistentVolumeClaim GVK for this builder.
func (builder *PVCBuilder) GetGVK() schema.GroupVersionKind {
	return persistentVolumeClaimGVK
}

// NewPVCBuilder creates a new structure for persistentvolumeclaim.
func NewPVCBuilder(apiClient *clients.Settings, name, nsname string) *PVCBuilder {
	return common.NewNamespacedBuilder[corev1.PersistentVolumeClaim, PVCBuilder](
		apiClient, corev1.AddToScheme, name, nsname)
}

// WithPVCAccessMode configure access mode for the PV.
func (builder *PVCBuilder) WithPVCAccessMode(accessMode string) (*PVCBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Set PVC accessMode: %s", accessMode)

	if accessMode == "" {
		klog.V(100).Infof("Empty accessMode for PVC %s", builder.Definition.Name)

		return builder, fmt.Errorf("empty accessMode for PVC requested")
	}

	if !validatePVCAccessMode(accessMode) {
		klog.V(100).Infof("Invalid accessMode for PVC %s", accessMode)

		return builder, fmt.Errorf("invalid accessMode for PVC %s", accessMode)
	}

	builder.Definition.Spec.AccessModes = append(
		builder.Definition.Spec.AccessModes, corev1.PersistentVolumeAccessMode(accessMode))

	return builder, nil
}

// validatePVCAccessMode validates if requested mode is valid for PVC.
func validatePVCAccessMode(accessMode string) bool {
	klog.V(100).Infof("Validating accessMode %s", accessMode)

	_, ok := validPVCModesMap[accessMode]

	return ok
}

// WithPVCCapacity configures the minimum resources the volume should have.
func (builder *PVCBuilder) WithPVCCapacity(capacity string) (*PVCBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	if capacity == "" {
		klog.V(100).Info("Capacity of the PersistentVolumeClaim is empty")

		return builder, fmt.Errorf("capacity of the PersistentVolumeClaim is empty")
	}

	parsedCapacity, err := resource.ParseQuantity(capacity)
	if err != nil {
		klog.V(100).Infof("Failed to parse %v", capacity)

		return builder, fmt.Errorf("failed to parse: %v", capacity)
	}

	builder.Definition.Spec.Resources = corev1.VolumeResourceRequirements{
		Requests: corev1.ResourceList{corev1.ResourceStorage: parsedCapacity},
	}

	return builder, nil
}

// WithStorageClass configures storageClass required by the claim.
func (builder *PVCBuilder) WithStorageClass(storageClass string) (*PVCBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Set storage class %s for the PersistentVolumeClaim", storageClass)

	if storageClass == "" {
		klog.V(100).Infof("Empty storageClass requested for the PersistentVolumeClaim")

		return builder, fmt.Errorf("empty storageClass requested for the PersistentVolumeClaim %s",
			builder.Definition.Name)
	}

	builder.Definition.Spec.StorageClassName = &storageClass

	return builder, nil
}

// WithVolumeMode configures what type of volume is required by the claim.
func (builder *PVCBuilder) WithVolumeMode(volumeMode string) (*PVCBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Set VolumeMode %s for the PersistentVolumeClaim", volumeMode)

	if volumeMode == "" {
		klog.V(100).Infof("Empty volumeMode requested for the PersistentVolumeClaim %s in %s namespace",
			builder.Definition.Name, builder.Definition.Namespace)

		return builder, fmt.Errorf("empty volumeMode requested for the PersistentVolumeClaim %s in %s namespace",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	if !validateVolumeMode(volumeMode) {
		klog.V(100).Infof("Unsupported VolumeMode: %s", volumeMode)

		return builder, fmt.Errorf("unsupported VolumeMode %q requested for %s PersistentVolumeClaim in %s namespace",
			volumeMode, builder.Definition.Name, builder.Definition.Name)
	}

	persistentVolumeMode := corev1.PersistentVolumeMode(volumeMode)
	builder.Definition.Spec.VolumeMode = &persistentVolumeMode

	return builder, nil
}

// DeleteAndWait deletes PersistentVolumeClaim and waits until it is removed from the cluster.
func (builder *PVCBuilder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Deleting PersistenVolumeClaim %s from %s namespace and waiting for the removal to complete",
		builder.Definition.Name, builder.Definition.Namespace)

	if err := builder.Delete(); err != nil {
		klog.V(100).Infof("Failed to delete PersistentVolumeClaim %s from %s namespace",
			builder.Definition.Name, builder.Definition.Namespace)
		klog.V(100).Infof("PersistenteVolumeClaim deletion error: %v", err)

		return err
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			_, err := common.Get(ctx, builder)

			return k8serrors.IsNotFound(err), nil
		})
}

// PullPersistentVolumeClaim gets an existing PersistentVolumeClaim
// from the cluster.
func PullPersistentVolumeClaim(
	apiClient *clients.Settings, name string, nsname string) (
	*PVCBuilder, error) {
	return common.PullNamespacedBuilder[corev1.PersistentVolumeClaim, PVCBuilder](
		context.TODO(), apiClient, corev1.AddToScheme, name, nsname)
}

// GetPersistentVolumeClaimGVR returns the GroupVersionResource for the PersistentVolumeClaim.
func GetPersistentVolumeClaimGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}
}

// validateVolumeMode validates if requested volume mode is valid for PVC.
func validateVolumeMode(volumeMode string) bool {
	klog.V(100).Infof("Validating volumeMode %s", volumeMode)

	validVolumeModes := []string{
		"Block",
		"Filesystem",
	}

	return slices.Contains(validVolumeModes, volumeMode)
}
