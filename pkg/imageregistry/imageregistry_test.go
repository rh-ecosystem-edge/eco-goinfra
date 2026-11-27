package imageregistry

import (
	"context"
	"fmt"
	"testing"
	"time"

	imageregistryv1 "github.com/openshift/api/imageregistry/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultImageRegistryName      = "cluster"
	defaultManagementState        = operatorv1.Managed
	defaultImageRegistryCondition = operatorv1.OperatorCondition{
		Type:   "Available",
		Status: operatorv1.ConditionTrue,
		Reason: "Ready",
	}
)

func TestImageRegistryPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		Pull, imageregistryv1.Install, imageRegistryGVK).ExecuteTests(t)
}

func TestImageRegistryMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[imageregistryv1.Config, Builder](
		imageregistryv1.Install, imageRegistryGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestImageRegistryWithManagementState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		testManagementState operatorv1.ManagementState
		valid               bool
	}{
		{
			name:                "removed management state",
			testManagementState: operatorv1.Removed,
			valid:               true,
		},
		{
			name:                "managed management state",
			testManagementState: operatorv1.Managed,
			valid:               true,
		},
		{
			name:                "unmanaged management state",
			testManagementState: operatorv1.Unmanaged,
			valid:               true,
		},
		{
			name:                "invalid builder",
			testManagementState: operatorv1.Managed,
			valid:               false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject())
			if !testCase.valid {
				testBuilder = buildInvalidImageRegistryBuilder(buildImageRegistryClientWithDummyObject())
			}

			result := testBuilder.WithManagementState(testCase.testManagementState)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			assert.NoError(t, result.GetError())
			assert.Equal(t, testCase.testManagementState, result.Definition.Spec.ManagementState)
		})
	}
}

func TestImageRegistryGetManagementState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		resourceExists bool
		valid          bool
		expectedError  error
	}{
		{
			name:           "existing resource",
			resourceExists: true,
			valid:          true,
		},
		{
			name:           "non-existent resource",
			resourceExists: false,
			valid:          true,
			expectedError:  fmt.Errorf("imageRegistry object does not exist"),
		},
		{
			name:           "invalid builder",
			resourceExists: true,
			valid:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := buildImageRegistryClientWithDummyObject()
			if !testCase.resourceExists {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: []clients.SchemeAttacher{imageregistryv1.Install},
				})
			}

			testBuilder := buildValidImageRegistryBuilder(testSettings)
			if !testCase.valid {
				testBuilder = buildInvalidImageRegistryBuilder(testSettings)
			}

			currentManagementState, err := testBuilder.GetManagementState()

			if !testCase.valid {
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
				assert.Nil(t, currentManagementState)

				return
			}

			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, err)
				assert.Nil(t, currentManagementState)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, currentManagementState)
			assert.Equal(t, testBuilder.Object.Spec.ManagementState, *currentManagementState)
		})
	}
}

func TestImageRegistryWithStorage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		testStorageConfig imageregistryv1.ImageRegistryConfigStorage
		valid             bool
	}{
		{
			name: "empty directory storage",
			testStorageConfig: imageregistryv1.ImageRegistryConfigStorage{
				EmptyDir: &imageregistryv1.ImageRegistryConfigStorageEmptyDir{},
			},
			valid: true,
		},
		{
			name:              "empty storage",
			testStorageConfig: imageregistryv1.ImageRegistryConfigStorage{},
			valid:             true,
		},
		{
			name:              "invalid builder",
			testStorageConfig: imageregistryv1.ImageRegistryConfigStorage{},
			valid:             false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject())
			if !testCase.valid {
				testBuilder = buildInvalidImageRegistryBuilder(buildImageRegistryClientWithDummyObject())
			}

			result := testBuilder.WithStorage(testCase.testStorageConfig)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			assert.NoError(t, result.GetError())
			assert.Equal(t, testCase.testStorageConfig, result.Definition.Spec.Storage)
		})
	}
}

func TestImageRegistryGetStorageConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		resourceExists bool
		valid          bool
		expectedError  error
	}{
		{
			name:           "existing resource",
			resourceExists: true,
			valid:          true,
		},
		{
			name:           "non-existent resource",
			resourceExists: false,
			valid:          true,
			expectedError:  fmt.Errorf("imageRegistry object does not exist"),
		},
		{
			name:           "invalid builder",
			resourceExists: true,
			valid:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := buildImageRegistryClientWithDummyObject()
			if !testCase.resourceExists {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: []clients.SchemeAttacher{imageregistryv1.Install},
				})
			}

			testBuilder := buildValidImageRegistryBuilder(testSettings)
			if !testCase.valid {
				testBuilder = buildInvalidImageRegistryBuilder(testSettings)
			}

			currentStorageConfig, err := testBuilder.GetStorageConfig()

			if !testCase.valid {
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
				assert.Nil(t, currentStorageConfig)

				return
			}

			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, err)
				assert.Nil(t, currentStorageConfig)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, currentStorageConfig)
			assert.Equal(t, testBuilder.Object.Spec.Storage, *currentStorageConfig)
		})
	}
}

func TestImageRegistryWaitForCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		conditionMet  bool
		valid         bool
		expectedError error
	}{
		{
			name:         "condition met",
			exists:       true,
			conditionMet: true,
			valid:        true,
		},
		{
			name:         "resource does not exist",
			exists:       false,
			conditionMet: true,
			valid:        true,
			expectedError: fmt.Errorf("imageRegistry object %s does not exist",
				defaultImageRegistryName),
		},
		{
			name:          "condition not met",
			exists:        true,
			conditionMet:  false,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:         "invalid builder",
			exists:       true,
			conditionMet: true,
			valid:        false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				imageRegistry := buildDummyImageRegistry(defaultImageRegistryName, defaultManagementState)

				if testCase.conditionMet {
					imageRegistry.Status.Conditions = append(
						imageRegistry.Status.Conditions, defaultImageRegistryCondition)
				}

				runtimeObjects = append(runtimeObjects, imageRegistry)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{imageregistryv1.Install},
			})

			testBuilder := buildValidImageRegistryBuilder(testSettings)
			if !testCase.valid {
				testBuilder = buildInvalidImageRegistryBuilder(testSettings)
			}

			_, err := testBuilder.WaitForCondition(defaultImageRegistryCondition, time.Second)

			if !testCase.valid {
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))

				return
			}

			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func buildValidImageRegistryBuilder(apiClient *clients.Settings) *Builder {
	return buildImageRegistryBuilder(apiClient, defaultImageRegistryName)
}

func buildInvalidImageRegistryBuilder(apiClient *clients.Settings) *Builder {
	return buildImageRegistryBuilder(apiClient, "")
}

func buildImageRegistryBuilder(apiClient *clients.Settings, name string) *Builder {
	builder := common.NewClusterScopedBuilder[imageregistryv1.Config, Builder](
		apiClient, imageregistryv1.Install, name)
	builder.Definition.Spec.ManagementState = defaultManagementState

	return builder
}

func buildImageRegistryClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyImageRegistryObject(),
		SchemeAttachers: []clients.SchemeAttacher{imageregistryv1.Install},
	})
}

func buildDummyImageRegistry(name string, managementState operatorv1.ManagementState) *imageregistryv1.Config {
	return &imageregistryv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: imageregistryv1.ImageRegistrySpec{
			OperatorSpec: operatorv1.OperatorSpec{
				ManagementState: managementState,
			},
			Storage: imageregistryv1.ImageRegistryConfigStorage{},
		},
	}
}

func buildDummyImageRegistryObject() []runtime.Object {
	return append([]runtime.Object{}, buildDummyImageRegistry(defaultImageRegistryName, defaultManagementState))
}
