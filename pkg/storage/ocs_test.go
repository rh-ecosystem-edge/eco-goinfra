package storage

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	ocsoperatorv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ocs/operatorv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultStorageClusterName      = "ocs-storagecluster"
	defaultStorageClusterNamespace = "openshift-storage"
	defaultStorageClassName        = "ocs-storagecluster-cephfs"
	defaultVolumeMode              = corev1.PersistentVolumeBlock
	ocsTestSchemes                 = []clients.SchemeAttacher{
		ocsoperatorv1.AddToScheme,
	}
)

func TestNewStorageClusterBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewStorageClusterBuilder,
		ocsoperatorv1.AddToScheme,
		storageClusterGVK,
	).ExecuteTests(t)
}

func TestPullStorageCluster(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullStorageCluster,
		ocsoperatorv1.AddToScheme,
		storageClusterGVK,
	).ExecuteTests(t)
}

func TestStorageClusterBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[ocsoperatorv1.StorageCluster, StorageClusterBuilder](
		ocsoperatorv1.AddToScheme,
		storageClusterGVK,
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

func TestStorageClusterGetManagedResources(t *testing.T) {
	t.Parallel()

	expectedManagedResources := ocsoperatorv1.ManagedResourcesSpec{
		CephBlockPools: ocsoperatorv1.ManageCephBlockPools{ReconcileStrategy: "manage"},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		dummyStorageCluster := buildDummyStorageCluster()
		dummyStorageCluster.Spec.ManagedResources = expectedManagedResources
		builder := buildValidStorageClusterBuilder(buildTestClientWithStorageCluster(dummyStorageCluster))

		managedResources, err := builder.GetManagedResources()

		require.NoError(t, err)
		assert.Equal(t, expectedManagedResources, *managedResources)
	})

	t.Run("resource does not exist", func(t *testing.T) {
		t.Parallel()

		builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())

		_, err := builder.GetManagedResources()

		require.EqualError(t, err, fmt.Sprintf(
			"storageCluster object %s does not exist in namespace %s",
			defaultStorageClusterName, defaultStorageClusterNamespace))
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())

		_, err := builder.GetManagedResources()

		require.Error(t, err)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
	})
}

func TestStorageClusterGetMonDataDirHostPath(t *testing.T) {
	t.Parallel()

	const expectedMonDataDirHostPath = "/var/lib/rook"

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		dummyStorageCluster := buildDummyStorageCluster()
		dummyStorageCluster.Spec.MonDataDirHostPath = expectedMonDataDirHostPath
		builder := buildValidStorageClusterBuilder(buildTestClientWithStorageCluster(dummyStorageCluster))

		monDataDirHostPath, err := builder.GetMonDataDirHostPath()

		require.NoError(t, err)
		assert.Equal(t, expectedMonDataDirHostPath, monDataDirHostPath)
	})

	t.Run("resource does not exist", func(t *testing.T) {
		t.Parallel()

		builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())

		_, err := builder.GetMonDataDirHostPath()

		require.EqualError(t, err, fmt.Sprintf(
			"storageCluster object %s does not exist in namespace %s",
			defaultStorageClusterName, defaultStorageClusterNamespace))
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())

		_, err := builder.GetMonDataDirHostPath()

		require.Error(t, err)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
	})
}

func TestStorageClusterGetMultiCloudGateway(t *testing.T) {
	t.Parallel()

	expectedMultiCloudGateway := &ocsoperatorv1.MultiCloudGatewaySpec{ReconcileStrategy: "manage"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		dummyStorageCluster := buildDummyStorageCluster()
		dummyStorageCluster.Spec.MultiCloudGateway = expectedMultiCloudGateway
		builder := buildValidStorageClusterBuilder(buildTestClientWithStorageCluster(dummyStorageCluster))

		multiCloudGateway, err := builder.GetMultiCloudGateway()

		require.NoError(t, err)
		assert.Equal(t, expectedMultiCloudGateway, multiCloudGateway)
	})

	t.Run("resource does not exist", func(t *testing.T) {
		t.Parallel()

		builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())

		_, err := builder.GetMultiCloudGateway()

		require.EqualError(t, err, fmt.Sprintf(
			"storageCluster object %s does not exist in namespace %s",
			defaultStorageClusterName, defaultStorageClusterNamespace))
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())

		_, err := builder.GetMultiCloudGateway()

		require.Error(t, err)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
	})
}

func TestStorageClusterGetStorageDeviceSets(t *testing.T) {
	t.Parallel()

	expectedStorageDeviceSets := []ocsoperatorv1.StorageDeviceSet{{Name: "local-block", Count: 3}}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		dummyStorageCluster := buildDummyStorageCluster()
		dummyStorageCluster.Spec.StorageDeviceSets = expectedStorageDeviceSets
		builder := buildValidStorageClusterBuilder(buildTestClientWithStorageCluster(dummyStorageCluster))

		storageDeviceSets, err := builder.GetStorageDeviceSets()

		require.NoError(t, err)
		assert.Equal(t, expectedStorageDeviceSets, storageDeviceSets)
	})

	t.Run("resource does not exist", func(t *testing.T) {
		t.Parallel()

		builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())

		_, err := builder.GetStorageDeviceSets()

		require.EqualError(t, err, fmt.Sprintf(
			"storageCluster object %s does not exist in namespace %s",
			defaultStorageClusterName, defaultStorageClusterNamespace))
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())

		_, err := builder.GetStorageDeviceSets()

		require.Error(t, err)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
	})
}

func TestStorageClusterWithFlexibleScaling(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		flexibleScaling bool
		valid           bool
	}{
		{name: "flexible scaling enabled", flexibleScaling: true, valid: true},
		{name: "flexible scaling disabled", flexibleScaling: false, valid: true},
		{name: "invalid builder", flexibleScaling: true, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			if !testCase.valid {
				builder = buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			}

			builder = builder.WithFlexibleScaling(testCase.flexibleScaling)

			if testCase.valid {
				require.NoError(t, builder.GetError())
				assert.Equal(t, testCase.flexibleScaling, builder.Definition.Spec.FlexibleScaling)
			} else {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			}
		})
	}
}

func TestStorageClusterWithManagedResources(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		managedResources    ocsoperatorv1.ManagedResourcesSpec
		expectedManagedSpec ocsoperatorv1.ManagedResourcesSpec
		valid               bool
	}{
		{
			name: "managed resources configured",
			managedResources: ocsoperatorv1.ManagedResourcesSpec{
				CephBlockPools: ocsoperatorv1.ManageCephBlockPools{ReconcileStrategy: "manage"},
			},
			expectedManagedSpec: ocsoperatorv1.ManagedResourcesSpec{
				CephBlockPools: ocsoperatorv1.ManageCephBlockPools{ReconcileStrategy: "manage"},
			},
			valid: true,
		},
		{
			name:                "empty managed resources",
			managedResources:    ocsoperatorv1.ManagedResourcesSpec{},
			expectedManagedSpec: ocsoperatorv1.ManagedResourcesSpec{},
			valid:               true,
		},
		{name: "invalid builder", valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			if !testCase.valid {
				builder = buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			}

			builder = builder.WithManagedResources(testCase.managedResources)

			if testCase.valid {
				require.NoError(t, builder.GetError())
				assert.Equal(t, testCase.expectedManagedSpec, builder.Definition.Spec.ManagedResources)
			} else {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			}
		})
	}
}

func TestStorageClusterWithMonDataDirHostPath(t *testing.T) {
	t.Parallel()

	const expectedMonDataDirHostPath = "/var/lib/rook"

	testCases := []struct {
		name                   string
		monDataDirHostPath     string
		expectedMonDataDirPath string
		expectedError          string
		valid                  bool
	}{
		{
			name:                   "valid mon data directory",
			monDataDirHostPath:     expectedMonDataDirHostPath,
			expectedMonDataDirPath: expectedMonDataDirHostPath,
			valid:                  true,
		},
		{
			name:               "empty mon data directory",
			monDataDirHostPath: "",
			expectedError:      "the expectedMonDataDirHostPath can not be empty",
			valid:              true,
		},
		{name: "invalid builder", monDataDirHostPath: expectedMonDataDirHostPath, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			if !testCase.valid {
				builder = buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			}

			builder = builder.WithMonDataDirHostPath(testCase.monDataDirHostPath)

			switch {
			case testCase.expectedError != "":
				require.EqualError(t, builder.GetError(), testCase.expectedError)
			case testCase.valid:
				require.NoError(t, builder.GetError())
				assert.Equal(t, testCase.expectedMonDataDirPath, builder.Definition.Spec.MonDataDirHostPath)
			default:
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			}
		})
	}
}

func TestStorageClusterWithMultiCloudGateway(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		multiCloudGateway ocsoperatorv1.MultiCloudGatewaySpec
		expectedGateway   *ocsoperatorv1.MultiCloudGatewaySpec
		valid             bool
	}{
		{
			name:              "multi cloud gateway configured",
			multiCloudGateway: ocsoperatorv1.MultiCloudGatewaySpec{ReconcileStrategy: "manage"},
			expectedGateway:   &ocsoperatorv1.MultiCloudGatewaySpec{ReconcileStrategy: "manage"},
			valid:             true,
		},
		{name: "invalid builder", valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			if !testCase.valid {
				builder = buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			}

			builder = builder.WithMultiCloudGateway(testCase.multiCloudGateway)

			if testCase.valid {
				require.NoError(t, builder.GetError())
				assert.Equal(t, testCase.expectedGateway, builder.Definition.Spec.MultiCloudGateway)
			} else {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			}
		})
	}
}

func TestStorageClusterWithStorageDeviceSet(t *testing.T) {
	t.Parallel()

	resourceList := map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceStorage: resource.MustParse("1"),
	}
	storageDeviceSet := ocsoperatorv1.StorageDeviceSet{
		Count:    3,
		Replica:  1,
		Portable: false,
		Name:     "local-block",
		DataPVCTemplate: corev1.PersistentVolumeClaim{
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: resourceList,
				},
				StorageClassName: &defaultStorageClassName,
				VolumeMode:       &defaultVolumeMode,
			},
		},
	}

	testCases := []struct {
		name             string
		storageDeviceSet ocsoperatorv1.StorageDeviceSet
		valid            bool
	}{
		{name: "storage device set configured", storageDeviceSet: storageDeviceSet, valid: true},
		{name: "empty storage device set", storageDeviceSet: ocsoperatorv1.StorageDeviceSet{}, valid: true},
		{name: "invalid builder", storageDeviceSet: storageDeviceSet, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			if !testCase.valid {
				builder = buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			}

			builder = builder.WithStorageDeviceSet(testCase.storageDeviceSet)

			if testCase.valid {
				require.NoError(t, builder.GetError())
				assert.Equal(t, []ocsoperatorv1.StorageDeviceSet{testCase.storageDeviceSet},
					builder.Definition.Spec.StorageDeviceSets)
			} else {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			}
		})
	}
}

func TestStorageClusterWithAnnotations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		annotations         map[string]string
		expectedAnnotations map[string]string
		expectedError       string
		valid               bool
	}{
		{
			name:                "annotations configured",
			annotations:         map[string]string{"uninstall.ocs.openshift.io/cleanup-policy": "delete"},
			expectedAnnotations: map[string]string{"uninstall.ocs.openshift.io/cleanup-policy": "delete"},
			valid:               true,
		},
		{
			name:          "empty annotations",
			annotations:   map[string]string{},
			expectedError: "'annotations' argument cannot be empty",
			valid:         true,
		},
		{
			name:        "invalid builder",
			annotations: map[string]string{"uninstall.ocs.openshift.io/cleanup-policy": "delete"},
			valid:       false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			if !testCase.valid {
				builder = buildInvalidStorageClusterBuilder(buildTestClientWithStorageClusterScheme())
			}

			builder = builder.WithAnnotations(testCase.annotations)

			switch {
			case testCase.expectedError != "":
				require.EqualError(t, builder.GetError(), testCase.expectedError)
			case testCase.valid:
				require.NoError(t, builder.GetError())
				assert.Equal(t, testCase.expectedAnnotations, builder.Definition.Annotations)
			default:
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			}
		})
	}
}

// buildValidStorageClusterBuilder returns a valid StorageClusterBuilder for testing.
func buildValidStorageClusterBuilder(apiClient *clients.Settings) *StorageClusterBuilder {
	return NewStorageClusterBuilder(apiClient, defaultStorageClusterName, defaultStorageClusterNamespace)
}

// buildInvalidStorageClusterBuilder returns an invalid StorageClusterBuilder for testing.
func buildInvalidStorageClusterBuilder(apiClient *clients.Settings) *StorageClusterBuilder {
	return NewStorageClusterBuilder(apiClient, "", defaultStorageClusterNamespace)
}

// buildTestClientWithStorageClusterScheme returns a client with the StorageCluster scheme.
func buildTestClientWithStorageClusterScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: ocsTestSchemes,
	})
}

// buildTestClientWithStorageCluster returns a client containing the provided StorageCluster.
func buildTestClientWithStorageCluster(storageCluster *ocsoperatorv1.StorageCluster) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{storageCluster},
		SchemeAttachers: ocsTestSchemes,
	})
}

// buildDummyStorageCluster returns a StorageCluster with the default test name and namespace.
func buildDummyStorageCluster() *ocsoperatorv1.StorageCluster {
	return &ocsoperatorv1.StorageCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultStorageClusterName,
			Namespace: defaultStorageClusterNamespace,
		},
	}
}
