package lso

import (
	"context"
	"fmt"

	lsov1 "github.com/openshift/local-storage-operator/api/v1"
	lsov1alpha1 "github.com/openshift/local-storage-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var localVolumeSetGVK = lsov1alpha1.GroupVersion.WithKind("LocalVolumeSet")

// LocalVolumeSetBuilder provides a struct for localVolumeSet object from the cluster and a localVolumeSet definition.
type LocalVolumeSetBuilder struct {
	common.EmbeddableBuilder[lsov1alpha1.LocalVolumeSet, *lsov1alpha1.LocalVolumeSet]
	common.EmbeddableCreator[
		lsov1alpha1.LocalVolumeSet, LocalVolumeSetBuilder, *lsov1alpha1.LocalVolumeSet, *LocalVolumeSetBuilder]
	common.EmbeddableDeleter[lsov1alpha1.LocalVolumeSet, *lsov1alpha1.LocalVolumeSet]
	common.EmbeddableUpdater[
		lsov1alpha1.LocalVolumeSet, LocalVolumeSetBuilder, *lsov1alpha1.LocalVolumeSet, *LocalVolumeSetBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *LocalVolumeSetBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the LocalVolumeSet GVK for this builder.
func (builder *LocalVolumeSetBuilder) GetGVK() schema.GroupVersionKind {
	return localVolumeSetGVK
}

// NewLocalVolumeSetBuilder creates new instance of LocalVolumeSetBuilder.
func NewLocalVolumeSetBuilder(apiClient *clients.Settings, name, nsname string) *LocalVolumeSetBuilder {
	return common.NewNamespacedBuilder[lsov1alpha1.LocalVolumeSet, LocalVolumeSetBuilder](
		apiClient, lsov1alpha1.AddToScheme, name, nsname)
}

// PullLocalVolumeSet retrieves an existing localVolumeSet object from the cluster.
func PullLocalVolumeSet(apiClient *clients.Settings, name, nsname string) (*LocalVolumeSetBuilder, error) {
	return common.PullNamespacedBuilder[lsov1alpha1.LocalVolumeSet, LocalVolumeSetBuilder](
		context.TODO(), apiClient, lsov1alpha1.AddToScheme, name, nsname)
}

// WithTolerations sets the localVolumeSet's tolerations.
func (builder *LocalVolumeSetBuilder) WithTolerations(
	tolerations []corev1.Toleration) *LocalVolumeSetBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding tolerations %v to localVolumeSet %s in namespace %s",
		tolerations, builder.Definition.Name, builder.Definition.Namespace)

	if len(tolerations) == 0 {
		klog.V(100).Info("The tolerations list is empty")

		builder.SetError(fmt.Errorf(errEmptyTolerations))

		return builder
	}

	builder.Definition.Spec.Tolerations = tolerations

	return builder
}

// WithNodeSelector sets the localVolumeSet's nodeSelector.
func (builder *LocalVolumeSetBuilder) WithNodeSelector(
	nodeSelector corev1.NodeSelector) *LocalVolumeSetBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding nodeSelector %v to localVolumeSet %s in namespace %s",
		nodeSelector, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.NodeSelector = &nodeSelector

	return builder
}

// WithStorageClassName sets the localVolumeSet's storageClassName.
func (builder *LocalVolumeSetBuilder) WithStorageClassName(
	storageClassName string) *LocalVolumeSetBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding storageClassName %s to localVolumeSet %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace, storageClassName)

	if storageClassName == "" {
		klog.V(100).Info("The storageClassName is empty")

		builder.SetError(fmt.Errorf("'storageClassName' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.StorageClassName = storageClassName

	return builder
}

// WithVolumeMode sets the localVolumeSet's volumeMode.
func (builder *LocalVolumeSetBuilder) WithVolumeMode(
	volumeMode lsov1.PersistentVolumeMode) *LocalVolumeSetBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding volumeMode %v to localVolumeSet %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace, volumeMode)

	if volumeMode == "" {
		klog.V(100).Info("The volumeMode is empty")

		builder.SetError(fmt.Errorf("'volumeMode' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.VolumeMode = volumeMode

	return builder
}

// WithFSType sets the localVolumeSet's fstype.
func (builder *LocalVolumeSetBuilder) WithFSType(
	fstype string) *LocalVolumeSetBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding fstype %s to localVolumeSet %s in namespace %s",
		fstype, builder.Definition.Name, builder.Definition.Namespace)

	if fstype == "" {
		klog.V(100).Info("The fstype is empty")

		builder.SetError(fmt.Errorf("'fstype' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.FSType = fstype

	return builder
}

// WithMaxDeviceCount sets the localVolumeSet's maxDeviceCount.
func (builder *LocalVolumeSetBuilder) WithMaxDeviceCount(
	maxDeviceCount int32) *LocalVolumeSetBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding maxDeviceCount %v to localVolumeSet %s in namespace %s",
		maxDeviceCount, builder.Definition.Name, builder.Definition.Namespace)

	if maxDeviceCount == 0 {
		klog.V(100).Info("The maxDeviceCount is zero")

		builder.SetError(fmt.Errorf("'maxDeviceCount' argument cannot be equal zero"))

		return builder
	}

	builder.Definition.Spec.MaxDeviceCount = &maxDeviceCount

	return builder
}

// WithDeviceInclusionSpec sets the localVolumeSet's deviceInclusionSpec.
func (builder *LocalVolumeSetBuilder) WithDeviceInclusionSpec(
	deviceInclusionSpec lsov1alpha1.DeviceInclusionSpec) *LocalVolumeSetBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding deviceInclusionSpec %v to localVolumeSet %s in namespace %s",
		deviceInclusionSpec, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.DeviceInclusionSpec = &deviceInclusionSpec

	return builder
}
