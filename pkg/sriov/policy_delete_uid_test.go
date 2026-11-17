package sriov

import (
	"context"
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestPolicyDeleteWithUID(t *testing.T) {
	expectedUID := types.UID("owned-policy")
	called := false
	apiClient := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{srIovV1.AddToScheme},
		InterceptorFuncs: interceptor.Funcs{Delete: func(_ context.Context, _ runtimeclient.WithWatch, _ runtimeclient.Object,
			options ...runtimeclient.DeleteOption) error {
			called = true
			opts := (&runtimeclient.DeleteOptions{}).ApplyOptions(options)
			require.NotNil(t, opts.Preconditions)
			require.NotNil(t, opts.Preconditions.UID)
			require.Equal(t, expectedUID, *opts.Preconditions.UID)

			return apierrors.NewConflict(
				schema.GroupResource{Group: "sriovnetwork.openshift.io", Resource: "sriovnetworknodepolicies"}, "policy", nil)
		}},
	})
	builder := NewPolicyBuilder(
		apiClient, "policy", "operator", "testresource", 2, []string{"eth1"}, map[string]string{"kubernetes.io/hostname": "node1"})
	require.Error(t, builder.DeleteWithUID(""))
	require.False(t, called)
	require.True(t, apierrors.IsConflict(builder.DeleteWithUID(expectedUID)))
	require.True(t, called)
}
