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

// ListClusterConfig returns SriovVrbClusterConfigList from given namespace.
func ListClusterConfig(
	apiClient *clients.Settings,
	nsname string,
	options ...client.ListOptions) ([]*ClusterConfigBuilder, error) {
	if len(options) > 1 {
		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if nsname == "" {
		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("SriovVrbClusterConfig", "", ""), commonerrors.BuilderFieldNamespace)
	}

	listOptions := common.ConvertListOptionsToOptions(options)
	listOptions = append(listOptions, client.InNamespace(nsname))

	return common.List[sriovvrbtypes.SriovVrbClusterConfig, sriovvrbtypes.SriovVrbClusterConfigList, ClusterConfigBuilder](
		context.TODO(), apiClient, sriovvrbtypes.AddToScheme, listOptions...)
}
