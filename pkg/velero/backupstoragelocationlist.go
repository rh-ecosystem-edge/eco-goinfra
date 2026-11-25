package velero

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListBackupStorageLocationBuilder returns backupstoragelocation inventory in the given namespace.
func ListBackupStorageLocationBuilder(
	apiClient *clients.Settings, nsname string, options ...runtimeclient.ListOptions) ([]*BackupStorageLocationBuilder, error) {
	if nsname == "" {
		klog.V(100).Info("backupstoragelocation 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("BackupStorageLocation", "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	allOptions := append(
		[]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)},
		common.ConvertListOptionsToOptions(options)...,
	)

	return common.List[velerov1.BackupStorageLocation, velerov1.BackupStorageLocationList, BackupStorageLocationBuilder](
		context.TODO(), apiClient, velerov1.AddToScheme, allOptions...)
}
