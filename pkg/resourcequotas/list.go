package resourcequotas

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// List returns resource quota inventory in the given namespace.
func List(apiClient *clients.Settings, nsname string, options ...metav1.ListOptions) ([]*Builder, error) {
	klog.V(100).Infof("Listing resource quotas in the namespace %s", nsname)

	if nsname == "" {
		klog.V(100).Info("resource quota 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list resource quotas, 'nsname' parameter is empty")
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	convertedOptions, err := common.ConvertMetaListOptionsToListOptions(options)
	if err != nil {
		return nil, err
	}

	allOptions := append([]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)}, convertedOptions...)

	return common.List[corev1.ResourceQuota, corev1.ResourceQuotaList, Builder](
		context.TODO(), apiClient, corev1.AddToScheme, allOptions...)
}
