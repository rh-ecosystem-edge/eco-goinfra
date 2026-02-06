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

// Updater is an interface for builders that have an Update method without a force parameter.
type Updater[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Update() (SB, error)
}

// ForceUpdater is an interface for builders that have an Update method with a force parameter.
type ForceUpdater[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Update(force bool) (SB, error)
}

// internalUpdateFunc is the internal function signature used by UpdateTestConfig. All of the other update functions
// must be able to be wrapped in this signature.
//
// This type is different than the [GenericUpdateFunc] because it makes stricter assumptions about the builder type that
// the common package does not. The constructor for the generic version enforces these constraints so they can be made
// equivalent with a thin wrapper.
type internalUpdateFunc[
	O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(ctx context.Context, builder SB, force bool) (SB, error)

// GenericUpdateFunc is the signature for the common.Update function that takes context and builder.
type GenericUpdateFunc[O any, SO common.ObjectPointer[O]] func(ctx context.Context, builder common.Builder[O, SO], force bool) error

// UpdateTestConfig provides the configuration needed to test an Update method without force.
type UpdateTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// updateFunc is a function that updates the resource and returns an error. It gets set by the constructor
	// methods and will handle the different signatures of the Update method.
	updateFunc internalUpdateFunc[O, B, SO, SB]
	// alsoRunForceTests is a flag that indicates if force update tests should also be run. It gets set by the
	// constructor methods.
	alsoRunForceTests bool
}

// NewUpdateTestConfig creates a new UpdateTestConfig with the given parameters for non-force updates. Force update
// tests will not be run.
func NewUpdateTestConfig[O, B any, SO common.ObjectPointer[O], SB Updater[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) UpdateTestConfig[O, B, SO, SB] {
	return UpdateTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		updateFunc: func(_ context.Context, builder SB, force bool) (SB, error) {
			return builder.Update()
		},
		alsoRunForceTests: false,
	}
}

// NewForceUpdateTestConfig creates a new UpdateTestConfig with the given parameters for force updates. When executing
// tests, additional tests will be run for force updates.
func NewForceUpdateTestConfig[O, B any, SO common.ObjectPointer[O], SB ForceUpdater[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) UpdateTestConfig[O, B, SO, SB] {
	return UpdateTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		updateFunc: func(_ context.Context, builder SB, force bool) (SB, error) {
			return builder.Update(force)
		},
		alsoRunForceTests: true,
	}
}

// NewGenericUpdateTestConfig creates a new UpdateTestConfig with a custom update function. This is useful for testing
// standalone functions like common.Update() rather than builder methods.
func NewGenericUpdateTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	updateFunc GenericUpdateFunc[O, SO],
	alsoRunForceTests bool,
) UpdateTestConfig[O, B, SO, SB] {
	return UpdateTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		updateFunc: func(ctx context.Context, builder SB, force bool) (SB, error) {
			return builder, updateFunc(ctx, builder, force)
		},
		alsoRunForceTests: alsoRunForceTests,
	}
}

// Name returns the name to use for running these tests.
func (config UpdateTestConfig[O, B, SO, SB]) Name() string {
	return "Update"
}

// ExecuteTests runs the standard set of Update tests (non-force) for the configured resource.
func (config UpdateTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	config.executeNonForceTests(t)

	if config.alsoRunForceTests {
		config.executeForceTests(t)
	}
}

func (config UpdateTestConfig[O, B, SO, SB]) executeNonForceTests(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name             string
		objectExists     bool
		builderError     error
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
	}{
		{
			name:         "valid update existing resource",
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
			assertError:  isAPICallFailedWithGet,
		},
		{
			name:             "failed update returns error",
			objectExists:     true,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate},
			assertError:      isAPICallFailedWithUpdate,
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
			builder.GetDefinition().SetAnnotations(map[string]string{testAnnotationKey: testAnnotationValue})

			result, err := config.updateFunc(t.Context(), builder, false)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				require.NotNil(t, result)
				require.NotNil(t, result.GetObject())
				assert.Equal(t, testResourceName, result.GetObject().GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testResourceNamespace, result.GetObject().GetNamespace())
				}

				assert.Equal(t, testAnnotationValue, result.GetObject().GetAnnotations()[testAnnotationKey])
			}
		})
	}
}

func (config UpdateTestConfig[O, B, SO, SB]) executeForceTests(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name             string
		objectExists     bool
		builderError     error
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
	}{
		{
			name:             "force update succeeds via delete and create when update fails",
			objectExists:     true,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate},
			assertError:      isErrorNil,
		},
		{
			name:             "force update fails if delete fails",
			objectExists:     true,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate, Delete: testFailingDelete},
			assertError:      isAPICallFailedWithDelete,
		},
		{
			name:             "force update fails if create fails after delete",
			objectExists:     true,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate, Create: testFailingCreate},
			assertError:      isAPICallFailedWithCreate,
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

			builder.GetDefinition().SetAnnotations(map[string]string{testAnnotationKey: testAnnotationValue})

			result, err := config.updateFunc(t.Context(), builder, true)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				require.NotNil(t, result)
				require.NotNil(t, result.GetObject())
				assert.Equal(t, testResourceName, result.GetObject().GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testResourceNamespace, result.GetObject().GetNamespace())
				}

				assert.Equal(t, testAnnotationValue, result.GetObject().GetAnnotations()[testAnnotationKey])
			}
		})
	}
}
