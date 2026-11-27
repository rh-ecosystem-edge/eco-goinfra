package kserve

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	kservev1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kserve/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	defaultISvcName      = "test-isvc"
	defaultISvcNamespace = "test-ns"
)

func TestNewInferenceServiceBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewInferenceServiceBuilder,
		kservev1beta1.AddToScheme,
		inferenceServiceGVK,
	).ExecuteTests(t)
}

func TestPullInferenceService(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullInferenceService,
		kservev1beta1.AddToScheme,
		inferenceServiceGVK,
	).ExecuteTests(t)
}

func TestInferenceServiceBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[kservev1beta1.InferenceService, InferenceServiceBuilder](
		kservev1beta1.AddToScheme,
		inferenceServiceGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		Run(t)
}

func TestInferenceServiceWithModelFormat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		modelFormat    string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:        "valid model format",
			modelFormat: "vllm-neuron",
		},
		{
			name:          "empty model format",
			expectedError: "InferenceService 'modelFormat' cannot be empty",
		},
		{
			name:           "invalid builder",
			modelFormat:    "vllm-neuron",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidISvcTestBuilder(buildISvcTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidISvcTestBuilder(buildISvcTestClient())
			}

			builder.WithModelFormat(testCase.modelFormat)

			switch {
			case testCase.invalidBuilder:
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.Predictor.Model.ModelFormat.Name)
			case testCase.expectedError != "":
				assert.EqualError(t, builder.GetError(), testCase.expectedError)
				assert.Empty(t, builder.Definition.Spec.Predictor.Model.ModelFormat.Name)
			default:
				require.NoError(t, builder.GetError())
				assert.Equal(t, testCase.modelFormat,
					builder.Definition.Spec.Predictor.Model.ModelFormat.Name)
			}
		})
	}
}

func TestInferenceServiceWithRuntime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		runtime        string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:    "valid runtime",
			runtime: "vllm-neuron-runtime",
		},
		{
			name:          "empty runtime",
			expectedError: "InferenceService 'runtime' cannot be empty",
		},
		{
			name:           "invalid builder",
			runtime:        "vllm-neuron-runtime",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidISvcTestBuilder(buildISvcTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidISvcTestBuilder(buildISvcTestClient())
			}

			builder.WithRuntime(testCase.runtime)

			switch {
			case testCase.invalidBuilder:
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Nil(t, builder.Definition.Spec.Predictor.Model.Runtime)
			case testCase.expectedError != "":
				assert.EqualError(t, builder.GetError(), testCase.expectedError)
				assert.Nil(t, builder.Definition.Spec.Predictor.Model.Runtime)
			default:
				require.NoError(t, builder.GetError())
				assert.Equal(t, testCase.runtime, *builder.Definition.Spec.Predictor.Model.Runtime)
			}
		})
	}
}

func TestInferenceServiceWithStorageURI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		uri            string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name: "valid storage URI",
			uri:  "s3://bucket/model",
		},
		{
			name:          "empty storage URI",
			expectedError: "InferenceService 'storageUri' cannot be empty",
		},
		{
			name:           "invalid builder",
			uri:            "s3://bucket/model",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidISvcTestBuilder(buildISvcTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidISvcTestBuilder(buildISvcTestClient())
			}

			builder.WithStorageURI(testCase.uri)

			switch {
			case testCase.invalidBuilder:
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Nil(t, builder.Definition.Spec.Predictor.Model.StorageURI)
			case testCase.expectedError != "":
				assert.EqualError(t, builder.GetError(), testCase.expectedError)
				assert.Nil(t, builder.Definition.Spec.Predictor.Model.StorageURI)
			default:
				require.NoError(t, builder.GetError())
				assert.Equal(t, testCase.uri, *builder.Definition.Spec.Predictor.Model.StorageURI)
			}
		})
	}
}

func TestInferenceServiceWithServiceAccountName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		serviceAccountName string
		invalidBuilder     bool
	}{
		{
			name:               "valid service account name",
			serviceAccountName: "test-sa",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidISvcTestBuilder(buildISvcTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidISvcTestBuilder(buildISvcTestClient())
			}

			builder.WithServiceAccountName(testCase.serviceAccountName)

			if testCase.invalidBuilder {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.Predictor.ServiceAccountName)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.serviceAccountName, builder.Definition.Spec.Predictor.ServiceAccountName)
		})
	}
}

func TestInferenceServiceWithResources(t *testing.T) {
	t.Parallel()

	requests := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("2Gi"),
	}
	limits := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("2"),
		corev1.ResourceMemory: resource.MustParse("4Gi"),
	}

	testCases := []struct {
		name           string
		invalidBuilder bool
	}{
		{
			name: "valid resources",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidISvcTestBuilder(buildISvcTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidISvcTestBuilder(buildISvcTestClient())
			}

			builder.WithResources(requests, limits)

			if testCase.invalidBuilder {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.Predictor.Model.Resources.Requests)
				assert.Empty(t, builder.Definition.Spec.Predictor.Model.Resources.Limits)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, requests, builder.Definition.Spec.Predictor.Model.Resources.Requests)
			assert.Equal(t, limits, builder.Definition.Spec.Predictor.Model.Resources.Limits)
		})
	}
}

func TestInferenceServiceWithNeuronResources(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		devices        int64
		memoryRequest  string
		memoryLimit    string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:          "valid neuron resources",
			devices:       2,
			memoryRequest: "8Gi",
			memoryLimit:   "16Gi",
		},
		{
			name:          "invalid memory request",
			devices:       1,
			memoryRequest: "not-a-quantity",
			memoryLimit:   "8Gi",
			expectedError: "failed to parse memory request",
		},
		{
			name:          "invalid memory limit",
			devices:       1,
			memoryRequest: "8Gi",
			memoryLimit:   "not-a-quantity",
			expectedError: "failed to parse memory limit",
		},
		{
			name:           "invalid builder",
			devices:        1,
			memoryRequest:  "8Gi",
			memoryLimit:    "8Gi",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidISvcTestBuilder(buildISvcTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidISvcTestBuilder(buildISvcTestClient())
			}

			builder.WithNeuronResources(testCase.devices, testCase.memoryRequest, testCase.memoryLimit)

			switch {
			case testCase.invalidBuilder:
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.Predictor.Model.Resources.Requests)
				assert.Empty(t, builder.Definition.Spec.Predictor.Model.Resources.Limits)
			case testCase.expectedError != "":
				assert.ErrorContains(t, builder.GetError(), testCase.expectedError)
				assert.Empty(t, builder.Definition.Spec.Predictor.Model.Resources.Requests)
				assert.Empty(t, builder.Definition.Spec.Predictor.Model.Resources.Limits)
			default:
				require.NoError(t, builder.GetError())

				neuronQuantity := resource.MustParse("2")
				memoryRequest := resource.MustParse(testCase.memoryRequest)
				memoryLimit := resource.MustParse(testCase.memoryLimit)

				assert.Equal(t, neuronQuantity,
					builder.Definition.Spec.Predictor.Model.Resources.Requests["aws.amazon.com/neuron"])
				assert.Equal(t, neuronQuantity,
					builder.Definition.Spec.Predictor.Model.Resources.Limits["aws.amazon.com/neuron"])
				assert.Equal(t, memoryRequest,
					builder.Definition.Spec.Predictor.Model.Resources.Requests[corev1.ResourceMemory])
				assert.Equal(t, memoryLimit,
					builder.Definition.Spec.Predictor.Model.Resources.Limits[corev1.ResourceMemory])
			}
		})
	}
}

func TestInferenceServiceWithAnnotation(t *testing.T) {
	t.Parallel()

	const (
		annotationKey   = "serving.kserve.io/deploymentMode"
		annotationValue = "RawDeployment"
	)

	testCases := []struct {
		name           string
		invalidBuilder bool
	}{
		{
			name: "valid annotation",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidISvcTestBuilder(buildISvcTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidISvcTestBuilder(buildISvcTestClient())
			}

			builder.WithAnnotation(annotationKey, annotationValue)

			if testCase.invalidBuilder {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Annotations)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, annotationValue, builder.Definition.Annotations[annotationKey])
		})
	}
}

func TestInferenceServiceWithEnv(t *testing.T) {
	t.Parallel()

	envVars := []corev1.EnvVar{
		{Name: "NEURON_CORES", Value: "2"},
		{Name: "MODEL_NAME", Value: "test-model"},
	}

	testCases := []struct {
		name           string
		invalidBuilder bool
	}{
		{
			name: "valid environment variables",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidISvcTestBuilder(buildISvcTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidISvcTestBuilder(buildISvcTestClient())
			}

			builder.WithEnv(envVars)

			if testCase.invalidBuilder {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.Predictor.Model.Env)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, envVars, builder.Definition.Spec.Predictor.Model.Env)
		})
	}
}

func buildValidISvcTestBuilder(apiClient *clients.Settings) *InferenceServiceBuilder {
	return NewInferenceServiceBuilder(apiClient, defaultISvcName, defaultISvcNamespace)
}

func buildInvalidISvcTestBuilder(apiClient *clients.Settings) *InferenceServiceBuilder {
	return NewInferenceServiceBuilder(apiClient, "", defaultISvcNamespace)
}

func buildISvcTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{kservev1beta1.AddToScheme},
	})
}
