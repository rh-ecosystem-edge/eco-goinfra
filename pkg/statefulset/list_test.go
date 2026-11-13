package statefulset

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*Builder, error) {
			return List(apiClient, nsname)
		},
		appsv1.AddToScheme,
		statefulSetGVK,
	).ExecuteTests(t)
}

func TestListInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListInAllNamespaces,
		appsv1.AddToScheme,
		statefulSetGVK,
	).ExecuteTests(t)
}

func TestListInAllNamespacesWithLabelSelector(t *testing.T) {
	t.Parallel()

	matching := testStatefulSetObject("matching", "namespace-one")
	matching.Labels = map[string]string{"demo": "matching"}
	nonMatching := testStatefulSetObject("non-matching", "namespace-two")
	nonMatching.Labels = map[string]string{"demo": "non-matching"}

	builders, err := ListInAllNamespaces(
		newStatefulSetTestClient(matching, nonMatching),
		metav1.ListOptions{LabelSelector: "demo=matching"},
	)

	require.NoError(t, err)
	require.Len(t, builders, 1)
	assert.Equal(t, "matching", builders[0].Definition.Name)
}
