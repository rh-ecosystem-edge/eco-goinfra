package oran

import (
	"testing"

	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListClusterTemplates(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListClusterTemplates,
		provisioningv1alpha1.AddToScheme,
		clusterTemplateGVK,
	).ExecuteTests(t)
}

func TestListClusterTemplatesRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	testSettings := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: provisioningTestSchemes,
	})

	builders, err := ListClusterTemplates(
		testSettings,
		runtimeclient.ListOptions{},
		runtimeclient.ListOptions{},
	)
	require.EqualError(t, err, "error: more than one ListOptions was passed")
	assert.Nil(t, builders)
}
