package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListFrrNodeState(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListFrrNodeState,
		frrtypes.AddToScheme,
		frrNodeStateGVK,
	).ExecuteTests(t)
}

func TestListFrrNodeStateRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	builders, err := ListFrrNodeState(
		clients.GetTestClients(clients.TestClientParams{}),
		runtimeclient.ListOptions{},
		runtimeclient.ListOptions{},
	)

	require.EqualError(t, err, "error: more than one ListOptions was passed")
	assert.Nil(t, builders)
}

func TestListFrrNodeStateWithLabelSelector(t *testing.T) {
	t.Parallel()

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildFrrNodeStateWithLabels("worker-0", map[string]string{"role": "worker"}),
			buildFrrNodeStateWithLabels("worker-1", map[string]string{"role": "worker"}),
			buildFrrNodeStateWithLabels("control-plane-0", map[string]string{"role": "control-plane"}),
		},
		SchemeAttachers: []clients.SchemeAttacher{frrtypes.AddToScheme},
	})

	builders, err := ListFrrNodeState(testSettings, runtimeclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"role": "worker"}),
	})

	require.NoError(t, err)
	require.Len(t, builders, 2)
	assert.ElementsMatch(t, []string{"worker-0", "worker-1"}, []string{
		builders[0].Definition.Name,
		builders[1].Definition.Name,
	})
}

// buildFrrNodeStateWithLabels returns an FRRNodeState with the provided name and labels.
func buildFrrNodeStateWithLabels(name string, resourceLabels map[string]string) *frrtypes.FRRNodeState {
	return &frrtypes.FRRNodeState{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: resourceLabels,
		},
	}
}
