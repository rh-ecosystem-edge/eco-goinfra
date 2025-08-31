package ovn

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultRouteAdvertisementName      = "test-routeadvertisement"
	defaultRouteAdvertisementNamespace = "test-namespace"
	defaultAdvertisements              = []ovnv1.AdvertisementType{ovnv1.PodNetworkAdvertisement}
)

func TestNewRouteAdvertisementBuilder(t *testing.T) {
	testCases := []struct {
		name           string
		routeName      string
		namespace      string
		advertisements []ovnv1.AdvertisementType
		expectedErrMsg string
		client         bool
	}{
		{
			name:           "valid RouteAdvertisement",
			routeName:      defaultRouteAdvertisementName,
			namespace:      defaultRouteAdvertisementNamespace,
			advertisements: defaultAdvertisements,
			expectedErrMsg: "",
			client:         true,
		},
		{
			name:           "empty RouteAdvertisement name",
			routeName:      "",
			namespace:      defaultRouteAdvertisementNamespace,
			advertisements: defaultAdvertisements,
			expectedErrMsg: "RouteAdvertisement 'name' cannot be empty",
			client:         true,
		},
		{
			name:           "empty RouteAdvertisement namespace",
			routeName:      defaultRouteAdvertisementName,
			namespace:      "",
			advertisements: defaultAdvertisements,
			expectedErrMsg: "RouteAdvertisement 'namespace' cannot be empty",
			client:         true,
		},
		{
			name:           "empty advertisements",
			routeName:      defaultRouteAdvertisementName,
			namespace:      defaultRouteAdvertisementNamespace,
			advertisements: []ovnv1.AdvertisementType{},
			expectedErrMsg: "RouteAdvertisement 'advertisements' cannot be empty",
			client:         true,
		},
		{
			name:      "too many advertisements",
			routeName: defaultRouteAdvertisementName,
			namespace: defaultRouteAdvertisementNamespace,
			advertisements: []ovnv1.AdvertisementType{
				ovnv1.PodNetworkAdvertisement,
				ovnv1.EgressIPAdvertisement,
				ovnv1.PodNetworkAdvertisement, // This makes it 3 items
			},
			expectedErrMsg: "RouteAdvertisement 'advertisements' cannot have more than 2 items",
			client:         true,
		},
		{
			name:      "duplicate advertisements",
			routeName: defaultRouteAdvertisementName,
			namespace: defaultRouteAdvertisementNamespace,
			advertisements: []ovnv1.AdvertisementType{
				ovnv1.PodNetworkAdvertisement,
				ovnv1.PodNetworkAdvertisement, // Duplicate
			},
			expectedErrMsg: "RouteAdvertisement 'advertisements' cannot contain duplicates: PodNetwork",
			client:         true,
		},
		{
			name:           "nil client",
			routeName:      defaultRouteAdvertisementName,
			namespace:      defaultRouteAdvertisementNamespace,
			advertisements: defaultAdvertisements,
			expectedErrMsg: "",
			client:         false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var testSettings *clients.Settings

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects: buildDummyRouteAdvertisement(),
				})
			}

			builder := NewRouteAdvertisementBuilder(testSettings, testCase.routeName, testCase.namespace, testCase.advertisements)

			if testCase.client {
				assert.NotNil(t, builder)

				if testCase.expectedErrMsg != "" {
					assert.Equal(t, testCase.expectedErrMsg, builder.errorMsg)
				}
			} else {
				assert.Nil(t, builder)
			}
		})
	}
}

func TestRouteAdvertisementGet(t *testing.T) {
	testCases := []struct {
		name                string
		routeAdvertisement  *ovnv1.RouteAdvertisement
		expectedError       bool
		addToRuntimeObjects bool
	}{
		{
			name:                "get existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedError:       false,
			addToRuntimeObjects: true,
		},
		{
			name:                "get non-existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedError:       true,
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects, testCase.routeAdvertisement)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})

			builder := buildValidRouteAdvertisementBuilder(testSettings)
			obj, err := builder.Get()

			if testCase.expectedError {
				assert.NotNil(t, err)
				assert.Nil(t, obj)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, obj)
			}
		})
	}
}

func TestRouteAdvertisementExists(t *testing.T) {
	testCases := []struct {
		name                string
		routeAdvertisement  *ovnv1.RouteAdvertisement
		expectedStatus      bool
		addToRuntimeObjects bool
	}{
		{
			name:                "existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedStatus:      true,
			addToRuntimeObjects: true,
		},
		{
			name:                "non-existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedStatus:      false,
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects, testCase.routeAdvertisement)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})

			builder := buildValidRouteAdvertisementBuilder(testSettings)
			exists := builder.Exists()

			assert.Equal(t, testCase.expectedStatus, exists)
		})
	}
}

func TestRouteAdvertisementCreate(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError error
	}{
		{
			name:          "create RouteAdvertisement",
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testSettings := clients.GetTestClients(clients.TestClientParams{})

			builder := buildValidRouteAdvertisementBuilder(testSettings)
			result, err := builder.Create()

			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestRouteAdvertisementDelete(t *testing.T) {
	testCases := []struct {
		name                string
		routeAdvertisement  *ovnv1.RouteAdvertisement
		expectedError       bool
		addToRuntimeObjects bool
	}{
		{
			name:                "delete existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedError:       false,
			addToRuntimeObjects: true,
		},
		{
			name:                "delete non-existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedError:       false,
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects, testCase.routeAdvertisement)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})

			builder := buildValidRouteAdvertisementBuilder(testSettings)
			_, err := builder.Delete()

			if testCase.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRouteAdvertisementUpdate(t *testing.T) {
	testCases := []struct {
		name                string
		routeAdvertisement  *ovnv1.RouteAdvertisement
		expectedError       bool
		addToRuntimeObjects bool
		force               bool
	}{
		{
			name:                "update existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedError:       false,
			addToRuntimeObjects: true,
			force:               false,
		},
		{
			name:                "update non-existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedError:       true,
			addToRuntimeObjects: false,
			force:               false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects, testCase.routeAdvertisement)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})

			builder := buildValidRouteAdvertisementBuilder(testSettings)
			_, err := builder.Update(testCase.force)

			if testCase.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRouteAdvertisementWithFRRConfigurationSelector(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{})

	builder := buildValidRouteAdvertisementBuilder(testSettings)

	selector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"test": "label",
		},
	}

	builder = builder.WithFRRConfigurationSelector(selector)

	assert.NotNil(t, builder)
	assert.Equal(t, selector, builder.Definition.Spec.FRRConfigurationSelector)
}

func TestRouteAdvertisementWithOptions(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{})

	builder := buildValidRouteAdvertisementBuilder(testSettings)

	// Test with valid option
	builder = builder.WithOptions(func(builder *RouteAdvertisementBuilder) (*RouteAdvertisementBuilder, error) {
		return builder, nil
	})

	assert.NotNil(t, builder)
	assert.Equal(t, "", builder.errorMsg)

	// Test with option that returns error
	builder = builder.WithOptions(func(builder *RouteAdvertisementBuilder) (*RouteAdvertisementBuilder, error) {
		return builder, fmt.Errorf("test error")
	})

	assert.NotNil(t, builder)
	assert.Equal(t, "test error", builder.errorMsg)
}

func TestPull(t *testing.T) {
	testCases := []struct {
		name                string
		routeAdvertisement  *ovnv1.RouteAdvertisement
		expectedError       bool
		addToRuntimeObjects bool
		client              bool
	}{
		{
			name:                "pull existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "pull non-existing RouteAdvertisement",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedError:       true,
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "pull with nil client",
			routeAdvertisement:  buildValidRouteAdvertisement(),
			expectedError:       true,
			addToRuntimeObjects: false,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object
			var testSettings *clients.Settings

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects, testCase.routeAdvertisement)
			}

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects: runtimeObjects,
				})
			}

			builder, err := Pull(testSettings, defaultRouteAdvertisementName, defaultRouteAdvertisementNamespace)

			if testCase.expectedError {
				assert.NotNil(t, err)
				assert.Nil(t, builder)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, builder)
			}
		})
	}
}

func buildValidRouteAdvertisement() *ovnv1.RouteAdvertisement {
	return &ovnv1.RouteAdvertisement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultRouteAdvertisementName,
			Namespace: defaultRouteAdvertisementNamespace,
		},
		Spec: ovnv1.RouteAdvertisementSpec{
			Advertisements: defaultAdvertisements,
		},
	}
}

func buildValidRouteAdvertisementBuilder(apiClient *clients.Settings) *RouteAdvertisementBuilder {
	return NewRouteAdvertisementBuilder(
		apiClient,
		defaultRouteAdvertisementName,
		defaultRouteAdvertisementNamespace,
		defaultAdvertisements,
	)
}

func buildDummyRouteAdvertisement() []runtime.Object {
	return append([]runtime.Object{}, buildValidRouteAdvertisement())
}
