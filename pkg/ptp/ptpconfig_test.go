package ptp

import (
	"testing"

	ptpv1 "github.com/k8snetworkplumbingwg/ptp-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

const (
	defaultPtpConfigName      = "test-ptp-config"
	defaultPtpConfigNamespace = "test-ns"
	testProfileName           = "test"
)

func TestNewPtpConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewPtpConfigBuilder,
		ptpv1.AddToScheme,
		ptpConfigGVK,
	).ExecuteTests(t)
}

func TestPullPtpConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPtpConfig,
		ptpv1.AddToScheme,
		ptpConfigGVK,
	).ExecuteTests(t)
}

func TestPtpConfigMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[ptpv1.PtpConfig, PtpConfigBuilder](
		ptpv1.AddToScheme,
		ptpConfigGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

//nolint:funlen // Table coverage includes all Intel plugin types and package-specific error paths.
func TestGetIntelPlugin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		invalidBuilder   bool
		pluginType       PluginType
		pluginJSON       []byte
		profileExists    bool
		emptyProfileName bool
		expectedError    string
	}{
		{
			name:             "empty profile name",
			emptyProfileName: true,
			expectedError:    "profileName cannot be empty",
		},
		{
			name:          "valid E810 plugin",
			pluginType:    PluginTypeE810,
			pluginJSON:    []byte(`{"enableDefaultConfig": true}`),
			profileExists: true,
		},
		{
			name:          "valid E825 plugin",
			pluginType:    PluginTypeE825,
			pluginJSON:    []byte(`{"enableDefaultConfig": true}`),
			profileExists: true,
		},
		{
			name:          "valid E830 plugin",
			pluginType:    PluginTypeE830,
			pluginJSON:    []byte(`{"enableDefaultConfig": true}`),
			profileExists: true,
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
			profileExists:  true,
		},
		{
			name:          "profile not found",
			expectedError: "ptpProfile test not found",
		},
		{
			name:          "profile has no Intel plugin",
			profileExists: true,
			expectedError: "ptpProfile test does not have an Intel plugin",
		},
		{
			name:          "invalid plugin JSON",
			pluginType:    PluginTypeE810,
			pluginJSON:    []byte(`{'`),
			profileExists: true,
			expectedError: "invalid character '\\'' looking for beginning of object key string",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildTestClientWithPtpScheme()

			testBuilder := buildValidPtpConfigBuilder(apiClient)
			if testCase.invalidBuilder {
				testBuilder = buildInvalidPtpConfigBuilder(apiClient)
			}

			if testCase.profileExists {
				if len(testCase.pluginJSON) > 0 {
					testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{
						buildProfileWithPlugin(testProfileName, testCase.pluginType, testCase.pluginJSON),
					}
				} else {
					testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{{
						Name: ptr.To(testProfileName),
					}}
				}
			}

			profileName := testProfileName
			if testCase.emptyProfileName {
				profileName = ""
			}

			plugin, err := testBuilder.GetIntelPlugin(profileName)
			if testCase.invalidBuilder {
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, plugin)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, plugin)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, plugin)
			assert.Equal(t, testCase.pluginType, plugin.Type)
			assert.True(t, plugin.EnableDefaultConfig)
		})
	}
}

//nolint:funlen // Table coverage includes package-specific validation and invalid-builder short-circuiting.
func TestWithIntelPlugin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		invalidBuilder   bool
		profileExists    bool
		pluginType       PluginType
		emptyProfileName bool
		pluginNil        bool
		expectedError    string
	}{
		{
			name:             "empty profile name",
			profileExists:    true,
			emptyProfileName: true,
			expectedError:    "cannot set Intel plugin: profileName cannot be empty",
		},
		{
			name:          "nil plugin",
			profileExists: true,
			pluginNil:     true,
			expectedError: "cannot set Intel plugin: plugin is nil",
		},
		{
			name:          "valid E810 plugin",
			profileExists: true,
			pluginType:    PluginTypeE810,
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
			profileExists:  true,
			pluginType:     PluginTypeE810,
		},
		{
			name:          "profile does not exist",
			pluginType:    PluginTypeE810,
			expectedError: "cannot set Intel plugin: ptpProfile test does not exist",
		},
		{
			name:          "plugin type not set",
			profileExists: true,
			expectedError: "cannot set Intel plugin: plugin Type is not set",
		},
		{
			name:          "plugin type not supported",
			profileExists: true,
			pluginType:    PluginType("unsupported"),
			expectedError: "cannot set Intel plugin: plugin type unsupported is not supported",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildTestClientWithPtpScheme()

			testBuilder := buildValidPtpConfigBuilder(apiClient)
			if testCase.invalidBuilder {
				testBuilder = buildInvalidPtpConfigBuilder(apiClient)
			}

			if testCase.profileExists {
				testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{{Name: ptr.To(testProfileName)}}
			}

			var plugin *IntelPlugin
			if !testCase.pluginNil {
				plugin = &IntelPlugin{Type: testCase.pluginType, EnableDefaultConfig: true}
			}

			profileName := testProfileName
			if testCase.emptyProfileName {
				profileName = ""
			}

			testBuilder.WithIntelPlugin(profileName, plugin)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.Profile[0].Plugins)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.NotNil(t, testBuilder.Definition.Spec.Profile[0].Plugins[string(testCase.pluginType)])
		})
	}
}

func TestGetPluginType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		invalidBuilder   bool
		pluginType       PluginType
		profileExists    bool
		emptyProfileName bool
		expectedError    string
	}{
		{
			name:             "empty profile name",
			emptyProfileName: true,
			expectedError:    "profileName cannot be empty",
		},
		{
			name:          "valid E810 plugin",
			pluginType:    PluginTypeE810,
			profileExists: true,
		},
		{
			name:          "valid E825 plugin",
			pluginType:    PluginTypeE825,
			profileExists: true,
		},
		{
			name:          "valid E830 plugin",
			pluginType:    PluginTypeE830,
			profileExists: true,
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
			pluginType:     PluginTypeE810,
			profileExists:  true,
		},
		{
			name:          "profile not found",
			expectedError: "ptpProfile test not found",
		},
		{
			name:          "profile has no Intel plugin",
			profileExists: true,
			expectedError: "ptpProfile test does not have an Intel plugin",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildTestClientWithPtpScheme()

			testBuilder := buildValidPtpConfigBuilder(apiClient)
			if testCase.invalidBuilder {
				testBuilder = buildInvalidPtpConfigBuilder(apiClient)
			}

			if testCase.profileExists {
				testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{{
					Name:    ptr.To(testProfileName),
					Plugins: map[string]*apiextensionsv1.JSON{string(testCase.pluginType): {Raw: []byte(`{}`)}},
				}}
			}

			profileName := testProfileName
			if testCase.emptyProfileName {
				profileName = ""
			}

			result, err := testBuilder.GetPluginType(profileName)
			if testCase.invalidBuilder {
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, testCase.pluginType, result)
		})
	}
}

func buildDummyPtpConfig(name, namespace string) *ptpv1.PtpConfig {
	return &ptpv1.PtpConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}
}

func buildTestClientWithPtpScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{ptpv1.AddToScheme},
	})
}

func buildTestClientWithDummyPtpConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyPtpConfig(defaultPtpConfigName, defaultPtpConfigNamespace)},
		SchemeAttachers: []clients.SchemeAttacher{ptpv1.AddToScheme},
	})
}

func buildValidPtpConfigBuilder(apiClient *clients.Settings) *PtpConfigBuilder {
	return NewPtpConfigBuilder(apiClient, defaultPtpConfigName, defaultPtpConfigNamespace)
}

func buildInvalidPtpConfigBuilder(apiClient *clients.Settings) *PtpConfigBuilder {
	return NewPtpConfigBuilder(apiClient, defaultPtpConfigName, "")
}

func buildProfileWithPlugin(name string, pluginType PluginType, raw []byte) ptpv1.PtpProfile {
	return ptpv1.PtpProfile{
		Name: ptr.To(name),
		Plugins: map[string]*apiextensionsv1.JSON{
			string(pluginType): {Raw: raw},
		},
	}
}
