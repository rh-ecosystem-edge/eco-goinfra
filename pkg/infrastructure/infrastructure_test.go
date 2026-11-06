package infrastructure

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var infrastructureGVK = configv1.GroupVersion.WithKind("Infrastructure")

func TestPullInfrastructure(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		Pull,
		configv1.Install,
		infrastructureGVK,
		infrastructureName,
	).ExecuteTests(t)
}

func TestInfrastructureBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[configv1.Infrastructure, Builder](
		configv1.Install, infrastructureGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		Run(t)
}
