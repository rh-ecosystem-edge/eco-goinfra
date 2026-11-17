package sriov

import (
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPolicyUpdate(t *testing.T) {
	apiClient := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{srIovV1.AddToScheme},
		K8sMockObjects:  []runtime.Object{}})
	builder := NewPolicyBuilder(
		apiClient, "update-policy", "operator", "testresource", 2, []string{"eth1"}, map[string]string{"kubernetes.io/hostname": "node1"})
	_, err := builder.Create()
	require.NoError(t, err)
	stale, err := PullPolicy(apiClient, "update-policy", "operator")
	require.NoError(t, err)

	builder.Definition.Spec.NumVfs = 4
	_, err = builder.Update()
	require.NoError(t, err)
	current, err := builder.Get()
	require.NoError(t, err)
	require.Equal(t, 4, current.Spec.NumVfs)
	require.Equal(t, "testresource", current.Spec.ResourceName)
	stale.Definition.Spec.NumVfs = 8
	_, err = stale.Update()
	require.True(t, apierrors.IsConflict(err), "expected update conflict, got %v", err)
}
