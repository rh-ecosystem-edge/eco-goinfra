package ovn

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultRouteAdvertisementName = "test-routeadvertisement"
	defaultAdvertisements         = []ovnv1.AdvertisementType{ovnv1.PodNetwork}
	defaultNodeSelector           = metav1.LabelSelector{
		MatchLabels: map[string]string{"test": "label"},
	}
	defaultFrrConfigurationSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{"frr": "config"},
	}
	defaultNetworkSelectors = types.NetworkSelectors{
		{
			NetworkSelectionType: types.DefaultNetwork,
		},
	}
	routeAdvertisementGVK = ovnv1.SchemeGroupVersion.WithKind("RouteAdvertisements")
)

func TestNewRouteAdvertisementBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		NewRouteAdvertisementBuilder,
		ovnv1.AddToScheme,
		routeAdvertisementGVK,
	).ExecuteTests(t)
}

func TestPullRouteAdvertisement(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullRouteAdvertisement,
		ovnv1.AddToScheme,
		routeAdvertisementGVK,
	).ExecuteTests(t)
}

func TestRouteAdvertisementBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[ovnv1.RouteAdvertisements, RouteAdvertisementBuilder](
		ovnv1.AddToScheme,
		routeAdvertisementGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestRouteAdvertisementWithTargetVRF(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		invalidBuilder bool
		targetVRF      string
	}{
		{
			name:      "sets target VRF",
			targetVRF: "test-vrf",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
			targetVRF:      "test-vrf",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildTestRouteAdvertisementBuilder(getTestRouteAdvertisementAPIClient())
			if testCase.invalidBuilder {
				builder = buildInvalidRouteAdvertisementTestBuilder(getTestRouteAdvertisementAPIClient())
			}

			builder.WithTargetVRF(testCase.targetVRF)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.targetVRF, builder.Definition.Spec.TargetVRF)
		})
	}
}

func TestRouteAdvertisementWithAdvertisements(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		invalidBuilder bool
		advertisements []ovnv1.AdvertisementType
		expectedError  string
	}{
		{
			name:           "sets advertisements",
			advertisements: []ovnv1.AdvertisementType{ovnv1.PodNetwork},
		},
		{
			name:           "empty advertisements",
			advertisements: []ovnv1.AdvertisementType{},
			expectedError:  "RouteAdvertisement 'advertisements' cannot be empty",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
			advertisements: []ovnv1.AdvertisementType{ovnv1.PodNetwork},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildTestRouteAdvertisementBuilder(getTestRouteAdvertisementAPIClient())
			if testCase.invalidBuilder {
				builder = buildInvalidRouteAdvertisementTestBuilder(getTestRouteAdvertisementAPIClient())
			}

			builder.WithAdvertisements(testCase.advertisements)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.advertisements, builder.Definition.Spec.Advertisements)
		})
	}
}

func TestRouteAdvertisementWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		invalidBuilder bool
		nodeSelector   metav1.LabelSelector
	}{
		{
			name:         "sets node selector",
			nodeSelector: metav1.LabelSelector{MatchLabels: map[string]string{"node": "test"}},
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
			nodeSelector:   metav1.LabelSelector{MatchLabels: map[string]string{"node": "test"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildTestRouteAdvertisementBuilder(getTestRouteAdvertisementAPIClient())
			if testCase.invalidBuilder {
				builder = buildInvalidRouteAdvertisementTestBuilder(getTestRouteAdvertisementAPIClient())
			}

			builder.WithNodeSelector(testCase.nodeSelector)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.nodeSelector, builder.Definition.Spec.NodeSelector)
		})
	}
}

func TestRouteAdvertisementWithFRRConfigurationSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                     string
		invalidBuilder           bool
		frrConfigurationSelector metav1.LabelSelector
	}{
		{
			name:                     "sets FRR configuration selector",
			frrConfigurationSelector: metav1.LabelSelector{MatchLabels: map[string]string{"frr": "test"}},
		},
		{
			name:                     "invalid builder",
			invalidBuilder:           true,
			frrConfigurationSelector: metav1.LabelSelector{MatchLabels: map[string]string{"frr": "test"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildTestRouteAdvertisementBuilder(getTestRouteAdvertisementAPIClient())
			if testCase.invalidBuilder {
				builder = buildInvalidRouteAdvertisementTestBuilder(getTestRouteAdvertisementAPIClient())
			}

			builder.WithFRRConfigurationSelector(testCase.frrConfigurationSelector)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.frrConfigurationSelector, builder.Definition.Spec.FRRConfigurationSelector)
		})
	}
}

func TestRouteAdvertisementWithNetworkSelectors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		invalidBuilder   bool
		networkSelectors types.NetworkSelectors
		expectedError    string
	}{
		{
			name: "sets network selectors",
			networkSelectors: types.NetworkSelectors{
				{
					NetworkSelectionType: types.ClusterUserDefinedNetworks,
					ClusterUserDefinedNetworkSelector: &types.ClusterUserDefinedNetworkSelector{
						NetworkSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{"network": "test"},
						},
					},
				},
			},
		},
		{
			name:             "empty network selectors",
			networkSelectors: types.NetworkSelectors{},
			expectedError:    "RouteAdvertisement 'networkSelectors' cannot be empty",
		},
		{
			name: "empty network selection type",
			networkSelectors: types.NetworkSelectors{
				{
					NetworkSelectionType: "",
				},
			},
			expectedError: "RouteAdvertisement 'networkSelectors' must have valid NetworkSelectionType",
		},
		{
			name: "invalid network selection type",
			networkSelectors: types.NetworkSelectors{
				{
					NetworkSelectionType: "InvalidNetwork",
				},
			},
			expectedError: "RouteAdvertisement 'networkSelectors[0]' has invalid NetworkSelectionType: InvalidNetwork",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
			networkSelectors: types.NetworkSelectors{
				{
					NetworkSelectionType: types.DefaultNetwork,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildTestRouteAdvertisementBuilder(getTestRouteAdvertisementAPIClient())
			if testCase.invalidBuilder {
				builder = buildInvalidRouteAdvertisementTestBuilder(getTestRouteAdvertisementAPIClient())
			}

			builder.WithNetworkSelectors(testCase.networkSelectors)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.networkSelectors, builder.Definition.Spec.NetworkSelectors)
		})
	}
}

func buildDummyRouteAdvertisement(name string) *ovnv1.RouteAdvertisements {
	return &ovnv1.RouteAdvertisements{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: ovnv1.RouteAdvertisementsSpec{
			Advertisements:           defaultAdvertisements,
			NodeSelector:             defaultNodeSelector,
			FRRConfigurationSelector: defaultFrrConfigurationSelector,
			NetworkSelectors:         defaultNetworkSelectors,
		},
	}
}

func buildTestRouteAdvertisementBuilder(apiClient *clients.Settings) *RouteAdvertisementBuilder {
	return NewRouteAdvertisementBuilder(apiClient, defaultRouteAdvertisementName).
		WithAdvertisements(defaultAdvertisements).
		WithNodeSelector(defaultNodeSelector).
		WithFRRConfigurationSelector(defaultFrrConfigurationSelector).
		WithNetworkSelectors(defaultNetworkSelectors)
}

func buildInvalidRouteAdvertisementTestBuilder(apiClient *clients.Settings) *RouteAdvertisementBuilder {
	return NewRouteAdvertisementBuilder(apiClient, "")
}

func getTestRouteAdvertisementAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}
