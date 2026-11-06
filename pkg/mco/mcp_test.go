package mco

import (
	"context"
	"testing"
	"time"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultMCPName = "test-machine-config-pool"

var updatingMCPCondition = mcv1.MachineConfigPoolCondition{
	Type:   mcv1.MachineConfigPoolUpdating,
	Status: corev1.ConditionTrue,
}

func TestNewMCPBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		NewMCPBuilder, mcv1.Install, mcpGVK).ExecuteTests(t)
}

func TestPullMachineConfigPool(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		Pull, mcv1.Install, mcpGVK).ExecuteTests(t)
}

func TestMachineConfigPoolBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[mcv1.MachineConfigPool, MCPBuilder](
		mcv1.Install,
		mcpGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestMachineConfigPoolWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[mcv1.MachineConfigPool, MCPBuilder](
			mcv1.Install,
			mcpGVK,
			testhelper.ResourceScopeClusterScoped,
		)).ExecuteTests(t)
}

func TestMachineConfigPoolWithMCSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		selector            map[string]string
		invalidBuilder      bool
		expectedError       string
		expectedMatchLabels map[string]string
	}{
		{
			name:                "sets selector on valid builder",
			selector:            map[string]string{"test": "test"},
			expectedMatchLabels: map[string]string{"test": "test"},
		},
		{
			name:          "nil selector sets builder error",
			selector:      nil,
			expectedError: "machineConfigSelector 'MatchLabels' field cannot be empty",
		},
		{
			name:          "empty selector sets builder error",
			selector:      map[string]string{},
			expectedError: "machineConfigSelector 'MatchLabels' field cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			selector:       map[string]string{"test": "test"},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := buildTestClientWithDummyMCP()

			var testBuilder *MCPBuilder
			if testCase.invalidBuilder {
				testBuilder = buildInvalidMCPTestBuilder(testSettings)
			} else {
				testBuilder = buildValidMCPTestBuilder(testSettings)
			}

			result := testBuilder.WithMcSelector(testCase.selector)
			require.Same(t, testBuilder, result)

			switch {
			case testCase.invalidBuilder:
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.MachineConfigSelector)
			case testCase.expectedError != "":
				assert.EqualError(t, result.GetError(), testCase.expectedError)
			default:
				require.NoError(t, result.GetError())
				assert.Equal(t, testCase.expectedMatchLabels, result.Definition.Spec.MachineConfigSelector.MatchLabels)
			}
		})
	}
}

func TestMachineConfigPoolWaitToBeInCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		valid         bool
		hasCondition  bool
		expectedError error
	}{
		{
			name:         "condition met",
			exists:       true,
			valid:        true,
			hasCondition: true,
		},
		{
			name:          "missing pool times out",
			exists:        false,
			valid:         true,
			hasCondition:  true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:         "invalid builder returns validation error",
			exists:       true,
			valid:        false,
			hasCondition: true,
		},
		{
			name:          "condition absent times out",
			exists:        true,
			valid:         true,
			hasCondition:  false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildMCPBuilderWithUpdatingCondition(
				testCase.exists, testCase.hasCondition, testCase.valid)
			err := testBuilder.WaitToBeInCondition(
				updatingMCPCondition.Type, updatingMCPCondition.Status, time.Second)

			switch {
			case !testCase.valid:
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
			case testCase.expectedError != nil:
				assert.ErrorIs(t, err, testCase.expectedError)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestMachineConfigPoolWaitForUpdate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		exists        bool
		updating      bool
		expectedError error
	}{
		{
			name:     "pool is not updating",
			valid:    true,
			exists:   true,
			updating: false,
		},
		{
			name:          "pool never finishes updating",
			valid:         true,
			exists:        true,
			updating:      true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:     "invalid builder returns validation error",
			valid:    false,
			exists:   true,
			updating: true,
		},
		{
			name:     "missing pool returns get error",
			valid:    true,
			exists:   false,
			updating: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildMCPBuilderWithUpdatingCondition(testCase.exists, testCase.updating, testCase.valid)
			err := testBuilder.WaitForUpdate(time.Second)

			switch {
			case !testCase.valid:
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
			case !testCase.exists:
				assert.True(t, commonerrors.IsAPICallFailedWithVerb(err, "get"))
			case testCase.expectedError != nil:
				assert.ErrorIs(t, err, testCase.expectedError)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestMachineConfigPoolWaitToBeStableFor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		exists        bool
		stable        bool
		expectedError error
	}{
		{
			name:   "pool is stable",
			valid:  true,
			exists: true,
			stable: true,
		},
		{
			name:   "invalid builder returns validation error",
			valid:  false,
			exists: true,
			stable: true,
		},
		{
			name:   "missing pool is treated as stable",
			valid:  true,
			exists: false,
			stable: true,
		},
		{
			name:          "unstable pool times out",
			valid:         true,
			exists:        true,
			stable:        false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				mcp := buildDummyMCP()

				if !testCase.stable {
					mcp.Status.DegradedMachineCount = 1
				}

				runtimeObjects = append(runtimeObjects, mcp)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{mcv1.Install},
			})

			var testBuilder *MCPBuilder
			if testCase.valid {
				testBuilder = buildValidMCPTestBuilder(testSettings)
			} else {
				testBuilder = buildInvalidMCPTestBuilder(testSettings)
			}

			err := testBuilder.WaitToBeStableFor(500*time.Millisecond, time.Second)

			switch {
			case !testCase.valid:
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
			case testCase.expectedError != nil:
				assert.ErrorIs(t, err, testCase.expectedError)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestMachineConfigPoolIsInCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		valid         bool
		hasCondition  bool
		isInCondition bool
	}{
		{
			name:          "condition present",
			exists:        true,
			valid:         true,
			hasCondition:  true,
			isInCondition: true,
		},
		{
			name:          "invalid builder",
			exists:        true,
			valid:         false,
			hasCondition:  true,
			isInCondition: false,
		},
		{
			name:          "pool missing",
			exists:        false,
			valid:         true,
			hasCondition:  true,
			isInCondition: false,
		},
		{
			name:          "condition absent",
			exists:        true,
			valid:         true,
			hasCondition:  false,
			isInCondition: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildMCPBuilderWithUpdatingCondition(
				testCase.exists, testCase.hasCondition, testCase.valid)

			assert.Equal(t, testCase.isInCondition, testBuilder.IsInCondition(updatingMCPCondition.Type))
		})
	}
}

// buildDummyMCP returns a dummy MachineConfigPool.
func buildDummyMCP() *mcv1.MachineConfigPool {
	return &mcv1.MachineConfigPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultMCPName,
		},
	}
}

// buildTestClientWithDummyMCP returns a client with a dummy MachineConfigPool.
func buildTestClientWithDummyMCP() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyMCP(),
		},
		SchemeAttachers: []clients.SchemeAttacher{mcv1.Install},
	})
}

// buildValidMCPTestBuilder returns a valid MCPBuilder for testing.
func buildValidMCPTestBuilder(apiClient *clients.Settings) *MCPBuilder {
	return NewMCPBuilder(apiClient, defaultMCPName)
}

// buildInvalidMCPTestBuilder returns an invalid MCPBuilder for testing.
func buildInvalidMCPTestBuilder(apiClient *clients.Settings) *MCPBuilder {
	return NewMCPBuilder(apiClient, "")
}

// buildMCPBuilderWithUpdatingCondition returns an MCPBuilder for testing, with the ability to configure whether it
// exists on the test client, has the updating condition, and is valid.
func buildMCPBuilderWithUpdatingCondition(exists, hasCondition, valid bool) *MCPBuilder {
	var runtimeObjects []runtime.Object

	if exists {
		mcp := buildDummyMCP()

		if hasCondition {
			mcp.Status.Conditions = []mcv1.MachineConfigPoolCondition{updatingMCPCondition}
		}

		runtimeObjects = append(runtimeObjects, mcp)
	}

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  runtimeObjects,
		SchemeAttachers: []clients.SchemeAttacher{mcv1.Install},
	})

	if valid {
		return buildValidMCPTestBuilder(testSettings)
	}

	return buildInvalidMCPTestBuilder(testSettings)
}
