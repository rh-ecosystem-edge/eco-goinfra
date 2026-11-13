package clusteroperator

import (
	"testing"

	configV1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	clusterOperatorGVK = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    APIKind,
	}
	defaultClusterOperatorName = "test-co"
)

func TestClusterOperatorPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		Pull,
		configV1.Install,
		clusterOperatorGVK,
	).ExecuteTests(t)
}

func TestClusterOperatorBuilderMethods(t *testing.T) {
	t.Parallel()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(newClusterOperatorCommonTestConfig())).
		With(testhelper.NewExistsTestConfig(newClusterOperatorCommonTestConfig())).
		Run(t)
}

func TestClusterOperatorHasDesiredVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		desiredVersion      string
		expectedOutput      bool
		testClusterOperator *Builder
		assertError         func(error) bool
	}{
		{
			name:                "empty desired version",
			desiredVersion:      "",
			expectedOutput:      false,
			testClusterOperator: buildValidClusterOperatorBuilder(buildClusterOperatorClientWithDummyObject()),
			assertError: func(err error) bool {
				return err != nil && err.Error() == "desiredVersion can't be empty"
			},
		},
		{
			name:           "desired version matches",
			desiredVersion: "4.14.0",
			expectedOutput: true,
			testClusterOperator: buildFakeClusterOperatorWithVersions(buildClusterOperatorClientWithDummyObject(),
				[]string{"4.13.2", "4.14.0"}),
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:                "nil builder",
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			testClusterOperator: nil,
			assertError:         commonerrors.IsBuilderNil,
		},
		{
			name:           "desired version does not match",
			desiredVersion: "4.14.0",
			expectedOutput: false,
			testClusterOperator: buildFakeClusterOperatorWithVersions(buildClusterOperatorClientWithDummyObject(),
				[]string{"4.13.2", "4.13.0"}),
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:           "invalid version does not match",
			desiredVersion: "4.14.0",
			expectedOutput: false,
			testClusterOperator: buildFakeClusterOperatorWithVersions(buildClusterOperatorClientWithDummyObject(),
				[]string{"4.13.2", "invalid"}),
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:                "undefined status versions",
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			testClusterOperator: buildFakeClusterOperatorWithVersions(buildClusterOperatorClientWithDummyObject(), []string{}),
			assertError: func(err error) bool {
				return err != nil && err.Error() == "undefined cluster operator status versions"
			},
		},
		{
			name:                "nil client",
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			testClusterOperator: buildNilClientClusterOperatorBuilder(),
			assertError:         commonerrors.IsAPIClientNil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := testCase.testClusterOperator.HasDesiredVersion(testCase.desiredVersion)
			assert.Equal(t, testCase.expectedOutput, result)
			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)
		})
	}
}

func buildValidClusterOperatorBuilder(apiClient *clients.Settings) *Builder {
	return newBuilder(apiClient, defaultClusterOperatorName, configV1.ClusterOperatorStatus{})
}

func buildNilClientClusterOperatorBuilder() *Builder {
	return newBuilder(nil, defaultClusterOperatorName, configV1.ClusterOperatorStatus{})
}

func buildClusterOperatorClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterOperatorConfig(),
		SchemeAttachers: []clients.SchemeAttacher{configV1.Install},
	})
}

func buildDummyClusterOperatorConfig() []runtime.Object {
	return append([]runtime.Object{}, &configV1.ClusterOperator{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultClusterOperatorName,
		},
		Spec: configV1.ClusterOperatorSpec{},
	})
}

func buildFakeClusterOperatorListWithDesiredVersion(apiClient *clients.Settings) []*Builder {
	desiredVersion := "4.14.0"

	return buildFakeClusterOperatorListWithVersions(apiClient, []string{desiredVersion,
		"4.13.24", "4.13.25"}, []string{"4.13.23", desiredVersion, "4.13.25"})
}

func buildFakeClusterOperatorListWithoutDesiredVersion(apiClient *clients.Settings) []*Builder {
	return buildFakeClusterOperatorListWithVersions(apiClient, []string{"4.13.23",
		"4.13.24", "4.13.25"}, []string{"4.13.23", "", "4.13.25"})
}

func buildFakeClusterOperatorListWithVersions(apiClient *clients.Settings,
	versionList1, versionList2 []string) []*Builder {
	listWithVersions := []*Builder{buildFakeClusterOperatorWithVersions(apiClient, versionList1),
		buildFakeClusterOperatorWithVersions(apiClient, versionList2)}

	return listWithVersions
}

// buildFakeClusterOperatorWithVersions creates a fake clusterOperator with
// a specific list of ClusterOperatorStatusVersions for testing purpose.
func buildFakeClusterOperatorWithVersions(apiClient *clients.Settings, versions []string) *Builder {
	clusterOperatorStatusVersions := createFakeClusterOperatorStatusVersions(versions)

	clusterOperatorStatus := createFakeClusterOperatorStatus(clusterOperatorStatusVersions)

	clusterOperator := newBuilder(apiClient, defaultClusterOperatorName, clusterOperatorStatus)
	clusterOperator.Object = clusterOperator.Definition

	return clusterOperator
}

// createFakeClusterOperatorStatusVersions creates a fake list of ClusterOperatorStatusVersions.
func createFakeClusterOperatorStatusVersions(versions []string) []configV1.OperandVersion {
	var clusterOperatorStatusVersions []configV1.OperandVersion

	for _, version := range versions {
		clusterOperatorStatusVersions = append(clusterOperatorStatusVersions,
			configV1.OperandVersion{Name: defaultClusterOperatorName, Version: version})
	}

	return clusterOperatorStatusVersions
}

// createFakeClusterOperatorStatus creates a fake lusterOperatorStatus containing a list of OperandVersions.
func createFakeClusterOperatorStatus(
	clusterOperatorStatusVersions []configV1.OperandVersion) configV1.ClusterOperatorStatus {
	clusterOperatorStatus := configV1.ClusterOperatorStatus{
		Versions: clusterOperatorStatusVersions,
	}

	return clusterOperatorStatus
}

// newBuilder creates a new Builder instance for tests that need a prefilled status.
func newBuilder(apiClient *clients.Settings, name string, status configV1.ClusterOperatorStatus) *Builder {
	builder := common.NewClusterScopedBuilder[configV1.ClusterOperator, Builder](
		apiClient, configV1.Install, name)
	builder.Definition.ResourceVersion = "999"
	builder.Definition.Status = status

	return builder
}

func newClusterOperatorCommonTestConfig() testhelper.CommonTestConfig[
	configV1.ClusterOperator, Builder, *configV1.ClusterOperator, *Builder] {
	return testhelper.NewCommonTestConfig[configV1.ClusterOperator, Builder](
		configV1.Install, clusterOperatorGVK, testhelper.ResourceScopeClusterScoped)
}
