package sriovfec

import (
	"errors"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	sriovfectypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/fectypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestClusterConfigList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*ClusterConfigBuilder, error) {
			return ListClusterConfig(apiClient, nsname)
		},
		sriovfectypes.AddToScheme,
		clusterConfigGVK,
	).ExecuteTests(t)
}

func TestClusterConfigListWithOptions(t *testing.T) {
	t.Parallel()

	selected := buildDummyClusterConfig("selected", defaultClusterConfigNamespace)
	selected.Labels = map[string]string{"fec": "selected"}

	testCases := []struct {
		name          string
		objects       []runtime.Object
		options       []runtimeclient.ListOptions
		expectedNames []string
		expectedError error
	}{
		{
			name:    "label selector filters cluster configs",
			objects: []runtime.Object{selected, buildDummyClusterConfig("other", defaultClusterConfigNamespace)},
			options: []runtimeclient.ListOptions{{
				LabelSelector: labels.SelectorFromSet(labels.Set{"fec": "selected"}),
			}},
			expectedNames: []string{"selected"},
		},
		{
			name: "namespace option is overridden by supplied namespace",
			objects: []runtime.Object{
				buildDummyClusterConfig("selected", defaultClusterConfigNamespace),
				buildDummyClusterConfig("wrong-namespace", "wrong-ns"),
			},
			options:       []runtimeclient.ListOptions{{Namespace: "wrong-ns"}},
			expectedNames: []string{"selected"},
		},
		{
			name:          "more than one list option returns error",
			objects:       []runtime.Object{selected},
			options:       []runtimeclient.ListOptions{{}, {}},
			expectedError: errors.New("error: more than one ListOptions was passed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  testCase.objects,
				SchemeAttachers: []clients.SchemeAttacher{sriovfectypes.AddToScheme},
			})

			builders, err := ListClusterConfig(testSettings, defaultClusterConfigNamespace, testCase.options...)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError != nil {
				assert.Nil(t, builders)

				return
			}

			require.NoError(t, err)

			names := make([]string, 0, len(builders))
			for _, builder := range builders {
				names = append(names, builder.Definition.Name)
			}

			assert.ElementsMatch(t, testCase.expectedNames, names)
		})
	}
}
