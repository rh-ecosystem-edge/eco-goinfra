package rbac

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const roleMustContainAtLeastOne = "role must contain at least one Verb"

var roleGVK = rbacv1.SchemeGroupVersion.WithKind("Role")

// RoleBuilder provides a struct for role object containing connection to the cluster and the role definitions.
type RoleBuilder struct {
	common.EmbeddableBuilder[rbacv1.Role, *rbacv1.Role]
	common.EmbeddableCreator[rbacv1.Role, RoleBuilder, *rbacv1.Role, *RoleBuilder]
	common.EmbeddableDeleter[rbacv1.Role, *rbacv1.Role]
	common.EmbeddableUpdater[rbacv1.Role, RoleBuilder, *rbacv1.Role, *RoleBuilder]
	common.EmbeddableWithOptions[rbacv1.Role, RoleBuilder, *rbacv1.Role, *RoleBuilder, RoleAdditionalOptions]

	// errorMsg mirrors the common builder error for the baseline ClusterRole tests that still reference RoleBuilder.
	// Role validation and modifiers use the embedded common builder error.
	errorMsg string
}

// RoleAdditionalOptions additional options for Role object.
type RoleAdditionalOptions func(builder *RoleBuilder) (*RoleBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *RoleBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the Role GVK for this builder.
func (builder *RoleBuilder) GetGVK() schema.GroupVersionKind {
	return roleGVK
}

// SetError updates the common builder error and its compatibility mirror.
func (builder *RoleBuilder) SetError(err error) {
	builder.EmbeddableBuilder.SetError(err)

	if err == nil {
		builder.errorMsg = ""

		return
	}

	builder.errorMsg = err.Error()
}

// NewRoleBuilder create a new instance of RoleBuilder.
func NewRoleBuilder(apiClient *clients.Settings, name, nsname string, rule rbacv1.PolicyRule) *RoleBuilder {
	klog.V(100).Infof(
		"Initializing new role structure with the following params: "+
			"name: %s, namespace: %s, rule %v", name, nsname, rule)

	builder := common.NewNamespacedBuilder[rbacv1.Role, RoleBuilder](
		apiClient, rbacv1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	builder.WithRules([]rbacv1.PolicyRule{rule})

	return builder
}

// WithRules adds the specified PolicyRule to the Role.
func (builder *RoleBuilder) WithRules(rules []rbacv1.PolicyRule) *RoleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding to role %s the following rules: %v",
		builder.Definition.Name, rules)

	if len(rules) == 0 {
		klog.V(100).Info("The list of rules is empty")

		builder.SetError(fmt.Errorf("cannot create role with empty rule"))

		return builder
	}

	for _, rule := range rules {
		if len(rule.Verbs) == 0 {
			klog.V(100).Info("The role has no verbs")

			builder.SetError(fmt.Errorf(roleMustContainAtLeastOne))

			return builder
		}

		if len(rule.Resources) == 0 {
			klog.V(100).Info("The role has no resources")

			builder.SetError(fmt.Errorf("role must contain at least one Resource"))

			return builder
		}

		if len(rule.APIGroups) == 0 {
			klog.V(100).Info("The role has no apigroups")

			builder.SetError(fmt.Errorf("role must contain at least one APIGroup"))

			return builder
		}
	}

	builder.Definition.Rules = append(builder.Definition.Rules, rules...)

	return builder
}

// PullRole pulls existing role from cluster.
func PullRole(apiClient *clients.Settings, name, nsname string) (*RoleBuilder, error) {
	klog.V(100).Infof("Pulling existing role name %s under namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[rbacv1.Role, RoleBuilder](
		context.TODO(), apiClient, rbacv1.AddToScheme, name, nsname)
}
