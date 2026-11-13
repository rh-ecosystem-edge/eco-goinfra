package olm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	oplmV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListCatalogSources returns catalogsource inventory in the given namespace.
func ListCatalogSources(
	apiClient *clients.Settings,
	nsname string,
	options ...runtimeclient.ListOptions) ([]*CatalogSourceBuilder, error) {
	if nsname == "" {
		klog.V(100).Info("catalogsource 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("CatalogSource", "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	allOptions := append(
		[]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)},
		common.ConvertListOptionsToOptions(options)...,
	)

	return common.List[oplmV1alpha1.CatalogSource, oplmV1alpha1.CatalogSourceList, CatalogSourceBuilder](
		context.TODO(), apiClient, oplmV1alpha1.AddToScheme, allOptions...)
}
