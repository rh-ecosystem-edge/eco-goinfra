package sriov

import (
	"fmt"
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListNetworkNodeState(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		ListNetworkNodeState,
		srIovV1.AddToScheme,
		networkNodeStateGVK,
	).ExecuteTests(t)
}

func TestListNetworkNodeStateRespectsNamespace(t *testing.T) {
	t.Parallel()

	settings := getNetworkNodeStateTestClient(
		buildNodeNetworkState(defaultNodeName, defaultNodeNsName),
		buildNodeNetworkState("other-node", "other-namespace"),
	)

	builders, err := ListNetworkNodeState(settings, defaultNodeNsName)
	require.NoError(t, err)
	require.Len(t, builders, 1)
	assert.Equal(t, defaultNodeName, builders[0].Definition.Name)
	assert.Equal(t, defaultNodeNsName, builders[0].Definition.Namespace)
	assert.Same(t, builders[0].Object, builders[0].Objects)
}

func TestListNetworkNodeStateRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	settings := getNetworkNodeStateTestClient()
	options := []runtimeclient.ListOptions{{}, {}}

	builders, err := ListNetworkNodeState(settings, defaultNodeNsName, options...)
	assert.Nil(t, builders)
	assert.Equal(t, fmt.Errorf("error: more than one ListOptions was passed"), err)
}
