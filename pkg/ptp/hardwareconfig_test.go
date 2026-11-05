package ptp

import (
	"testing"

	ptpv2alpha1 "github.com/k8snetworkplumbingwg/ptp-operator/api/v2alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var hardwareConfigGVK = ptpv2alpha1.GroupVersion.WithKind("HardwareConfig")

func TestPullHardwareConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullHardwareConfig,
		ptpv2alpha1.AddToScheme,
		hardwareConfigGVK,
	).ExecuteTests(t)
}

func TestListHardwareConfigs(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListHardwareConfigs,
		ptpv2alpha1.AddToScheme,
		hardwareConfigGVK,
	).ExecuteTests(t)
}

func TestHardwareConfigMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[ptpv2alpha1.HardwareConfig, HardwareConfigBuilder](
		ptpv2alpha1.AddToScheme,
		hardwareConfigGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}
