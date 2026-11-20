package velero

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

const (
	defaultBackupName      = "backup-test-name"
	defaultBackupNamespace = "backup-test-namespace"
)

var v1TestSchemes = []clients.SchemeAttacher{
	velerov1.AddToScheme,
}

func TestNewBackupBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(NewBackupBuilder, velerov1.AddToScheme, backupGVK).ExecuteTests(t)
}

func TestPullBackup(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(PullBackup, velerov1.AddToScheme, backupGVK).ExecuteTests(t)
}

func TestBackupBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[velerov1.Backup, BackupBuilder](
		velerov1.AddToScheme,
		backupGVK,
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

func TestWithStorageLocation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		location      string
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "storage location set",
			location:     "default",
			validBuilder: true,
		},
		{
			name:          "empty storage location",
			location:      "",
			validBuilder:  true,
			expectedError: "backup storage location cannot be an empty string",
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

			testBuilder := buildBackupTestBuilder(testCase.validBuilder)
			initialError := testBuilder.GetError()
			result := testBuilder.WithStorageLocation(testCase.location)

			require.Same(t, testBuilder, result)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Labels["velero.io/storage-location"])

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Labels)

				return
			}

			assert.NoError(t, result.GetError())
			assert.Equal(t, testCase.location, result.Definition.Labels["velero.io/storage-location"])
		})
	}
}

func TestWithIncludedNamespace(t *testing.T) {
	t.Parallel()

	testBackupStringSliceModifier(t,
		func(builder *BackupBuilder, value string) *BackupBuilder {
			return builder.WithIncludedNamespace(value)
		},
		func(builder *BackupBuilder) []string {
			return builder.Definition.Spec.IncludedNamespaces
		},
		"backup includedNamespace cannot be an empty string",
	)
}

func TestWithIncludedClusterScopedResource(t *testing.T) {
	t.Parallel()

	testBackupStringSliceModifier(t,
		func(builder *BackupBuilder, value string) *BackupBuilder {
			return builder.WithIncludedClusterScopedResource(value)
		},
		func(builder *BackupBuilder) []string {
			return builder.Definition.Spec.IncludedClusterScopedResources
		},
		"backup includedClusterScopedResource cannot be an empty string",
	)
}

func TestWithIncludedNamespaceScopedResource(t *testing.T) {
	t.Parallel()

	testBackupStringSliceModifier(t,
		func(builder *BackupBuilder, value string) *BackupBuilder {
			return builder.WithIncludedNamespaceScopedResource(value)
		},
		func(builder *BackupBuilder) []string {
			return builder.Definition.Spec.IncludedNamespaceScopedResources
		},
		"backup includedNamespaceScopedResource cannot be an empty string",
	)
}

func TestWithExcludedClusterScopedResource(t *testing.T) {
	t.Parallel()

	testBackupStringSliceModifier(t,
		func(builder *BackupBuilder, value string) *BackupBuilder {
			return builder.WithExcludedClusterScopedResource(value)
		},
		func(builder *BackupBuilder) []string {
			return builder.Definition.Spec.ExcludedClusterScopedResources
		},
		"backup excludedClusterScopedResource cannot be an empty string",
	)
}

func TestWithExcludedNamespaceScopedResources(t *testing.T) {
	t.Parallel()

	testBackupStringSliceModifier(t,
		func(builder *BackupBuilder, value string) *BackupBuilder {
			return builder.WithExcludedNamespaceScopedResources(value)
		},
		func(builder *BackupBuilder) []string {
			return builder.Definition.Spec.ExcludedNamespaceScopedResources
		},
		"backup excludedNamespaceScopedResource cannot be an empty string",
	)
}

func testBackupStringSliceModifier(
	t *testing.T,
	modifier func(*BackupBuilder, string) *BackupBuilder,
	values func(*BackupBuilder) []string,
	expectedValidationError string,
) {
	t.Helper()

	testCases := []struct {
		name          string
		value         string
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "value appended",
			value:        "test-value",
			validBuilder: true,
		},
		{
			name:          "empty value",
			value:         "",
			validBuilder:  true,
			expectedError: expectedValidationError,
		},
		{
			name:         "invalid builder",
			value:        "test-value",
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBackupTestBuilder(testCase.validBuilder)
			initialError := testBuilder.GetError()
			result := modifier(testBuilder, testCase.value)

			require.Same(t, testBuilder, result)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))
				assert.Empty(t, values(result))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Empty(t, values(result))

				return
			}

			assert.NoError(t, result.GetError())
			assert.Equal(t, []string{testCase.value}, values(result))
		})
	}
}

func buildBackupTestBuilder(valid bool) *BackupBuilder {
	apiClient := clients.GetTestClients(clients.TestClientParams{})
	if valid {
		return NewBackupBuilder(apiClient, defaultBackupName, defaultBackupNamespace)
	}

	return NewBackupBuilder(apiClient, defaultBackupName, "")
}
