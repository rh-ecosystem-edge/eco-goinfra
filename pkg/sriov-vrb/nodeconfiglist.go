package sriovvrb

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	sriovvrbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/vrbtypes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListNodeConfig returns SriovVrbNodeConfigList from given namespace.
func ListNodeConfig(
	apiClient *clients.Settings,
	nsname string,
	options ...client.ListOptions) ([]*NodeConfigBuilder, error) {
	if len(options) > 1 {
		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if nsname == "" {
		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("SriovVrbNodeConfig", "", ""), commonerrors.BuilderFieldNamespace)
	}

	listOptions := common.ConvertListOptionsToOptions(options)
	listOptions = append(listOptions, client.InNamespace(nsname))

	return common.List[sriovvrbtypes.SriovVrbNodeConfig, sriovvrbtypes.SriovVrbNodeConfigList, NodeConfigBuilder](
		context.TODO(), apiClient, sriovvrbtypes.AddToScheme, listOptions...)
}
