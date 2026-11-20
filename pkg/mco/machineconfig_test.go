package mco

import (
	"testing"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultMachineConfigName = "test-machine-config"

var testSchemes = []clients.SchemeAttacher{mcv1.Install}

func TestNewMachineConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		NewMCBuilder, mcv1.Install, machineConfigGVK).ExecuteTests(t)
}

func TestPullMachineConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullMachineConfig, mcv1.Install, machineConfigGVK).ExecuteTests(t)
}

func TestMachineConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newMachineConfigCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestMachineConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newMachineConfigCommonTestConfig()).ExecuteTests(t)
}

func TestMachineConfigWithLabel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *MCBuilder
		key         string
		value       string
		assertError func(error) bool
	}{
		{
			name: "sets label on valid builder",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			key:         "test-key",
			value:       "test-value",
			assertError: isNilError,
		},
		{
			name: "empty key sets builder error",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			key:         "",
			value:       "test-value",
			assertError: hasErrorText(errEmptyKey),
		},
		{
			name: "invalid builder short circuits",
			builder: func() *MCBuilder {
				return buildInvalidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			key:         "test-key",
			value:       "test-value",
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithLabel(testCase.key, testCase.value)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.value, result.Definition.Labels[testCase.key])
			} else {
				assert.Empty(t, result.Definition.Labels)
			}
		})
	}
}

func TestMachineConfigWithKernelArguments(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *MCBuilder
		kernelArgs  []string
		assertError func(error) bool
	}{
		{
			name: "sets kernel arguments on valid builder",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			kernelArgs:  []string{"test"},
			assertError: isNilError,
		},
		{
			name: "empty kernel arguments sets builder error",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			kernelArgs:  nil,
			assertError: hasErrorText("'kernelArgs' cannot be empty"),
		},
		{
			name: "invalid builder short circuits",
			builder: func() *MCBuilder {
				return buildInvalidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			kernelArgs:  []string{"test"},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithKernelArguments(testCase.kernelArgs)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.kernelArgs, result.Definition.Spec.KernelArguments)
			} else {
				assert.Nil(t, result.Definition.Spec.KernelArguments)
			}
		})
	}
}

func TestMachineConfigWithExtensions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *MCBuilder
		extensions  []string
		assertError func(error) bool
	}{
		{
			name: "sets extensions on valid builder",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			extensions:  []string{"test"},
			assertError: isNilError,
		},
		{
			name: "empty extensions sets builder error",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			extensions:  nil,
			assertError: hasErrorText("'extensions' cannot be empty"),
		},
		{
			name: "invalid builder short circuits",
			builder: func() *MCBuilder {
				return buildInvalidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			extensions:  []string{"test"},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithExtensions(testCase.extensions)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.extensions, result.Definition.Spec.Extensions)
			} else {
				assert.Nil(t, result.Definition.Spec.Extensions)
			}
		})
	}
}

func TestMachineConfigWithRawConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *MCBuilder
		rawConfig   []byte
		assertError func(error) bool
	}{
		{
			name: "sets raw config on valid builder",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			rawConfig:   []byte{0, 0, 0},
			assertError: isNilError,
		},
		{
			name: "empty raw config sets builder error",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			rawConfig:   []byte{},
			assertError: hasErrorText("'Config.Raw' cannot be empty"),
		},
		{
			name: "invalid builder short circuits",
			builder: func() *MCBuilder {
				return buildInvalidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			rawConfig:   []byte{0, 0, 0},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithRawConfig(testCase.rawConfig)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.rawConfig, result.Definition.Spec.Config.Raw)
			} else {
				assert.Nil(t, result.Definition.Spec.Config.Raw)
			}
		})
	}
}

func TestMachineConfigWithFIPS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *MCBuilder
		fips        bool
		assertError func(error) bool
	}{
		{
			name: "sets FIPS on valid builder",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			fips:        true,
			assertError: isNilError,
		},
		{
			name: "invalid builder short circuits",
			builder: func() *MCBuilder {
				return buildInvalidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			fips:        true,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithFIPS(testCase.fips)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.fips, result.Definition.Spec.FIPS)
			} else {
				assert.False(t, result.Definition.Spec.FIPS)
			}
		})
	}
}

func TestMachineConfigWithKernelType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *MCBuilder
		kernelType  string
		assertError func(error) bool
	}{
		{
			name: "sets kernel type on valid builder",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			kernelType:  "test",
			assertError: isNilError,
		},
		{
			name: "empty kernel type sets builder error",
			builder: func() *MCBuilder {
				return buildValidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			kernelType:  "",
			assertError: hasErrorText("'kernelType' cannot be empty"),
		},
		{
			name: "invalid builder short circuits",
			builder: func() *MCBuilder {
				return buildInvalidMachineConfigTestBuilder(newMachineConfigTestClient())
			},
			kernelType:  "test",
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithKernelType(testCase.kernelType)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.kernelType, result.Definition.Spec.KernelType)
			} else {
				assert.Empty(t, result.Definition.Spec.KernelType)
			}
		})
	}
}

func newMachineConfigCommonTestConfig() testhelper.CommonTestConfig[
	mcv1.MachineConfig, MCBuilder, *mcv1.MachineConfig, *MCBuilder,
] {
	return testhelper.NewCommonTestConfig[mcv1.MachineConfig, MCBuilder](
		mcv1.Install, machineConfigGVK, testhelper.ResourceScopeClusterScoped)
}

func newMachineConfigTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

func buildValidMachineConfigTestBuilder(apiClient *clients.Settings) *MCBuilder {
	return NewMCBuilder(apiClient, defaultMachineConfigName)
}

func buildInvalidMachineConfigTestBuilder(apiClient *clients.Settings) *MCBuilder {
	return NewMCBuilder(apiClient, "")
}

func isNilError(err error) bool {
	return err == nil
}

func hasErrorText(expected string) func(error) bool {
	return func(err error) bool {
		return err != nil && err.Error() == expected
	}
}
