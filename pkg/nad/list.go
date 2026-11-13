package nad

import (
	"context"

	nadV1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// List returns NADs inventory in the given namespace.
func List(apiClient *clients.Settings, nsname string) ([]*Builder, error) {
	if nsname == "" {
		klog.V(100).Info("nad 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("NetworkAttachmentDefinition", "", ""), commonerrors.BuilderFieldNamespace)
	}

	return common.List[nadV1.NetworkAttachmentDefinition, nadV1.NetworkAttachmentDefinitionList, Builder](
		context.TODO(), apiClient, nadV1.AddToScheme, runtimeclient.InNamespace(nsname))
}
