package ocm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	operatorv1 "open-cluster-management.io/api/operator/v1"
)

func TestNewKlusterletBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		NewKlusterletBuilder, operatorv1.Install, klusterletGVK).ExecuteTests(t)
}

func TestPullKlusterlet(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullKlusterlet, operatorv1.Install, klusterletGVK).ExecuteTests(t)
}

func TestKlusterletBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[operatorv1.Klusterlet, KlusterletBuilder](
		operatorv1.Install,
		klusterletGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}
