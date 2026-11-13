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

// List returns SriovFecNodeConfigList. The supplied namespace is validated but, for backward compatibility, is not
// added to the list options; callers can scope the request with client.InNamespace.
func List(apiClient *clients.Settings, nsname string, options ...runtimeclient.ListOptions) ([]*NodeConfigBuilder, error) {
	if nsname == "" {
		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("SriovFecNodeConfig", "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		return nil, errors.New("error: more than one ListOptions was passed")
	}

	return common.List[sriovfectypes.SriovFecNodeConfig, sriovfectypes.SriovFecNodeConfigList, NodeConfigBuilder](
		context.TODO(), apiClient, sriovfectypes.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
