package sriovvrb

import (
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	sriovvrbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/vrbtypes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListNodeConfig returns SriovVrbNodeConfigList from given namespace.
func ListNodeConfig(
	apiClient *clients.Settings,
	nsname string,
	options ...client.ListOptions) ([]*NodeConfigBuilder, error) {
	if apiClient == nil {
		klog.V(100).Info("SriovVrbNodeConfigList 'apiClient' parameter can not be empty")

		return nil, fmt.Errorf("failed to list SriovVrbNodeConfig, 'apiClient' parameter is empty")
	}

	err := apiClient.AttachScheme(sriovvrbtypes.AddToScheme)
	if err != nil {
		klog.V(100).Info("Failed to add sriov-vrb scheme to client schemes")

		return nil, err
	}

	if nsname == "" {
		klog.V(100).Info("SriovVrbNodeConfigList 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list SriovVrbNodeConfig, 'nsname' parameter is empty")
	}

	logMessage := fmt.Sprintf("Listing SriovVrbNodeConfig in the namespace %s", nsname)
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

	sfncList := new(sriovvrbtypes.SriovVrbNodeConfigList)

	err = apiClient.List(logging.DiscardContext(), sfncList, &passedOptions)
	if err != nil {
		klog.V(100).Infof("Failed to list SriovVrbNodeConfigs in the namespace %s due to %s", nsname, err.Error())

		return nil, err
	}

	var sfncBuilderList []*NodeConfigBuilder

	for _, sfnc := range sfncList.Items {
		copiedObject := sfnc
		sfncBuilder := &NodeConfigBuilder{
			apiClient:  apiClient.Client,
			Object:     &copiedObject,
			Definition: &copiedObject,
		}

		sfncBuilderList = append(sfncBuilderList, sfncBuilder)
	}

	return sfncBuilderList, nil
}
