package bmh

import (
	"testing"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var dataImageGVK = bmhv1alpha1.GroupVersion.WithKind("DataImage")

func TestPullDataImage(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullDataImage,
		bmhv1alpha1.AddToScheme,
		dataImageGVK,
	).ExecuteTests(t)
}

func TestDataImageMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[bmhv1alpha1.DataImage, DataImageBuilder](
		bmhv1alpha1.AddToScheme,
		dataImageGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		Run(t)
}
