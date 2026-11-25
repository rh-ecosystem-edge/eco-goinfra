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
	defaultNodeConfigName      = "test-node-config"
	defaultNodeConfigNamespace = "test-ns"
)

var nodeConfigGVK = sriovvrbtypes.GroupVersion.WithKind("SriovVrbNodeConfig")

func TestVrbNewNodeConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewNodeConfigBuilder,
		sriovvrbtypes.AddToScheme,
		nodeConfigGVK,
	).ExecuteTests(t)
}

func TestVrbPullNodeConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullNodeConfig,
		sriovvrbtypes.AddToScheme,
		nodeConfigGVK,
	).ExecuteTests(t)
}

func TestVrbNodeConfigMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newNodeConfigCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestVrbNodeConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newNodeConfigCommonTestConfig()).ExecuteTests(t)
}

func TestVrbNodeConfigWithOptionsMutatesSpec(t *testing.T) {
	t.Parallel()

	builder := buildValidNodeConfigTestBuilder(buildNodeConfigTestClient())
	builder = builder.WithOptions(func(builder *NodeConfigBuilder) (*NodeConfigBuilder, error) {
		builder.Definition.Spec.DrainSkip = true

		return builder, nil
	})

	assert.NoError(t, builder.GetError())
	assert.True(t, builder.Definition.Spec.DrainSkip)

	builder = buildInvalidNodeConfigTestBuilder(buildNodeConfigTestClient())
	builder = builder.WithOptions(func(builder *NodeConfigBuilder) (*NodeConfigBuilder, error) {
		builder.Definition.Spec.DrainSkip = true

		return builder, nil
	})

	assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
	assert.False(t, builder.Definition.Spec.DrainSkip)
}

func TestVrbNodeConfigGetGVR(t *testing.T) {
	t.Parallel()

	gvr := GetSriovVrbNodeConfigIoGVR()
	assert.Equal(t, APIGroup, gvr.Group)
	assert.Equal(t, APIVersion, gvr.Version)
	assert.Equal(t, NodeConfigsResource, gvr.Resource)
}

// newNodeConfigCommonTestConfig returns the shared test configuration for SriovVrbNodeConfig.
func newNodeConfigCommonTestConfig() testhelper.CommonTestConfig[
	sriovvrbtypes.SriovVrbNodeConfig, NodeConfigBuilder, *sriovvrbtypes.SriovVrbNodeConfig, *NodeConfigBuilder,
] {
	return testhelper.NewCommonTestConfig[
		sriovvrbtypes.SriovVrbNodeConfig, NodeConfigBuilder,
	](sriovvrbtypes.AddToScheme, nodeConfigGVK, testhelper.ResourceScopeNamespaced)
}

// buildNodeConfigTestClient returns a test client with the SriovVrbNodeConfig scheme attached.
func buildNodeConfigTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{sriovvrbtypes.AddToScheme},
	})
}

// buildValidNodeConfigTestBuilder returns a valid SriovVrbNodeConfig builder for testing.
func buildValidNodeConfigTestBuilder(apiClient *clients.Settings) *NodeConfigBuilder {
	return NewNodeConfigBuilder(apiClient, defaultNodeConfigName, defaultNodeConfigNamespace)
}

// buildInvalidNodeConfigTestBuilder returns a SriovVrbNodeConfig builder with an empty namespace.
func buildInvalidNodeConfigTestBuilder(apiClient *clients.Settings) *NodeConfigBuilder {
	return NewNodeConfigBuilder(apiClient, defaultNodeConfigName, "")
}
