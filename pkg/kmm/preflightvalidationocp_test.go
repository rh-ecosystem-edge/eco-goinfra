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
	defaultPreflightOCPName      = "preflightocp"
	defaultPreflightOCPNamespace = "preflightocpns"
)

var preflightOCPGVK = kmmv1beta2.GroupVersion.WithKind("PreflightValidationOCP")

func TestNewPreflightValidationOCPBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewPreflightValidationOCPBuilder,
		kmmv1beta2.AddToScheme,
		preflightOCPGVK,
	).ExecuteTests(t)
}

func TestPullPreflightValidationOCP(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPreflightValidationOCP,
		kmmv1beta2.AddToScheme,
		preflightOCPGVK,
	).ExecuteTests(t)
}

func TestPreflightValidationOCPBuilderMethods(t *testing.T) {
	t.Parallel()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(preflightOCPCommonTestConfig())).
		With(testhelper.NewExistsTestConfig(preflightOCPCommonTestConfig())).
		With(testhelper.NewCreateTestConfig(preflightOCPCommonTestConfig())).
		With(testhelper.NewDeleteReturnerTestConfig(preflightOCPCommonTestConfig())).
		With(testhelper.NewUpdateTestConfig(preflightOCPCommonTestConfig())).
		Run(t)
}

func TestPreflightValidationOCPWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig[
		kmmv1beta2.PreflightValidationOCP,
		PreflightValidationOCPBuilder,
		*kmmv1beta2.PreflightValidationOCP,
		*PreflightValidationOCPBuilder,
		PreflightValidationOCPAdditionalOptions,
	](preflightOCPCommonTestConfig()).ExecuteTests(t)
}

func TestPreflightValidationOCPWithOptionsAppliesResourceOption(t *testing.T) {
	t.Parallel()

	builder := buildValidPreflightOCPBuilder(newPreflightOCPTestClient())
	builder = builder.WithOptions(func(builder *PreflightValidationOCPBuilder) (*PreflightValidationOCPBuilder, error) {
		builder.Definition.Spec.KernelVersion = "custom-kernel"

		return builder, nil
	})

	require.NoError(t, builder.GetError())
	assert.Equal(t, "custom-kernel", builder.Definition.Spec.KernelVersion)
}

func TestPreflightValidationOCPWithDtkImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		dtkImage      string
		valid         bool
		expectedError error
	}{
		{
			name:     "sets DTK image",
			dtkImage: "some-image",
			valid:    true,
		},
		{
			name:          "empty DTK image",
			dtkImage:      "",
			valid:         true,
			expectedError: fmt.Errorf("invalid 'dtkImage' argument can not be nil"),
		},
		{
			name:     "invalid builder",
			dtkImage: "some-image",
			valid:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPreflightOCPBuilderForTestCase(testCase.valid)
			builder = builder.WithDtkImage(testCase.dtkImage)

			assertPreflightOCPModifierResult(
				t, builder, testCase.valid, testCase.expectedError, builder.Definition.Spec.DTKImage, testCase.dtkImage)
		})
	}
}

func TestPreflightValidationOCPWithKernelVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		kernelVersion string
		valid         bool
		expectedError error
	}{
		{
			name:          "sets kernel version",
			kernelVersion: "some-image",
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
			kernelVersion: "some-image",
			valid:         false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPreflightOCPBuilderForTestCase(testCase.valid)
			builder = builder.WithKernelVersion(testCase.kernelVersion)

			assertPreflightOCPModifierResult(
				t, builder, testCase.valid, testCase.expectedError,
				builder.Definition.Spec.KernelVersion, testCase.kernelVersion)
		})
	}
}

func TestPreflightValidationOCPWithPushBuiltImage(t *testing.T) {
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

			builder := buildPreflightOCPBuilderForTestCase(testCase.valid)
			builder = builder.WithPushBuiltImage(testCase.push)

			assertPreflightOCPModifierResult(
				t, builder, testCase.valid, nil, builder.Definition.Spec.PushBuiltImage, testCase.push)
		})
	}
}

// preflightOCPCommonTestConfig returns the shared testhelper configuration for the PreflightValidationOCP builder.
func preflightOCPCommonTestConfig() testhelper.CommonTestConfig[
	kmmv1beta2.PreflightValidationOCP,
	PreflightValidationOCPBuilder,
	*kmmv1beta2.PreflightValidationOCP,
	*PreflightValidationOCPBuilder,
] {
	return testhelper.NewCommonTestConfig[
		kmmv1beta2.PreflightValidationOCP,
		PreflightValidationOCPBuilder,
	](kmmv1beta2.AddToScheme, preflightOCPGVK, testhelper.ResourceScopeNamespaced)
}

// buildPreflightOCPBuilderForTestCase returns a valid or invalid builder for modifier tests.
func buildPreflightOCPBuilderForTestCase(valid bool) *PreflightValidationOCPBuilder {
	if valid {
		return buildValidPreflightOCPBuilder(newPreflightOCPTestClient())
	}

	return buildInvalidPreflightOCPBuilder(newPreflightOCPTestClient())
}

// assertPreflightOCPModifierResult checks the error and field value after a modifier is applied.
func assertPreflightOCPModifierResult[T any](
	t *testing.T,
	builder *PreflightValidationOCPBuilder,
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

// buildValidPreflightOCPBuilder returns a valid PreflightValidationOCPBuilder for testing.
func buildValidPreflightOCPBuilder(apiClient *clients.Settings) *PreflightValidationOCPBuilder {
	return NewPreflightValidationOCPBuilder(apiClient, defaultPreflightOCPName, defaultPreflightOCPNamespace)
}

// buildInvalidPreflightOCPBuilder returns an invalid PreflightValidationOCPBuilder for testing.
func buildInvalidPreflightOCPBuilder(apiClient *clients.Settings) *PreflightValidationOCPBuilder {
	return NewPreflightValidationOCPBuilder(apiClient, defaultPreflightOCPName, "")
}

// newPreflightOCPTestClient returns a test client with the KMM v1beta2 scheme attached.
func newPreflightOCPTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{kmmv1beta2.AddToScheme},
	})
}
