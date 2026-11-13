package nvidiagpu

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	nvidiagputypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/nvidiagpu/nvidiagputypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultClusterPolicyName = "default-cluster-policy"
	almExample               = `[
 {
    "apiVersion": "nvidia.com/v1",
    "kind": "ClusterPolicy",
    "metadata": {
      "name": "default-cluster-policy"
    },
    "spec": {
      "driver": {
        "enabled": true
      }
    }
  }
]
`
)

func TestNvidiaGPUNewBuilderFromObjectString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		almExample        string
		client            bool
		expectedErrorText string
	}{
		{
			name:       "valid ALM example",
			almExample: almExample,
			client:     true,
		},
		{
			name:              "empty ALM example",
			almExample:        "",
			client:            true,
			expectedErrorText: "error initializing ClusterPolicy from alm-examples: almExample is an empty string",
		},
		{
			name:              "invalid ALM example",
			almExample:        "{ invalid: data }",
			client:            true,
			expectedErrorText: "error initializing ClusterPolicy from alm-examples:",
		},
		{
			name:              "empty ALM example list",
			almExample:        "[]",
			client:            true,
			expectedErrorText: "error initializing ClusterPolicy from alm-examples: failed to get alm examples",
		},
		{
			name:              "no client",
			almExample:        almExample,
			client:            false,
			expectedErrorText: "apiClient for ClusterPolicy is nil",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var client *clients.Settings
			if testCase.client {
				client = buildTestClientWithNvidiaGPUScheme()
			}

			builder := NewBuilderFromObjectString(client, testCase.almExample)

			require.NotNil(t, builder)
			assert.Equal(t, clusterPolicyGVK, builder.GetGVK())

			if testCase.expectedErrorText != "" {
				assert.ErrorContains(t, builder.GetError(), testCase.expectedErrorText)

				_, err := builder.Get()
				assert.ErrorContains(t, err, testCase.expectedErrorText)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition)
			assert.Equal(t, defaultClusterPolicyName, builder.Definition.Name)
		})
	}
}

func TestNvidiaGPUPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		Pull, nvidiagputypes.AddToScheme, clusterPolicyGVK).ExecuteTests(t)
}

func TestNvidiaGPUBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[nvidiagputypes.ClusterPolicy, Builder](
		nvidiagputypes.AddToScheme,
		clusterPolicyGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func buildTestClientWithNvidiaGPUScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{nvidiagputypes.AddToScheme},
	})
}
