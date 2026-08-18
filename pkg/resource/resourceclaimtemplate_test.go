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
	defaultRCTName      = "test-rct"
	defaultRCTNamespace = "test-ns"
)

func TestNewResourceClaimTemplateBuilder(t *testing.T) {
	testCases := []struct {
		name        string
		namespace   string
		expectedErr string
		client      bool
	}{
		{
			name:        defaultRCTName,
			namespace:   defaultRCTNamespace,
			expectedErr: "",
			client:      true,
		},
		{
			name:        "",
			namespace:   defaultRCTNamespace,
			expectedErr: errRCTNameEmpty,
			client:      true,
		},
		{
			name:        defaultRCTName,
			namespace:   "",
			expectedErr: "ResourceClaimTemplate 'namespace' cannot be empty",
			client:      true,
		},
		{
			name:        defaultRCTName,
			namespace:   defaultRCTNamespace,
			expectedErr: "",
			client:      false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: testResourceSchemes,
			})
		}

		testBuilder := NewResourceClaimTemplateBuilder(testSettings, testCase.name, testCase.namespace)

		switch {
		case !testCase.client:
			assert.Nil(t, testBuilder)
		case testCase.expectedErr == "":
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
			assert.Empty(t, testBuilder.errorMsg)
		default:
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestWithDeviceRequest(t *testing.T) {
	testCases := []struct {
		requestName     string
		deviceClassName string
		count           int64
		expectedErr     string
	}{
		{
			requestName:     "test-request",
			deviceClassName: "test-class",
			count:           1,
			expectedErr:     "",
		},
		{
			requestName:     "",
			deviceClassName: "test-class",
			count:           1,
			expectedErr:     "device request 'name' cannot be empty",
		},
		{
			requestName:     "test-request",
			deviceClassName: "",
			count:           1,
			expectedErr:     "device request 'deviceClassName' cannot be empty",
		},
		{
			requestName:     "test-request",
			deviceClassName: "test-class",
			count:           0,
			expectedErr:     "device request 'count' must be at least 1",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidResourceClaimTemplate()
		testBuilder.WithDeviceRequest(testCase.requestName, testCase.deviceClassName, testCase.count)

		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Len(t, testBuilder.Definition.Spec.Spec.Devices.Requests, 1)
			assert.Equal(t, testCase.requestName,
				testBuilder.Definition.Spec.Spec.Devices.Requests[0].Name)
		}
	}
}

func TestWithDeviceRequestSubrequestName(t *testing.T) {
	testBuilder := buildValidResourceClaimTemplate()
	testBuilder.WithDeviceRequest("my-request", "test-class", 1)

	assert.Empty(t, testBuilder.errorMsg)
	assert.Len(t, testBuilder.Definition.Spec.Spec.Devices.Requests, 1)

	request := testBuilder.Definition.Spec.Spec.Devices.Requests[0]
	assert.Len(t, request.FirstAvailable, 1)
	assert.Equal(t, "default", request.FirstAvailable[0].Name)
}

func TestWithDeviceRequestMultiple(t *testing.T) {
	testBuilder := buildValidResourceClaimTemplate()
	testBuilder.WithDeviceRequest("request-1", "class-1", 1)
	testBuilder.WithDeviceRequest("request-2", "class-2", 2)

	assert.Empty(t, testBuilder.errorMsg)
	assert.Len(t, testBuilder.Definition.Spec.Spec.Devices.Requests, 2)
}

func TestWithDeviceRequestInvalidBuilder(t *testing.T) {
	testBuilder := buildInvalidResourceClaimTemplate()
	origErr := testBuilder.errorMsg

	testBuilder.WithDeviceRequest("test", "test-class", 1)
	assert.Equal(t, origErr, testBuilder.errorMsg)
}

func TestPullResourceClaimTemplate(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		expectedError       error
		addToRuntimeObjects bool
		client              bool
	}{
		{
			name:                defaultRCTName,
			namespace:           defaultRCTNamespace,
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultRCTNamespace,
			expectedError:       fmt.Errorf("resourceClaimTemplate 'name' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                defaultRCTName,
			namespace:           "",
			expectedError:       fmt.Errorf("resourceClaimTemplate 'namespace' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:      defaultRCTName,
			namespace: defaultRCTNamespace,
			expectedError: fmt.Errorf(
				"resourceClaimTemplate object %s does not exist in namespace %s",
				defaultRCTName, defaultRCTNamespace),
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                defaultRCTName,
			namespace:           defaultRCTNamespace,
			expectedError:       fmt.Errorf("resourceClaimTemplate 'apiClient' cannot be nil"),
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
			runtimeObjects = append(runtimeObjects,
				generateResourceClaimTemplate(testCase.name, testCase.namespace))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testResourceSchemes,
			})
		}

		builderResult, err := PullResourceClaimTemplate(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
			assert.Equal(t, testCase.name, builderResult.Definition.Name)
		}
	}
}

func TestResourceClaimTemplateCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ResourceClaimTemplateBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidResourceClaimTemplate(),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidResourceClaimTemplate(),
			expectedError: errRCTNameEmpty,
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

func TestResourceClaimTemplateDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *ResourceClaimTemplateBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidResourceClaimTemplateWithDummyObject(),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidResourceClaimTemplate(),
			expectedError: fmt.Errorf(errRCTNameEmpty),
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

func TestResourceClaimTemplateExists(t *testing.T) {
	testCases := []struct {
		testBuilder    *ResourceClaimTemplateBuilder
		expectedStatus bool
	}{
		{
			testBuilder:    buildValidResourceClaimTemplateWithDummyObject(),
			expectedStatus: true,
		},
		{
			testBuilder:    buildInvalidResourceClaimTemplate(),
			expectedStatus: false,
		},
		{
			testBuilder:    buildValidResourceClaimTemplate(),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func buildValidResourceClaimTemplate() *ResourceClaimTemplateBuilder {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testResourceSchemes,
	})

	return NewResourceClaimTemplateBuilder(testSettings, defaultRCTName, defaultRCTNamespace)
}

func buildValidResourceClaimTemplateWithDummyObject() *ResourceClaimTemplateBuilder {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			generateResourceClaimTemplate(defaultRCTName, defaultRCTNamespace),
		},
		SchemeAttachers: testResourceSchemes,
	})

	return NewResourceClaimTemplateBuilder(testSettings, defaultRCTName, defaultRCTNamespace)
}

func buildInvalidResourceClaimTemplate() *ResourceClaimTemplateBuilder {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testResourceSchemes,
	})

	return NewResourceClaimTemplateBuilder(testSettings, "", defaultRCTNamespace)
}

func generateResourceClaimTemplate(name, namespace string) *resourcev1.ResourceClaimTemplate {
	return &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
