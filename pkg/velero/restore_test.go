package velero

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

const (
	defaultRestoreName      = "restore-test-name"
	defaultRestoreNamespace = "restore-test-namespace"
	defaultRestoreBackup    = "backup-test-name"
)

func TestNewRestoreBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		func(apiClient *clients.Settings, name, nsname string) *RestoreBuilder {
			return NewRestoreBuilder(apiClient, name, nsname, defaultRestoreBackup)
		},
		velerov1.AddToScheme,
		restoreGVK,
	).ExecuteTests(t)

	t.Run("backup name is set", func(t *testing.T) {
		t.Parallel()

		builder := NewRestoreBuilder(
			clients.GetTestClients(clients.TestClientParams{}),
			defaultRestoreName,
			defaultRestoreNamespace,
			defaultRestoreBackup,
		)

		require.NoError(t, builder.GetError())
		assert.Equal(t, defaultRestoreBackup, builder.Definition.Spec.BackupName)
	})

	t.Run("empty backup name returns error", func(t *testing.T) {
		t.Parallel()

		builder := NewRestoreBuilder(
			clients.GetTestClients(clients.TestClientParams{}),
			defaultRestoreName,
			defaultRestoreNamespace,
			"",
		)

		require.NotNil(t, builder)
		assert.EqualError(t, builder.GetError(), errEmptyBackupName)
	})
}

func TestPullRestore(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(PullRestore, velerov1.AddToScheme, restoreGVK).ExecuteTests(t)
}

func TestRestoreBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[velerov1.Restore, RestoreBuilder](
		velerov1.AddToScheme,
		restoreGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestRestoreWithStorageLocation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		location       string
		validBuilder   bool
		expectedError  string
		expectedLabels map[string]string
	}{
		{
			name:         "storage location set",
			location:     "default",
			validBuilder: true,
			expectedLabels: map[string]string{
				"velero.io/storage-location": "default",
			},
		},
		{
			name:          "empty storage location",
			location:      "",
			validBuilder:  true,
			expectedError: "restore storage location cannot be an empty string",
		},
		{
			name:         "invalid builder",
			location:     "default",
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildRestoreTestBuilder(testCase.validBuilder)
			initialError := testBuilder.GetError()
			result := testBuilder.WithStorageLocation(testCase.location)

			require.Same(t, testBuilder, result)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.Equal(t, initialError, result.GetError())
				assert.Nil(t, result.Definition.Labels)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Labels)

				return
			}

			assert.NoError(t, result.GetError())
			assert.Equal(t, testCase.expectedLabels, result.Definition.Labels)
		})
	}
}

// buildRestoreTestBuilder returns a RestoreBuilder in the requested validity state.
func buildRestoreTestBuilder(valid bool) *RestoreBuilder {
	apiClient := clients.GetTestClients(clients.TestClientParams{})
	if valid {
		return NewRestoreBuilder(apiClient, defaultRestoreName, defaultRestoreNamespace, defaultRestoreBackup)
	}

	return NewRestoreBuilder(apiClient, defaultRestoreName, defaultRestoreNamespace, "")
}
