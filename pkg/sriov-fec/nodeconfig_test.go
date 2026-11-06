package sriovfec

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	sriovfectypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/fec/fectypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultNodeConfigName      = "test-node-config"
	defaultNodeConfigNamespace = "test-ns"
)

var nodeConfigGVK = sriovfectypes.GroupVersion.WithKind("SriovFecNodeConfig")

func TestFecNewNodeConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		func(apiClient *clients.Settings, name, nsname string) *NodeConfigBuilder {
			return NewNodeConfigBuilder(apiClient, name, nsname, nil)
		},
		sriovfectypes.AddToScheme,
		nodeConfigGVK,
	).ExecuteTests(t)
}

func TestFecPullNodeConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(Pull, sriovfectypes.AddToScheme, nodeConfigGVK).ExecuteTests(t)
}

func TestFecNodeConfigBuilderMethods(t *testing.T) {
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

func TestFecNodeConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newNodeConfigCommonTestConfig()).ExecuteTests(t)
}

func TestFecNodeConfigWithOptionsMutatesSpec(t *testing.T) {
	t.Parallel()

	builder := NewNodeConfigBuilder(
		clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: []clients.SchemeAttacher{sriovfectypes.AddToScheme},
		}),
		defaultNodeConfigName,
		defaultNodeConfigNamespace,
		nil,
	)

	result := builder.WithOptions(func(builder *NodeConfigBuilder) (*NodeConfigBuilder, error) {
		builder.Definition.Spec.DrainSkip = true

		return builder, nil
	})

	require.Same(t, builder, result)
	assert.NoError(t, result.GetError())
	assert.True(t, result.Definition.Spec.DrainSkip)
}

func TestFecNodeConfigGetGVR(t *testing.T) {
	t.Parallel()

	gvr := GetSriovFecNodeConfigIoGVR()
	assert.Equal(t, APIGroup, gvr.Group)
	assert.Equal(t, APIVersion, gvr.Version)
	assert.Equal(t, NodeConfigsResource, gvr.Resource)
}

func newNodeConfigCommonTestConfig() testhelper.CommonTestConfig[
	sriovfectypes.SriovFecNodeConfig,
	NodeConfigBuilder,
	*sriovfectypes.SriovFecNodeConfig,
	*NodeConfigBuilder,
] {
	return testhelper.NewCommonTestConfig[sriovfectypes.SriovFecNodeConfig, NodeConfigBuilder](
		sriovfectypes.AddToScheme,
		nodeConfigGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

// buildDummyNodeConfig returns a NodeConfig with the provided name and namespace.
func buildDummyNodeConfig(name, nsname string) *sriovfectypes.SriovFecNodeConfig {
	return &sriovfectypes.SriovFecNodeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}
