package monitoring

import (
	"fmt"
	"testing"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

var (
	defaultServiceMonitorName      = "test-monitor-name"
	defaultServiceMonitorNamespace = "test-monitor-namespace"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewBuilder, monv1.AddToScheme, serviceMonitorGVK).ExecuteTests(t)
}

func TestPullServiceMonitor(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull, monv1.AddToScheme, serviceMonitorGVK).ExecuteTests(t)
}

func TestServiceMonitorBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[monv1.ServiceMonitor, Builder](
		monv1.AddToScheme,
		serviceMonitorGVK,
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

func TestServiceMonitorWithEndpoints(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		endpoints     []monv1.Endpoint
		valid         bool
		expectedError error
	}{
		{
			name: "single endpoint",
			endpoints: []monv1.Endpoint{{
				Port:   "http",
				Scheme: ptr.To(monv1.Scheme("http")),
			}},
			valid: true,
		},
		{
			name: "multiple endpoints",
			endpoints: []monv1.Endpoint{
				{
					Port:   "http",
					Scheme: ptr.To(monv1.Scheme("http")),
				},
				{
					Port:   "sctp",
					Scheme: ptr.To(monv1.Scheme("sctp")),
				},
			},
			valid: true,
		},
		{
			name:          "empty endpoints",
			endpoints:     []monv1.Endpoint{},
			valid:         true,
			expectedError: fmt.Errorf("'endpoints' argument cannot be empty"),
		},
		{
			name: "invalid builder",
			endpoints: []monv1.Endpoint{{
				Port:   "http",
				Scheme: ptr.To(monv1.Scheme("http")),
			}},
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testBuilder *Builder
			if testCase.valid {
				testBuilder = buildValidServiceMonitorTestBuilder(buildServiceMonitorTestClient())
			} else {
				testBuilder = buildInvalidServiceMonitorTestBuilder(buildServiceMonitorTestClient())
			}

			result := testBuilder.WithEndpoints(testCase.endpoints)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.Endpoints)

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.endpoints, result.Definition.Spec.Endpoints)
			}
		})
	}
}

func TestServiceMonitorWithLabels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		labels        map[string]string
		valid         bool
		expectedError error
	}{
		{
			name:   "single label",
			labels: map[string]string{"test-label-key": "test-label-value"},
			valid:  true,
		},
		{
			name: "multiple labels",
			labels: map[string]string{
				"test-label1-key": "test-label1-value",
				"test-label2-key": "test-label2-value",
				"test-label3-key": "test-label3-value",
			},
			valid: true,
		},
		{
			name:   "empty label value",
			labels: map[string]string{"test-label-key": ""},
			valid:  true,
		},
		{
			name:          "empty label key",
			labels:        map[string]string{"": "test-label-value"},
			valid:         true,
			expectedError: fmt.Errorf("can not apply a labels with an empty key"),
		},
		{
			name:          "empty labels",
			labels:        map[string]string{},
			valid:         true,
			expectedError: fmt.Errorf(errEmptyLabels),
		},
		{
			name:   "invalid builder",
			labels: map[string]string{"test-label-key": "test-label-value"},
			valid:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testBuilder *Builder
			if testCase.valid {
				testBuilder = buildValidServiceMonitorTestBuilder(buildServiceMonitorTestClient())
			} else {
				testBuilder = buildInvalidServiceMonitorTestBuilder(buildServiceMonitorTestClient())
			}

			result := testBuilder.WithLabels(testCase.labels)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Labels)

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.labels, result.Definition.Labels)
			}
		})
	}
}

func TestServiceMonitorWithSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		selector      map[string]string
		valid         bool
		expectedError error
	}{
		{
			name:     "selector with value",
			selector: map[string]string{"test-selector-key": "test-selector-value"},
			valid:    true,
		},
		{
			name:     "selector with empty value",
			selector: map[string]string{"test-selector-key": ""},
			valid:    true,
		},
		{
			name:          "empty selector key",
			selector:      map[string]string{"": "test-selector-value"},
			valid:         true,
			expectedError: fmt.Errorf("can not apply a selector with an empty key"),
		},
		{
			name:          "empty selector",
			selector:      map[string]string{},
			valid:         true,
			expectedError: fmt.Errorf("selector can not be empty"),
		},
		{
			name:     "invalid builder",
			selector: map[string]string{"test-selector-key": "test-selector-value"},
			valid:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testBuilder *Builder
			if testCase.valid {
				testBuilder = buildValidServiceMonitorTestBuilder(buildServiceMonitorTestClient())
			} else {
				testBuilder = buildInvalidServiceMonitorTestBuilder(buildServiceMonitorTestClient())
			}

			result := testBuilder.WithSelector(testCase.selector)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Selector.MatchLabels)

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.selector, result.Definition.Spec.Selector.MatchLabels)
			}
		})
	}
}

func TestServiceMonitorWithNamespaceSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		namespaceSelector []string
		valid             bool
		expectedError     error
	}{
		{
			name:              "single namespace",
			namespaceSelector: []string{"test-ns-selector"},
			valid:             true,
		},
		{
			name:              "multiple namespaces",
			namespaceSelector: []string{"test-ns-selector1", "test-ns-selector2"},
			valid:             true,
		},
		{
			name:              "empty namespace selector",
			namespaceSelector: []string{},
			valid:             true,
			expectedError:     fmt.Errorf("namespaceSelector can not be empty"),
		},
		{
			name:              "invalid builder",
			namespaceSelector: []string{"test-ns-selector"},
			valid:             false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testBuilder *Builder
			if testCase.valid {
				testBuilder = buildValidServiceMonitorTestBuilder(buildServiceMonitorTestClient())
			} else {
				testBuilder = buildInvalidServiceMonitorTestBuilder(buildServiceMonitorTestClient())
			}

			result := testBuilder.WithNamespaceSelector(testCase.namespaceSelector)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.NamespaceSelector.MatchNames)

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.namespaceSelector, result.Definition.Spec.NamespaceSelector.MatchNames)
			}
		})
	}
}

func buildValidServiceMonitorTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultServiceMonitorName, defaultServiceMonitorNamespace)
}

func buildInvalidServiceMonitorTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "", defaultServiceMonitorNamespace)
}

func buildServiceMonitorTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{monv1.AddToScheme},
	})
}
