package assisted

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testInfraEnvName       = "test-infraenv"
	testInfraEnvNamespace  = "test-infraenv-namespace"
	testInfraEnvNamespace2 = "test-infraenv-namespace-2"
)

var infraEnvTestSchemes = []clients.SchemeAttacher{
	agentInstallV1Beta1.AddToScheme,
}

func TestListInfraEnvsInAllNamespaces(t *testing.T) {
	testCases := []struct {
		name          string
		infraEnvs     []*agentInstallV1Beta1.InfraEnv
		listOptions   []runtimeclient.ListOption
		client        bool
		expectedError error
		expectedNames []string
	}{
		{
			name: "lists all infraenvs",
			infraEnvs: []*agentInstallV1Beta1.InfraEnv{
				buildDummyInfraEnv(testInfraEnvName, testInfraEnvNamespace),
				buildDummyInfraEnv("another-infraenv", testInfraEnvNamespace2),
			},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
			expectedNames: []string{testInfraEnvName, "another-infraenv"},
		},
		{
			name: "lists infraenvs with namespace option",
			infraEnvs: []*agentInstallV1Beta1.InfraEnv{
				buildDummyInfraEnv(testInfraEnvName, testInfraEnvNamespace),
				buildDummyInfraEnv("another-infraenv", testInfraEnvNamespace2),
			},
			listOptions: []runtimeclient.ListOption{
				&runtimeclient.ListOptions{Namespace: testInfraEnvNamespace},
			},
			client:        true,
			expectedError: nil,
			expectedNames: []string{testInfraEnvName},
		},
		{
			name:          "nil client",
			infraEnvs:     nil,
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("the apiClient is nil"),
			expectedNames: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var testSettings *clients.Settings

			if testCase.client {
				testSettings = buildTestClientWithInfraEnvs(testCase.infraEnvs...)
			}

			builders, err := ListInfraEnvsInAllNamespaces(testSettings, testCase.listOptions...)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError != nil {
				assert.Nil(t, builders)

				return
			}

			names := make([]string, 0, len(builders))
			for _, builder := range builders {
				names = append(names, builder.Definition.Name)
			}

			assert.ElementsMatch(t, testCase.expectedNames, names)
		})
	}
}

func TestListInfraEnvs(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		infraEnvs     []*agentInstallV1Beta1.InfraEnv
		client        bool
		expectedError error
		expectedNames []string
	}{
		{
			name:      "lists infraenvs in namespace",
			namespace: testInfraEnvNamespace,
			infraEnvs: []*agentInstallV1Beta1.InfraEnv{
				buildDummyInfraEnv(testInfraEnvName, testInfraEnvNamespace),
				buildDummyInfraEnv("node-infraenv", testInfraEnvNamespace),
				buildDummyInfraEnv("other-ns-infraenv", testInfraEnvNamespace2),
			},
			client:        true,
			expectedError: nil,
			expectedNames: []string{testInfraEnvName, "node-infraenv"},
		},
		{
			name:          "empty namespace",
			namespace:     "",
			infraEnvs:     nil,
			client:        true,
			expectedError: fmt.Errorf("namespace to list infraenvs cannot be empty"),
			expectedNames: nil,
		},
		{
			name:          "nil client",
			namespace:     testInfraEnvNamespace,
			infraEnvs:     nil,
			client:        false,
			expectedError: fmt.Errorf("the apiClient is nil"),
			expectedNames: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var testSettings *clients.Settings

			if testCase.client {
				testSettings = buildTestClientWithInfraEnvs(testCase.infraEnvs...)
			}

			builders, err := ListInfraEnvs(testSettings, testCase.namespace)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError != nil {
				assert.Nil(t, builders)

				return
			}

			names := make([]string, 0, len(builders))
			for _, builder := range builders {
				names = append(names, builder.Definition.Name)
			}

			assert.ElementsMatch(t, testCase.expectedNames, names)
		})
	}
}

func buildDummyInfraEnv(name, namespace string) *agentInstallV1Beta1.InfraEnv {
	return &agentInstallV1Beta1.InfraEnv{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func buildTestClientWithInfraEnvs(infraEnvs ...*agentInstallV1Beta1.InfraEnv) *clients.Settings {
	runtimeObjects := make([]runtime.Object, 0, len(infraEnvs))
	for _, infraEnv := range infraEnvs {
		runtimeObjects = append(runtimeObjects, infraEnv)
	}

	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  runtimeObjects,
		SchemeAttachers: infraEnvTestSchemes,
	})
}
