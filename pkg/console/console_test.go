package console

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var consoleGVK = configv1.GroupVersion.WithKind("Console")

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(NewBuilder, configv1.Install, consoleGVK).ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(Pull, configv1.Install, consoleGVK).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[configv1.Console, Builder](
		configv1.Install,
		consoleGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}
