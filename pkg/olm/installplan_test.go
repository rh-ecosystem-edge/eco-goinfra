package olm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	oplmV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var installPlanGVK = oplmV1alpha1.SchemeGroupVersion.WithKind("InstallPlan")

func TestNewInstallPlanBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewInstallPlanBuilder,
		oplmV1alpha1.AddToScheme,
		installPlanGVK,
	).ExecuteTests(t)
}

func TestPullInstallPlan(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullInstallPlan,
		oplmV1alpha1.AddToScheme,
		installPlanGVK,
	).ExecuteTests(t)
}

func TestInstallPlanBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[oplmV1alpha1.InstallPlan, InstallPlanBuilder](
		oplmV1alpha1.AddToScheme,
		installPlanGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func buildDummyInstallPlan() []runtime.Object {
	return []runtime.Object{&oplmV1alpha1.InstallPlan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceInstallPlan,
			Namespace: "test-namespace",
		},
		Spec: oplmV1alpha1.InstallPlanSpec{
			CatalogSource: "test",
		},
	}}
}
