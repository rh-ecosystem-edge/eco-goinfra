package ocm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

// buildDummyPlacementRule returns a PlacementRule with the provided name and namespace.
func buildDummyPlacementRule(name, nsname string) *placementrulev1.PlacementRule {
	return &placementrulev1.PlacementRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyPlacementRule returns a client with a mock dummy PlacementRule.
func buildTestClientWithDummyPlacementRule() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPlacementRule(defaultPlacementRuleName, defaultPlacementRuleNsName),
		},
		SchemeAttachers: []clients.SchemeAttacher{
			placementrulev1.AddToScheme,
		},
	})
}
