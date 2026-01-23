package testhelper

import (
	"context"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

// Exister is an interface for builders that have an Exists method.
type Exister[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Exists() bool
}

// internalExistsFunc defines the signature for exists operations.
type internalExistsFunc[
	O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(ctx context.Context, builder SB) bool

// GenericExistsFunc is the signature for the common.Exists function that takes context and builder.
type GenericExistsFunc[O any, SO common.ObjectPointer[O]] func(ctx context.Context, builder common.Builder[O, SO]) bool

// ExistsTestConfig provides the configuration needed to test an Exists method.
type ExistsTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// existsFunc is a function that checks if the resource exists and returns a boolean.
	existsFunc internalExistsFunc[O, B, SO, SB]
}

// NewExistsTestConfig creates a new ExistsTestConfig with the given parameters for builders that implement the Exister
// interface.
func NewExistsTestConfig[O, B any, SO common.ObjectPointer[O], SB Exister[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) ExistsTestConfig[O, B, SO, SB] {
	return ExistsTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		existsFunc: func(_ context.Context, builder SB) bool {
			return builder.Exists()
		},
	}
}

// NewGenericExistsTestConfig creates a new ExistsTestConfig with a custom exists function. This is useful for testing
// standalone functions like common.Exists() rather than builder methods.
func NewGenericExistsTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	existsFunc GenericExistsFunc[O, SO],
) ExistsTestConfig[O, B, SO, SB] {
	return ExistsTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		existsFunc: func(ctx context.Context, builder SB) bool {
			return existsFunc(ctx, builder)
		},
	}
}

// Name returns the name to use for running these tests.
func (config ExistsTestConfig[O, B, SO, SB]) Name() string {
	return "Exists"
}

// ExecuteTests runs the standard set of Exists tests for the configured resource.
func (config ExistsTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	testCases := []struct {
		name           string
		objectExists   bool
		builderError   error
		expectedResult bool
	}{
		{
			name:           "valid exists returns true when resource exists",
			objectExists:   true,
			expectedResult: true,
		},
		{
			name:           "invalid builder returns false",
			objectExists:   true,
			builderError:   errInvalidBuilder,
			expectedResult: false,
		},
		{
			name:           "resource does not exist returns false",
			objectExists:   false,
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object

			if testCase.objectExists {
				var namespace string
				if config.ResourceScope.IsNamespaced() {
					namespace = testResourceNamespace
				}

				objects = append(objects, buildDummyObject[O, SO](testResourceName, namespace))
			}

			client := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  objects,
				SchemeAttachers: []clients.SchemeAttacher{config.SchemeAttacher},
			})

			var builder SB
			if config.ResourceScope.IsNamespaced() {
				builder = common.NewNamespacedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName, testResourceNamespace)
			} else {
				builder = common.NewClusterScopedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName)
			}

			builder.SetError(testCase.builderError)

			result := config.existsFunc(t.Context(), builder)

			assert.Equal(t, testCase.expectedResult, result)

			if testCase.expectedResult {
				require.NotNil(t, builder.GetObject())
				assert.Equal(t, testResourceName, builder.GetObject().GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testResourceNamespace, builder.GetObject().GetNamespace())
				}
			}
		})
	}
}
