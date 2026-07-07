package namespace

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

var namespaceGVK = corev1.SchemeGroupVersion.WithKind("Namespace")

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(NewBuilder, corev1.AddToScheme, namespaceGVK).
		ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(Pull, corev1.AddToScheme, namespaceGVK).
		ExecuteTests(t)
}

func TestList(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		List,
		corev1.AddToScheme,
		namespaceGVK,
	).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newNamespaceCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newNamespaceCommonTestConfig()).ExecuteTests(t)
}

func TestWithLabel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		key           string
		value         string
		expectedError string
		builder       func() *Builder
	}{
		{
			name:    "valid label",
			key:     "test-key",
			value:   "test-value",
			builder: func() *Builder { return buildValidNamespaceTestBuilder(newNamespaceTestClient()) },
		},
		{
			name:          "empty key",
			key:           "",
			value:         "test-value",
			expectedError: "'key' cannot be empty",
			builder:       func() *Builder { return buildValidNamespaceTestBuilder(newNamespaceTestClient()) },
		},
		{
			name:    "invalid builder short circuits",
			key:     "test-key",
			value:   "test-value",
			builder: func() *Builder { return buildInvalidNamespaceTestBuilder(newNamespaceTestClient()) },
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := testCase.builder()
			require.NotNil(t, testBuilder)

			result := testBuilder.WithLabel(testCase.key, testCase.value)
			require.Same(t, testBuilder, result)

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)
			} else if result.GetError() == nil {
				assert.Equal(t, testCase.value, result.Definition.Labels[testCase.key])
			}
		})
	}
}

func TestWithMultipleLabels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		labels map[string]string
	}{
		{
			name:   "valid labels",
			labels: map[string]string{"key1": "value1", "key2": "value2"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidNamespaceTestBuilder(newNamespaceTestClient())
			result := testBuilder.WithMultipleLabels(testCase.labels)
			require.Same(t, testBuilder, result)
			assert.Nil(t, result.GetError())

			for k, v := range testCase.labels {
				assert.Equal(t, v, result.Definition.Labels[k])
			}
		})
	}
}

func TestNamespaceRemoveLabels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		labels        map[string]string
		expectedError string
		builder       func() *Builder
	}{
		{
			name:   "valid remove",
			labels: map[string]string{"key1": "value1"},
			builder: func() *Builder {
				return buildValidNamespaceTestBuilder(newNamespaceTestClient()).
					WithMultipleLabels(map[string]string{"key1": "value1"})
			},
		},
		{
			name:          "empty labels",
			labels:        map[string]string{},
			expectedError: "labels to be removed cannot be empty",
			builder:       func() *Builder { return buildValidNamespaceTestBuilder(newNamespaceTestClient()) },
		},
		{
			name:   "invalid builder short circuits",
			labels: map[string]string{"key1": "value1"},
			builder: func() *Builder {
				return buildInvalidNamespaceTestBuilder(newNamespaceTestClient())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := testCase.builder()
			require.NotNil(t, testBuilder)

			result := testBuilder.RemoveLabels(testCase.labels)
			require.Same(t, testBuilder, result)

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)
			} else if result.GetError() == nil {
				assert.Equal(t, 0, len(result.Definition.Labels))
			}
		})
	}
}

func newNamespaceCommonTestConfig() testhelper.CommonTestConfig[corev1.Namespace, Builder, *corev1.Namespace, *Builder] {
	return testhelper.NewCommonTestConfig[corev1.Namespace, Builder](
		corev1.AddToScheme, namespaceGVK, testhelper.ResourceScopeClusterScoped)
}

func newNamespaceTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
}

func buildValidNamespaceTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "test-namespace")
}

func buildInvalidNamespaceTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "")
}
