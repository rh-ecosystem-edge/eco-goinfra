package neuron

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/neuron/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListDeviceConfigs(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		func(apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*Builder, error) {
			return ListDeviceConfigs(apiClient, "", options...)
		},
		v1beta1.AddToScheme,
		deviceConfigGVK,
	).ExecuteTests(t)
}

func TestListDeviceConfigsNamespaceBehavior(t *testing.T) {
	t.Parallel()

	t.Run("list multiple DeviceConfigs in namespace", func(t *testing.T) {
		t.Parallel()

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: []runtime.Object{
				buildDummyDeviceConfig("neuron-1", defaultDeviceConfigNamespace),
				buildDummyDeviceConfig("neuron-2", defaultDeviceConfigNamespace),
			},
			SchemeAttachers: testSchemes,
		})

		builders, err := ListDeviceConfigs(testSettings, defaultDeviceConfigNamespace)
		require.NoError(t, err)
		require.Len(t, builders, 2)

		for _, builder := range builders {
			assert.Equal(t, defaultDeviceConfigNamespace, builder.Definition.Namespace)
		}
	})

	t.Run("namespace filter returns empty", func(t *testing.T) {
		t.Parallel()

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: []runtime.Object{
				buildDummyDeviceConfig(defaultDeviceConfigName, defaultDeviceConfigNamespace),
			},
			SchemeAttachers: testSchemes,
		})

		builders, err := ListDeviceConfigs(testSettings, "other-namespace")
		require.NoError(t, err)
		require.NotNil(t, builders)
		assert.Empty(t, builders)
	})

	t.Run("empty result returns empty slice", func(t *testing.T) {
		t.Parallel()

		testSettings := buildTestClientWithDeviceConfigScheme()

		builders, err := ListDeviceConfigs(testSettings, defaultDeviceConfigNamespace)
		require.NoError(t, err)
		require.NotNil(t, builders)
		assert.Empty(t, builders)
	})

	t.Run("empty namespace lists all namespaces", func(t *testing.T) {
		t.Parallel()

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: []runtime.Object{
				buildDummyDeviceConfig("neuron-1", defaultDeviceConfigNamespace),
				buildDummyDeviceConfig("neuron-2", "other-namespace"),
			},
			SchemeAttachers: testSchemes,
		})

		builders, err := ListDeviceConfigs(testSettings, "")
		require.NoError(t, err)
		assert.Len(t, builders, 2)
	})

	t.Run("namespace argument overrides list option", func(t *testing.T) {
		t.Parallel()

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: []runtime.Object{
				buildDummyDeviceConfig(defaultDeviceConfigName, defaultDeviceConfigNamespace),
			},
			SchemeAttachers: testSchemes,
		})

		builders, err := ListDeviceConfigs(
			testSettings,
			defaultDeviceConfigNamespace,
			runtimeclient.ListOptions{Namespace: "other-namespace"},
		)
		require.NoError(t, err)
		require.Len(t, builders, 1)
		assert.Equal(t, defaultDeviceConfigName, builders[0].Definition.Name)
	})
}
