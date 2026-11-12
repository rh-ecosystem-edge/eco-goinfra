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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// List returns sriov networks in the given namespace.
func List(apiClient *clients.Settings, nsname string, options ...client.ListOptions) ([]*NetworkBuilder, error) {
	resourceKey := commonkey.NewResourceKey(sriovNetworkGVK.Kind, "", "")

	if apiClient == nil {
		klog.V(100).Info("sriov network 'apiClient' parameter can not be empty")

		return nil, commonerrors.NewAPIClientNil(resourceKey)
	}

	if nsname == "" {
		klog.V(100).Info("sriov network 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(resourceKey, commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	listOptions := append([]client.ListOption{client.InNamespace(nsname)}, common.ConvertListOptionsToOptions(options)...)

	return common.List[srIovV1.SriovNetwork, srIovV1.SriovNetworkList, NetworkBuilder](
		context.TODO(), apiClient, srIovV1.AddToScheme, listOptions...)
}

// CleanAllNetworksByTargetNamespace deletes all networks matched by their NetworkNamespace spec.
func CleanAllNetworksByTargetNamespace(
	apiClient *clients.Settings,
	operatornsname string,
	targetnsname string,
	options ...client.ListOptions) error {
	klog.V(100).Infof("Cleaning up sriov networks in the %s namespace with %s NetworkNamespace spec",
		operatornsname, targetnsname)

	if operatornsname == "" {
		klog.V(100).Info("'operatornsname' parameter can not be empty")

		return fmt.Errorf("failed to clean up sriov networks, 'operatornsname' parameter is empty")
	}

	if targetnsname == "" {
		klog.V(100).Info("'targetnsname' parameter can not be empty")

		return fmt.Errorf("failed to clean up sriov networks, 'targetnsname' parameter is empty")
	}

	networks, err := List(apiClient, operatornsname, options...)
	if err != nil {
		klog.V(100).Infof("Failed to list sriov networks in namespace: %s", operatornsname)

		return err
	}

	for _, network := range networks {
		if network.Object.Spec.NetworkNamespace == targetnsname {
			err = network.Delete()
			if err != nil {
				klog.V(100).Infof("Failed to delete sriov networks: %s", network.Object.Name)

				return err
			}
		}
	}

	return nil
}
