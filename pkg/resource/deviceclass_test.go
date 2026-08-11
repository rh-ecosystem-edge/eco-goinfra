package resource

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testResourceSchemes = []clients.SchemeAttacher{
		resourcev1.AddToScheme,
	}
	defaultDeviceClassName = "test-device-class"
)

func TestNewDeviceClassBuilder(t *testing.T) {
	testCases := []struct {
		name        string
		expectedErr string
		client      bool
	}{
		{
			name:        defaultDeviceClassName,
			expectedErr: "",
			client:      true,
		},
		{
			name:        "",
			expectedErr: "deviceClass 'name' cannot be empty",
			client:      true,
		},
		{
			name:        defaultDeviceClassName,
			expectedErr: "",
			client:      false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testResourceSchemes})
		}

		testBuilder := NewDeviceClassBuilder(testSettings, testCase.name)

		switch {
		case !testCase.client:
			assert.Nil(t, testBuilder)
		case testCase.expectedErr == "":
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Empty(t, testBuilder.errorMsg)
		default:
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestDeviceClassWithSelector(t *testing.T) {
	testCases := []struct {
		selector    resourcev1.DeviceSelector
		expectedErr string
	}{
		{
			selector: resourcev1.DeviceSelector{
				CEL: &resourcev1.CELDeviceSelector{Expression: "device.driver == 'test'"},
			},
			expectedErr: "",
		},
		{
			selector:    resourcev1.DeviceSelector{},
			expectedErr: "selector must have a CEL expression",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDeviceClass()
		testBuilder.WithSelector(testCase.selector)

		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Len(t, testBuilder.Definition.Spec.Selectors, 1)
			assert.Equal(t, testCase.selector, testBuilder.Definition.Spec.Selectors[0])
		} else {
			assert.Empty(t, testBuilder.Definition.Spec.Selectors)
		}
	}
}

func TestDeviceClassWithSelectorMultiple(t *testing.T) {
	testBuilder := buildValidDeviceClass()
	testBuilder.WithSelector(resourcev1.DeviceSelector{
		CEL: &resourcev1.CELDeviceSelector{Expression: "device.driver == 'test'"},
	})
	testBuilder.WithSelector(resourcev1.DeviceSelector{
		CEL: &resourcev1.CELDeviceSelector{Expression: "device.attr == 'gpu'"},
	})

	assert.Empty(t, testBuilder.errorMsg)
	assert.Len(t, testBuilder.Definition.Spec.Selectors, 2)
}

func TestDeviceClassWithCELSelector(t *testing.T) {
	testCases := []struct {
		expression  string
		expectedErr string
	}{
		{
			expression:  "device.driver == 'test'",
			expectedErr: "",
		},
		{
			expression:  "",
			expectedErr: "CEL 'expression' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDeviceClass()
		testBuilder.WithCELSelector(testCase.expression)

		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Len(t, testBuilder.Definition.Spec.Selectors, 1)
			assert.NotNil(t, testBuilder.Definition.Spec.Selectors[0].CEL)
			assert.Equal(t, testCase.expression, testBuilder.Definition.Spec.Selectors[0].CEL.Expression)
		}
	}
}

func TestDeviceClassWithConfig(t *testing.T) {
	testCases := []struct {
		config      resourcev1.DeviceClassConfiguration
		expectedErr string
	}{
		{
			config: resourcev1.DeviceClassConfiguration{
				DeviceConfiguration: resourcev1.DeviceConfiguration{
					Opaque: &resourcev1.OpaqueDeviceConfiguration{
						Driver:     "test-driver",
						Parameters: runtime.RawExtension{Raw: []byte(`{"key":"value"}`)},
					},
				},
			},
			expectedErr: "",
		},
		{
			config:      resourcev1.DeviceClassConfiguration{},
			expectedErr: "config must have an Opaque device configuration",
		},
		{
			config: resourcev1.DeviceClassConfiguration{
				DeviceConfiguration: resourcev1.DeviceConfiguration{
					Opaque: &resourcev1.OpaqueDeviceConfiguration{
						Driver: "",
					},
				},
			},
			expectedErr: "config Opaque 'Driver' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDeviceClass()
		testBuilder.WithConfig(testCase.config)

		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Len(t, testBuilder.Definition.Spec.Config, 1)
			assert.Equal(t, testCase.config, testBuilder.Definition.Spec.Config[0])
		} else {
			assert.Empty(t, testBuilder.Definition.Spec.Config)
		}
	}
}

func TestDeviceClassWithLabel(t *testing.T) {
	testCases := []struct {
		key         string
		value       string
		expectedErr string
	}{
		{
			key:         "kmm.node.kubernetes.io/module.name",
			value:       "test-module",
			expectedErr: "",
		},
		{
			key:         "",
			value:       "test",
			expectedErr: "label 'key' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDeviceClass()
		testBuilder.WithLabel(testCase.key, testCase.value)

		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.value, testBuilder.Definition.Labels[testCase.key])
		}
	}
}

func TestDeviceClassWithLabelMultiple(t *testing.T) {
	testBuilder := buildValidDeviceClass()
	testBuilder.WithLabel("key1", "value1").WithLabel("key2", "value2")

	assert.Empty(t, testBuilder.errorMsg)
	assert.Len(t, testBuilder.Definition.Labels, 2)
	assert.Equal(t, "value1", testBuilder.Definition.Labels["key1"])
	assert.Equal(t, "value2", testBuilder.Definition.Labels["key2"])
}

func TestPullDeviceClass(t *testing.T) {
	testCases := []struct {
		name                string
		expectedError       error
		addToRuntimeObjects bool
		client              bool
	}{
		{
			name:                defaultDeviceClassName,
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			expectedError:       fmt.Errorf("deviceClass 'name' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                defaultDeviceClassName,
			expectedError:       fmt.Errorf("deviceClass object %s does not exist", defaultDeviceClassName),
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                defaultDeviceClassName,
			expectedError:       fmt.Errorf("deviceClass 'apiClient' cannot be nil"),
			addToRuntimeObjects: true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, generateDeviceClass(testCase.name))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testResourceSchemes,
			})
		}

		builderResult, err := PullDeviceClass(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
			assert.Equal(t, testCase.name, builderResult.Definition.Name)
		}
	}
}

func TestDeviceClassCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *DeviceClassBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidDeviceClassWithClient(),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidDeviceClass(),
			expectedError: "deviceClass 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		result, err := testCase.testBuilder.Create()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, result.Object)
			assert.Equal(t, result.Definition.Name, result.Object.Name)
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestDeviceClassDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *DeviceClassBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidDeviceClassWithDummyObject(),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidDeviceClass(),
			expectedError: fmt.Errorf("deviceClass 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestDeviceClassExists(t *testing.T) {
	testCases := []struct {
		testBuilder    *DeviceClassBuilder
		expectedStatus bool
	}{
		{
			testBuilder:    buildValidDeviceClassWithDummyObject(),
			expectedStatus: true,
		},
		{
			testBuilder:    buildInvalidDeviceClass(),
			expectedStatus: false,
		},
		{
			testBuilder:    buildValidDeviceClassWithClient(),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestListDeviceClasses(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
		expectedCount       int
		client              bool
		expectedError       error
	}{
		{
			addToRuntimeObjects: true,
			expectedCount:       1,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: false,
			expectedCount:       0,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: true,
			expectedCount:       0,
			client:              false,
			expectedError:       fmt.Errorf("deviceClass 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, generateDeviceClass(defaultDeviceClassName))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testResourceSchemes,
			})
		}

		builders, err := ListDeviceClasses(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Len(t, builders, testCase.expectedCount)
		}
	}
}

func TestDeviceClassBuilderChaining(t *testing.T) {
	testBuilder := buildValidDeviceClass()
	testBuilder.
		WithCELSelector("device.driver == 'test'").
		WithConfig(resourcev1.DeviceClassConfiguration{
			DeviceConfiguration: resourcev1.DeviceConfiguration{
				Opaque: &resourcev1.OpaqueDeviceConfiguration{
					Driver:     "test-driver",
					Parameters: runtime.RawExtension{Raw: []byte(`{}`)},
				},
			},
		}).
		WithLabel("key", "value")

	assert.Empty(t, testBuilder.errorMsg)
	assert.Len(t, testBuilder.Definition.Spec.Selectors, 1)
	assert.Len(t, testBuilder.Definition.Spec.Config, 1)
	assert.Equal(t, "value", testBuilder.Definition.Labels["key"])
}

func TestDeviceClassBuilderWithInvalidBuilder(t *testing.T) {
	testBuilder := buildInvalidDeviceClass()
	origErr := testBuilder.errorMsg

	testBuilder.WithSelector(resourcev1.DeviceSelector{
		CEL: &resourcev1.CELDeviceSelector{Expression: "test"},
	})
	assert.Equal(t, origErr, testBuilder.errorMsg)

	testBuilder.WithCELSelector("test")
	assert.Equal(t, origErr, testBuilder.errorMsg)

	testBuilder.WithConfig(resourcev1.DeviceClassConfiguration{
		DeviceConfiguration: resourcev1.DeviceConfiguration{
			Opaque: &resourcev1.OpaqueDeviceConfiguration{Driver: "d"},
		},
	})
	assert.Equal(t, origErr, testBuilder.errorMsg)

	testBuilder.WithLabel("key", "value")
	assert.Equal(t, origErr, testBuilder.errorMsg)
}

func buildValidDeviceClass() *DeviceClassBuilder {
	testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testResourceSchemes})

	return NewDeviceClassBuilder(testSettings, defaultDeviceClassName)
}

func buildValidDeviceClassWithClient() *DeviceClassBuilder {
	testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testResourceSchemes})

	return NewDeviceClassBuilder(testSettings, defaultDeviceClassName)
}

func buildValidDeviceClassWithDummyObject() *DeviceClassBuilder {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{generateDeviceClass(defaultDeviceClassName)},
		SchemeAttachers: testResourceSchemes,
	})

	return NewDeviceClassBuilder(testSettings, defaultDeviceClassName)
}

func buildInvalidDeviceClass() *DeviceClassBuilder {
	testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testResourceSchemes})

	return NewDeviceClassBuilder(testSettings, "")
}

func generateDeviceClass(name string) *resourcev1.DeviceClass {
	return &resourcev1.DeviceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
