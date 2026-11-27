package nodesconfig

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
	nodesConfigGVK         = configV1.GroupVersion.WithKind(APIKind)
	defaultNodesConfigName = "cluster"
)

func TestNodesConfigPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		Pull,
		configV1.Install,
		nodesConfigGVK,
	).ExecuteTests(t)
}

func TestNodesConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newNodesConfigCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestNodesConfigGetCGroupMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		builder      *Builder
		expectedMode configV1.CgroupMode
		assertError  func(error) bool
	}{
		{
			name:         "returns configured cgroup mode",
			builder:      buildValidNodeConfigBuilder(buildNodeConfigClientWithDummyObject(configV1.CgroupModeV2)),
			expectedMode: configV1.CgroupModeV2,
			assertError:  func(err error) bool { return err == nil },
		},
		{
			name:    "missing resource returns error",
			builder: buildValidNodeConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			assertError: func(err error) bool {
				return err != nil && err.Error() == "nodesConfig object does not exist"
			},
		},
		{
			name:        "nil builder returns error",
			assertError: commonerrors.IsBuilderNil,
		},
		{
			name: "invalid builder returns error",
			builder: common.NewClusterScopedBuilder[configV1.Node, Builder](
				clients.GetTestClients(clients.TestClientParams{}), configV1.Install, ""),
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cgroupMode, err := testCase.builder.GetCGroupMode()
			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)

			if err == nil {
				assert.Equal(t, testCase.expectedMode, cgroupMode)
			}
		})
	}
}

func TestNodesConfigWithCGroupMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		cgroupMode     configV1.CgroupMode
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:       "sets cgroup mode",
			cgroupMode: configV1.CgroupModeV2,
		},
		{
			name:          "empty cgroup mode sets builder error",
			cgroupMode:    configV1.CgroupModeEmpty,
			expectedError: "the cGroupMode value can not be empty",
		},
		{
			name:           "invalid builder short circuits",
			cgroupMode:     configV1.CgroupModeV2,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidNodeConfigBuilder(clients.GetTestClients(clients.TestClientParams{}))
			if testCase.invalidBuilder {
				testBuilder = common.NewClusterScopedBuilder[configV1.Node, Builder](
					clients.GetTestClients(clients.TestClientParams{}), configV1.Install, "")
			}

			result := testBuilder.WithCGroupMode(testCase.cgroupMode)
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
			assert.Equal(t, testCase.cgroupMode, result.Definition.Spec.CgroupMode)
		})
	}
}

func TestGetNodesConfigGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, GetNodesConfigIoGVR(), schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: APIKind,
	})
}

func buildValidNodeConfigBuilder(apiClient *clients.Settings) *Builder {
	return common.NewClusterScopedBuilder[configV1.Node, Builder](
		apiClient, configV1.Install, defaultNodesConfigName)
}

func buildNodeConfigClientWithDummyObject(cgroupMode configV1.CgroupMode) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{&configV1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultNodesConfigName,
			},
			Spec: configV1.NodeSpec{
				CgroupMode: cgroupMode,
			},
		}},
		SchemeAttachers: []clients.SchemeAttacher{configV1.Install},
	})
}

func newNodesConfigCommonTestConfig() testhelper.CommonTestConfig[
	configV1.Node, Builder, *configV1.Node, *Builder] {
	return testhelper.NewCommonTestConfig[configV1.Node, Builder](
		configV1.Install, nodesConfigGVK, testhelper.ResourceScopeClusterScoped)
}
