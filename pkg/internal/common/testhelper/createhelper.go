package testhelper

import (
	"context"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// Creator is an interface for builders that have a Create method.
type Creator[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Create() (SB, error)
}

// internalCreateFunc is the internal function signature used by CreateTestConfig. All of the other create functions
// must be able to be wrapped in this signature.
//
// This type is different than the [GenericCreateFunc] because it makes stricter assumptions about the builder type that
// the common package does not. The constructor for the generic version enforces these constraints so they can be made
// equivalent with a thin wrapper.
type internalCreateFunc[
	O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(ctx context.Context, builder SB) (SB, error)

// GenericCreateFunc is the signature for the common.Create function that takes context and builder.
type GenericCreateFunc[O any, SO common.ObjectPointer[O]] func(ctx context.Context, builder common.Builder[O, SO]) error

// CreateTestConfig provides the configuration needed to test a Create method.
type CreateTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// createFunc is a function that creates the resource and returns the builder and an error.
	createFunc internalCreateFunc[O, B, SO, SB]
}

// NewCreateTestConfig creates a new CreateTestConfig with the given parameters for builders that implement the Creator
// interface.
func NewCreateTestConfig[O, B any, SO common.ObjectPointer[O], SB Creator[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) CreateTestConfig[O, B, SO, SB] {
	return CreateTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		createFunc: func(_ context.Context, builder SB) (SB, error) {
			return builder.Create()
		},
	}
}

// NewGenericCreateTestConfig creates a new CreateTestConfig with a custom create function. This is useful for testing
// standalone functions like common.Create() rather than builder methods.
func NewGenericCreateTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	createFunc GenericCreateFunc[O, SO],
) CreateTestConfig[O, B, SO, SB] {
	return CreateTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		createFunc: func(ctx context.Context, builder SB) (SB, error) {
			return builder, createFunc(ctx, builder)
		},
	}
}

// Name returns the name to use for running these tests.
func (config CreateTestConfig[O, B, SO, SB]) Name() string {
	return "Create"
}

// ExecuteTests runs the standard set of Create tests for the configured resource.
func (config CreateTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	testCases := []struct {
		name             string
		objectExists     bool
		builderError     error
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
		expectObjectSet  bool
	}{
		{
			name:            "valid create new resource",
			objectExists:    false,
			assertError:     isErrorNil,
			expectObjectSet: true,
		},
		{
			name:            "invalid builder returns error",
			objectExists:    false,
			builderError:    errInvalidBuilder,
			assertError:     isInvalidBuilder,
			expectObjectSet: false,
		},
		{
			name:            "resource already exists succeeds",
			objectExists:    true,
			assertError:     isErrorNil,
			expectObjectSet: false, // Create does not set Object when resource already exists
		},
		{
			name:             "failed creation returns error",
			objectExists:     false,
			interceptorFuncs: interceptor.Funcs{Create: testFailingCreate},
			assertError:      isAPICallFailedWithCreate,
			expectObjectSet:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object

			if testCase.objectExists {
				objects = append(objects, buildDummyObject[O, SO](testResourceName, testResourceNamespace))
			}

			client := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:   objects,
				SchemeAttachers:  []clients.SchemeAttacher{config.SchemeAttacher},
				InterceptorFuncs: testCase.interceptorFuncs,
			})

			var builder SB
			if config.ResourceScope.IsNamespaced() {
				builder = common.NewNamespacedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName, testResourceNamespace)
			} else {
				builder = common.NewClusterScopedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName)
			}

			builder.SetError(testCase.builderError)

			result, err := config.createFunc(t.Context(), builder)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				require.NotNil(t, result)
				require.NotNil(t, result.GetDefinition())

				assert.Equal(t, testResourceName, result.GetDefinition().GetName())
				assert.Equal(t, config.ExpectedGVK, result.GetGVK())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testResourceNamespace, result.GetDefinition().GetNamespace())
				}

				if testCase.expectObjectSet {
					require.NotNil(t, result.GetObject())
					assert.Equal(t, testResourceName, result.GetObject().GetName())

					if config.ResourceScope.IsNamespaced() {
						assert.Equal(t, testResourceNamespace, result.GetObject().GetNamespace())
					}
				}
			}
		})
	}
}
