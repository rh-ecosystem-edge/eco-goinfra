package sriov

import (
	"context"
	"fmt"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	"k8s.io/klog/v2"
)

// ListPolicy returns SriovNetworkNodePolicies inventory in the given namespace.
func ListPolicy(apiClient *clients.Settings, nsname string, options ...client.ListOptions) ([]*PolicyBuilder, error) {
	if nsname == "" {
		klog.V(100).Info("SriovNetworkNodePolicies 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey(policyGVK.Kind, "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	allOptions := append([]client.ListOption{client.InNamespace(nsname)},
		common.ConvertListOptionsToOptions(options)...)

	return common.List[srIovV1.SriovNetworkNodePolicy, srIovV1.SriovNetworkNodePolicyList, PolicyBuilder](
		context.TODO(), apiClient, srIovV1.AddToScheme, allOptions...)
}

// CleanAllNetworkNodePolicies removes all SriovNetworkNodePolicies that are not set as default.
func CleanAllNetworkNodePolicies(
	apiClient *clients.Settings, operatornsname string, options ...client.ListOptions) error {
	klog.V(100).Infof("Cleaning up SriovNetworkNodePolicies in the %s namespace", operatornsname)

	if operatornsname == "" {
		klog.V(100).Info("'operatornsname' parameter can not be empty")

		return fmt.Errorf("failed to clean up SriovNetworkNodePolicies, 'operatornsname' parameter is empty")
	}

	policies, err := ListPolicy(apiClient, operatornsname, options...)
	if err != nil {
		klog.V(100).Infof("Failed to list SriovNetworkNodePolicies in namespace: %s", operatornsname)

		return err
	}

	for _, policy := range policies {
		// The sriovOperatorConfigName SriovNetworkNodePolicy is both mandatory and the default option.
		if policy.Object.Name != sriovOperatorConfigName {
			err = policy.Delete()
			if err != nil {
				klog.V(100).Infof("Failed to delete SriovNetworkNodePolicy: %s", policy.Object.Name)

				return err
			}
		}
	}

	return nil
}
