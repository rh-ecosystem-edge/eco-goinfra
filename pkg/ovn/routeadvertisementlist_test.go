package ovn

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestListRouteAdvertisements(t *testing.T) {
	testCases := []struct {
		name                 string
		addToRuntimeObjects  bool
		expectedError        bool
		client               bool
		listOptions          []metav1.ListOptions
		expectedObjectsCount int
	}{
		{
			name:                 "list RouteAdvertisements (cluster-scoped)",
			addToRuntimeObjects:  true,
			expectedError:        false,
			client:               true,
			listOptions:          nil,
			expectedObjectsCount: 1,
		},
		{
			name:                 "list RouteAdvertisements with no objects",
			addToRuntimeObjects:  false,
			expectedError:        false,
			client:               true,
			listOptions:          nil,
			expectedObjectsCount: 0,
		},
		{
			name:                 "list RouteAdvertisements with nil client",
			addToRuntimeObjects:  false,
			expectedError:        true,
			client:               false,
			listOptions:          nil,
			expectedObjectsCount: 0,
		},
		{
			name:                 "list RouteAdvertisements with multiple ListOptions",
			addToRuntimeObjects:  true,
			expectedError:        true,
			client:               true,
			listOptions:          []metav1.ListOptions{{}, {}},
			expectedObjectsCount: 0,
		},
		{
			name:                 "list RouteAdvertisements with single ListOption",
			addToRuntimeObjects:  true,
			expectedError:        false,
			client:               true,
			listOptions:          []metav1.ListOptions{{}},
			expectedObjectsCount: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object
			var testSettings *clients.Settings

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects, buildRouteAdvertisementWithName("test-routeadvertisement"))
			}

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects: runtimeObjects,
				})
			}

			routeAdvertisements, err := ListRouteAdvertisements(testSettings, testCase.listOptions...)

			if testCase.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Len(t, routeAdvertisements, testCase.expectedObjectsCount)
			}
		})
	}
}

func buildRouteAdvertisementWithName(name string) *ovnv1.RouteAdvertisements {
	return &ovnv1.RouteAdvertisements{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: ovnv1.RouteAdvertisementsSpec{
			Advertisements: defaultAdvertisements,
		},
	}
}
