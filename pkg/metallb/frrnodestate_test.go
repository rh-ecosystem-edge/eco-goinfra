package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
)

var defaultNodeName = "worker-0"

func TestNewFrrNodeStateBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		newFRRNodeStateBuilder, frrtypes.AddToScheme, frrNodeStateGVK).ExecuteTests(t)
}

func TestPullFrrNodeState(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullFrrNodeState, frrtypes.AddToScheme, frrNodeStateGVK).ExecuteTests(t)
}

func TestFrrNodeStateBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[frrtypes.FRRNodeState, FrrNodeStateBuilder](
		frrtypes.AddToScheme,
		frrNodeStateGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}

// newFRRNodeStateBuilder creates a new FRRNodeState builder for tests.
func newFRRNodeStateBuilder(apiClient *clients.Settings, name string) *FrrNodeStateBuilder {
	return common.NewClusterScopedBuilder[frrtypes.FRRNodeState, FrrNodeStateBuilder](
		apiClient, frrtypes.AddToScheme, name)
}
