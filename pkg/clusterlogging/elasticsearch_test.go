package clusterlogging

import (
	"testing"

	eskv1 "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultElasticsearchName      = "elasticsearch"
	defaultElasticsearchNamespace = "openshift-logging"
)

func TestNewElasticsearchBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewElasticsearchBuilder, eskv1.AddToScheme, elasticsearchGVK).ExecuteTests(t)
}

func TestPullElasticsearch(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullElasticsearch, eskv1.AddToScheme, elasticsearchGVK).ExecuteTests(t)
}

func TestElasticsearchBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[eskv1.Elasticsearch, ElasticsearchBuilder](
		eskv1.AddToScheme,
		elasticsearchGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestElasticsearchWithManagementState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		managementState eskv1.ManagementState
		valid           bool
	}{
		{
			name:            "managed state",
			managementState: eskv1.ManagementStateManaged,
			valid:           true,
		},
		{
			name:            "unmanaged state",
			managementState: eskv1.ManagementStateUnmanaged,
			valid:           true,
		},
		{
			name:            "invalid builder short-circuits",
			managementState: eskv1.ManagementStateManaged,
			valid:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildElasticsearchTestClient()

			builder := buildValidElasticsearchTestBuilder(apiClient)
			if !testCase.valid {
				builder = buildInvalidElasticsearchTestBuilder(apiClient)
			}

			result := builder.WithManagementState(testCase.managementState)

			require.Same(t, builder, result)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.ManagementState)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.managementState, result.Definition.Spec.ManagementState)
		})
	}
}

func TestElasticsearchGetManagementState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		managementState eskv1.ManagementState
		objectExists    bool
		valid           bool
	}{
		{
			name:            "managed state",
			managementState: eskv1.ManagementStateManaged,
			objectExists:    true,
			valid:           true,
		},
		{
			name:            "unmanaged state",
			managementState: eskv1.ManagementStateUnmanaged,
			objectExists:    true,
			valid:           true,
		},
		{
			name:         "resource does not exist",
			objectExists: false,
			valid:        true,
		},
		{
			name:            "invalid builder short-circuits",
			managementState: eskv1.ManagementStateManaged,
			objectExists:    true,
			valid:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object

			if testCase.objectExists {
				elasticsearch := buildDummyElasticsearch()
				elasticsearch.Spec.ManagementState = testCase.managementState
				objects = append(objects, elasticsearch)
			}

			apiClient := buildElasticsearchTestClient(objects...)

			builder := buildValidElasticsearchTestBuilder(apiClient)
			if !testCase.valid {
				builder = buildInvalidElasticsearchTestBuilder(apiClient)
			}

			managementState, err := builder.GetManagementState()

			switch {
			case !testCase.valid:
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
				assert.Nil(t, managementState)
			case !testCase.objectExists:
				assert.EqualError(t, err, "elasticsearch object does not exist")
				assert.Nil(t, managementState)
			default:
				require.NoError(t, err)
				require.NotNil(t, managementState)
				assert.Equal(t, testCase.managementState, *managementState)
			}
		})
	}
}

func buildValidElasticsearchTestBuilder(apiClient *clients.Settings) *ElasticsearchBuilder {
	return NewElasticsearchBuilder(apiClient, defaultElasticsearchName, defaultElasticsearchNamespace)
}

func buildInvalidElasticsearchTestBuilder(apiClient *clients.Settings) *ElasticsearchBuilder {
	return NewElasticsearchBuilder(apiClient, "", defaultElasticsearchNamespace)
}

func buildElasticsearchTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{eskv1.AddToScheme},
	})
}

func buildDummyElasticsearch() *eskv1.Elasticsearch {
	return &eskv1.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultElasticsearchName,
			Namespace: defaultElasticsearchNamespace,
		},
	}
}
