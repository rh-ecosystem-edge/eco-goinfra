package storage

import (
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultPVCName      = "persistentvolumeclaim-test"
	defaultPVCNamespace = "persistentvolumeclaim-namespace"
)

func TestNewPersistentVolumeClaimBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewPVCBuilder,
		corev1.AddToScheme,
		persistentVolumeClaimGVK,
	).ExecuteTests(t)
}

func TestPullPersistentVolumeClaim(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPersistentVolumeClaim,
		corev1.AddToScheme,
		persistentVolumeClaimGVK,
	).ExecuteTests(t)
}

func TestPersistentVolumeClaimBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[corev1.PersistentVolumeClaim, PVCBuilder](
		corev1.AddToScheme,
		persistentVolumeClaimGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestPersistentVolumeClaimDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		claimExists   bool
		expectedError func(error) bool
	}{
		{
			name:          "deletes existing claim",
			valid:         true,
			claimExists:   true,
			expectedError: func(err error) bool { return err == nil },
		},
		{
			name:          "succeeds when claim is absent",
			valid:         true,
			claimExists:   false,
			expectedError: func(err error) bool { return err == nil },
		},
		{
			name:          "invalid builder short circuits",
			valid:         false,
			claimExists:   true,
			expectedError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var apiClient *clients.Settings
			if testCase.claimExists {
				apiClient = buildTestClientWithDummyPVC()
			} else {
				apiClient = buildTestClientWithPVCScheme()
			}

			var builder *PVCBuilder
			if testCase.valid {
				builder = buildValidPVCTestBuilder(apiClient)
			} else {
				builder = buildInvalidPVCTestBuilder(apiClient)
			}

			err := builder.DeleteAndWait(time.Second)

			require.Truef(t, testCase.expectedError(err), "unexpected error: %v", err)

			if err == nil {
				assert.Nil(t, builder.GetObject())
			}
		})
	}
}

func TestPersistentVolumeClaimWithVolumeMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		volumeMode     string
		invalidBuilder bool
		expectedError  func(error) bool
	}{
		{
			name:          "sets block volume mode",
			volumeMode:    "Block",
			expectedError: func(err error) bool { return err == nil },
		},
		{
			name:          "sets filesystem volume mode",
			volumeMode:    "Filesystem",
			expectedError: func(err error) bool { return err == nil },
		},
		{
			name:       "unsupported volume mode returns error",
			volumeMode: "Disk",
			expectedError: func(err error) bool {
				return err != nil && err.Error() == "unsupported VolumeMode \"Disk\" requested for "+
					"persistentvolumeclaim-test PersistentVolumeClaim in persistentvolumeclaim-test namespace"
			},
		},
		{
			name:       "empty volume mode returns error",
			volumeMode: "",
			expectedError: func(err error) bool {
				return err != nil && err.Error() == "empty volumeMode requested for the PersistentVolumeClaim "+
					"persistentvolumeclaim-test in persistentvolumeclaim-namespace namespace"
			},
		},
		{
			name:           "invalid builder short circuits",
			volumeMode:     "Block",
			invalidBuilder: true,
			expectedError:  commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPVCTestBuilder(buildTestClientWithDummyPVC())
			if testCase.invalidBuilder {
				builder = buildInvalidPVCTestBuilder(buildTestClientWithDummyPVC())
			}

			result, err := builder.WithVolumeMode(testCase.volumeMode)
			require.Same(t, builder, result)
			require.Truef(t, testCase.expectedError(err), "unexpected error: %v", err)

			if err == nil {
				require.NotNil(t, result.Definition.Spec.VolumeMode)
				assert.Equal(t, corev1.PersistentVolumeMode(testCase.volumeMode), *result.Definition.Spec.VolumeMode)
			} else {
				assert.Nil(t, result.Definition.Spec.VolumeMode)
			}
		})
	}
}

func TestPersistentVolumeClaimWithStorageClass(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		storageClass   string
		invalidBuilder bool
		expectedError  func(error) bool
	}{
		{
			name:          "sets storage class",
			storageClass:  "test-storage-class",
			expectedError: func(err error) bool { return err == nil },
		},
		{
			name:         "empty storage class returns error",
			storageClass: "",
			expectedError: func(err error) bool {
				return err != nil && err.Error() ==
					"empty storageClass requested for the PersistentVolumeClaim persistentvolumeclaim-test"
			},
		},
		{
			name:           "invalid builder short circuits",
			storageClass:   "test-storage-class",
			invalidBuilder: true,
			expectedError:  commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPVCTestBuilder(buildTestClientWithDummyPVC())
			if testCase.invalidBuilder {
				builder = buildInvalidPVCTestBuilder(buildTestClientWithDummyPVC())
			}

			result, err := builder.WithStorageClass(testCase.storageClass)
			require.Same(t, builder, result)
			require.Truef(t, testCase.expectedError(err), "unexpected error: %v", err)

			if err == nil {
				require.NotNil(t, result.Definition.Spec.StorageClassName)
				assert.Equal(t, testCase.storageClass, *result.Definition.Spec.StorageClassName)
			} else {
				assert.Nil(t, result.Definition.Spec.StorageClassName)
			}
		})
	}
}

func TestPersistentVolumeClaimWithPVCCapacity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		capacity       string
		invalidBuilder bool
		expectedError  func(error) bool
	}{
		{
			name:          "sets capacity",
			capacity:      "5Gi",
			expectedError: func(err error) bool { return err == nil },
		},
		{
			name:     "empty capacity returns error",
			capacity: "",
			expectedError: func(err error) bool {
				return err != nil && err.Error() == "capacity of the PersistentVolumeClaim is empty"
			},
		},
		{
			name:     "invalid capacity returns error",
			capacity: "invalid",
			expectedError: func(err error) bool {
				return err != nil && err.Error() == "failed to parse: invalid"
			},
		},
		{
			name:           "invalid builder short circuits",
			capacity:       "5Gi",
			invalidBuilder: true,
			expectedError:  commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPVCTestBuilder(buildTestClientWithDummyPVC())
			if testCase.invalidBuilder {
				builder = buildInvalidPVCTestBuilder(buildTestClientWithDummyPVC())
			}

			result, err := builder.WithPVCCapacity(testCase.capacity)
			require.Same(t, builder, result)
			require.Truef(t, testCase.expectedError(err), "unexpected error: %v", err)

			if err == nil {
				capacity := resource.MustParse(testCase.capacity)
				assert.Equal(t, capacity, result.Definition.Spec.Resources.Requests[corev1.ResourceStorage])
			} else {
				assert.Empty(t, result.Definition.Spec.Resources.Requests)
			}
		})
	}
}

func TestPersistentVolumeClaimWithPVCAccessMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		accessMode     string
		invalidBuilder bool
		expectedError  func(error) bool
	}{
		{
			name:          "sets access mode",
			accessMode:    accessModeReadWriteOnce,
			expectedError: func(err error) bool { return err == nil },
		},
		{
			name:       "empty access mode returns error",
			accessMode: "",
			expectedError: func(err error) bool {
				return err != nil && err.Error() == "empty accessMode for PVC requested"
			},
		},
		{
			name:       "invalid access mode returns error",
			accessMode: "Invalid",
			expectedError: func(err error) bool {
				return err != nil && err.Error() == "invalid accessMode for PVC Invalid"
			},
		},
		{
			name:           "invalid builder short circuits",
			accessMode:     accessModeReadWriteOnce,
			invalidBuilder: true,
			expectedError:  commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPVCTestBuilder(buildTestClientWithDummyPVC())
			if testCase.invalidBuilder {
				builder = buildInvalidPVCTestBuilder(buildTestClientWithDummyPVC())
			}

			result, err := builder.WithPVCAccessMode(testCase.accessMode)
			require.Same(t, builder, result)
			require.Truef(t, testCase.expectedError(err), "unexpected error: %v", err)

			if err == nil {
				require.Len(t, result.Definition.Spec.AccessModes, 1)
				assert.Equal(t, corev1.PersistentVolumeAccessMode(testCase.accessMode),
					result.Definition.Spec.AccessModes[0])
			} else {
				assert.Empty(t, result.Definition.Spec.AccessModes)
			}
		})
	}
}

func TestGetPersistentVolumeClaimGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "persistentvolumeclaims", GetPersistentVolumeClaimGVR().Resource)
}

func buildValidPVCTestBuilder(apiClient *clients.Settings) *PVCBuilder {
	return NewPVCBuilder(apiClient, defaultPVCName, defaultPVCNamespace)
}

func buildInvalidPVCTestBuilder(apiClient *clients.Settings) *PVCBuilder {
	return NewPVCBuilder(apiClient, "", defaultPVCNamespace)
}

func buildTestClientWithDummyPVC() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPersistentVolumeClaim(),
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
}

func buildTestClientWithPVCScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
}

func buildDummyPersistentVolumeClaim() []runtime.Object {
	return []runtime.Object{
		&corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultPVCName,
				Namespace: defaultPVCNamespace,
			},
		},
	}
}
