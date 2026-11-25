package nmstate

import (
	"fmt"
	"testing"

	nmstateV1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPolicy(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(ListPolicy, nmstateV1.AddToScheme, policyGVK).ExecuteTests(t)
}

func TestListPolicyRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	testSettings := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{nmstateV1.AddToScheme},
	})

	builders, err := ListPolicy(testSettings, client.ListOptions{}, client.ListOptions{})

	assert.EqualError(t, err, "error: more than one ListOptions was passed")
	assert.Empty(t, builders)
}

func TestCleanAllNMStatePolicies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		listOptions          []client.ListOptions
		runtimeObject        bool
		client               bool
		expectedError        error
		expectAPIClientError bool
	}{
		{
			name:          "deletes existing resources",
			runtimeObject: true,
			client:        true,
		},
		{
			name:                 "nil client",
			runtimeObject:        true,
			expectAPIClientError: true,
		},
		{
			name:          "single list option",
			listOptions:   []client.ListOptions{{Continue: "test"}},
			runtimeObject: true,
			client:        true,
		},
		{
			name:          "multiple list options",
			listOptions:   []client.ListOptions{{Continue: "test"}, {Namespace: "test"}},
			runtimeObject: true,
			client:        true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			name:   "no resources",
			client: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				testSettings   *clients.Settings
				runtimeObjects []runtime.Object
			)

			if testCase.runtimeObject {
				runtimeObjects = buildDummyPolicyObject()
			}

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  runtimeObjects,
					SchemeAttachers: v1TestSchemes,
				})
			}

			err := CleanAllNMStatePolicies(testSettings, testCase.listOptions...)
			if testCase.expectAPIClientError {
				require.Error(t, err)
				assert.True(t, commonerrors.IsAPIClientNil(err), "expected a nil API client error, got %v", err)

				return
			}

			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError != nil {
				return
			}

			builders, err := ListPolicy(testSettings, testCase.listOptions...)
			assert.NoError(t, err)
			assert.Empty(t, builders)
		})
	}
}
