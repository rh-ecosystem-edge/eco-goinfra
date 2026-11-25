package statefulset

import (
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var statefulSetGVK = appsv1.SchemeGroupVersion.WithKind("StatefulSet")

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig[appsv1.StatefulSet, Builder](
			func(apiClient *clients.Settings, name, nsname string) *Builder {
				return NewBuilder(apiClient, name, nsname, validTestLabels(), validTestContainer())
			},
			appsv1.AddToScheme,
			statefulSetGVK,
		).ExecuteTests(t)
	})

	t.Run("valid required parameters are applied", func(t *testing.T) {
		t.Parallel()

		labels := validTestLabels()
		container := validTestContainer()
		builder := NewBuilder(newStatefulSetTestClient(), "test-name", "test-namespace", labels, container)

		require.NoError(t, builder.GetError())
		assert.Equal(t, labels, builder.Definition.Spec.Selector.MatchLabels)
		assert.Equal(t, labels, builder.Definition.Spec.Template.Labels)
		require.Len(t, builder.Definition.Spec.Template.Spec.Containers, 1)
		assert.Equal(t, *container, builder.Definition.Spec.Template.Spec.Containers[0])
	})

	t.Run("nil labels set builder error", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder(newStatefulSetTestClient(), "test-name", "test-namespace", nil, validTestContainer())

		require.EqualError(t, builder.GetError(), "statefulset 'labels' cannot be empty")
	})

	t.Run("nil container spec sets builder error", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder(newStatefulSetTestClient(), "test-name", "test-namespace", validTestLabels(), nil)

		require.EqualError(t, builder.GetError(), "statefulset 'containerSpec' cannot be nil")
	})
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[appsv1.StatefulSet, Builder](
		Pull,
		appsv1.AddToScheme,
		statefulSetGVK,
	).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newStatefulSetCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		Run(t)
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newStatefulSetCommonTestConfig()).ExecuteTests(t)
}

func TestWithAdditionalContainerSpecs(t *testing.T) {
	t.Parallel()

	t.Run("appends valid container specs", func(t *testing.T) {
		t.Parallel()

		builder := buildValidStatefulSetTestBuilder(newStatefulSetTestClient())
		result := builder.WithAdditionalContainerSpecs([]corev1.Container{
			{Name: "additional-one", Image: "additional-one-image"},
			{Name: "additional-two", Image: "additional-two-image"},
		})

		require.Same(t, builder, result)
		require.NoError(t, result.GetError())
		require.Len(t, result.Definition.Spec.Template.Spec.Containers, 3)
		assert.Equal(t, "test-container", result.Definition.Spec.Template.Spec.Containers[0].Name)
		assert.Equal(t, "additional-one", result.Definition.Spec.Template.Spec.Containers[1].Name)
		assert.Equal(t, "additional-two", result.Definition.Spec.Template.Spec.Containers[2].Name)
	})

	t.Run("nil specs set builder error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidStatefulSetTestBuilder(newStatefulSetTestClient())
		result := builder.WithAdditionalContainerSpecs(nil)

		require.Same(t, builder, result)
		require.EqualError(t, result.GetError(), "cannot accept nil or empty list as container specs")
		assert.Len(t, result.Definition.Spec.Template.Spec.Containers, 1)
	})

	t.Run("invalid builder short circuits", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidStatefulSetTestBuilder(newStatefulSetTestClient())
		result := builder.WithAdditionalContainerSpecs([]corev1.Container{
			{Name: "additional", Image: "additional-image"},
		})

		require.Same(t, builder, result)
		require.EqualError(t, result.GetError(), "statefulset 'labels' cannot be empty")
		assert.Empty(t, result.Definition.Spec.Template.Spec.Containers)
	})
}

func TestWithPodAnnotations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		annotations         map[string]string
		existingAnnotations map[string]string
		expectedAnnotations map[string]string
		expectedError       string
	}{
		{
			name: "adds annotations to empty pod template",
			annotations: map[string]string{
				"app.kubernetes.io/version": "v1.0.0",
				"custom.annotation/key":     "value",
			},
			expectedAnnotations: map[string]string{
				"app.kubernetes.io/version": "v1.0.0",
				"custom.annotation/key":     "value",
			},
		},
		{
			name: "merges annotations with existing ones",
			annotations: map[string]string{
				"new.annotation/key": "new-value",
			},
			existingAnnotations: map[string]string{
				"existing.annotation/key": "existing-value",
			},
			expectedAnnotations: map[string]string{
				"existing.annotation/key": "existing-value",
				"new.annotation/key":      "new-value",
			},
		},
		{
			name: "overwrites existing annotation",
			annotations: map[string]string{
				"shared.annotation/key": "new-value",
			},
			existingAnnotations: map[string]string{
				"shared.annotation/key": "old-value",
				"other.annotation/key":  "other-value",
			},
			expectedAnnotations: map[string]string{
				"shared.annotation/key": "new-value",
				"other.annotation/key":  "other-value",
			},
		},
		{
			name:          "nil annotations set builder error",
			annotations:   nil,
			expectedError: "cannot accept nil or empty annotations",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidStatefulSetTestBuilder(newStatefulSetTestClient())
			builder.Definition.Spec.Template.Annotations = testCase.existingAnnotations

			result := builder.WithPodAnnotations(testCase.annotations)

			require.Same(t, builder, result)

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.expectedAnnotations, result.Definition.Spec.Template.Annotations)
		})
	}

	t.Run("invalid builder short circuits", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidStatefulSetTestBuilder(newStatefulSetTestClient())
		result := builder.WithPodAnnotations(map[string]string{"test": "value"})

		require.Same(t, builder, result)
		require.EqualError(t, result.GetError(), "statefulset 'labels' cannot be empty")
		assert.Empty(t, result.Definition.Spec.Template.Annotations)
	})
}

func TestIsReady(t *testing.T) {
	t.Parallel()

	t.Run("returns true for ready statefulset", func(t *testing.T) {
		t.Parallel()

		statefulSet := testStatefulSetObject("test-name", "test-namespace")
		statefulSet.Status.Replicas = 2
		statefulSet.Status.ReadyReplicas = 2
		builder := buildValidStatefulSetTestBuilder(newStatefulSetTestClient(statefulSet))

		assert.True(t, builder.IsReady(time.Second))
	})

	t.Run("returns false when statefulset does not exist", func(t *testing.T) {
		t.Parallel()

		builder := buildValidStatefulSetTestBuilder(newStatefulSetTestClient())

		assert.False(t, builder.IsReady(time.Second))
	})

	t.Run("returns false for invalid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidStatefulSetTestBuilder(newStatefulSetTestClient())

		assert.False(t, builder.IsReady(time.Second))
	})
}

func TestGetGVR(t *testing.T) {
	t.Parallel()

	testGVR := GetGVR()
	assert.Equal(t, "statefulsets", testGVR.Resource)
	assert.Equal(t, "v1", testGVR.Version)
	assert.Equal(t, "apps", testGVR.Group)
}

func newStatefulSetCommonTestConfig() testhelper.CommonTestConfig[appsv1.StatefulSet, Builder, *appsv1.StatefulSet, *Builder] {
	return testhelper.NewCommonTestConfig[appsv1.StatefulSet, Builder](
		appsv1.AddToScheme, statefulSetGVK, testhelper.ResourceScopeNamespaced)
}

func newStatefulSetTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{appsv1.AddToScheme},
	})
}

func buildValidStatefulSetTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "test-name", "test-namespace", validTestLabels(), validTestContainer())
}

func buildInvalidStatefulSetTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "test-name", "test-namespace", nil, validTestContainer())
}

func validTestLabels() map[string]string {
	return map[string]string{"app": "test-statefulset"}
}

func validTestContainer() *corev1.Container {
	return &corev1.Container{Name: "test-container", Image: "test-image"}
}

func testStatefulSetObject(name, namespace string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}
}
