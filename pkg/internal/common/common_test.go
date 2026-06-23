package common_test

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// testSchemeAttacher is a valid scheme attacher for testing.
	testSchemeAttacher clients.SchemeAttacher = corev1.AddToScheme

	// clusterScopedGVK is the GVK for cluster-scoped test resources.
	clusterScopedGVK = corev1.SchemeGroupVersion.WithKind("Namespace")
	// namespacedGVK is the GVK for namespaced test resources.
	namespacedGVK = corev1.SchemeGroupVersion.WithKind("ConfigMap")
)

func TestNewClusterScopedBuilder(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.Namespace, mockClusterScopedBuilder](
		testSchemeAttacher, clusterScopedGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewGenericClusterScopedBuilderTestConfig(commonConfig, common.NewClusterScopedBuilder).ExecuteTests(t)
}

func TestNewNamespacedBuilder(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.ConfigMap, mockNamespacedBuilder](
		testSchemeAttacher, namespacedGVK, testhelper.ResourceScopeNamespaced)

	testhelper.NewGenericNamespacedBuilderTestConfig(commonConfig, common.NewNamespacedBuilder).ExecuteTests(t)
}

func TestPullClusterScopedBuilder(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.Namespace, mockClusterScopedBuilder](
		testSchemeAttacher, clusterScopedGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewGenericClusterScopedPullTestConfig(commonConfig, common.PullClusterScopedBuilder).ExecuteTests(t)
}

func TestPullNamespacedBuilder(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.ConfigMap, mockNamespacedBuilder](
		testSchemeAttacher, namespacedGVK, testhelper.ResourceScopeNamespaced)

	testhelper.NewGenericNamespacedPullTestConfig(commonConfig, common.PullNamespacedBuilder).ExecuteTests(t)
}

func TestGet(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.Namespace, mockClusterScopedBuilder](
		testSchemeAttacher, clusterScopedGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewGenericGetTestConfig(commonConfig, common.Get).ExecuteTests(t)
}

func TestExists(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.Namespace, mockClusterScopedBuilder](
		testSchemeAttacher, clusterScopedGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewGenericExistsTestConfig(commonConfig, common.Exists).ExecuteTests(t)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.Namespace, mockClusterScopedBuilder](
		testSchemeAttacher, clusterScopedGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewGenericCreateTestConfig(commonConfig, common.Create).ExecuteTests(t)
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.Namespace, mockClusterScopedBuilder](
		testSchemeAttacher, clusterScopedGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewGenericUpdateTestConfig(commonConfig, common.Update, true).ExecuteTests(t)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.Namespace, mockClusterScopedBuilder](
		testSchemeAttacher, clusterScopedGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewGenericDeleteTestConfig(commonConfig, common.Delete).ExecuteTests(t)
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.Namespace, mockClusterScopedBuilder](
		testSchemeAttacher, clusterScopedGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewWithOptionsTestConfig(commonConfig).ExecuteTests(t)
}

func TestList(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.ConfigMap, mockNamespacedBuilder](
		testSchemeAttacher, namespacedGVK, testhelper.ResourceScopeNamespaced)

	testhelper.NewGenericListTestConfig(commonConfig, common.List[corev1.ConfigMap, corev1.ConfigMapList]).ExecuteTests(t)
}

func TestValidate(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.Namespace, mockClusterScopedBuilder](
		testSchemeAttacher, clusterScopedGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewGenericValidateTestConfig(commonConfig, common.Validate).ExecuteTests(t)
}

func TestConvertListOptionsToOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		options         []runtimeclient.ListOptions
		expectedOptions []runtimeclient.ListOption
	}{
		{
			name:            "valid conversion",
			options:         []runtimeclient.ListOptions{{}},
			expectedOptions: []runtimeclient.ListOption{&runtimeclient.ListOptions{}},
		},
		{
			name:            "nil options",
			options:         nil,
			expectedOptions: []runtimeclient.ListOption{},
		},
		{
			name:            "empty options",
			options:         []runtimeclient.ListOptions{},
			expectedOptions: []runtimeclient.ListOption{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			options := common.ConvertListOptionsToOptions(testCase.options)
			assert.Equal(t, testCase.expectedOptions, options)

			for i, option := range options {
				_, ok := option.(*runtimeclient.ListOptions)
				assert.Truef(t, ok, "option %d is not a runtimeclient.ListOptions", i)
			}
		})
	}
}

func TestConvertMetaListOptionsToListOptions(t *testing.T) {
	t.Parallel()

	t.Run("single option with all fields populated", func(t *testing.T) {
		t.Parallel()

		result, err := common.ConvertMetaListOptionsToListOptions([]metav1.ListOptions{{
			LabelSelector: "app=foo",
			FieldSelector: "metadata.name=bar",
			Limit:         10,
			Continue:      "token",
		}})

		require.NoError(t, err)
		require.Len(t, result, 1)

		runtimeOption, ok := result[0].(*runtimeclient.ListOptions)
		require.True(t, ok, "option should be *runtimeclient.ListOptions")

		assert.Nil(t, runtimeOption.Raw)
		assert.NotNil(t, runtimeOption.LabelSelector)
		assert.NotNil(t, runtimeOption.FieldSelector)
		assert.Equal(t, int64(10), runtimeOption.Limit)
		assert.Equal(t, "token", runtimeOption.Continue)
	})

	testCases := []struct {
		name          string
		options       []metav1.ListOptions
		expectedLen   int
		expectedError bool
	}{
		{
			name:        "empty input returns empty slice",
			options:     nil,
			expectedLen: 0,
		},
		{
			name:          "invalid label selector returns error",
			options:       []metav1.ListOptions{{LabelSelector: "app in (foo"}},
			expectedError: true,
		},
		{
			name:          "invalid field selector returns error",
			options:       []metav1.ListOptions{{FieldSelector: "a b"}},
			expectedError: true,
		},
		{
			name: "multiple options converted independently",
			options: []metav1.ListOptions{
				{LabelSelector: "app=foo", Limit: 5},
				{LabelSelector: "app=bar", Limit: 10},
			},
			expectedLen: 2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := common.ConvertMetaListOptionsToListOptions(testCase.options)

			if testCase.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)

				return
			}

			require.NoError(t, err)
			assert.Len(t, result, testCase.expectedLen)

			for _, opt := range result {
				rtOpt, ok := opt.(*runtimeclient.ListOptions)
				require.True(t, ok, "option should be *runtimeclient.ListOptions")

				assert.Nil(t, rtOpt.Raw)
			}
		})
	}
}

// mockClusterScopedBuilder implements the Builder interface for testing using a cluster-scoped resource.
type mockClusterScopedBuilder struct {
	common.EmbeddableBuilder[corev1.Namespace, *corev1.Namespace]
}

// Compile-time check to ensure MockClusterScopedBuilder implements Builder interface.
var _ common.Builder[corev1.Namespace, *corev1.Namespace] = (*mockClusterScopedBuilder)(nil)

// GetGVK returns the GVK for the mock cluster-scoped builder.
func (builder *mockClusterScopedBuilder) GetGVK() schema.GroupVersionKind {
	return clusterScopedGVK
}

// WithOptions delegates to the common package implementation for tests that exercise the generic helper.
func (builder *mockClusterScopedBuilder) WithOptions(
	options ...func(*mockClusterScopedBuilder) (*mockClusterScopedBuilder, error),
) *mockClusterScopedBuilder {
	return common.WithOptions(builder, options...)
}

// mockNamespacedBuilder implements the Builder interface for testing using a namespaced resource.
type mockNamespacedBuilder struct {
	common.EmbeddableBuilder[corev1.ConfigMap, *corev1.ConfigMap]
}

// Compile-time check to ensure MockNamespacedBuilder implements Builder interface.
var _ common.Builder[corev1.ConfigMap, *corev1.ConfigMap] = (*mockNamespacedBuilder)(nil)

// GetGVK returns the GVK for the mock namespaced builder.
func (builder *mockNamespacedBuilder) GetGVK() schema.GroupVersionKind {
	return namespacedGVK
}
