package kmm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemesBootModuleConfig = []clients.SchemeAttacher{
		v1beta1.AddToScheme,
	}
	defaultBootModuleConfigName      = "testbootmoduleconfig"
	defaultBootModuleConfigNamespace = "testns"
)

func TestNewBootModuleConfigBuilder(t *testing.T) {
	testCases := []struct {
		name        string
		namespace   string
		expectedErr string
		client      bool
	}{
		{
			name:        defaultBootModuleConfigName,
			namespace:   defaultBootModuleConfigNamespace,
			expectedErr: "",
			client:      true,
		},
		{
			name:        defaultBootModuleConfigName,
			namespace:   defaultBootModuleConfigNamespace,
			expectedErr: "",
			client:      false,
		},
		{
			name:        defaultBootModuleConfigName,
			namespace:   "",
			expectedErr: "bootmoduleconfig 'namespace' cannot be empty",
			client:      true,
		},
		{
			name:        "",
			namespace:   defaultBootModuleConfigNamespace,
			expectedErr: "bootmoduleconfig 'name' cannot be empty",
			client:      true,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemesBootModuleConfig})
		}

		testBuilder := NewBootModuleConfigBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedErr == "" {
			if testCase.client {
				assert.NotNil(t, testBuilder)
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestBootModuleConfigPull(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		expectedError       error
		addToRuntimeObjects bool
		client              bool
	}{
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			namespace:           "testns",
			expectedError:       fmt.Errorf("bootmoduleconfig 'name' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			expectedError:       fmt.Errorf("bootmoduleconfig 'namespace' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("bootmoduleconfig object test does not exist in namespace testns"),
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("bootmoduleconfig 'apiClient' cannot be empty"),
			addToRuntimeObjects: false,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, generateBootModuleConfig(testCase.name, testCase.namespace))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemesBootModuleConfig,
			})
		}

		testBuilder, err := PullBootModuleConfig(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestBootModuleConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *BootModuleConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: fmt.Errorf("bootmoduleconfig 'namespace' cannot be empty"),
		},
		{
			testBuilder:   buildValidBootModuleConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("bootmoduleconfigs.kmm.sigs.x-k8s.io \"testbootmoduleconfig\" not found"),
		},
	}

	for _, testCase := range testCases {
		bootmoduleconfig, err := testCase.testBuilder.Get()

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.NotNil(t, bootmoduleconfig)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestBootModuleConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder    *BootModuleConfigBuilder
		expectedStatus bool
	}{
		{
			testBuilder:    buildValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testBuilder:    buildInValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testBuilder:    buildValidBootModuleConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

func TestBootModuleConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *BootModuleConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: fmt.Errorf("bootmoduleconfig 'namespace' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		bootmoduleconfig, err := testCase.testBuilder.Create()

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.NotNil(t, bootmoduleconfig)
			assert.NotNil(t, bootmoduleconfig.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestBootModuleConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *BootModuleConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: fmt.Errorf("bootmoduleconfig 'namespace' cannot be empty"),
		},
		{
			testBuilder:   buildValidBootModuleConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		bootmoduleconfig, err := testCase.testBuilder.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.NotNil(t, bootmoduleconfig)
			assert.Nil(t, bootmoduleconfig.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestBootModuleConfigUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *BootModuleConfigBuilder
		expectedError string
		newKernelName string
	}{
		{
			testBuilder:   buildValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: "",
			newKernelName: "newkernelmodule",
		},
		{
			testBuilder:   buildInValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: "bootmoduleconfig 'namespace' cannot be empty",
			newKernelName: "newkernelmodule",
		},
	}

	for _, testCase := range testCases {
		// For valid test case, fetch the existing object first
		if testCase.expectedError == "" {
			obj, err := testCase.testBuilder.Get()
			assert.Nil(t, err)
			assert.NotNil(t, obj)
			testCase.testBuilder.Definition = obj

			assert.Equal(t, "test-module", testCase.testBuilder.Definition.Spec.KernelModuleName)
			testCase.testBuilder.Definition.Spec.KernelModuleName = testCase.newKernelName
		} else {
			testCase.testBuilder.Definition.Spec.KernelModuleName = testCase.newKernelName
		}

		bootmoduleconfig, err := testCase.testBuilder.Update()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, bootmoduleconfig)
			assert.Equal(t, testCase.newKernelName, bootmoduleconfig.Definition.Spec.KernelModuleName)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedError)
		}
	}
}

func TestBootModuleConfigWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *BootModuleConfigBuilder
		expectedError error
		options       BootModuleConfigAdditionalOptions
	}{
		{
			testBuilder:   buildValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: nil,
			options: func(builder *BootModuleConfigBuilder) (*BootModuleConfigBuilder, error) {
				builder.Definition.Spec.WorkerImage = "test-worker-image"

				return builder, nil
			},
		},
		{
			testBuilder:   buildValidBootModuleConfigBuilder(buildBootModuleConfigTestClientWithDummyObject()),
			expectedError: fmt.Errorf("error adding additional option"),
			options: func(builder *BootModuleConfigBuilder) (*BootModuleConfigBuilder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithOptions(testCase.options)

		if testCase.expectedError == nil {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, "", testBuilder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), testBuilder.errorMsg)
		}
	}
}

// buildValidBootModuleConfigBuilder returns a valid BootModuleConfigBuilder for testing.
func buildValidBootModuleConfigBuilder(apiClient *clients.Settings) *BootModuleConfigBuilder {
	return NewBootModuleConfigBuilder(
		apiClient,
		defaultBootModuleConfigName,
		defaultBootModuleConfigNamespace,
	)
}

// buildInValidBootModuleConfigBuilder returns an invalid BootModuleConfigBuilder for testing.
func buildInValidBootModuleConfigBuilder(apiClient *clients.Settings) *BootModuleConfigBuilder {
	return NewBootModuleConfigBuilder(
		apiClient,
		defaultBootModuleConfigName,
		"",
	)
}

// buildBootModuleConfigTestClientWithDummyObject returns a test client with a dummy BootModuleConfig object.
func buildBootModuleConfigTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyBootModuleConfig(),
		SchemeAttachers: testSchemesBootModuleConfig,
	})
}

// buildDummyBootModuleConfig returns a slice of runtime.Object with a dummy BootModuleConfig.
func buildDummyBootModuleConfig() []runtime.Object {
	return []runtime.Object{
		generateBootModuleConfig(defaultBootModuleConfigName, defaultBootModuleConfigNamespace),
	}
}

// generateBootModuleConfig returns a BootModuleConfig object with the given name and namespace.
func generateBootModuleConfig(name, nsname string) *v1beta1.BootModuleConfig {
	return &v1beta1.BootModuleConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		Spec: v1beta1.BootModuleConfigSpec{
			MachineConfigName:     "test-machine-config",
			MachineConfigPoolName: "test-pool",
			KernelModuleImage:     "test-image:latest",
			KernelModuleName:      "test-module",
		},
	}
}
