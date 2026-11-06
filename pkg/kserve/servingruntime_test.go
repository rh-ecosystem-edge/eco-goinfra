package kserve

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	kservev1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kserve/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultSRName      = "test-runtime"
	defaultSRNamespace = "test-ns"
)

func TestNewServingRuntimeBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewServingRuntimeBuilder,
		kservev1alpha1.AddToScheme,
		servingRuntimeGVK,
	).ExecuteTests(t)

	t.Run("initializes MultiModel to false", func(t *testing.T) {
		t.Parallel()

		builder := buildValidSRTestBuilder(buildSRTestClient())

		require.NoError(t, builder.GetError())
		require.NotNil(t, builder.Definition.Spec.MultiModel)
		assert.False(t, *builder.Definition.Spec.MultiModel)
	})
}

func TestPullServingRuntime(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullServingRuntime,
		kservev1alpha1.AddToScheme,
		servingRuntimeGVK,
	).ExecuteTests(t)
}

func TestServingRuntimeBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[kservev1alpha1.ServingRuntime, ServingRuntimeBuilder](
		kservev1alpha1.AddToScheme,
		servingRuntimeGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		Run(t)
}

func TestServingRuntimeWithModelFormat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		modelFormat    string
		autoSelect     bool
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:        "valid model format",
			modelFormat: "vllm-neuron",
			autoSelect:  true,
		},
		{
			name:          "empty model format",
			expectedError: "ServingRuntime model format 'name' cannot be empty",
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

			builder := buildValidSRTestBuilder(buildSRTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidSRTestBuilder(buildSRTestClient())
			}

			builder.WithModelFormat(testCase.modelFormat, testCase.autoSelect)

			switch {
			case testCase.invalidBuilder:
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.SupportedModelFormats)
			case testCase.expectedError != "":
				assert.EqualError(t, builder.GetError(), testCase.expectedError)
				assert.Empty(t, builder.Definition.Spec.SupportedModelFormats)
			default:
				require.NoError(t, builder.GetError())
				require.Len(t, builder.Definition.Spec.SupportedModelFormats, 1)
				assert.Equal(t, testCase.modelFormat, builder.Definition.Spec.SupportedModelFormats[0].Name)
				assert.Equal(t, testCase.autoSelect, builder.Definition.Spec.SupportedModelFormats[0].AutoSelect)
			}
		})
	}
}

func TestServingRuntimeWithContainer(t *testing.T) {
	t.Parallel()

	container := corev1.Container{
		Name:  "vllm",
		Image: "registry.redhat.io/rhaiis/vllm-neuron-rhel9:3",
		Ports: []corev1.ContainerPort{
			{ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
		},
	}

	testCases := []struct {
		name           string
		invalidBuilder bool
	}{
		{
			name: "valid container",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidSRTestBuilder(buildSRTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidSRTestBuilder(buildSRTestClient())
			}

			builder.WithContainer(container)

			if testCase.invalidBuilder {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.Containers)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, []corev1.Container{container}, builder.Definition.Spec.Containers)
		})
	}
}

func TestServingRuntimeWithVolume(t *testing.T) {
	t.Parallel()

	volume := corev1.Volume{
		Name: "shm",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumMemory,
			},
		},
	}

	testCases := []struct {
		name           string
		invalidBuilder bool
	}{
		{
			name: "valid volume",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidSRTestBuilder(buildSRTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidSRTestBuilder(buildSRTestClient())
			}

			builder.WithVolume(volume)

			if testCase.invalidBuilder {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.Volumes)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, []corev1.Volume{volume}, builder.Definition.Spec.Volumes)
		})
	}
}

func TestServingRuntimeWithAnnotation(t *testing.T) {
	t.Parallel()

	const (
		annotationKey   = "serving.kserve.io/autoscalerClass"
		annotationValue = "hpa"
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

			builder := buildValidSRTestBuilder(buildSRTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidSRTestBuilder(buildSRTestClient())
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

func buildValidSRTestBuilder(apiClient *clients.Settings) *ServingRuntimeBuilder {
	return NewServingRuntimeBuilder(apiClient, defaultSRName, defaultSRNamespace)
}

func buildInvalidSRTestBuilder(apiClient *clients.Settings) *ServingRuntimeBuilder {
	return NewServingRuntimeBuilder(apiClient, "", defaultSRNamespace)
}

func buildSRTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{kservev1alpha1.AddToScheme},
	})
}
