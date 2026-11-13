package deployment

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// List returns deployment inventory in the given namespace.
func List(apiClient *clients.Settings, nsname string, options ...metav1.ListOptions) ([]*Builder, error) {
	if nsname == "" {
		klog.V(100).Info("deployment 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("Deployment", "", ""), commonerrors.BuilderFieldNamespace)
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

	return common.List[appsv1.Deployment, appsv1.DeploymentList, Builder](
		context.TODO(), apiClient, appsv1.AddToScheme, allOptions...)
}

// ListInAllNamespaces returns deployment inventory in the all the namespaces.
func ListInAllNamespaces(apiClient *clients.Settings, options ...metav1.ListOptions) ([]*Builder, error) {
	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be either empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	convertedOptions, err := common.ConvertMetaListOptionsToListOptions(options)
	if err != nil {
		return nil, err
	}

	return common.List[appsv1.Deployment, appsv1.DeploymentList, Builder](
		context.TODO(), apiClient, appsv1.AddToScheme, convertedOptions...)
}
