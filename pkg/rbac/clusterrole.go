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

const clusterroleRuleMustContainAtLeast = "clusterrole rule must contain at least one Verb entry"

var clusterRoleGVK = rbacv1.SchemeGroupVersion.WithKind("ClusterRole")

/*
ClusterRoleBuilder provides struct for clusterrole object

	containing connection to the cluster and the clusterrole definitions.
*/
type ClusterRoleBuilder struct {
	common.EmbeddableBuilder[rbacv1.ClusterRole, *rbacv1.ClusterRole]
	common.EmbeddableCreator[rbacv1.ClusterRole, ClusterRoleBuilder, *rbacv1.ClusterRole, *ClusterRoleBuilder]
	common.EmbeddableDeleter[rbacv1.ClusterRole, *rbacv1.ClusterRole]
	common.EmbeddableUpdater[rbacv1.ClusterRole, ClusterRoleBuilder, *rbacv1.ClusterRole, *ClusterRoleBuilder]
	common.EmbeddableWithOptions[rbacv1.ClusterRole, ClusterRoleBuilder, *rbacv1.ClusterRole, *ClusterRoleBuilder,
		ClusterRoleAdditionalOptions]
}

// ClusterRoleAdditionalOptions additional options for ClusterRole object.
type ClusterRoleAdditionalOptions func(builder *ClusterRoleBuilder) (*ClusterRoleBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ClusterRoleBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the ClusterRole GVK for this builder.
func (builder *ClusterRoleBuilder) GetGVK() schema.GroupVersionKind {
	return clusterRoleGVK
}

// NewClusterRoleBuilder creates new instance of ClusterRoleBuilder.
func NewClusterRoleBuilder(apiClient *clients.Settings, name string, rule rbacv1.PolicyRule) *ClusterRoleBuilder {
	klog.V(100).Infof(
		"Initializing new clusterrole structure with the following params: "+
			"name: %s, policy rule: %v",
		name, rule)

	builder := common.NewClusterScopedBuilder[rbacv1.ClusterRole, ClusterRoleBuilder](
		apiClient, rbacv1.AddToScheme, name)
	if builder.GetError() != nil {
		return builder
	}

	builder.WithRules([]rbacv1.PolicyRule{rule})

	return builder
}

// WithRules appends additional rules to the clusterrole definition.
func (builder *ClusterRoleBuilder) WithRules(rules []rbacv1.PolicyRule) *ClusterRoleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Appending to the definition of clusterrole %s these additional rules %v",
		builder.Definition.Name, rules)

	if len(rules) == 0 {
		klog.V(100).Info("The list of rules is empty")

		builder.SetError(fmt.Errorf("cannot accept nil or empty slice as rules"))

		return builder
	}

	for _, rule := range rules {
		if len(rule.Verbs) == 0 {
			klog.V(100).Info("The clusterrole rule must contain at least one Verb entry")

			builder.SetError(fmt.Errorf(clusterroleRuleMustContainAtLeast))

			return builder
		}
	}

	builder.Definition.Rules = append(builder.Definition.Rules, rules...)

	return builder
}

// PullClusterRole pulls existing clusterrole from cluster.
func PullClusterRole(apiClient *clients.Settings, name string) (*ClusterRoleBuilder, error) {
	klog.V(100).Infof("Pulling existing clusterrole name %s from cluster", name)

	return common.PullClusterScopedBuilder[rbacv1.ClusterRole, ClusterRoleBuilder](
		context.TODO(), apiClient, rbacv1.AddToScheme, name)
}
