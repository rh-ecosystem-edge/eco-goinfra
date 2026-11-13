package olm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	oplmV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
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
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...client.ListOptions) (
			[]*ClusterServiceVersionBuilder, error) {
			return ListClusterServiceVersion(apiClient, nsname)
		},
		oplmV1alpha1.AddToScheme,
		clusterServiceVersionGVK,
	).ExecuteTests(t)
}

func TestListClusterServiceVersionRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})

	builders, err := ListClusterServiceVersion(
		testSettings,
		"test-namespace",
		client.ListOptions{Continue: "true"},
		client.ListOptions{Limit: 100},
	)

	assert.EqualError(t, err, "error: more than one ListOptions was passed")
	assert.Nil(t, builders)
}

func TestListClusterServiceVersionWithNamePattern(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...client.ListOptions) (
			[]*ClusterServiceVersionBuilder, error) {
			return ListClusterServiceVersionWithNamePattern(apiClient, "resource", nsname)
		},
		oplmV1alpha1.AddToScheme,
		clusterServiceVersionGVK,
	).ExecuteTests(t)
}

func TestListClusterServiceVersionWithNamePatternSpecificBehavior(t *testing.T) {
	t.Parallel()

	t.Run("filters matching resources", func(t *testing.T) {
		t.Parallel()

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  buildDummyClusterService(nil, ""),
			SchemeAttachers: testSchemes,
		})

		builders, err := ListClusterServiceVersionWithNamePattern(
			testSettings, "cluster", "test-namespace")

		assert.NoError(t, err)
		assert.Len(t, builders, 1)
	})

	t.Run("does not return non-matching resources", func(t *testing.T) {
		t.Parallel()

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  buildDummyClusterService(nil, ""),
			SchemeAttachers: testSchemes,
		})

		builders, err := ListClusterServiceVersionWithNamePattern(
			testSettings, "not-present", "test-namespace")

		assert.NoError(t, err)
		assert.Empty(t, builders)
	})

	t.Run("rejects empty name pattern", func(t *testing.T) {
		t.Parallel()

		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})

		builders, err := ListClusterServiceVersionWithNamePattern(
			testSettings, "", "test-namespace")

		assert.EqualError(t, err,
			"the namePattern field to filter out all relevant clusterserviceversion cannot be empty")
		assert.Nil(t, builders)
	})

	t.Run("rejects nil client", func(t *testing.T) {
		t.Parallel()

		builders, err := ListClusterServiceVersionWithNamePattern(nil, "cluster", "test-namespace")

		assert.True(t, commonerrors.IsAPIClientNil(err))
		assert.Nil(t, builders)
	})

	t.Run("rejects empty namespace", func(t *testing.T) {
		t.Parallel()

		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})

		builders, err := ListClusterServiceVersionWithNamePattern(testSettings, "cluster", "")

		assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
		assert.Nil(t, builders)
	})
}

func TestListClusterServiceVersionInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		func(apiClient *clients.Settings, _ ...client.ListOptions) (
			[]*ClusterServiceVersionBuilder, error) {
			return ListClusterServiceVersionInAllNamespaces(apiClient)
		},
		oplmV1alpha1.AddToScheme,
		clusterServiceVersionGVK,
	).ExecuteTests(t)
}

func TestListClusterServiceVersionInAllNamespacesRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})

	builders, err := ListClusterServiceVersionInAllNamespaces(
		testSettings,
		client.ListOptions{Continue: "true"},
		client.ListOptions{Limit: 100},
	)

	assert.EqualError(t, err, "error: more than one ListOptions was passed")
	assert.Nil(t, builders)
}

func TestListInstallPlan(t *testing.T) {
	testCases := []struct {
		installPlan   []*InstallPlanBuilder
		nsName        string
		listOptions   []client.ListOptions
		expectedError error
		client        bool
	}{
		{
			installPlan: []*InstallPlanBuilder{
				buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			client:        true,
		},
		{
			installPlan: []*InstallPlanBuilder{
				buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject())},
			nsName:        "",
			expectedError: fmt.Errorf("the nsname of the installplan is empty"),
			client:        true,
		},
		{
			installPlan: []*InstallPlanBuilder{
				buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			installPlan: []*InstallPlanBuilder{
				buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			installPlan: []*InstallPlanBuilder{
				buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("failed to list installPlan, 'apiClient' parameter is empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyInstallPlan(),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := ListInstallPlan(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.installPlan))
		}
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
