package keda

import (
	"testing"

	kedav1alpha1 "github.com/kedacore/keda-olm-operator/api/keda/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultKedaControllerName      = "keda"
	defaultKedaControllerNamespace = "openshift-keda"
)

func TestNewControllerBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewControllerBuilder, kedav1alpha1.AddToScheme, controllerGVK).ExecuteTests(t)
}

func TestPullController(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullController, kedav1alpha1.AddToScheme, controllerGVK).ExecuteTests(t)
}

func TestControllerBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[kedav1alpha1.KedaController, ControllerBuilder](
		kedav1alpha1.AddToScheme,
		controllerGVK,
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

func TestControllerWithAdmissionWebhooks(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		admissionWebhooks   kedav1alpha1.KedaAdmissionWebhooksSpec
		admissionWebhooksOK bool
	}{
		{
			name: "configured admission webhooks",
			admissionWebhooks: kedav1alpha1.KedaAdmissionWebhooksSpec{
				LogLevel:   "info",
				LogEncoder: "console",
			},
			admissionWebhooksOK: true,
		},
		{
			name:                "empty admission webhooks",
			admissionWebhooks:   kedav1alpha1.KedaAdmissionWebhooksSpec{},
			admissionWebhooksOK: true,
		},
		{
			name: "invalid builder short-circuits",
			admissionWebhooks: kedav1alpha1.KedaAdmissionWebhooksSpec{
				LogLevel: "info",
			},
			admissionWebhooksOK: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildControllerClientWithDummyObject()

			builder := buildValidControllerBuilder(apiClient)
			if !testCase.admissionWebhooksOK {
				builder = buildInvalidControllerBuilder(apiClient)
			}

			result := builder.WithAdmissionWebhooks(testCase.admissionWebhooks)

			require.Same(t, builder, result)

			if !testCase.admissionWebhooksOK {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.AdmissionWebhooks)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.admissionWebhooks, result.Definition.Spec.AdmissionWebhooks)
		})
	}
}

func TestControllerWithOperator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		operator   kedav1alpha1.KedaOperatorSpec
		operatorOK bool
	}{
		{
			name: "configured operator",
			operator: kedav1alpha1.KedaOperatorSpec{
				LogLevel:   "info",
				LogEncoder: "console",
			},
			operatorOK: true,
		},
		{
			name:       "empty operator",
			operator:   kedav1alpha1.KedaOperatorSpec{},
			operatorOK: true,
		},
		{
			name:       "invalid builder short-circuits",
			operator:   kedav1alpha1.KedaOperatorSpec{LogLevel: "info"},
			operatorOK: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildControllerClientWithDummyObject()

			builder := buildValidControllerBuilder(apiClient)
			if !testCase.operatorOK {
				builder = buildInvalidControllerBuilder(apiClient)
			}

			result := builder.WithOperator(testCase.operator)

			require.Same(t, builder, result)

			if !testCase.operatorOK {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.Operator)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.operator, result.Definition.Spec.Operator)
		})
	}
}

func TestControllerWithMetricsServer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		metricsServer kedav1alpha1.KedaMetricsServerSpec
		metricsOK     bool
	}{
		{
			name:          "configured metrics server",
			metricsServer: kedav1alpha1.KedaMetricsServerSpec{LogLevel: "0"},
			metricsOK:     true,
		},
		{
			name:          "empty metrics server",
			metricsServer: kedav1alpha1.KedaMetricsServerSpec{},
			metricsOK:     true,
		},
		{
			name:          "invalid builder short-circuits",
			metricsServer: kedav1alpha1.KedaMetricsServerSpec{LogLevel: "0"},
			metricsOK:     false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildControllerClientWithDummyObject()

			builder := buildValidControllerBuilder(apiClient)
			if !testCase.metricsOK {
				builder = buildInvalidControllerBuilder(apiClient)
			}

			result := builder.WithMetricsServer(testCase.metricsServer)

			require.Same(t, builder, result)

			if !testCase.metricsOK {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.MetricsServer)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.metricsServer, result.Definition.Spec.MetricsServer)
		})
	}
}

func TestControllerWithWatchNamespace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		watchNamespace string
		expectedError  string
		valid          bool
	}{
		{
			name:           "configured watch namespace",
			watchNamespace: "test-app",
			valid:          true,
		},
		{
			name:          "empty watch namespace",
			expectedError: "'watchNamespace' argument cannot be empty",
			valid:         true,
		},
		{
			name:           "invalid builder short-circuits",
			watchNamespace: "test-app",
			valid:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildControllerClientWithDummyObject()

			builder := buildValidControllerBuilder(apiClient)
			if !testCase.valid {
				builder = buildInvalidControllerBuilder(apiClient)
			}

			result := builder.WithWatchNamespace(testCase.watchNamespace)

			require.Same(t, builder, result)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.WatchNamespace)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Empty(t, result.Definition.Spec.WatchNamespace)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.watchNamespace, result.Definition.Spec.WatchNamespace)
		})
	}
}

func buildValidControllerBuilder(apiClient *clients.Settings) *ControllerBuilder {
	return NewControllerBuilder(apiClient, defaultKedaControllerName, defaultKedaControllerNamespace)
}

func buildInvalidControllerBuilder(apiClient *clients.Settings) *ControllerBuilder {
	return NewControllerBuilder(apiClient, "", defaultKedaControllerNamespace)
}

func buildControllerClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyController(),
		SchemeAttachers: []clients.SchemeAttacher{kedav1alpha1.AddToScheme},
	})
}

func buildDummyController() []runtime.Object {
	return []runtime.Object{
		&kedav1alpha1.KedaController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultKedaControllerName,
				Namespace: defaultKedaControllerNamespace,
			},
		},
	}
}
