package olm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	operatorv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/package-server/operators/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPackageManifest returns PackageManifest inventory in the given namespace.
func ListPackageManifest(
	apiClient *clients.Settings,
	nsname string,
	options ...runtimeclient.ListOptions) ([]*PackageManifestBuilder, error) {
	if nsname == "" {
		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey(operatorv1.PackageManifestKind, "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	allOptions := append([]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)},
		common.ConvertListOptionsToOptions(options)...)

	return common.List[operatorv1.PackageManifest, operatorv1.PackageManifestList, PackageManifestBuilder](
		context.TODO(), apiClient, operatorv1.AddToScheme, allOptions...)
}
