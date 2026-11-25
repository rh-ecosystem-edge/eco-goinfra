package storage

import (
	"context"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultPersistentVolumeName = "persistentvolume-test"

func TestPullPersistentVolume(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullPersistentVolume,
		corev1.AddToScheme,
		persistentVolumeGVK,
	).ExecuteTests(t)
}

func TestPersistentVolumeBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[corev1.PersistentVolume, PVBuilder](
		corev1.AddToScheme,
		persistentVolumeGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestPersistentVolumeDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		objectExists   bool
		invalidBuilder bool
	}{
		{
			name:         "deletes existing persistent volume",
			objectExists: true,
		},
		{
			name: "persistent volume already absent",
		},
		{
			name:           "invalid builder short circuits",
			objectExists:   true,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.objectExists {
				objects = append(objects, buildDummyPersistentVolume(defaultPersistentVolumeName))
			}

			apiClient := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  objects,
				SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
			})

			builder := buildValidPersistentVolumeTestBuilder(apiClient)
			if testCase.invalidBuilder {
				builder = buildInvalidPersistentVolumeTestBuilder(apiClient)
			}

			err := builder.DeleteAndWait(time.Second)

			if testCase.invalidBuilder {
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))

				return
			}

			require.NoError(t, err)
			assert.Nil(t, builder.Object)
		})
	}
}

func TestPersistentVolumeWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		objectExists   bool
		invalidBuilder bool
		expectTimeout  bool
	}{
		{
			name:          "times out while persistent volume remains",
			objectExists:  true,
			expectTimeout: true,
		},
		{
			name: "persistent volume already absent",
		},
		{
			name:           "invalid builder short circuits",
			objectExists:   true,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.objectExists {
				objects = append(objects, buildDummyPersistentVolume(defaultPersistentVolumeName))
			}

			apiClient := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  objects,
				SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
			})

			builder := buildValidPersistentVolumeTestBuilder(apiClient)
			if testCase.invalidBuilder {
				builder = buildInvalidPersistentVolumeTestBuilder(apiClient)
			}

			err := builder.WaitUntilDeleted(time.Second)

			switch {
			case testCase.invalidBuilder:
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
			case testCase.expectTimeout:
				assert.Equal(t, context.DeadlineExceeded, err)
			default:
				require.NoError(t, err)
			}
		})
	}
}

// buildDummyPersistentVolume returns a PersistentVolume with the specified name.
func buildDummyPersistentVolume(name string) *corev1.PersistentVolume {
	return &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildValidPersistentVolumeTestBuilder returns a valid PVBuilder with the default name and specified client.
func buildValidPersistentVolumeTestBuilder(apiClient *clients.Settings) *PVBuilder {
	return common.NewClusterScopedBuilder[corev1.PersistentVolume, PVBuilder](
		apiClient, corev1.AddToScheme, defaultPersistentVolumeName)
}

// buildInvalidPersistentVolumeTestBuilder returns an invalid PVBuilder with an empty name and the specified client.
func buildInvalidPersistentVolumeTestBuilder(apiClient *clients.Settings) *PVBuilder {
	return common.NewClusterScopedBuilder[corev1.PersistentVolume, PVBuilder](
		apiClient, corev1.AddToScheme, "")
}
