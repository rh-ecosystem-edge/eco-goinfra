package sriovvrb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	sriovvrbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/vrbtypes"
	"github.com/stretchr/testify/assert"
)

const (
	defaultClusterConfigName      = "test-cluster-config"
	defaultClusterConfigNamespace = "test-ns"
)

var clusterConfigGVK = sriovvrbtypes.GroupVersion.WithKind("SriovVrbClusterConfig")

func TestVrbNewClusterConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewClusterConfigBuilder,
		sriovvrbtypes.AddToScheme,
		clusterConfigGVK,
	).ExecuteTests(t)
}

func TestVrbPullClusterConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullClusterConfig,
		sriovvrbtypes.AddToScheme,
		clusterConfigGVK,
	).ExecuteTests(t)
}

func TestVrbClusterConfigMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newClusterConfigCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestVrbClusterConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newClusterConfigCommonTestConfig()).ExecuteTests(t)
}

func TestVrbClusterConfigWithOptionsMutatesSpec(t *testing.T) {
	t.Parallel()

	builder := buildValidClusterConfigTestBuilder(buildClusterConfigTestClient())
	builder = builder.WithOptions(func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error) {
		builder.Definition.Spec.Priority = 10

		return builder, nil
	})

	assert.NoError(t, builder.GetError())
	assert.Equal(t, 10, builder.Definition.Spec.Priority)

	builder = buildInvalidClusterConfigTestBuilder(buildClusterConfigTestClient())
	builder = builder.WithOptions(func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error) {
		builder.Definition.Spec.Priority = 10

		return builder, nil
	})

	assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
	assert.Zero(t, builder.Definition.Spec.Priority)
}

func TestVrbClusterConfigGetGVR(t *testing.T) {
	t.Parallel()

	gvr := GetSriovVrbClusterConfigIoGVR()
	assert.Equal(t, APIGroup, gvr.Group)
	assert.Equal(t, APIVersion, gvr.Version)
	assert.Equal(t, ClusterConfigsResource, gvr.Resource)
}

// newClusterConfigCommonTestConfig returns the shared test configuration for SriovVrbClusterConfig.
func newClusterConfigCommonTestConfig() testhelper.CommonTestConfig[
	sriovvrbtypes.SriovVrbClusterConfig, ClusterConfigBuilder, *sriovvrbtypes.SriovVrbClusterConfig,
	*ClusterConfigBuilder,
] {
	return testhelper.NewCommonTestConfig[
		sriovvrbtypes.SriovVrbClusterConfig, ClusterConfigBuilder,
	](sriovvrbtypes.AddToScheme, clusterConfigGVK, testhelper.ResourceScopeNamespaced)
}

// buildClusterConfigTestClient returns a test client with the SriovVrbClusterConfig scheme attached.
func buildClusterConfigTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{sriovvrbtypes.AddToScheme},
	})
}

// buildValidClusterConfigTestBuilder returns a valid SriovVrbClusterConfig builder for testing.
func buildValidClusterConfigTestBuilder(apiClient *clients.Settings) *ClusterConfigBuilder {
	return NewClusterConfigBuilder(apiClient, defaultClusterConfigName, defaultClusterConfigNamespace)
}

// buildInvalidClusterConfigTestBuilder returns a SriovVrbClusterConfig builder with an empty namespace.
func buildInvalidClusterConfigTestBuilder(apiClient *clients.Settings) *ClusterConfigBuilder {
	return NewClusterConfigBuilder(apiClient, defaultClusterConfigName, "")
}
