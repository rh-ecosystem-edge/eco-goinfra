package lso

import (
	"testing"
	"time"

	lsov1alpha1 "github.com/openshift/local-storage-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultLocalVolumeDiscoveryName      = "auto-discover-devices"
	defaultLocalVolumeDiscoveryNamespace = "test-lvdspace"
)

func TestNewLocalVolumeDiscoveryBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[
		lsov1alpha1.LocalVolumeDiscovery, LocalVolumeDiscoveryBuilder,
	](NewLocalVolumeDiscoveryBuilder, lsov1alpha1.AddToScheme, localVolumeDiscoveryGVK).ExecuteTests(t)
}

func TestPullLocalVolumeDiscovery(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[
		lsov1alpha1.LocalVolumeDiscovery, LocalVolumeDiscoveryBuilder,
	](PullLocalVolumeDiscovery, lsov1alpha1.AddToScheme, localVolumeDiscoveryGVK).ExecuteTests(t)
}

func TestLocalVolumeDiscoveryBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[
		lsov1alpha1.LocalVolumeDiscovery, LocalVolumeDiscoveryBuilder,
	](lsov1alpha1.AddToScheme, localVolumeDiscoveryGVK, testhelper.ResourceScopeNamespaced)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestLocalVolumeDiscoveryWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		nodeSelector   corev1.NodeSelector
		invalidBuilder bool
	}{
		{
			name: "node selector with matching expression",
			nodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{""},
					}},
				}},
			},
		},
		{
			name: "node selector with exists expression",
			nodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: corev1.NodeSelectorOpExists,
					}},
				}},
			},
		},
		{
			name: "node selector with multiple expressions",
			nodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "cluster.ocs.openshift.io/openshift-storage",
							Operator: corev1.NodeSelectorOpExists,
						},
						{
							Key:      "machineconfiguration.openshift.io/role",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"customcnf", "worker"},
						},
					},
				}},
			},
		},
		{
			name:           "invalid builder short circuits",
			nodeSelector:   corev1.NodeSelector{},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *LocalVolumeDiscoveryBuilder
			if testCase.invalidBuilder {
				builder = buildInvalidLocalVolumeDiscoveryTestBuilder(getTestLocalVolumeDiscoveryAPIClient())
			} else {
				builder = buildValidLocalVolumeDiscoveryTestBuilder(getTestLocalVolumeDiscoveryAPIClient())
			}

			initialError := builder.GetError()
			result := builder.WithNodeSelector(testCase.nodeSelector)

			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.NodeSelector)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.NodeSelector)
			assert.Equal(t, testCase.nodeSelector, *result.Definition.Spec.NodeSelector)
		})
	}
}

func TestLocalVolumeDiscoveryWithTolerations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		tolerations    []corev1.Toleration
		invalidBuilder bool
		expectedError  string
	}{
		{
			name: "valid tolerations",
			tolerations: []corev1.Toleration{{
				Key:      "node.ocs.openshift.io/storage",
				Operator: corev1.TolerationOpEqual,
				Value:    "true",
				Effect:   corev1.TaintEffectNoSchedule,
			}},
		},
		{
			name:          "empty tolerations",
			tolerations:   []corev1.Toleration{},
			expectedError: errEmptyTolerations,
		},
		{
			name:           "invalid builder short circuits",
			tolerations:    []corev1.Toleration{},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *LocalVolumeDiscoveryBuilder
			if testCase.invalidBuilder {
				builder = buildInvalidLocalVolumeDiscoveryTestBuilder(getTestLocalVolumeDiscoveryAPIClient())
			} else {
				builder = buildValidLocalVolumeDiscoveryTestBuilder(getTestLocalVolumeDiscoveryAPIClient())
			}

			initialError := builder.GetError()
			result := builder.WithTolerations(testCase.tolerations)

			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Tolerations)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Spec.Tolerations)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.tolerations, result.Definition.Spec.Tolerations)
		})
	}
}

func TestLocalVolumeDiscoveryGetPhase(t *testing.T) {
	t.Parallel()

	t.Run("returns discovering phase", func(t *testing.T) {
		t.Parallel()

		builder := buildValidLocalVolumeDiscoveryTestBuilder(
			newLocalVolumeDiscoveryTestClient(lsov1alpha1.Discovering))

		phase, err := builder.GetPhase()

		require.NoError(t, err)
		assert.Equal(t, lsov1alpha1.Discovering, phase)
	})

	t.Run("resource not found", func(t *testing.T) {
		t.Parallel()

		builder := buildValidLocalVolumeDiscoveryTestBuilder(getTestLocalVolumeDiscoveryAPIClient())

		phase, err := builder.GetPhase()

		assert.Equal(t, lsov1alpha1.DiscoveryPhase(""), phase)
		assert.EqualError(t, err, "auto-discover-devices localVolumeDiscovery not found in test-lvdspace namespace")
	})

	t.Run("invalid builder short circuits", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidLocalVolumeDiscoveryTestBuilder(
			newLocalVolumeDiscoveryTestClient(lsov1alpha1.Discovering))

		phase, err := builder.GetPhase()

		assert.Equal(t, lsov1alpha1.DiscoveryPhase(""), phase)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
	})
}

func TestLocalVolumeDiscoveryIsDiscovering(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *LocalVolumeDiscoveryBuilder
		expectedResult bool
	}{
		{
			name: "returns true while discovering",
			builder: buildValidLocalVolumeDiscoveryTestBuilder(
				newLocalVolumeDiscoveryTestClient(lsov1alpha1.Discovering)),
			expectedResult: true,
		},
		{
			name: "returns false when discovery failed",
			builder: buildValidLocalVolumeDiscoveryTestBuilder(
				newLocalVolumeDiscoveryTestClient(lsov1alpha1.DiscoveryFailed)),
			expectedResult: false,
		},
		{
			name: "invalid builder short circuits",
			builder: buildInvalidLocalVolumeDiscoveryTestBuilder(
				newLocalVolumeDiscoveryTestClient(lsov1alpha1.Discovering)),
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expectedResult, testCase.builder.IsDiscovering(time.Second))
		})
	}
}

func buildValidLocalVolumeDiscoveryTestBuilder(apiClient *clients.Settings) *LocalVolumeDiscoveryBuilder {
	return NewLocalVolumeDiscoveryBuilder(
		apiClient, defaultLocalVolumeDiscoveryName, defaultLocalVolumeDiscoveryNamespace)
}

func buildInvalidLocalVolumeDiscoveryTestBuilder(apiClient *clients.Settings) *LocalVolumeDiscoveryBuilder {
	return NewLocalVolumeDiscoveryBuilder(apiClient, "", defaultLocalVolumeDiscoveryNamespace)
}

func getTestLocalVolumeDiscoveryAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{lsov1alpha1.AddToScheme},
	})
}

func newLocalVolumeDiscoveryTestClient(phase ...lsov1alpha1.DiscoveryPhase) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyLocalVolumeDiscovery(phase...),
		SchemeAttachers: []clients.SchemeAttacher{lsov1alpha1.AddToScheme},
	})
}

func buildDummyLocalVolumeDiscovery(phase ...lsov1alpha1.DiscoveryPhase) []runtime.Object {
	discoveryPhase := lsov1alpha1.Discovering

	if len(phase) > 0 {
		discoveryPhase = phase[0]
	}

	return []runtime.Object{&lsov1alpha1.LocalVolumeDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultLocalVolumeDiscoveryName,
			Namespace: defaultLocalVolumeDiscoveryNamespace,
		},
		Status: lsov1alpha1.LocalVolumeDiscoveryStatus{
			Phase: discoveryPhase,
		},
	}}
}
