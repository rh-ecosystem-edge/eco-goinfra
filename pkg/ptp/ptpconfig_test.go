package ptp

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	ptpv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ptp/v1"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

const (
	defaultPtpConfigName      = "test-ptp-config"
	defaultPtpConfigNamespace = "test-ns"
)

var testSchemes = []clients.SchemeAttacher{
	ptpv1.AddToScheme,
}

func TestNewPtpConfigBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		nsname        string
		client        bool
		expectedError string
	}{
		{
			name:          defaultPtpConfigName,
			nsname:        defaultPtpConfigNamespace,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			nsname:        defaultPtpConfigNamespace,
			client:        true,
			expectedError: "ptpConfig 'name' cannot be empty",
		},
		{
			name:          defaultPtpConfigName,
			nsname:        "",
			client:        true,
			expectedError: "ptpConfig 'nsname' cannot be empty",
		},
		{
			name:          defaultPtpConfigName,
			nsname:        defaultPtpConfigNamespace,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithPtpScheme()
		}

		testBuilder := NewPtpConfigBuilder(testSettings, testCase.name, testCase.nsname)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

//nolint:funlen // long due to the number of test cases
func TestGetIntelPlugin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		ptpConfigValid bool
		pluginType     PluginType
		pluginJSON     []byte
		profileExists  bool
		expectedError  string
	}{
		{
			name:           "valid E810 plugin",
			ptpConfigValid: true,
			pluginType:     PluginTypeE810,
			pluginJSON:     []byte(`{"enableDefaultConfig": true}`),
			profileExists:  true,
			expectedError:  "",
		},
		{
			name:           "valid E825 plugin",
			ptpConfigValid: true,
			pluginType:     PluginTypeE825,
			pluginJSON:     []byte(`{"enableDefaultConfig": true}`),
			profileExists:  true,
			expectedError:  "",
		},
		{
			name:           "valid E830 plugin",
			ptpConfigValid: true,
			pluginType:     PluginTypeE830,
			pluginJSON:     []byte(`{"enableDefaultConfig": true}`),
			profileExists:  true,
			expectedError:  "",
		},
		{
			name:           "invalid ptpConfig",
			ptpConfigValid: false,
			profileExists:  true,
			expectedError:  "ptpConfig 'nsname' cannot be empty",
		},
		{
			name:           "profile not found",
			ptpConfigValid: true,
			profileExists:  false,
			expectedError:  "ptpProfile test not found",
		},
		{
			name:           "no Intel plugin",
			ptpConfigValid: true,
			profileExists:  true,
			expectedError:  "ptpProfile test does not have an Intel plugin",
		},
		{
			name:           "invalid plugin JSON",
			ptpConfigValid: true,
			pluginType:     PluginTypeE810,
			pluginJSON:     []byte(`{'`),
			profileExists:  true,
			expectedError:  "invalid character '\\'' looking for beginning of object key string",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testSettings := buildTestClientWithPtpScheme()
			testBuilder := buildValidPtpConfigBuilder(testSettings)

			if !testCase.ptpConfigValid {
				testBuilder = buildInvalidPtpConfigBuilder(testSettings)
			}

			if testCase.profileExists {
				if len(testCase.pluginJSON) > 0 {
					testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{
						buildProfileWithPlugin("test", testCase.pluginType, testCase.pluginJSON),
					}
				} else {
					testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{{Name: ptr.To("test"), Plugins: nil}}
				}
			}

			plugin, err := testBuilder.GetIntelPlugin("test")

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, plugin)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, plugin)
				assert.Equal(t, testCase.pluginType, plugin.Type)
				assert.True(t, plugin.EnableDefaultConfig)
			}
		})
	}
}

func TestWithIntelPlugin(t *testing.T) {
	testCases := []struct {
		name           string
		ptpConfigValid bool
		profileExists  bool
		pluginType     PluginType
		expectedError  string
	}{
		{
			name:           "valid E810 plugin",
			ptpConfigValid: true,
			profileExists:  true,
			pluginType:     PluginTypeE810,
			expectedError:  "",
		},
		{
			name:           "invalid ptpConfig",
			ptpConfigValid: false,
			profileExists:  true,
			pluginType:     PluginTypeE810,
			expectedError:  "ptpConfig 'nsname' cannot be empty",
		},
		{
			name:           "profile does not exist",
			ptpConfigValid: true,
			profileExists:  false,
			pluginType:     PluginTypeE810,
			expectedError:  "cannot set Intel plugin: ptpProfile test does not exist",
		},
		{
			name:           "plugin type not set",
			ptpConfigValid: true,
			profileExists:  true,
			pluginType:     "",
			expectedError:  "cannot set Intel plugin: plugin Type is not set",
		},
		{
			name:           "plugin type not supported",
			ptpConfigValid: true,
			profileExists:  true,
			pluginType:     PluginType("unsupported"),
			expectedError:  "cannot set Intel plugin: plugin type unsupported is not supported",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testSettings := buildTestClientWithPtpScheme()
			testBuilder := buildValidPtpConfigBuilder(testSettings)

			if !testCase.ptpConfigValid {
				testBuilder = buildInvalidPtpConfigBuilder(testSettings)
			}

			if testCase.profileExists {
				testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{{Name: ptr.To("test")}}
			}

			plugin := &IntelPlugin{
				Type:                testCase.pluginType,
				EnableDefaultConfig: true,
			}

			testBuilder = testBuilder.WithIntelPlugin("test", plugin)
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				plugins := testBuilder.Definition.Spec.Profile[0].Plugins

				assert.NotNil(t, plugins)
				assert.NotNil(t, plugins[string(testCase.pluginType)])
			}
		})
	}
}

func TestGetPluginType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		ptpConfigValid bool
		pluginType     PluginType
		profileExists  bool
		expectedError  string
	}{
		{
			name:           "valid E810 plugin",
			ptpConfigValid: true,
			pluginType:     PluginTypeE810,
			profileExists:  true,
			expectedError:  "",
		},
		{
			name:           "valid E825 plugin",
			ptpConfigValid: true,
			pluginType:     PluginTypeE825,
			profileExists:  true,
			expectedError:  "",
		},
		{
			name:           "valid E830 plugin",
			ptpConfigValid: true,
			pluginType:     PluginTypeE830,
			profileExists:  true,
			expectedError:  "",
		},
		{
			name:           "invalid ptpConfig",
			ptpConfigValid: false,
			profileExists:  true,
			expectedError:  "ptpConfig 'nsname' cannot be empty",
		},
		{
			name:           "profile not found",
			ptpConfigValid: true,
			profileExists:  false,
			expectedError:  "ptpProfile test not found",
		},
		{
			name:           "no Intel plugin",
			ptpConfigValid: true,
			profileExists:  true,
			expectedError:  "ptpProfile test does not have an Intel plugin",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testSettings := buildTestClientWithPtpScheme()
			testBuilder := buildValidPtpConfigBuilder(testSettings)

			if !testCase.ptpConfigValid {
				testBuilder = buildInvalidPtpConfigBuilder(testSettings)
			}

			if testCase.profileExists {
				if testCase.pluginType != "" {
					testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{
						buildProfileWithPlugin("test", testCase.pluginType, []byte(`{}`)),
					}
				} else {
					testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{
						{Name: ptr.To("test"), Plugins: map[string]*apiextensionsv1.JSON{}},
					}
				}
			}

			result, err := testBuilder.GetPluginType("test")

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, testCase.pluginType, result)
		})
	}
}

func TestPullPtpConfig(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultPtpConfigName,
			nsname:              defaultPtpConfigNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultPtpConfigNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("ptpConfig 'name' cannot be empty"),
		},
		{
			name:                defaultPtpConfigName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("ptpConfig 'nsname' cannot be empty"),
		},
		{
			name:                defaultPtpConfigName,
			nsname:              defaultPtpConfigNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"ptpConfig object %s does not exist in namespace %s", defaultPtpConfigName, defaultPtpConfigNamespace),
		},
		{
			name:                defaultPtpConfigName,
			nsname:              defaultPtpConfigNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("ptpConfig 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPtpConfig := buildDummyPtpConfig(testCase.name, testCase.nsname)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPtpConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullPtpConfig(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPtpConfig.Name, testBuilder.Definition.Name)
			assert.Equal(t, testPtpConfig.Namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestPtpConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *PtpConfigBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: "ptpConfig 'nsname' cannot be empty",
		},
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			expectedError: "ptpconfigs.ptp.openshift.io \"test-ptp-config\" not found",
		},
	}

	for _, testCase := range testCases {
		ptpConfig, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, ptpConfig.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, ptpConfig.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestPtpConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PtpConfigBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			exists:      false,
		},
		{
			testBuilder: buildValidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPtpConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *PtpConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			expectedError: fmt.Errorf("ptpConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
			assert.Equal(t, testBuilder.Definition.Namespace, testBuilder.Object.Namespace)
		}
	}
}

func TestPtpConfigUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists bool
		expectedError error
	}{
		{
			alreadyExists: false,
			expectedError: fmt.Errorf("cannot update non-existent ptpConfig"),
		},
		{
			alreadyExists: true,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithPtpScheme()

		if testCase.alreadyExists {
			testSettings = buildTestClientWithDummyPtpConfig()
		}

		testBuilder := buildValidPtpConfigBuilder(testSettings)

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.Spec.Profile)

		testBuilder.Definition.Spec.Profile = []ptpv1.PtpProfile{{}}

		testBuilder, err := testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotEmpty(t, testBuilder.Object.Spec.Profile)
		}
	}
}

func TestPtpConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PtpConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidPtpConfigBuilder(buildTestClientWithPtpScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPtpConfigBuilder(buildTestClientWithDummyPtpConfig()),
			expectedError: fmt.Errorf("ptpConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestPtpConfigValidate(t *testing.T) {
	testCases := []struct {
		builderNil      bool
		definitionNil   bool
		apiClientNil    bool
		builderErrorMsg string
		expectedError   error
	}{
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   nil,
		},
		{
			builderNil:      true,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("error: received nil ptpConfig builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined ptpConfig"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("ptpConfig builder cannot have nil apiClient"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "test error",
			expectedError:   fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPtpConfigBuilder(buildTestClientWithPtpScheme())

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			testBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := testBuilder.validate()
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, valid)
	}
}

// buildDummyPtpConfig returns a PtpConfig with the provided name and namespace.
func buildDummyPtpConfig(name, namespace string) *ptpv1.PtpConfig {
	return &ptpv1.PtpConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// buildTestClientWithDummyPtpConfig returns a client with a mock PtpConfig.
func buildTestClientWithDummyPtpConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPtpConfig(defaultPtpConfigName, defaultPtpConfigNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildTestClientWithPtpScheme returns a client with no objects but the ptp v1 scheme attached.
func buildTestClientWithPtpScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

// buildValidPtpConfigBuilder returns a valid PtpConfigBuilder for testing.
func buildValidPtpConfigBuilder(apiClient *clients.Settings) *PtpConfigBuilder {
	return NewPtpConfigBuilder(apiClient, defaultPtpConfigName, defaultPtpConfigNamespace)
}

// buildInvalidPtpConfigBuilder returns an invalid PtpConfigBuilder for testing.
func buildInvalidPtpConfigBuilder(apiClient *clients.Settings) *PtpConfigBuilder {
	return NewPtpConfigBuilder(apiClient, defaultPtpConfigName, "")
}

// buildProfileWithPlugin returns a PtpProfile with the specified plugin type and raw JSON.
func buildProfileWithPlugin(name string, pluginType PluginType, raw []byte) ptpv1.PtpProfile {
	return ptpv1.PtpProfile{
		Name: ptr.To(name),
		Plugins: map[string]*apiextensionsv1.JSON{
			string(pluginType): {Raw: raw},
		},
	}
}
