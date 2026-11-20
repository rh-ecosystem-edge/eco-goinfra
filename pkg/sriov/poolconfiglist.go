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

// ListPoolConfigs returns a sriovNetworkPoolConfig list in a given namespace.
func ListPoolConfigs(apiClient *clients.Settings, namespace string) ([]*PoolConfigBuilder, error) {
	if namespace == "" {
		klog.V(100).Info("sriovNetworkPoolConfigs 'namespace' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey(poolConfigGVK.Kind, "", ""), commonerrors.BuilderFieldNamespace)
	}

	return common.List[srIovV1.SriovNetworkPoolConfig, srIovV1.SriovNetworkPoolConfigList, PoolConfigBuilder](
		context.TODO(), apiClient, srIovV1.AddToScheme, runtimeclient.InNamespace(namespace))
}

// CleanAllPoolConfigs removes all sriovNetworkPoolConfigs.
func CleanAllPoolConfigs(
	apiClient *clients.Settings, operatornsname string) error {
	klog.V(100).Infof("Cleaning up SriovNetworkPoolConfigs in the %s namespace", operatornsname)

	if operatornsname == "" {
		klog.V(100).Info("'operatornsname' parameter can not be empty")

		return fmt.Errorf("failed to clean up SriovNetworkPoolConfigs, 'operatornsname' parameter is empty")
	}

	poolConfigs, err := ListPoolConfigs(apiClient, operatornsname)
	if err != nil {
		klog.V(100).Infof("Failed to list SriovNetworkPoolConfigs in namespace: %s", operatornsname)

		return err
	}

	for _, poolConfig := range poolConfigs {
		err = poolConfig.Delete()
		if err != nil {
			klog.V(100).Infof("Failed to delete SriovNetworkPoolConfigs: %s", poolConfig.Object.Name)

			return err
		}
	}

	return nil
}
