package siteconfig

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	siteconfigv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/siteconfig/v1alpha1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListClusterInstances returns a list of ClusterInstances in all namespaces, using the provided options.
func ListClusterInstances(
	apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*CIBuilder, error) {
	if len(options) > 1 {
		klog.V(100).Info("ClusterInstances 'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[
		siteconfigv1alpha1.ClusterInstance,
		siteconfigv1alpha1.ClusterInstanceList,
		CIBuilder,
	](context.TODO(), apiClient, siteconfigv1alpha1.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
