package proxy

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var proxyGVK = schema.GroupVersionKind{
	Group:   configv1.GroupVersion.Group,
	Version: configv1.GroupVersion.Version,
	Kind:    "Proxy",
}

func TestPullProxy(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		Pull,
		configv1.Install,
		proxyGVK,
		clusterProxyName,
	).ExecuteTests(t)
}

func TestProxyBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[configv1.Proxy, Builder](
		configv1.Install,
		proxyGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}
