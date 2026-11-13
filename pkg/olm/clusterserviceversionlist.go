package olm

import (
	"context"
	"fmt"
	"strings"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	oplmV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListClusterServiceVersion returns clusterserviceversion inventory in the given namespace.
func ListClusterServiceVersion(
	apiClient *clients.Settings,
	nsname string,
	options ...runtimeclient.ListOptions) ([]*ClusterServiceVersionBuilder, error) {
	if nsname == "" {
		klog.V(100).Info("clusterserviceversion 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey(clusterServiceVersionGVK.Kind, "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	allOptions := append(
		[]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)},
		common.ConvertListOptionsToOptions(options)...,
	)

	return common.List[oplmV1alpha1.ClusterServiceVersion, oplmV1alpha1.ClusterServiceVersionList,
		ClusterServiceVersionBuilder](context.TODO(), apiClient, oplmV1alpha1.AddToScheme, allOptions...)
}

// ListClusterServiceVersionWithNamePattern returns a cluster-wide clusterserviceversion inventory
// filtered by the name pattern.
func ListClusterServiceVersionWithNamePattern(
	apiClient *clients.Settings,
	namePattern string,
	nsname string,
	options ...runtimeclient.ListOptions) ([]*ClusterServiceVersionBuilder, error) {
	if apiClient == nil {
		klog.V(100).Info("The apiClient cannot be nil")

		return nil, commonerrors.NewAPIClientNil(
			commonkey.NewResourceKey(clusterServiceVersionGVK.Kind, "", ""))
	}

	if namePattern == "" {
		klog.V(100).Info(
			"The namePattern field to filter out all relevant clusterserviceversion cannot be empty")

		return nil, fmt.Errorf(
			"the namePattern field to filter out all relevant clusterserviceversion cannot be empty")
	}

	klog.V(100).Infof("Listing clusterserviceversion filtered by the name pattern %s in %s namespace",
		namePattern, nsname)

	notFilteredCsvList, err := ListClusterServiceVersion(apiClient, nsname, options...)
	if err != nil {
		klog.V(100).Infof("Failed to list all clusterserviceversions in namespace %s due to %s",
			nsname, err.Error())

		return nil, err
	}

	var finalCsvList []*ClusterServiceVersionBuilder

	for _, foundCsv := range notFilteredCsvList {
		if strings.Contains(foundCsv.Definition.Name, namePattern) {
			finalCsvList = append(finalCsvList, foundCsv)
		}
	}

	return finalCsvList, nil
}

// ListClusterServiceVersionInAllNamespaces returns cluster-wide clusterserviceversion inventory.
func ListClusterServiceVersionInAllNamespaces(
	apiClient *clients.Settings,
	options ...runtimeclient.ListOptions) ([]*ClusterServiceVersionBuilder, error) {
	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[oplmV1alpha1.ClusterServiceVersion, oplmV1alpha1.ClusterServiceVersionList,
		ClusterServiceVersionBuilder](context.TODO(), apiClient, oplmV1alpha1.AddToScheme,
		common.ConvertListOptionsToOptions(options)...)
}
