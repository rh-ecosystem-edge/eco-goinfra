package ocm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	placementrulev1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/placementrule/v1"
)

const (
	defaultPlacementRuleName   = "placementrule-test"
	defaultPlacementRuleNsName = "test-ns"
)

func TestNewPlacementRuleBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewPlacementRuleBuilder, placementrulev1.AddToScheme, placementRuleGVK).ExecuteTests(t)
}

func TestPullPlacementRule(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPlacementRule, placementrulev1.AddToScheme, placementRuleGVK).ExecuteTests(t)
}

func TestPlacementRuleBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[placementrulev1.PlacementRule, PlacementRuleBuilder](
		placementrulev1.AddToScheme,
		placementRuleGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}
