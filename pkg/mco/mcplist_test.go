package mco

import (
	"context"
	"fmt"
	"testing"
	"time"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultMCPListLabel = "test-label-value"

func TestListMCP(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListMCP,
		mcv1.Install,
		mcpGVK,
	).ExecuteTests(t)

	t.Run("single list option succeeds", func(t *testing.T) {
		t.Parallel()

		builders, err := ListMCP(buildTestClientWithDummyMCP(), runtimeclient.ListOptions{Continue: "test"})

		require.NoError(t, err)
		assert.Len(t, builders, 1)
	})

	t.Run("multiple list options return error", func(t *testing.T) {
		t.Parallel()

		_, err := ListMCP(
			buildTestClientWithDummyMCP(),
			runtimeclient.ListOptions{Continue: "test"},
			runtimeclient.ListOptions{Continue: "test"},
		)

		assert.EqualError(t, err, "error: more than one ListOptions was passed")
	})
}

func TestListMCPByMachineConfigSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		client        bool
		hasLabel      bool
		expectedError error
	}{
		{
			client:        true,
			hasLabel:      true,
			expectedError: nil,
		},
		{
			client:   false,
			hasLabel: true,
		},
		{
			client:   true,
			hasLabel: false,
			expectedError: fmt.Errorf(
				"cannot find MachineConfigPool that targets machineConfig with label: %s", defaultMCPListLabel),
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("client=%t hasLabel=%t", testCase.client, testCase.hasLabel), func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				mcp := buildDummyMCP()

				if testCase.hasLabel {
					mcp.Spec.MachineConfigSelector = &metav1.LabelSelector{
						MatchLabels: map[string]string{"test": defaultMCPListLabel},
					}
				}

				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  []runtime.Object{mcp},
					SchemeAttachers: testSchemes,
				})
			}

			testBuilder, err := ListMCPByMachineConfigSelector(testSettings, defaultMCPListLabel)

			if !testCase.client {
				assert.True(t, commonerrors.IsAPIClientNil(err))

				return
			}

			if testCase.expectedError == nil {
				require.NoError(t, err)
				require.NotNil(t, testBuilder)
				assert.Equal(t, defaultMCPName, testBuilder.Definition.Name)

				return
			}

			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestListMCPWaitToBeStableFor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		client        bool
		stable        bool
		expectedError error
	}{
		{
			client:        true,
			stable:        true,
			expectedError: nil,
		},
		{
			client: false,
			stable: true,
		},
		{
			client:        true,
			stable:        false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("client=%t stable=%t", testCase.client, testCase.stable), func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				mcp := buildDummyMCP()

				if !testCase.stable {
					mcp.Status.DegradedMachineCount = 1
				}

				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  []runtime.Object{mcp},
					SchemeAttachers: testSchemes,
				})
			}

			err := ListMCPWaitToBeStableFor(testSettings, 500*time.Millisecond, time.Second)

			if !testCase.client {
				assert.True(t, commonerrors.IsAPIClientNil(err))
			} else {
				assert.Equal(t, testCase.expectedError, err)
			}
		})
	}
}
