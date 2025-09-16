package sriovvrb

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	sriovvrbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/vrbtypes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListClusterConfig returns SriovVrbClusterConfigList from given namespace.
func ListClusterConfig(
	apiClient *clients.Settings,
	nsname string,
	options ...client.ListOptions) ([]*ClusterConfigBuilder, error) {
	if apiClient == nil {
		klog.V(100).Info("SriovVrbClusterConfigList 'apiClient' parameter can not be empty")

		return nil, fmt.Errorf("failed to list SriovVrbClusterConfig, 'apiClient' parameter is empty")
	}

	err := apiClient.AttachScheme(sriovvrbtypes.AddToScheme)
	if err != nil {
		klog.V(100).Info("Failed to add sriov-vrb scheme to client schemes")

		return nil, err
	}

	if nsname == "" {
		klog.V(100).Info("SriovVrbClusterConfigList 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list SriovVrbClusterConfig, 'nsname' parameter is empty")
	}

	logMessage := fmt.Sprintf("Listing SriovVrbClusterConfig in the namespace %s", nsname)
	passedOptions := client.ListOptions{}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	klog.V(100).Infof("%v", logMessage)

	passedOptions.Namespace = nsname

	sfncList := new(sriovvrbtypes.SriovVrbClusterConfigList)

	err = apiClient.List(context.TODO(), sfncList, &passedOptions)
	if err != nil {
		klog.V(100).Infof("Failed to list SriovVrbClusterConfigs in the namespace %s due to %s", nsname, err.Error())

		return nil, err
	}

	var sfncBuilderList []*ClusterConfigBuilder

	for _, sfnc := range sfncList.Items {
		copiedObject := sfnc
		sfncBuilder := &ClusterConfigBuilder{
			apiClient:  apiClient.Client,
			Object:     &copiedObject,
			Definition: &copiedObject,
		}

		sfncBuilderList = append(sfncBuilderList, sfncBuilder)
	}

	return sfncBuilderList, nil
}
