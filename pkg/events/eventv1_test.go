package events

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	eventsv1 "k8s.io/api/events/v1"
)

var eventV1GVK = eventsv1.SchemeGroupVersion.WithKind("Event")

func TestNewEventV1Builder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[eventsv1.Event, EventV1Builder](
		NewEventV1Builder, eventsv1.AddToScheme, eventV1GVK).
		ExecuteTests(t)
}

func TestPullEventV1(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[eventsv1.Event, EventV1Builder](
		PullEventV1, eventsv1.AddToScheme, eventV1GVK).
		ExecuteTests(t)
}

func TestEventV1BuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newEventV1CommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		Run(t)
}

// newEventV1CommonTestConfig returns the shared testhelper configuration for EventV1Builder tests.
func newEventV1CommonTestConfig() testhelper.CommonTestConfig[
	eventsv1.Event, EventV1Builder, *eventsv1.Event, *EventV1Builder] {
	return testhelper.NewCommonTestConfig[eventsv1.Event, EventV1Builder](
		eventsv1.AddToScheme, eventV1GVK, testhelper.ResourceScopeNamespaced)
}

func TestListEventV1s(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(ListEventV1s, eventsv1.AddToScheme, eventV1GVK).ExecuteTests(t)
}
