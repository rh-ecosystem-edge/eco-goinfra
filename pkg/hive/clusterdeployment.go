package hive

import (
	"context"
	"errors"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	hiveextV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/hiveextension/v1beta1"
	hiveV1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/hive/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/hive/api/v1/agent"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	kindAgentClusterInstall       = "AgentClusterInstall"
	errEmptyName                  = "clusterdeployment 'name' cannot be empty"
	errEmptyNamespace             = "clusterdeployment 'namespace' cannot be empty"
	errEmptyClusterName           = "clusterdeployment 'clusterName' cannot be empty"
	errEmptyBaseDomain            = "clusterdeployment 'baseDomain' cannot be empty"
	errEmptyClusterInstallRefName = "clusterdeployment 'clusterInstallRef.name' cannot be empty"
)

var clusterDeploymentGVK = hiveV1.SchemeGroupVersion.WithKind("ClusterDeployment")

// ClusterDeploymentBuilder provides struct for the clusterdeployment object containing connection to
// the cluster and the clusterdeployment definitions.
type ClusterDeploymentBuilder struct {
	common.EmbeddableBuilder[hiveV1.ClusterDeployment, *hiveV1.ClusterDeployment]
	common.EmbeddableCreator[
		hiveV1.ClusterDeployment, ClusterDeploymentBuilder, *hiveV1.ClusterDeployment, *ClusterDeploymentBuilder]
	common.EmbeddableDeleter[hiveV1.ClusterDeployment, *hiveV1.ClusterDeployment]
	common.EmbeddableForceUpdater[
		hiveV1.ClusterDeployment, ClusterDeploymentBuilder, *hiveV1.ClusterDeployment, *ClusterDeploymentBuilder]
	common.EmbeddableWithOptions[
		hiveV1.ClusterDeployment,
		ClusterDeploymentBuilder,
		*hiveV1.ClusterDeployment,
		*ClusterDeploymentBuilder,
		ClusterDeploymentAdditionalOptions,
	]
}

// ClusterDeploymentAdditionalOptions additional options for ClusterDeployment object.
type ClusterDeploymentAdditionalOptions func(builder *ClusterDeploymentBuilder) (*ClusterDeploymentBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ClusterDeploymentBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the ClusterDeployment GVK for this builder.
func (builder *ClusterDeploymentBuilder) GetGVK() schema.GroupVersionKind {
	return clusterDeploymentGVK
}

// NewABMClusterDeploymentBuilder creates a new instance of
// ClusterDeploymentBuilder with platform type set to agentBareMetal.
func NewABMClusterDeploymentBuilder(
	apiClient *clients.Settings,
	name string,
	nsname string,
	clusterName string,
	baseDomain string,
	clusterInstallRef string,
	agentSelector metav1.LabelSelector) *ClusterDeploymentBuilder {
	klog.V(100).Infof(
		`Initializing new agentbaremetal clusterdeployment structure with the following params: name: %s, namespace: %s,
		  clusterName: %s, baseDomain: %s, clusterInstallRef: %s, agentSelector: %s`,
		name, nsname, clusterName, baseDomain, clusterInstallRef, agentSelector)

	return NewClusterDeploymentByInstallRefBuilder(apiClient, name, nsname, clusterName, baseDomain,
		hiveV1.ClusterInstallLocalReference{
			Group:   hiveextV1Beta1.Group,
			Version: hiveextV1Beta1.Version,
			Kind:    kindAgentClusterInstall,
			Name:    clusterInstallRef,
		}, hiveV1.Platform{
			AgentBareMetal: &agent.BareMetalPlatform{
				AgentSelector: agentSelector,
			},
		})
}

// NewClusterDeploymentByInstallRefBuilder creates a new instance of
// ClusterDeploymentBuilder with the provided install reference and platform.
func NewClusterDeploymentByInstallRefBuilder(
	apiClient *clients.Settings,
	name string,
	nsname string,
	clusterName string,
	baseDomain string,
	clusterInstallRef hiveV1.ClusterInstallLocalReference,
	platform hiveV1.Platform) *ClusterDeploymentBuilder {
	klog.V(100).Infof(
		`Initializing new clusterdeployment structure with the following params: name: %s, namespace: %s,
		  clusterName: %s, baseDomain: %s, clusterInstallRef: %v, platform: %v`,
		name, nsname, clusterName, baseDomain, clusterInstallRef, platform)

	builder := common.NewNamespacedBuilder[hiveV1.ClusterDeployment, ClusterDeploymentBuilder](
		apiClient, hiveV1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	if clusterName == "" {
		klog.V(100).Info("The clusterName of the clusterdeployment is empty")

		builder.SetError(errors.New(errEmptyClusterName))

		return builder
	}

	if baseDomain == "" {
		klog.V(100).Info("The baseDomain of the clusterdeployment is empty")

		builder.SetError(errors.New(errEmptyBaseDomain))

		return builder
	}

	if clusterInstallRef.Name == "" {
		klog.V(100).Info("The clusterInstallRef name of the clusterdeployment is empty")

		builder.SetError(errors.New(errEmptyClusterInstallRefName))

		return builder
	}

	builder.Definition.Spec = hiveV1.ClusterDeploymentSpec{
		ClusterName:       clusterName,
		BaseDomain:        baseDomain,
		ClusterInstallRef: &clusterInstallRef,
		Platform:          platform,
	}

	return builder
}

// WithAdditionalAgentSelectorLabels inserts additional labels
// into the clusterdeployment label selector.
func (builder *ClusterDeploymentBuilder) WithAdditionalAgentSelectorLabels(
	agentSelector map[string]string) *ClusterDeploymentBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding agentSelectors %s to clusterdeployment %s in namespace %s",
		agentSelector, builder.Definition.Name, builder.Definition.Namespace)

	if builder.Definition.Spec.Platform.AgentBareMetal == nil {
		klog.V(100).Info("The clusterdeployment platform is not agentBareMetal")

		builder.SetError(errors.New("clusterdeployment type must be AgentBareMetal to use agentSelector"))

		return builder
	}

	if len(agentSelector) == 0 {
		klog.V(100).Info("The clusterdeployment agentSelector is empty")

		builder.SetError(errors.New("agentSelector cannot be empty"))

		return builder
	}

	if len(builder.Definition.Spec.Platform.AgentBareMetal.AgentSelector.MatchLabels) == 0 {
		builder.Definition.Spec.Platform.AgentBareMetal.AgentSelector.MatchLabels = agentSelector
	} else {
		for k, v := range agentSelector {
			builder.Definition.Spec.Platform.AgentBareMetal.AgentSelector.MatchLabels[k] = v
		}
	}

	return builder
}

// WithPullSecret adds a pull-secret reference to the clusterdeployment.
func (builder *ClusterDeploymentBuilder) WithPullSecret(psName string) *ClusterDeploymentBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding pull-secret ref %s to clusterdeployment %s in namespace %s",
		psName, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.PullSecretRef = &corev1.LocalObjectReference{Name: psName}

	return builder
}

// PullClusterDeployment pulls existing clusterdeployment from cluster.
func PullClusterDeployment(apiClient *clients.Settings, name, nsname string) (*ClusterDeploymentBuilder, error) {
	klog.V(100).Infof("Pulling existing clusterdeployment name %s under namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[hiveV1.ClusterDeployment, ClusterDeploymentBuilder](
		context.TODO(), apiClient, hiveV1.AddToScheme, name, nsname)
}
