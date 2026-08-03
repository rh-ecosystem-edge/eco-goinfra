package bmh

import (
	"testing"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var hardwareDataGVK = bmhv1alpha1.GroupVersion.WithKind("HardwareData")

func TestPullHardwareData(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullHardwareData,
		bmhv1alpha1.AddToScheme,
		hardwareDataGVK,
	).ExecuteTests(t)
}

func TestListHardwareData(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListHardwareData,
		bmhv1alpha1.AddToScheme,
		hardwareDataGVK,
	).ExecuteTests(t)
}

func TestHardwareDataMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[bmhv1alpha1.HardwareData, HardwareDataBuilder](
		bmhv1alpha1.AddToScheme,
		hardwareDataGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}
