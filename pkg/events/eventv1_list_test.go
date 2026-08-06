package events

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	eventsv1 "k8s.io/api/events/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListEventV1(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig[eventsv1.Event, EventV1Builder](
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*EventV1Builder, error) {
			return ListEventV1(apiClient, nsname)
		},
		eventsv1.AddToScheme,
		eventV1GVK,
	).ExecuteTests(t)
}
