package oran

import (
	"context"
	"fmt"

	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListClusterTemplates returns a list of ClusterTemplates in all namespaces, using the provided options.
func ListClusterTemplates(
	apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*ClusterTemplateBuilder, error) {
	if len(options) > 1 {
		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[
		provisioningv1alpha1.ClusterTemplate,
		provisioningv1alpha1.ClusterTemplateList,
		ClusterTemplateBuilder,
	](context.TODO(), apiClient, provisioningv1alpha1.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
