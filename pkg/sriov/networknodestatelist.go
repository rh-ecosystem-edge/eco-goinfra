package sriov

import (
	"context"
	"fmt"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListNetworkNodeState returns SriovNetworkNodeStates inventory in the given namespace.
func ListNetworkNodeState(
	apiClient *clients.Settings, nsname string, options ...runtimeclient.ListOptions) ([]*NetworkNodeStateBuilder, error) {
	if nsname == "" {
		klog.V(100).Info("SriovNetworkNodeStates 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey(networkNodeStateGVK.Kind, "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	convertedOptions := common.ConvertListOptionsToOptions(options)
	allOptions := append([]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)}, convertedOptions...)

	return common.List[srIovV1.SriovNetworkNodeState, srIovV1.SriovNetworkNodeStateList, NetworkNodeStateBuilder](
		context.TODO(), apiClient, srIovV1.AddToScheme, allOptions...)
}
