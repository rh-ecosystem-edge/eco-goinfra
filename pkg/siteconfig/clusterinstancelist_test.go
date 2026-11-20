package siteconfig

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	siteconfigv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/siteconfig/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListClusterInstances(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListClusterInstances,
		siteconfigv1alpha1.AddToScheme,
		clusterInstanceGVK,
	).ExecuteTests(t)
}

func TestListClusterInstancesWithNamespaceOption(t *testing.T) {
	t.Parallel()

	matching := generateClusterInstance()
	other := generateClusterInstance()
	other.Name = "other-cluster-instance"
	other.Namespace = "other-namespace"

	testSettings := buildClusterInstanceTestClient(matching, other)

	builders, err := ListClusterInstances(testSettings, runtimeclient.ListOptions{Namespace: testClusterInstance})

	require.NoError(t, err)
	require.Len(t, builders, 1)
	assert.Equal(t, testClusterInstance, builders[0].Definition.Name)
	assert.Equal(t, testClusterInstance, builders[0].Definition.Namespace)
	assert.Equal(t, clusterInstanceGVK, builders[0].GetGVK())
}

func TestListClusterInstancesRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	builders, err := ListClusterInstances(
		buildClusterInstanceClientWithDummyObject(),
		runtimeclient.ListOptions{Namespace: testClusterInstance},
		runtimeclient.ListOptions{Continue: "test"},
	)

	assert.EqualError(t, err, "error: more than one ListOptions was passed")
	assert.Nil(t, builders)
}

func buildClusterInstanceClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{generateClusterInstance()},
		SchemeAttachers: testSchemes,
	})
}

func TestListClusterInstancesReturnsSingleObject(t *testing.T) {
	t.Parallel()

	builders, err := ListClusterInstances(buildClusterInstanceClientWithDummyObject())

	require.NoError(t, err)
	require.Len(t, builders, 1)
	assert.Equal(t, testClusterInstance, builders[0].Object.Name)
	assert.Equal(t, testClusterInstance, builders[0].Object.Namespace)
}

func TestListClusterInstancesWithContinueOption(t *testing.T) {
	t.Parallel()

	builders, err := ListClusterInstances(
		buildClusterInstanceClientWithDummyObject(),
		runtimeclient.ListOptions{Continue: "test"},
	)

	require.NoError(t, err)
	require.Len(t, builders, 1)
	assert.Equal(t, testClusterInstance, builders[0].Definition.Name)
}
