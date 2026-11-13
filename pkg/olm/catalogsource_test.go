package olm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	oplmV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
)

var (
	testSchemes = []clients.SchemeAttacher{
		oplmV1alpha1.AddToScheme,
	}
	catalogSourceGVK = oplmV1alpha1.SchemeGroupVersion.WithKind("CatalogSource")
)

func TestNewCatalogSourceBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewCatalogSourceBuilder,
		oplmV1alpha1.AddToScheme,
		catalogSourceGVK,
	).ExecuteTests(t)
}

func TestPullCatalogSource(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullCatalogSource,
		oplmV1alpha1.AddToScheme,
		catalogSourceGVK,
	).ExecuteTests(t)
}

func TestCatalogSourceBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[oplmV1alpha1.CatalogSource, CatalogSourceBuilder](
		oplmV1alpha1.AddToScheme,
		catalogSourceGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}
