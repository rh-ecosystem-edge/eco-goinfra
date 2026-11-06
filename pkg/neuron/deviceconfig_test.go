package neuron

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/neuron/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		v1beta1.AddToScheme,
	}
	defaultDeviceConfigName      = "neuron"
	defaultDeviceConfigNamespace = "ai-operator-on-aws"
	defaultDriversImage          = "public.ecr.aws/os-partners/neuron-openshift/neuron-kernel-module:2.25.4.0"
	defaultDriverVersion         = "2.24.7.0"
	defaultDevicePluginImage     = "public.ecr.aws/neuron/neuron-device-plugin:2.29.16.0"
	defaultDRADriverImage        = "public.ecr.aws/neuron/neuron-dra-driver:1.0.0"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, namespace string) *Builder {
				return NewBuilder(
					apiClient, name, namespace, defaultDriversImage, defaultDriverVersion, defaultDevicePluginImage)
			},
			v1beta1.AddToScheme,
			deviceConfigGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name              string
		driversImage      string
		driverVersion     string
		devicePluginImage string
		expectedError     string
	}{
		{
			name:              "valid images",
			driversImage:      defaultDriversImage,
			driverVersion:     defaultDriverVersion,
			devicePluginImage: defaultDevicePluginImage,
		},
		{
			name:              "empty driversImage",
			driverVersion:     defaultDriverVersion,
			devicePluginImage: defaultDevicePluginImage,
			expectedError:     errDriversImageEmpty,
		},
		{
			name:              "empty driverVersion",
			driversImage:      defaultDriversImage,
			devicePluginImage: defaultDevicePluginImage,
			expectedError:     errDriverVersionEmpty,
		},
		{
			name:          "empty devicePluginImage",
			driversImage:  defaultDriversImage,
			driverVersion: defaultDriverVersion,
			expectedError: errDevicePluginImageEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewBuilder(
				buildTestClientWithDeviceConfigScheme(),
				defaultDeviceConfigName,
				defaultDeviceConfigNamespace,
				testCase.driversImage,
				testCase.driverVersion,
				testCase.devicePluginImage,
			)

			if testCase.expectedError != "" {
				require.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.driversImage, builder.Definition.Spec.DriversImage)
			assert.Equal(t, testCase.driverVersion, builder.Definition.Spec.DriverVersion)
			assert.Equal(t, testCase.devicePluginImage, builder.Definition.Spec.DevicePluginImage)
		})
	}
}

func TestNewBuilderVariants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		newBuilder func(apiClient *clients.Settings, name, namespace string) *Builder
	}{
		{
			name: "in-cluster build",
			newBuilder: func(apiClient *clients.Settings, name, namespace string) *Builder {
				return NewBuilderWithInClusterBuild(
					apiClient, name, namespace, defaultDriverVersion, defaultDevicePluginImage)
			},
		},
		{
			name: "DRA",
			newBuilder: func(apiClient *clients.Settings, name, namespace string) *Builder {
				return NewBuilderWithDRA(
					apiClient, name, namespace, defaultDriversImage, defaultDriverVersion, defaultDRADriverImage)
			},
		},
		{
			name: "DRA in-cluster build",
			newBuilder: func(apiClient *clients.Settings, name, namespace string) *Builder {
				return NewBuilderWithInClusterBuildDRA(
					apiClient, name, namespace, defaultDriverVersion, defaultDRADriverImage)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testhelper.NewNamespacedBuilderTestConfig(
				testCase.newBuilder, v1beta1.AddToScheme, deviceConfigGVK).ExecuteTests(t)
		})
	}
}

type newBuilderValidationTestCase struct {
	name          string
	build         func(apiClient *clients.Settings) *Builder
	expectedSpec  v1beta1.DeviceConfigSpec
	expectedError string
}

// runNewBuilderValidationTests executes the package-specific constructor validation cases.
func runNewBuilderValidationTests(t *testing.T, testCases []newBuilderValidationTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.build(buildTestClientWithDeviceConfigScheme())

			if testCase.expectedError != "" {
				require.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.expectedSpec, builder.Definition.Spec)
		})
	}
}

func TestNewBuilderWithInClusterBuildValidation(t *testing.T) {
	t.Parallel()

	runNewBuilderValidationTests(t, []newBuilderValidationTestCase{
		{
			name: "valid",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithInClusterBuild(
					apiClient,
					defaultDeviceConfigName,
					defaultDeviceConfigNamespace,
					defaultDriverVersion,
					defaultDevicePluginImage,
				)
			},
			expectedSpec: v1beta1.DeviceConfigSpec{
				DriverVersion:     defaultDriverVersion,
				DevicePluginImage: defaultDevicePluginImage,
			},
		},
		{
			name: "empty driverVersion",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithInClusterBuild(
					apiClient, defaultDeviceConfigName, defaultDeviceConfigNamespace, "", defaultDevicePluginImage)
			},
			expectedError: errDriverVersionEmpty,
		},
		{
			name: "empty devicePluginImage",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithInClusterBuild(
					apiClient, defaultDeviceConfigName, defaultDeviceConfigNamespace, defaultDriverVersion, "")
			},
			expectedError: errDevicePluginImageEmpty,
		},
	})
}

func TestNewBuilderWithDRAValidation(t *testing.T) {
	t.Parallel()

	runNewBuilderValidationTests(t, []newBuilderValidationTestCase{
		{
			name: "valid",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithDRA(
					apiClient,
					defaultDeviceConfigName,
					defaultDeviceConfigNamespace,
					defaultDriversImage,
					defaultDriverVersion,
					defaultDRADriverImage,
				)
			},
			expectedSpec: v1beta1.DeviceConfigSpec{
				DriversImage:   defaultDriversImage,
				DriverVersion:  defaultDriverVersion,
				DRADriverImage: defaultDRADriverImage,
			},
		},
		{
			name: "empty driversImage",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithDRA(
					apiClient,
					defaultDeviceConfigName,
					defaultDeviceConfigNamespace,
					"",
					defaultDriverVersion,
					defaultDRADriverImage,
				)
			},
			expectedError: errDriversImageEmpty,
		},
		{
			name: "empty driverVersion",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithDRA(
					apiClient,
					defaultDeviceConfigName,
					defaultDeviceConfigNamespace,
					defaultDriversImage,
					"",
					defaultDRADriverImage,
				)
			},
			expectedError: errDriverVersionEmpty,
		},
		{
			name: "empty draDriverImage",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithDRA(
					apiClient,
					defaultDeviceConfigName,
					defaultDeviceConfigNamespace,
					defaultDriversImage,
					defaultDriverVersion,
					"",
				)
			},
			expectedError: errDRADriverImageEmpty,
		},
	})
}

func TestNewBuilderWithInClusterBuildDRAValidation(t *testing.T) {
	t.Parallel()

	runNewBuilderValidationTests(t, []newBuilderValidationTestCase{
		{
			name: "valid",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithInClusterBuildDRA(
					apiClient,
					defaultDeviceConfigName,
					defaultDeviceConfigNamespace,
					defaultDriverVersion,
					defaultDRADriverImage,
				)
			},
			expectedSpec: v1beta1.DeviceConfigSpec{
				DriverVersion:  defaultDriverVersion,
				DRADriverImage: defaultDRADriverImage,
			},
		},
		{
			name: "empty driverVersion",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithInClusterBuildDRA(
					apiClient, defaultDeviceConfigName, defaultDeviceConfigNamespace, "", defaultDRADriverImage)
			},
			expectedError: errDriverVersionEmpty,
		},
		{
			name: "empty draDriverImage",
			build: func(apiClient *clients.Settings) *Builder {
				return NewBuilderWithInClusterBuildDRA(
					apiClient, defaultDeviceConfigName, defaultDeviceConfigNamespace, defaultDriverVersion, "")
			},
			expectedError: errDRADriverImageEmpty,
		},
	})
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull, v1beta1.AddToScheme, deviceConfigGVK).ExecuteTests(t)
}

func TestDeviceConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[v1beta1.DeviceConfig, Builder](
		v1beta1.AddToScheme,
		deviceConfigGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestDeviceConfigWithSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		testSelector  map[string]string
		valid         bool
		expectedError string
	}{
		{
			name:         "valid selector",
			testSelector: map[string]string{"feature.node.kubernetes.io/aws-neuron": "true"},
			valid:        true,
		},
		{
			name:          "empty selector",
			testSelector:  map[string]string{},
			valid:         true,
			expectedError: errSelectorEmpty,
		},
		{
			name:         "invalid builder",
			testSelector: map[string]string{"feature.node.kubernetes.io/aws-neuron": "true"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildDeviceConfigBuilderForTest(testCase.valid)
			result := builder.WithSelector(testCase.testSelector)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.testSelector, result.Definition.Spec.Selector)
		})
	}
}

func TestDeviceConfigWithScheduler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		schedulerImage string
		extensionImage string
		valid          bool
		expectedError  string
	}{
		{
			name:           "valid scheduler",
			schedulerImage: "public.ecr.aws/eks-distro/kubernetes/kube-scheduler:v1.32.9-eks-1-32-24",
			extensionImage: "public.ecr.aws/neuron/neuron-scheduler:2.29.16.0",
			valid:          true,
		},
		{
			name:           "empty schedulerImage",
			extensionImage: "public.ecr.aws/neuron/neuron-scheduler:2.29.16.0",
			valid:          true,
			expectedError:  errSchedulerImageEmpty,
		},
		{
			name:           "empty extensionImage",
			schedulerImage: "public.ecr.aws/eks-distro/kubernetes/kube-scheduler:v1.32.9-eks-1-32-24",
			valid:          true,
			expectedError:  errExtensionImageEmpty,
		},
		{
			name:           "invalid builder",
			schedulerImage: "public.ecr.aws/eks-distro/kubernetes/kube-scheduler:v1.32.9-eks-1-32-24",
			extensionImage: "public.ecr.aws/neuron/neuron-scheduler:2.29.16.0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildDeviceConfigBuilderForTest(testCase.valid)
			result := builder.WithScheduler(testCase.schedulerImage, testCase.extensionImage)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.schedulerImage, result.Definition.Spec.CustomSchedulerImage)
			assert.Equal(t, testCase.extensionImage, result.Definition.Spec.SchedulerExtensionImage)
		})
	}
}

func TestDeviceConfigWithImageRepoSecret(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		secretName    string
		valid         bool
		expectedError string
	}{
		{
			name:       "valid secret",
			secretName: "my-secret",
			valid:      true,
		},
		{
			name:          "empty secret",
			valid:         true,
			expectedError: errImageRepoSecretEmpty,
		},
		{
			name:       "invalid builder",
			secretName: "my-secret",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildDeviceConfigBuilderForTest(testCase.valid)
			result := builder.WithImageRepoSecret(testCase.secretName)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.ImageRepoSecret)
			assert.Equal(t, testCase.secretName, result.Definition.Spec.ImageRepoSecret.Name)
		})
	}
}

func TestDeviceConfigWithDriverVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		driverVersion string
		valid         bool
		expectedError string
	}{
		{
			name:          "valid driver version",
			driverVersion: "2.25.0.0",
			valid:         true,
		},
		{
			name:          "empty driver version",
			valid:         true,
			expectedError: errDriverVersionEmpty,
		},
		{
			name:          "invalid builder",
			driverVersion: "2.25.0.0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildDeviceConfigBuilderForTest(testCase.valid)
			result := builder.WithDriverVersion(testCase.driverVersion)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.driverVersion, result.Definition.Spec.DriverVersion)
		})
	}
}

func TestDeviceConfigWithNodeMetricsImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		nodeMetricsImage string
		valid            bool
		expectedError    string
	}{
		{
			name:             "valid node metrics image",
			nodeMetricsImage: "public.ecr.aws/neuron/neuron-monitor:1.3.0",
			valid:            true,
		},
		{
			name:          "empty node metrics image",
			valid:         true,
			expectedError: errNodeMetricsImageEmpty,
		},
		{
			name:             "invalid builder",
			nodeMetricsImage: "public.ecr.aws/neuron/neuron-monitor:1.3.0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildDeviceConfigBuilderForTest(testCase.valid)
			result := builder.WithNodeMetricsImage(testCase.nodeMetricsImage)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.nodeMetricsImage, result.Definition.Spec.NodeMetricsImage)
		})
	}
}

func TestDeviceConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[v1beta1.DeviceConfig, Builder](
			v1beta1.AddToScheme,
			deviceConfigGVK,
			testhelper.ResourceScopeNamespaced,
		)).ExecuteTests(t)
}

func TestDeviceConfigWithDRADriverImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		draDriverImage string
		valid          bool
		expectedError  string
	}{
		{
			name:           "valid DRA driver image",
			draDriverImage: defaultDRADriverImage,
			valid:          true,
		},
		{
			name:          "empty DRA driver image",
			valid:         true,
			expectedError: errDRADriverImageEmpty,
		},
		{
			name:           "invalid builder",
			draDriverImage: defaultDRADriverImage,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildDeviceConfigBuilderForTest(testCase.valid)
			result := builder.WithDRADriverImage(testCase.draDriverImage)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.draDriverImage, result.Definition.Spec.DRADriverImage)
		})
	}
}

func TestDeviceConfigWithDeviceClasses(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		deviceClasses []v1beta1.DeviceClassSpec
		valid         bool
		expectedError string
	}{
		{
			name: "valid device classes",
			deviceClasses: []v1beta1.DeviceClassSpec{
				{Name: "neuron-device"},
			},
			valid: true,
		},
		{
			name:          "empty device classes",
			deviceClasses: []v1beta1.DeviceClassSpec{},
			valid:         true,
			expectedError: errDeviceClassesEmpty,
		},
		{
			name:          "nil device classes",
			valid:         true,
			expectedError: errDeviceClassesEmpty,
		},
		{
			name: "invalid builder",
			deviceClasses: []v1beta1.DeviceClassSpec{
				{Name: "neuron-device"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildDeviceConfigBuilderForTest(testCase.valid)
			result := builder.WithDeviceClasses(testCase.deviceClasses)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.deviceClasses, result.Definition.Spec.DeviceClasses)
		})
	}
}

// buildDummyDeviceConfig returns a DeviceConfig with the provided name and namespace.
func buildDummyDeviceConfig(name, namespace string) *v1beta1.DeviceConfig {
	return &v1beta1.DeviceConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.DeviceConfigSpec{
			DriversImage:      defaultDriversImage,
			DriverVersion:     defaultDriverVersion,
			DevicePluginImage: defaultDevicePluginImage,
		},
	}
}

// buildTestClientWithDeviceConfigScheme returns a client with the DeviceConfig scheme attached.
func buildTestClientWithDeviceConfigScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
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

// buildDeviceConfigBuilderForTest returns either a valid or invalid DeviceConfig builder.
func buildDeviceConfigBuilderForTest(valid bool) *Builder {
	if valid {
		return buildValidDeviceConfigBuilder(buildTestClientWithDummyObject())
	}

	return buildInvalidDeviceConfigBuilder(buildTestClientWithDummyObject())
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
		defaultDeviceConfigName,
		"",
		defaultDriversImage,
		defaultDriverVersion,
		defaultDevicePluginImage,
	)
}
