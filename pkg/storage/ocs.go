package storage

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	ocsoperatorv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ocs/operatorv1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var storageClusterGVK = schema.GroupVersion{Group: APIGroup, Version: APIVersion}.WithKind(StorageClusterKind)

// StorageClusterBuilder provides struct for StorageCluster object containing connection
// to the cluster and the storageCluster definitions.
type StorageClusterBuilder struct {
	common.EmbeddableBuilder[ocsoperatorv1.StorageCluster, *ocsoperatorv1.StorageCluster]
	common.EmbeddableCreator[
		ocsoperatorv1.StorageCluster, StorageClusterBuilder, *ocsoperatorv1.StorageCluster, *StorageClusterBuilder]
	common.EmbeddableDeleter[ocsoperatorv1.StorageCluster, *ocsoperatorv1.StorageCluster]
	common.EmbeddableUpdater[
		ocsoperatorv1.StorageCluster, StorageClusterBuilder, *ocsoperatorv1.StorageCluster, *StorageClusterBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *StorageClusterBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the StorageCluster GVK for this builder.
func (builder *StorageClusterBuilder) GetGVK() schema.GroupVersionKind {
	return storageClusterGVK
}

// NewStorageClusterBuilder creates a new instance of StorageClusterBuilder.
func NewStorageClusterBuilder(apiClient *clients.Settings, name, nsname string) *StorageClusterBuilder {
	return common.NewNamespacedBuilder[ocsoperatorv1.StorageCluster, StorageClusterBuilder](
		apiClient, ocsoperatorv1.AddToScheme, name, nsname)
}

// PullStorageCluster gets an existing storageCluster object from the cluster.
func PullStorageCluster(apiClient *clients.Settings, name, namespace string) (*StorageClusterBuilder, error) {
	return common.PullNamespacedBuilder[ocsoperatorv1.StorageCluster, StorageClusterBuilder](
		context.TODO(), apiClient, ocsoperatorv1.AddToScheme, name, namespace)
}

// GetManagedResources fetches storageCluster managedResources value.
func (builder *StorageClusterBuilder) GetManagedResources() (*ocsoperatorv1.ManagedResourcesSpec, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting storageCluster %s in namespace %s managedResources configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil, fmt.Errorf("storageCluster object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return &builder.Object.Spec.ManagedResources, nil
}

// GetMonDataDirHostPath fetches storageCluster monDataDirHostPath value.
func (builder *StorageClusterBuilder) GetMonDataDirHostPath() (string, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Infof("Getting storageCluster %s in namespace %s monDataDirHostPath configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return "", fmt.Errorf("storageCluster object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return builder.Object.Spec.MonDataDirHostPath, nil
}

// GetMultiCloudGateway fetches storageCluster multiCloudGateway value.
func (builder *StorageClusterBuilder) GetMultiCloudGateway() (*ocsoperatorv1.MultiCloudGatewaySpec, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting storageCluster %s in namespace %s multiCloudGateway configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil, fmt.Errorf("storageCluster object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return builder.Object.Spec.MultiCloudGateway, nil
}

// GetStorageDeviceSets fetches storageCluster storageDeviceSets value.
func (builder *StorageClusterBuilder) GetStorageDeviceSets() ([]ocsoperatorv1.StorageDeviceSet, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting storageCluster %s in namespace %s storageDeviceSets configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil, fmt.Errorf("storageCluster object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return builder.Object.Spec.StorageDeviceSets, nil
}

// WithFlexibleScaling sets the storageCluster's flexibleScaling value.
func (builder *StorageClusterBuilder) WithFlexibleScaling(flexibleScaling bool) *StorageClusterBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with flexibleScaling value: %t",
		builder.Definition.Name, builder.Definition.Namespace, flexibleScaling)

	builder.Definition.Spec.FlexibleScaling = flexibleScaling

	return builder
}

// WithManagedResources sets the storageCluster's managedResources value.
func (builder *StorageClusterBuilder) WithManagedResources(
	expectedManagedResources ocsoperatorv1.ManagedResourcesSpec) *StorageClusterBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with managedResources value: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedManagedResources)

	builder.Definition.Spec.ManagedResources = expectedManagedResources

	return builder
}

// WithMonDataDirHostPath sets the storageCluster's monDataDirHostPath value.
func (builder *StorageClusterBuilder) WithMonDataDirHostPath(
	expectedMonDataDirHostPath string) *StorageClusterBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with monDataDirHostPath value: %s",
		builder.Definition.Name, builder.Definition.Namespace, expectedMonDataDirHostPath)

	if expectedMonDataDirHostPath == "" {
		klog.V(100).Info("the expectedMonDataDirHostPath can not be empty")

		builder.SetError(fmt.Errorf("the expectedMonDataDirHostPath can not be empty"))

		return builder
	}

	builder.Definition.Spec.MonDataDirHostPath = expectedMonDataDirHostPath

	return builder
}

// WithMultiCloudGateway sets the storageCluster's multiCloudGateway value.
func (builder *StorageClusterBuilder) WithMultiCloudGateway(
	expectedMultiCloudGateway ocsoperatorv1.MultiCloudGatewaySpec) *StorageClusterBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with multiCloudGateway value: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedMultiCloudGateway)

	builder.Definition.Spec.MultiCloudGateway = &expectedMultiCloudGateway

	return builder
}

// WithStorageDeviceSet sets the storageCluster's storageDeviceSets value.
func (builder *StorageClusterBuilder) WithStorageDeviceSet(
	expectedStorageDeviceSet ocsoperatorv1.StorageDeviceSet) *StorageClusterBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with storageDeviceSets value: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedStorageDeviceSet)

	builder.Definition.Spec.StorageDeviceSets =
		append(builder.Definition.Spec.StorageDeviceSets, expectedStorageDeviceSet)

	return builder
}

// WithAnnotations sets the storageCluster's annotations value.
func (builder *StorageClusterBuilder) WithAnnotations(
	annotations map[string]string) *StorageClusterBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with annotations values: %v",
		builder.Definition.Name, builder.Definition.Namespace, annotations)

	if len(annotations) == 0 {
		klog.V(100).Info("'annotations' argument cannot be empty")

		builder.SetError(fmt.Errorf("'annotations' argument cannot be empty"))

		return builder
	}

	builder.Definition.Annotations = annotations

	return builder
}
