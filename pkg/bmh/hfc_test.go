package bmh

import (
	"testing"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var hfcGVK = bmhv1alpha1.GroupVersion.WithKind("HostFirmwareComponents")

func TestPullHFC(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(PullHFC, bmhv1alpha1.AddToScheme, hfcGVK).ExecuteTests(t)
}

func TestHFCBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[bmhv1alpha1.HostFirmwareComponents, HFCBuilder](
		bmhv1alpha1.AddToScheme,
		hfcGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}
