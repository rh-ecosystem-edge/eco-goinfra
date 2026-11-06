package hive

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	hivev1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/hive/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultClusterImageSetName    = "imageset"
	defaultClusterImageSetRelease = "imageset-release-1-1"
)

func TestNewClusterImageSetBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig[hivev1.ClusterImageSet, ClusterImageSetBuilder](
		func(apiClient *clients.Settings, name string) *ClusterImageSetBuilder {
			return NewClusterImageSetBuilder(apiClient, name, defaultClusterImageSetRelease)
		},
		hivev1.AddToScheme,
		clusterImageSetGVK,
	).ExecuteTests(t)

	t.Run("sets release image", func(t *testing.T) {
		t.Parallel()

		builder := NewClusterImageSetBuilder(
			clients.GetTestClients(clients.TestClientParams{}), defaultClusterImageSetName, defaultClusterImageSetRelease)

		require.NoError(t, builder.GetError())
		assert.Equal(t, defaultClusterImageSetRelease, builder.Definition.Spec.ReleaseImage)
	})

	t.Run("empty release image returns error", func(t *testing.T) {
		t.Parallel()

		builder := NewClusterImageSetBuilder(
			clients.GetTestClients(clients.TestClientParams{}), defaultClusterImageSetName, "")

		assert.EqualError(t, builder.GetError(), "clusterimageset 'releaseImage' cannot be empty")
	})
}

func TestPullClusterImageSet(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig[hivev1.ClusterImageSet, ClusterImageSetBuilder](
		PullClusterImageSet, hivev1.AddToScheme, clusterImageSetGVK).ExecuteTests(t)
}

func TestClusterImageSetBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[hivev1.ClusterImageSet, ClusterImageSetBuilder](
		hivev1.AddToScheme,
		clusterImageSetGVK,
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

func TestClusterImageSetWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[hivev1.ClusterImageSet, ClusterImageSetBuilder](
			hivev1.AddToScheme,
			clusterImageSetGVK,
			testhelper.ResourceScopeClusterScoped,
		)).ExecuteTests(t)
}

func TestClusterImageSetWithReleaseImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		releaseImage   string
		expectedError  error
		invalidBuilder bool
	}{
		{
			name:         "sets release image",
			releaseImage: "release-2-1",
		},
		{
			name:          "rejects empty release image",
			releaseImage:  "",
			expectedError: fmt.Errorf("cannot set releaseImage to empty string"),
		},
		{
			name:           "invalid builder stops processing",
			releaseImage:   "release-2-1",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := clients.GetTestClients(clients.TestClientParams{})

			var builder *ClusterImageSetBuilder
			if testCase.invalidBuilder {
				builder = NewClusterImageSetBuilder(apiClient, "", defaultClusterImageSetRelease)
			} else {
				builder = NewClusterImageSetBuilder(apiClient, defaultClusterImageSetName, defaultClusterImageSetRelease)
			}

			result := builder.WithReleaseImage(testCase.releaseImage)

			switch {
			case testCase.invalidBuilder:
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
			case testCase.expectedError != nil:
				assert.Equal(t, testCase.expectedError, result.GetError())
				assert.Equal(t, defaultClusterImageSetRelease, result.Definition.Spec.ReleaseImage)
			default:
				assert.NoError(t, result.GetError())
				assert.Equal(t, testCase.releaseImage, result.Definition.Spec.ReleaseImage)
			}
		})
	}
}

// buildClusterImageSetClientWithDummyObject returns a test client with a ClusterImageSet object.
func buildClusterImageSetClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterImageSetConfig(),
		SchemeAttachers: testSchemes,
	})
}

// buildDummyClusterImageSetConfig returns a ClusterImageSet runtime object for tests.
func buildDummyClusterImageSetConfig() []runtime.Object {
	return append([]runtime.Object{}, &hivev1.ClusterImageSet{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "999",
			Name:            defaultClusterImageSetName,
		},
		Spec: hivev1.ClusterImageSetSpec{},
	})
}
