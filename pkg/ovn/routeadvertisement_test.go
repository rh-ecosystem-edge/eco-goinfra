package ovn

import (
	"fmt"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

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
	defaultFRRConfigurationSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{"frr": "config"},
	}
	defaultNetworkSelectors = types.NetworkSelectors{
		{
			NetworkSelectionType: types.DefaultNetwork,
		},
	}
)

func TestNewRouteAdvertisementBuilder(t *testing.T) {
	testCases := []struct {
		name                     string
		advertisements           []ovnv1.AdvertisementType
		nodeSelector             metav1.LabelSelector
		frrConfigurationSelector metav1.LabelSelector
		networkSelectors         types.NetworkSelectors
		expectedErrorText        string
	}{
		{
			name:                     defaultRouteAdvertisementName,
			advertisements:           defaultAdvertisements,
			nodeSelector:             defaultNodeSelector,
			frrConfigurationSelector: defaultFRRConfigurationSelector,
			networkSelectors:         defaultNetworkSelectors,
			expectedErrorText:        "",
		},
		{
			name:                     "",
			advertisements:           defaultAdvertisements,
			nodeSelector:             defaultNodeSelector,
			frrConfigurationSelector: defaultFRRConfigurationSelector,
			networkSelectors:         defaultNetworkSelectors,
			expectedErrorText:        "RouteAdvertisement 'name' cannot be empty",
		},
		{
			name:                     defaultRouteAdvertisementName,
			advertisements:           []ovnv1.AdvertisementType{},
			nodeSelector:             defaultNodeSelector,
			frrConfigurationSelector: defaultFRRConfigurationSelector,
			networkSelectors:         defaultNetworkSelectors,
			expectedErrorText:        "RouteAdvertisement 'advertisements' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testRouteAdvertisementBuilder := NewRouteAdvertisementBuilder(
			testSettings.Client,
			testCase.name,
			testCase.advertisements,
			testCase.nodeSelector,
			testCase.frrConfigurationSelector,
			testCase.networkSelectors)

		if testCase.expectedErrorText == "" {
			if testRouteAdvertisementBuilder == nil {
				t.Errorf("Unexpected nil RouteAdvertisementBuilder")
			}

			if testRouteAdvertisementBuilder.errorMsg != "" {
				t.Errorf("Unexpected error message %s", testRouteAdvertisementBuilder.errorMsg)
			}

			if testRouteAdvertisementBuilder.Definition.Name != testCase.name {
				t.Errorf("Expected RouteAdvertisement name %s, got %s",
					testCase.name, testRouteAdvertisementBuilder.Definition.Name)
			}
		} else {
			if testRouteAdvertisementBuilder == nil {
				t.Errorf("Expected RouteAdvertisementBuilder, got nil")
			}

			if testRouteAdvertisementBuilder.errorMsg != testCase.expectedErrorText {
				t.Errorf("Expected error message %s, got %s",
					testCase.expectedErrorText, testRouteAdvertisementBuilder.errorMsg)
			}
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

		routeAdvertisementBuilder, err := PullRouteAdvertisement(testSettings.Client, testCase.routeAdvertisement.Name)

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

		routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings.Client)

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

		routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings.Client)
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

		routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings.Client)
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
	testSettings := clients.GetTestClients(clients.TestClientParams{})
	routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings.Client)

	targetVRF := "test-vrf"
	routeAdvertisementBuilder.WithTargetVRF(targetVRF)

	assert.Equal(t, targetVRF, routeAdvertisementBuilder.Definition.Spec.TargetVRF)
}

func TestRouteAdvertisementWithAdvertisements(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{})
	routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings.Client)

	advertisements := []ovnv1.AdvertisementType{ovnv1.PodNetwork, ovnv1.EgressIP}
	routeAdvertisementBuilder.WithAdvertisements(advertisements)

	assert.Equal(t, advertisements, routeAdvertisementBuilder.Definition.Spec.Advertisements)
}

func TestRouteAdvertisementWithNodeSelector(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{})
	routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings.Client)

	nodeSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{"node": "test"},
	}
	routeAdvertisementBuilder.WithNodeSelector(nodeSelector)

	assert.Equal(t, nodeSelector, routeAdvertisementBuilder.Definition.Spec.NodeSelector)
}

func TestRouteAdvertisementWithFRRConfigurationSelector(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{})
	routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings.Client)

	frrConfigurationSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{"frr": "test"},
	}
	routeAdvertisementBuilder.WithFRRConfigurationSelector(frrConfigurationSelector)

	assert.Equal(t, frrConfigurationSelector, routeAdvertisementBuilder.Definition.Spec.FRRConfigurationSelector)
}

func TestRouteAdvertisementWithNetworkSelectors(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{})
	routeAdvertisementBuilder := buildTestRouteAdvertisementBuilder(testSettings.Client)

	networkSelectors := types.NetworkSelectors{
		{
			NetworkSelectionType: types.ClusterUserDefinedNetworks,
			ClusterUserDefinedNetworkSelector: &types.ClusterUserDefinedNetworkSelector{
				NetworkSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{"network": "test"},
				},
			},
		},
	}
	routeAdvertisementBuilder.WithNetworkSelectors(networkSelectors)

	assert.Equal(t, networkSelectors, routeAdvertisementBuilder.Definition.Spec.NetworkSelectors)
}

func buildDummyRouteAdvertisement(name string) *ovnv1.RouteAdvertisements {
	return &ovnv1.RouteAdvertisements{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: ovnv1.RouteAdvertisementsSpec{
			Advertisements:           defaultAdvertisements,
			NodeSelector:             defaultNodeSelector,
			FRRConfigurationSelector: defaultFRRConfigurationSelector,
			NetworkSelectors:         defaultNetworkSelectors,
		},
	}
}

func buildTestRouteAdvertisementBuilder(apiClient client.Client) *RouteAdvertisementBuilder {
	return NewRouteAdvertisementBuilder(
		apiClient,
		defaultRouteAdvertisementName,
		defaultAdvertisements,
		defaultNodeSelector,
		defaultFRRConfigurationSelector,
		defaultNetworkSelectors,
	)
}
