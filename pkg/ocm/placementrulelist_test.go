package ocm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	placementrulev1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/placementrule/v1"
)

func TestListPlacementrulesInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListPlacementrulesInAllNamespaces,
		placementrulev1.AddToScheme,
		placementRuleGVK,
	).ExecuteTests(t)
}
