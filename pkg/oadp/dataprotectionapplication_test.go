package oadp

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	oadpv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/oadp/api/v1alpha1"
	velerov1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/oadp/velero/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const testDataProtectionApplication = "test-dataprotectionapplication"

func TestNewDPABuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *DPABuilder {
				return NewDPABuilder(apiClient, name, nsname, buildTestApplicationConfig())
			},
			oadpv1alpha1.AddToScheme,
			dataProtectionApplicationGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		config        oadpv1alpha1.ApplicationConfig
		expectedError error
	}{
		{
			name:          "valid velero config",
			config:        buildTestApplicationConfig(),
			expectedError: nil,
		},
		{
			name:          "empty velero config",
			config:        oadpv1alpha1.ApplicationConfig{},
			expectedError: fmt.Errorf("dataprotectionapplication velero config cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := NewDPABuilder(
				clients.GetTestClients(clients.TestClientParams{}),
				testDataProtectionApplication,
				testDataProtectionApplication,
				testCase.config,
			)

			assert.Equal(t, testCase.expectedError, testBuilder.GetError())

			if testCase.expectedError == nil {
				require.NotNil(t, testBuilder.Definition.Spec.Configuration)
				assert.Equal(t, testCase.config, *testBuilder.Definition.Spec.Configuration)
			}
		})
	}
}

func TestPullDPA(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullDPA, oadpv1alpha1.AddToScheme, dataProtectionApplicationGVK).ExecuteTests(t)
}

func TestListDataProtectionApplication(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		ListDataProtectionApplication, oadpv1alpha1.AddToScheme, dataProtectionApplicationGVK).ExecuteTests(t)
}

func TestListDataProtectionApplicationWithMultipleOptions(t *testing.T) {
	t.Parallel()

	_, err := ListDataProtectionApplication(
		clients.GetTestClients(clients.TestClientParams{}),
		testDataProtectionApplication,
		runtimeClient.ListOptions{},
		runtimeClient.ListOptions{},
	)

	assert.EqualError(t, err, "error: more than one ListOptions was passed")
}

func TestDPABuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[oadpv1alpha1.DataProtectionApplication, DPABuilder](
		oadpv1alpha1.AddToScheme,
		dataProtectionApplicationGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestDPAWithBackupLocation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		backupLocation oadpv1alpha1.BackupLocation
		validBuilder   bool
		expectedError  error
	}{
		{
			name:           "valid backup location",
			backupLocation: buildTestBackupLocation(),
			validBuilder:   true,
			expectedError:  nil,
		},
		{
			name:           "empty velero config",
			backupLocation: oadpv1alpha1.BackupLocation{},
			validBuilder:   true,
			expectedError: fmt.Errorf(
				"dataprotectionapplication backuplocation cannot have empty velero config"),
		},
		{
			name:           "invalid builder",
			backupLocation: buildTestBackupLocation(),
			validBuilder:   false,
			expectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := clients.GetTestClients(clients.TestClientParams{})

			var testBuilder *DPABuilder
			if testCase.validBuilder {
				testBuilder = buildValidDPATestBuilder(apiClient)
			} else {
				testBuilder = buildInvalidDPATestBuilder(apiClient)
			}

			testBuilder = testBuilder.WithBackupLocation(testCase.backupLocation)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))

				return
			}

			assert.Equal(t, testCase.expectedError, testBuilder.GetError())

			if testCase.expectedError == nil {
				require.Len(t, testBuilder.Definition.Spec.BackupLocations, 1)
				assert.Equal(t, testCase.backupLocation, testBuilder.Definition.Spec.BackupLocations[0])
			}
		})
	}
}

func buildValidDPATestBuilder(apiClient *clients.Settings) *DPABuilder {
	return NewDPABuilder(apiClient, testDataProtectionApplication, testDataProtectionApplication,
		buildTestApplicationConfig())
}

func buildInvalidDPATestBuilder(apiClient *clients.Settings) *DPABuilder {
	return NewDPABuilder(apiClient, testDataProtectionApplication, "", buildTestApplicationConfig())
}

func buildTestApplicationConfig() oadpv1alpha1.ApplicationConfig {
	return oadpv1alpha1.ApplicationConfig{
		Velero: &oadpv1alpha1.VeleroConfig{
			DefaultPlugins: []oadpv1alpha1.DefaultPlugin{
				oadpv1alpha1.DefaultPluginAWS,
				oadpv1alpha1.DefaultPluginOpenShift,
			},
			ResourceTimeout: "10m",
		},
	}
}

func buildTestBackupLocation() oadpv1alpha1.BackupLocation {
	return oadpv1alpha1.BackupLocation{
		Velero: &velerov1.BackupStorageLocationSpec{
			Provider: "aws",
			StorageType: velerov1.StorageType{
				ObjectStorage: &velerov1.ObjectStorageLocation{
					Bucket: "test-bucket",
					Prefix: "backup",
				},
			},
			Config: map[string]string{
				"insecureSkipTLSVerify": "true",
				"s3Url":                 "http://example.com/",
			},
		},
	}
}
