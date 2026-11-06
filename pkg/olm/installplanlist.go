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

// ListInstallPlan returns a list of installplans found for specific namespace.
func ListInstallPlan(
	apiClient *clients.Settings, nsname string, options ...runtimeclient.ListOptions) ([]*InstallPlanBuilder, error) {
	if nsname == "" {
		klog.V(100).Info("The nsname of the installplan is empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("InstallPlan", "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	allOptions := append(
		[]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)},
		common.ConvertListOptionsToOptions(options)...,
	)

	installPlanObjects, err := common.List[
		oplmV1alpha1.InstallPlan, oplmV1alpha1.InstallPlanList, InstallPlanBuilder](
		context.TODO(), apiClient, oplmV1alpha1.AddToScheme, allOptions...)
	if err != nil {
		klog.V(100).Infof("Failed to list all installplans in namespace %s due to %v", nsname, err)

		return nil, err
	}

	if len(installPlanObjects) == 0 {
		return nil, fmt.Errorf("installplan not found in namespace %s", nsname)
	}

	return installPlanObjects, nil
}
