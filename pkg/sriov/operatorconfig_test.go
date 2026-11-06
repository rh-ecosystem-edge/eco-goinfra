package sriov

import (
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultOperatorConfigNsName = "operatorconfig-testnamespace"

var operatorConfigGVK = srIovV1.GroupVersion.WithKind("SriovOperatorConfig")

func TestNewOperatorConfigBuilder(t *testing.T) {
	t.Parallel()

	t.Run("valid builder creation", func(t *testing.T) {
		t.Parallel()

		builder := NewOperatorConfigBuilder(newOperatorConfigTestClient(), defaultOperatorConfigNsName)

		require.NotNil(t, builder)
		require.NoError(t, builder.GetError())
		assert.Equal(t, sriovOperatorConfigName, builder.Definition.Name)
		assert.Equal(t, defaultOperatorConfigNsName, builder.Definition.Namespace)
		assert.Equal(t, operatorConfigGVK, builder.GetGVK())
	})

	t.Run("nil client returns error-bearing builder", func(t *testing.T) {
		t.Parallel()

		builder := NewOperatorConfigBuilder(nil, defaultOperatorConfigNsName)

		require.NotNil(t, builder)
		assert.True(t, commonerrors.IsAPIClientNil(builder.GetError()))
	})

	t.Run("empty namespace returns error-bearing builder", func(t *testing.T) {
		t.Parallel()

		builder := NewOperatorConfigBuilder(newOperatorConfigTestClient(), "")

		require.NotNil(t, builder)
		assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
	})
}

func TestPullOperatorConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid pull existing resource", func(t *testing.T) {
		t.Parallel()

		client := newOperatorConfigTestClient(buildDummyOperatorConfig(defaultOperatorConfigNsName))
		builder, err := PullOperatorConfig(client, defaultOperatorConfigNsName)

		require.NoError(t, err)
		require.NotNil(t, builder)
		require.NotNil(t, builder.Object)
		assert.Equal(t, sriovOperatorConfigName, builder.Definition.Name)
		assert.Equal(t, defaultOperatorConfigNsName, builder.Definition.Namespace)
		assert.Equal(t, operatorConfigGVK, builder.GetGVK())
	})

	t.Run("nil client returns error", func(t *testing.T) {
		t.Parallel()

		builder, err := PullOperatorConfig(nil, defaultOperatorConfigNsName)

		assert.Nil(t, builder)
		assert.True(t, commonerrors.IsAPIClientNil(err))
	})

	t.Run("non-existent resource returns not found", func(t *testing.T) {
		t.Parallel()

		builder, err := PullOperatorConfig(newOperatorConfigTestClient(), defaultOperatorConfigNsName)

		assert.Nil(t, builder)
		assert.True(t, k8serrors.IsNotFound(err))
	})

	t.Run("empty namespace returns error", func(t *testing.T) {
		t.Parallel()

		builder, err := PullOperatorConfig(newOperatorConfigTestClient(), "")

		assert.Nil(t, builder)
		assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
	})
}

func TestOperatorConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[srIovV1.SriovOperatorConfig, OperatorConfigBuilder](
		srIovV1.AddToScheme,
		operatorConfigGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestOperatorConfigWithInjector(t *testing.T) {
	t.Parallel()

	t.Run("valid values", func(t *testing.T) {
		t.Parallel()

		for _, injector := range []bool{true, false} {
			testBuilder := buildValidOperatorConfigTestBuilder().WithInjector(injector)

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, injector, testBuilder.Definition.Spec.EnableInjector)
		}
	})

	t.Run("invalid builder short-circuits", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildInvalidOperatorConfigTestBuilder().WithInjector(true)

		assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
		assert.False(t, testBuilder.Definition.Spec.EnableInjector)
	})
}

func TestOperatorConfigWithOperatorWebhook(t *testing.T) {
	t.Parallel()

	t.Run("valid values", func(t *testing.T) {
		t.Parallel()

		for _, webhook := range []bool{true, false} {
			testBuilder := buildValidOperatorConfigTestBuilder().WithOperatorWebhook(webhook)

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, webhook, testBuilder.Definition.Spec.EnableOperatorWebhook)
		}
	})

	t.Run("invalid builder short-circuits", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildInvalidOperatorConfigTestBuilder().WithOperatorWebhook(true)

		assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
		assert.False(t, testBuilder.Definition.Spec.EnableOperatorWebhook)
	})
}

func TestOperatorConfigWithConfigDaemonNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                      string
		configDaemonNodeSelector  map[string]string
		invalidBuilder            bool
		expectedError             string
		expectedConfigDaemonNodes map[string]string
	}{
		{
			name:                      "valid selector",
			configDaemonNodeSelector:  map[string]string{"test-node-selector-key": "test-node-selector-value"},
			expectedConfigDaemonNodes: map[string]string{"test-node-selector-key": "test-node-selector-value"},
		},
		{
			name:                      "empty selector value is allowed",
			configDaemonNodeSelector:  map[string]string{"test-node-selector-key": ""},
			expectedConfigDaemonNodes: map[string]string{"test-node-selector-key": ""},
		},
		{
			name:                     "empty selector returns error",
			configDaemonNodeSelector: map[string]string{},
			expectedError:            "can not apply empty configDaemonNodeSelector",
		},
		{
			name:                     "empty selector key returns error",
			configDaemonNodeSelector: map[string]string{"": "test-node-selector-value"},
			expectedError:            "can not apply configDaemonNodeSelector with an empty selectorKey value",
		},
		{
			name:                     "invalid builder short-circuits",
			configDaemonNodeSelector: map[string]string{"test-node-selector-key": "test-node-selector-value"},
			invalidBuilder:           true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidOperatorConfigTestBuilder()
			if testCase.invalidBuilder {
				testBuilder = buildInvalidOperatorConfigTestBuilder()
			}

			testBuilder.WithConfigDaemonNodeSelector(testCase.configDaemonNodeSelector)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.ConfigDaemonNodeSelector)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.ConfigDaemonNodeSelector)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.expectedConfigDaemonNodes,
				testBuilder.Definition.Spec.ConfigDaemonNodeSelector)
		})
	}
}

func TestOperatorConfigWithDisablePlugins(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		plugins         []string
		invalidBuilder  bool
		expectedError   string
		expectedPlugins srIovV1.PluginNameSlice
	}{
		{
			name:            "valid plugin",
			plugins:         []string{pluginMellanox},
			expectedPlugins: srIovV1.PluginNameSlice{srIovV1.PluginNameValue(pluginMellanox)},
		},
		{
			name:    "empty plugin list is allowed",
			plugins: []string{},
		},
		{
			name:          "invalid plugin returns error",
			plugins:       []string{"test"},
			expectedError: "invalid plugin parameter",
		},
		{
			name:           "invalid builder short-circuits",
			plugins:        []string{pluginMellanox},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidOperatorConfigTestBuilder()
			if testCase.invalidBuilder {
				testBuilder = buildInvalidOperatorConfigTestBuilder()
			}

			testBuilder.WithDisablePlugins(testCase.plugins)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.DisablePlugins)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.DisablePlugins)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.expectedPlugins, testBuilder.Definition.Spec.DisablePlugins)
		})
	}
}

func TestOperatorConfigRemoveDisablePlugins(t *testing.T) {
	t.Parallel()

	t.Run("removes configured plugins", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildValidOperatorConfigTestBuilder().WithDisablePlugins([]string{pluginMellanox})
		require.NoError(t, testBuilder.GetError())

		testBuilder.RemoveDisablePlugins()

		require.NoError(t, testBuilder.GetError())
		assert.Nil(t, testBuilder.Definition.Spec.DisablePlugins)
	})

	t.Run("invalid builder short-circuits", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildInvalidOperatorConfigTestBuilder()
		testBuilder.Definition.Spec.DisablePlugins = srIovV1.PluginNameSlice{
			srIovV1.PluginNameValue(pluginMellanox),
		}

		testBuilder.RemoveDisablePlugins()

		assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
		assert.Equal(t, srIovV1.PluginNameSlice{srIovV1.PluginNameValue(pluginMellanox)},
			testBuilder.Definition.Spec.DisablePlugins)
	})
}

func newOperatorConfigTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: objects,
		SchemeAttachers: []clients.SchemeAttacher{
			srIovV1.AddToScheme,
		},
	})
}

func buildDummyOperatorConfig(namespace string) *srIovV1.SriovOperatorConfig {
	return &srIovV1.SriovOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sriovOperatorConfigName,
			Namespace: namespace,
		},
	}
}

func buildValidOperatorConfigTestBuilder() *OperatorConfigBuilder {
	return NewOperatorConfigBuilder(newOperatorConfigTestClient(), defaultOperatorConfigNsName)
}

func buildInvalidOperatorConfigTestBuilder() *OperatorConfigBuilder {
	return NewOperatorConfigBuilder(newOperatorConfigTestClient(), "")
}
