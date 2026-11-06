package kmm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	kmmv1beta2 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultPreflightName      = "preflight"
	defaultPreflightNamespace = "preflightns"
)

func TestNewPreflightValidationBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewPreflightValidationBuilder,
		kmmv1beta2.AddToScheme,
		preflightValidationGVK,
	).ExecuteTests(t)
}

func TestPullPreflightValidation(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPreflightValidation,
		kmmv1beta2.AddToScheme,
		preflightValidationGVK,
	).ExecuteTests(t)
}

func TestPreflightValidationBuilderMethods(t *testing.T) {
	t.Parallel()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(preflightValidationCommonTestConfig())).
		With(testhelper.NewExistsTestConfig(preflightValidationCommonTestConfig())).
		With(testhelper.NewCreateTestConfig(preflightValidationCommonTestConfig())).
		With(testhelper.NewDeleteReturnerTestConfig(preflightValidationCommonTestConfig())).
		With(testhelper.NewUpdateTestConfig(preflightValidationCommonTestConfig())).
		Run(t)
}

func TestPreflightValidationWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig[
		kmmv1beta2.PreflightValidation,
		PreflightValidationBuilder,
		*kmmv1beta2.PreflightValidation,
		*PreflightValidationBuilder,
		PreflightValidationAdditionalOptions,
	](preflightValidationCommonTestConfig()).ExecuteTests(t)
}

func TestPreflightValidationWithOptionsAppliesResourceOption(t *testing.T) {
	t.Parallel()

	builder := buildValidPreflightValidationBuilder(newPreflightValidationTestClient())
	builder = builder.WithOptions(func(builder *PreflightValidationBuilder) (*PreflightValidationBuilder, error) {
		builder.Definition.Spec.KernelVersion = "custom-kernel"

		return builder, nil
	})

	require.NoError(t, builder.GetError())
	assert.Equal(t, "custom-kernel", builder.Definition.Spec.KernelVersion)
}

func TestPreflightValidationWithKernelVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		kernelVersion string
		valid         bool
		expectedError error
	}{
		{
			name:          "sets kernel version",
			kernelVersion: "some-kernel",
			valid:         true,
		},
		{
			name:          "empty kernel version",
			kernelVersion: "",
			valid:         true,
			expectedError: fmt.Errorf("%s", errNilKernelVersion),
		},
		{
			name:          "invalid builder",
			kernelVersion: "some-kernel",
			valid:         false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPreflightValidationBuilderForTestCase(testCase.valid)
			builder = builder.WithKernelVersion(testCase.kernelVersion)

			assertPreflightValidationModifierResult(
				t, builder, testCase.valid, testCase.expectedError,
				builder.Definition.Spec.KernelVersion, testCase.kernelVersion)
		})
	}
}

func TestPreflightValidationWithPushBuiltImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		push  bool
		valid bool
	}{
		{
			name:  "enables pushing built images",
			push:  true,
			valid: true,
		},
		{
			name:  "disables pushing built images",
			push:  false,
			valid: true,
		},
		{
			name:  "invalid builder",
			push:  true,
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPreflightValidationBuilderForTestCase(testCase.valid)
			builder = builder.WithPushBuiltImage(testCase.push)

			assertPreflightValidationModifierResult(
				t, builder, testCase.valid, nil, builder.Definition.Spec.PushBuiltImage, testCase.push)
		})
	}
}

// preflightValidationCommonTestConfig returns the shared testhelper configuration for the builder.
func preflightValidationCommonTestConfig() testhelper.CommonTestConfig[
	kmmv1beta2.PreflightValidation,
	PreflightValidationBuilder,
	*kmmv1beta2.PreflightValidation,
	*PreflightValidationBuilder,
] {
	return testhelper.NewCommonTestConfig[
		kmmv1beta2.PreflightValidation,
		PreflightValidationBuilder,
	](kmmv1beta2.AddToScheme, preflightValidationGVK, testhelper.ResourceScopeNamespaced)
}

// buildPreflightValidationBuilderForTestCase returns a valid or invalid builder for modifier tests.
func buildPreflightValidationBuilderForTestCase(valid bool) *PreflightValidationBuilder {
	if valid {
		return buildValidPreflightValidationBuilder(newPreflightValidationTestClient())
	}

	return buildInvalidPreflightValidationBuilder(newPreflightValidationTestClient())
}

// assertPreflightValidationModifierResult checks the error and field value after a modifier is applied.
func assertPreflightValidationModifierResult[T any](
	t *testing.T,
	builder *PreflightValidationBuilder,
	valid bool,
	expectedError error,
	actualValue T,
	expectedValue T,
) {
	t.Helper()

	if !valid {
		require.Error(t, builder.GetError())
		assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))

		return
	}

	assert.Equal(t, expectedError, builder.GetError())

	if expectedError == nil {
		assert.Equal(t, expectedValue, actualValue)
	}
}

// buildValidPreflightValidationBuilder returns a valid PreflightValidationBuilder for testing.
func buildValidPreflightValidationBuilder(apiClient *clients.Settings) *PreflightValidationBuilder {
	return NewPreflightValidationBuilder(apiClient, defaultPreflightName, defaultPreflightNamespace)
}

// buildInvalidPreflightValidationBuilder returns an invalid PreflightValidationBuilder for testing.
func buildInvalidPreflightValidationBuilder(apiClient *clients.Settings) *PreflightValidationBuilder {
	return NewPreflightValidationBuilder(apiClient, defaultPreflightName, "")
}

// newPreflightValidationTestClient returns a test client with the KMM v1beta2 scheme attached.
func newPreflightValidationTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{kmmv1beta2.AddToScheme},
	})
}
