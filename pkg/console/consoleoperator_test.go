package console

import (
	"testing"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	consoleOperatorGVK         = operatorv1.GroupVersion.WithKind("Console")
	defaultConsoleOperatorName = "cluster"
	defaultPluginsList         = []string{"monitoring-plugin", "nmstate-console-plugin"}
)

func TestPullConsoleOperator(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullConsoleOperator, operatorv1.Install, consoleOperatorGVK).ExecuteTests(t)
}

func TestConsoleOperatorBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newConsoleOperatorCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestConsoleOperatorGetPlugins(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *ConsoleOperatorBuilder
		expected    []string
		assertError func(error) bool
	}{
		{
			name: "returns plugins from existing console operator",
			builder: func() *ConsoleOperatorBuilder {
				return buildValidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject())
			},
			expected:    defaultPluginsList,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name: "invalid builder returns error",
			builder: func() *ConsoleOperatorBuilder {
				return buildInvalidConsoleOperatorBuilder(buildConsoleOperatorClientWithDummyObject())
			},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
		{
			name: "missing console operator returns error",
			builder: func() *ConsoleOperatorBuilder {
				return buildValidConsoleOperatorBuilder(newConsoleOperatorTestClient())
			},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "consoleOperator cluster object does not exist"
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			plugins, err := builder.GetPlugins()

			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)

			if err == nil {
				require.NotNil(t, plugins)
				assert.Equal(t, testCase.expected, *plugins)
			} else {
				assert.Nil(t, plugins)
			}
		})
	}
}

func TestConsoleOperatorWithPlugins(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		plugins         []string
		redefine        bool
		expectedPlugins []string
		assertError     func(error) bool
		invalidBuilder  bool
	}{
		{
			name:            "redefines plugins",
			plugins:         []string{"test-new-plugin"},
			redefine:        true,
			expectedPlugins: []string{"test-new-plugin"},
			assertError:     func(err error) bool { return err == nil },
		},
		{
			name:            "appends new plugins",
			plugins:         []string{"test-new-plugin"},
			expectedPlugins: []string{"monitoring-plugin", "nmstate-console-plugin", "test-new-plugin"},
			assertError:     func(err error) bool { return err == nil },
		},
		{
			name:            "does not duplicate existing plugins",
			plugins:         defaultPluginsList,
			expectedPlugins: defaultPluginsList,
			assertError:     func(err error) bool { return err == nil },
		},
		{
			name:            "empty plugins sets builder error",
			plugins:         []string{},
			expectedPlugins: defaultPluginsList,
			assertError: func(err error) bool {
				return err != nil && err.Error() == "the newPluginsList can not be empty"
			},
		},
		{
			name:            "invalid builder short-circuits",
			plugins:         []string{"test-new-plugin"},
			expectedPlugins: defaultPluginsList,
			assertError:     commonerrors.IsBuilderNameEmpty,
			invalidBuilder:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildConsoleOperatorClientWithDummyObject()
			builder := buildValidConsoleOperatorBuilder(apiClient)

			if testCase.invalidBuilder {
				builder = buildInvalidConsoleOperatorBuilder(apiClient)
			}

			result := builder.WithPlugins(testCase.plugins, testCase.redefine)

			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())
			assert.Equal(t, testCase.expectedPlugins, result.Definition.Spec.Plugins)
		})
	}
}

func TestConsoleOperatorUpdatePlugins(t *testing.T) {
	t.Parallel()

	builder, err := PullConsoleOperator(buildConsoleOperatorClientWithDummyObject(), defaultConsoleOperatorName)
	require.NoError(t, err)

	expectedPlugins := []string{"odf-console"}
	builder.WithPlugins(expectedPlugins, true)
	require.NoError(t, builder.GetError())

	updatedBuilder, err := builder.Update()
	require.NoError(t, err)
	require.NotNil(t, updatedBuilder.Object)
	assert.Equal(t, expectedPlugins, updatedBuilder.Object.Spec.Plugins)
}

func newConsoleOperatorCommonTestConfig() testhelper.CommonTestConfig[
	operatorv1.Console, ConsoleOperatorBuilder, *operatorv1.Console, *ConsoleOperatorBuilder] {
	return testhelper.NewCommonTestConfig[operatorv1.Console, ConsoleOperatorBuilder](
		operatorv1.Install,
		consoleOperatorGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

func buildValidConsoleOperatorBuilder(apiClient *clients.Settings) *ConsoleOperatorBuilder {
	builder := common.NewClusterScopedBuilder[operatorv1.Console, ConsoleOperatorBuilder](
		apiClient, operatorv1.Install, defaultConsoleOperatorName)
	builder.Definition.Spec.Plugins = append([]string(nil), defaultPluginsList...)

	return builder
}

func buildInvalidConsoleOperatorBuilder(apiClient *clients.Settings) *ConsoleOperatorBuilder {
	builder := common.NewClusterScopedBuilder[operatorv1.Console, ConsoleOperatorBuilder](
		apiClient, operatorv1.Install, "")
	builder.Definition.Spec.Plugins = append([]string(nil), defaultPluginsList...)

	return builder
}

func buildConsoleOperatorClientWithDummyObject() *clients.Settings {
	return newConsoleOperatorTestClient(buildDummyConsoleOperator()...)
}

func newConsoleOperatorTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{operatorv1.Install},
	})
}

func buildDummyConsoleOperator() []runtime.Object {
	return []runtime.Object{
		&operatorv1.Console{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultConsoleOperatorName,
			},
			Spec: operatorv1.ConsoleSpec{
				Plugins: append([]string(nil), defaultPluginsList...),
			},
		},
	}
}
