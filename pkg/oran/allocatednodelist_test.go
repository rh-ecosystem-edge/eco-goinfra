package oran

import (
	"testing"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListAllocatedNodes(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListAllocatedNodes,
		hardwaremanagementv1alpha1.AddToScheme,
		allocatedNodeGVK,
	).ExecuteTests(t)

	t.Run("multiple list options return error", func(t *testing.T) {
		t.Parallel()

		testSettings := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: pluginsTestSchemes,
		})

		_, err := ListAllocatedNodes(testSettings, runtimeclient.ListOptions{}, runtimeclient.ListOptions{})
		assert.EqualError(t, err, "error: more than one ListOptions was passed")
	})
}
