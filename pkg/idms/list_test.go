package idms

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListImageDigestMirrorSets(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListImageDigestMirrorSets,
		configv1.AddToScheme,
		idmsGVK,
	).ExecuteTests(t)
}

func TestListImageDigestMirrorSetsOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		options       []runtimeClient.ListOptions
		expectedCount int
		expectedError string
	}{
		{
			name: "single option",
			options: []runtimeClient.ListOptions{
				{LabelSelector: labels.Everything()},
			},
			expectedCount: 1,
		},
		{
			name: "multiple options rejected",
			options: []runtimeClient.ListOptions{
				{LabelSelector: labels.Everything()},
				{Namespace: "test"},
			},
			expectedError: "error: more than one ListOptions was passed",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{generateImagDigestMirrorSet()},
				SchemeAttachers: []clients.SchemeAttacher{configv1.AddToScheme},
			})

			builders, err := ListImageDigestMirrorSets(apiClient, testCase.options...)
			if testCase.expectedError != "" {
				require.EqualError(t, err, testCase.expectedError)
				assert.Empty(t, builders)

				return
			}

			require.NoError(t, err)
			assert.Len(t, builders, testCase.expectedCount)
		})
	}
}
