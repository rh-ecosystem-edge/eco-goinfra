package kmm

import (
	"fmt"
	"testing"

	moduleV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestNewRegExKernelMappingBuilder(t *testing.T) {
	testCases := []struct {
		regex         string
		expectedError string
	}{
		{
			regex:         "^.+$",
			expectedError: "",
		},
		{
			regex:         "",
			expectedError: "'regex' parameter can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder(testCase.regex)

		assert.NotNil(t, testBuilder)
		assert.NotNil(t, testBuilder.definition)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.regex, testBuilder.definition.Regexp)
		}
	}
}

func TestNewLiteralKernelMappingBuilder(t *testing.T) {
	testCases := []struct {
		literal       string
		expectedError string
	}{
		{
			literal:       "5.14.0-70.58.1.el9_0.x86_64",
			expectedError: "",
		},
		{
			literal:       "",
			expectedError: "'literal' parameter can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewLiteralKernelMappingBuilder(testCase.literal)

		assert.NotNil(t, testBuilder)
		assert.NotNil(t, testBuilder.definition)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.literal, testBuilder.definition.Literal)
		}
	}
}

func TestBuildKernelMappingConfig(t *testing.T) {
	testCases := []struct {
		regex         string
		expectedError string
	}{
		{
			regex:         "^.+$",
			expectedError: "",
		},
		{
			regex:         "",
			expectedError: "error building KernelMappingConfig config due to :'regex' parameter can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder(testCase.regex)
		config, err := testBuilder.BuildKernelMappingConfig()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, config)
			assert.Equal(t, testCase.regex, config.Regexp)
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.Nil(t, config)
		}
	}
}

func TestWithContainerImage(t *testing.T) {
	testCases := []struct {
		image         string
		expectedError string
	}{
		{
			image:         "quay.io/myrepo/myimage:v1.0",
			expectedError: "",
		},
		{
			image:         "",
			expectedError: "'image' parameter can not be empty for KernelMapping",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.WithContainerImage(testCase.image)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.image, testBuilder.definition.ContainerImage)
		}
	}
}

func TestWithBuildArg(t *testing.T) {
	testCases := []struct {
		argName       string
		argValue      string
		expectedError string
	}{
		{
			argName:       "KERNEL_VERSION",
			argValue:      "5.14.0",
			expectedError: "",
		},
		{
			argName:       "",
			argValue:      "somevalue",
			expectedError: "'argName' parameter can not be empty for KernelMapping BuildArg",
		},
		{
			argName:       "somename",
			argValue:      "",
			expectedError: "'argValue' parameter can not be empty for KernelMapping BuildArg",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.WithBuildArg(testCase.argName, testCase.argValue)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, testBuilder.definition.Build)
			assert.Len(t, testBuilder.definition.Build.BuildArgs, 1)
			assert.Equal(t, testCase.argName, testBuilder.definition.Build.BuildArgs[0].Name)
			assert.Equal(t, testCase.argValue, testBuilder.definition.Build.BuildArgs[0].Value)
		}
	}
}

func TestWithBuildArgMultiple(t *testing.T) {
	testBuilder := NewRegExKernelMappingBuilder("^.+$")
	testBuilder.WithBuildArg("ARG1", "value1")
	testBuilder.WithBuildArg("ARG2", "value2")

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.NotNil(t, testBuilder.definition.Build)
	assert.Len(t, testBuilder.definition.Build.BuildArgs, 2)
	assert.Equal(t, "ARG1", testBuilder.definition.Build.BuildArgs[0].Name)
	assert.Equal(t, "value1", testBuilder.definition.Build.BuildArgs[0].Value)
	assert.Equal(t, "ARG2", testBuilder.definition.Build.BuildArgs[1].Name)
	assert.Equal(t, "value2", testBuilder.definition.Build.BuildArgs[1].Value)
}

func TestWithBuildSecret(t *testing.T) {
	testCases := []struct {
		secret        string
		expectedError string
	}{
		{
			secret:        "my-registry-secret",
			expectedError: "",
		},
		{
			secret:        "",
			expectedError: "'secret' parameter can not be empty for KernelMapping Secret",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.WithBuildSecret(testCase.secret)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, testBuilder.definition.Build)
			assert.Len(t, testBuilder.definition.Build.Secrets, 1)
			assert.Equal(t, testCase.secret, testBuilder.definition.Build.Secrets[0].Name)
		}
	}
}

func TestWithBuildSecretMultiple(t *testing.T) {
	testBuilder := NewRegExKernelMappingBuilder("^.+$")
	testBuilder.WithBuildSecret("secret1")
	testBuilder.WithBuildSecret("secret2")

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.NotNil(t, testBuilder.definition.Build)
	assert.Len(t, testBuilder.definition.Build.Secrets, 2)
	assert.Equal(t, "secret1", testBuilder.definition.Build.Secrets[0].Name)
	assert.Equal(t, "secret2", testBuilder.definition.Build.Secrets[1].Name)
}

func TestWithBuildImageRegistryTLS(t *testing.T) {
	testCases := []struct {
		insecure      bool
		skipTLSVerify bool
	}{
		{
			insecure:      true,
			skipTLSVerify: true,
		},
		{
			insecure:      false,
			skipTLSVerify: false,
		},
		{
			insecure:      true,
			skipTLSVerify: false,
		},
		{
			insecure:      false,
			skipTLSVerify: true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.WithBuildImageRegistryTLS(testCase.insecure, testCase.skipTLSVerify)

		assert.Equal(t, "", testBuilder.errorMsg)
		assert.NotNil(t, testBuilder.definition.Build)
		assert.Equal(t, testCase.insecure, testBuilder.definition.Build.BaseImageRegistryTLS.Insecure)
		assert.Equal(t, testCase.skipTLSVerify, testBuilder.definition.Build.BaseImageRegistryTLS.InsecureSkipTLSVerify)
	}
}

func TestWithBuildDockerCfgFile(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "my-dockerfile-configmap",
			expectedError: "",
		},
		{
			name:          "",
			expectedError: "'name' parameter can not be empty for KernelMapping Docker file",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.WithBuildDockerCfgFile(testCase.name)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, testBuilder.definition.Build)
			assert.NotNil(t, testBuilder.definition.Build.DockerfileConfigMap)
			assert.Equal(t, testCase.name, testBuilder.definition.Build.DockerfileConfigMap.Name)
		}
	}
}

func TestWithSign(t *testing.T) {
	testCases := []struct {
		certSecret    string
		keySecret     string
		filesToSign   []string
		expectedError string
	}{
		{
			certSecret:    "signing-cert",
			keySecret:     "signing-key",
			filesToSign:   []string{"/opt/lib/modules/driver.ko"},
			expectedError: "",
		},
		{
			certSecret:    "",
			keySecret:     "signing-key",
			filesToSign:   []string{"/opt/lib/modules/driver.ko"},
			expectedError: "'certSecret' parameter can not be empty for KernelMapping Sign",
		},
		{
			certSecret:    "signing-cert",
			keySecret:     "",
			filesToSign:   []string{"/opt/lib/modules/driver.ko"},
			expectedError: "'keySecret' parameter can not be empty for KernelMapping Sign",
		},
		{
			certSecret:    "signing-cert",
			keySecret:     "signing-key",
			filesToSign:   []string{},
			expectedError: "'fileToSign' parameter can not be empty for KernelMapping Sign",
		},
		{
			certSecret:    "signing-cert",
			keySecret:     "signing-key",
			filesToSign:   nil,
			expectedError: "'fileToSign' parameter can not be empty for KernelMapping Sign",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.WithSign(testCase.certSecret, testCase.keySecret, testCase.filesToSign)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, testBuilder.definition.Sign)
			assert.Equal(t, testCase.certSecret, testBuilder.definition.Sign.CertSecret.Name)
			assert.Equal(t, testCase.keySecret, testBuilder.definition.Sign.KeySecret.Name)
			assert.Equal(t, testCase.filesToSign, testBuilder.definition.Sign.FilesToSign)
		}
	}
}

func TestRegistryTLS(t *testing.T) {
	testCases := []struct {
		insecure      bool
		skipTLSVerify bool
	}{
		{
			insecure:      true,
			skipTLSVerify: true,
		},
		{
			insecure:      false,
			skipTLSVerify: false,
		},
		{
			insecure:      true,
			skipTLSVerify: false,
		},
		{
			insecure:      false,
			skipTLSVerify: true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.RegistryTLS(testCase.insecure, testCase.skipTLSVerify)

		assert.Equal(t, "", testBuilder.errorMsg)
		assert.NotNil(t, testBuilder.definition.RegistryTLS)
		assert.Equal(t, testCase.insecure, testBuilder.definition.RegistryTLS.Insecure)
		assert.Equal(t, testCase.skipTLSVerify, testBuilder.definition.RegistryTLS.InsecureSkipTLSVerify)
	}
}

func TestWithInTreeModuleToRemove(t *testing.T) {
	testCases := []struct {
		existingModule string
		expectedError  string
	}{
		{
			existingModule: "i915",
			expectedError:  "",
		},
		{
			existingModule: "",
			expectedError:  "'existingModule' parameter can not be empty for KernelMapping inTreeModuleToRemove",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.WithInTreeModuleToRemove(testCase.existingModule)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, []string{testCase.existingModule}, testBuilder.definition.InTreeModulesToRemove)
		}
	}
}

func TestWithInTreeModulesToRemove(t *testing.T) {
	testCases := []struct {
		existingModulesList []string
		expectedError       string
	}{
		{
			existingModulesList: []string{"i915"},
			expectedError:       "",
		},
		{
			existingModulesList: []string{"i915", "nouveau"},
			expectedError:       "",
		},
		{
			existingModulesList: []string{},
			expectedError:       "'existingModuleList' parameter can not be empty for KernelMapping inTreeModulesToRemove",
		},
		{
			existingModulesList: nil,
			expectedError:       "'existingModuleList' parameter can not be empty for KernelMapping inTreeModulesToRemove",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.WithInTreeModulesToRemove(testCase.existingModulesList)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.existingModulesList, testBuilder.definition.InTreeModulesToRemove)
		}
	}
}

func TestWithOptions(t *testing.T) {
	testCases := []struct {
		options       KernelMappingAdditionalOptions
		expectedError string
	}{
		{
			options: func(builder *KernelMappingBuilder) (*KernelMappingBuilder, error) {
				builder.definition.ContainerImage = "test-image:latest"

				return builder, nil
			},
			expectedError: "",
		},
		{
			options: func(builder *KernelMappingBuilder) (*KernelMappingBuilder, error) {
				return builder, fmt.Errorf("error adding additional option")
			},
			expectedError: "error adding additional option",
		},
		{
			options:       nil,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewRegExKernelMappingBuilder("^.+$")
		testBuilder.WithOptions(testCase.options)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" && testCase.options != nil {
			assert.Equal(t, "test-image:latest", testBuilder.definition.ContainerImage)
		}
	}
}

func TestWithOptionsMultiple(t *testing.T) {
	testBuilder := NewRegExKernelMappingBuilder("^.+$")
	testBuilder.WithOptions(
		func(builder *KernelMappingBuilder) (*KernelMappingBuilder, error) {
			builder.definition.ContainerImage = "test-image:latest"

			return builder, nil
		},
		func(builder *KernelMappingBuilder) (*KernelMappingBuilder, error) {
			builder.definition.RegistryTLS = &moduleV1Beta1.TLSOptions{Insecure: true}

			return builder, nil
		},
	)

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.Equal(t, "test-image:latest", testBuilder.definition.ContainerImage)
}

func TestKernelMappingBuilderChaining(t *testing.T) {
	testBuilder := NewRegExKernelMappingBuilder("^.+$").
		WithContainerImage("quay.io/myrepo/driver:v1.0").
		WithBuildArg("KERNEL_VERSION", "5.14.0").
		WithBuildSecret("my-secret").
		WithBuildDockerCfgFile("my-dockerfile").
		RegistryTLS(true, false)

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.Equal(t, "^.+$", testBuilder.definition.Regexp)
	assert.Equal(t, "quay.io/myrepo/driver:v1.0", testBuilder.definition.ContainerImage)
	assert.NotNil(t, testBuilder.definition.Build)
	assert.Len(t, testBuilder.definition.Build.BuildArgs, 1)
	assert.Len(t, testBuilder.definition.Build.Secrets, 1)
	assert.NotNil(t, testBuilder.definition.Build.DockerfileConfigMap)
	assert.NotNil(t, testBuilder.definition.RegistryTLS)
	assert.True(t, testBuilder.definition.RegistryTLS.Insecure)
	assert.False(t, testBuilder.definition.RegistryTLS.InsecureSkipTLSVerify)
}

func TestKernelMappingBuilderWithInvalidBuilder(t *testing.T) {
	// Test that methods return early when builder has error
	testBuilder := NewRegExKernelMappingBuilder("")

	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	// Calling methods on invalid builder should not change error message
	testBuilder.WithContainerImage("test-image")
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	testBuilder.WithBuildArg("arg", "value")
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	testBuilder.WithBuildSecret("secret")
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	testBuilder.WithBuildImageRegistryTLS(true, true)
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	testBuilder.WithBuildDockerCfgFile("dockerfile")
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	testBuilder.WithSign("cert", "key", []string{"file"})
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	testBuilder.RegistryTLS(true, true)
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	testBuilder.WithInTreeModuleToRemove("module")
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	testBuilder.WithInTreeModulesToRemove([]string{"module1", "module2"})
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)

	testBuilder.WithOptions(func(builder *KernelMappingBuilder) (*KernelMappingBuilder, error) {
		return builder, nil
	})
	assert.Equal(t, "'regex' parameter can not be empty", testBuilder.errorMsg)
}

func TestLiteralKernelMappingBuilderChaining(t *testing.T) {
	testBuilder := NewLiteralKernelMappingBuilder("5.14.0-70.58.1.el9_0.x86_64").
		WithContainerImage("quay.io/myrepo/driver:v1.0").
		WithSign("cert-secret", "key-secret", []string{"/lib/modules/driver.ko"})

	assert.Equal(t, "", testBuilder.errorMsg)
	assert.Equal(t, "5.14.0-70.58.1.el9_0.x86_64", testBuilder.definition.Literal)
	assert.Equal(t, "quay.io/myrepo/driver:v1.0", testBuilder.definition.ContainerImage)
	assert.NotNil(t, testBuilder.definition.Sign)

	config, err := testBuilder.BuildKernelMappingConfig()
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "5.14.0-70.58.1.el9_0.x86_64", config.Literal)
}
