package velero

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var backupGVK = velerov1.SchemeGroupVersion.WithKind("Backup")

// BackupBuilder provides a struct for backup object from the cluster and a backup definition.
type BackupBuilder struct {
	common.EmbeddableBuilder[velerov1.Backup, *velerov1.Backup]
	common.EmbeddableCreator[velerov1.Backup, BackupBuilder, *velerov1.Backup, *BackupBuilder]
	common.EmbeddableDeleteReturner[velerov1.Backup, BackupBuilder, *velerov1.Backup, *BackupBuilder]
	common.EmbeddableUpdater[velerov1.Backup, BackupBuilder, *velerov1.Backup, *BackupBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *BackupBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the Backup GVK for this builder.
func (builder *BackupBuilder) GetGVK() schema.GroupVersionKind {
	return backupGVK
}

// NewBackupBuilder creates a new instance of BackupBuilder.
func NewBackupBuilder(apiClient *clients.Settings, name, nsname string) *BackupBuilder {
	return common.NewNamespacedBuilder[velerov1.Backup, BackupBuilder](
		apiClient, velerov1.AddToScheme, name, nsname)
}

// PullBackup loads an existing backup into BackupBuilder struct.
func PullBackup(apiClient *clients.Settings, name, nsname string) (*BackupBuilder, error) {
	return common.PullNamespacedBuilder[velerov1.Backup, BackupBuilder](
		context.TODO(), apiClient, velerov1.AddToScheme, name, nsname)
}

// WithStorageLocation adds a storage location to the backup.
func (builder *BackupBuilder) WithStorageLocation(location string) *BackupBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding storage location %s to backup %s in namespace %s",
		location, builder.Definition.Name, builder.Definition.Namespace)

	if location == "" {
		klog.V(100).Info("Backup storage location is empty")

		builder.SetError(fmt.Errorf("backup storage location cannot be an empty string"))

		return builder
	}

	if builder.Definition.Labels == nil {
		builder.Definition.Labels = make(map[string]string)
	}

	builder.Definition.Labels["velero.io/storage-location"] = location

	return builder
}

// WithIncludedNamespace adds the specified namespace for inclusion when performing a backup.
func (builder *BackupBuilder) WithIncludedNamespace(namespace string) *BackupBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding namespace %s to backup %s in namespace %s includedNamespaces field",
		namespace, builder.Definition.Name, builder.Definition.Namespace)

	if namespace == "" {
		klog.V(100).Info("Backup includedNamespace is empty")

		builder.SetError(fmt.Errorf("backup includedNamespace cannot be an empty string"))

		return builder
	}

	builder.Definition.Spec.IncludedNamespaces = append(builder.Definition.Spec.IncludedNamespaces, namespace)

	return builder
}

// WithIncludedClusterScopedResource adds the specified cluster-scoped crd for inclusion when performing a backup.
func (builder *BackupBuilder) WithIncludedClusterScopedResource(crd string) *BackupBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding custom resource %s to backup %s in namespace %s includedClusterScopedResources field",
		crd, builder.Definition.Name, builder.Definition.Namespace)

	if crd == "" {
		klog.V(100).Info("Backup includedClusterScopedResource is empty")

		builder.SetError(fmt.Errorf("backup includedClusterScopedResource cannot be an empty string"))

		return builder
	}

	builder.Definition.Spec.IncludedClusterScopedResources =
		append(builder.Definition.Spec.IncludedClusterScopedResources, crd)

	return builder
}

// WithIncludedNamespaceScopedResource adds the specified namespace-scoped crd for inclusion when performing a backup.
func (builder *BackupBuilder) WithIncludedNamespaceScopedResource(crd string) *BackupBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding custom resource %s to backup %s in namespace %s includedNamespaceScopedResources field",
		crd, builder.Definition.Name, builder.Definition.Namespace)

	if crd == "" {
		klog.V(100).Info("Backup includedNamespaceScopedResource is empty")

		builder.SetError(fmt.Errorf("backup includedNamespaceScopedResource cannot be an empty string"))

		return builder
	}

	builder.Definition.Spec.IncludedNamespaceScopedResources =
		append(builder.Definition.Spec.IncludedNamespaceScopedResources, crd)

	return builder
}

// WithExcludedClusterScopedResource adds the specified cluster-scoped crd for exclusion when performing a backup.
func (builder *BackupBuilder) WithExcludedClusterScopedResource(crd string) *BackupBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding custom resource %s to backup %s in namespace %s excludedClusterScopedResources field",
		crd, builder.Definition.Name, builder.Definition.Namespace)

	if crd == "" {
		klog.V(100).Info("Backup excludedClusterScopedResource is empty")

		builder.SetError(fmt.Errorf("backup excludedClusterScopedResource cannot be an empty string"))

		return builder
	}

	builder.Definition.Spec.ExcludedClusterScopedResources =
		append(builder.Definition.Spec.ExcludedClusterScopedResources, crd)

	return builder
}

// WithExcludedNamespaceScopedResources adds the specified namespace-scoped crd for exclusion when performing a backup.
func (builder *BackupBuilder) WithExcludedNamespaceScopedResources(crd string) *BackupBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding custom resource %s to backup %s in namespace %s excludedNamespaceScopedResources field",
		crd, builder.Definition.Name, builder.Definition.Namespace)

	if crd == "" {
		klog.V(100).Info("Backup excludedNamespaceScopedResource is empty")

		builder.SetError(fmt.Errorf("backup excludedNamespaceScopedResource cannot be an empty string"))

		return builder
	}

	builder.Definition.Spec.ExcludedNamespaceScopedResources =
		append(builder.Definition.Spec.ExcludedNamespaceScopedResources, crd)

	return builder
}
