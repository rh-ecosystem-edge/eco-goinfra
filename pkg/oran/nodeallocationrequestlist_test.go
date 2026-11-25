package oran

import (
	"testing"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListNodeAllocationRequests(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListNodeAllocationRequests,
		hardwaremanagementv1alpha1.AddToScheme,
		nodeAllocationRequestGVK,
	).ExecuteTests(t)
}

func TestListNodeAllocationRequestsOptions(t *testing.T) {
	t.Parallel()

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildListNodeAllocationRequest("matching", map[string]string{"selected": "true"}),
			buildListNodeAllocationRequest("not-matching", nil),
		},
		SchemeAttachers: pluginsTestSchemes,
	})

	builders, err := ListNodeAllocationRequests(testSettings, runtimeclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{"selected": "true"}),
	})
	require.NoError(t, err)
	require.Len(t, builders, 1)
	assert.Equal(t, "matching", builders[0].Definition.Name)

	builders, err = ListNodeAllocationRequests(testSettings,
		runtimeclient.ListOptions{}, runtimeclient.ListOptions{})
	assert.Nil(t, builders)
	assert.EqualError(t, err, "error: more than one ListOptions was passed")
}

func buildListNodeAllocationRequest(
	name string, objectLabels map[string]string) *hardwaremanagementv1alpha1.NodeAllocationRequest {
	return &hardwaremanagementv1alpha1.NodeAllocationRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test-namespace",
			Labels:    objectLabels,
		},
	}
}
