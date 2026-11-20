package olm

import (
	"fmt"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	operatorsv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/package-server/operators/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPullPackageManifest(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPackageManifest, operatorsv1.AddToScheme, packageManifestGVK).ExecuteTests(t)
}

//nolint:funlen
func TestPullPackageManifestByCatalog(t *testing.T) {
	t.Parallel()

	expectError := func(message string) func(error) bool {
		return func(err error) bool {
			return err != nil && err.Error() == message
		}
	}

	testCases := []struct {
		description         string
		name                string
		namespace           string
		catalog             string
		addToRuntimeObjects bool
		assertError         func(error) bool
		client              bool
	}{
		{
			description:         "matching catalog",
			name:                "packagemanifest-test",
			namespace:           "test-namespace",
			catalog:             "test",
			addToRuntimeObjects: true,
			assertError:         func(err error) bool { return err == nil },
			client:              true,
		},
		{
			description:         "non-matching catalog",
			name:                "packagemanifest-test",
			namespace:           "test-namespace",
			catalog:             "invalid",
			addToRuntimeObjects: true,
			assertError:         expectError("no matching PackageManifests were found"),
			client:              true,
		},
		{
			description:         "empty name",
			name:                "",
			namespace:           "test-namespace",
			catalog:             "test",
			addToRuntimeObjects: true,
			assertError:         expectError("packageManifest 'name' cannot be empty"),
			client:              true,
		},
		{
			description:         "empty namespace",
			name:                "packagemanifest",
			namespace:           "",
			catalog:             "test",
			addToRuntimeObjects: true,
			assertError:         commonerrors.IsBuilderNamespaceEmpty,
			client:              true,
		},
		{
			description:         "empty catalog",
			name:                "packagemanifest",
			namespace:           "test-namespace",
			catalog:             "",
			addToRuntimeObjects: true,
			assertError:         expectError("packageManifest 'catalog' cannot be empty"),
			client:              true,
		},
		{
			description:         "resource does not exist",
			name:                "packagemanifest",
			namespace:           "test-namespace",
			catalog:             "test",
			addToRuntimeObjects: false,
			assertError:         expectError("no matching PackageManifests were found"),
			client:              true,
		},
		{
			description:         "nil client",
			name:                "packagemanifest",
			namespace:           "test-namespace",
			catalog:             "test",
			addToRuntimeObjects: true,
			assertError:         commonerrors.IsAPIClientNil,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			var testSettings *clients.Settings

			packageManifest := buildPackageManifestDefinition(testCase.name, testCase.namespace)

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects, packageManifest)
			}

			if testCase.client {
				var builder *fake.ClientBuilder

				testSettings, builder = clients.GetModifiableTestClients(clients.TestClientParams{
					K8sMockObjects:  runtimeObjects,
					SchemeAttachers: []clients.SchemeAttacher{operatorsv1.AddToScheme}})

				testSettings.Client = builder.WithIndex(
					&operatorsv1.PackageManifest{}, "metadata.name", nameMetadataIndexer).Build()
			}

			builderResult, err := PullPackageManifestByCatalog(
				testSettings, testCase.name, testCase.namespace, testCase.catalog)
			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)

			if err == nil {
				require.NotNil(t, builderResult)
				assert.Equal(t, testCase.name, builderResult.Object.Name)
				assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
			} else {
				assert.Nil(t, builderResult)
			}
		})
	}
}

func TestPackageManifestMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[operatorsv1.PackageManifest, PackageManifestBuilder](
		operatorsv1.AddToScheme,
		packageManifestGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func buildPackageManifestDefinition(name, nsName string) *operatorsv1.PackageManifest {
	return &operatorsv1.PackageManifest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsName,
			Labels:    map[string]string{"catalog": "test"},
		},
	}
}

func nameMetadataIndexer(obj client.Object) []string {
	packageManifest, found := obj.(*operatorsv1.PackageManifest)
	if !found {
		panic(fmt.Errorf("indexer function for type %T's spec.replicas field received"+
			" object of type %T, this should never happen", operatorsv1.PackageManifest{}, obj))
	}

	return []string{packageManifest.Name}
}
