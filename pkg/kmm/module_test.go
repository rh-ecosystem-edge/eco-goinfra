package kmm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemesV1beta1 = []clients.SchemeAttacher{
		v1beta1.AddToScheme,
	}
	defaultModuleName      = "testmodule"
	defaultModuleNamespace = "testns"
)

func TestNewModuleBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewModuleBuilder, v1beta1.AddToScheme, moduleGVK).ExecuteTests(t)
}

func TestModulePull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(Pull, v1beta1.AddToScheme, moduleGVK).ExecuteTests(t)
}

func TestModuleBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[v1beta1.Module, ModuleBuilder](
		v1beta1.AddToScheme,
		moduleGVK,
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

func TestModuleWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		nodeSelector map[string]string
		expectedErr  string
	}{
		{
			nodeSelector: map[string]string{"test": "test"},
			expectedErr:  "",
		},
		{
			nodeSelector: map[string]string{},
			expectedErr:  "Module 'nodeSelector' cannot be empty map",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithNodeSelector(testCase.nodeSelector)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.Selector, testCase.nodeSelector)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithLoadServiceAccount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		loadServiceAccount string
		expectedErr        string
	}{
		{
			loadServiceAccount: "test",
			expectedErr:        "",
		},
		{
			loadServiceAccount: "",
			expectedErr:        canNotRedefineModuleWithEmpty,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithLoadServiceAccount(testCase.loadServiceAccount)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.ModuleLoader.ServiceAccountName, testCase.loadServiceAccount)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDevicePluginServiceAccount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		devicePluginServiceAccount string
		expectedErr                string
	}{
		{
			devicePluginServiceAccount: "test",
			expectedErr:                "",
		},
		{
			devicePluginServiceAccount: "",
			expectedErr:                canNotRedefineModuleWithEmpty,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDevicePluginServiceAccount(testCase.devicePluginServiceAccount)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.DevicePlugin.ServiceAccountName, testCase.devicePluginServiceAccount)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithImageRepoSecret(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		repoSecret  string
		expectedErr string
	}{
		{
			repoSecret:  "test",
			expectedErr: "",
		},
		{
			repoSecret:  "",
			expectedErr: "can not redefine module with empty imageRepoSecret",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithImageRepoSecret(testCase.repoSecret)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.ImageRepoSecret.Name, testCase.repoSecret)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDevicePluginAutomountServiceAccountToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		automount     bool
		expectedValue bool
	}{
		{
			automount:     false,
			expectedValue: false,
		},
		{
			automount:     true,
			expectedValue: true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDevicePluginAutomountServiceAccountToken(testCase.automount)

		assert.NotNil(t, testBuilder.Definition.Spec.DevicePlugin)
		assert.NotNil(t, testBuilder.Definition.Spec.DevicePlugin.AutomountServiceAccountToken)
		assert.Equal(t, testCase.expectedValue,
			*testBuilder.Definition.Spec.DevicePlugin.AutomountServiceAccountToken)
	}
}

func TestModuleWithDevicePluginVolume(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		pluginVolumeName string
		configMapName    string
		expectedErr      string
	}{
		{
			pluginVolumeName: "test",
			configMapName:    "test",
			expectedErr:      "",
		},
		{
			pluginVolumeName: "",
			configMapName:    "test",
			expectedErr:      "cannot redefine with empty volume 'name'",
		},
		{
			pluginVolumeName: "test",
			configMapName:    "",
			expectedErr:      "cannot redefine with empty 'configMapName'",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDevicePluginVolume(testCase.pluginVolumeName, testCase.configMapName)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.DevicePlugin.Volumes[0].Name, testCase.pluginVolumeName)
			assert.Equal(t, testBuilder.Definition.Spec.DevicePlugin.Volumes[0].ConfigMap.Name, testCase.configMapName)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithModuleLoaderContainer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		containerSpec *v1beta1.ModuleLoaderContainerSpec
		expectedErr   string
	}{
		{
			containerSpec: &v1beta1.ModuleLoaderContainerSpec{ContainerImage: "test"},
			expectedErr:   "",
		},
		{
			containerSpec: nil,
			expectedErr:   errNilContainer,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithModuleLoaderContainer(testCase.containerSpec)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.ModuleLoader.Container, *testCase.containerSpec)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDevicePluginContainer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		containerSpec *v1beta1.DevicePluginContainerSpec
		expectedErr   string
	}{
		{
			containerSpec: &v1beta1.DevicePluginContainerSpec{Image: "test"},
			expectedErr:   "",
		},
		{
			containerSpec: nil,
			expectedErr:   errNilContainer,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDevicePluginContainer(testCase.containerSpec)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.DevicePlugin.Container, *testCase.containerSpec)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithOptions(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[v1beta1.Module, ModuleBuilder](
		v1beta1.AddToScheme,
		moduleGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewWithOptionsTestConfig(commonTestConfig).ExecuteTests(t)
}

func TestModuleWithToleration(t *testing.T) {
	t.Parallel()

	var tenSeconds = int64(10)

	testCases := []struct {
		key         string
		operator    string
		value       string
		effect      string
		seconds     *int64
		expectedErr string
	}{
		{
			key:         "",
			operator:    "test",
			value:       "test",
			effect:      "test",
			seconds:     &tenSeconds,
			expectedErr: "cannot redefine with empty 'key' value",
		},
		{
			key:         "test",
			operator:    "",
			value:       "test",
			effect:      "test",
			seconds:     &tenSeconds,
			expectedErr: "cannot redefine with empty 'operator' value",
		},
		{
			key:         "test",
			operator:    "test",
			value:       "test",
			effect:      "",
			seconds:     &tenSeconds,
			expectedErr: "cannot redefine with empty 'effect' value",
		},
		{
			key:         "testkey",
			operator:    "Equals",
			value:       "testvalue",
			effect:      "NoSchedule",
			expectedErr: "",
		},
		{
			key:         "testkey",
			operator:    "Equals",
			value:       "testvalue",
			effect:      "NoSchedule",
			expectedErr: "",
		},
	}
	for _, testCase := range testCases {
		testSettings := buildModuleTestClientWithDummyObject()
		testBuilder := buildValidTestModule(testSettings).WithToleration(
			testCase.key, testCase.operator, testCase.value, testCase.effect, testCase.seconds)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.key, testBuilder.Definition.Spec.Tolerations[0].Key)
			assert.Equal(t, corev1.TaintEffect(testCase.effect), testBuilder.Definition.Spec.Tolerations[0].Effect)
			assert.Equal(t, corev1.TolerationOperator(testCase.operator), testBuilder.Definition.Spec.Tolerations[0].Operator)
			assert.Equal(t, testCase.value, testBuilder.Definition.Spec.Tolerations[0].Value)
			assert.Equal(t, testCase.seconds, testBuilder.Definition.Spec.Tolerations[0].TolerationSeconds)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithImageRebuildTriggerGeneration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		generation  int
		expectedErr string
	}{
		{
			generation:  1,
			expectedErr: "",
		},
		{
			generation:  0,
			expectedErr: "",
		},
		{
			generation:  42,
			expectedErr: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithImageRebuildTriggerGeneration(testCase.generation)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder.Definition.Spec.ImageRebuildTriggerGeneration)
			assert.Equal(t, testCase.generation, *testBuilder.Definition.Spec.ImageRebuildTriggerGeneration)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleBuildModuleSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		testModule *ModuleBuilder
		valid      bool
	}{
		{
			name:       "valid builder",
			testModule: buildValidTestModule(buildModuleTestClientWithDummyObject()),
			valid:      true,
		},
		{
			name:       "invalid builder",
			testModule: buildInValidTestModule(buildModuleTestClientWithDummyObject()),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			moduleSpec, err := testCase.testModule.BuildModuleSpec()
			if testCase.valid {
				assert.NoError(t, err)
				assert.NotEmpty(t, moduleSpec)
			} else {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
			}
		})
	}
}

func buildValidTestModule(apiClient *clients.Settings) *ModuleBuilder {
	moduleBuilder := NewModuleBuilder(apiClient, defaultModuleName, defaultModuleNamespace)
	moduleBuilder.Definition.Spec.ModuleLoader = &v1beta1.ModuleLoaderSpec{ServiceAccountName: "Test"}

	return moduleBuilder
}

func buildInValidTestModule(apiClient *clients.Settings) *ModuleBuilder {
	moduleBuilder := NewModuleBuilder(apiClient, defaultModuleName, "")

	return moduleBuilder
}

func buildModuleTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyModule(),
		SchemeAttachers: testSchemesV1beta1,
	})
}

func buildDummyModule() []runtime.Object {
	return append([]runtime.Object{}, &v1beta1.Module{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultModuleName,
			Namespace: defaultModuleNamespace,
		},

		Spec: v1beta1.ModuleSpec{},
	})
}

// assertModuleBuilderError checks the package-specific error message stored on a Module builder.
func assertModuleBuilderError(t *testing.T, expectedError string, builder *ModuleBuilder) {
	t.Helper()

	if expectedError == "" {
		assert.NoError(t, builder.GetError())

		return
	}

	assert.EqualError(t, builder.GetError(), expectedError)
}

func TestModuleWithDRA(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		dra         *v1beta1.DRASpec
		expectedErr string
	}{
		{
			dra:         &v1beta1.DRASpec{DriverName: "test.example.com"},
			expectedErr: "",
		},
		{
			dra:         nil,
			expectedErr: "invalid 'dra' argument cannot be nil",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDRA(testCase.dra)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder.Definition.Spec.DRA)
			assert.Equal(t, testCase.dra.DriverName, testBuilder.Definition.Spec.DRA.DriverName)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDRAContainer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		container   *v1beta1.CommonContainerSpec
		expectedErr string
	}{
		{
			container:   &v1beta1.CommonContainerSpec{Image: "test-image:latest"},
			expectedErr: "",
		},
		{
			container:   nil,
			expectedErr: errNilContainer,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDRAContainer(testCase.container)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder.Definition.Spec.DRA)
			assert.Equal(t, testCase.container.Image, testBuilder.Definition.Spec.DRA.Container.Image)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDRAInitContainer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		container   *v1beta1.CommonContainerSpec
		expectedErr string
	}{
		{
			container:   &v1beta1.CommonContainerSpec{Image: "init-image:latest"},
			expectedErr: "",
		},
		{
			container:   nil,
			expectedErr: errNilContainer,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDRAInitContainer(testCase.container)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder.Definition.Spec.DRA)
			assert.NotNil(t, testBuilder.Definition.Spec.DRA.InitContainer)
			assert.Equal(t, testCase.container.Image, testBuilder.Definition.Spec.DRA.InitContainer.Image)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDRAServiceAccount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		serviceAccount string
		expectedErr    string
	}{
		{
			serviceAccount: "test-sa",
			expectedErr:    "",
		},
		{
			serviceAccount: "",
			expectedErr:    canNotRedefineModuleWithEmpty,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDRAServiceAccount(testCase.serviceAccount)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder.Definition.Spec.DRA)
			assert.Equal(t, testCase.serviceAccount, testBuilder.Definition.Spec.DRA.ServiceAccountName)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDRADriverName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		driverName  string
		expectedErr string
	}{
		{
			driverName:  "gpu.example.com",
			expectedErr: "",
		},
		{
			driverName:  "",
			expectedErr: "cannot set empty DRA driverName",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDRADriverName(testCase.driverName)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder.Definition.Spec.DRA)
			assert.Equal(t, testCase.driverName, testBuilder.Definition.Spec.DRA.DriverName)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDRADeviceClass(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		deviceClass v1beta1.DeviceClassSpec
		expectedErr string
	}{
		{
			deviceClass: v1beta1.DeviceClassSpec{Name: "test-device"},
			expectedErr: "",
		},
		{
			deviceClass: v1beta1.DeviceClassSpec{Name: ""},
			expectedErr: "cannot append DeviceClass with empty name",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDRADeviceClass(testCase.deviceClass)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder.Definition.Spec.DRA)
			assert.Len(t, testBuilder.Definition.Spec.DRA.DeviceClasses, 1)
			assert.Equal(t, testCase.deviceClass.Name, testBuilder.Definition.Spec.DRA.DeviceClasses[0].Name)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDRADeviceClassMultiple(t *testing.T) {
	t.Parallel()

	testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
	testBuilder.WithDRADeviceClass(v1beta1.DeviceClassSpec{Name: "device-a"})
	testBuilder.WithDRADeviceClass(v1beta1.DeviceClassSpec{Name: "device-b"})

	assert.NoError(t, testBuilder.GetError())
	assert.Len(t, testBuilder.Definition.Spec.DRA.DeviceClasses, 2)
	assert.Equal(t, "device-a", testBuilder.Definition.Spec.DRA.DeviceClasses[0].Name)
	assert.Equal(t, "device-b", testBuilder.Definition.Spec.DRA.DeviceClasses[1].Name)
}

func TestModuleWithDRAVolume(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		configMapName string
		expectedErr   string
	}{
		{
			name:          "test-vol",
			configMapName: "test-cm",
			expectedErr:   "",
		},
		{
			name:          "",
			configMapName: "test-cm",
			expectedErr:   "cannot add DRA volume with empty 'name'",
		},
		{
			name:          "test-vol",
			configMapName: "",
			expectedErr:   "cannot add DRA volume with empty 'configMapName'",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDRAVolume(testCase.name, testCase.configMapName)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder.Definition.Spec.DRA)
			assert.Len(t, testBuilder.Definition.Spec.DRA.Volumes, 1)
			assert.Equal(t, testCase.name, testBuilder.Definition.Spec.DRA.Volumes[0].Name)
			assert.Equal(t, testCase.configMapName,
				testBuilder.Definition.Spec.DRA.Volumes[0].ConfigMap.Name)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDRAHostPathVolume(t *testing.T) {
	t.Parallel()

	dirType := corev1.HostPathDirectory

	testCases := []struct {
		name        string
		path        string
		pathType    *corev1.HostPathType
		expectedErr string
	}{
		{
			name:        "kubelet-plugins",
			path:        "/var/lib/kubelet/plugins",
			pathType:    &dirType,
			expectedErr: "",
		},
		{
			name:        "",
			path:        "/var/lib/kubelet/plugins",
			pathType:    &dirType,
			expectedErr: "cannot add DRA hostPath volume with empty 'name'",
		},
		{
			name:        "kubelet-plugins",
			path:        "",
			pathType:    &dirType,
			expectedErr: "cannot add DRA hostPath volume with empty 'path'",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDRAHostPathVolume(testCase.name, testCase.path, testCase.pathType)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder.Definition.Spec.DRA)
			assert.Len(t, testBuilder.Definition.Spec.DRA.Volumes, 1)
			assert.Equal(t, testCase.name, testBuilder.Definition.Spec.DRA.Volumes[0].Name)
			assert.Equal(t, testCase.path, testBuilder.Definition.Spec.DRA.Volumes[0].HostPath.Path)
			assert.Equal(t, testCase.pathType, testBuilder.Definition.Spec.DRA.Volumes[0].HostPath.Type)
		} else {
			assertModuleBuilderError(t, testCase.expectedErr, testBuilder)
		}
	}
}

func TestModuleWithDRAAutomountServiceAccountToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		automount     bool
		expectedValue bool
	}{
		{
			automount:     false,
			expectedValue: false,
		},
		{
			automount:     true,
			expectedValue: true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
		testBuilder.WithDRAAutomountServiceAccountToken(testCase.automount)

		assert.NotNil(t, testBuilder.Definition.Spec.DRA)
		assert.NotNil(t, testBuilder.Definition.Spec.DRA.AutomountServiceAccountToken)
		assert.Equal(t, testCase.expectedValue,
			*testBuilder.Definition.Spec.DRA.AutomountServiceAccountToken)
	}
}

func TestModuleWithServiceAccountDRAPath(t *testing.T) {
	t.Parallel()

	testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
	testBuilder.WithDRAServiceAccount("dra-sa")

	assert.NoError(t, testBuilder.GetError())
	assert.NotNil(t, testBuilder.Definition.Spec.DRA)
	assert.Equal(t, "dra-sa", testBuilder.Definition.Spec.DRA.ServiceAccountName)
}

func TestModuleWithServiceAccountInvalidType(t *testing.T) {
	t.Parallel()

	testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
	testBuilder.withServiceAccount("test-sa", "invalid")

	assert.EqualError(t, testBuilder.GetError(),
		"invalid account type parameter. Supported parameters are: 'module', 'device', 'dra'")
}

func TestModuleWithDRAChaining(t *testing.T) {
	t.Parallel()

	testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
	testBuilder.
		WithDRADriverName("gpu.example.com").
		WithDRAContainer(&v1beta1.CommonContainerSpec{Image: "driver:v1"}).
		WithDRAServiceAccount("dra-sa").
		WithDRADeviceClass(v1beta1.DeviceClassSpec{Name: "test-device"}).
		WithDRAVolume("config", "config-cm").
		WithDRAAutomountServiceAccountToken(true)

	assert.NoError(t, testBuilder.GetError())
	assert.Equal(t, "gpu.example.com", testBuilder.Definition.Spec.DRA.DriverName)
	assert.Equal(t, "driver:v1", testBuilder.Definition.Spec.DRA.Container.Image)
	assert.Equal(t, "dra-sa", testBuilder.Definition.Spec.DRA.ServiceAccountName)
	assert.Len(t, testBuilder.Definition.Spec.DRA.DeviceClasses, 1)
	assert.Len(t, testBuilder.Definition.Spec.DRA.Volumes, 1)
	assert.True(t, *testBuilder.Definition.Spec.DRA.AutomountServiceAccountToken)
}

func TestModuleModifiersOnInvalidBuilder(t *testing.T) {
	t.Parallel()

	testBuilder := buildInValidTestModule(buildModuleTestClientWithDummyObject())
	origErr := testBuilder.GetError()

	testBuilder.WithNodeSelector(map[string]string{"test": "test"})
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithLoadServiceAccount("test")
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDevicePluginServiceAccount("test")
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithImageRepoSecret("test")
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDevicePluginAutomountServiceAccountToken(true)
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDevicePluginVolume("test", "test")
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithToleration("test", "Equals", "test", "NoSchedule", nil)
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithModuleLoaderContainer(&v1beta1.ModuleLoaderContainerSpec{})
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDevicePluginContainer(&v1beta1.DevicePluginContainerSpec{})
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithImageRebuildTriggerGeneration(1)
	assert.Equal(t, origErr, testBuilder.GetError())
}

func TestModuleDRAModifiersOnInvalidBuilder(t *testing.T) {
	t.Parallel()

	testBuilder := buildInValidTestModule(buildModuleTestClientWithDummyObject())
	origErr := testBuilder.GetError()

	testBuilder.WithDRA(&v1beta1.DRASpec{DriverName: "test"})
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDRAContainer(&v1beta1.CommonContainerSpec{Image: "test"})
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDRAInitContainer(&v1beta1.CommonContainerSpec{Image: "test"})
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDRADriverName("test")
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDRAServiceAccount("test")
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDRADeviceClass(v1beta1.DeviceClassSpec{Name: "test"})
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDRAVolume("test", "test")
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDRAHostPathVolume("test", "/test", nil)
	assert.Equal(t, origErr, testBuilder.GetError())

	testBuilder.WithDRAAutomountServiceAccountToken(true)
	assert.Equal(t, origErr, testBuilder.GetError())
}

func TestModuleWithDRAInitializesDRASpec(t *testing.T) {
	t.Parallel()

	testBuilder := buildValidTestModule(buildModuleTestClientWithDummyObject())
	assert.Nil(t, testBuilder.Definition.Spec.DRA)

	testBuilder.WithDRADriverName("test.example.com")
	assert.NotNil(t, testBuilder.Definition.Spec.DRA)
	assert.Equal(t, "test.example.com", testBuilder.Definition.Spec.DRA.DriverName)
}
