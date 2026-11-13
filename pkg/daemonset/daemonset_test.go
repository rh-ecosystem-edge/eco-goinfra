package daemonset

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultDaemonSetLabels    = map[string]string{"test-key": "test-value"}
	defaultDaemonSetContainer = corev1.Container{Name: "test-container"}
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[appsv1.DaemonSet, Builder](
		func(apiClient *clients.Settings, name, nsname string) *Builder {
			return NewBuilder(apiClient, name, nsname, defaultDaemonSetLabels, defaultDaemonSetContainer)
		},
		appsv1.AddToScheme,
		daemonSetGVK,
	).ExecuteTests(t)
}

func TestNewBuilderSetsSpec(t *testing.T) {
	t.Parallel()

	builder := buildValidDaemonSetTestBuilder()
	require.NoError(t, builder.GetError())
	require.NotNil(t, builder.Definition.Spec.Selector)
	assert.Equal(t, defaultDaemonSetLabels, builder.Definition.Spec.Selector.MatchLabels)
	assert.Equal(t, defaultDaemonSetLabels, builder.Definition.Spec.Template.Labels)
	assert.Equal(t, []corev1.Container{defaultDaemonSetContainer}, builder.Definition.Spec.Template.Spec.Containers)
}

func TestNewBuilderRejectsEmptyLabels(t *testing.T) {
	t.Parallel()

	builder := NewBuilder(
		newDaemonSetTestClient(), "test-name", "test-namespace", map[string]string{}, defaultDaemonSetContainer)

	require.Error(t, builder.GetError())
	assert.Equal(t, "daemonset 'labels' cannot be empty", builder.GetError().Error())
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[appsv1.DaemonSet, Builder](
		Pull, appsv1.AddToScheme, daemonSetGVK,
	).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newDaemonSetCommonTestConfig()

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

	testhelper.NewWithOptionsTestConfig(newDaemonSetCommonTestConfig()).ExecuteTests(t)
}

func TestWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *Builder
		selector    map[string]string
		assertError func(error) bool
	}{
		{
			name:        "sets node selector",
			builder:     buildValidDaemonSetTestBuilder,
			selector:    map[string]string{"test-node-selector-key": "test-node-selector-value"},
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:     "empty selector sets builder error",
			builder:  buildValidDaemonSetTestBuilder,
			selector: map[string]string{},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "cannot accept empty map as nodeselector"
			},
		},
		{
			name:        "invalid builder short circuits",
			builder:     buildInvalidDaemonSetTestBuilder,
			selector:    map[string]string{"test-node-selector-key": "test-node-selector-value"},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithNodeSelector(testCase.selector)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.selector, result.Definition.Spec.Template.Spec.NodeSelector)
			}
		})
	}
}

func TestWithHostNetwork(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *Builder
		assertError func(error) bool
	}{
		{
			name:        "enables host network",
			builder:     buildValidDaemonSetTestBuilder,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "invalid builder short circuits",
			builder:     buildInvalidDaemonSetTestBuilder,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithHostNetwork()
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.True(t, result.Definition.Spec.Template.Spec.HostNetwork)
			}
		})
	}
}

func TestWithPodAffinity(t *testing.T) {
	t.Parallel()

	affinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "kubernetes.io/hostname",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"host1", "host2"},
					}},
				}},
			},
		},
	}

	testCases := []struct {
		name        string
		builder     func() *Builder
		affinity    *corev1.Affinity
		assertError func(error) bool
	}{
		{
			name:        "sets pod affinity",
			builder:     buildValidDaemonSetTestBuilder,
			affinity:    affinity,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:     "nil affinity sets builder error",
			builder:  buildValidDaemonSetTestBuilder,
			affinity: nil,
			assertError: func(err error) bool {
				return err != nil && err.Error() == "affinity parameter is empty"
			},
		},
		{
			name:        "invalid builder short circuits",
			builder:     buildInvalidDaemonSetTestBuilder,
			affinity:    affinity,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithPodAffinity(testCase.affinity)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.affinity, result.Definition.Spec.Template.Spec.Affinity)
			}
		})
	}
}

func TestWithVolume(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *Builder
		volume      corev1.Volume
		assertError func(error) bool
	}{
		{
			name:        "adds volume",
			builder:     buildValidDaemonSetTestBuilder,
			volume:      corev1.Volume{Name: "test-volume"},
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:    "empty volume name sets builder error",
			builder: buildValidDaemonSetTestBuilder,
			volume:  corev1.Volume{},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "volume name parameter is empty"
			},
		},
		{
			name:        "invalid builder short circuits",
			builder:     buildInvalidDaemonSetTestBuilder,
			volume:      corev1.Volume{Name: "test-volume"},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithVolume(testCase.volume)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Contains(t, result.Definition.Spec.Template.Spec.Volumes, testCase.volume)
			}
		})
	}
}

func TestWithAdditionalContainerSpecs(t *testing.T) {
	t.Parallel()

	additionalContainer := corev1.Container{Name: "test-additional-container"}

	testCases := []struct {
		name        string
		builder     func() *Builder
		specs       []corev1.Container
		assertError func(error) bool
	}{
		{
			name:        "appends container specs",
			builder:     buildValidDaemonSetTestBuilder,
			specs:       []corev1.Container{additionalContainer},
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:    "empty specs set builder error",
			builder: buildValidDaemonSetTestBuilder,
			specs:   []corev1.Container{},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "cannot accept empty list as container specs"
			},
		},
		{
			name:        "invalid builder short circuits",
			builder:     buildInvalidDaemonSetTestBuilder,
			specs:       []corev1.Container{additionalContainer},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithAdditionalContainerSpecs(testCase.specs)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, append([]corev1.Container{defaultDaemonSetContainer}, testCase.specs...),
					result.Definition.Spec.Template.Spec.Containers)
			}
		})
	}
}

func TestGetGVR(t *testing.T) {
	t.Parallel()

	testGVR := GetGVR()
	assert.Equal(t, "apps", testGVR.Group)
	assert.Equal(t, "v1", testGVR.Version)
	assert.Equal(t, "daemonsets", testGVR.Resource)
}

func newDaemonSetCommonTestConfig() testhelper.CommonTestConfig[appsv1.DaemonSet, Builder, *appsv1.DaemonSet, *Builder] {
	return testhelper.NewCommonTestConfig[appsv1.DaemonSet, Builder](
		appsv1.AddToScheme, daemonSetGVK, testhelper.ResourceScopeNamespaced)
}

func newDaemonSetTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{appsv1.AddToScheme},
	})
}

func buildValidDaemonSetTestBuilder() *Builder {
	return NewBuilder(
		newDaemonSetTestClient(), "test-name", "test-namespace", defaultDaemonSetLabels, defaultDaemonSetContainer)
}

func buildInvalidDaemonSetTestBuilder() *Builder {
	return NewBuilder(
		newDaemonSetTestClient(), "", "test-namespace", defaultDaemonSetLabels, defaultDaemonSetContainer)
}
