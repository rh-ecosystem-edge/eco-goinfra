package icsp

import (
	"testing"

	v1alpha1 "github.com/openshift/api/operator/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultICSPName   = "test-icsp"
	defaultICSPSource = "test-source"
)

var defaultICSPMirrors = []string{"test-mirror"}

func TestNewICSPBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common cluster-scoped builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewClusterScopedBuilderTestConfig(
			func(apiClient *clients.Settings, name string) *ICSPBuilder {
				return NewICSPBuilder(apiClient, name, defaultICSPSource, defaultICSPMirrors)
			},
			v1alpha1.Install,
			icspGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		source        string
		mirrors       []string
		expectedError string
	}{
		{
			name:    "valid source and mirrors",
			source:  defaultICSPSource,
			mirrors: defaultICSPMirrors,
		},
		{
			name:          "empty source",
			source:        "",
			mirrors:       defaultICSPMirrors,
			expectedError: errEmptySource,
		},
		{
			name:          "empty mirrors",
			source:        defaultICSPSource,
			mirrors:       nil,
			expectedError: errEmptyMirrors,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := NewICSPBuilder(
				getTestICSPAPIClient(), defaultICSPName, testCase.source, testCase.mirrors)

			if testCase.expectedError != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, []v1alpha1.RepositoryDigestMirrors{{
				Source:  testCase.source,
				Mirrors: testCase.mirrors,
			}}, testBuilder.Definition.Spec.RepositoryDigestMirrors)
		})
	}
}

func TestPullICSP(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(Pull, v1alpha1.Install, icspGVK).ExecuteTests(t)
}

func TestICSPBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newICSPCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestICSPWithRepositoryDigestMirror(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		source        string
		mirrors       []string
		expectedError string
	}{
		{
			name:    "adds repository digest mirror",
			valid:   true,
			source:  "additional-source",
			mirrors: []string{"additional-mirror"},
		},
		{
			name:    "invalid builder short circuits",
			valid:   false,
			source:  "additional-source",
			mirrors: []string{"additional-mirror"},
		},
		{
			name:          "empty source",
			valid:         true,
			source:        "",
			mirrors:       []string{"additional-mirror"},
			expectedError: errEmptySource,
		},
		{
			name:          "empty mirrors",
			valid:         true,
			source:        "additional-source",
			mirrors:       nil,
			expectedError: errEmptyMirrors,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testBuilder *ICSPBuilder
			if testCase.valid {
				testBuilder = buildValidICSPBuilder(getTestICSPAPIClient())
			} else {
				testBuilder = buildInvalidICSPBuilder(getTestICSPAPIClient())
			}

			initialError := testBuilder.GetError()
			result := testBuilder.WithRepositoryDigestMirror(testCase.source, testCase.mirrors)
			require.Same(t, testBuilder, result)

			if !testCase.valid {
				assert.Equal(t, initialError, result.GetError())
				assert.Empty(t, result.Definition.Spec.RepositoryDigestMirrors)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Len(t, result.Definition.Spec.RepositoryDigestMirrors, 1)

				return
			}

			require.NoError(t, result.GetError())
			require.Len(t, result.Definition.Spec.RepositoryDigestMirrors, 2)
			assert.Equal(t, v1alpha1.RepositoryDigestMirrors{
				Source:  testCase.source,
				Mirrors: testCase.mirrors,
			}, result.Definition.Spec.RepositoryDigestMirrors[1])
		})
	}
}

func TestICSPWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newICSPCommonTestConfig()).ExecuteTests(t)
}

func newICSPCommonTestConfig() testhelper.CommonTestConfig[
	v1alpha1.ImageContentSourcePolicy, ICSPBuilder, *v1alpha1.ImageContentSourcePolicy, *ICSPBuilder,
] {
	return testhelper.NewCommonTestConfig[v1alpha1.ImageContentSourcePolicy, ICSPBuilder](
		v1alpha1.Install,
		icspGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

func buildValidICSPBuilder(apiClient *clients.Settings) *ICSPBuilder {
	return NewICSPBuilder(apiClient, defaultICSPName, defaultICSPSource, defaultICSPMirrors)
}

func buildInvalidICSPBuilder(apiClient *clients.Settings) *ICSPBuilder {
	return NewICSPBuilder(apiClient, defaultICSPName, defaultICSPSource, nil)
}

func getTestICSPAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}
