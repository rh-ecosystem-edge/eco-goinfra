package oadp

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	oadpv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/oadp/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var dataProtectionApplicationGVK = oadpv1alpha1.GroupVersion.WithKind("DataProtectionApplication")

// DPABuilder provides a struct for backup object from the cluster and a backup definition.
type DPABuilder struct {
	common.EmbeddableBuilder[oadpv1alpha1.DataProtectionApplication, *oadpv1alpha1.DataProtectionApplication]
	common.EmbeddableCreator[
		oadpv1alpha1.DataProtectionApplication, DPABuilder, *oadpv1alpha1.DataProtectionApplication, *DPABuilder]
	common.EmbeddableDeleter[oadpv1alpha1.DataProtectionApplication, *oadpv1alpha1.DataProtectionApplication]
	common.EmbeddableForceUpdater[
		oadpv1alpha1.DataProtectionApplication, DPABuilder, *oadpv1alpha1.DataProtectionApplication, *DPABuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *DPABuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the DataProtectionApplication GVK for this builder.
func (builder *DPABuilder) GetGVK() schema.GroupVersionKind {
	return dataProtectionApplicationGVK
}

// NewDPABuilder creates a new instance of DPABuilder.
func NewDPABuilder(
	apiClient *clients.Settings, name, namespace string, config oadpv1alpha1.ApplicationConfig) *DPABuilder {
	builder := common.NewNamespacedBuilder[oadpv1alpha1.DataProtectionApplication, DPABuilder](
		apiClient, oadpv1alpha1.AddToScheme, name, namespace)
	if builder.GetError() != nil {
		return builder
	}

	if config.Velero == nil {
		klog.V(100).Info("The velero config of the dataprotectionapplication is empty")

		builder.SetError(fmt.Errorf("dataprotectionapplication velero config cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Configuration = &config

	return builder
}

// PullDPA pulls existing dataprotectionapplication from cluster.
func PullDPA(apiClient *clients.Settings, name, nsname string) (*DPABuilder, error) {
	return common.PullNamespacedBuilder[oadpv1alpha1.DataProtectionApplication, DPABuilder](
		context.TODO(), apiClient, oadpv1alpha1.AddToScheme, name, nsname)
}

// WithBackupLocation configures the dataprotectionapplication with the specified backup location.
func (builder *DPABuilder) WithBackupLocation(backupLocation oadpv1alpha1.BackupLocation) *DPABuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding backuplocation to dataprotectionapplication %s in namespace %s: %v",
		builder.Definition.Name, builder.Definition.Namespace, backupLocation)

	if backupLocation.Velero == nil {
		klog.V(100).Info("The backuplocation velero config of the dataprotectionapplication is empty")

		builder.SetError(fmt.Errorf("dataprotectionapplication backuplocation cannot have empty velero config"))

		return builder
	}

	builder.Definition.Spec.BackupLocations = append(builder.Definition.Spec.BackupLocations, backupLocation)

	return builder
}
