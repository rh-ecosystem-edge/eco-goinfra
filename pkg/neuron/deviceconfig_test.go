package neuron

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/neuron/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		v1alpha1.AddToScheme,
	}
	defaultDeviceConfigName      = "neuron"
	defaultDeviceConfigNamespace = "ai-operator-on-aws"
	defaultDriversImage          = "public.ecr.aws/q5p6u7h8/neuron-openshift/neuron-kernel-module:2.24.7.0"
	defaultDriverVersion         = "2.24.11.0"
	defaultDevicePluginImage     = "public.ecr.aws/neuron/neuron-device-plugin:2.24.23.0"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("valid config with client", func(t *testing.T) {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
		testBuilder := NewBuilder(testSettings, defaultDeviceConfigName, defaultDeviceConfigNamespace,
			defaultDriversImage, defaultDriverVersion, defaultDevicePluginImage)

		assert.NotNil(t, testBuilder)
		assert.Equal(t, defaultDeviceConfigName, testBuilder.Definition.Name)
		assert.Equal(t, defaultDeviceConfigNamespace, testBuilder.Definition.Namespace)
		assert.Equal(t, defaultDriversImage, testBuilder.Definition.Spec.DriversImage)
		assert.Equal(t, defaultDriverVersion, testBuilder.Definition.Spec.DriverVersion)
		assert.Equal(t, defaultDevicePluginImage, testBuilder.Definition.Spec.DevicePluginImage)
	})

	t.Run("nil client", func(t *testing.T) {
		testBuilder := NewBuilder(nil, defaultDeviceConfigName, defaultDeviceConfigNamespace,
			defaultDriversImage, defaultDriverVersion, defaultDevicePluginImage)
		assert.Nil(t, testBuilder)
	})

	t.Run("empty name", func(t *testing.T) {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
		testBuilder := NewBuilder(testSettings, "", defaultDeviceConfigNamespace,
			defaultDriversImage, defaultDriverVersion, defaultDevicePluginImage)
		assert.Equal(t, "DeviceConfig 'name' cannot be empty", testBuilder.errorMsg)
	})

	t.Run("empty namespace", func(t *testing.T) {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
		testBuilder := NewBuilder(testSettings, defaultDeviceConfigName, "",
			defaultDriversImage, defaultDriverVersion, defaultDevicePluginImage)
		assert.Equal(t, "DeviceConfig 'namespace' cannot be empty", testBuilder.errorMsg)
	})

	t.Run("empty driversImage", func(t *testing.T) {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
		testBuilder := NewBuilder(testSettings, defaultDeviceConfigName, defaultDeviceConfigNamespace,
			"", defaultDriverVersion, defaultDevicePluginImage)
		assert.Equal(t, "DeviceConfig 'driversImage' cannot be empty", testBuilder.errorMsg)
	})

	t.Run("empty driverVersion", func(t *testing.T) {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
		testBuilder := NewBuilder(testSettings, defaultDeviceConfigName, defaultDeviceConfigNamespace,
			defaultDriversImage, "", defaultDevicePluginImage)
		assert.Equal(t, "DeviceConfig 'driverVersion' cannot be empty", testBuilder.errorMsg)
	})

	t.Run("empty devicePluginImage", func(t *testing.T) {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
		testBuilder := NewBuilder(testSettings, defaultDeviceConfigName, defaultDeviceConfigNamespace,
			defaultDriversImage, defaultDriverVersion, "")
		assert.Equal(t, "DeviceConfig 'devicePluginImage' cannot be empty", testBuilder.errorMsg)
	})
}

func TestPull(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		expectedError       error
		addToRuntimeObjects bool
		client              bool
	}{
		{
			name:                "neuron",
			namespace:           "ai-operator-on-aws",
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			namespace:           "ai-operator-on-aws",
			expectedError:       fmt.Errorf("deviceConfig 'name' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "neuron",
			namespace:           "",
			expectedError:       fmt.Errorf("deviceConfig 'namespace' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "neuron",
			namespace:           "ai-operator-on-aws",
			expectedError:       fmt.Errorf("deviceConfig object neuron does not exist in namespace ai-operator-on-aws"),
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "neuron",
			namespace:           "ai-operator-on-aws",
			expectedError:       fmt.Errorf("deviceConfig 'apiClient' cannot be nil"),
			addToRuntimeObjects: false,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testDeviceConfig := buildDummyDeviceConfig(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testDeviceConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := Pull(testSettings, testCase.name, testCase.namespace)

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

func TestDeviceConfigGet(t *testing.T) {
	testCases := []struct {
		testDeviceConfig *Builder
		expectedError    error
	}{
		{
			testDeviceConfig: buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError:    nil,
		},
		{
			testDeviceConfig: buildInvalidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError:    fmt.Errorf("DeviceConfig 'name' cannot be empty"),
		},
		{
			testDeviceConfig: buildValidDeviceConfigBuilder(clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: testSchemes,
			})),
			expectedError: fmt.Errorf("deviceconfigs.k8s.aws \"neuron\" not found"),
		},
	}

	for _, testCase := range testCases {
		deviceConfig, err := testCase.testDeviceConfig.Get()

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.NotNil(t, deviceConfig)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestDeviceConfigExists(t *testing.T) {
	testCases := []struct {
		testDeviceConfig *Builder
		expectedStatus   bool
	}{
		{
			testDeviceConfig: buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedStatus:   true,
		},
		{
			testDeviceConfig: buildInvalidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedStatus:   false,
		},
		{
			testDeviceConfig: buildValidDeviceConfigBuilder(clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: testSchemes,
			})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testDeviceConfig.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

func TestDeviceConfigCreate(t *testing.T) {
	testCases := []struct {
		testDeviceConfig *Builder
		expectedError    error
	}{
		{
			testDeviceConfig: buildValidDeviceConfigBuilder(clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: testSchemes,
			})),
			expectedError: nil,
		},
		{
			testDeviceConfig: buildInvalidDeviceConfigBuilder(clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: testSchemes,
			})),
			expectedError: fmt.Errorf("DeviceConfig 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		deviceConfigBuilder, err := testCase.testDeviceConfig.Create()

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.NotNil(t, deviceConfigBuilder)
			assert.Equal(t, deviceConfigBuilder.Definition, deviceConfigBuilder.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestDeviceConfigDelete(t *testing.T) {
	testCases := []struct {
		testDeviceConfig *Builder
		expectedError    error
	}{
		{
			testDeviceConfig: buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError:    nil,
		},
		{
			testDeviceConfig: buildInvalidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError:    fmt.Errorf("DeviceConfig 'name' cannot be empty"),
		},
		{
			testDeviceConfig: buildValidDeviceConfigBuilder(clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: testSchemes,
			})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testDeviceConfig.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.Nil(t, testCase.testDeviceConfig.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestDeviceConfigUpdate(t *testing.T) {
	testCases := []struct {
		testDeviceConfig *Builder
		expectedError    error
		driverVersion    string
	}{
		{
			testDeviceConfig: buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError:    nil,
			driverVersion:    "2.25.0.0",
		},
		{
			testDeviceConfig: buildInvalidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			expectedError:    fmt.Errorf("DeviceConfig 'name' cannot be empty"),
			driverVersion:    "2.25.0.0",
		},
	}

	for _, testCase := range testCases {
		assert.NotNil(t, testCase.testDeviceConfig.Definition)
		assert.Equal(t, defaultDriverVersion, testCase.testDeviceConfig.Definition.Spec.DriverVersion)

		testCase.testDeviceConfig.Definition.Spec.DriverVersion = testCase.driverVersion
		testCase.testDeviceConfig.Definition.ResourceVersion = "999"

		_, err := testCase.testDeviceConfig.Update(false)

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.Equal(t, testCase.driverVersion, testCase.testDeviceConfig.Definition.Spec.DriverVersion)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestDeviceConfigWithSelector(t *testing.T) {
	testCases := []struct {
		testSelector  map[string]string
		expectedError string
	}{
		{
			testSelector:  map[string]string{"feature.node.kubernetes.io/aws-neuron": "true"},
			expectedError: "",
		},
		{
			testSelector:  map[string]string{},
			expectedError: "DeviceConfig 'selector' cannot be empty map",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDeviceConfigBuilder(buildTestClientWithDummyObject())
		result := testBuilder.WithSelector(testCase.testSelector)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.testSelector, result.Definition.Spec.Selector)
		} else {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		}
	}
}

func TestDeviceConfigWithScheduler(t *testing.T) {
	testCases := []struct {
		schedulerImage string
		extensionImage string
		expectedError  string
	}{
		{
			schedulerImage: "public.ecr.aws/eks-distro/kubernetes/kube-scheduler:v1.32.9-eks-1-32-24",
			extensionImage: "public.ecr.aws/neuron/neuron-scheduler:2.24.23.0",
			expectedError:  "",
		},
		{
			schedulerImage: "",
			extensionImage: "public.ecr.aws/neuron/neuron-scheduler:2.24.23.0",
			expectedError:  "DeviceConfig 'schedulerImage' cannot be empty",
		},
		{
			schedulerImage: "public.ecr.aws/eks-distro/kubernetes/kube-scheduler:v1.32.9-eks-1-32-24",
			extensionImage: "",
			expectedError:  "DeviceConfig 'extensionImage' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDeviceConfigBuilder(buildTestClientWithDummyObject())
		result := testBuilder.WithScheduler(testCase.schedulerImage, testCase.extensionImage)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.schedulerImage, result.Definition.Spec.CustomSchedulerImage)
			assert.Equal(t, testCase.extensionImage, result.Definition.Spec.SchedulerExtensionImage)
		} else {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		}
	}
}

func TestDeviceConfigWithImageRepoSecret(t *testing.T) {
	testCases := []struct {
		secretName    string
		expectedError string
	}{
		{
			secretName:    "my-secret",
			expectedError: "",
		},
		{
			secretName:    "",
			expectedError: "DeviceConfig 'imageRepoSecret' name cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDeviceConfigBuilder(buildTestClientWithDummyObject())
		result := testBuilder.WithImageRepoSecret(testCase.secretName)

		if testCase.expectedError == "" {
			assert.NotNil(t, result.Definition.Spec.ImageRepoSecret)
			assert.Equal(t, testCase.secretName, result.Definition.Spec.ImageRepoSecret.Name)
		} else {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		}
	}
}

func TestDeviceConfigWithDriverVersion(t *testing.T) {
	testCases := []struct {
		driverVersion string
		expectedError string
	}{
		{
			driverVersion: "2.25.0.0",
			expectedError: "",
		},
		{
			driverVersion: "",
			expectedError: "DeviceConfig 'driverVersion' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDeviceConfigBuilder(buildTestClientWithDummyObject())
		result := testBuilder.WithDriverVersion(testCase.driverVersion)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.driverVersion, result.Definition.Spec.DriverVersion)
		} else {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		}
	}
}

func TestDeviceConfigWithNodeMetricsImage(t *testing.T) {
	testCases := []struct {
		nodeMetricsImage string
		expectedError    string
	}{
		{
			nodeMetricsImage: "public.ecr.aws/neuron/neuron-monitor:1.3.0",
			expectedError:    "",
		},
		{
			nodeMetricsImage: "",
			expectedError:    "DeviceConfig 'nodeMetricsImage' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDeviceConfigBuilder(buildTestClientWithDummyObject())
		result := testBuilder.WithNodeMetricsImage(testCase.nodeMetricsImage)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.nodeMetricsImage, result.Definition.Spec.NodeMetricsImage)
		} else {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		}
	}
}

func TestDeviceConfigWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		options       AdditionalOptions
		expectedError string
	}{
		{
			testBuilder: buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			options: func(builder *Builder) (*Builder, error) {
				builder.Definition.Spec.DriverVersion = "2.26.0.0"

				return builder, nil
			},
			expectedError: "",
		},
		{
			testBuilder: buildValidDeviceConfigBuilder(buildTestClientWithDummyObject()),
			options: func(builder *Builder) (*Builder, error) {
				return builder, fmt.Errorf("error in mutation function")
			},
			expectedError: "error in mutation function",
		},
	}

	for _, testCase := range testCases {
		result := testCase.testBuilder.WithOptions(testCase.options)

		if testCase.expectedError == "" {
			assert.Empty(t, result.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		}
	}
}

// buildDummyDeviceConfig returns a DeviceConfig with the provided name and namespace.
func buildDummyDeviceConfig(name, namespace string) *v1alpha1.DeviceConfig {
	return &v1alpha1.DeviceConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.DeviceConfigSpec{
			DriversImage:      defaultDriversImage,
			DriverVersion:     defaultDriverVersion,
			DevicePluginImage: defaultDevicePluginImage,
		},
	}
}

// buildTestClientWithDummyObject returns a client with a mock DeviceConfig object.
func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyDeviceConfig(defaultDeviceConfigName, defaultDeviceConfigNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildValidDeviceConfigBuilder returns a valid Builder for testing.
func buildValidDeviceConfigBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(
		apiClient,
		defaultDeviceConfigName,
		defaultDeviceConfigNamespace,
		defaultDriversImage,
		defaultDriverVersion,
		defaultDevicePluginImage,
	)
}

// buildInvalidDeviceConfigBuilder returns an invalid Builder for testing.
func buildInvalidDeviceConfigBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(
		apiClient,
		"",
		defaultDeviceConfigNamespace,
		defaultDriversImage,
		defaultDriverVersion,
		defaultDevicePluginImage,
	)
}
