package ovn

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListRouteAdvertisements(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListRouteAdvertisements,
		ovnv1.AddToScheme,
		routeAdvertisementGVK,
	).ExecuteTests(t)
}

func TestListRouteAdvertisementsWithOptions(t *testing.T) {
	t.Parallel()

	matchingRouteAdvertisement := buildDummyRouteAdvertisement(defaultRouteAdvertisementName)
	matchingRouteAdvertisement.Labels = map[string]string{"test": "label"}
	nonMatchingRouteAdvertisement := buildDummyRouteAdvertisement("non-matching-routeadvertisement")

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			matchingRouteAdvertisement,
			nonMatchingRouteAdvertisement,
		},
		SchemeAttachers: []clients.SchemeAttacher{ovnv1.AddToScheme},
	})

	builders, err := ListRouteAdvertisements(testSettings, runtimeClient.MatchingLabels{"test": "label"})

	require.NoError(t, err)
	require.Len(t, builders, 1)
	assert.Equal(t, defaultRouteAdvertisementName, builders[0].Definition.Name)
}
