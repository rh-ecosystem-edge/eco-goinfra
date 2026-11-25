package mco

import (
	"context"
	"fmt"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListMC returns a list of builders for MachineConfigs.
func ListMC(apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*MCBuilder, error) {
	logMessage := "Listing all MC resources"

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		logMessage += fmt.Sprintf(" with the options %v", options[0])
	}

	klog.V(100).Infof("%v", logMessage)

	return common.List[mcv1.MachineConfig, mcv1.MachineConfigList, MCBuilder](
		context.TODO(), apiClient, mcv1.Install, common.ConvertListOptionsToOptions(options)...)
}
