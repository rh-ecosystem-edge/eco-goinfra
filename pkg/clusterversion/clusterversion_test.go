package clusterversion

import (
	"context"
	"fmt"
	"testing"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultRelease = configv1.Release{
		Version: "4.17.1",
		Image:   "test-image",
	}
	testSchemes = []clients.SchemeAttacher{
		configv1.Install,
	}
	clusterVersionGVK = configv1.GroupVersion.WithKind("ClusterVersion")
)

func TestPullClusterVersion(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		Pull,
		configv1.Install,
		clusterVersionGVK,
		clusterVersionName,
	).ExecuteTests(t)
}

func TestClusterVersionBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[configv1.ClusterVersion, Builder](
		configv1.Install, clusterVersionGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestClusterVersionWithDesiredUpdateImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		desiredUpdateImage string
		force              bool
		expectedError      string
		invalidBuilder     bool
	}{
		{
			name:               "sets desired update image",
			desiredUpdateImage: "test-image",
			force:              true,
			expectedError:      "",
		},
		{
			name:               "empty image sets builder error",
			desiredUpdateImage: "",
			force:              true,
			expectedError:      "clusterversion 'desiredUpdateImage' cannot be empty",
		},
		{
			name:               "invalid builder short circuits",
			desiredUpdateImage: "test-image",
			force:              true,
			invalidBuilder:     true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := newClusterVersionBuilder(clients.GetTestClients(clients.TestClientParams{}))
			if testCase.invalidBuilder {
				testBuilder = common.NewClusterScopedBuilder[configv1.ClusterVersion, Builder](
					clients.GetTestClients(clients.TestClientParams{}), configv1.Install, "")
			}

			result := testBuilder.WithDesiredUpdateImage(testCase.desiredUpdateImage, testCase.force)
			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.desiredUpdateImage, result.Definition.Spec.DesiredUpdate.Image)
			assert.Equal(t, testCase.force, result.Definition.Spec.DesiredUpdate.Force)
		})
	}
}

func TestClusterVersionWithDesiredUpdateChannel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		desiredUpdateChannel string
		expectedError        string
		invalidBuilder       bool
	}{
		{
			name:                 "sets desired update channel",
			desiredUpdateChannel: "test-channel",
			expectedError:        "",
		},
		{
			name:                 "empty channel sets builder error",
			desiredUpdateChannel: "",
			expectedError:        "clusterversion 'updateChannel' cannot be empty",
		},
		{
			name:                 "invalid builder short circuits",
			desiredUpdateChannel: "test-channel",
			invalidBuilder:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := newClusterVersionBuilder(clients.GetTestClients(clients.TestClientParams{}))
			if testCase.invalidBuilder {
				testBuilder = common.NewClusterScopedBuilder[configv1.ClusterVersion, Builder](
					clients.GetTestClients(clients.TestClientParams{}), configv1.Install, "")
			}

			result := testBuilder.WithDesiredUpdateChannel(testCase.desiredUpdateChannel)
			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.desiredUpdateChannel, result.Definition.Spec.Channel)
		})
	}
}

func TestClusterVersionWaitUntilProgressing(t *testing.T) {
	t.Parallel()

	testWaitUntilConditionTrueHelper(t, configv1.OperatorProgressing, func(builder *Builder) error {
		return builder.WaitUntilProgressing(time.Second)
	})
}

func TestClusterVersionWaitUntilAvailable(t *testing.T) {
	t.Parallel()

	testWaitUntilConditionTrueHelper(t, configv1.OperatorAvailable, func(builder *Builder) error {
		return builder.WaitUntilAvailable(time.Second)
	})
}

func TestClusterVersionWaitUntilConditionTrue(t *testing.T) {
	t.Parallel()

	testWaitUntilConditionTrueHelper(t, configv1.OperatorAvailable, func(builder *Builder) error {
		return builder.WaitUntilConditionTrue(configv1.OperatorAvailable, time.Second)
	})
}

func TestClusterVersionWaitUntilUpdateIsStarted(t *testing.T) {
	t.Parallel()

	testWaitUntilUpdateHistoryStateHelper(t, configv1.PartialUpdate, func(builder *Builder) error {
		return builder.WaitUntilUpdateIsStarted(time.Second)
	})
}

func TestClusterVersionWaitUntilUpdateIsCompleted(t *testing.T) {
	t.Parallel()

	testWaitUntilUpdateHistoryStateHelper(t, configv1.CompletedUpdate, func(builder *Builder) error {
		return builder.WaitUntilUpdateIsCompleted(time.Second)
	})
}

func TestClusterVersionWaitUntilUpdateHistoryStateTrue(t *testing.T) {
	t.Parallel()

	testWaitUntilUpdateHistoryStateHelper(t, configv1.CompletedUpdate, func(builder *Builder) error {
		return builder.WaitUntilUpdateHistoryStateTrue(configv1.CompletedUpdate, time.Second)
	})
}

func TestClusterVersionGetNextUpdateVersionImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		stream            string
		acceptConditional bool
		exists            bool
		hasAvailable      bool
		hasConditional    bool
		expectedError     error
	}{
		{
			name:              "available Z stream update",
			stream:            Z,
			acceptConditional: false,
			exists:            true,
			hasAvailable:      true,
			hasConditional:    false,
			expectedError:     nil,
		},
		{
			name:              "conditional Z stream update",
			stream:            Z,
			acceptConditional: true,
			exists:            true,
			hasAvailable:      false,
			hasConditional:    true,
			expectedError:     nil,
		},
		{
			name:              "empty stream returns error",
			stream:            "",
			acceptConditional: false,
			exists:            true,
			hasAvailable:      true,
			hasConditional:    false,
			expectedError:     fmt.Errorf("stream can not be empty"),
		},
		{
			name:              "missing clusterversion returns error",
			stream:            Z,
			acceptConditional: false,
			exists:            false,
			hasAvailable:      false,
			hasConditional:    false,
			expectedError:     fmt.Errorf("clusterversion object %s does not exist", clusterVersionName),
		},
		{
			name:              "matching update not found returns error",
			stream:            Z,
			acceptConditional: false,
			exists:            true,
			hasAvailable:      false,
			hasConditional:    false,
			expectedError:     fmt.Errorf("update version in Z stream not found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: buildClusterVersionUpdateObjects(
					testCase.exists, testCase.hasAvailable, testCase.hasConditional),
				SchemeAttachers: testSchemes,
			})
			testBuilder := newClusterVersionBuilder(testSettings)

			nextUpdate, err := testBuilder.GetNextUpdateVersionImage(testCase.stream, testCase.acceptConditional)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, "test-image", nextUpdate)
			}
		})
	}
}

func buildClusterVersionUpdateObjects(exists, hasAvailable, hasConditional bool) []runtime.Object {
	if !exists {
		return nil
	}

	clusterVersion := buildDummyClusterVersion()
	clusterVersion.Status.Desired.Version = "4.17.0"

	if hasAvailable {
		clusterVersion.Status.AvailableUpdates = []configv1.Release{defaultRelease}
	}

	if hasConditional {
		clusterVersion.Status.ConditionalUpdates = []configv1.ConditionalUpdate{{Release: defaultRelease}}
	}

	return []runtime.Object{clusterVersion}
}

// testWaitUntilConditionTrueHelper is a helper for functions that call WaitUntilConditionTrue. It will run the test
// cases and ensure the provided condition type is set then call testFunc and compare errors.
func testWaitUntilConditionTrueHelper(
	t *testing.T, condType configv1.ClusterStatusConditionType, testFunc func(builder *Builder) error) {
	t.Helper()

	testCases := []struct {
		name          string
		exists        bool
		hasCondType   bool
		expectedError error
	}{
		{
			name:          "condition is true",
			exists:        true,
			hasCondType:   true,
			expectedError: nil,
		},
		{
			name:          "clusterversion does not exist",
			exists:        false,
			hasCondType:   true,
			expectedError: fmt.Errorf("clusterversion object %s does not exist", clusterVersionName),
		},
		{
			name:          "condition is missing",
			exists:        true,
			hasCondType:   false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				clusterVersion := buildDummyClusterVersion()

				if testCase.hasCondType {
					clusterVersion.Status.Conditions = []configv1.ClusterOperatorStatusCondition{{
						Type:   condType,
						Status: configv1.ConditionTrue,
					}}
				}

				runtimeObjects = append(runtimeObjects, clusterVersion)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
			testBuilder := newClusterVersionBuilder(testSettings)

			err := testFunc(testBuilder)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

// testWaitUntilUpdateHistoryStateHelper is a helper for functions that call WaitUntilUpdateHistoryStateTrue. It will
// run the test cases and set the provided state then call testFunc and compare errors.
func testWaitUntilUpdateHistoryStateHelper(
	t *testing.T, state configv1.UpdateState, testFunc func(builder *Builder) error) {
	t.Helper()

	testCases := []struct {
		name          string
		exists        bool
		inState       bool
		expectedError error
	}{
		{
			name:          "update history state is true",
			exists:        true,
			inState:       true,
			expectedError: nil,
		},
		{
			name:          "clusterversion does not exist",
			exists:        false,
			inState:       true,
			expectedError: fmt.Errorf("clusterversion object %s does not exist", clusterVersionName),
		},
		{
			name:          "update history state is not true",
			exists:        true,
			inState:       false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				clusterVersion := buildDummyClusterVersion()
				clusterVersion.Status.Desired.Image = "test-image"

				if testCase.inState {
					clusterVersion.Status.History = []configv1.UpdateHistory{{
						Image: "test-image",
						State: state,
					}}
				}

				runtimeObjects = append(runtimeObjects, clusterVersion)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
			testBuilder := newClusterVersionBuilder(testSettings)

			err := testFunc(testBuilder)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

// buildDummyClusterVersion returns a ClusterVersion with the clusterVersionName.
func buildDummyClusterVersion() *configv1.ClusterVersion {
	return &configv1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterVersionName,
		},
	}
}

// newClusterVersionBuilder returns a Builder with the provided apiClient and the clusterVersionName.
func newClusterVersionBuilder(apiClient *clients.Settings) *Builder {
	return common.NewClusterScopedBuilder[configv1.ClusterVersion, Builder](
		apiClient, configv1.Install, clusterVersionName)
}
