package assisted

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/models"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nonExistentMsg = "cannot update non-existent agent"
)

var agentGVK = agentInstallV1Beta1.GroupVersion.WithKind("Agent")

// agentBuilder provides struct for the agent object containing connection to
// the cluster and the agent definitions.
type agentBuilder struct {
	common.EmbeddableBuilder[agentInstallV1Beta1.Agent, *agentInstallV1Beta1.Agent]
	common.EmbeddableUpdater[
		agentInstallV1Beta1.Agent, agentBuilder, *agentInstallV1Beta1.Agent, *agentBuilder]
	common.EmbeddableDeleter[agentInstallV1Beta1.Agent, *agentInstallV1Beta1.Agent]
	common.EmbeddableWithOptions[
		agentInstallV1Beta1.Agent, agentBuilder, *agentInstallV1Beta1.Agent, *agentBuilder, AgentAdditionalOptions]
}

// AgentAdditionalOptions additional options for agent object.
type AgentAdditionalOptions func(builder *agentBuilder) (*agentBuilder, error)

// runtimeClientPointer ensures clients held by other assisted builders are wrapped in a pointer for common.Validate.
type runtimeClientPointer struct {
	goclient.Client
}

// AttachMixins wires the embedded CRUD and options mixins to this builder instance.
func (builder *agentBuilder) AttachMixins() {
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the Agent GVK for this builder.
func (builder *agentBuilder) GetGVK() schema.GroupVersionKind {
	return agentGVK
}

// newAgentBuilder creates a new instance of agentBuilder
// Users cannot create agent resources themselves as they are generated from the operator.
func newAgentBuilder(apiClient goclient.Client, definition *agentInstallV1Beta1.Agent) *agentBuilder {
	if definition == nil {
		return nil
	}

	klog.V(100).Infof("Initializing new agent structure for the following agent %s",
		definition.Name)

	var client goclient.Client
	if apiClient != nil {
		client = &runtimeClientPointer{Client: apiClient}
	}

	builder := common.NewNamespacedBuilder[agentInstallV1Beta1.Agent, agentBuilder](
		client, agentInstallV1Beta1.AddToScheme, definition.Name, definition.Namespace)

	builder.Definition = definition
	builder.Object = definition

	return builder
}

// PullAgent pulls existing agent from cluster.
func PullAgent(apiClient *clients.Settings, name, nsname string) (*agentBuilder, error) {
	return common.PullNamespacedBuilder[agentInstallV1Beta1.Agent, agentBuilder](
		context.TODO(), apiClient, agentInstallV1Beta1.AddToScheme, name, nsname)
}

// WithHostName sets the hostname of the agent resource.
func (builder *agentBuilder) WithHostName(hostname string) *agentBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting agent %s in namespace %s hostname to %s",
		builder.Definition.Name, builder.Definition.Namespace, hostname)

	if !builder.Exists() {
		klog.V(100).Infof("agent %s in namespace %s does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.SetError(fmt.Errorf("%s", nonExistentMsg))

		return builder
	}

	builder.Definition.Spec.Hostname = hostname

	return builder
}

// WithRole sets the role of the agent resource.
func (builder *agentBuilder) WithRole(role string) *agentBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting agent %s in namespace %s to role %s",
		builder.Definition.Name, builder.Definition.Namespace, role)

	if !builder.Exists() {
		klog.V(100).Infof("agent %s in namespace %s does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.SetError(fmt.Errorf("%s", nonExistentMsg))

		return builder
	}

	builder.Definition.Spec.Role = models.HostRole(role)

	return builder
}

// WithInstallationDisk sets the installationDiskID of the agent.
func (builder *agentBuilder) WithInstallationDisk(diskID string) *agentBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting agent %s in namespace %s installation disk id to %s",
		builder.Definition.Name, builder.Definition.Namespace, diskID)

	builder.Definition.Spec.InstallationDiskID = diskID

	return builder
}

// WithIgnitionConfigOverride sets the ignitionConfigOverrides of the agent.
func (builder *agentBuilder) WithIgnitionConfigOverride(override string) *agentBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting agent %s in namespace %s ignitionConfigOverride to %s",
		builder.Definition.Name, builder.Definition.Namespace, override)

	builder.Definition.Spec.IgnitionConfigOverrides = override

	return builder
}

// WithApproval sets the approved field of the agent.
func (builder *agentBuilder) WithApproval(approved bool) *agentBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting agent %s in namespace %s approval to %v",
		builder.Definition.Name, builder.Definition.Namespace, approved)

	builder.Definition.Spec.Approved = approved

	return builder
}

// WaitForState waits the specified timeout for the agent to report the specified state.
func (builder *agentBuilder) WaitForState(state string, timeout time.Duration) (*agentBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Waiting for agent %s in namespace %s to report state %s",
		builder.Definition.Name, builder.Definition.Namespace, state)

	// Polls every retryInterval to determine if agent is in desired state.
	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			return builder.Object.Status.DebugInfo.State == state, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}

// WaitForStateInfo waits the specified timeout for the agent to report the specified stateInfo.
func (builder *agentBuilder) WaitForStateInfo(stateInfo string, timeout time.Duration) (*agentBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Waiting for agent %s in namespace %s to report stateInfo %s",
		builder.Definition.Name, builder.Definition.Namespace, stateInfo)

	// Polls every retryInterval to determine if agent is in desired state.
	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), retryInterval, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			return builder.Object.Status.DebugInfo.StateInfo == stateInfo, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}
