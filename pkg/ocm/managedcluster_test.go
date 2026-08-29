package ocm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ocm/clusterv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultManagedClusterName = "managedcluster-test"

func TestNewManagedClusterBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		NewManagedClusterBuilder, clusterv1.Install, managedClusterGVK).ExecuteTests(t)
}

func TestPullManagedCluster(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullManagedCluster, clusterv1.Install, managedClusterGVK).ExecuteTests(t)
}

func TestManagedClusterBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[clusterv1.ManagedCluster, ManagedClusterBuilder](
		clusterv1.Install,
		managedClusterGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestManagedClusterWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[clusterv1.ManagedCluster, ManagedClusterBuilder](
			clusterv1.Install,
			managedClusterGVK,
			testhelper.ResourceScopeClusterScoped,
		)).ExecuteTests(t)
}

func TestWithHubAcceptsClient(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		accept bool
		valid  bool
	}{
		{
			name:   "accept true",
			accept: true,
			valid:  true,
		},
		{
			name:   "accept false",
			accept: false,
			valid:  true,
		},
		{
			name:   "invalid builder",
			accept: true,
			valid:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var managedClusterBuilder *ManagedClusterBuilder
			if testCase.valid {
				managedClusterBuilder = buildValidManagedClusterTestBuilder(buildTestClientWithManagedClusterScheme())
			} else {
				managedClusterBuilder = buildInvalidManagedClusterTestBuilder(buildTestClientWithManagedClusterScheme())
			}

			managedClusterBuilder = managedClusterBuilder.WithHubAcceptsClient(testCase.accept)

			switch testCase.name {
			case "invalid builder":
				require.Error(t, managedClusterBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(managedClusterBuilder.GetError()))
			default:
				assert.NoError(t, managedClusterBuilder.GetError())
				assert.Equal(t, testCase.accept, managedClusterBuilder.Definition.Spec.HubAcceptsClient)
			}
		})
	}
}

func TestManagedClusterDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		clusterExists bool
	}{
		{
			name:          "deleted cluster",
			valid:         true,
			clusterExists: true,
		},
		{
			name:          "cluster already absent",
			valid:         true,
			clusterExists: false,
		},
		{
			name:          "invalid builder",
			valid:         false,
			clusterExists: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings
			if testCase.clusterExists {
				testSettings = buildTestClientWithDummyManagedCluster()
			} else {
				testSettings = buildTestClientWithManagedClusterScheme()
			}

			var managedClusterBuilder *ManagedClusterBuilder
			if testCase.valid {
				managedClusterBuilder = buildValidManagedClusterTestBuilder(testSettings)
			} else {
				managedClusterBuilder = buildInvalidManagedClusterTestBuilder(testSettings)
			}

			err := managedClusterBuilder.DeleteAndWait(time.Second)

			switch testCase.name {
			case "invalid builder":
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
			default:
				assert.NoError(t, err)
				assert.Nil(t, managedClusterBuilder.Object)
			}
		})
	}
}

func TestManagedClusterWaitForLabel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		valid         bool
		hasLabel      bool
		expectedError error
	}{
		{
			name:     "label found",
			exists:   true,
			valid:    true,
			hasLabel: true,
		},
		{
			name:          "not exists",
			exists:        false,
			valid:         true,
			hasLabel:      true,
			expectedError: fmt.Errorf("managedCluster object %s does not exist", defaultManagedClusterName),
		},
		{
			name:     "invalid builder",
			exists:   true,
			valid:    false,
			hasLabel: true,
		},
		{
			name:          "timeout waiting for label",
			exists:        true,
			valid:         true,
			hasLabel:      false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				mcl := buildDummyManagedCluster(defaultManagedClusterName)

				if testCase.hasLabel {
					mcl.Labels = map[string]string{"test": ""}
				}

				runtimeObjects = append(runtimeObjects, mcl)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{
					clusterv1.Install,
				},
			})

			var testBuilder *ManagedClusterBuilder
			if testCase.valid {
				testBuilder = buildValidManagedClusterTestBuilder(testSettings)
			} else {
				testBuilder = buildInvalidManagedClusterTestBuilder(testSettings)
			}

			_, err := testBuilder.WaitForLabel("test", time.Second)

			switch testCase.name {
			case "invalid builder":
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
			case "timeout waiting for label", "not exists":
				assert.Equal(t, testCase.expectedError, err)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

func TestManagedClusterWaitForCondition(t *testing.T) {
	t.Parallel()

	testCases := []managedClusterWaitForConditionTestCase{
		{
			name:         "condition found",
			exists:       true,
			valid:        true,
			hasCondition: true,
			condition: metav1.Condition{
				Type: clusterv1.ManagedClusterConditionAvailable, Status: metav1.ConditionTrue,
			},
		},
		{
			name:         "not exists",
			exists:       false,
			valid:        true,
			hasCondition: true,
			condition: metav1.Condition{
				Type: clusterv1.ManagedClusterConditionAvailable, Status: metav1.ConditionTrue,
			},
			expectedError: fmt.Errorf("cannot wait for non-existent ManagedCluster"),
		},
		{
			name:         "invalid builder",
			exists:       true,
			valid:        false,
			hasCondition: true,
			condition: metav1.Condition{
				Type: clusterv1.ManagedClusterConditionAvailable, Status: metav1.ConditionTrue,
			},
		},
		{
			name:         "timeout waiting for condition",
			exists:       true,
			valid:        true,
			hasCondition: false,
			condition: metav1.Condition{
				Type: clusterv1.ManagedClusterConditionAvailable, Status: metav1.ConditionTrue,
			},
			expectedError: context.DeadlineExceeded,
		},
		{
			name:         "condition with reason and message",
			exists:       true,
			valid:        true,
			hasCondition: true,
			condition: metav1.Condition{
				Type:    clusterv1.ManagedClusterConditionAvailable,
				Status:  metav1.ConditionTrue,
				Reason:  "TestReason",
				Message: "Test message",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runManagedClusterWaitForConditionTestCase(t, testCase)
		})
	}
}

type managedClusterWaitForConditionTestCase struct {
	name          string
	exists        bool
	valid         bool
	hasCondition  bool
	condition     metav1.Condition
	expectedError error
}

func runManagedClusterWaitForConditionTestCase(t *testing.T, testCase managedClusterWaitForConditionTestCase) {
	t.Helper()

	var runtimeObjects []runtime.Object

	if testCase.exists {
		mcl := buildDummyManagedCluster(defaultManagedClusterName)

		if testCase.hasCondition {
			mcl.Status.Conditions = []metav1.Condition{testCase.condition}
		}

		runtimeObjects = append(runtimeObjects, mcl)
	}

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: runtimeObjects,
		SchemeAttachers: []clients.SchemeAttacher{
			clusterv1.Install,
		},
	})

	var testBuilder *ManagedClusterBuilder
	if testCase.valid {
		testBuilder = buildValidManagedClusterTestBuilder(testSettings)
	} else {
		testBuilder = buildInvalidManagedClusterTestBuilder(testSettings)
	}

	_, err := testBuilder.WaitForCondition(testCase.condition, time.Second)

	switch testCase.name {
	case "invalid builder":
		require.Error(t, err)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
	case "timeout waiting for condition", "not exists":
		assert.Equal(t, testCase.expectedError, err)
	default:
		assert.NoError(t, err)
	}
}

// buildDummyManagedCluster returns a ManagedCluster with the provided name.
func buildDummyManagedCluster(name string) *clusterv1.ManagedCluster {
	return &clusterv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyManagedCluster returns a client with a mock dummy ManagedCluster.
func buildTestClientWithDummyManagedCluster() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyManagedCluster(defaultManagedClusterName),
		},
		SchemeAttachers: []clients.SchemeAttacher{
			clusterv1.Install,
		},
	})
}

// buildTestClientWithManagedClusterScheme returns a client with no objects but the ManagedCluster scheme.
func buildTestClientWithManagedClusterScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{
			clusterv1.Install,
		},
	})
}

// buildValidManagedClusterTestBuilder returns a valid ManagedCluster for testing.
func buildValidManagedClusterTestBuilder(apiClient *clients.Settings) *ManagedClusterBuilder {
	return NewManagedClusterBuilder(apiClient, defaultManagedClusterName)
}

// buildInvalidManagedClusterTestBuilder returns an invalid ManagedCluster for testing.
func buildInvalidManagedClusterTestBuilder(apiClient *clients.Settings) *ManagedClusterBuilder {
	return NewManagedClusterBuilder(apiClient, "")
}
