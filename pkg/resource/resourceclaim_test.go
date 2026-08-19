package resource

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var resourceClaimGVK = resourcev1.SchemeGroupVersion.WithKind("ResourceClaim")

func TestNewResourceClaimBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewResourceClaimBuilder,
		resourcev1.AddToScheme,
		resourceClaimGVK,
	).ExecuteTests(t)
}

func TestPullResourceClaim(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullResourceClaim,
		resourcev1.AddToScheme,
		resourceClaimGVK,
	).ExecuteTests(t)
}

func TestListResourceClaims(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		namespace           string
		client              bool
		expectedCount       int
		expectedError       bool
	}{
		{
			name:                "valid with claims",
			addToRuntimeObjects: true,
			namespace:           "test-ns",
			client:              true,
			expectedCount:       1,
		},
		{
			name:                "valid no claims",
			addToRuntimeObjects: false,
			namespace:           "test-ns",
			client:              true,
			expectedCount:       0,
		},
		{
			name:          "nil apiClient",
			namespace:     "test-ns",
			client:        false,
			expectedError: true,
		},
		{
			name:          "empty namespace",
			namespace:     "",
			client:        true,
			expectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				runtimeObjects []runtime.Object
				testSettings   *clients.Settings
			)

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects,
					generateResourceClaim("test-ns"))
			}

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  runtimeObjects,
					SchemeAttachers: testResourceSchemes,
				})
			}

			builders, err := ListResourceClaims(testSettings, testCase.namespace)

			if testCase.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, builders, testCase.expectedCount)
			}
		})
	}
}

func TestListResourceClaimsNamespaceOverride(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			generateResourceClaim("correct-ns"),
		},
		SchemeAttachers: testResourceSchemes,
	})

	builders, err := ListResourceClaims(testSettings, "correct-ns",
		runtimeclient.InNamespace("wrong-ns"))
	assert.NoError(t, err)
	assert.Len(t, builders, 1)
	assert.Equal(t, "test-claim", builders[0].Object.Name)
}

func TestResourceClaimBuilderExists(t *testing.T) {
	testCases := []struct {
		name           string
		addToRuntime   bool
		expectedStatus bool
	}{
		{
			name:           "object exists",
			addToRuntime:   true,
			expectedStatus: true,
		},
		{
			name:           "object does not exist",
			addToRuntime:   false,
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object
			if testCase.addToRuntime {
				runtimeObjects = append(runtimeObjects,
					generateResourceClaim("test-ns"))
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testResourceSchemes,
			})

			builder := NewResourceClaimBuilder(testSettings, "test-claim", "test-ns")
			assert.Equal(t, testCase.expectedStatus, builder.Exists())
		})
	}
}

func TestResourceClaimBuilderDelete(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			generateResourceClaim("test-ns"),
		},
		SchemeAttachers: testResourceSchemes,
	})

	builder := NewResourceClaimBuilder(testSettings, "test-claim", "test-ns")
	err := builder.Delete()
	assert.NoError(t, err)
	assert.Nil(t, builder.Object)
}

func TestPullResourceClaimNotFound(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testResourceSchemes,
	})

	builder, err := PullResourceClaim(testSettings, "nonexistent", "test-ns")
	assert.Error(t, err)
	assert.Nil(t, builder)
}

func TestResourceClaimBuilderGetGVK(t *testing.T) {
	builder := &ResourceClaimBuilder{}
	assert.Equal(t, resourceClaimGVK, builder.GetGVK())
}

func generateResourceClaim(namespace string) *resourcev1.ResourceClaim {
	return &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-claim",
			Namespace: namespace,
		},
	}
}
