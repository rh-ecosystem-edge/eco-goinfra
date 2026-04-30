package kserve

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	kservev1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kserve/v1alpha1"
	kservev1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kserve/v1beta1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultSRName      = "test-runtime"
	defaultSRNamespace = "test-ns"
)

var srTestSchemes = []clients.SchemeAttacher{
	kservev1alpha1.AddToScheme,
	kservev1beta1.AddToScheme,
}

func TestNewServingRuntimeBuilder(t *testing.T) {
	testCases := []struct {
		name             string
		namespace        string
		expectedErrorMsg string
	}{
		{
			name:             defaultSRName,
			namespace:        defaultSRNamespace,
			expectedErrorMsg: "",
		},
		{
			name:             "",
			namespace:        defaultSRNamespace,
			expectedErrorMsg: "ServingRuntime 'name' cannot be empty",
		},
		{
			name:             defaultSRName,
			namespace:        "",
			expectedErrorMsg: "ServingRuntime 'namespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: srTestSchemes,
		})
		builder := NewServingRuntimeBuilder(testSettings, testCase.name, testCase.namespace)
		assert.NotNil(t, builder)
		assert.NotNil(t, builder.Definition)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.name, builder.Definition.Name)
			assert.Equal(t, testCase.namespace, builder.Definition.Namespace)
			assert.NotNil(t, builder.Definition.Spec.MultiModel)
			assert.False(t, *builder.Definition.Spec.MultiModel)
			assert.Equal(t, "", builder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedErrorMsg, builder.errorMsg)
		}
	}
}

func TestNewServingRuntimeBuilderNilClient(t *testing.T) {
	builder := NewServingRuntimeBuilder(nil, defaultSRName, defaultSRNamespace)
	assert.Nil(t, builder)
}

func TestServingRuntimeWithModelFormat(t *testing.T) {
	testCases := []struct {
		name             string
		autoSelect       bool
		expectedErrorMsg string
	}{
		{
			name:             "vllm-neuron",
			autoSelect:       true,
			expectedErrorMsg: "",
		},
		{
			name:             "",
			autoSelect:       false,
			expectedErrorMsg: "ServingRuntime model format 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		builder := buildValidSRBuilder(buildSRTestClient())
		builder.WithModelFormat(testCase.name, testCase.autoSelect)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, 1, len(builder.Definition.Spec.SupportedModelFormats))
			assert.Equal(t, testCase.name, builder.Definition.Spec.SupportedModelFormats[0].Name)
			assert.Equal(t, testCase.autoSelect, builder.Definition.Spec.SupportedModelFormats[0].AutoSelect)
			assert.Equal(t, "", builder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedErrorMsg, builder.errorMsg)
		}
	}
}

func TestServingRuntimeWithContainer(t *testing.T) {
	builder := buildValidSRBuilder(buildSRTestClient())

	container := corev1.Container{
		Name:  "vllm",
		Image: "registry.redhat.io/rhaiis/vllm-neuron-rhel9:3",
		Ports: []corev1.ContainerPort{
			{ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
		},
	}

	builder.WithContainer(container)

	assert.Equal(t, 1, len(builder.Definition.Spec.Containers))
	assert.Equal(t, "vllm", builder.Definition.Spec.Containers[0].Name)
	assert.Equal(t, "", builder.errorMsg)
}

func TestServingRuntimeWithVolume(t *testing.T) {
	builder := buildValidSRBuilder(buildSRTestClient())

	volume := corev1.Volume{
		Name: "shm",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumMemory,
			},
		},
	}

	builder.WithVolume(volume)

	assert.Equal(t, 1, len(builder.Definition.Spec.Volumes))
	assert.Equal(t, "shm", builder.Definition.Spec.Volumes[0].Name)
	assert.Equal(t, "", builder.errorMsg)
}

func TestServingRuntimeWithAnnotation(t *testing.T) {
	builder := buildValidSRBuilder(buildSRTestClient())
	builder.WithAnnotation("serving.kserve.io/autoscalerClass", "hpa")

	assert.Equal(t, "hpa", builder.Definition.Annotations["serving.kserve.io/autoscalerClass"])
	assert.Equal(t, "", builder.errorMsg)
}

func TestPullServingRuntime(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                defaultSRName,
			namespace:           defaultSRNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultSRNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("servingRuntime 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultSRName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("servingRuntime 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "nonexistent",
			namespace:           defaultSRNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf(
				"servingRuntime nonexistent does not exist in namespace %s", defaultSRNamespace),
			client: true,
		},
		{
			name:                defaultSRName,
			namespace:           defaultSRNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("servingRuntime 'apiClient' cannot be nil"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testSR := buildDummyServingRuntime(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testSR)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: srTestSchemes,
			})
		}

		builderResult, err := PullServingRuntime(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError != nil {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testSR.Name, builderResult.Object.Name)
			assert.Equal(t, testSR.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestServingRuntimeExists(t *testing.T) {
	testCases := []struct {
		testSR         *ServingRuntimeBuilder
		expectedStatus bool
	}{
		{
			testSR:         buildValidSRBuilder(buildSRClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testSR:         buildInvalidSRBuilder(buildSRClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testSR:         buildValidSRBuilder(buildSRTestClient()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testSR.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestServingRuntimeCreate(t *testing.T) {
	testCases := []struct {
		testSR      *ServingRuntimeBuilder
		expectedErr bool
	}{
		{
			testSR:      buildValidSRBuilder(buildSRTestClient()),
			expectedErr: false,
		},
		{
			testSR:      buildInvalidSRBuilder(buildSRTestClient()),
			expectedErr: true,
		},
	}

	for _, testCase := range testCases {
		result, err := testCase.testSR.Create()

		if testCase.expectedErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, result.Object)
			assert.Equal(t, defaultSRName, result.Object.Name)
		}
	}
}

func TestServingRuntimeDelete(t *testing.T) {
	testCases := []struct {
		testSR      *ServingRuntimeBuilder
		expectedErr bool
	}{
		{
			testSR:      buildValidSRBuilder(buildSRClientWithDummyObject()),
			expectedErr: false,
		},
		{
			testSR:      buildValidSRBuilder(buildSRTestClient()),
			expectedErr: false,
		},
		{
			testSR:      buildInvalidSRBuilder(buildSRTestClient()),
			expectedErr: true,
		},
	}

	for _, testCase := range testCases {
		result, err := testCase.testSR.Delete()

		if testCase.expectedErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Nil(t, result.Object)
		}
	}
}

func TestServingRuntimeValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		errorMsg      string
		expectedValid bool
	}{
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			errorMsg:      "",
			expectedValid: true,
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			errorMsg:      "",
			expectedValid: false,
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			errorMsg:      "",
			expectedValid: false,
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			errorMsg:      "test error",
			expectedValid: false,
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: srTestSchemes,
		})

		builder := &ServingRuntimeBuilder{
			apiClient:  testSettings,
			Definition: &kservev1alpha1.ServingRuntime{},
			errorMsg:   testCase.errorMsg,
		}

		if testCase.definitionNil {
			builder.Definition = nil
		}

		if testCase.apiClientNil {
			builder.apiClient = nil
		}

		valid, err := builder.validate()
		assert.Equal(t, testCase.expectedValid, valid)

		if !testCase.expectedValid {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func buildValidSRBuilder(apiClient *clients.Settings) *ServingRuntimeBuilder {
	return NewServingRuntimeBuilder(apiClient, defaultSRName, defaultSRNamespace)
}

func buildInvalidSRBuilder(apiClient *clients.Settings) *ServingRuntimeBuilder {
	return NewServingRuntimeBuilder(apiClient, "", defaultSRNamespace)
}

func buildSRTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: srTestSchemes,
	})
}

func buildSRClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummySRObjects(),
		SchemeAttachers: srTestSchemes,
	})
}

func buildDummySRObjects() []runtime.Object {
	return []runtime.Object{
		buildDummyServingRuntime(defaultSRName, defaultSRNamespace),
	}
}

func buildDummyServingRuntime(name, namespace string) *kservev1alpha1.ServingRuntime {
	return &kservev1alpha1.ServingRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
