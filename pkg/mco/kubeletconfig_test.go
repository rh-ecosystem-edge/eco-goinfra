package mco

import (
	"testing"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

const (
	defaultKubeletConfigName    = "test-kubeletconfig"
	defaultMCPoolSelectorKey    = "test-pool-selector"
	defaultMCPoolSelectorValue  = ""
	defaultSystemReservedCPU    = "100m"
	defaultSystemReservedMemory = "100Gi"
)

func TestNewKubeletConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		NewKubeletConfigBuilder, mcv1.Install, kubeletConfigGVK).ExecuteTests(t)
}

func TestPullKubeletConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullKubeletConfig, mcv1.Install, kubeletConfigGVK).ExecuteTests(t)
}

func TestKubeletConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[mcv1.KubeletConfig, KubeletConfigBuilder](
		mcv1.Install,
		kubeletConfigGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestKubeletConfigWithMCPoolSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		key           string
		valid         bool
		expectedError string
	}{
		{
			name:  "valid key",
			key:   defaultMCPoolSelectorKey,
			valid: true,
		},
		{
			name:          "empty key",
			key:           "",
			valid:         true,
			expectedError: errEmptyKey,
		},
		{
			name:  "invalid builder short circuits",
			key:   defaultMCPoolSelectorKey,
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := clients.GetTestClients(clients.TestClientParams{})

			testBuilder := buildValidKubeletConfigBuilder(apiClient)
			if !testCase.valid {
				testBuilder = buildInvalidKubeletConfigBuilder(apiClient)
			}

			result := testBuilder.WithMCPoolSelector(testCase.key, defaultMCPoolSelectorValue)
			require.Same(t, testBuilder, result)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(
				t, defaultMCPoolSelectorValue, result.Definition.Spec.MachineConfigPoolSelector.MatchLabels[testCase.key])
		})
	}
}

func TestKubeletConfigWithSystemReserved(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		cpu           string
		memory        string
		valid         bool
		expectedError string
	}{
		{
			name:   "valid system reserved",
			cpu:    defaultSystemReservedCPU,
			memory: defaultSystemReservedMemory,
			valid:  true,
		},
		{
			name:   "invalid builder short circuits",
			cpu:    defaultSystemReservedCPU,
			memory: defaultSystemReservedMemory,
			valid:  false,
		},
		{
			name:          "empty cpu",
			cpu:           "",
			memory:        defaultSystemReservedMemory,
			valid:         true,
			expectedError: "'cpu' cannot be empty",
		},
		{
			name:          "empty memory",
			cpu:           defaultSystemReservedCPU,
			memory:        "",
			valid:         true,
			expectedError: "'memory' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := clients.GetTestClients(clients.TestClientParams{})

			testBuilder := buildValidKubeletConfigBuilder(apiClient)
			if !testCase.valid {
				testBuilder = buildInvalidKubeletConfigBuilder(apiClient)
			}

			result := testBuilder.WithSystemReserved(testCase.cpu, testCase.memory)
			require.Same(t, testBuilder, result)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())

			systemReserved, ok := result.Definition.Spec.KubeletConfig.Object.(*kubeletconfigv1beta1.KubeletConfiguration)
			require.True(t, ok)

			assert.Equal(t, testCase.cpu, systemReserved.SystemReserved["cpu"])
			assert.Equal(t, testCase.memory, systemReserved.SystemReserved["memory"])
		})
	}
}

func TestKubeletConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[mcv1.KubeletConfig, KubeletConfigBuilder](
			mcv1.Install,
			kubeletConfigGVK,
			testhelper.ResourceScopeClusterScoped,
		)).ExecuteTests(t)
}

// buildValidKubeletConfigBuilder returns a valid KubeletConfigBuilder for testing.
func buildValidKubeletConfigBuilder(apiClient *clients.Settings) *KubeletConfigBuilder {
	return NewKubeletConfigBuilder(apiClient, defaultKubeletConfigName)
}

// buildInvalidKubeletConfigBuilder returns a invalid KubeletConfigBuilder for testing.
func buildInvalidKubeletConfigBuilder(apiClient *clients.Settings) *KubeletConfigBuilder {
	return NewKubeletConfigBuilder(apiClient, "")
}
