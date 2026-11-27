package ptp

import (
	"context"
	"fmt"

	ptpv1 "github.com/k8snetworkplumbingwg/ptp-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPtpConfigs returns a list of PtpConfigs in all namespaces, using the provided options.
func ListPtpConfigs(
	apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*PtpConfigBuilder, error) {
	if len(options) > 1 {
		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[ptpv1.PtpConfig, ptpv1.PtpConfigList, PtpConfigBuilder](
		context.TODO(), apiClient, ptpv1.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
