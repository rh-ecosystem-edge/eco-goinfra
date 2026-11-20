package rbac

import (
	"context"
	"errors"
	"slices"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var clusterRoleBindingGVK = rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding")

// ClusterRoleBindingBuilder provides struct for clusterrolebinding object
// containing connection to the cluster and the clusterrolebinding definitions.
type ClusterRoleBindingBuilder struct {
	common.EmbeddableBuilder[rbacv1.ClusterRoleBinding, *rbacv1.ClusterRoleBinding]
	common.EmbeddableCreator[
		rbacv1.ClusterRoleBinding, ClusterRoleBindingBuilder, *rbacv1.ClusterRoleBinding, *ClusterRoleBindingBuilder]
	common.EmbeddableDeleter[rbacv1.ClusterRoleBinding, *rbacv1.ClusterRoleBinding]
	common.EmbeddableUpdater[
		rbacv1.ClusterRoleBinding, ClusterRoleBindingBuilder, *rbacv1.ClusterRoleBinding, *ClusterRoleBindingBuilder]
	common.EmbeddableWithOptions[
		rbacv1.ClusterRoleBinding, ClusterRoleBindingBuilder, *rbacv1.ClusterRoleBinding, *ClusterRoleBindingBuilder,
		ClusterRoleBindingAdditionalOptions]
}

// ClusterRoleBindingAdditionalOptions additional options for ClusterRoleBinding object.
type ClusterRoleBindingAdditionalOptions func(
	builder *ClusterRoleBindingBuilder) (*ClusterRoleBindingBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ClusterRoleBindingBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the ClusterRoleBinding GVK for this builder.
func (builder *ClusterRoleBindingBuilder) GetGVK() schema.GroupVersionKind {
	return clusterRoleBindingGVK
}

// NewClusterRoleBindingBuilder creates a new instance of ClusterRoleBindingBuilder.
func NewClusterRoleBindingBuilder(
	apiClient *clients.Settings, name, clusterRole string, subject rbacv1.Subject) *ClusterRoleBindingBuilder {
	klog.V(100).Infof(
		"Initializing new clusterrolebinding structure with the following params: "+
			"name: %s, clusterrole: %s, subject %v",
		name, clusterRole, subject)

	builder := common.NewClusterScopedBuilder[rbacv1.ClusterRoleBinding, ClusterRoleBindingBuilder](
		apiClient, rbacv1.AddToScheme, name)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Name:     clusterRole,
		Kind:     "ClusterRole",
	}

	builder.WithSubjects([]rbacv1.Subject{subject})

	return builder
}

// WithSubjects appends additional subjects to clusterrolebinding definition.
func (builder *ClusterRoleBindingBuilder) WithSubjects(subjects []rbacv1.Subject) *ClusterRoleBindingBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Appending to the definition of clusterrolebinding %s these additional subjects %v",
		builder.Definition.Name, subjects)

	if len(subjects) == 0 {
		klog.V(100).Info("The list of subjects is empty")

		builder.SetError(errors.New("cannot accept nil or empty slice as subjects"))

		return builder
	}

	for _, subject := range subjects {
		if !slices.Contains(allowedSubjectKinds(), subject.Kind) {
			klog.V(100).Info("The clusterrolebinding subject kind must be one of 'ServiceAccount', 'User', or 'Group'")

			builder.SetError(errors.New("clusterrolebinding subject kind must be one of 'ServiceAccount', 'User', or 'Group'"))

			return builder
		}

		if subject.Name == "" {
			klog.V(100).Info("The clusterrolebinding subject name cannot be empty")

			builder.SetError(errors.New("clusterrolebinding subject name cannot be empty"))

			return builder
		}
	}

	builder.Definition.Subjects = append(builder.Definition.Subjects, subjects...)

	return builder
}

// PullClusterRoleBinding pulls existing clusterrolebinding from cluster.
func PullClusterRoleBinding(apiClient *clients.Settings, name string) (*ClusterRoleBindingBuilder, error) {
	klog.V(100).Infof("Pulling existing clusterrolebinding name %s from cluster", name)

	return common.PullClusterScopedBuilder[rbacv1.ClusterRoleBinding, ClusterRoleBindingBuilder](
		context.TODO(), apiClient, rbacv1.AddToScheme, name)
}
