package cgu

import (
	"testing"

	"github.com/openshift-kni/cluster-group-upgrades-operator/pkg/api/clustergroupupgrades/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var preCachingConfigGVK = v1alpha1.SchemeGroupVersion.WithKind("PreCachingConfig")

func TestNewPreCachingConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewPreCachingConfigBuilder,
		v1alpha1.AddToScheme,
		preCachingConfigGVK,
	).ExecuteTests(t)
}

func TestPullPreCachingConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPreCachingConfig,
		v1alpha1.AddToScheme,
		preCachingConfigGVK,
	).ExecuteTests(t)
}

func TestPreCachingConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[v1alpha1.PreCachingConfig, PreCachingConfigBuilder](
		v1alpha1.AddToScheme,
		preCachingConfigGVK,
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
