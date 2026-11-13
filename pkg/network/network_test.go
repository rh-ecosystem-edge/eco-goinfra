package network

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var networkConfigGVK = configv1.GroupVersion.WithKind("Network")

func TestPullConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		PullConfig,
		configv1.Install,
		networkConfigGVK,
		clusterNetworkName,
	).ExecuteTests(t)
}

func TestConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[configv1.Network, ConfigBuilder](
		configv1.Install, networkConfigGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		Run(t)
}
