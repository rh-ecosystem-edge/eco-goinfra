package nmstate

import (
	"context"
	"fmt"

	nmstateV1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/klog/v2"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPolicy returns a list of NodeNetworkConfigurationPolicy.
func ListPolicy(apiClient *clients.Settings, options ...goclient.ListOptions) ([]*PolicyBuilder, error) {
	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	klog.V(100).Infof("Listing NodeNetworkConfigurationPolicy with the options %v", options)

	return common.List[
		nmstateV1.NodeNetworkConfigurationPolicy,
		nmstateV1.NodeNetworkConfigurationPolicyList,
		PolicyBuilder,
	](context.TODO(), apiClient, nmstateV1.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}

// CleanAllNMStatePolicies removes all NodeNetworkConfigurationPolicies.
func CleanAllNMStatePolicies(apiClient *clients.Settings, options ...goclient.ListOptions) error {
	klog.V(100).Info("Cleaning up NodeNetworkConfigurationPolicies")

	nncpList, err := ListPolicy(apiClient, options...)
	if err != nil {
		klog.V(100).Info("Failed to list NodeNetworkConfigurationPolicies")

		return err
	}

	for _, nncpPolicy := range nncpList {
		_, err = nncpPolicy.Delete()
		if err != nil {
			klog.V(100).Infof("Failed to delete NodeNetworkConfigurationPolicy: %s", nncpPolicy.Object.Name)

			return err
		}
	}

	return nil
}
