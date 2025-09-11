package ovn

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/types"
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
	routeAdvertisementTestSchemes = []clients.SchemeAttacher{
		ovnv1.AddToScheme,
	}
)

func TestNewRouteAdvertisementBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		expectedErrorText string
	}{
		{
			name:              defaultRouteAdvertisementName,
			expectedErrorText: "",
		},
		{
			name:              "",
			expectedErrorText: "RouteAdvertisement 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: routeAdvertisementTestSchemes})
		testRouteAdvertisementBuilder := NewRouteAdvertisementBuilder(testSettings, testCase.name)

		assert.Equal(t, testCase.expectedErrorText, testRouteAdvertisementBuilder.errorMsg)
		assert.NotNil(t, testRouteAdvertisementBuilder.Definition)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.name, testRouteAdvertisementBuilder.Definition.Name)
		}
	}
}

func TestRouteAdvertisementGet(t *testing.T) {
	testCases := []struct {
		routeAdvertisement *ovnv1.RouteAdvertisements
		expectedError      bool
	}{
		{
			routeAdvertisement: buildDummyRouteAdvertisement(defaultRouteAdvertisementName),
			expectedError:      false,
		},
		{
			routeAdvertisement: buildDummyRouteAdvertisement(""),
			expectedError:      true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.routeAdvertisement != nil {
			runtimeObjects = append(runtimeObjects, testCase.routeAdvertisement)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
			SchemeAttachers: []clients.SchemeAttacher{
				ovnv1.AddToScheme,
			},
		})

		routeAdvertisementBuilder, err := PullRouteAdvertisement(testSettings, testCase.routeAdvertisement.Name)

		if testCase.expectedError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.routeAdvertisement.Name, routeAdvertisementBuilder.Definition.Name)
		}
	}
}

func TestRouteAdvertisementExists(t *testing.T) {
	testCases := []struct {
		testRouteAdvertisement *ovnv1.RouteAdvertisements
		expectedStatus         bool
	}{
		{
			testRouteAdvertisement: buildDummyRouteAdvertisement(defaultRouteAdvertisementName),
			expectedStatus:         true,
		},
		{
			testRouteAdvertisement: buildDummyRouteAdvertisement(""),
			expectedStatus:         false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.testRouteAdvertisement != nil {
			runtimeObjects = append(runtimeObjects, testCase.testRouteAdvertisement)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
			SchemeAttachers: []clients.SchemeAttacher{
				ovnv1.AddToScheme,
			},
		})

		routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings)

		if testCase.testRouteAdvertisement != nil {
			routeAdvertisementBuilder.Definition.Name = testCase.testRouteAdvertisement.Name
		}

		assert.Equal(t, testCase.expectedStatus, routeAdvertisementBuilder.Exists())
	}
}

func TestRouteAdvertisementCreate(t *testing.T) {
	testCases := []struct {
		testRouteAdvertisement *ovnv1.RouteAdvertisements
		expectedError          error
	}{
		{
			testRouteAdvertisement: buildDummyRouteAdvertisement(defaultRouteAdvertisementName),
			expectedError:          nil,
		},
		{
			testRouteAdvertisement: buildDummyRouteAdvertisement(""),
			expectedError:          fmt.Errorf("RouteAdvertisement 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: []clients.SchemeAttacher{
				ovnv1.AddToScheme,
			},
		})

		routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings)
		routeAdvertisementBuilder.Definition = testCase.testRouteAdvertisement

		result, err := routeAdvertisementBuilder.Create()

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testRouteAdvertisement.Name, result.Definition.Name)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestRouteAdvertisementDelete(t *testing.T) {
	testCases := []struct {
		testRouteAdvertisement *ovnv1.RouteAdvertisements
		expectedError          error
	}{
		{
			testRouteAdvertisement: buildDummyRouteAdvertisement(defaultRouteAdvertisementName),
			expectedError:          nil,
		},
		{
			testRouteAdvertisement: buildDummyRouteAdvertisement(""),
			expectedError:          fmt.Errorf("RouteAdvertisement 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.testRouteAdvertisement != nil {
			runtimeObjects = append(runtimeObjects, testCase.testRouteAdvertisement)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
			SchemeAttachers: []clients.SchemeAttacher{
				ovnv1.AddToScheme,
			},
		})

		routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings)
		routeAdvertisementBuilder.Definition = testCase.testRouteAdvertisement

		err := routeAdvertisementBuilder.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.Nil(t, routeAdvertisementBuilder.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestRouteAdvertisementWithTargetVRF(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: routeAdvertisementTestSchemes})
	routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings)

	targetVRF := "test-vrf"
	routeAdvertisementBuilder.WithTargetVRF(targetVRF)

	assert.Equal(t, targetVRF, routeAdvertisementBuilder.Definition.Spec.TargetVRF)
}

func TestRouteAdvertisementWithAdvertisements(t *testing.T) {
	testCases := []struct {
		advertisements []ovnv1.AdvertisementType
		expectedError  string
	}{
		{
			advertisements: []ovnv1.AdvertisementType{ovnv1.PodNetwork},
			expectedError:  "",
		},
		{
			advertisements: []ovnv1.AdvertisementType{},
			expectedError:  "RouteAdvertisement 'advertisements' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: routeAdvertisementTestSchemes})
		routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings)

		routeAdvertisementBuilder.WithAdvertisements(testCase.advertisements)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.advertisements, routeAdvertisementBuilder.Definition.Spec.Advertisements)
			assert.Equal(t, "", routeAdvertisementBuilder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError, routeAdvertisementBuilder.errorMsg)
		}
	}
}

func TestRouteAdvertisementWithNodeSelector(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: routeAdvertisementTestSchemes})
	routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings)

	nodeSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{"node": "test"},
	}
	routeAdvertisementBuilder.WithNodeSelector(nodeSelector)

	assert.Equal(t, nodeSelector, routeAdvertisementBuilder.Definition.Spec.NodeSelector)
}

func TestRouteAdvertisementWithFRRConfigurationSelector(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: routeAdvertisementTestSchemes})
	routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings)

	frrConfigurationSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{"frr": "test"},
	}
	routeAdvertisementBuilder.WithFRRConfigurationSelector(frrConfigurationSelector)

	assert.Equal(t, frrConfigurationSelector, routeAdvertisementBuilder.Definition.Spec.FrrConfigurationSelector)
}

func TestRouteAdvertisementWithNetworkSelectors(t *testing.T) {
	testCases := []struct {
		networkSelectors types.NetworkSelectors
		expectedError    string
	}{
		{
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
			expectedError: "",
		},
		{
			networkSelectors: types.NetworkSelectors{},
			expectedError:    "RouteAdvertisement 'networkSelectors' cannot be empty",
		},
		{
			networkSelectors: types.NetworkSelectors{
				{
					NetworkSelectionType: "",
				},
			},
			expectedError: "RouteAdvertisement 'networkSelectors' must have valid NetworkSelectionType",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: routeAdvertisementTestSchemes})
		routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings)

		routeAdvertisementBuilder.WithNetworkSelectors(testCase.networkSelectors)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.networkSelectors, routeAdvertisementBuilder.Definition.Spec.NetworkSelectors)
			assert.Equal(t, "", routeAdvertisementBuilder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError, routeAdvertisementBuilder.errorMsg)
		}
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
			FrrConfigurationSelector: defaultFrrConfigurationSelector,
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
