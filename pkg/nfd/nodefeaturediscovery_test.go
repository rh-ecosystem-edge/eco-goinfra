package nfd

import (
	"fmt"
	"testing"

	nfdv1 "github.com/openshift/cluster-nfd-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const almExampleName = "default-cluster-policy"

var almExample = fmt.Sprintf(`[
 {
    "apiVersion": "nfd.openshift.io/v1",
    "kind": "NodeFeatureDiscovery",
    "metadata": {
      "name": "%s",
      "namespace": "test-namespace"
    },
    "spec": {
      "operand": {
        "image": "registry.k8s.io/nfd/node-feature-discovery"
      }
    }
  }
]`, almExampleName)

func TestNFDNewBuilderFromObjectString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		almString         string
		client            bool
		expectedErrorText string
	}{
		{
			name:      "valid ALM example",
			almString: almExample,
			client:    true,
		},
		{
			name:              "empty ALM example",
			almString:         "",
			client:            true,
			expectedErrorText: "error initializing NodeFeatureDiscovery from alm-examples: almExample is an empty string",
		},
		{
			name:      "invalid ALM example",
			almString: "{invalid}",
			client:    true,
			expectedErrorText: "error initializing NodeFeatureDiscovery from alm-examples:" +
				" invalid character 'i' looking for beginning of object key string",
		},
		{
			name:              "empty ALM example list",
			almString:         "[]",
			client:            true,
			expectedErrorText: "error initializing NodeFeatureDiscovery from alm-examples: failed to get alm examples",
		},
		{
			name:      "no client",
			almString: almExample,
			client:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var client *clients.Settings
			if testCase.client {
				client = buildTestClientWithNFDScheme()
			}

			builder := NewBuilderFromObjectString(client, testCase.almString)

			if !testCase.client {
				assert.Nil(t, builder)

				return
			}

			require.NotNil(t, builder)
			assert.Equal(t, nodeFeatureDiscoveryGVK, builder.GetGVK())

			if testCase.expectedErrorText != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedErrorText)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition)
			assert.Equal(t, almExampleName, builder.Definition.Name)
			assert.Equal(t, "test-namespace", builder.Definition.Namespace)
		})
	}
}

func TestNFDPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull, nfdv1.AddToScheme, nodeFeatureDiscoveryGVK).ExecuteTests(t)
}

func TestNFDBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[nfdv1.NodeFeatureDiscovery, Builder](
		nfdv1.AddToScheme,
		nodeFeatureDiscoveryGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func buildTestClientWithNFDScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{nfdv1.AddToScheme},
	})
}
