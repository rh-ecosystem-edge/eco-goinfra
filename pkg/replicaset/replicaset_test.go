package replicaset

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

var (
	defaultReplicaSetName      = "test-name"
	defaultReplicaSetNamespace = "test-namespace"
	defaultReplicaSetLabel     = map[string]string{"testLabels": "testLabelValue"}
	defaultReplicaSetContainer = []corev1.Container{{Name: "test-container"}}
)

func TestNewReplicaSetBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		func(apiClient *clients.Settings, name, nsname string) *Builder {
			return NewBuilder(apiClient, name, nsname, defaultReplicaSetLabel, defaultReplicaSetContainer)
		},
		appsv1.AddToScheme,
		replicaSetGVK,
	).ExecuteTests(t)

	testCases := []struct {
		name             string
		labels           map[string]string
		containers       []corev1.Container
		expectedErrorMsg string
	}{
		{
			name:             "valid labels and containers",
			labels:           defaultReplicaSetLabel,
			containers:       defaultReplicaSetContainer,
			expectedErrorMsg: "",
		},
		{
			name:             "empty labels",
			labels:           map[string]string{},
			containers:       defaultReplicaSetContainer,
			expectedErrorMsg: "replicaset 'labels' cannot be empty",
		},
		{
			name:             "empty containers",
			labels:           defaultReplicaSetLabel,
			containers:       []corev1.Container{},
			expectedErrorMsg: "replicaset 'containerSpec' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewBuilder(
				newReplicaSetTestClient(),
				defaultReplicaSetName,
				defaultReplicaSetNamespace,
				testCase.labels,
				testCase.containers,
			)
			require.NotNil(t, builder)

			if testCase.expectedErrorMsg != "" {
				require.EqualError(t, builder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.labels, builder.Definition.Labels)
			assert.Equal(t, testCase.containers, builder.Definition.Spec.Template.Spec.Containers)
		})
	}
}

func TestPullReplicaSet(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull,
		appsv1.AddToScheme,
		replicaSetGVK,
	).ExecuteTests(t)
}

func TestReplicaSetBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newReplicaSetCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestReplicaSetWithLabel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		labels          map[string]string
		invalidBuilder  bool
		expectedErrText string
	}{
		{
			name:   "valid labels",
			labels: map[string]string{"test-label-key": "test-label-value"},
		},
		{
			name:   "empty label value is allowed",
			labels: map[string]string{"test-label-key": ""},
		},
		{
			name:            "empty label key",
			labels:          map[string]string{"": "test-label-value"},
			expectedErrText: canNotApplyLabelsWithAn,
		},
		{
			name:            "empty labels",
			labels:          map[string]string{},
			expectedErrText: canNotApplyEmptyLabels,
		},
		{
			name:           "invalid builder short circuits",
			labels:         map[string]string{"test-label-key": "test-label-value"},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidReplicaSetTestBuilder()
			if testCase.invalidBuilder {
				builder = buildInvalidReplicaSetTestBuilder()
			}

			result := builder.WithLabel(testCase.labels)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())

				return
			}

			if testCase.expectedErrText != "" {
				require.EqualError(t, result.GetError(), testCase.expectedErrText)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.labels, result.Definition.Labels)
		})
	}
}

func TestReplicaSetWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		nodeSelector    map[string]string
		invalidBuilder  bool
		expectedErrText string
	}{
		{
			name:         "valid node selector",
			nodeSelector: map[string]string{"test-node-selector-key": "test-node-selector-value"},
		},
		{
			name:         "empty node selector value is allowed",
			nodeSelector: map[string]string{"test-node-selector-key": ""},
		},
		{
			name:            "empty node selector key",
			nodeSelector:    map[string]string{"": "test-node-selector-value"},
			expectedErrText: "can not apply a nodeSelector with an empty key value",
		},
		{
			name:            "empty node selector",
			nodeSelector:    map[string]string{},
			expectedErrText: "can not apply empty nodeSelector",
		},
		{
			name:           "invalid builder short circuits",
			nodeSelector:   map[string]string{"test-node-selector-key": "test-node-selector-value"},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidReplicaSetTestBuilder()
			if testCase.invalidBuilder {
				builder = buildInvalidReplicaSetTestBuilder()
			}

			result := builder.WithNodeSelector(testCase.nodeSelector)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())

				return
			}

			if testCase.expectedErrText != "" {
				require.EqualError(t, result.GetError(), testCase.expectedErrText)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.nodeSelector, result.Definition.Spec.Template.Spec.NodeSelector)
		})
	}
}

func TestReplicaSetWithVolume(t *testing.T) {
	t.Parallel()

	testVolume := corev1.Volume{
		Name: "test-volume",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	testCases := []struct {
		name            string
		volume          corev1.Volume
		invalidBuilder  bool
		expectedErrText string
	}{
		{
			name:   "valid volume",
			volume: testVolume,
		},
		{
			name:            "empty volume name",
			volume:          corev1.Volume{},
			expectedErrText: volumeNameParameterIsEmpty,
		},
		{
			name:           "invalid builder short circuits",
			volume:         testVolume,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidReplicaSetTestBuilder()
			if testCase.invalidBuilder {
				builder = buildInvalidReplicaSetTestBuilder()
			}

			result := builder.WithVolume(testCase.volume)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())
				assert.Empty(t, result.Definition.Spec.Template.Spec.Volumes)

				return
			}

			if testCase.expectedErrText != "" {
				require.EqualError(t, result.GetError(), testCase.expectedErrText)
				assert.Empty(t, result.Definition.Spec.Template.Spec.Volumes)

				return
			}

			require.NoError(t, result.GetError())
			require.Len(t, result.Definition.Spec.Template.Spec.Volumes, 1)
			assert.Equal(t, testCase.volume, result.Definition.Spec.Template.Spec.Volumes[0])
		})
	}
}

func TestReplicaSetWithAdditionalContainerSpecs(t *testing.T) {
	t.Parallel()

	additionalContainer := corev1.Container{Name: "test-additional-container"}

	testCases := []struct {
		name            string
		specs           []corev1.Container
		invalidBuilder  bool
		expectedErrText string
	}{
		{
			name:  "valid container specs",
			specs: []corev1.Container{additionalContainer},
		},
		{
			name:            "empty container specs",
			specs:           []corev1.Container{},
			expectedErrText: "cannot accept empty list as container specs",
		},
		{
			name:           "invalid builder short circuits",
			specs:          []corev1.Container{additionalContainer},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidReplicaSetTestBuilder()
			if testCase.invalidBuilder {
				builder = buildInvalidReplicaSetTestBuilder()
			}

			result := builder.WithAdditionalContainerSpecs(testCase.specs)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())
				assert.Equal(t, defaultReplicaSetContainer, result.Definition.Spec.Template.Spec.Containers)

				return
			}

			if testCase.expectedErrText != "" {
				require.EqualError(t, result.GetError(), testCase.expectedErrText)
				assert.Equal(t, defaultReplicaSetContainer, result.Definition.Spec.Template.Spec.Containers)

				return
			}

			require.NoError(t, result.GetError())
			require.Len(t, result.Definition.Spec.Template.Spec.Containers, 2)
			assert.Equal(t, additionalContainer, result.Definition.Spec.Template.Spec.Containers[1])
		})
	}
}

func TestReplicaSetWaitMethods(t *testing.T) {
	t.Parallel()

	t.Run("CreateAndWaitUntilReady", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder(
			newReplicaSetTestClient(),
			defaultReplicaSetName,
			defaultReplicaSetNamespace,
			defaultReplicaSetLabel,
			defaultReplicaSetContainer,
		)

		result, err := builder.CreateAndWaitUntilReady(time.Second)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Object)
		assert.Equal(t, defaultReplicaSetName, result.Object.Name)
	})

	t.Run("DeleteAndWait", func(t *testing.T) {
		t.Parallel()

		builder := buildValidReplicaSetTestBuilder(buildDummyReplicaSet())
		require.NoError(t, builder.DeleteAndWait(time.Second))
		assert.Nil(t, builder.Object)
	})

	t.Run("IsReady", func(t *testing.T) {
		t.Parallel()

		builder := buildValidReplicaSetTestBuilder(buildDummyReplicaSet())
		assert.True(t, builder.IsReady(time.Second))
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidReplicaSetTestBuilder()

		result, err := builder.CreateAndWaitUntilReady(time.Second)
		require.Same(t, builder, result)
		require.Error(t, err)
		require.Error(t, builder.DeleteAndWait(time.Second))
		assert.False(t, builder.IsReady(time.Second))
	})
}

func TestGetGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "apps", GetGVR().Group)
	assert.Equal(t, "v1", GetGVR().Version)
	assert.Equal(t, "replicasets", GetGVR().Resource)
}

func newReplicaSetCommonTestConfig() testhelper.CommonTestConfig[appsv1.ReplicaSet, Builder, *appsv1.ReplicaSet, *Builder] {
	return testhelper.NewCommonTestConfig[appsv1.ReplicaSet, Builder](
		appsv1.AddToScheme,
		replicaSetGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

func newReplicaSetTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{appsv1.AddToScheme},
	})
}

func buildValidReplicaSetTestBuilder(objects ...runtime.Object) *Builder {
	return NewBuilder(
		newReplicaSetTestClient(objects...),
		defaultReplicaSetName,
		defaultReplicaSetNamespace,
		defaultReplicaSetLabel,
		defaultReplicaSetContainer,
	)
}

func buildInvalidReplicaSetTestBuilder() *Builder {
	return NewBuilder(
		newReplicaSetTestClient(),
		defaultReplicaSetName,
		defaultReplicaSetNamespace,
		map[string]string{},
		defaultReplicaSetContainer,
	)
}

func buildDummyReplicaSet() *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultReplicaSetName,
			Namespace: defaultReplicaSetNamespace,
			Labels:    defaultReplicaSetLabel,
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: defaultReplicaSetContainer,
				},
			},
		},
	}
}
