package ibgu

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/imagebasedgroupupgrades/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testIbguName      = "test-ibgu"
	testIbguNamespace = "test-namespace"
)

func TestNewIbguBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewIbguBuilder, v1alpha1.AddToScheme, ibguGVK).ExecuteTests(t)

	t.Run("initializes an empty plan", func(t *testing.T) {
		t.Parallel()

		builder := buildValidIbguTestBuilder(getTestIbguAPIClient())

		require.NotNil(t, builder.Definition.Spec.Plan)
		assert.Empty(t, builder.Definition.Spec.Plan)
	})
}

func TestPullIbgu(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullIbgu, v1alpha1.AddToScheme, ibguGVK).ExecuteTests(t)
}

func TestIbguBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[v1alpha1.ImageBasedGroupUpgrade, IbguBuilder](
		v1alpha1.AddToScheme, ibguGVK, testhelper.ResourceScopeNamespaced)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestIbguWithClusterLabelSelectors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		labels        map[string]string
		expectedError string
	}{
		{
			name:   "labels are set",
			valid:  true,
			labels: map[string]string{"key": "value"},
		},
		{
			name:          "empty labels return error",
			valid:         true,
			labels:        map[string]string{},
			expectedError: "can not apply empty cluster label selectors to the IBGU",
		},
		{
			name:   "invalid builder short circuits",
			valid:  false,
			labels: map[string]string{"key": "value"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildIbguTestBuilder(getTestIbguAPIClient(), testCase.valid)
			result := builder.WithClusterLabelSelectors(testCase.labels)

			require.Same(t, builder, result)

			if !testCase.valid {
				require.Truef(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()),
					"unexpected error: %v", result.GetError())
				assert.Nil(t, result.Definition.Spec.ClusterLabelSelectors)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			require.Len(t, result.Definition.Spec.ClusterLabelSelectors, 1)
			assert.Equal(t, testCase.labels, result.Definition.Spec.ClusterLabelSelectors[0].MatchLabels)
		})
	}
}

func TestIbguWithAutoRollbackOnFailure(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		valid              bool
		initMonitorTimeout int
		expectedError      string
	}{
		{
			name:               "timeout is set",
			valid:              true,
			initMonitorTimeout: 10,
		},
		{
			name:               "negative timeout returns error",
			valid:              true,
			initMonitorTimeout: -1,
			expectedError:      "initMonitorTimeout cannot be undefined",
		},
		{
			name:               "invalid builder short circuits",
			valid:              false,
			initMonitorTimeout: 10,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildIbguTestBuilder(getTestIbguAPIClient(), testCase.valid)
			result := builder.WithAutoRollbackOnFailure(testCase.initMonitorTimeout)

			require.Same(t, builder, result)

			if !testCase.valid {
				require.Truef(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()),
					"unexpected error: %v", result.GetError())
				assert.Nil(t, result.Definition.Spec.IBUSpec.AutoRollbackOnFailure)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Spec.IBUSpec.AutoRollbackOnFailure)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.IBUSpec.AutoRollbackOnFailure)
			assert.Equal(t, testCase.initMonitorTimeout,
				result.Definition.Spec.IBUSpec.AutoRollbackOnFailure.InitMonitorTimeoutSeconds)
		})
	}
}

func TestIbguWithSeedImageRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		seedImage     string
		seedVersion   string
		expectedError string
	}{
		{
			name:        "seed image reference is set",
			valid:       true,
			seedImage:   "test-image",
			seedVersion: "v1.0",
		},
		{
			name:          "empty image returns error",
			valid:         true,
			seedVersion:   "v1.0",
			expectedError: "seedImage cannot be empty",
		},
		{
			name:          "empty version returns error",
			valid:         true,
			seedImage:     "test-image",
			expectedError: "seedVersion cannot be empty",
		},
		{
			name:        "invalid builder short circuits",
			valid:       false,
			seedImage:   "test-image",
			seedVersion: "v1.0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildIbguTestBuilder(getTestIbguAPIClient(), testCase.valid)
			result := builder.WithSeedImageRef(testCase.seedImage, testCase.seedVersion)

			require.Same(t, builder, result)

			if !testCase.valid {
				require.Truef(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()),
					"unexpected error: %v", result.GetError())
				assert.Empty(t, result.Definition.Spec.IBUSpec.SeedImageRef)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Empty(t, result.Definition.Spec.IBUSpec.SeedImageRef)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.seedImage, result.Definition.Spec.IBUSpec.SeedImageRef.Image)
			assert.Equal(t, testCase.seedVersion, result.Definition.Spec.IBUSpec.SeedImageRef.Version)
		})
	}
}

func TestIbguWithOadpContent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		oadpName      string
		oadpNamespace string
		expectedError string
	}{
		{
			name:          "OADP content is appended",
			valid:         true,
			oadpName:      "test-oadp",
			oadpNamespace: "test-ns",
		},
		{
			name:          "empty name returns error",
			valid:         true,
			oadpNamespace: "test-ns",
			expectedError: "oadp content name cannot be empty",
		},
		{
			name:          "empty namespace returns error",
			valid:         true,
			oadpName:      "test-oadp",
			expectedError: "oadp content namespace cannot be empty",
		},
		{
			name:          "invalid builder short circuits",
			valid:         false,
			oadpName:      "test-oadp",
			oadpNamespace: "test-ns",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildIbguTestBuilder(getTestIbguAPIClient(), testCase.valid)
			result := builder.WithOadpContent(testCase.oadpName, testCase.oadpNamespace)

			require.Same(t, builder, result)

			if !testCase.valid {
				require.Truef(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()),
					"unexpected error: %v", result.GetError())
				assert.Empty(t, result.Definition.Spec.IBUSpec.OADPContent)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Empty(t, result.Definition.Spec.IBUSpec.OADPContent)

				return
			}

			require.NoError(t, result.GetError())
			require.Len(t, result.Definition.Spec.IBUSpec.OADPContent, 1)
			assert.Equal(t, testCase.oadpName, result.Definition.Spec.IBUSpec.OADPContent[0].Name)
			assert.Equal(t, testCase.oadpNamespace, result.Definition.Spec.IBUSpec.OADPContent[0].Namespace)
		})
	}
}

func TestIbguWithPlan(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		valid          bool
		actions        []string
		maxConcurrency int
		timeout        int
		expectedError  string
	}{
		{
			name:           "plan is appended",
			valid:          true,
			actions:        []string{"action1", "action2"},
			maxConcurrency: 2,
			timeout:        300,
		},
		{
			name:           "empty actions return error",
			valid:          true,
			actions:        []string{},
			maxConcurrency: 2,
			timeout:        300,
			expectedError:  "plan actions cannot be empty",
		},
		{
			name:          "zero max concurrency returns error",
			valid:         true,
			actions:       []string{"action1"},
			timeout:       300,
			expectedError: "maxConcurrency must be greater than 0",
		},
		{
			name:           "zero timeout returns error",
			valid:          true,
			actions:        []string{"action1"},
			maxConcurrency: 2,
			expectedError:  "timeout must be greater than 0",
		},
		{
			name:           "invalid builder short circuits",
			valid:          false,
			actions:        []string{"action1"},
			maxConcurrency: 2,
			timeout:        300,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildIbguTestBuilder(getTestIbguAPIClient(), testCase.valid)
			result := builder.WithPlan(testCase.actions, testCase.maxConcurrency, testCase.timeout)

			require.Same(t, builder, result)

			if !testCase.valid {
				require.Truef(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()),
					"unexpected error: %v", result.GetError())
				assert.Empty(t, result.Definition.Spec.Plan)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Empty(t, result.Definition.Spec.Plan)

				return
			}

			require.NoError(t, result.GetError())
			require.Len(t, result.Definition.Spec.Plan, 1)
			assert.Equal(t, testCase.actions, result.Definition.Spec.Plan[0].Actions)
			assert.Equal(t, testCase.maxConcurrency, result.Definition.Spec.Plan[0].RolloutStrategy.MaxConcurrency)
			assert.Equal(t, testCase.timeout, result.Definition.Spec.Plan[0].RolloutStrategy.Timeout)
		})
	}
}

func TestIbguDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		ibguExists    bool
		expectedError error
	}{
		{
			name:       "existing IBGU is deleted",
			valid:      true,
			ibguExists: true,
		},
		{
			name:  "absent IBGU succeeds",
			valid: true,
		},
		{
			name:       "invalid builder returns error",
			ibguExists: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := getTestIbguAPIClient()
			if testCase.ibguExists {
				apiClient = buildTestClientWithDummyIbgu()
			}

			builder := buildIbguTestBuilder(apiClient, testCase.valid)
			result, err := builder.DeleteAndWait(time.Second)

			require.Same(t, builder, result)

			if !testCase.valid {
				require.Truef(t, commonerrors.IsBuilderNamespaceEmpty(err), "unexpected error: %v", err)

				return
			}

			require.Equal(t, testCase.expectedError, err)
			assert.Nil(t, builder.Object)
		})
	}
}

func TestIbguWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		ibguExists    bool
		expectedError error
	}{
		{
			name:  "IBGU already absent",
			valid: true,
		},
		{
			name:          "IBGU remains present",
			valid:         true,
			ibguExists:    true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:       "invalid builder returns error",
			ibguExists: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := getTestIbguAPIClient()
			if testCase.ibguExists {
				apiClient = buildTestClientWithDummyIbgu()
			}

			builder := buildIbguTestBuilder(apiClient, testCase.valid)
			err := builder.WaitUntilDeleted(time.Second)

			if !testCase.valid {
				require.Truef(t, commonerrors.IsBuilderNamespaceEmpty(err), "unexpected error: %v", err)

				return
			}

			require.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestIbguWaitForCondition(t *testing.T) {
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
			name:         "condition is met",
			condition:    conditionComplete,
			exists:       true,
			conditionMet: true,
			valid:        true,
		},
		{
			name:         "IBGU does not exist",
			condition:    conditionComplete,
			conditionMet: true,
			valid:        true,
			expectedError: fmt.Errorf(
				"ibgu object %s does not exist in namespace %s", testIbguName, testIbguNamespace),
		},
		{
			name:          "condition is not met",
			condition:     conditionComplete,
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:         "invalid builder returns error",
			condition:    conditionComplete,
			exists:       true,
			conditionMet: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				ibgu := buildDummyIbgu(testIbguName, testIbguNamespace)

				if testCase.conditionMet {
					ibgu.Status.Conditions = append(ibgu.Status.Conditions, testCase.condition)
				}

				runtimeObjects = append(runtimeObjects, ibgu)
			}

			apiClient := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
			})

			builder := buildIbguTestBuilder(apiClient, testCase.valid)
			result, err := builder.WaitForCondition(testCase.condition, time.Second)

			require.Same(t, builder, result)

			if !testCase.valid {
				require.Truef(t, commonerrors.IsBuilderNamespaceEmpty(err), "unexpected error: %v", err)

				return
			}

			require.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestIbguWaitUntilComplete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		complete      bool
		expectedError error
	}{
		{
			name:     "IBGU is complete",
			complete: true,
		},
		{
			name:          "IBGU does not complete",
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ibgu := buildDummyIbgu(testIbguName, testIbguNamespace)

			if testCase.complete {
				ibgu.Status.Conditions = append(ibgu.Status.Conditions, conditionComplete)
			}

			apiClient := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{ibgu},
				SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
			})

			builder := buildValidIbguTestBuilder(apiClient)
			_, err := builder.WaitUntilComplete(time.Second)

			require.Equal(t, testCase.expectedError, err)
		})
	}
}

func getTestIbguAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}

func buildValidIbguTestBuilder(apiClient *clients.Settings) *IbguBuilder {
	return NewIbguBuilder(apiClient, testIbguName, testIbguNamespace)
}

func buildInvalidIbguTestBuilder(apiClient *clients.Settings) *IbguBuilder {
	return NewIbguBuilder(apiClient, testIbguName, "")
}

func buildIbguTestBuilder(apiClient *clients.Settings, valid bool) *IbguBuilder {
	if valid {
		return buildValidIbguTestBuilder(apiClient)
	}

	return buildInvalidIbguTestBuilder(apiClient)
}

func buildTestClientWithDummyIbgu() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyIbgu(testIbguName, testIbguNamespace)},
		SchemeAttachers: []clients.SchemeAttacher{v1alpha1.AddToScheme},
	})
}

func buildDummyIbgu(name, namespace string) *v1alpha1.ImageBasedGroupUpgrade {
	return &v1alpha1.ImageBasedGroupUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
