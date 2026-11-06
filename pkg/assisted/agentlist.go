package assisted

import (
	"context"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const clusterTemplateTemplateIDLabel = "clustertemplates.clcm.openshift.io/templateId"

// ListAgents returns all Agent CRs across all namespaces.
func ListAgents(apiClient *clients.Settings, options ...goclient.ListOption) ([]*agentBuilder, error) {
	return common.List[agentInstallV1Beta1.Agent, agentInstallV1Beta1.AgentList, agentBuilder](
		context.TODO(), apiClient, agentInstallV1Beta1.AddToScheme, options...)
}

// ListORANEligibleAgents returns Agent CRs that are eligible for ORAN collection:
// installed (Installed condition is True) and provisioned (cluster template templateId label is present).
func ListORANEligibleAgents(apiClient *clients.Settings, options ...goclient.ListOption) ([]*agentBuilder, error) {
	agents, err := ListAgents(apiClient, options...)
	if err != nil {
		return nil, err
	}

	eligibleAgents := make([]*agentBuilder, 0, len(agents))

	for _, agent := range agents {
		if isORANEligibleAgent(agent.Object) {
			eligibleAgents = append(eligibleAgents, agent)
		}
	}

	return eligibleAgents, nil
}

func isORANEligibleAgent(agent *agentInstallV1Beta1.Agent) bool {
	if agent == nil {
		return false
	}

	installed := conditionsv1.FindStatusCondition(agent.Status.Conditions, agentInstallV1Beta1.InstalledCondition)
	if installed == nil || installed.Status == corev1.ConditionFalse {
		return false
	}

	_, found := agent.Labels[clusterTemplateTemplateIDLabel]

	return found
}
