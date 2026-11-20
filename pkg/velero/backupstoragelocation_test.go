package velero

import (
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testBackupStorageLocation          = "test-bsl"
	testBackupStorageLocationNamespace = "test-ns"
	testBackupStorageLocationOtherNS   = "other-ns"
	testBackupStorageLocationProvider  = "aws"
	testBackupStorageLocationBucket    = "test-bucket"
)

func TestNewBackupStorageLocationBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, namespace string) *BackupStorageLocationBuilder {
				return NewBackupStorageLocationBuilder(
					apiClient,
					name,
					namespace,
					testBackupStorageLocationProvider,
					velerov1.ObjectStorageLocation{Bucket: testBackupStorageLocationBucket},
				)
			},
			velerov1.AddToScheme,
			backupStorageLocationGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		provider      string
		objectStorage velerov1.ObjectStorageLocation
		expectedError error
	}{
		{
			name:     "valid required fields",
			provider: testBackupStorageLocationProvider,
			objectStorage: velerov1.ObjectStorageLocation{
				Bucket: testBackupStorageLocationBucket,
				Prefix: "backups",
			},
		},
		{
			name:          "empty provider",
			objectStorage: velerov1.ObjectStorageLocation{Bucket: testBackupStorageLocationBucket},
			expectedError: fmt.Errorf("backupstoragelocation provider cannot be empty"),
		},
		{
			name:          "empty object storage bucket",
			provider:      testBackupStorageLocationProvider,
			objectStorage: velerov1.ObjectStorageLocation{Prefix: "backups"},
			expectedError: fmt.Errorf("backupstoragelocation objectstorage bucket cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewBackupStorageLocationBuilder(
				newBackupStorageLocationTestClient(), testBackupStorageLocation, testBackupStorageLocationNamespace,
				testCase.provider, testCase.objectStorage)

			require.NotNil(t, builder)
			assert.Equal(t, testCase.expectedError, builder.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.provider, builder.Definition.Spec.Provider)
				require.NotNil(t, builder.Definition.Spec.ObjectStorage)
				assert.Equal(t, testCase.objectStorage, *builder.Definition.Spec.ObjectStorage)
			}
		})
	}
}

func TestPullBackupStorageLocation(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullBackupStorageLocationBuilder, velerov1.AddToScheme, backupStorageLocationGVK).ExecuteTests(t)
}

func TestListBackupStorageLocationBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		ListBackupStorageLocationBuilder, velerov1.AddToScheme, backupStorageLocationGVK).ExecuteTests(t)
}

func TestListBackupStorageLocationBuilderScopeAndOptions(t *testing.T) {
	t.Parallel()

	selectedLabels := map[string]string{"selected": "true"}
	otherLabels := map[string]string{"selected": "false"}
	selected := buildDummyBackupStorageLocation("selected", testBackupStorageLocationNamespace)
	selected.Labels = selectedLabels
	notSelected := buildDummyBackupStorageLocation("not-selected", testBackupStorageLocationNamespace)
	notSelected.Labels = otherLabels
	otherNamespace := buildDummyBackupStorageLocation("other-namespace", testBackupStorageLocationOtherNS)
	otherNamespace.Labels = selectedLabels
	testClient := newBackupStorageLocationTestClient(selected, notSelected, otherNamespace)

	t.Run("scopes results to the requested namespace", func(t *testing.T) {
		t.Parallel()

		builders, err := ListBackupStorageLocationBuilder(testClient, testBackupStorageLocationNamespace)
		require.NoError(t, err)
		assert.Len(t, builders, 2)

		for _, builder := range builders {
			assert.Equal(t, testBackupStorageLocationNamespace, builder.Definition.Namespace)
		}
	})

	t.Run("applies the single list option with the namespace", func(t *testing.T) {
		t.Parallel()

		builders, err := ListBackupStorageLocationBuilder(
			testClient,
			testBackupStorageLocationNamespace,
			runtimeclient.ListOptions{LabelSelector: labels.SelectorFromSet(selectedLabels)},
		)
		require.NoError(t, err)
		require.Len(t, builders, 1)
		assert.Equal(t, "selected", builders[0].Definition.Name)
		assert.Equal(t, testBackupStorageLocationNamespace, builders[0].Definition.Namespace)
	})

	t.Run("rejects multiple list options", func(t *testing.T) {
		t.Parallel()

		builders, err := ListBackupStorageLocationBuilder(
			testClient,
			testBackupStorageLocationNamespace,
			runtimeclient.ListOptions{},
			runtimeclient.ListOptions{},
		)
		assert.EqualError(t, err, "error: more than one ListOptions was passed")
		assert.Nil(t, builders)
	})
}

func TestBackupStorageLocationBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[velerov1.BackupStorageLocation, BackupStorageLocationBuilder](
		velerov1.AddToScheme,
		backupStorageLocationGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestBackupStorageLocationWithConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		config        map[string]string
		valid         bool
		expectedError error
	}{
		{
			name:   "valid config",
			config: map[string]string{"testConfig": "testData"},
			valid:  true,
		},
		{
			name:          "empty config",
			config:        map[string]string{},
			valid:         true,
			expectedError: fmt.Errorf("backupstoragelocation cannot have empty config"),
		},
		{
			name:  "invalid builder",
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testClient := newBackupStorageLocationTestClient()

			var builder *BackupStorageLocationBuilder
			if testCase.valid {
				builder = buildValidBackupStorageLocationTestBuilder(testClient)
			} else {
				builder = buildInvalidBackupStorageLocationTestBuilder(testClient)
			}

			result := builder.WithConfig(testCase.config)

			require.Same(t, builder, result)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.config, result.Definition.Spec.Config)
			}
		})
	}
}

func TestBackupStorageLocationWaitUntilAvailable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		valid         bool
		phase         velerov1.BackupStorageLocationPhase
		expectedError string
	}{
		{
			name:   "available",
			exists: true,
			valid:  true,
			phase:  velerov1.BackupStorageLocationPhaseAvailable,
		},
		{
			name:          "timeout",
			exists:        true,
			valid:         true,
			phase:         velerov1.BackupStorageLocationPhaseUnavailable,
			expectedError: "error waiting for backupstoragelocation to become available",
		},
		{
			name:          "does not exist",
			valid:         true,
			expectedError: "cannot wait for backupstoragelocation that does not exist",
		},
		{
			name:   "invalid builder",
			exists: true,
			valid:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			objects := buildBackupStorageLocationWaitObjects(testCase.exists, testCase.phase)
			testClient := newBackupStorageLocationTestClient(objects...)

			var builder *BackupStorageLocationBuilder
			if testCase.valid {
				builder = buildValidBackupStorageLocationTestBuilder(testClient)
			} else {
				builder = buildInvalidBackupStorageLocationTestBuilder(testClient)
			}

			result, err := builder.WaitUntilAvailable(time.Second)

			switch {
			case !testCase.valid:
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
			case testCase.expectedError != "":
				require.ErrorContains(t, err, testCase.expectedError)
			default:
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, velerov1.BackupStorageLocationPhaseAvailable, result.Object.Status.Phase)
			}
		})
	}
}

func TestBackupStorageLocationWaitUntilUnavailable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		valid         bool
		phase         velerov1.BackupStorageLocationPhase
		expectedError string
	}{
		{
			name:   "unavailable",
			exists: true,
			valid:  true,
			phase:  velerov1.BackupStorageLocationPhaseUnavailable,
		},
		{
			name:          "timeout",
			exists:        true,
			valid:         true,
			phase:         velerov1.BackupStorageLocationPhaseAvailable,
			expectedError: "error waiting for backupstoragelocation to become unavailable",
		},
		{
			name:          "does not exist",
			valid:         true,
			expectedError: "cannot wait for backupstoragelocation that does not exist",
		},
		{
			name:   "invalid builder",
			exists: true,
			valid:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			objects := buildBackupStorageLocationWaitObjects(testCase.exists, testCase.phase)
			testClient := newBackupStorageLocationTestClient(objects...)

			var builder *BackupStorageLocationBuilder
			if testCase.valid {
				builder = buildValidBackupStorageLocationTestBuilder(testClient)
			} else {
				builder = buildInvalidBackupStorageLocationTestBuilder(testClient)
			}

			result, err := builder.WaitUntilUnavailable(time.Second)

			switch {
			case !testCase.valid:
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
			case testCase.expectedError != "":
				require.ErrorContains(t, err, testCase.expectedError)
			default:
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, velerov1.BackupStorageLocationPhaseUnavailable, result.Object.Status.Phase)
			}
		})
	}
}

// newBackupStorageLocationTestClient returns a fake client configured for Velero resources.
func newBackupStorageLocationTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{velerov1.AddToScheme},
	})
}

// buildValidBackupStorageLocationTestBuilder returns a valid BackupStorageLocation builder for tests.
func buildValidBackupStorageLocationTestBuilder(apiClient *clients.Settings) *BackupStorageLocationBuilder {
	return NewBackupStorageLocationBuilder(
		apiClient,
		testBackupStorageLocation,
		testBackupStorageLocationNamespace,
		testBackupStorageLocationProvider,
		velerov1.ObjectStorageLocation{Bucket: testBackupStorageLocationBucket},
	)
}

// buildInvalidBackupStorageLocationTestBuilder returns an error-bearing BackupStorageLocation builder for tests.
func buildInvalidBackupStorageLocationTestBuilder(apiClient *clients.Settings) *BackupStorageLocationBuilder {
	return NewBackupStorageLocationBuilder(
		apiClient,
		testBackupStorageLocation,
		"",
		testBackupStorageLocationProvider,
		velerov1.ObjectStorageLocation{Bucket: testBackupStorageLocationBucket},
	)
}

// buildBackupStorageLocationWaitObjects returns mock objects for wait tests.
func buildBackupStorageLocationWaitObjects(
	exists bool, phase velerov1.BackupStorageLocationPhase) []runtime.Object {
	if !exists {
		return nil
	}

	backupStorageLocation := buildDummyBackupStorageLocation(
		testBackupStorageLocation, testBackupStorageLocationNamespace)
	backupStorageLocation.Status.Phase = phase

	return []runtime.Object{backupStorageLocation}
}

// buildDummyBackupStorageLocation returns a BackupStorageLocation with the provided identity.
func buildDummyBackupStorageLocation(name, namespace string) *velerov1.BackupStorageLocation {
	return &velerov1.BackupStorageLocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: velerov1.BackupStorageLocationSpec{
			Provider: testBackupStorageLocationProvider,
			StorageType: velerov1.StorageType{
				ObjectStorage: &velerov1.ObjectStorageLocation{
					Bucket: testBackupStorageLocationBucket,
					Prefix: "backup",
				},
			},
		},
	}
}
