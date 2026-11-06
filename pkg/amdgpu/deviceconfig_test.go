package amdgpu

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	amdgpuv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/amd/gpu-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDeviceConfigName      = "test-deviceconfig"
	testDeviceConfigNamespace = "test-openshift-amd-gpu"
)

var almExample = fmt.Sprintf(`
[
	{
		"apiVersion": "amd.com/v1alpha1",
		"kind": "DeviceConfig",
		"metadata": {
			"name": "%s",
			"namespace": "%s"
		},
		"spec": {
			"driver": {
				"enable": true
			}
		}
	}
]
`, testDeviceConfigName, testDeviceConfigNamespace)

func TestAMDGPUNewBuilderFromObjectString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		almExample        string
		client            bool
		expectedErrorText string
	}{
		{
			name:       "valid ALM example",
			almExample: almExample,
			client:     true,
		},
		{
			name:              "empty ALM example",
			almExample:        "",
			client:            true,
			expectedErrorText: "error initializing DeviceConfig from alm-examples: almExample is an empty string",
		},
		{
			name:              "invalid ALM example",
			almExample:        "{ invalid: data }",
			client:            true,
			expectedErrorText: "error initializing DeviceConfig from alm-examples:",
		},
		{
			name:              "empty ALM example list",
			almExample:        "[]",
			client:            true,
			expectedErrorText: "error initializing DeviceConfig from alm-examples: failed to get alm examples",
		},
		{
			name:       "no client",
			almExample: almExample,
			client:     false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var client *clients.Settings
			if testCase.client {
				client = buildTestClientWithAMDGPUScheme()
			}

			builder := NewBuilderFromObjectString(client, testCase.almExample)

			if !testCase.client {
				assert.Nil(t, builder)

				return
			}

			require.NotNil(t, builder)
			assert.Equal(t, deviceConfigGVK, builder.GetGVK())

			if testCase.expectedErrorText != "" {
				assert.ErrorContains(t, builder.GetError(), testCase.expectedErrorText)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition)
			assert.Equal(t, testDeviceConfigName, builder.Definition.Name)
			assert.Equal(t, testDeviceConfigNamespace, builder.Definition.Namespace)
			require.NotNil(t, builder.Definition.Spec.Driver.Enable)
			assert.True(t, *builder.Definition.Spec.Driver.Enable)
		})
	}
}

func TestAMDGPUPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull, amdgpuv1.AddToScheme, deviceConfigGVK).ExecuteTests(t)
}

func TestAMDGPUBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[amdgpuv1.DeviceConfig, Builder](
		amdgpuv1.AddToScheme,
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

func buildTestClientWithAMDGPUScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{amdgpuv1.AddToScheme},
	})
}
