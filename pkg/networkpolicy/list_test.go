package networkpolicy

import (
	"errors"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig[netv1.NetworkPolicy, NetworkPolicyBuilder](
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*NetworkPolicyBuilder, error) {
			return List(apiClient, nsname)
		},
		netv1.AddToScheme,
		networkPolicyGVK,
	).ExecuteTests(t)
}

func TestListRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	testClient := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{netv1.AddToScheme},
	})

	builders, err := List(testClient, defaultNetworkPolicyNamespace, metav1.ListOptions{}, metav1.ListOptions{})

	assert.Equal(t, errors.New("error: more than one ListOptions was passed"), err)
	assert.Empty(t, builders)
}
