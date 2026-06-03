package cgu

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/cluster-group-upgrades-operator/pkg/api/clustergroupupgrades/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultCguName           = "cgu-test"
	defaultCguNsName         = "test-ns"
	defaultCguMaxConcurrency = 1
	defaultCguClusterName    = "test-cluster"
	defaultCguClusterState   = v1alpha1.NotStarted
	defaultCguCondition      = conditionComplete
)

var cguGVK = v1alpha1.SchemeGroupVersion.WithKind("ClusterGroupUpgrade")

func TestPullCgu(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(Pull, v1alpha1.AddToScheme, cguGVK).ExecuteTests(t)
}

func TestNewCguBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *CguBuilder {
				return NewCguBuilder(apiClient, name, nsname, defaultCguMaxConcurrency)
			},
			v1alpha1.AddToScheme,
			cguGVK,
		).ExecuteTests(t)
	})

	t.Run("maxConcurrency less than 1 returns error", func(t *testing.T) {
		t.Parallel()

		testBuilder := NewCguBuilder(
			clients.GetTestClients(clients.TestClientParams{}),
			defaultCguName,
			defaultCguNsName,
			0,
		)

		assert.Equal(t, errInvalidCguMaxConcurrency, testBuilder.GetError())
	})

	t.Run("valid maxConcurrency sets remediation strategy", func(t *testing.T) {
		t.Parallel()

		testBuilder := NewCguBuilder(
			clients.GetTestClients(clients.TestClientParams{}),
			defaultCguName,
			defaultCguNsName,
			defaultCguMaxConcurrency,
		)

		require.NoError(t, testBuilder.GetError())
		require.NotNil(t, testBuilder.Definition.Spec.RemediationStrategy)
		assert.Equal(t, defaultCguMaxConcurrency, testBuilder.Definition.Spec.RemediationStrategy.MaxConcurrency)
	})
}

func TestCguBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[v1alpha1.ClusterGroupUpgrade, CguBuilder](
		v1alpha1.AddToScheme,
		cguGVK,
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

func TestCguWithCluster(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		cluster       string
		expectedError error
	}{
		{
			name:          "non-empty cluster appends to spec",
			cluster:       "test-cluster",
			expectedError: nil,
		},
		{
			name:          "empty cluster returns error",
			cluster:       "",
			expectedError: fmt.Errorf("cluster in CGU cluster spec cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidCguTestBuilder(buildTestClientWithDummyCguObject())
			testBuilder.WithCluster(testCase.cluster)

			assert.Equal(t, testCase.expectedError, testBuilder.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, []string{testCase.cluster}, testBuilder.Definition.Spec.Clusters)
			}
		})
	}

	t.Run("invalid builder short-circuits", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject())
		testBuilder.WithCluster("test-cluster")

		assert.Equal(t, errInvalidCguMaxConcurrency, testBuilder.GetError())
	})
}

func TestCguWithManagedPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		policy        string
		expectedError error
	}{
		{
			name:          "non-empty policy appends to spec",
			policy:        "test-policy",
			expectedError: nil,
		},
		{
			name:          "empty policy returns error",
			policy:        "",
			expectedError: fmt.Errorf("policy in CGU managedpolicies spec cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidCguTestBuilder(buildTestClientWithDummyCguObject())
			testBuilder.WithManagedPolicy(testCase.policy)

			assert.Equal(t, testCase.expectedError, testBuilder.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, []string{testCase.policy}, testBuilder.Definition.Spec.ManagedPolicies)
			}
		})
	}

	t.Run("invalid builder short-circuits", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject())
		testBuilder.WithManagedPolicy("test-policy")

		assert.Equal(t, errInvalidCguMaxConcurrency, testBuilder.GetError())
	})
}

func TestCguWithCanary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		canary        string
		expectedError error
	}{
		{
			name:          "non-empty canary appends to spec",
			canary:        "test-canary",
			expectedError: nil,
		},
		{
			name:          "empty canary returns error",
			canary:        "",
			expectedError: fmt.Errorf("canary in CGU remediationstrategy spec cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidCguTestBuilder(buildTestClientWithDummyCguObject())
			testBuilder.WithCanary(testCase.canary)

			assert.Equal(t, testCase.expectedError, testBuilder.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, []string{testCase.canary}, testBuilder.Definition.Spec.RemediationStrategy.Canaries)
			}
		})
	}

	t.Run("invalid builder short-circuits", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject())
		testBuilder.WithCanary("test-canary")

		assert.Equal(t, errInvalidCguMaxConcurrency, testBuilder.GetError())
	})
}

func TestCguDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		testCgu       *CguBuilder
		expectedError error
	}{
		{
			name:          "deletes existing cgu and waits",
			testCgu:       buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: nil,
		},
		{
			name:          "deletes created cgu and waits",
			testCgu:       buildValidCguTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			name:          "invalid builder returns validation error",
			testCgu:       buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: errInvalidCguMaxConcurrency,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := testCase.testCgu.DeleteAndWait(time.Second)

			if testCase.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}

			if testCase.expectedError == nil {
				assert.Nil(t, testCase.testCgu.Object)
			}
		})
	}
}

func TestCguWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		testCgu       *CguBuilder
		expectedError error
	}{
		{
			name:          "waits until deleted after delete",
			testCgu:       buildValidCguTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			name:          "times out when cgu still exists",
			testCgu:       buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "invalid builder returns validation error",
			testCgu:       buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: errInvalidCguMaxConcurrency,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.testCgu.WaitUntilDeleted(time.Second)

			if testCase.expectedError == nil {
				assert.NoError(t, err)
				assert.Nil(t, testCase.testCgu.Object)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestCguWaitForCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		condition     metav1.Condition
		exists        bool
		conditionMet  bool
		valid         bool
		expectedError error
	}{
		{
			name:          "condition met",
			condition:     defaultCguCondition,
			exists:        true,
			conditionMet:  true,
			valid:         true,
			expectedError: nil,
		},
		{
			name:          "cgu does not exist",
			condition:     defaultCguCondition,
			exists:        false,
			conditionMet:  true,
			valid:         true,
			expectedError: errCguObjectNotExists(defaultCguName, defaultCguNsName),
		},
		{
			name:          "condition not met times out",
			condition:     defaultCguCondition,
			exists:        true,
			conditionMet:  false,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "invalid builder returns validation error",
			condition:     defaultCguCondition,
			exists:        true,
			conditionMet:  true,
			valid:         false,
			expectedError: errInvalidCguMaxConcurrency,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

				if testCase.conditionMet {
					cgu.Status.Conditions = append(cgu.Status.Conditions, testCase.condition)
				}

				runtimeObjects = append(runtimeObjects, cgu)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
			})

			var cguBuilder *CguBuilder
			if testCase.valid {
				cguBuilder = buildValidCguTestBuilder(testSettings)
			} else {
				cguBuilder = buildInvalidCguTestBuilder(testSettings)
			}

			_, err := cguBuilder.WaitForCondition(testCase.condition, time.Second)

			if testCase.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestCguWaitUntilComplete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		complete      bool
		expectedError error
	}{
		{
			name:          "cgu complete",
			complete:      true,
			expectedError: nil,
		},
		{
			name:          "cgu not complete times out",
			complete:      false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

			if testCase.complete {
				cgu.Status.Conditions = append(cgu.Status.Conditions, conditionComplete)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{cgu},
				SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
			})

			cguBuilder := buildValidCguTestBuilder(testSettings)
			_, err := cguBuilder.WaitUntilComplete(time.Second)

			if testCase.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

//nolint:funlen // table-driven test with multiple wait scenarios.
func TestCguWaitUntilClusterInState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		cluster       string
		state         string
		exists        bool
		inState       bool
		valid         bool
		expectedError error
	}{
		{
			name:          "cluster in state",
			cluster:       defaultCguClusterName,
			state:         defaultCguClusterState,
			exists:        true,
			inState:       true,
			valid:         true,
			expectedError: nil,
		},
		{
			name:          "empty cluster name",
			cluster:       "",
			state:         defaultCguClusterState,
			exists:        true,
			inState:       true,
			valid:         true,
			expectedError: errClusterNameEmpty,
		},
		{
			name:          "empty state",
			cluster:       defaultCguClusterName,
			state:         "",
			exists:        true,
			inState:       true,
			valid:         true,
			expectedError: errStateEmpty,
		},
		{
			name:          "cgu does not exist",
			cluster:       defaultCguClusterName,
			state:         defaultCguClusterState,
			exists:        false,
			inState:       true,
			valid:         true,
			expectedError: errCguObjectNotExists(defaultCguName, defaultCguNsName),
		},
		{
			name:          "cluster not in state times out",
			cluster:       defaultCguClusterName,
			state:         defaultCguClusterState,
			exists:        true,
			inState:       false,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "invalid builder returns validation error",
			cluster:       defaultCguClusterName,
			state:         defaultCguClusterState,
			exists:        true,
			inState:       true,
			valid:         false,
			expectedError: errInvalidCguMaxConcurrency,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

				if testCase.inState {
					cgu.Status.Status.CurrentBatchRemediationProgress = map[string]*v1alpha1.ClusterRemediationProgress{
						testCase.cluster: {State: testCase.state},
					}
				}

				runtimeObjects = append(runtimeObjects, cgu)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
			})

			var cguBuilder *CguBuilder
			if testCase.valid {
				cguBuilder = buildValidCguTestBuilder(testSettings)
			} else {
				cguBuilder = buildInvalidCguTestBuilder(testSettings)
			}

			_, err := cguBuilder.WaitUntilClusterInState(testCase.cluster, testCase.state, time.Second)

			if testCase.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestCguWaitUntilClusterComplete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		complete      bool
		expectedError error
	}{
		{
			name:          "cluster complete",
			complete:      true,
			expectedError: nil,
		},
		{
			name:          "cluster not complete times out",
			complete:      false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

			if testCase.complete {
				cgu.Status.Status.CurrentBatchRemediationProgress = map[string]*v1alpha1.ClusterRemediationProgress{
					defaultCguClusterName: {State: v1alpha1.Completed},
				}
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{cgu},
				SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
			})

			cguBuilder := buildValidCguTestBuilder(testSettings)
			_, err := cguBuilder.WaitUntilClusterComplete(defaultCguClusterName, time.Second)

			if testCase.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestCguWaitUntilClusterInProgress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		inProgress    bool
		expectedError error
	}{
		{
			name:          "cluster in progress",
			inProgress:    true,
			expectedError: nil,
		},
		{
			name:          "cluster not in progress times out",
			inProgress:    false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

			if testCase.inProgress {
				cgu.Status.Status.CurrentBatchRemediationProgress = map[string]*v1alpha1.ClusterRemediationProgress{
					defaultCguClusterName: {State: v1alpha1.InProgress},
				}
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{cgu},
				SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
			})

			cguBuilder := buildValidCguTestBuilder(testSettings)
			_, err := cguBuilder.WaitUntilClusterInProgress(defaultCguClusterName, time.Second)

			if testCase.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestWaitUntilBackupStarts(t *testing.T) {
	t.Parallel()

	cguObject := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)
	cguObject.Status.Backup = &v1alpha1.BackupStatus{}

	cguBuilder := buildValidCguTestBuilder(clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{cguObject},
		SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
	}))
	cguBuilder, err := cguBuilder.WaitUntilBackupStarts(5 * time.Second)

	assert.Nil(t, err)
	assert.Equal(t, defaultCguName, cguBuilder.Object.Name)
	assert.Equal(t, defaultCguNsName, cguBuilder.Object.Namespace)
}

func buildTestClientWithDummyCguObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyCguObject(),
		SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
	})
}

func buildDummyCguObject() []runtime.Object {
	return []runtime.Object{buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)}
}

func buildDummyCgu(name, namespace string, maxConcurrency int) *v1alpha1.ClusterGroupUpgrade {
	return &v1alpha1.ClusterGroupUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ClusterGroupUpgradeSpec{
			RemediationStrategy: &v1alpha1.RemediationStrategySpec{
				MaxConcurrency: maxConcurrency,
			},
		},
	}
}

func buildValidCguTestBuilder(apiClient *clients.Settings) *CguBuilder {
	return NewCguBuilder(apiClient, defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)
}

func buildInvalidCguTestBuilder(apiClient *clients.Settings) *CguBuilder {
	return NewCguBuilder(apiClient, defaultCguName, defaultCguNsName, 0)
}
