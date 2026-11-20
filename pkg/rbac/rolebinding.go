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

var roleBindingGVK = rbacv1.SchemeGroupVersion.WithKind("RoleBinding")

// RoleBindingBuilder provides struct for RoleBinding object containing connection
// to the cluster RoleBinding definition.
type RoleBindingBuilder struct {
	common.EmbeddableBuilder[rbacv1.RoleBinding, *rbacv1.RoleBinding]
	common.EmbeddableCreator[
		rbacv1.RoleBinding, RoleBindingBuilder, *rbacv1.RoleBinding, *RoleBindingBuilder]
	common.EmbeddableDeleter[rbacv1.RoleBinding, *rbacv1.RoleBinding]
	common.EmbeddableUpdater[
		rbacv1.RoleBinding, RoleBindingBuilder, *rbacv1.RoleBinding, *RoleBindingBuilder]
	common.EmbeddableWithOptions[
		rbacv1.RoleBinding, RoleBindingBuilder, *rbacv1.RoleBinding, *RoleBindingBuilder,
		RoleBindingAdditionalOptions]
}

// RoleBindingAdditionalOptions additional options for RoleBinding object.
type RoleBindingAdditionalOptions func(builder *RoleBindingBuilder) (*RoleBindingBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *RoleBindingBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the RoleBinding GVK for this builder.
func (builder *RoleBindingBuilder) GetGVK() schema.GroupVersionKind {
	return roleBindingGVK
}

// NewRoleBindingBuilder creates new instance of RoleBindingBuilder.
func NewRoleBindingBuilder(apiClient *clients.Settings,
	name, nsname, role string,
	subject rbacv1.Subject) *RoleBindingBuilder {
	builder := common.NewNamespacedBuilder[rbacv1.RoleBinding, RoleBindingBuilder](
		apiClient, rbacv1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Name:     role,
		Kind:     "Role",
	}

	builder.WithSubjects([]rbacv1.Subject{subject})

	return builder
}

// WithSubjects adds specified Subject to the RoleBinding.
func (builder *RoleBindingBuilder) WithSubjects(subjects []rbacv1.Subject) *RoleBindingBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding to the rolebinding %s these specified subjects: %v",
		builder.Definition.Name, subjects)

	if len(subjects) == 0 {
		klog.V(100).Info("The list of subjects is empty")

		builder.SetError(errors.New("cannot create rolebinding with empty subject"))

		return builder
	}

	for _, subject := range subjects {
		if !slices.Contains(allowedSubjectKinds(), subject.Kind) {
			klog.V(100).Info("The rolebinding subject kind must be one of 'ServiceAccount', 'User', or 'Group'")

			builder.SetError(errors.New("rolebinding subject kind must be one of 'ServiceAccount', 'User', 'Group'"))

			return builder
		}

		if subject.Name == "" {
			klog.V(100).Info("The rolebinding subject name cannot be empty")

			builder.SetError(errors.New("rolebinding subject name cannot be empty"))

			return builder
		}
	}

	builder.Definition.Subjects = append(builder.Definition.Subjects, subjects...)

	return builder
}

// PullRoleBinding pulls existing rolebinding from cluster.
func PullRoleBinding(apiClient *clients.Settings, name, nsname string) (*RoleBindingBuilder, error) {
	return common.PullNamespacedBuilder[rbacv1.RoleBinding, RoleBindingBuilder](
		context.TODO(), apiClient, rbacv1.AddToScheme, name, nsname)
}
