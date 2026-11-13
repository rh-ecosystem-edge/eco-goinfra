package sriovvrb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	sriovvrbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/vrbtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestVrbClusterConfigList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		ListClusterConfig,
		sriovvrbtypes.AddToScheme,
		clusterConfigGVK,
	).ExecuteTests(t)
}

func TestVrbClusterConfigListWithOptions(t *testing.T) {
	t.Parallel()

	selector := labels.SelectorFromSet(labels.Set{"selected": "true"})
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildClusterConfigWithLabels(
				"selected", defaultClusterConfigNamespace, map[string]string{"selected": "true"}),
			buildClusterConfigWithLabels("other-namespace", "other-ns", map[string]string{"selected": "true"}),
			buildClusterConfigWithLabels("unselected", defaultClusterConfigNamespace, nil),
		},
		SchemeAttachers: []clients.SchemeAttacher{sriovvrbtypes.AddToScheme},
	})

	builders, err := ListClusterConfig(
		testSettings,
		defaultClusterConfigNamespace,
		client.ListOptions{Namespace: "ignored", LabelSelector: selector},
	)

	require.NoError(t, err)
	require.Len(t, builders, 1)
	assert.Equal(t, "selected", builders[0].Definition.Name)
}

func TestVrbClusterConfigListRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	testSettings := buildClusterConfigTestClient()
	builders, err := ListClusterConfig(
		testSettings,
		defaultClusterConfigNamespace,
		client.ListOptions{},
		client.ListOptions{},
	)

	require.Nil(t, builders)
	assert.EqualError(t, err, "error: more than one ListOptions was passed")
}

// buildClusterConfigWithLabels returns a SriovVrbClusterConfig with the provided metadata.
func buildClusterConfigWithLabels(
	name, namespace string, clusterLabels map[string]string) *sriovvrbtypes.SriovVrbClusterConfig {
	return &sriovvrbtypes.SriovVrbClusterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    clusterLabels,
		},
	}
}
