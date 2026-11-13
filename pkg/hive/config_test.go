package hive

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	hiveV1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/hive/api/v1"
)

var configGVK = hiveV1.SchemeGroupVersion.WithKind("HiveConfig")

func TestNewConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		NewConfigBuilder, hiveV1.AddToScheme, configGVK).ExecuteTests(t)
}

func TestPullConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullConfig, hiveV1.AddToScheme, configGVK).ExecuteTests(t)
}

func TestConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[hiveV1.HiveConfig, ConfigBuilder](
		hiveV1.AddToScheme,
		configGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[hiveV1.HiveConfig, ConfigBuilder](
			hiveV1.AddToScheme,
			configGVK,
			testhelper.ResourceScopeClusterScoped,
		),
	).ExecuteTests(t)
}
