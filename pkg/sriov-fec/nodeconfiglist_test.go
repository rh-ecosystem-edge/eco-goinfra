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

func TestNodeConfigList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*NodeConfigBuilder, error) {
			return List(apiClient, nsname)
		},
		sriovfectypes.AddToScheme,
		nodeConfigGVK,
	).ExecuteTests(t)
}

func TestNodeConfigListWithOptions(t *testing.T) {
	t.Parallel()

	selected := buildDummyNodeConfig("selected", defaultNodeConfigNamespace)
	selected.Labels = map[string]string{"fec": "selected"}

	testCases := []struct {
		name          string
		objects       []runtime.Object
		options       []runtimeclient.ListOptions
		expectedNames []string
		expectedError error
	}{
		{
			name:    "label selector filters node configs",
			objects: []runtime.Object{selected, buildDummyNodeConfig("other", defaultNodeConfigNamespace)},
			options: []runtimeclient.ListOptions{{
				LabelSelector: labels.SelectorFromSet(labels.Set{"fec": "selected"}),
			}},
			expectedNames: []string{"selected"},
		},
		{
			name: "supplied namespace does not implicitly scope the list",
			objects: []runtime.Object{
				buildDummyNodeConfig("default-namespace", defaultNodeConfigNamespace),
				buildDummyNodeConfig("other-namespace", "other-ns"),
			},
			expectedNames: []string{"default-namespace", "other-namespace"},
		},
		{
			name: "namespace option scopes the list",
			objects: []runtime.Object{
				buildDummyNodeConfig("selected", defaultNodeConfigNamespace),
				buildDummyNodeConfig("wrong-namespace", "wrong-ns"),
			},
			options:       []runtimeclient.ListOptions{{Namespace: defaultNodeConfigNamespace}},
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

			builders, err := List(testSettings, defaultNodeConfigNamespace, testCase.options...)
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
