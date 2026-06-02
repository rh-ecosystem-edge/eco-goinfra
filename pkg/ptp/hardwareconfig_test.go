package ptp

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	ptpv2alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ptp/v2alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
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

	testhelper.NewListTestConfig[ptpv2alpha1.HardwareConfig, HardwareConfigBuilder](
		func(apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*HardwareConfigBuilder, error) {
			if len(options) == 0 {
				return ListHardwareConfigs(apiClient)
			}

			return ListHardwareConfigs(apiClient, &options[0])
		},
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
