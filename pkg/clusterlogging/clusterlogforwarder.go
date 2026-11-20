package clusterlogging

import (
	"context"
	"fmt"

	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var clusterLogForwarderGVK = observabilityv1.GroupVersion.WithKind("ClusterLogForwarder")

// ClusterLogForwarderBuilder provides a struct for clusterlogforwarder object from the
// cluster and a clusterlogforwarder definition.
type ClusterLogForwarderBuilder struct {
	common.EmbeddableBuilder[observabilityv1.ClusterLogForwarder, *observabilityv1.ClusterLogForwarder]
	common.EmbeddableCreator[observabilityv1.ClusterLogForwarder, ClusterLogForwarderBuilder,
		*observabilityv1.ClusterLogForwarder, *ClusterLogForwarderBuilder]
	common.EmbeddableDeleter[observabilityv1.ClusterLogForwarder, *observabilityv1.ClusterLogForwarder]
	common.EmbeddableForceUpdater[observabilityv1.ClusterLogForwarder, ClusterLogForwarderBuilder,
		*observabilityv1.ClusterLogForwarder, *ClusterLogForwarderBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ClusterLogForwarderBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the ClusterLogForwarder GVK for this builder.
func (builder *ClusterLogForwarderBuilder) GetGVK() schema.GroupVersionKind {
	return clusterLogForwarderGVK
}

// NewClusterLogForwarderBuilder method creates new instance of builder.
func NewClusterLogForwarderBuilder(
	apiClient *clients.Settings, name, nsname string) *ClusterLogForwarderBuilder {
	klog.V(100).Infof("Initializing new clusterlogforwarder structure with the following params: "+
		"name: %s, namespace: %s", name, nsname)

	return common.NewNamespacedBuilder[observabilityv1.ClusterLogForwarder, ClusterLogForwarderBuilder](
		apiClient, observabilityv1.AddToScheme, name, nsname)
}

// WithManagementState sets the clusterlogforwarder operator's managementState configuration.
func (builder *ClusterLogForwarderBuilder) WithManagementState(
	managementState observabilityv1.ManagementState) *ClusterLogForwarderBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if managementState != observabilityv1.ManagementStateManaged &&
		managementState != observabilityv1.ManagementStateUnmanaged {
		klog.V(100).Infof("The management state of the clusterlogforwarder is unsupported: %s;"+
			"accepted only %s or %s",
			managementState, observabilityv1.ManagementStateManaged, observabilityv1.ManagementStateUnmanaged)

		builder.SetError(fmt.Errorf("the management state of the clusterlogforwarder is unsupported: \"%s\";"+
			"accepted only %s or %s states",
			managementState, observabilityv1.ManagementStateManaged, observabilityv1.ManagementStateUnmanaged))

		return builder
	}

	klog.V(100).Infof(
		"Setting clusterlogforwarder %s in namespace %s with the managementState config: %v",
		builder.Definition.Name, builder.Definition.Namespace, managementState)

	builder.Definition.Spec.ManagementState = managementState

	return builder
}

// WithServiceAccount sets the clusterlogforwarder operator's serviceAccount configuration.
func (builder *ClusterLogForwarderBuilder) WithServiceAccount(serviceAccount string) *ClusterLogForwarderBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if serviceAccount == "" {
		klog.V(100).Info("The serviceAccount of the clusterlogforwarder is empty")

		builder.SetError(fmt.Errorf("clusterlogforwarder 'serviceAccount' cannot be empty"))

		return builder
	}

	klog.V(100).Infof(
		"Setting clusterlogforwarder %s in namespace %s with the serviceAccount config: %v",
		builder.Definition.Name, builder.Definition.Namespace, serviceAccount)

	builder.Definition.Spec.ServiceAccount.Name = serviceAccount

	return builder
}

// WithOutput sets the output on the clusterlogforwarder definition.
func (builder *ClusterLogForwarderBuilder) WithOutput(
	outputSpec *observabilityv1.OutputSpec) *ClusterLogForwarderBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting output %v on clusterlogforwarder %s in namespace %s",
		outputSpec, builder.Definition.Name, builder.Definition.Namespace)

	if outputSpec == nil {
		klog.V(100).Info("The 'outputSpec' of the deployment is empty")

		builder.SetError(fmt.Errorf("'outputSpec' parameter is empty"))

		return builder
	}

	builder.Definition.Spec.Outputs = append(builder.Definition.Spec.Outputs, *outputSpec)

	return builder
}

// WithPipeline sets the pipeline on the clusterlogforwarder definition.
func (builder *ClusterLogForwarderBuilder) WithPipeline(
	pipelineSpec *observabilityv1.PipelineSpec) *ClusterLogForwarderBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting pipeline %v on clusterlogforwarder %s in namespace %s",
		pipelineSpec, builder.Definition.Name, builder.Definition.Namespace)

	if pipelineSpec == nil {
		klog.V(100).Info("The 'pipelineSpec' of the deployment is empty")

		builder.SetError(fmt.Errorf("'pipelineSpec' parameter is empty"))

		return builder
	}

	builder.Definition.Spec.Pipelines = append(builder.Definition.Spec.Pipelines, *pipelineSpec)

	return builder
}

// PullClusterLogForwarder retrieves an existing clusterlogforwarder object from the cluster.
func PullClusterLogForwarder(apiClient *clients.Settings, name, nsname string) (*ClusterLogForwarderBuilder, error) {
	klog.V(100).Infof("Pulling existing clusterlogforwarder %s in nsname %s", name, nsname)

	return common.PullNamespacedBuilder[observabilityv1.ClusterLogForwarder, ClusterLogForwarderBuilder](
		context.TODO(), apiClient, observabilityv1.AddToScheme, name, nsname)
}
