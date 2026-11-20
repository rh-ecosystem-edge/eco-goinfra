package olm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListCatalogSources(t *testing.T) {
	testCases := []struct {
		catalogSource []*CatalogSourceBuilder
		nsName        string
		listOptions   []client.ListOptions
		expectedError error
		client        bool
	}{
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			client:        true,
		},
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "",
			expectedError: fmt.Errorf("failed to list catalogsource, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("failed to list catalogSource, 'apiClient' parameter is empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyCatalogSource(),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := ListCatalogSources(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.catalogSource))
		}
	}
}

func TestListClusterServiceVersion(t *testing.T) {
	testCases := []struct {
		clusterVersion []*ClusterServiceVersionBuilder
		nsName         string
		listOptions    []client.ListOptions
		expectedError  error
		client         bool
	}{
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "",
			expectedError: fmt.Errorf("failed to list clusterserviceversion, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("clusterserviceversion 'apiClient' cannot be empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyClusterService(nil, ""),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := ListClusterServiceVersion(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.clusterVersion))
		}
	}
}

func TestListClusterServiceVersionWithNamePattern(t *testing.T) {
	testCases := []struct {
		clusterVersion []*ClusterServiceVersionBuilder
		nsName         string
		namePattern    string
		listOptions    []client.ListOptions
		expectedError  error
		client         bool
	}{
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "cluster",
			expectedError: nil,
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "",
			namePattern:   "cluster",
			expectedError: fmt.Errorf("failed to list clusterserviceversion, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "",
			expectedError: fmt.Errorf("the namePattern field to filter out all relevant clusterserviceversion cannot be empty"),
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "cluster",
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "cluster",
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "cluster",
			expectedError: fmt.Errorf("clusterserviceversion 'apiClient' cannot be empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyClusterService(nil, ""),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := ListClusterServiceVersionWithNamePattern(
			testSettings, testCase.namePattern, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.clusterVersion))
		}
	}
}

func TestListClusterServiceVersionInAllNamespaces(t *testing.T) {
	testCases := []struct {
		clusterVersion []*ClusterServiceVersionBuilder
		listOptions    []client.ListOptions
		expectedError  error
		client         bool
	}{
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			expectedError: nil,
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			expectedError: fmt.Errorf("clusterserviceversion 'apiClient' cannot be empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyClusterService(nil, ""),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := ListClusterServiceVersionInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.clusterVersion))
		}
	}
}

func TestListInstallPlan(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		nsName        string
		listOptions   []client.ListOptions
		objectsExist  bool
		client        bool
		expectedCount int
		expectedError string
		assertError   func(error) bool
	}{
		{
			name:          "valid list",
			nsName:        "test-namespace",
			objectsExist:  true,
			client:        true,
			expectedCount: 1,
		},
		{
			name:        "empty namespace returns error",
			nsName:      "",
			client:      true,
			assertError: commonerrors.IsBuilderNamespaceEmpty,
		},
		{
			name:          "valid list with options",
			nsName:        "test-namespace",
			listOptions:   []client.ListOptions{{Continue: "true"}},
			objectsExist:  true,
			client:        true,
			expectedCount: 1,
		},
		{
			name:          "multiple options return error",
			nsName:        "test-namespace",
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
			expectedError: "error: more than one ListOptions was passed",
		},
		{
			name:        "nil client returns error",
			nsName:      "test-namespace",
			listOptions: []client.ListOptions{},
			client:      false,
			assertError: commonerrors.IsAPIClientNil,
		},
		{
			name:          "missing resources return error",
			nsName:        "test-namespace",
			client:        true,
			expectedError: "installplan not found in namespace test-namespace",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				testSettings   *clients.Settings
				runtimeObjects = []runtime.Object{}
			)

			if testCase.objectsExist {
				runtimeObjects = buildDummyInstallPlan()
			}

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  runtimeObjects,
					SchemeAttachers: testSchemes,
				})
			}

			installPlanBuilders, err := ListInstallPlan(testSettings, testCase.nsName, testCase.listOptions...)

			switch {
			case testCase.assertError != nil:
				assert.Truef(t, testCase.assertError(err), "unexpected error: %v", err)
				assert.Nil(t, installPlanBuilders)
			case testCase.expectedError != "":
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, installPlanBuilders)
			default:
				assert.NoError(t, err)
				assert.Len(t, installPlanBuilders, testCase.expectedCount)
			}
		})
	}
}

func TestListPackageManifest(t *testing.T) {
	testCases := []struct {
		packageManifest []*PackageManifestBuilder
		nsName          string
		listOptions     []client.ListOptions
		expectedError   error
		client          bool
	}{
		{
			packageManifest: []*PackageManifestBuilder{
				buildValidPackageManifestBuilder(buildPackageManifestTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			client:        true,
		},
		{
			packageManifest: []*PackageManifestBuilder{
				buildValidPackageManifestBuilder(buildPackageManifestTestClientWithDummyObject())},
			nsName:        "",
			expectedError: fmt.Errorf("failed to list packagemanifests, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			packageManifest: []*PackageManifestBuilder{
				buildValidPackageManifestBuilder(buildPackageManifestTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			packageManifest: []*PackageManifestBuilder{
				buildValidPackageManifestBuilder(buildPackageManifestTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			packageManifest: []*PackageManifestBuilder{
				buildValidPackageManifestBuilder(buildPackageManifestTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("failed to list packageManifest, 'apiClient' parameter is empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyPackageManifest(),
				SchemeAttachers: operatorsv1Scheme,
			})
		}

		packageManifests, err := ListPackageManifest(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(packageManifests), len(testCase.packageManifest))
		}
	}
}
