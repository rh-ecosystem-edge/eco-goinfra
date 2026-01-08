package kmm

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestNewModLoaderContainerBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "kmod",
			expectedError: "",
		},
		{
			name:          "",
			expectedError: "'modName' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testModuleLoaderContainerBuilder := NewModLoaderContainerBuilder(testCase.name)
		assert.Equal(t, testCase.expectedError, testModuleLoaderContainerBuilder.errorMsg)
		assert.NotNil(t, testModuleLoaderContainerBuilder.definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testModuleLoaderContainerBuilder.definition.Modprobe.ModuleName)
		}
	}
}

func TestModuleLoaderContainerWithModprobeSpec(t *testing.T) {
	testCases := []struct {
		dirName            string
		fwPath             string
		parameters         []string
		args               []string
		rawargs            []string
		moduleLoadingOrder []string
	}{
		{
			dirName:            "",
			fwPath:             "",
			parameters:         nil,
			moduleLoadingOrder: nil,
			args:               nil,
			rawargs:            nil,
		},
		{
			dirName:            "test",
			fwPath:             "test",
			parameters:         []string{"one", "two"},
			moduleLoadingOrder: []string{"one", "two"},
			args:               []string{},
			rawargs:            []string{},
		},
		{
			dirName:            "test",
			fwPath:             "test",
			parameters:         []string{"one", "two"},
			moduleLoadingOrder: []string{"one", "two"},
			args:               []string{"arg"},
			rawargs:            []string{},
		},
		{
			dirName:            "test",
			fwPath:             "test",
			parameters:         []string{"one", "two"},
			moduleLoadingOrder: []string{"one", "two"},
			args:               []string{},
			rawargs:            []string{"arg"},
		},
		{
			dirName:            "test",
			fwPath:             "test",
			parameters:         []string{"one", "two"},
			moduleLoadingOrder: []string{"one", "two"},
			args:               []string{"arg"},
			rawargs:            []string{"rawarg1", "rawargs2"},
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewModLoaderContainerBuilder("test")
		testBuilder.WithModprobeSpec(testCase.dirName, testCase.fwPath,
			testCase.parameters, testCase.args, testCase.rawargs, testCase.moduleLoadingOrder)

		assert.Equal(t, testCase.dirName, testBuilder.definition.Modprobe.DirName)
		assert.Equal(t, testCase.fwPath, testBuilder.definition.Modprobe.FirmwarePath)
		assert.Equal(t, testCase.parameters, testBuilder.definition.Modprobe.Parameters)
		assert.Equal(t, testCase.moduleLoadingOrder, testBuilder.definition.Modprobe.ModulesLoadingOrder)

		if len(testCase.args) > 0 {
			assert.Equal(t, testCase.args, testBuilder.definition.Modprobe.Args.Load)
		}

		if len(testCase.rawargs) > 0 {
			assert.Equal(t, testCase.rawargs, testBuilder.definition.Modprobe.RawArgs.Load)
		}
	}
}

func TestModuleLoaderContainerWithImagePullPolicy(t *testing.T) {
	testCases := []struct {
		imagePolicy   string
		expectedError string
	}{
		{
			imagePolicy:   "",
			expectedError: "'policy' can not be empty",
		},
		{
			imagePolicy:   "SomePolicy",
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewModLoaderContainerBuilder("test")
		testBuilder.WithImagePullPolicy(testCase.imagePolicy)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, corev1.PullPolicy(testCase.imagePolicy), testBuilder.definition.ImagePullPolicy)
		}
	}
}

func TestModuleLoaderContainerWithKernelMapping(t *testing.T) {
	testCases := []struct {
		mapping       *v1beta1.KernelMapping
		expectedError string
	}{
		{
			mapping:       buildRegExKernelMapping(""),
			expectedError: "'mapping' can not be empty nil",
		},
		{
			mapping:       buildRegExKernelMapping("^.+$"),
			expectedError: "",
		},
		{
			mapping:       buildLiteralKernelMapping("5.14.0-70.58.1.el9_0.x86_64"),
			expectedError: "",
		},
		{
			mapping:       buildLiteralKernelMapping(""),
			expectedError: "'mapping' can not be empty nil",
		},
	}

	for _, testcase := range testCases {
		testBuilder := NewModLoaderContainerBuilder("test")
		testBuilder.WithKernelMapping(testcase.mapping)

		if testcase.expectedError != "" {
			assert.Equal(t, testcase.expectedError, testBuilder.errorMsg)
		} else {
			assert.Equal(t, testBuilder.definition.KernelMappings[0], *testcase.mapping)
		}
	}
}

func TestModuleLoaderContainerWithOptions(t *testing.T) {
	testBuilder := NewModLoaderContainerBuilder("test").WithOptions(
		func(builder *ModuleLoaderContainerBuilder) (*ModuleLoaderContainerBuilder, error) {
			return builder, nil
		})
	assert.Equal(t, "", testBuilder.errorMsg)

	testBuilder = NewModLoaderContainerBuilder("test").WithOptions(
		func(builder *ModuleLoaderContainerBuilder) (*ModuleLoaderContainerBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestModuleLoaderContainerWithVersion(t *testing.T) {
	testCases := []struct {
		version       string
		expectedError string
	}{
		{
			version:       "",
			expectedError: "'version' can not be empty",
		},
		{
			version:       "1.1",
			expectedError: "",
		},
	}

	for _, testcase := range testCases {
		testBuilder := NewModLoaderContainerBuilder("test")
		testBuilder.WithVersion(testcase.version)

		if testcase.expectedError != "" {
			assert.Equal(t, testcase.expectedError, testBuilder.errorMsg)
		} else {
			assert.Equal(t, testBuilder.definition.Version, testcase.version)
		}
	}
}

func TestModuleLoaderContainerBuildModuleLoaderContainerCfg(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
		mutate        bool
	}{
		{
			name:          "kmod",
			expectedError: "",
			mutate:        false,
		},
		{
			name:          "",
			expectedError: "'modName' cannot be empty",
			mutate:        false,
		},
		{
			name:          "kmod",
			expectedError: "'mapping' can not be empty nil",
			mutate:        true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewModLoaderContainerBuilder(testCase.name)

		if testCase.mutate {
			testBuilder.WithKernelMapping(nil)
		}

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
		assert.NotNil(t, testBuilder.definition)

		if testCase.expectedError == "" || testCase.name != "" {
			assert.Equal(t, testCase.name, testBuilder.definition.Modprobe.ModuleName)
		}
	}
}

func TestNewDevicePluginContainerBuilder(t *testing.T) {
	testCases := []struct {
		image         string
		expectedError string
	}{
		{
			image:         "quay.io/myrepo/device-plugin:v1.0",
			expectedError: "",
		},
		{
			image:         "",
			expectedError: "invalid parameter 'image' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewDevicePluginContainerBuilder(testCase.image)

		assert.NotNil(t, testBuilder)
		assert.NotNil(t, testBuilder.definition)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.image, testBuilder.definition.Image)
		}
	}
}

func TestDevicePluginContainerWithEnv(t *testing.T) {
	testCases := []struct {
		name          string
		value         string
		expectedError string
	}{
		{
			name:          "MY_ENV_VAR",
			value:         "my-value",
			expectedError: "",
		},
		{
			name:          "",
			value:         "some-value",
			expectedError: "'name' can not be empty for DevicePlugin Env",
		},
		{
			name:          "some-name",
			value:         "",
			expectedError: "'value' can not be empty for DevicePlugin Env",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewDevicePluginContainerBuilder("test-image:latest")
		testBuilder.WithEnv(testCase.name, testCase.value)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Len(t, testBuilder.definition.Env, 1)
			assert.Equal(t, testCase.name, testBuilder.definition.Env[0].Name)
			assert.Equal(t, testCase.value, testBuilder.definition.Env[0].Value)
		}
	}
}

func TestDevicePluginContainerWithEnvMultiple(t *testing.T) {
	testBuilder := NewDevicePluginContainerBuilder("test-image:latest")
	testBuilder.WithEnv("ENV1", "value1")
	testBuilder.WithEnv("ENV2", "value2")

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.Len(t, testBuilder.definition.Env, 2)
	assert.Equal(t, "ENV1", testBuilder.definition.Env[0].Name)
	assert.Equal(t, "value1", testBuilder.definition.Env[0].Value)
	assert.Equal(t, "ENV2", testBuilder.definition.Env[1].Name)
	assert.Equal(t, "value2", testBuilder.definition.Env[1].Value)
}

func TestDevicePluginContainerWithVolumeMount(t *testing.T) {
	testCases := []struct {
		mountPath     string
		name          string
		expectedError string
	}{
		{
			mountPath:     "/dev/vfio",
			name:          "vfio-volume",
			expectedError: "",
		},
		{
			mountPath:     "/some/path",
			name:          "",
			expectedError: "'name' can not be empty for DevicePlugin mountPath",
		},
		{
			mountPath:     "",
			name:          "some-name",
			expectedError: "'mountPath' can not be empty for DevicePlugin mountPath",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewDevicePluginContainerBuilder("test-image:latest")
		testBuilder.WithVolumeMount(testCase.mountPath, testCase.name)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Len(t, testBuilder.definition.VolumeMounts, 1)
			assert.Equal(t, testCase.name, testBuilder.definition.VolumeMounts[0].Name)
			assert.Equal(t, testCase.mountPath, testBuilder.definition.VolumeMounts[0].MountPath)
		}
	}
}

func TestDevicePluginContainerWithVolumeMountMultiple(t *testing.T) {
	testBuilder := NewDevicePluginContainerBuilder("test-image:latest")
	testBuilder.WithVolumeMount("/dev/vfio", "vfio-volume")
	testBuilder.WithVolumeMount("/var/run/plugin", "plugin-socket")

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.Len(t, testBuilder.definition.VolumeMounts, 2)
	assert.Equal(t, "vfio-volume", testBuilder.definition.VolumeMounts[0].Name)
	assert.Equal(t, "/dev/vfio", testBuilder.definition.VolumeMounts[0].MountPath)
	assert.Equal(t, "plugin-socket", testBuilder.definition.VolumeMounts[1].Name)
	assert.Equal(t, "/var/run/plugin", testBuilder.definition.VolumeMounts[1].MountPath)
}

func TestGetDevicePluginContainerConfig(t *testing.T) {
	testCases := []struct {
		image         string
		expectedError string
	}{
		{
			image:         "test-image:latest",
			expectedError: "",
		},
		{
			image:         "",
			expectedError: "error building DevicePluginContainerSpec config due to :invalid parameter 'image' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewDevicePluginContainerBuilder(testCase.image)
		config, err := testBuilder.GetDevicePluginContainerConfig()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, config)
			assert.Equal(t, testCase.image, config.Image)
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.Nil(t, config)
		}
	}
}

func TestDevicePluginContainerBuilderChaining(t *testing.T) {
	testBuilder := NewDevicePluginContainerBuilder("quay.io/myrepo/plugin:v1.0").
		WithEnv("DEBUG", "true").
		WithEnv("LOG_LEVEL", "info").
		WithVolumeMount("/dev/vfio", "vfio-volume").
		WithVolumeMount("/var/run/plugin", "plugin-socket")

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.Equal(t, "quay.io/myrepo/plugin:v1.0", testBuilder.definition.Image)
	assert.Len(t, testBuilder.definition.Env, 2)
	assert.Len(t, testBuilder.definition.VolumeMounts, 2)

	config, err := testBuilder.GetDevicePluginContainerConfig()
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "quay.io/myrepo/plugin:v1.0", config.Image)
}

func TestDevicePluginContainerBuilderWithInvalidBuilder(t *testing.T) {
	// Test that methods return early when builder has error
	testBuilder := NewDevicePluginContainerBuilder("")

	assert.Equal(t, "invalid parameter 'image' cannot be empty", testBuilder.errorMsg)

	// Calling methods on invalid builder should not change error message
	testBuilder.WithEnv("ENV", "value")
	assert.Equal(t, "invalid parameter 'image' cannot be empty", testBuilder.errorMsg)

	testBuilder.WithVolumeMount("/path", "name")
	assert.Equal(t, "invalid parameter 'image' cannot be empty", testBuilder.errorMsg)
}

func TestModuleLoaderContainerBuilderWithInvalidBuilder(t *testing.T) {
	// Test that methods return early when builder has error
	testBuilder := NewModLoaderContainerBuilder("")

	assert.Equal(t, "'modName' cannot be empty", testBuilder.errorMsg)

	// Calling methods on invalid builder should not change error message
	testBuilder.WithModprobeSpec("dir", "path", nil, nil, nil, nil)
	assert.Equal(t, "'modName' cannot be empty", testBuilder.errorMsg)

	testBuilder.WithKernelMapping(nil)
	assert.Equal(t, "'modName' cannot be empty", testBuilder.errorMsg)

	testBuilder.WithImagePullPolicy("Always")
	assert.Equal(t, "'modName' cannot be empty", testBuilder.errorMsg)

	testBuilder.WithVersion("1.0")
	assert.Equal(t, "'modName' cannot be empty", testBuilder.errorMsg)

	testBuilder.WithOptions(func(builder *ModuleLoaderContainerBuilder) (*ModuleLoaderContainerBuilder, error) {
		return builder, nil
	})
	assert.Equal(t, "'modName' cannot be empty", testBuilder.errorMsg)
}

func TestModuleLoaderContainerWithKernelMappingMultiple(t *testing.T) {
	regexMapping := buildRegExKernelMapping("^.+$")
	literalMapping := buildLiteralKernelMapping("5.14.0-70.58.1.el9_0.x86_64")

	testBuilder := NewModLoaderContainerBuilder("test")
	testBuilder.WithKernelMapping(regexMapping)
	testBuilder.WithKernelMapping(literalMapping)

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.Len(t, testBuilder.definition.KernelMappings, 2)
	assert.Equal(t, *regexMapping, testBuilder.definition.KernelMappings[0])
	assert.Equal(t, *literalMapping, testBuilder.definition.KernelMappings[1])
}

func TestModuleLoaderContainerWithOptionsMultiple(t *testing.T) {
	testBuilder := NewModLoaderContainerBuilder("test").WithOptions(
		func(builder *ModuleLoaderContainerBuilder) (*ModuleLoaderContainerBuilder, error) {
			builder.definition.Version = "1.0"

			return builder, nil
		},
		func(builder *ModuleLoaderContainerBuilder) (*ModuleLoaderContainerBuilder, error) {
			builder.definition.ImagePullPolicy = corev1.PullAlways

			return builder, nil
		},
	)

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.Equal(t, "1.0", testBuilder.definition.Version)
	assert.Equal(t, corev1.PullAlways, testBuilder.definition.ImagePullPolicy)
}

func TestModuleLoaderContainerWithOptionsNil(t *testing.T) {
	testBuilder := NewModLoaderContainerBuilder("test").WithOptions(nil)

	assert.Equal(t, "", testBuilder.errorMsg)
}

func TestBuildModuleLoaderContainerCfgSuccess(t *testing.T) {
	testBuilder := NewModLoaderContainerBuilder("kmod").
		WithVersion("1.0").
		WithImagePullPolicy("Always")

	config, err := testBuilder.BuildModuleLoaderContainerCfg()

	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "kmod", config.Modprobe.ModuleName)
	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, corev1.PullPolicy("Always"), config.ImagePullPolicy)
}

func TestModuleLoaderContainerChaining(t *testing.T) {
	regexMapping := buildRegExKernelMapping("^.+$")

	testBuilder := NewModLoaderContainerBuilder("test-module").
		WithModprobeSpec("/opt/lib/modules", "/lib/firmware", []string{"param1"}, []string{"arg1"}, nil, []string{"mod1", "mod2"}).
		WithKernelMapping(regexMapping).
		WithImagePullPolicy("Always").
		WithVersion("2.0")

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.Equal(t, "test-module", testBuilder.definition.Modprobe.ModuleName)
	assert.Equal(t, "/opt/lib/modules", testBuilder.definition.Modprobe.DirName)
	assert.Equal(t, "/lib/firmware", testBuilder.definition.Modprobe.FirmwarePath)
	assert.Equal(t, []string{"param1"}, testBuilder.definition.Modprobe.Parameters)
	assert.Equal(t, []string{"arg1"}, testBuilder.definition.Modprobe.Args.Load)
	assert.Equal(t, []string{"mod1", "mod2"}, testBuilder.definition.Modprobe.ModulesLoadingOrder)
	assert.Len(t, testBuilder.definition.KernelMappings, 1)
	assert.Equal(t, corev1.PullPolicy("Always"), testBuilder.definition.ImagePullPolicy)
	assert.Equal(t, "2.0", testBuilder.definition.Version)

	config, err := testBuilder.BuildModuleLoaderContainerCfg()
	assert.Nil(t, err)
	assert.NotNil(t, config)
}

func buildRegExKernelMapping(regexp string) *v1beta1.KernelMapping {
	reg := NewRegExKernelMappingBuilder(regexp)
	regexBuild, _ := reg.BuildKernelMappingConfig()

	return regexBuild
}

func buildLiteralKernelMapping(literal string) *v1beta1.KernelMapping {
	lit := NewLiteralKernelMappingBuilder(literal)
	litBuild, _ := lit.BuildKernelMappingConfig()

	return litBuild
}
