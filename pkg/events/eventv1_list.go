package events

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListEventV1 returns events.k8s.io/v1 Event inventory in the given namespace.
func ListEventV1(
	apiClient *clients.Settings, nsname string, options ...metav1.ListOptions) ([]*EventV1Builder, error) {
	if nsname == "" {
		klog.V(100).Info("events.k8s.io/v1 Event 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("Event", "", ""), commonerrors.BuilderFieldNamespace)
	}

	convertedOptions, err := common.ConvertMetaListOptionsToListOptions(options)
	if err != nil {
		return nil, err
	}

	allOptions := append([]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)}, convertedOptions...)

	return common.List[eventsv1.Event, eventsv1.EventList, EventV1Builder](
		context.TODO(), apiClient, eventsv1.AddToScheme, allOptions...)
}
