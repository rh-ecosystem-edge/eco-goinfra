package argocd

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	argocdoperator "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/argocd/argocdoperator"
)

var argoCdGVK = argocdoperator.GroupVersion.WithKind("ArgoCD")

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[argocdoperator.ArgoCD, Builder](
		NewBuilder, argocdoperator.AddToScheme, argoCdGVK,
	).ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[argocdoperator.ArgoCD, Builder](
		Pull, argocdoperator.AddToScheme, argoCdGVK,
	).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[argocdoperator.ArgoCD, Builder](
		argocdoperator.AddToScheme,
		argoCdGVK,
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
