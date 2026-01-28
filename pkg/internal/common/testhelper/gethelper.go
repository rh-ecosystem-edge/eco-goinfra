package testhelper

import (
	"context"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// Getter is an interface for builders that have a Get method.
type Getter[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Get() (SO, error)
}

// GenericGetFunc is the signature for the common.Get function that takes context and builder.
type GenericGetFunc[O any, SO common.ObjectPointer[O]] func(ctx context.Context, builder common.Builder[O, SO]) (SO, error)

// internalGetFunc is the internal function signature used by GetTestConfig. All of the other get functions must be able
// to be wrapped in this signature.
//
// This type is different than the [GenericGetFunc] because it makes stricter assumptions about the builder type that
// the common package does not. The constructor for the generic version enforces these constraints so they can be made
// equivalent with a thin wrapper.
type internalGetFunc[
	O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(ctx context.Context, builder SB) (SO, error)

// GetTestConfig provides the configuration needed to test a Get method.
type GetTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// getFunc is a function that gets the resource and returns the object and an error.
	getFunc internalGetFunc[O, B, SO, SB]
}

// NewGetTestConfig creates a new GetTestConfig with the given parameters for builders that implement the Getter
// interface.
func NewGetTestConfig[O, B any, SO common.ObjectPointer[O], SB Getter[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) GetTestConfig[O, B, SO, SB] {
	return GetTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		getFunc: func(_ context.Context, builder SB) (SO, error) {
			return builder.Get()
		},
	}
}

// NewGenericGetTestConfig creates a new GetTestConfig with a custom get function. This is useful for testing
// standalone functions like common.Get() rather than builder methods.
func NewGenericGetTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	getFunc GenericGetFunc[O, SO],
) GetTestConfig[O, B, SO, SB] {
	return GetTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		getFunc: func(ctx context.Context, builder SB) (SO, error) {
			return getFunc(ctx, builder)
		},
	}
}

// Name returns the name to use for running these tests.
func (config GetTestConfig[O, B, SO, SB]) Name() string {
	return "Get"
}

// ExecuteTests runs the standard set of Get tests for the configured resource.
func (config GetTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
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
			name:         "valid get existing resource",
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
			name:         "resource does not exist returns not found",
			objectExists: false,
			assertError:  k8serrors.IsNotFound,
		},
		{
			name:             "get failure returns error",
			objectExists:     true,
			interceptorFuncs: interceptor.Funcs{Get: testFailingGet},
			assertError:      isAPICallFailedWithGet,
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

			result, err := config.getFunc(t.Context(), builder)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				require.NotNil(t, result)
				assert.Equal(t, testResourceName, result.GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testResourceNamespace, result.GetNamespace())
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
