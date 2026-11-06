package assisted

import (
	"testing"

	"github.com/google/uuid"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testAgentName      = "test-agent"
	testAgentNamespace = "test-agent-namespace"
)

var agentTestSchemes = []clients.SchemeAttacher{
	agentInstallV1Beta1.AddToScheme,
}

func TestListAgents(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListAgents, agentInstallV1Beta1.AddToScheme, agentGVK).ExecuteTests(t)
}

func TestListAgentsWithOptions(t *testing.T) {
	t.Parallel()

	selectedAgent := buildDummyAgent(testAgentName)
	selectedAgent.Labels = map[string]string{"selected": "true"}
	otherAgent := buildDummyAgent("other-agent")

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{selectedAgent, otherAgent},
		SchemeAttachers: agentTestSchemes,
	})

	builders, err := ListAgents(testSettings, runtimeclient.MatchingLabels{"selected": "true"})

	assert.NoError(t, err)
	assert.Len(t, builders, 1)

	if len(builders) == 1 {
		assert.Equal(t, testAgentName, builders[0].Definition.Name)
	}
}

func TestListORANEligibleAgents(t *testing.T) {
	t.Parallel()

	templateID := uuid.New().String()

	testCases := []struct {
		name          string
		agents        []*agentInstallV1Beta1.Agent
		expectedNames []string
	}{
		{
			name: "returns only ORAN-eligible agents",
			agents: []*agentInstallV1Beta1.Agent{
				buildORANEligibleAgent(testAgentName, templateID),
				buildDummyAgent("not-yet-installed"),
			},
			expectedNames: []string{testAgentName},
		},
		{
			name: "filters agent without template id label",
			agents: []*agentInstallV1Beta1.Agent{
				buildORANEligibleAgent(testAgentName, templateID),
				func() *agentInstallV1Beta1.Agent {
					agent := buildORANEligibleAgent("missing-template-label", templateID)
					delete(agent.Labels, clusterTemplateTemplateIDLabel)

					return agent
				}(),
			},
			expectedNames: []string{testAgentName},
		},
		{
			name: "filters agent with Installed condition false",
			agents: []*agentInstallV1Beta1.Agent{
				buildORANEligibleAgent(testAgentName, templateID),
				func() *agentInstallV1Beta1.Agent {
					agent := buildORANEligibleAgent("not-installed", templateID)
					agent.Status.Conditions = []conditionsv1.Condition{{
						Type:   agentInstallV1Beta1.InstalledCondition,
						Status: corev1.ConditionFalse,
					}}

					return agent
				}(),
			},
			expectedNames: []string{testAgentName},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runtimeObjects := make([]runtime.Object, 0, len(testCase.agents))
			for _, agent := range testCase.agents {
				runtimeObjects = append(runtimeObjects, agent)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: agentTestSchemes,
			})

			builders, err := ListORANEligibleAgents(testSettings)
			assert.NoError(t, err)
			assert.Len(t, builders, len(testCase.expectedNames))

			names := make([]string, 0, len(builders))
			for _, builder := range builders {
				names = append(names, builder.Definition.Name)
			}

			assert.ElementsMatch(t, testCase.expectedNames, names)
		})
	}
}

func TestIsORANEligibleAgent(t *testing.T) {
	t.Parallel()

	templateID := uuid.New().String()

	testCases := []struct {
		name  string
		agent *agentInstallV1Beta1.Agent
	}{
		{
			name:  "eligible installed agent",
			agent: buildORANEligibleAgent(testAgentName, templateID),
		},
		{
			name: "installed condition unknown",
			agent: func() *agentInstallV1Beta1.Agent {
				agent := buildORANEligibleAgent(testAgentName, templateID)
				agent.Status.Conditions = []conditionsv1.Condition{{
					Type:   agentInstallV1Beta1.InstalledCondition,
					Status: corev1.ConditionUnknown,
				}}

				return agent
			}(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, isORANEligibleAgent(testCase.agent))
		})
	}
}

// buildDummyAgent returns an Agent with the provided name.
func buildDummyAgent(name string) *agentInstallV1Beta1.Agent {
	return &agentInstallV1Beta1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testAgentNamespace,
		},
	}
}

// buildTestClientWithDummyAgent returns a client with a mock dummy Agent.
func buildTestClientWithDummyAgent() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyAgent(testAgentName),
		},
		SchemeAttachers: agentTestSchemes,
	})
}

// buildValidAgentTestBuilder returns a valid Agent builder for testing.
func buildValidAgentTestBuilder(apiClient *clients.Settings) *agentBuilder {
	return newAgentBuilder(apiClient.Client, buildDummyAgent(testAgentName))
}

// buildORANEligibleAgent returns an Agent that passes all ORAN eligibility checks.
func buildORANEligibleAgent(name, templateID string) *agentInstallV1Beta1.Agent {
	agent := buildDummyAgent(name)
	agent.Labels = map[string]string{
		clusterTemplateTemplateIDLabel: templateID,
	}
	agent.Status.Conditions = []conditionsv1.Condition{{
		Type:   agentInstallV1Beta1.InstalledCondition,
		Status: corev1.ConditionTrue,
	}}

	return agent
}
