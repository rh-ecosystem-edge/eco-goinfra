package storage

import (
	"context"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testStorageClassBuilderName    = "storageclass-test"
	defaultStorageClassProvisioner = "test-provisioner"
	defaultParameterKey            = "readOnly"
	defaultParameterValue          = "false"
)

func TestNewClassBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		func(apiClient *clients.Settings, name string) *ClassBuilder {
			return NewClassBuilder(apiClient, name, defaultStorageClassProvisioner)
		},
		storageV1.AddToScheme,
		storageClassGVK,
	).ExecuteTests(t)

	testCases := []struct {
		name          string
		provisioner   string
		expectedError string
	}{
		{
			name:        "valid provisioner",
			provisioner: defaultStorageClassProvisioner,
		},
		{
			name:          "empty provisioner",
			provisioner:   "",
			expectedError: errEmptyProvisioner,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewClassBuilder(
				buildTestClientWithoutDummyStorageClass(),
				testStorageClassBuilderName,
				testCase.provisioner,
			)

			if testCase.expectedError == "" {
				require.NoError(t, builder.GetError())
			} else {
				require.EqualError(t, builder.GetError(), testCase.expectedError)
			}

			assert.Equal(t, testCase.provisioner, builder.Definition.Provisioner)
		})
	}
}

func TestPullClass(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullClass,
		storageV1.AddToScheme,
		storageClassGVK,
	).ExecuteTests(t)

	t.Run("pulls storageclass provisioner", func(t *testing.T) {
		t.Parallel()

		builder, err := PullClass(buildTestClientWithDummyStorageClass(), testStorageClassBuilderName)
		require.NoError(t, err)
		assert.Equal(t, defaultStorageClassProvisioner, builder.Definition.Provisioner)
	})
}

func TestClassBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newStorageClassCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestClassWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newStorageClassCommonTestConfig()).ExecuteTests(t)
}

func TestClassWithReclaimPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		reclaimPolicy corev1.PersistentVolumeReclaimPolicy
		validBuilder  bool
		expectedError string
	}{
		{
			name:          "valid reclaim policy",
			reclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			validBuilder:  true,
		},
		{
			name:          "empty reclaim policy",
			reclaimPolicy: "",
			validBuilder:  true,
			expectedError: "storageclass 'reclaimPolicy' cannot be empty",
		},
		{
			name:          "invalid builder",
			reclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			validBuilder:  false,
			expectedError: errEmptyProvisioner,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			classBuilder := buildStorageClassTestBuilder(testCase.validBuilder)
			classBuilder.WithReclaimPolicy(testCase.reclaimPolicy)

			if testCase.expectedError == "" {
				require.NoError(t, classBuilder.GetError())
				require.NotNil(t, classBuilder.Definition.ReclaimPolicy)
				assert.Equal(t, testCase.reclaimPolicy, *classBuilder.Definition.ReclaimPolicy)
			} else {
				require.EqualError(t, classBuilder.GetError(), testCase.expectedError)
			}
		})
	}
}

func TestClassWithVolumeBindingMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		volumeBindingMode storageV1.VolumeBindingMode
		validBuilder      bool
		expectedError     string
	}{
		{
			name:              "valid volume binding mode",
			volumeBindingMode: storageV1.VolumeBindingImmediate,
			validBuilder:      true,
		},
		{
			name:              "empty volume binding mode",
			volumeBindingMode: "",
			validBuilder:      true,
			expectedError:     "storageclass 'bindingMode' cannot be empty",
		},
		{
			name:              "invalid builder",
			volumeBindingMode: storageV1.VolumeBindingWaitForFirstConsumer,
			validBuilder:      false,
			expectedError:     errEmptyProvisioner,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			classBuilder := buildStorageClassTestBuilder(testCase.validBuilder)
			classBuilder.WithVolumeBindingMode(testCase.volumeBindingMode)

			if testCase.expectedError == "" {
				require.NoError(t, classBuilder.GetError())
				require.NotNil(t, classBuilder.Definition.VolumeBindingMode)
				assert.Equal(t, testCase.volumeBindingMode, *classBuilder.Definition.VolumeBindingMode)
			} else {
				require.EqualError(t, classBuilder.GetError(), testCase.expectedError)
			}
		})
	}
}

func TestClassWithParameter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		parameterKey   string
		parameterValue string
		validBuilder   bool
		expectedError  string
	}{
		{
			name:           "valid parameter",
			parameterKey:   defaultParameterKey,
			parameterValue: defaultParameterValue,
			validBuilder:   true,
		},
		{
			name:           "empty parameter key",
			parameterValue: defaultParameterValue,
			validBuilder:   true,
			expectedError:  "storageclass parameter key cannot be empty",
		},
		{
			name:          "empty parameter value",
			parameterKey:  defaultParameterKey,
			validBuilder:  true,
			expectedError: "storageclass parameter value cannot be empty",
		},
		{
			name:           "invalid builder",
			parameterKey:   defaultParameterKey,
			parameterValue: defaultParameterValue,
			validBuilder:   false,
			expectedError:  errEmptyProvisioner,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			classBuilder := buildStorageClassTestBuilder(testCase.validBuilder)
			classBuilder.WithParameter(testCase.parameterKey, testCase.parameterValue)

			if testCase.expectedError == "" {
				require.NoError(t, classBuilder.GetError())
				assert.Equal(t, testCase.parameterValue, classBuilder.Definition.Parameters[testCase.parameterKey])
			} else {
				require.EqualError(t, classBuilder.GetError(), testCase.expectedError)
			}
		})
	}
}

func TestClassDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		storageExists bool
		validBuilder  bool
		expectedError string
	}{
		{
			name:          "deletes existing storageclass",
			storageExists: true,
			validBuilder:  true,
		},
		{
			name:          "storageclass already absent",
			storageExists: false,
			validBuilder:  true,
		},
		{
			name:          "invalid builder",
			storageExists: true,
			validBuilder:  false,
			expectedError: errEmptyProvisioner,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var apiClient *clients.Settings
			if testCase.storageExists {
				apiClient = buildTestClientWithDummyStorageClass()
			} else {
				apiClient = buildTestClientWithoutDummyStorageClass()
			}

			classBuilder := buildValidClassTestBuilder(apiClient)
			if !testCase.validBuilder {
				classBuilder = buildInvalidClassTestBuilder(apiClient)
			}

			err := classBuilder.DeleteAndWait(time.Second)
			if testCase.expectedError == "" {
				require.NoError(t, err)
				assert.Nil(t, classBuilder.Object)
			} else {
				require.ErrorContains(t, err, testCase.expectedError)
			}
		})
	}
}

func TestClassWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		storageExists bool
		validBuilder  bool
		expectedError error
	}{
		{
			name:          "times out while storageclass exists",
			storageExists: true,
			validBuilder:  true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "storageclass already absent",
			storageExists: false,
			validBuilder:  true,
		},
		{
			name:          "invalid builder",
			storageExists: true,
			validBuilder:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var apiClient *clients.Settings
			if testCase.storageExists {
				apiClient = buildTestClientWithDummyStorageClass()
			} else {
				apiClient = buildTestClientWithoutDummyStorageClass()
			}

			classBuilder := buildValidClassTestBuilder(apiClient)
			if !testCase.validBuilder {
				classBuilder = buildInvalidClassTestBuilder(apiClient)
			}

			err := classBuilder.WaitUntilDeleted(time.Second)

			switch {
			case testCase.expectedError != nil:
				require.ErrorIs(t, err, testCase.expectedError)
			case testCase.validBuilder:
				require.NoError(t, err)
			default:
				require.ErrorContains(t, err, errEmptyProvisioner)
			}
		})
	}
}

// newStorageClassCommonTestConfig returns the shared test configuration for StorageClass.
func newStorageClassCommonTestConfig() testhelper.CommonTestConfig[
	storageV1.StorageClass, ClassBuilder, *storageV1.StorageClass, *ClassBuilder,
] {
	return testhelper.NewCommonTestConfig[storageV1.StorageClass, ClassBuilder](
		storageV1.AddToScheme,
		storageClassGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

// buildStorageClassTestBuilder returns a valid or invalid ClassBuilder for modifier tests.
func buildStorageClassTestBuilder(valid bool) *ClassBuilder {
	apiClient := buildTestClientWithoutDummyStorageClass()
	if valid {
		return buildValidClassTestBuilder(apiClient)
	}

	return buildInvalidClassTestBuilder(apiClient)
}

// buildDummyStorageClass returns a StorageClass with the provided name and provisioner.
func buildDummyStorageClass(name, provisioner string) *storageV1.StorageClass {
	return &storageV1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Provisioner: provisioner,
	}
}

// buildTestClientWithDummyStorageClass returns a client with a mock StorageClass.
func buildTestClientWithDummyStorageClass() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyStorageClass(testStorageClassBuilderName, defaultStorageClassProvisioner),
		},
		SchemeAttachers: []clients.SchemeAttacher{storageV1.AddToScheme},
	})
}

// buildTestClientWithoutDummyStorageClass returns a client with the StorageClass scheme and no objects.
func buildTestClientWithoutDummyStorageClass() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{storageV1.AddToScheme},
	})
}

// buildValidClassTestBuilder returns a valid ClassBuilder for testing.
func buildValidClassTestBuilder(apiClient *clients.Settings) *ClassBuilder {
	return NewClassBuilder(apiClient, testStorageClassBuilderName, defaultStorageClassProvisioner)
}

// buildInvalidClassTestBuilder returns an invalid ClassBuilder for testing.
func buildInvalidClassTestBuilder(apiClient *clients.Settings) *ClassBuilder {
	return NewClassBuilder(apiClient, testStorageClassBuilderName, "")
}
