package sriovfec

import (
	"context"
	"errors"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	sriovfectypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/fectypes"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListClusterConfig returns SriovFecClusterConfigList from given namespace.
func ListClusterConfig(
	apiClient *clients.Settings,
	nsname string,
	options ...runtimeclient.ListOptions) ([]*ClusterConfigBuilder, error) {
	if nsname == "" {
		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("SriovFecClusterConfig", "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		return nil, errors.New("error: more than one ListOptions was passed")
	}

	allOptions := append(common.ConvertListOptionsToOptions(options), runtimeclient.InNamespace(nsname))

	return common.List[sriovfectypes.SriovFecClusterConfig, sriovfectypes.SriovFecClusterConfigList, ClusterConfigBuilder](
		context.TODO(), apiClient, sriovfectypes.AddToScheme, allOptions...)
}
