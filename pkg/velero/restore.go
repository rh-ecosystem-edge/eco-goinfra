package velero

import (
	"context"
	"errors"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const errEmptyBackupName = "restore backupName cannot be an empty string"

var restoreGVK = velerov1.SchemeGroupVersion.WithKind("Restore")

// RestoreBuilder provides a struct for restore object from the cluster and a restore definition.
type RestoreBuilder struct {
	common.EmbeddableBuilder[velerov1.Restore, *velerov1.Restore]
	common.EmbeddableCreator[velerov1.Restore, RestoreBuilder, *velerov1.Restore, *RestoreBuilder]
	common.EmbeddableDeleteReturner[velerov1.Restore, RestoreBuilder, *velerov1.Restore, *RestoreBuilder]
	common.EmbeddableUpdater[velerov1.Restore, RestoreBuilder, *velerov1.Restore, *RestoreBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *RestoreBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the Restore GVK for this builder.
func (builder *RestoreBuilder) GetGVK() schema.GroupVersionKind {
	return restoreGVK
}

// NewRestoreBuilder creates a new instance of RestoreBuilder.
func NewRestoreBuilder(apiClient *clients.Settings, name, nsname, backupName string) *RestoreBuilder {
	klog.V(100).Infof(
		"Initializing new restore structure with the following params: "+
			"name: %s, namespace: %s, backupName: %s", name, nsname, backupName)

	builder := common.NewNamespacedBuilder[velerov1.Restore, RestoreBuilder](
		apiClient, velerov1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	if backupName == "" {
		klog.V(100).Info("The backupName of the restore is empty")

		builder.SetError(errors.New(errEmptyBackupName))

		return builder
	}

	builder.Definition.Spec.BackupName = backupName

	return builder
}

// PullRestore loads an existing restore into RestoreBuilder struct.
func PullRestore(apiClient *clients.Settings, name, nsname string) (*RestoreBuilder, error) {
	return common.PullNamespacedBuilder[velerov1.Restore, RestoreBuilder](
		context.TODO(), apiClient, velerov1.AddToScheme, name, nsname)
}

// WithStorageLocation adds a storage location to the restore.
func (builder *RestoreBuilder) WithStorageLocation(location string) *RestoreBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding storage location %s to restore %s in namespace %s",
		location, builder.Definition.Name, builder.Definition.Namespace)

	if location == "" {
		klog.V(100).Info("Backup storage location is empty")

		builder.SetError(fmt.Errorf("restore storage location cannot be an empty string"))

		return builder
	}

	if builder.Definition.Labels == nil {
		builder.Definition.Labels = make(map[string]string)
	}

	builder.Definition.Labels["velero.io/storage-location"] = location

	return builder
}
