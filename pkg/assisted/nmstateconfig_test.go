package assisted

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	assistedv1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestNewNmStateConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[assistedv1beta1.NMStateConfig, NmStateConfigBuilder](
		NewNmStateConfigBuilder,
		assistedv1beta1.AddToScheme,
		nmStateConfigGVK,
	).ExecuteTests(t)
}

func TestNmStateConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newNmStateConfigCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		Run(t)
}

func TestListNmStateConfigsInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		func(apiClient *clients.Settings, _ ...runtimeclient.ListOptions) ([]*NmStateConfigBuilder, error) {
			return ListNmStateConfigsInAllNamespaces(apiClient)
		},
		assistedv1beta1.AddToScheme,
		nmStateConfigGVK,
	).ExecuteTests(t)
}

func TestListNmStateConfigs(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*NmStateConfigBuilder, error) {
			return ListNmStateConfigs(apiClient, nsname)
		},
		assistedv1beta1.AddToScheme,
		nmStateConfigGVK,
	).ExecuteTests(t)
}

func newNmStateConfigCommonTestConfig() testhelper.CommonTestConfig[
	assistedv1beta1.NMStateConfig, NmStateConfigBuilder, *assistedv1beta1.NMStateConfig, *NmStateConfigBuilder,
] {
	return testhelper.NewCommonTestConfig[assistedv1beta1.NMStateConfig, NmStateConfigBuilder](
		assistedv1beta1.AddToScheme,
		nmStateConfigGVK,
		testhelper.ResourceScopeNamespaced,
	)
}
