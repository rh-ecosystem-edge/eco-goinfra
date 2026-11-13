package assisted

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNewAgentServiceConfigBuilder(t *testing.T) {
	t.Parallel()

	databaseStorage, err := GetDefaultStorageSpec("50Gi")
	require.NoError(t, err)

	filesystemStorage, err := GetDefaultStorageSpec("40Gi")
	require.NoError(t, err)

	t.Run("valid builder", func(t *testing.T) {
		t.Parallel()

		builder := NewAgentServiceConfigBuilder(
			buildAgentServiceConfigTestClient(), databaseStorage, filesystemStorage)

		require.NotNil(t, builder)
		assert.NoError(t, builder.GetError())
		assert.Equal(t, agentServiceConfigName, builder.Definition.Name)
		assert.Equal(t, databaseStorage, builder.Definition.Spec.DatabaseStorage)
		assert.Equal(t, filesystemStorage, builder.Definition.Spec.FileSystemStorage)
		assert.Equal(t, agentServiceConfigGVK, builder.GetGVK())
	})

	t.Run("nil client returns error-bearing builder", func(t *testing.T) {
		t.Parallel()

		builder := NewAgentServiceConfigBuilder(nil, databaseStorage, filesystemStorage)

		require.NotNil(t, builder)
		assert.True(t, commonerrors.IsAPIClientNil(builder.GetError()))
	})
}

func TestNewDefaultAgentServiceConfigBuilder(t *testing.T) {
	t.Parallel()

	t.Run("valid builder", func(t *testing.T) {
		t.Parallel()

		builder := NewDefaultAgentServiceConfigBuilder(buildAgentServiceConfigTestClient())

		require.NotNil(t, builder)
		assert.NoError(t, builder.GetError())
		assert.Equal(t, agentServiceConfigName, builder.Definition.Name)
		assert.Equal(t, resource.MustParse(defaultDatabaseStorageSize),
			*builder.Definition.Spec.DatabaseStorage.Resources.Requests.Storage())
		assert.Equal(t, resource.MustParse(defaultFilesystemStorageSize),
			*builder.Definition.Spec.FileSystemStorage.Resources.Requests.Storage())
		assert.Equal(t, resource.MustParse(defaultImageStoreStorageSize),
			*builder.Definition.Spec.ImageStorage.Resources.Requests.Storage())
		assert.Equal(t, agentServiceConfigGVK, builder.GetGVK())
	})

	t.Run("nil client returns error-bearing builder", func(t *testing.T) {
		t.Parallel()

		builder := NewDefaultAgentServiceConfigBuilder(nil)

		require.NotNil(t, builder)
		assert.True(t, commonerrors.IsAPIClientNil(builder.GetError()))
	})
}

func TestPullAgentServiceConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		PullAgentServiceConfig,
		agentInstallV1Beta1.AddToScheme,
		agentServiceConfigGVK,
		agentServiceConfigName,
	).ExecuteTests(t)
}

func TestAgentServiceConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[
		agentInstallV1Beta1.AgentServiceConfig, AgentServiceConfigBuilder,
	](
		agentInstallV1Beta1.AddToScheme,
		agentServiceConfigGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestAgentServiceConfigWithImageStorage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		valid bool
		spec  corev1.PersistentVolumeClaimSpec
	}{
		{
			name:  "read write once",
			valid: true,
			spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{accessModeReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("30Gi"),
					},
				},
			},
		},
		{
			name:  "read write many",
			valid: true,
			spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteMany"},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("50Gi"),
					},
				},
			},
		},
		{
			name:  "invalid builder",
			valid: false,
			spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteMany"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidAgentServiceConfigTestBuilder()

			originalImageStorage := *builder.Definition.Spec.ImageStorage
			if !testCase.valid {
				builder = buildInvalidAgentServiceConfigTestBuilder()
			}

			builder.WithImageStorage(testCase.spec)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.Equal(t, originalImageStorage, *builder.Definition.Spec.ImageStorage)

				return
			}

			assert.NoError(t, builder.GetError())
			assert.Equal(t, testCase.spec, *builder.Definition.Spec.ImageStorage)
		})
	}
}

func TestAgentServiceConfigWithMirrorRegistryRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		configMapName string
		expectedError string
	}{
		{
			name:          "valid configmap ref",
			valid:         true,
			configMapName: "mirror-config",
		},
		{
			name:          "empty configmap name",
			valid:         true,
			configMapName: "",
			expectedError: "cannot add agentserviceconfig mirrorRegistryRef with empty configmap name",
		},
		{
			name:          "invalid builder",
			valid:         false,
			configMapName: "mirror-config",
			expectedError: "invalid builder",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidAgentServiceConfigTestBuilder()
			if !testCase.valid {
				builder = buildInvalidAgentServiceConfigTestBuilder()
			}

			builder.WithMirrorRegistryRef(testCase.configMapName)

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			assert.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.MirrorRegistryRef)
			assert.Equal(t, testCase.configMapName, builder.Definition.Spec.MirrorRegistryRef.Name)
		})
	}
}

func TestAgentServiceConfigWithOSImage(t *testing.T) {
	t.Parallel()

	firstImage := agentInstallV1Beta1.OSImage{
		OpenshiftVersion: "4.12",
		Version:          "412.86.202402272018-0",
		Url:              "https://mirror.openshift.com/rhcos-live-4.12.iso",
		CPUArchitecture:  "x86_64",
	}
	secondImage := agentInstallV1Beta1.OSImage{
		OpenshiftVersion: "4.13",
		Version:          "413.92.202307260246-0",
		Url:              "https://mirror.openshift.com/rhcos-live-4.13.iso",
		CPUArchitecture:  "x86_64",
	}

	testCases := []struct {
		name     string
		valid    bool
		osImages []agentInstallV1Beta1.OSImage
	}{
		{
			name:     "single image",
			valid:    true,
			osImages: []agentInstallV1Beta1.OSImage{firstImage},
		},
		{
			name:     "multiple images",
			valid:    true,
			osImages: []agentInstallV1Beta1.OSImage{firstImage, secondImage},
		},
		{
			name:     "invalid builder",
			valid:    false,
			osImages: []agentInstallV1Beta1.OSImage{firstImage},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidAgentServiceConfigTestBuilder()
			if !testCase.valid {
				builder = buildInvalidAgentServiceConfigTestBuilder()
			}

			for _, image := range testCase.osImages {
				builder.WithOSImage(image)
			}

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.Empty(t, builder.Definition.Spec.OSImages)

				return
			}

			assert.NoError(t, builder.GetError())
			assert.Equal(t, testCase.osImages, builder.Definition.Spec.OSImages)
		})
	}
}

func TestAgentServiceConfigWithUnauthenticatedRegistry(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		valid              bool
		registries         []string
		expectedRegistries []string
		expectedError      string
	}{
		{
			name:               "single registry",
			valid:              true,
			registries:         []string{"docker.io"},
			expectedRegistries: []string{"docker.io"},
		},
		{
			name:               "multiple registries",
			valid:              true,
			registries:         []string{"docker.io", "quay.io"},
			expectedRegistries: []string{"docker.io", "quay.io"},
		},
		{
			name:               "empty registry after valid registry",
			valid:              true,
			registries:         []string{"quay.io", ""},
			expectedRegistries: []string{"quay.io"},
			expectedError:      "agentserviceconfig cannot have empty unauthenticated registry",
		},
		{
			name:          "invalid builder",
			valid:         false,
			registries:    []string{"docker.io"},
			expectedError: "invalid builder",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidAgentServiceConfigTestBuilder()
			if !testCase.valid {
				builder = buildInvalidAgentServiceConfigTestBuilder()
			}

			for _, registry := range testCase.registries {
				builder.WithUnauthenticatedRegistry(registry)
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)
			} else {
				assert.NoError(t, builder.GetError())
			}

			assert.Equal(t, testCase.expectedRegistries, builder.Definition.Spec.UnauthenticatedRegistries)
		})
	}
}

func TestAgentServiceConfigWithIPXEHTTPRoute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		route         string
		expectedError string
	}{
		{
			name:  "enabled route",
			valid: true,
			route: ipxeEnabled,
		},
		{
			name:  "disabled route",
			valid: true,
			route: "disabled",
		},
		{
			name:  "invalid route",
			valid: true,
			route: "",
			expectedError: fmt.Sprintf(
				"agentserviceconfig passed invalid ipxeroute: , valid options: %v", validIPXEOptions),
		},
		{
			name:          "invalid builder",
			valid:         false,
			route:         ipxeEnabled,
			expectedError: "invalid builder",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidAgentServiceConfigTestBuilder()
			if !testCase.valid {
				builder = buildInvalidAgentServiceConfigTestBuilder()
			}

			builder.WithIPXEHTTPRoute(testCase.route)

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)
				assert.Empty(t, builder.Definition.Spec.IPXEHTTPRoute)

				return
			}

			assert.NoError(t, builder.GetError())
			assert.Equal(t, testCase.route, builder.Definition.Spec.IPXEHTTPRoute)
		})
	}
}

func TestAgentServiceConfigWithOptions(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[
		agentInstallV1Beta1.AgentServiceConfig, AgentServiceConfigBuilder,
	](
		agentInstallV1Beta1.AddToScheme,
		agentServiceConfigGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewWithOptionsTestConfig(commonTestConfig).ExecuteTests(t)

	t.Run("agent-specific option", func(t *testing.T) {
		t.Parallel()

		builder := buildValidAgentServiceConfigTestBuilder()
		builder.WithOptions(func(builder *AgentServiceConfigBuilder) (*AgentServiceConfigBuilder, error) {
			builder.Definition.Spec.MustGatherImages = []agentInstallV1Beta1.MustGatherImage{{
				Name:             "acm-must-gather",
				OpenshiftVersion: "4.16",
			}}

			return builder, nil
		})

		assert.NoError(t, builder.GetError())
		require.Len(t, builder.Definition.Spec.MustGatherImages, 1)
		assert.Equal(t, "acm-must-gather", builder.Definition.Spec.MustGatherImages[0].Name)
		assert.Equal(t, "4.16", builder.Definition.Spec.MustGatherImages[0].OpenshiftVersion)
	})
}

func TestAgentServiceConfigDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		valid         bool
		expectedError string
	}{
		{
			name:   "deleted resource",
			exists: true,
			valid:  true,
		},
		{
			name:   "resource already absent",
			exists: false,
			valid:  true,
		},
		{
			name:          "invalid builder",
			exists:        true,
			valid:         false,
			expectedError: "invalid builder",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object
			if testCase.exists {
				runtimeObjects = append(runtimeObjects, generateAgentServiceConfig())
			}

			builder := NewDefaultAgentServiceConfigBuilder(buildAgentServiceConfigTestClient(runtimeObjects...))
			if !testCase.valid {
				builder.SetError(errors.New("invalid builder"))
			}

			err := builder.DeleteAndWait(time.Second)

			if testCase.expectedError != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Nil(t, builder.Object)
		})
	}
}

func TestAgentServiceConfigWaitUntilDeployed(t *testing.T) {
	t.Parallel()

	healthyCondition := conditionsv1.Condition{
		Type:   agentInstallV1Beta1.ConditionDeploymentsHealthy,
		Status: corev1.ConditionTrue,
	}

	testCases := []struct {
		name          string
		exists        bool
		valid         bool
		conditions    []conditionsv1.Condition
		expectedError error
	}{
		{
			name:       "deployment healthy",
			exists:     true,
			valid:      true,
			conditions: []conditionsv1.Condition{healthyCondition},
		},
		{
			name:          "resource does not exist",
			exists:        false,
			valid:         true,
			expectedError: errors.New("cannot wait for non-existent agentserviceconfig to be deployed"),
		},
		{
			name:   "invalid builder",
			exists: true,
			valid:  false,
		},
		{
			name:          "timeout waiting for healthy condition",
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				agentServiceConfig := generateAgentServiceConfig()
				agentServiceConfig.Status.Conditions = testCase.conditions
				runtimeObjects = append(runtimeObjects, agentServiceConfig)
			}

			builder := NewDefaultAgentServiceConfigBuilder(buildAgentServiceConfigTestClient(runtimeObjects...))
			if !testCase.valid {
				builder.SetError(errors.New("invalid builder"))
			}

			_, err := builder.WaitUntilDeployed(time.Second)

			switch testCase.name {
			case "invalid builder":
				require.Error(t, err)
				assert.ErrorContains(t, err, "invalid builder")
			case "timeout waiting for healthy condition":
				assert.ErrorIs(t, err, testCase.expectedError)
			case "resource does not exist":
				assert.Equal(t, testCase.expectedError, err)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

func TestAgentServiceConfigGetDefaultStorageSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		size          string
		expectedError string
	}{
		{name: "database default", size: defaultDatabaseStorageSize},
		{name: "image store default", size: defaultImageStoreStorageSize},
		{name: "filesystem default", size: defaultFilesystemStorageSize},
		{name: "invalid unit", size: "30GiB", expectedError: "the storage size is in wrong format"},
		{name: "empty size", size: "", expectedError: "the storage size is in wrong format"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pvc, err := GetDefaultStorageSpec(testCase.size)

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
				assert.Empty(t, pvc)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, []corev1.PersistentVolumeAccessMode{accessModeReadWriteOnce}, pvc.AccessModes)
			assert.Equal(t, resource.MustParse(testCase.size), *pvc.Resources.Requests.Storage())
		})
	}
}

// generateAgentServiceConfig returns a default AgentServiceConfig test object.
func generateAgentServiceConfig() *agentInstallV1Beta1.AgentServiceConfig {
	return &agentInstallV1Beta1.AgentServiceConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: agentServiceConfigName,
		},
		Spec: agentInstallV1Beta1.AgentServiceConfigSpec{
			DatabaseStorage: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{accessModeReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(defaultDatabaseStorageSize),
					},
				},
			},
			FileSystemStorage: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{accessModeReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(defaultFilesystemStorageSize),
					},
				},
			},
		},
	}
}

// buildAgentServiceConfigTestClient returns a test client with the AgentServiceConfig scheme and optional objects.
func buildAgentServiceConfigTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: objects,
		SchemeAttachers: []clients.SchemeAttacher{
			agentInstallV1Beta1.AddToScheme,
		},
	})
}

// buildValidAgentServiceConfigTestBuilder returns a valid AgentServiceConfig builder for testing.
func buildValidAgentServiceConfigTestBuilder() *AgentServiceConfigBuilder {
	return NewDefaultAgentServiceConfigBuilder(buildAgentServiceConfigTestClient())
}

// buildInvalidAgentServiceConfigTestBuilder returns an error-bearing AgentServiceConfig builder for testing.
func buildInvalidAgentServiceConfigTestBuilder() *AgentServiceConfigBuilder {
	builder := buildValidAgentServiceConfigTestBuilder()
	builder.SetError(errors.New("invalid builder"))

	return builder
}
