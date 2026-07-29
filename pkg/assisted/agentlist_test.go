package assisted

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	testCases := []struct {
		agents        []*agentBuilder
		listOptions   []runtimeclient.ListOption
		client        bool
		expectedError error
	}{
		{
			agents: []*agentBuilder{
				buildValidAgentTestBuilder(buildTestClientWithDummyAgent()),
			},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
		},
		{
			agents: []*agentBuilder{
				buildValidAgentTestBuilder(buildTestClientWithDummyAgent()),
			},
			listOptions:   []runtimeclient.ListOption{&runtimeclient.ListOptions{LabelSelector: labels.NewSelector()}},
			client:        true,
			expectedError: nil,
		},
		{
			agents: []*agentBuilder{
				buildValidAgentTestBuilder(buildTestClientWithDummyAgent()),
			},
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("the apiClient is nil"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyAgent()
		}

		builders, err := ListAgents(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.agents), len(builders))
		}
	}
}

func TestListORANEligibleAgents(t *testing.T) {
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
