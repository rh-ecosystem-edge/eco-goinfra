package idms

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const TestIDMS = "test-image-digest-mirror-set"

func TestNewIDMSBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common cluster-scoped builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewClusterScopedBuilderTestConfig(
			func(apiClient *clients.Settings, name string) *Builder {
				return NewBuilder(apiClient, name, testImageDigestMirror())
			},
			configv1.AddToScheme,
			idmsGVK,
		).ExecuteTests(t)
	})

	t.Run("mirror is added to definition", func(t *testing.T) {
		t.Parallel()

		mirror := testImageDigestMirror()
		testBuilder := NewBuilder(getTestIDMSAPIClient(), TestIDMS, mirror)

		require.NoError(t, testBuilder.GetError())
		require.Len(t, testBuilder.Definition.Spec.ImageDigestMirrors, 1)
		assert.Equal(t, mirror, testBuilder.Definition.Spec.ImageDigestMirrors[0])
	})
}

func TestIDMSPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(Pull, configv1.AddToScheme, idmsGVK).ExecuteTests(t)
}

func TestIDMSBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[configv1.ImageDigestMirrorSet, Builder](
		configv1.AddToScheme,
		idmsGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestIDMSWithMirror(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		invalid       bool
		expectedError bool
	}{
		{
			name:    "adds mirror",
			invalid: false,
		},
		{
			name:          "invalid builder short circuits",
			invalid:       true,
			expectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testBuilder *Builder
			if testCase.invalid {
				testBuilder = buildInvalidIDMSTestBuilder(getTestIDMSAPIClient())
			} else {
				testBuilder = buildValidIDMSTestBuilder(getTestIDMSAPIClient())
			}

			initialError := testBuilder.GetError()
			mirror := configv1.ImageDigestMirrors{
				Mirrors: []configv1.ImageMirror{"cloned.test.org"},
				Source:  "test.org",
			}

			result := testBuilder.WithMirror(mirror)
			require.Same(t, testBuilder, result)

			if testCase.expectedError {
				require.Error(t, result.GetError())
				assert.Equal(t, initialError, result.GetError())
				assert.Empty(t, result.Definition.Spec.ImageDigestMirrors)

				return
			}

			require.NoError(t, result.GetError())
			require.Len(t, result.Definition.Spec.ImageDigestMirrors, 2)
			assert.Equal(t, mirror, result.Definition.Spec.ImageDigestMirrors[1])
		})
	}
}

func buildValidIDMSTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, TestIDMS, testImageDigestMirror())
}

func buildInvalidIDMSTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "", testImageDigestMirror())
}

func getTestIDMSAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{configv1.AddToScheme},
	})
}

func testImageDigestMirror() configv1.ImageDigestMirrors {
	return configv1.ImageDigestMirrors{
		Mirrors: []configv1.ImageMirror{"cloned.registry.org"},
		Source:  "registry.org",
	}
}

func generateImagDigestMirrorSet() *configv1.ImageDigestMirrorSet {
	return &configv1.ImageDigestMirrorSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: TestIDMS,
		},
		Spec: configv1.ImageDigestMirrorSetSpec{
			ImageDigestMirrors: []configv1.ImageDigestMirrors{testImageDigestMirror()},
		},
	}
}
