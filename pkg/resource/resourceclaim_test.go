package resource

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListResourceClaims(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
		namespace           string
		expectedCount       int
		client              bool
		expectedError       error
	}{
		{
			addToRuntimeObjects: true,
			namespace:           "test-ns",
			expectedCount:       1,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: false,
			namespace:           "test-ns",
			expectedCount:       0,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: true,
			namespace:           "test-ns",
			expectedCount:       0,
			client:              false,
			expectedError:       fmt.Errorf("resourceClaim 'apiClient' cannot be nil"),
		},
		{
			addToRuntimeObjects: false,
			namespace:           "",
			expectedCount:       0,
			client:              true,
			expectedError:       fmt.Errorf("resourceClaim 'namespace' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects,
				generateResourceClaim("test-claim", "test-ns"))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testResourceSchemes,
			})
		}

		claims, err := ListResourceClaims(testSettings, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Len(t, claims, testCase.expectedCount)
		}
	}
}

func TestListResourceClaimsNamespaceOverride(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			generateResourceClaim("test-claim", "correct-ns"),
		},
		SchemeAttachers: testResourceSchemes,
	})

	claims, err := ListResourceClaims(testSettings, "correct-ns",
		goclient.InNamespace("wrong-ns"))
	assert.Nil(t, err)
	assert.Len(t, claims, 1)
	assert.Equal(t, "test-claim", claims[0].Name)
}

func TestGetResourceClaim(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                "test-claim",
			namespace:           "test-ns",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "test-claim",
			namespace:           "test-ns",
			addToRuntimeObjects: false,
			client:              false,
			expectedError:       fmt.Errorf("resourceClaim 'apiClient' cannot be nil"),
		},
		{
			name:                "",
			namespace:           "test-ns",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("resourceClaim 'name' cannot be empty"),
		},
		{
			name:                "test-claim",
			namespace:           "",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("resourceClaim 'namespace' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects,
				generateResourceClaim(testCase.name, testCase.namespace))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testResourceSchemes,
			})
		}

		claim, err := GetResourceClaim(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, claim)
			assert.Equal(t, testCase.name, claim.Name)
		}
	}
}

func TestGetResourceClaimNotFound(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testResourceSchemes,
	})

	claim, err := GetResourceClaim(testSettings, "nonexistent", "test-ns")
	assert.NotNil(t, err)
	assert.Nil(t, claim)
}

func generateResourceClaim(name, namespace string) *resourcev1.ResourceClaim {
	return &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
