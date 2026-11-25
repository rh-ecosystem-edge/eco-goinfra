package kmm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	bootModuleConfigGVK = v1beta1.GroupVersion.WithKind("BootModuleConfig")

	defaultBootModuleConfigName      = "testbootmoduleconfig"
	defaultBootModuleConfigNamespace = "testns"
)

func TestNewBootModuleConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewBootModuleConfigBuilder,
		v1beta1.AddToScheme,
		bootModuleConfigGVK,
	).ExecuteTests(t)
}

func TestBootModuleConfigPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullBootModuleConfig,
		v1beta1.AddToScheme,
		bootModuleConfigGVK,
	).ExecuteTests(t)
}

func TestBootModuleConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(bootModuleConfigCommonTestConfig())).
		With(testhelper.NewExistsTestConfig(bootModuleConfigCommonTestConfig())).
		With(testhelper.NewCreateTestConfig(bootModuleConfigCommonTestConfig())).
		With(testhelper.NewDeleteReturnerTestConfig(bootModuleConfigCommonTestConfig())).
		With(testhelper.NewUpdateTestConfig(bootModuleConfigCommonTestConfig())).
		Run(t)
}

func TestBootModuleConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig[
		v1beta1.BootModuleConfig,
		BootModuleConfigBuilder,
		*v1beta1.BootModuleConfig,
		*BootModuleConfigBuilder,
		BootModuleConfigAdditionalOptions,
	](bootModuleConfigCommonTestConfig()).ExecuteTests(t)
}

func TestBootModuleConfigWithOptionsAppliesResourceOption(t *testing.T) {
	t.Parallel()

	testBuilder := buildValidBootModuleConfigBuilder(buildBootModuleConfigTestClient())
	testBuilder = testBuilder.WithOptions(func(builder *BootModuleConfigBuilder) (*BootModuleConfigBuilder, error) {
		builder.Definition.Spec.WorkerImage = "test-worker-image"

		return builder, nil
	})

	require.NoError(t, testBuilder.GetError())
	assert.Equal(t, "test-worker-image", testBuilder.Definition.Spec.WorkerImage)
}

func TestBootModuleConfigWithMachineConfigName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		value         string
		valid         bool
		expectedError error
	}{
		{
			name:  "sets machine config name",
			value: "test-machine-config",
			valid: true,
		},
		{
			name:          "empty machine config name",
			value:         "",
			valid:         true,
			expectedError: fmt.Errorf("bootmoduleconfig 'machineConfigName' cannot be empty"),
		},
		{
			name:  "invalid builder",
			value: "test-machine-config",
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBootModuleConfigBuilderForTestCase(testCase.valid)
			testBuilder = testBuilder.WithMachineConfigName(testCase.value)

			assertModifierResult(t, testBuilder, testCase.valid, testCase.expectedError,
				testBuilder.Definition.Spec.MachineConfigName, testCase.value)
		})
	}
}

func TestBootModuleConfigWithMachineConfigPoolName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		value         string
		valid         bool
		expectedError error
	}{
		{
			name:  "sets machine config pool name",
			value: "worker",
			valid: true,
		},
		{
			name:          "empty machine config pool name",
			value:         "",
			valid:         true,
			expectedError: fmt.Errorf("bootmoduleconfig 'machineConfigPoolName' cannot be empty"),
		},
		{
			name:  "invalid builder",
			value: "worker",
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBootModuleConfigBuilderForTestCase(testCase.valid)
			testBuilder = testBuilder.WithMachineConfigPoolName(testCase.value)

			assertModifierResult(t, testBuilder, testCase.valid, testCase.expectedError,
				testBuilder.Definition.Spec.MachineConfigPoolName, testCase.value)
		})
	}
}

func TestBootModuleConfigWithKernelModuleImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		value         string
		valid         bool
		expectedError error
	}{
		{
			name:  "sets kernel module image",
			value: "registry.example.com/driver:v1.0",
			valid: true,
		},
		{
			name:          "empty kernel module image",
			value:         "",
			valid:         true,
			expectedError: fmt.Errorf("bootmoduleconfig 'kernelModuleImage' cannot be empty"),
		},
		{
			name:  "invalid builder",
			value: "registry.example.com/driver:v1.0",
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBootModuleConfigBuilderForTestCase(testCase.valid)
			testBuilder = testBuilder.WithKernelModuleImage(testCase.value)

			assertModifierResult(t, testBuilder, testCase.valid, testCase.expectedError,
				testBuilder.Definition.Spec.KernelModuleImage, testCase.value)
		})
	}
}

func TestBootModuleConfigWithKernelModuleName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		value         string
		valid         bool
		expectedError error
	}{
		{
			name:  "sets kernel module name",
			value: "my_driver",
			valid: true,
		},
		{
			name:          "empty kernel module name",
			value:         "",
			valid:         true,
			expectedError: fmt.Errorf("bootmoduleconfig 'kernelModuleName' cannot be empty"),
		},
		{
			name:  "invalid builder",
			value: "my_driver",
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBootModuleConfigBuilderForTestCase(testCase.valid)
			testBuilder = testBuilder.WithKernelModuleName(testCase.value)

			assertModifierResult(t, testBuilder, testCase.valid, testCase.expectedError,
				testBuilder.Definition.Spec.KernelModuleName, testCase.value)
		})
	}
}

func TestBootModuleConfigWithInTreeModulesToRemove(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value []string
		valid bool
	}{
		{
			name:  "sets modules",
			value: []string{"old_driver", "legacy_module"},
			valid: true,
		},
		{
			name:  "sets empty modules",
			value: []string{},
			valid: true,
		},
		{
			name:  "sets nil modules",
			value: nil,
			valid: true,
		},
		{
			name:  "invalid builder",
			value: []string{"old_driver"},
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBootModuleConfigBuilderForTestCase(testCase.valid)
			testBuilder = testBuilder.WithInTreeModulesToRemove(testCase.value)

			assertModifierResult(t, testBuilder, testCase.valid, nil,
				testBuilder.Definition.Spec.InTreeModulesToRemove, testCase.value)
		})
	}
}

func TestBootModuleConfigWithFirmwareFilesPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "sets firmware files path",
			value: "/lib/firmware",
			valid: true,
		},
		{
			name:  "sets empty firmware files path",
			value: "",
			valid: true,
		},
		{
			name:  "invalid builder",
			value: "/lib/firmware",
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBootModuleConfigBuilderForTestCase(testCase.valid)
			testBuilder = testBuilder.WithFirmwareFilesPath(testCase.value)

			assertModifierResult(t, testBuilder, testCase.valid, nil,
				testBuilder.Definition.Spec.FirmwareFilesPath, testCase.value)
		})
	}
}

func TestBootModuleConfigWithWorkerImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "sets worker image",
			value: "registry.example.com/kmm-worker:v1.0",
			valid: true,
		},
		{
			name:  "sets empty worker image",
			value: "",
			valid: true,
		},
		{
			name:  "invalid builder",
			value: "registry.example.com/kmm-worker:v1.0",
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBootModuleConfigBuilderForTestCase(testCase.valid)
			testBuilder = testBuilder.WithWorkerImage(testCase.value)

			assertModifierResult(t, testBuilder, testCase.valid, nil,
				testBuilder.Definition.Spec.WorkerImage, testCase.value)
		})
	}
}

// bootModuleConfigCommonTestConfig creates the shared test configuration for the BootModuleConfig builder.
func bootModuleConfigCommonTestConfig() testhelper.CommonTestConfig[
	v1beta1.BootModuleConfig,
	BootModuleConfigBuilder,
	*v1beta1.BootModuleConfig,
	*BootModuleConfigBuilder,
] {
	return testhelper.NewCommonTestConfig[
		v1beta1.BootModuleConfig,
		BootModuleConfigBuilder,
	](v1beta1.AddToScheme, bootModuleConfigGVK, testhelper.ResourceScopeNamespaced)
}

// buildBootModuleConfigBuilderForTestCase returns a valid or invalid builder for modifier tests.
func buildBootModuleConfigBuilderForTestCase(valid bool) *BootModuleConfigBuilder {
	if valid {
		return buildValidBootModuleConfigBuilder(buildBootModuleConfigTestClient())
	}

	return buildInValidBootModuleConfigBuilder(buildBootModuleConfigTestClient())
}

// assertModifierResult checks the error and field value after a modifier is applied.
func assertModifierResult[T any](
	t *testing.T, testBuilder *BootModuleConfigBuilder, valid bool, expectedError error, actualValue T, expectedValue T) {
	t.Helper()

	if !valid {
		require.Error(t, testBuilder.GetError())
		assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))

		return
	}

	assert.Equal(t, expectedError, testBuilder.GetError())

	if expectedError == nil {
		assert.Equal(t, expectedValue, actualValue)
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

// buildBootModuleConfigTestClient returns a test client with the BootModuleConfig scheme attached.
func buildBootModuleConfigTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{v1beta1.AddToScheme},
	})
}
