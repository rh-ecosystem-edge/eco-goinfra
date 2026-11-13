package machine

import (
	"context"
	"fmt"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListWorkerMachineSets returns a slice of SetBuilder objects in a namespace on a cluster.
func ListWorkerMachineSets(
	apiClient *clients.Settings,
	namespace string,
	workerLabel string,
	options ...metav1.ListOptions) ([]*SetBuilder, error) {
	if namespace == "" {
		klog.V(100).Info("machineSet 'namespace' parameter can not be empty")

		return nil, fmt.Errorf("failed to list MachineSets, 'namespace' parameter is empty")
	}

	if workerLabel == "" {
		klog.V(100).Info("machineSet 'workerLabel' parameter can not be empty")

		return nil, fmt.Errorf("failed to list MachineSets, 'workerLabel' parameter is empty")
	}

	logMessage := fmt.Sprintf("Listing all workerMachinesSets in the namespace %s", namespace)
	listOptions := []runtimeclient.ListOption{runtimeclient.InNamespace(namespace)}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		logMessage += fmt.Sprintf(" with the options %v", options[0])

		convertedOptions, err := common.ConvertMetaListOptionsToListOptions(options)
		if err != nil {
			return nil, err
		}

		listOptions = append(listOptions, convertedOptions...)
	}

	klog.V(100).Infof("%v", logMessage)

	machineSetBuilders, err := common.List[machinev1beta1.MachineSet, machinev1beta1.MachineSetList, SetBuilder](
		context.TODO(), apiClient, machinev1beta1.AddToScheme, listOptions...)
	if err != nil {
		klog.V(100).Infof("Failed to list MachineSets in the namespace %s due to %s",
			namespace, err.Error())

		return nil, err
	}

	var machineSetObjects []*SetBuilder

	for _, machineSetBuilder := range machineSetBuilders {
		if val, ok := machineSetBuilder.Definition.Spec.Template.Labels[workerLabel]; ok && val == "worker" {
			machineSetObjects = append(machineSetObjects, machineSetBuilder)
		}
	}

	return machineSetObjects, nil
}
