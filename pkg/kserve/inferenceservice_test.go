package kserve

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	kservev1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kserve/v1beta1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultISvcName      = "test-isvc"
	defaultISvcNamespace = "test-ns"
)

var isvcTestSchemes = []clients.SchemeAttacher{
	kservev1beta1.AddToScheme,
}

func TestNewInferenceServiceBuilder(t *testing.T) {
	testCases := []struct {
		name             string
		namespace        string
		expectedErrorMsg string
	}{
		{
			name:             defaultISvcName,
			namespace:        defaultISvcNamespace,
			expectedErrorMsg: "",
		},
		{
			name:             "",
			namespace:        defaultISvcNamespace,
			expectedErrorMsg: "InferenceService 'name' cannot be empty",
		},
		{
			name:             defaultISvcName,
			namespace:        "",
			expectedErrorMsg: "InferenceService 'namespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: isvcTestSchemes,
		})
		builder := NewInferenceServiceBuilder(testSettings, testCase.name, testCase.namespace)
		assert.NotNil(t, builder)
		assert.NotNil(t, builder.Definition)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.name, builder.Definition.Name)
			assert.Equal(t, testCase.namespace, builder.Definition.Namespace)
			assert.Equal(t, "", builder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedErrorMsg, builder.errorMsg)
		}
	}
}

func TestNewInferenceServiceBuilderNilClient(t *testing.T) {
	builder := NewInferenceServiceBuilder(nil, defaultISvcName, defaultISvcNamespace)
	assert.Nil(t, builder)
}

func TestInferenceServiceWithModelFormat(t *testing.T) {
	testCases := []struct {
		modelFormat      string
		expectedErrorMsg string
	}{
		{
			modelFormat:      "vllm-neuron",
			expectedErrorMsg: "",
		},
		{
			modelFormat:      "",
			expectedErrorMsg: "InferenceService 'modelFormat' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		builder := buildValidISvcBuilder(buildISvcTestClient())
		builder.WithModelFormat(testCase.modelFormat)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.modelFormat,
				builder.Definition.Spec.Predictor.Model.ModelFormat.Name)
			assert.Equal(t, "", builder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedErrorMsg, builder.errorMsg)
		}
	}
}

func TestInferenceServiceWithRuntime(t *testing.T) {
	testCases := []struct {
		runtime          string
		expectedErrorMsg string
	}{
		{
			runtime:          "vllm-neuron-runtime",
			expectedErrorMsg: "",
		},
		{
			runtime:          "",
			expectedErrorMsg: "InferenceService 'runtime' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		builder := buildValidISvcBuilder(buildISvcTestClient())
		builder.WithRuntime(testCase.runtime)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.runtime, *builder.Definition.Spec.Predictor.Model.Runtime)
			assert.Equal(t, "", builder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedErrorMsg, builder.errorMsg)
		}
	}
}

func TestInferenceServiceWithStorageURI(t *testing.T) {
	testCases := []struct {
		uri              string
		expectedErrorMsg string
	}{
		{
			uri:              "s3://bucket/model",
			expectedErrorMsg: "",
		},
		{
			uri:              "",
			expectedErrorMsg: "InferenceService 'storageUri' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		builder := buildValidISvcBuilder(buildISvcTestClient())
		builder.WithStorageURI(testCase.uri)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.uri, *builder.Definition.Spec.Predictor.Model.StorageURI)
			assert.Equal(t, "", builder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedErrorMsg, builder.errorMsg)
		}
	}
}

func TestInferenceServiceWithServiceAccountName(t *testing.T) {
	builder := buildValidISvcBuilder(buildISvcTestClient())
	builder.WithServiceAccountName("test-sa")

	assert.Equal(t, "test-sa", builder.Definition.Spec.Predictor.ServiceAccountName)
	assert.Equal(t, "", builder.errorMsg)
}

func TestInferenceServiceWithResources(t *testing.T) {
	builder := buildValidISvcBuilder(buildISvcTestClient())

	requests := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("2Gi"),
	}
	limits := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("2"),
		corev1.ResourceMemory: resource.MustParse("4Gi"),
	}

	builder.WithResources(requests, limits)

	assert.Equal(t, requests, builder.Definition.Spec.Predictor.Model.Resources.Requests)
	assert.Equal(t, limits, builder.Definition.Spec.Predictor.Model.Resources.Limits)
	assert.Equal(t, "", builder.errorMsg)
}

func TestInferenceServiceWithNeuronResources(t *testing.T) {
	testCases := []struct {
		description      string
		devices          int64
		memoryRequest    string
		memoryLimit      string
		expectedErrorMsg string
	}{
		{
			description:      "valid neuron resources",
			devices:          2,
			memoryRequest:    "8Gi",
			memoryLimit:      "16Gi",
			expectedErrorMsg: "",
		},
		{
			description:      "invalid memory request",
			devices:          1,
			memoryRequest:    "not-a-quantity",
			memoryLimit:      "8Gi",
			expectedErrorMsg: "failed to parse memory request",
		},
		{
			description:      "invalid memory limit",
			devices:          1,
			memoryRequest:    "8Gi",
			memoryLimit:      "not-a-quantity",
			expectedErrorMsg: "failed to parse memory limit",
		},
	}

	for _, testCase := range testCases {
		builder := buildValidISvcBuilder(buildISvcTestClient())
		builder.WithNeuronResources(testCase.devices, testCase.memoryRequest, testCase.memoryLimit)

		if testCase.expectedErrorMsg == "" {
			neuronQuantity := resource.MustParse(fmt.Sprintf("%d", testCase.devices))
			assert.Equal(t, neuronQuantity,
				builder.Definition.Spec.Predictor.Model.Resources.Requests["aws.amazon.com/neuron"])
			assert.Equal(t, neuronQuantity,
				builder.Definition.Spec.Predictor.Model.Resources.Limits["aws.amazon.com/neuron"])
			assert.Equal(t, "", builder.errorMsg)
		} else {
			assert.Contains(t, builder.errorMsg, testCase.expectedErrorMsg)
		}
	}
}

func TestInferenceServiceWithAnnotation(t *testing.T) {
	builder := buildValidISvcBuilder(buildISvcTestClient())
	builder.WithAnnotation("serving.kserve.io/deploymentMode", "RawDeployment")

	assert.Equal(t, "RawDeployment",
		builder.Definition.Annotations["serving.kserve.io/deploymentMode"])
	assert.Equal(t, "", builder.errorMsg)
}

func TestInferenceServiceWithEnv(t *testing.T) {
	builder := buildValidISvcBuilder(buildISvcTestClient())

	envVars := []corev1.EnvVar{
		{Name: "NEURON_CORES", Value: "2"},
		{Name: "MODEL_NAME", Value: "test-model"},
	}

	builder.WithEnv(envVars)

	assert.Equal(t, 2, len(builder.Definition.Spec.Predictor.Model.Env))
	assert.Equal(t, "NEURON_CORES", builder.Definition.Spec.Predictor.Model.Env[0].Name)
	assert.Equal(t, "2", builder.Definition.Spec.Predictor.Model.Env[0].Value)
	assert.Equal(t, "", builder.errorMsg)
}

func TestPullInferenceService(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                defaultISvcName,
			namespace:           defaultISvcNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultISvcNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("inferenceService 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultISvcName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("inferenceService 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "nonexistent",
			namespace:           defaultISvcNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf(
				"inferenceService nonexistent does not exist in namespace %s", defaultISvcNamespace),
			client: true,
		},
		{
			name:                defaultISvcName,
			namespace:           defaultISvcNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("inferenceService 'apiClient' cannot be nil"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testISvc := buildDummyInferenceService(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testISvc)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: isvcTestSchemes,
			})
		}

		builderResult, err := PullInferenceService(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError != nil {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testISvc.Name, builderResult.Object.Name)
			assert.Equal(t, testISvc.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestInferenceServiceExists(t *testing.T) {
	testCases := []struct {
		testISvc       *InferenceServiceBuilder
		expectedStatus bool
	}{
		{
			testISvc:       buildValidISvcBuilder(buildISvcClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testISvc:       buildInvalidISvcBuilder(buildISvcClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testISvc:       buildValidISvcBuilder(buildISvcTestClient()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testISvc.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestInferenceServiceCreate(t *testing.T) {
	testCases := []struct {
		testISvc    *InferenceServiceBuilder
		expectedErr bool
	}{
		{
			testISvc:    buildValidISvcBuilder(buildISvcTestClient()),
			expectedErr: false,
		},
		{
			testISvc:    buildInvalidISvcBuilder(buildISvcTestClient()),
			expectedErr: true,
		},
	}

	for _, testCase := range testCases {
		result, err := testCase.testISvc.Create()

		if testCase.expectedErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, result.Object)
			assert.Equal(t, defaultISvcName, result.Object.Name)
		}
	}
}

func TestInferenceServiceDelete(t *testing.T) {
	testCases := []struct {
		testISvc    *InferenceServiceBuilder
		expectedErr bool
	}{
		{
			testISvc:    buildValidISvcBuilder(buildISvcClientWithDummyObject()),
			expectedErr: false,
		},
		{
			testISvc:    buildValidISvcBuilder(buildISvcTestClient()),
			expectedErr: false,
		},
		{
			testISvc:    buildInvalidISvcBuilder(buildISvcTestClient()),
			expectedErr: true,
		},
	}

	for _, testCase := range testCases {
		result, err := testCase.testISvc.Delete()

		if testCase.expectedErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Nil(t, result.Object)
		}
	}
}

func TestInferenceServiceValidate(t *testing.T) {
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
			SchemeAttachers: isvcTestSchemes,
		})

		builder := &InferenceServiceBuilder{
			apiClient:  testSettings,
			Definition: &kservev1beta1.InferenceService{},
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

func buildValidISvcBuilder(apiClient *clients.Settings) *InferenceServiceBuilder {
	return NewInferenceServiceBuilder(apiClient, defaultISvcName, defaultISvcNamespace)
}

func buildInvalidISvcBuilder(apiClient *clients.Settings) *InferenceServiceBuilder {
	return NewInferenceServiceBuilder(apiClient, "", defaultISvcNamespace)
}

func buildISvcTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: isvcTestSchemes,
	})
}

func buildISvcClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyISvcObjects(),
		SchemeAttachers: isvcTestSchemes,
	})
}

func buildDummyISvcObjects() []runtime.Object {
	return []runtime.Object{
		buildDummyInferenceService(defaultISvcName, defaultISvcNamespace),
	}
}

func buildDummyInferenceService(name, namespace string) *kservev1beta1.InferenceService {
	return &kservev1beta1.InferenceService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
