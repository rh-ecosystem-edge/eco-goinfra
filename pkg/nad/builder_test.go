package nad

import (
	"fmt"
	"testing"

	nadV1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultNetName   = "nadtest"
	defaultNetNsName = "nadnamespace"
)

var nadGVK = nadV1.SchemeGroupVersion.WithKind("NetworkAttachmentDefinition")

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(NewBuilder, nadV1.AddToScheme, nadGVK).ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(Pull, nadV1.AddToScheme, nadGVK).ExecuteTests(t)
}

func TestNADBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newNADCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestNADWithMasterPlugin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       func() *Builder
		plugin        *MasterPlugin
		expectedError error
		assertError   func(error) bool
		expectConfig  string
	}{
		{
			name:    "sets master plugin config",
			builder: func() *Builder { return buildValidNADNetworkTestBuilder(newNADTestClient()) },
			plugin:  &MasterPlugin{CniVersion: cniVersion031, Name: "test", Type: cniTypeIpvlan},
		},
		{
			name:          "nil master plugin returns error",
			builder:       func() *Builder { return buildValidNADNetworkTestBuilder(newNADTestClient()) },
			plugin:        nil,
			expectedError: fmt.Errorf("error 'masterPlugin' is empty"),
		},
		{
			name: "does not redefine existing config",
			builder: func() *Builder {
				return buildValidNADNetworkTestBuilder(newNADTestClient()).
					WithPlugins("existing", &[]Plugin{{Name: "existing"}})
			},
			plugin:        &MasterPlugin{CniVersion: cniVersion031, Name: "test", Type: cniTypeIpvlan},
			expectedError: fmt.Errorf("error to redefine predefine NAD"),
			expectConfig:  `"name":"existing"`,
		},
		{
			name:        "invalid builder short circuits",
			builder:     func() *Builder { return buildInvalidNADNetworkTestBuilder(newNADTestClient()) },
			plugin:      &MasterPlugin{CniVersion: cniVersion031, Name: "test", Type: cniTypeIpvlan},
			assertError: commonerrors.IsBuilderNamespaceEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := testCase.builder()
			result := testBuilder.WithMasterPlugin(testCase.plugin)
			require.Same(t, testBuilder, result)

			if testCase.assertError != nil {
				require.Error(t, result.GetError())
				assert.True(t, testCase.assertError(result.GetError()))

				return
			}

			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, result.GetError())
			} else {
				assert.NoError(t, result.GetError())
				assert.Contains(t, result.Definition.Spec.Config, testCase.plugin.Name)
			}

			if testCase.expectConfig != "" {
				assert.Contains(t, result.Definition.Spec.Config, testCase.expectConfig)
			}
		})
	}
}

func TestNADWithPlugins(t *testing.T) {
	t.Parallel()

	plugins := []Plugin{{Name: "test"}, {Name: "test2"}}

	testCases := []struct {
		name        string
		builder     func() *Builder
		assertError func(error) bool
	}{
		{
			name:    "sets plugins config",
			builder: func() *Builder { return buildValidNADNetworkTestBuilder(newNADTestClient()) },
		},
		{
			name:        "invalid builder short circuits",
			builder:     func() *Builder { return buildInvalidNADNetworkTestBuilder(newNADTestClient()) },
			assertError: commonerrors.IsBuilderNamespaceEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := testCase.builder()
			result := testBuilder.WithPlugins("test", &plugins)
			require.Same(t, testBuilder, result)

			if testCase.assertError != nil {
				require.Error(t, result.GetError())
				assert.True(t, testCase.assertError(result.GetError()))

				return
			}

			require.NoError(t, result.GetError())

			for _, plugin := range plugins {
				assert.Contains(t, result.Definition.Spec.Config, plugin.Name)
			}
		})
	}
}

func TestNADGetString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *Builder
		assertError func(error) bool
	}{
		{
			name:    "valid builder",
			builder: func() *Builder { return buildValidNADNetworkTestBuilder(newNADTestClient()) },
		},
		{
			name:        "invalid builder short circuits",
			builder:     func() *Builder { return buildInvalidNADNetworkTestBuilder(newNADTestClient()) },
			assertError: commonerrors.IsBuilderNamespaceEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := testCase.builder()
			result, err := testBuilder.GetString()

			if testCase.assertError != nil {
				require.Error(t, err)
				assert.True(t, testCase.assertError(err))

				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestGetGVR(t *testing.T) {
	t.Parallel()

	testGVR := GetGVR()
	assert.Equal(t, "network-attachment-definitions", testGVR.Resource)
	assert.Equal(t, "v1", testGVR.Version)
	assert.Equal(t, "k8s.cni.cncf.io", testGVR.Group)
}

func newNADCommonTestConfig() testhelper.CommonTestConfig[
	nadV1.NetworkAttachmentDefinition, Builder, *nadV1.NetworkAttachmentDefinition, *Builder] {
	return testhelper.NewCommonTestConfig[nadV1.NetworkAttachmentDefinition, Builder](
		nadV1.AddToScheme, nadGVK, testhelper.ResourceScopeNamespaced)
}

func newNADTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{nadV1.AddToScheme},
	})
}

func buildValidNADNetworkTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultNetName, defaultNetNsName)
}

func buildInvalidNADNetworkTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultNetName, "")
}
