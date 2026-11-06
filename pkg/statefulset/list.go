package statefulset

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// List returns statefulset inventory in the given namespace.
func List(apiClient *clients.Settings, nsname string, options ...metav1.ListOptions) ([]*Builder, error) {
	if nsname == "" {
		klog.V(100).Info("statefulset 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("StatefulSet", "", ""), commonerrors.BuilderFieldNamespace)
	}

	convertedOptions, err := common.ConvertMetaListOptionsToListOptions(options)
	if err != nil {
		return nil, err
	}

	allOptions := append([]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)}, convertedOptions...)

	return common.List[appsv1.StatefulSet, appsv1.StatefulSetList, Builder](
		context.TODO(), apiClient, appsv1.AddToScheme, allOptions...)
}

// ListInAllNamespaces returns statefulset inventory in all namespaces.
func ListInAllNamespaces(apiClient *clients.Settings, options ...metav1.ListOptions) ([]*Builder, error) {
	convertedOptions, err := common.ConvertMetaListOptionsToListOptions(options)
	if err != nil {
		return nil, err
	}

	return common.List[appsv1.StatefulSet, appsv1.StatefulSetList, Builder](
		context.TODO(), apiClient, appsv1.AddToScheme, convertedOptions...)
}
