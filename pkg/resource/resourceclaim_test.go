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
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*ResourceClaimBuilder, error) {
			return ListResourceClaims(apiClient, nsname)
		},
		resourcev1.AddToScheme,
		resourceClaimGVK,
	).ExecuteTests(t)
}

func TestResourceClaimBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[resourcev1.ResourceClaim, ResourceClaimBuilder](
		resourcev1.AddToScheme,
		resourceClaimGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
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
