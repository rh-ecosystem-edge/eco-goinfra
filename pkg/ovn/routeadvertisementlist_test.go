package ovn

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListRouteAdvertisements(t *testing.T) {
	testCases := []struct {
		name                 string
		namespace            string
		addToRuntimeObjects  bool
		expectedError        bool
		client               bool
		listOptions          []runtimeClient.ListOptions
		expectedObjectsCount int
	}{
		{
			name:                 "list RouteAdvertisements in namespace",
			namespace:            defaultRouteAdvertisementNamespace,
			addToRuntimeObjects:  true,
			expectedError:        false,
			client:               true,
			listOptions:          nil,
			expectedObjectsCount: 1,
		},
		{
			name:                 "list RouteAdvertisements in empty namespace",
			namespace:            defaultRouteAdvertisementNamespace,
			addToRuntimeObjects:  false,
			expectedError:        false,
			client:               true,
			listOptions:          nil,
			expectedObjectsCount: 0,
		},
		{
			name:                 "list RouteAdvertisements with nil client",
			namespace:            defaultRouteAdvertisementNamespace,
			addToRuntimeObjects:  false,
			expectedError:        true,
			client:               false,
			listOptions:          nil,
			expectedObjectsCount: 0,
		},
		{
			name:                 "list RouteAdvertisements with multiple ListOptions",
			namespace:            defaultRouteAdvertisementNamespace,
			addToRuntimeObjects:  true,
			expectedError:        true,
			client:               true,
			listOptions:          []runtimeClient.ListOptions{{}, {}},
			expectedObjectsCount: 0,
		},
		{
			name:                 "list RouteAdvertisements with single ListOption",
			namespace:            defaultRouteAdvertisementNamespace,
			addToRuntimeObjects:  true,
			expectedError:        false,
			client:               true,
			listOptions:          []runtimeClient.ListOptions{{}},
			expectedObjectsCount: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object
			var testSettings *clients.Settings

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects, buildValidRouteAdvertisement())
			}

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects: runtimeObjects,
				})
			}

			routeAdvertisements, err := ListRouteAdvertisements(testSettings, testCase.namespace, testCase.listOptions...)

			if testCase.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Len(t, routeAdvertisements, testCase.expectedObjectsCount)
			}
		})
	}
}

func buildRouteAdvertisementWithName(name string) *ovnv1.RouteAdvertisement {
	return &ovnv1.RouteAdvertisement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultRouteAdvertisementNamespace,
		},
		Spec: ovnv1.RouteAdvertisementSpec{
			Advertisements: defaultAdvertisements,
		},
	}
}
