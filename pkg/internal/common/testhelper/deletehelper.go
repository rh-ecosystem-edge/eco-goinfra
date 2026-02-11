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

// Deleter is an interface for builders that have a Delete method returning only an error.
type Deleter[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Delete() error
}

// DeleteReturner is an interface for builders that have a Delete method returning the builder and an error.
type DeleteReturner[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Delete() (SB, error)
}

// internalDeleteFunc is the internal function signature used by DeleteTestConfig. All of the other delete functions
// must be able to be wrapped in this signature.
//
// This type is different than the [GenericDeleteFunc] because it makes stricter assumptions about the builder type that
// the common package does not. The constructor for the generic version enforces these constraints so they can be made
// equivalent with a thin wrapper.
type internalDeleteFunc[
	O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(ctx context.Context, builder SB) error

// GenericDeleteFunc is the signature for the common.Delete function that takes context and builder.
type GenericDeleteFunc[O any, SO common.ObjectPointer[O]] func(ctx context.Context, builder common.Builder[O, SO]) error

// DeleteTestConfig provides the configuration needed to test a Delete method.
type DeleteTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// deleteFunc is a function that deletes the resource and returns an error. It gets set by the constructor
	// methods and will handle the different signatures of the Delete method.
	deleteFunc internalDeleteFunc[O, B, SO, SB]
}

// NewDeleterTestConfig creates a new DeleteTestConfig for builders that implement the Deleter interface.
func NewDeleterTestConfig[O, B any, SO common.ObjectPointer[O], SB Deleter[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) DeleteTestConfig[O, B, SO, SB] {
	return DeleteTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		deleteFunc: func(_ context.Context, builder SB) error {
			return builder.Delete()
		},
	}
}

// NewDeleteReturnerTestConfig creates a new DeleteTestConfig for builders that implement the DeleteReturner interface.
func NewDeleteReturnerTestConfig[O, B any, SO common.ObjectPointer[O], SB DeleteReturner[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) DeleteTestConfig[O, B, SO, SB] {
	return DeleteTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		deleteFunc: func(_ context.Context, builder SB) error {
			_, err := builder.Delete()

			return err
		},
	}
}

// NewGenericDeleteTestConfig creates a new DeleteTestConfig with a custom delete function. This is useful for testing
// standalone functions like common.Delete() rather than builder methods.
func NewGenericDeleteTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	deleteFunc GenericDeleteFunc[O, SO],
) DeleteTestConfig[O, B, SO, SB] {
	return DeleteTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		deleteFunc: func(ctx context.Context, builder SB) error {
			return deleteFunc(ctx, builder)
		},
	}
}

// Name returns the name to use for running these tests.
func (config DeleteTestConfig[O, B, SO, SB]) Name() string {
	return "Delete"
}

// ExecuteTests runs the standard set of Delete tests for the configured resource.
func (config DeleteTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	testCases := []struct {
		name             string
		objectExists     bool
		builderError     error
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
	}{
		{
			name:         "valid delete existing resource",
			objectExists: true,
			assertError:  isErrorNil,
		},
		{
			name:         "invalid builder returns error",
			objectExists: true,
			builderError: errInvalidBuilder,
			assertError:  isInvalidBuilder,
		},
		{
			name:         "resource does not exist succeeds",
			objectExists: false,
			assertError:  isErrorNil,
		},
		{
			name:             "failed deletion returns error",
			objectExists:     true,
			interceptorFuncs: interceptor.Funcs{Delete: testFailingDelete},
			assertError:      isAPICallFailedWithDelete,
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

			err := config.deleteFunc(t.Context(), builder)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				assert.Nil(t, builder.GetObject())
			}
		})
	}
}
