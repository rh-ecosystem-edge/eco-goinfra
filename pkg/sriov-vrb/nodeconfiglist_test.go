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

func TestVrbNodeConfigList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		ListNodeConfig,
		sriovvrbtypes.AddToScheme,
		nodeConfigGVK,
	).ExecuteTests(t)
}

func TestVrbNodeConfigListWithOptions(t *testing.T) {
	t.Parallel()

	selector := labels.SelectorFromSet(labels.Set{"selected": "true"})
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildNodeConfigWithLabels("selected", defaultNodeConfigNamespace, map[string]string{"selected": "true"}),
			buildNodeConfigWithLabels("other-namespace", "other-ns", map[string]string{"selected": "true"}),
			buildNodeConfigWithLabels("unselected", defaultNodeConfigNamespace, nil),
		},
		SchemeAttachers: []clients.SchemeAttacher{sriovvrbtypes.AddToScheme},
	})

	builders, err := ListNodeConfig(
		testSettings,
		defaultNodeConfigNamespace,
		client.ListOptions{Namespace: "ignored", LabelSelector: selector},
	)

	require.NoError(t, err)
	require.Len(t, builders, 1)
	assert.Equal(t, "selected", builders[0].Definition.Name)
}

func TestVrbNodeConfigListRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	testSettings := buildNodeConfigTestClient()
	builders, err := ListNodeConfig(
		testSettings,
		defaultNodeConfigNamespace,
		client.ListOptions{},
		client.ListOptions{},
	)

	require.Nil(t, builders)
	assert.EqualError(t, err, "error: more than one ListOptions was passed")
}

// buildNodeConfigWithLabels returns a SriovVrbNodeConfig with the provided metadata.
func buildNodeConfigWithLabels(name, namespace string, nodeLabels map[string]string) *sriovvrbtypes.SriovVrbNodeConfig {
	return &sriovvrbtypes.SriovVrbNodeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    nodeLabels,
		},
	}
}
