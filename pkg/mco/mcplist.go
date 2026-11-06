package mco

import (
	"context"
	"fmt"
	"time"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/util/wait"
)

// ListMCP returns a list of MachineConfigPoolBuilder.
func ListMCP(apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*MCPBuilder, error) {
	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[mcv1.MachineConfigPool, mcv1.MachineConfigPoolList, MCPBuilder](
		context.TODO(), apiClient, mcv1.Install, common.ConvertListOptionsToOptions(options)...)
}

// ListMCPByMachineConfigSelector returns a list of MachineConfigurationPoolBuilders for given selector.
func ListMCPByMachineConfigSelector(
	apiClient *clients.Settings, mcpLabel string, options ...runtimeclient.ListOptions) (*MCPBuilder, error) {
	klog.V(100).Infof("GetByLabel returns MachineConfigPool with the specified label: %v", mcpLabel)

	mcpList, err := ListMCP(apiClient, options...)
	if err != nil {
		return nil, err
	}

	for _, mcp := range mcpList {
		if mcp.Object.Spec.MachineConfigSelector == nil {
			continue
		}

		for _, label := range mcp.Object.Spec.MachineConfigSelector.MatchExpressions {
			for _, value := range label.Values {
				if value == mcpLabel {
					return mcp, nil
				}
			}
		}

		for _, label := range mcp.Object.Spec.MachineConfigSelector.MatchLabels {
			if label == mcpLabel {
				return mcp, nil
			}
		}
	}

	return nil, fmt.Errorf("cannot find MachineConfigPool that targets machineConfig with label: %s", mcpLabel)
}

// ListMCPWaitToBeStableFor waits for a given MachineConfigurationPool to be stable for a given period.
func ListMCPWaitToBeStableFor(
	apiClient *clients.Settings, stableDuration, timeout time.Duration, options ...runtimeclient.ListOptions) error {
	if apiClient == nil {
		klog.V(100).Info("MachineConfigPool 'apiClient' can not be empty")

		return commonerrors.NewAPIClientNil(commonkey.NewResourceKey(mcpGVK.Kind, "", ""))
	}

	klog.V(100).Infof("WaitForMcpListToBeStableFor waits up to duration of %v for "+
		"MachineConfigPoolList to be stable for %v", timeout, stableDuration)

	isMcpListStable := true

	// Wait 5 secs in each iteration before condition function () returns true or errors or times out
	// after stableDuration
	err := wait.PollUntilContextTimeout(
		context.TODO(), fiveScds, timeout, true, func(ctx context.Context) (bool, error) {
			isMcpListStable = true

			// check if cluster is stable every 5 seconds during entire stableDuration time period
			// Here we need to run through the entire stableDuration till it times out.
			_ = wait.PollUntilContextTimeout(
				context.TODO(), fiveScds, stableDuration, true, func(ctx2 context.Context) (done bool, err error) {
					mcpList, err := ListMCP(apiClient, options...)
					if err != nil {
						return false, err
					}

					// iterate through the MachineConfigPools in the list.
					for _, mcp := range mcpList {
						if mcp.Object.Status.ReadyMachineCount != mcp.Object.Status.MachineCount ||
							mcp.Object.Status.MachineCount != mcp.Object.Status.UpdatedMachineCount ||
							mcp.Object.Status.DegradedMachineCount != 0 {
							isMcpListStable = false

							klog.V(100).Infof("MachineConfigPool: %v degraded and has a mismatch in "+
								"machineCount: %v "+"vs machineCountUpdated: "+"%v vs readyMachineCount: %v and "+
								"degradedMachineCount is : %v \n", mcp.Object.Name,
								mcp.Object.Status.MachineCount, mcp.Object.Status.UpdatedMachineCount,
								mcp.Object.Status.ReadyMachineCount, mcp.Object.Status.DegradedMachineCount)

							return true, err
						}
					}

					// Here we are always returning "false, nil" so we keep iterating throughout the stableInterval
					// of the inner wait.PollUntilContextTimeout loop, until we time out.
					return false, nil
				})

			if isMcpListStable {
				klog.V(100).Infof("MachineConfigPools were stable during during stableDuration: %v",
					stableDuration)

				// exit the outer wait.PollUntilContextTimeout block since the mcps were stable during stableDuration.
				return true, nil
			}

			klog.V(100).Infof("MachineConfigPools were not stable during stableDuration: %v, retrying ...",
				stableDuration)

			// keep iterating in the outer wait.PollUntilContextTimeout waiting for cluster to be stable.
			return false, nil
		})
	if err == nil {
		klog.V(100).Infof("Cluster was stable during stableDuration: %v", stableDuration)
	} else {
		// Here err is "timed out waiting for the condition"
		klog.V(100).Infof("Cluster was Un-stable during stableDuration: %v", stableDuration)
	}

	return err
}
