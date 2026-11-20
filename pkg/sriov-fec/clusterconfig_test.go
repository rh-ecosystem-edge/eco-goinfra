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
	defaultClusterConfigName      = "test-cluster-config"
	defaultClusterConfigNamespace = "test-ns"
)

var clusterConfigGVK = sriovfectypes.GroupVersion.WithKind("SriovFecClusterConfig")

func TestFecNewClusterConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(NewClusterConfigBuilder, sriovfectypes.AddToScheme, clusterConfigGVK).
		ExecuteTests(t)
}

func TestFecPullClusterConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(PullClusterConfig, sriovfectypes.AddToScheme, clusterConfigGVK).
		ExecuteTests(t)
}

func TestFecClusterConfigBuilderMethods(t *testing.T) {
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

func TestFecClusterConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newClusterConfigCommonTestConfig()).ExecuteTests(t)
}

func TestFecClusterConfigWithOptionsMutatesSpec(t *testing.T) {
	t.Parallel()

	builder := NewClusterConfigBuilder(
		clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: []clients.SchemeAttacher{sriovfectypes.AddToScheme},
		}),
		defaultClusterConfigName,
		defaultClusterConfigNamespace,
	)

	result := builder.WithOptions(func(builder *ClusterConfigBuilder) (*ClusterConfigBuilder, error) {
		builder.Definition.Spec.Priority = 10

		return builder, nil
	})

	require.Same(t, builder, result)
	assert.NoError(t, result.GetError())
	assert.Equal(t, 10, result.Definition.Spec.Priority)
}

func TestFecClusterConfigGetGVR(t *testing.T) {
	t.Parallel()

	gvr := GetSriovFecClusterConfigIoGVR()
	assert.Equal(t, APIGroup, gvr.Group)
	assert.Equal(t, APIVersion, gvr.Version)
	assert.Equal(t, ClusterConfigsResource, gvr.Resource)
}

func newClusterConfigCommonTestConfig() testhelper.CommonTestConfig[
	sriovfectypes.SriovFecClusterConfig,
	ClusterConfigBuilder,
	*sriovfectypes.SriovFecClusterConfig,
	*ClusterConfigBuilder,
] {
	return testhelper.NewCommonTestConfig[sriovfectypes.SriovFecClusterConfig, ClusterConfigBuilder](
		sriovfectypes.AddToScheme,
		clusterConfigGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

// buildDummyClusterConfig returns a ClusterConfig with the provided name and namespace.
func buildDummyClusterConfig(name, nsname string) *sriovfectypes.SriovFecClusterConfig {
	return &sriovfectypes.SriovFecClusterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		Spec: sriovfectypes.SriovFecClusterConfigSpec{
			Priority: 1,
		},
	}
}
