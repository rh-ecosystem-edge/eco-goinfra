package servicemesh

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	istiov2 "maistra.io/api/core/v2"
)

var (
	defaultControlPlaneName = "basic"
	dummyIPAddress          = "10.22.22.1"
	emptyIPAddress          = ""
	isTrue                  = true
	isFalse                 = false
)

func TestPullControlPlane(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullControlPlane,
		istiov2.AddToScheme,
		controlPlaneGVK,
	).ExecuteTests(t)
}

func TestNewControlPlaneBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewControlPlaneBuilder,
		istiov2.AddToScheme,
		controlPlaneGVK,
	).ExecuteTests(t)
}

func TestControlPlaneBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[istiov2.ServiceMeshControlPlane, ControlPlaneBuilder](
		istiov2.AddToScheme,
		controlPlaneGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestControlPlaneWithAllAddonsDisabled(t *testing.T) {
	t.Parallel()

	t.Run("disables all addons", func(t *testing.T) {
		t.Parallel()

		builder := buildControlPlaneBuilderForTest(true)
		result := builder.WithAllAddonsDisabled()

		require.NoError(t, result.GetError())
		require.NotNil(t, result.Definition.Spec.Addons)
		assert.False(t, *result.Definition.Spec.Addons.Prometheus.Enabled)
		assert.False(t, *result.Definition.Spec.Addons.Grafana.Enabled)
		assert.False(t, *result.Definition.Spec.Addons.Kiali.Enabled)
		assert.False(t, *result.Definition.Spec.Addons.ThreeScale.Enabled)
	})

	t.Run("invalid builder short-circuits", func(t *testing.T) {
		t.Parallel()

		builder := buildControlPlaneBuilderForTest(false)
		result := builder.WithAllAddonsDisabled()

		require.Error(t, result.GetError())
		assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
		assert.Nil(t, result.Definition.Spec.Addons)
	})
}

//nolint:funlen // Table-driven test cases keep the addon behavior together.
func TestControlPlaneWithGrafanaAddon(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		enablement     bool
		install        *istiov2.GrafanaInstallConfig
		address        string
		validBuilder   bool
		expectedErrMsg string
	}{
		{
			name:         "enabled addon",
			enablement:   true,
			install:      &istiov2.GrafanaInstallConfig{SelfManaged: false},
			address:      dummyIPAddress,
			validBuilder: true,
		},
		{
			name:         "disabled addon",
			install:      &istiov2.GrafanaInstallConfig{SelfManaged: false},
			address:      dummyIPAddress,
			validBuilder: true,
		},
		{
			name:           "enabled addon without install config",
			enablement:     true,
			address:        dummyIPAddress,
			validBuilder:   true,
			expectedErrMsg: "the Grafana addon 'grafanaInstallConfig' cannot be empty when Grafana addon is enabled",
		},
		{
			name:         "disabled addon without install config",
			address:      dummyIPAddress,
			validBuilder: true,
		},
		{
			name:         "disabled addon without address",
			install:      &istiov2.GrafanaInstallConfig{SelfManaged: false},
			validBuilder: true,
		},
		{
			name:           "enabled addon without address",
			enablement:     true,
			install:        &istiov2.GrafanaInstallConfig{SelfManaged: false},
			validBuilder:   true,
			expectedErrMsg: "the Grafana addon 'address' cannot be empty when Grafana addon is enabled",
		},
		{
			name:         "empty address with disabled addon",
			install:      &istiov2.GrafanaInstallConfig{SelfManaged: false},
			address:      emptyIPAddress,
			validBuilder: true,
		},
		{
			name:         "invalid builder",
			enablement:   true,
			install:      &istiov2.GrafanaInstallConfig{SelfManaged: false},
			address:      dummyIPAddress,
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := buildControlPlaneBuilderForTest(testCase.validBuilder).
				WithGrafanaAddon(testCase.enablement, testCase.install, testCase.address)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Addons)

				return
			}

			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedErrMsg)
				assert.Nil(t, result.Definition.Spec.Addons)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Addons)
			require.NotNil(t, result.Definition.Spec.Addons.Grafana)
			assert.Equal(t, testCase.enablement, *result.Definition.Spec.Addons.Grafana.Enabled)
			assert.Equal(t, testCase.install, result.Definition.Spec.Addons.Grafana.Install)
			assert.Equal(t, testCase.address, *result.Definition.Spec.Addons.Grafana.Address)
		})
	}
}

func TestControlPlaneWithJaegerAddon(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		addonName      string
		install        *istiov2.JaegerInstallConfig
		validBuilder   bool
		expectedErrMsg string
	}{
		{
			name:         "enabled ingress",
			addonName:    "jaegertest",
			install:      &istiov2.JaegerInstallConfig{Ingress: &istiov2.JaegerIngressConfig{Enablement: istiov2.Enablement{Enabled: &isTrue}}},
			validBuilder: true,
		},
		{
			name:         "disabled ingress",
			addonName:    "jaegertest",
			install:      &istiov2.JaegerInstallConfig{Ingress: &istiov2.JaegerIngressConfig{Enablement: istiov2.Enablement{Enabled: &isFalse}}},
			validBuilder: true,
		},
		{
			name:           "empty addon name",
			install:        &istiov2.JaegerInstallConfig{Ingress: &istiov2.JaegerIngressConfig{Enablement: istiov2.Enablement{Enabled: &isTrue}}},
			validBuilder:   true,
			expectedErrMsg: "the Jaeger addon 'name' cannot be empty",
		},
		{
			name:           "empty install config",
			addonName:      "jaegertest",
			validBuilder:   true,
			expectedErrMsg: "the Jaeger addon 'jaegerInstallConfig' cannot be empty",
		},
		{
			name:         "invalid builder",
			addonName:    "jaegertest",
			install:      &istiov2.JaegerInstallConfig{},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := buildControlPlaneBuilderForTest(testCase.validBuilder).
				WithJaegerAddon(testCase.addonName, testCase.install)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Addons)

				return
			}

			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedErrMsg)
				assert.Nil(t, result.Definition.Spec.Addons)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Addons)
			require.NotNil(t, result.Definition.Spec.Addons.Jaeger)
			assert.Equal(t, testCase.addonName, result.Definition.Spec.Addons.Jaeger.Name)
			assert.Equal(t, testCase.install, result.Definition.Spec.Addons.Jaeger.Install)
		})
	}
}

//nolint:funlen // Table-driven test cases keep the addon behavior together.
func TestControlPlaneWithKialiAddon(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		enablement     bool
		addonName      string
		install        *istiov2.KialiInstallConfig
		validBuilder   bool
		expectedErrMsg string
	}{
		{
			name:         "enabled addon",
			enablement:   true,
			addonName:    "testkiali",
			install:      &istiov2.KialiInstallConfig{Dashboard: &istiov2.KialiDashboardConfig{ViewOnly: &isTrue}},
			validBuilder: true,
		},
		{
			name:         "disabled addon",
			addonName:    "testkiali",
			install:      &istiov2.KialiInstallConfig{Dashboard: &istiov2.KialiDashboardConfig{ViewOnly: &isTrue}},
			validBuilder: true,
		},
		{
			name:         "enabled with disabled dashboard features",
			enablement:   true,
			addonName:    "testkiali",
			install:      &istiov2.KialiInstallConfig{Dashboard: &istiov2.KialiDashboardConfig{ViewOnly: &isFalse}},
			validBuilder: true,
		},
		{
			name:         "disabled addon without install config",
			addonName:    "testkiali",
			validBuilder: true,
		},
		{
			name:         "disabled addon without addon name",
			install:      &istiov2.KialiInstallConfig{Dashboard: &istiov2.KialiDashboardConfig{ViewOnly: &isFalse}},
			validBuilder: true,
		},
		{
			name:           "enabled addon without addon name",
			enablement:     true,
			install:        &istiov2.KialiInstallConfig{Dashboard: &istiov2.KialiDashboardConfig{ViewOnly: &isFalse}},
			validBuilder:   true,
			expectedErrMsg: "the Kiali addon 'name' cannot be empty when Kiali addon is enabled",
		},
		{
			name:           "enabled addon without install config",
			enablement:     true,
			addonName:      "testkiali",
			validBuilder:   true,
			expectedErrMsg: "the Kiali addon 'kialiInstallConfig' cannot be empty when Kiali addon is enabled",
		},
		{
			name:         "invalid builder",
			enablement:   true,
			addonName:    "testkiali",
			install:      &istiov2.KialiInstallConfig{},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := buildControlPlaneBuilderForTest(testCase.validBuilder).
				WithKialiAddon(testCase.enablement, testCase.addonName, testCase.install)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Addons)

				return
			}

			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedErrMsg)
				assert.Nil(t, result.Definition.Spec.Addons)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Addons)
			require.NotNil(t, result.Definition.Spec.Addons.Kiali)
			assert.Equal(t, testCase.enablement, *result.Definition.Spec.Addons.Kiali.Enabled)
			assert.Equal(t, testCase.addonName, result.Definition.Spec.Addons.Kiali.Name)
			assert.Equal(t, testCase.install, result.Definition.Spec.Addons.Kiali.Install)
		})
	}
}

//nolint:funlen // Table-driven test cases keep the addon behavior together.
func TestControlPlaneWithPrometheusAddon(t *testing.T) {
	t.Parallel()

	prometheusInstallConfig := &istiov2.PrometheusInstallConfig{
		SelfManaged:    false,
		ScrapeInterval: "5",
		UseTLS:         &isFalse,
	}

	testCases := []struct {
		name                  string
		enablement            bool
		scrape                bool
		metricsExpiryDuration string
		address               string
		install               *istiov2.PrometheusInstallConfig
		validBuilder          bool
		expectedErrMsg        string
	}{
		{
			name:                  "enabled addon",
			enablement:            true,
			scrape:                true,
			metricsExpiryDuration: "100",
			address:               dummyIPAddress,
			install:               prometheusInstallConfig,
			validBuilder:          true,
		},
		{
			name:                  "disabled addon",
			scrape:                true,
			metricsExpiryDuration: "100",
			address:               dummyIPAddress,
			install:               prometheusInstallConfig,
			validBuilder:          true,
		},
		{
			name:                  "scrape disabled",
			enablement:            true,
			metricsExpiryDuration: "100",
			address:               dummyIPAddress,
			install:               prometheusInstallConfig,
			validBuilder:          true,
		},
		{
			name:         "disabled addon with empty expiry duration",
			scrape:       true,
			address:      dummyIPAddress,
			install:      prometheusInstallConfig,
			validBuilder: true,
		},
		{
			name:                  "disabled addon with empty address",
			scrape:                true,
			metricsExpiryDuration: "100",
			install:               prometheusInstallConfig,
			validBuilder:          true,
		},
		{
			name:                  "disabled addon without install config",
			scrape:                true,
			metricsExpiryDuration: "100",
			address:               dummyIPAddress,
			validBuilder:          true,
		},
		{
			name:           "enabled addon without expiry duration",
			enablement:     true,
			scrape:         true,
			address:        dummyIPAddress,
			install:        prometheusInstallConfig,
			validBuilder:   true,
			expectedErrMsg: "the Prometheus addon 'metricsExpiryDuration' cannot be empty when Prometheus addon is enabled",
		},
		{
			name:                  "enabled addon without address",
			enablement:            true,
			scrape:                true,
			metricsExpiryDuration: "100",
			install:               prometheusInstallConfig,
			validBuilder:          true,
			expectedErrMsg:        "the Prometheus addon 'address' cannot be empty when Prometheus addon is enabled",
		},
		{
			name:                  "enabled addon without install config",
			enablement:            true,
			scrape:                true,
			metricsExpiryDuration: "100",
			address:               dummyIPAddress,
			validBuilder:          true,
			expectedErrMsg: "the Prometheus addon 'prometheusInstallConfig' cannot be empty " +
				"when Prometheus addon is enabled",
		},
		{
			name:                  "invalid builder",
			enablement:            true,
			scrape:                true,
			metricsExpiryDuration: "100",
			address:               dummyIPAddress,
			install:               prometheusInstallConfig,
			validBuilder:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := buildControlPlaneBuilderForTest(testCase.validBuilder).WithPrometheusAddon(
				testCase.enablement,
				testCase.scrape,
				testCase.metricsExpiryDuration,
				testCase.address,
				testCase.install,
			)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Addons)

				return
			}

			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedErrMsg)
				assert.Nil(t, result.Definition.Spec.Addons)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Addons)
			require.NotNil(t, result.Definition.Spec.Addons.Prometheus)
			assert.Equal(t, testCase.enablement, *result.Definition.Spec.Addons.Prometheus.Enabled)
			assert.Equal(t, testCase.scrape, *result.Definition.Spec.Addons.Prometheus.Scrape)
			assert.Equal(t, testCase.metricsExpiryDuration,
				result.Definition.Spec.Addons.Prometheus.MetricsExpiryDuration)
			assert.Equal(t, testCase.address, *result.Definition.Spec.Addons.Prometheus.Address)
			assert.Equal(t, testCase.install, result.Definition.Spec.Addons.Prometheus.Install)
		})
	}
}

func TestControlPlaneWithGatewaysEnablement(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		enablement   bool
		validBuilder bool
	}{
		{name: "enabled gateways", enablement: true, validBuilder: true},
		{name: "disabled gateways", enablement: false, validBuilder: true},
		{name: "invalid builder", enablement: true, validBuilder: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := buildControlPlaneBuilderForTest(testCase.validBuilder).
				WithGatewaysEnablement(testCase.enablement)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Gateways)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Gateways)
			assert.Equal(t, testCase.enablement, *result.Definition.Spec.Gateways.Enabled)
		})
	}
}

func buildControlPlaneBuilderForTest(valid bool) *ControlPlaneBuilder {
	if !valid {
		return NewControlPlaneBuilder(
			clients.GetTestClients(clients.TestClientParams{}),
			"",
			defaultServiceMeshNamespace,
		)
	}

	return NewControlPlaneBuilder(
		clients.GetTestClients(clients.TestClientParams{}),
		defaultControlPlaneName,
		defaultServiceMeshNamespace,
	)
}
