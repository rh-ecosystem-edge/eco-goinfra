package testhelper

import (
	"context"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// ListInAllNamespacesFunc is a List function signature for listing resources in all namespaces (e.g.,
// ListInAllNamespaces).
type ListInAllNamespacesFunc[SB any] func(
	apiClient *clients.Settings,
	options ...runtimeclient.ListOptions,
) ([]SB, error)

// NamespacedListFunc is a List function signature for listing resources in a specific namespace (e.g., List with
// namespace parameter).
type NamespacedListFunc[SB any] func(
	apiClient *clients.Settings,
	nsname string,
	options ...runtimeclient.ListOptions,
) ([]SB, error)

// GenericListFunc is the signature for the common.List function that takes context, client, and scheme attacher.
type GenericListFunc[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(
	ctx context.Context,
	apiClient runtimeclient.Client,
	schemeAttacher clients.SchemeAttacher,
	options ...runtimeclient.ListOption,
) ([]SB, error)

// internalListFunc is the unified internal function signature used by ListTestConfig. All of the other list functions
// must be able to be wrapped in this signature.
//
// We use the clients.Settings type instead of the runtimeclient.Client type because clients.Settings may be used in
// place of runtimeclient.Client, but not vice versa.
type internalListFunc[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(
	ctx context.Context,
	apiClient *clients.Settings,
	schemeAttacher clients.SchemeAttacher,
	nsname string,
) ([]SB, error)

// ListTestConfig provides the configuration needed to test a List function.
type ListTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// listFunc is a unified function that wraps the actual list function being tested.
	listFunc internalListFunc[O, B, SO, SB]

	// testSchemeAttacher indicates whether scheme attacher failures should be tested. This is true for generic
	// (common.List) tests and false for wrapper function tests.
	testSchemeAttacher bool

	// testEmptyNamespace indicates whether empty namespace validation should be tested. This is true for wrapper
	// functions that take a namespace parameter and false for generic (common.List) tests that don't have namespace
	// validation.
	testEmptyNamespace bool
}

// NewListTestConfig creates a new ListTestConfig for ListInAllNamespaces-style functions.
func NewListTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	listFunc ListInAllNamespacesFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) ListTestConfig[O, B, SO, SB] {
	return ListTestConfig[O, B, SO, SB]{
		CommonTestConfig: CommonTestConfig[O, B, SO, SB]{
			SchemeAttacher: schemeAttacher,
			ExpectedGVK:    expectedGVK,
		},
		listFunc: func(_ context.Context, apiClient *clients.Settings, _ clients.SchemeAttacher, _ string) ([]SB, error) {
			return listFunc(apiClient)
		},
		testSchemeAttacher: false,
	}
}

// NewNamespacedListTestConfig creates a new ListTestConfig for namespaced List functions.
func NewNamespacedListTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	listFunc NamespacedListFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) ListTestConfig[O, B, SO, SB] {
	return ListTestConfig[O, B, SO, SB]{
		CommonTestConfig: CommonTestConfig[O, B, SO, SB]{
			SchemeAttacher: schemeAttacher,
			ExpectedGVK:    expectedGVK,
		},
		listFunc: func(_ context.Context, apiClient *clients.Settings, _ clients.SchemeAttacher, nsname string) ([]SB, error) {
			return listFunc(apiClient, nsname)
		},
		testSchemeAttacher: false,
		testEmptyNamespace: true,
	}
}

// NewGenericListTestConfig creates a new ListTestConfig with a custom list function. This is useful for testing
// standalone functions like common.List() rather than wrapper functions.
func NewGenericListTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	listFunc GenericListFunc[O, B, SO, SB],
) ListTestConfig[O, B, SO, SB] {
	return ListTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		listFunc: func(ctx context.Context, apiClient *clients.Settings, schemeAttacher clients.SchemeAttacher, _ string) ([]SB, error) {
			return listFunc(ctx, apiClient, schemeAttacher)
		},
		testSchemeAttacher: true,
		testEmptyNamespace: false,
	}
}

// Name returns the name to use for running these tests.
func (config ListTestConfig[O, B, SO, SB]) Name() string {
	return "List"
}

// ExecuteTests runs the standard set of tests for a List function.
//
//nolint:funlen // Test function with multiple test cases.
func (config ListTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	type testCase struct {
		name             string
		clientNil        bool
		schemeAttacher   clients.SchemeAttacher
		nsname           string
		objectsExist     bool
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
		expectedCount    int
	}

	testCases := []testCase{
		{
			name:           "valid list with resources",
			clientNil:      false,
			schemeAttacher: config.SchemeAttacher,
			nsname:         testResourceNamespace,
			objectsExist:   true,
			assertError:    isErrorNil,
			expectedCount:  2,
		},
		{
			name:           "valid list empty",
			clientNil:      false,
			schemeAttacher: config.SchemeAttacher,
			nsname:         testResourceNamespace,
			objectsExist:   false,
			assertError:    isErrorNil,
			expectedCount:  0,
		},
		{
			name:           "nil client returns error",
			clientNil:      true,
			schemeAttacher: config.SchemeAttacher,
			nsname:         testResourceNamespace,
			objectsExist:   false,
			assertError:    commonerrors.IsAPIClientNil,
			expectedCount:  0,
		},
		{
			name:             "failed list call returns error",
			clientNil:        false,
			schemeAttacher:   config.SchemeAttacher,
			nsname:           testResourceNamespace,
			objectsExist:     false,
			interceptorFuncs: interceptor.Funcs{List: testFailingList},
			assertError:      isAPICallFailedWithList,
			expectedCount:    0,
		},
	}

	if config.testSchemeAttacher {
		testCases = append(testCases, testCase{
			name:           "scheme attachment failure returns error",
			clientNil:      false,
			schemeAttacher: testFailingSchemeAttacher,
			nsname:         testResourceNamespace,
			objectsExist:   false,
			assertError:    commonerrors.IsSchemeAttacherFailed,
			expectedCount:  0,
		})
	}

	if config.testEmptyNamespace {
		testCases = append(testCases, testCase{
			name:           "empty namespace returns error",
			clientNil:      false,
			schemeAttacher: config.SchemeAttacher,
			nsname:         "",
			objectsExist:   false,
			assertError:    commonerrors.IsBuilderNamespaceEmpty,
			expectedCount:  0,
		})
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				client  *clients.Settings
				objects []runtime.Object
			)

			if !testCase.clientNil {
				if testCase.objectsExist {
					// We always use the namespace here since it is not important for the test. We
					// care about whether the List function takes a namespace parameter, not whether
					// the objects are in a namespace.
					objects = append(objects,
						buildDummyObject[O, SO]("resource-1", testResourceNamespace),
						buildDummyObject[O, SO]("resource-2", testResourceNamespace),
					)
				}

				client = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:   objects,
					SchemeAttachers:  []clients.SchemeAttacher{config.SchemeAttacher},
					InterceptorFuncs: testCase.interceptorFuncs,
				})
			}

			builders, err := config.listFunc(t.Context(), client, testCase.schemeAttacher, testCase.nsname)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				assert.Len(t, builders, testCase.expectedCount)

				for _, builder := range builders {
					require.NotNil(t, builder.GetDefinition())
					require.NotNil(t, builder.GetObject())
					require.NotNil(t, builder.GetClient())
					assert.Equal(t, config.ExpectedGVK, builder.GetGVK())
				}
			} else {
				assert.Empty(t, builders)
			}
		})
	}
}
