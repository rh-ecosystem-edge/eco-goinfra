package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
)

// Validator is an interface for builders that have a Validate method.
type Validator[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Validate() error
}

// GenericValidateFunc is the signature for the common.Validate function that takes a builder.
type GenericValidateFunc[O any, SO common.ObjectPointer[O]] func(builder common.Builder[O, SO]) error

// internalValidateFunc is the internal function signature used by ValidateTestConfig.
type internalValidateFunc[
	O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(builder SB) error

// ValidateTestConfig provides the configuration needed to test a Validate method.
type ValidateTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// validateFunc is a function that validates the builder and returns an error.
	validateFunc internalValidateFunc[O, B, SO, SB]
}

// NewValidateTestConfig creates a new ValidateTestConfig with the given parameters for builders that implement the
// Validator interface.
func NewValidateTestConfig[O, B any, SO common.ObjectPointer[O], SB Validator[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) ValidateTestConfig[O, B, SO, SB] {
	return ValidateTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		validateFunc: func(builder SB) error {
			return builder.Validate()
		},
	}
}

// NewGenericValidateTestConfig creates a new ValidateTestConfig with a custom validate function. This is useful for
// testing standalone functions like common.Validate() rather than builder methods.
func NewGenericValidateTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	validateFunc GenericValidateFunc[O, SO],
) ValidateTestConfig[O, B, SO, SB] {
	return ValidateTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		validateFunc: func(builder SB) error {
			return validateFunc(builder)
		},
	}
}

// Name returns the name to use for running these tests.
func (config ValidateTestConfig[O, B, SO, SB]) Name() string {
	return "Validate"
}

// ExecuteTests runs the standard set of Validate tests for the configured resource.
func (config ValidateTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	testCases := []struct {
		name          string
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		builderError  error
		assertError   func(error) bool
	}{
		{
			name:        "valid builder",
			assertError: isErrorNil,
		},
		{
			name:        "nil builder",
			builderNil:  true,
			assertError: commonerrors.IsBuilderNil,
		},
		{
			name:          "nil definition",
			definitionNil: true,
			assertError:   commonerrors.IsBuilderDefinitionNil,
		},
		{
			name:         "nil apiClient",
			apiClientNil: true,
			assertError:  commonerrors.IsAPIClientNil,
		},
		{
			name:         "error message set",
			builderError: errInvalidBuilder,
			assertError:  isInvalidBuilder,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder SB

			if !testCase.builderNil {
				client := clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: []clients.SchemeAttacher{config.SchemeAttacher},
				})

				if config.ResourceScope.IsNamespaced() {
					builder = common.NewNamespacedBuilder[O, B, SO, SB](
						client, config.SchemeAttacher, testResourceName, testResourceNamespace)
				} else {
					builder = common.NewClusterScopedBuilder[O, B, SO, SB](
						client, config.SchemeAttacher, testResourceName)
				}

				if testCase.definitionNil {
					builder.SetDefinition(nil)
				}

				if testCase.apiClientNil {
					builder.SetClient(nil)
				}

				if testCase.builderError != nil {
					builder.SetError(testCase.builderError)
				}
			}

			err := config.validateFunc(builder)

			assert.Truef(t, testCase.assertError(err), "got error %v", err)
		})
	}
}
