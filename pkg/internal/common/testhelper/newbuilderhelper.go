package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// NewClusterScopedBuilderFunc is a NewClusterScopedBuilder function signature.
type NewClusterScopedBuilderFunc[SB any] func(apiClient *clients.Settings, name string) SB

// NewNamespacedBuilderFunc is a NewNamespacedBuilder function signature.
type NewNamespacedBuilderFunc[SB any] func(apiClient *clients.Settings, name, nsname string) SB

// GenericNewClusterScopedBuilderFunc is a generic new builder function signature for cluster-scoped resources.
type GenericNewClusterScopedBuilderFunc[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(
	apiClient runtimeclient.Client,
	schemeAttacher clients.SchemeAttacher,
	name string,
) SB

// GenericNewNamespacedBuilderFunc is a generic new builder function signature for namespaced resources.
type GenericNewNamespacedBuilderFunc[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(
	apiClient runtimeclient.Client,
	schemeAttacher clients.SchemeAttacher,
	name, nsname string,
) SB

// internalNewBuilderFunc is the unified internal function signature used by NewBuilderTestConfig. It should be possible
// to make a thin wrapper around every other NewBuilder function to have this signature. This allows for unifying the
// test cases across the different signatures.
//
// We use the clients.Settings type instead of the runtimeclient.Client type because clients.Settings may be used in
// place of runtimeclient.Client, but not vice versa.
type internalNewBuilderFunc[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(
	apiClient *clients.Settings,
	schemeAttacher clients.SchemeAttacher,
	name, nsname string,
) SB

// NewBuilderTestConfig provides the configuration needed to test NewClusterScopedBuilder or NewNamespacedBuilder
// functions.
type NewBuilderTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// newBuilderFunc is a unified function that wraps the actual new builder function being tested.
	newBuilderFunc internalNewBuilderFunc[O, B, SO, SB]

	// testSchemeAttacher indicates whether scheme attacher failures should be tested.
	testSchemeAttacher bool
}

// NewClusterScopedBuilderTestConfig creates a new NewBuilderTestConfig for cluster-scoped NewBuilder functions.
func NewClusterScopedBuilderTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	newBuilderFunc NewClusterScopedBuilderFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) NewBuilderTestConfig[O, B, SO, SB] {
	return NewBuilderTestConfig[O, B, SO, SB]{
		CommonTestConfig: CommonTestConfig[O, B, SO, SB]{
			SchemeAttacher: schemeAttacher,
			ExpectedGVK:    expectedGVK,
			ResourceScope:  ResourceScopeClusterScoped,
		},
		newBuilderFunc: func(apiClient *clients.Settings, _ clients.SchemeAttacher, name, _ string) SB {
			return newBuilderFunc(apiClient, name)
		},
		testSchemeAttacher: false,
	}
}

// NewNamespacedBuilderTestConfig creates a new NewBuilderTestConfig for namespaced NewBuilder functions.
func NewNamespacedBuilderTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	newBuilderFunc NewNamespacedBuilderFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) NewBuilderTestConfig[O, B, SO, SB] {
	return NewBuilderTestConfig[O, B, SO, SB]{
		CommonTestConfig: CommonTestConfig[O, B, SO, SB]{
			SchemeAttacher: schemeAttacher,
			ExpectedGVK:    expectedGVK,
			ResourceScope:  ResourceScopeNamespaced,
		},
		newBuilderFunc: func(apiClient *clients.Settings, _ clients.SchemeAttacher, name, nsname string) SB {
			return newBuilderFunc(apiClient, name, nsname)
		},
		testSchemeAttacher: false,
	}
}

// NewGenericClusterScopedBuilderTestConfig creates a new NewBuilderTestConfig for testing the generic
// common.NewClusterScopedBuilder function.
func NewGenericClusterScopedBuilderTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	newBuilderFunc GenericNewClusterScopedBuilderFunc[O, B, SO, SB],
) NewBuilderTestConfig[O, B, SO, SB] {
	return NewBuilderTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		newBuilderFunc: func(apiClient *clients.Settings, schemeAttacher clients.SchemeAttacher, name, _ string) SB {
			return newBuilderFunc(apiClient, schemeAttacher, name)
		},
		testSchemeAttacher: true,
	}
}

// NewGenericNamespacedBuilderTestConfig creates a new NewBuilderTestConfig for testing the generic
// common.NewNamespacedBuilder function.
func NewGenericNamespacedBuilderTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	newBuilderFunc GenericNewNamespacedBuilderFunc[O, B, SO, SB],
) NewBuilderTestConfig[O, B, SO, SB] {
	return NewBuilderTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		newBuilderFunc: func(apiClient *clients.Settings, schemeAttacher clients.SchemeAttacher, name, nsname string) SB {
			return newBuilderFunc(apiClient, schemeAttacher, name, nsname)
		},
		testSchemeAttacher: true,
	}
}

// Name returns the name to use for running these tests.
func (config NewBuilderTestConfig[O, B, SO, SB]) Name() string {
	return "NewBuilder"
}

// ExecuteTests runs the standard set of tests for a NewBuilder function.
//
//nolint:funlen // Test function with multiple test cases.
func (config NewBuilderTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	type testCase struct {
		name           string
		clientNil      bool
		builderName    string
		builderNsName  string
		schemeAttacher clients.SchemeAttacher
		assertError    func(error) bool
	}

	testCases := []testCase{
		{
			name:           "valid builder creation",
			clientNil:      false,
			builderName:    testResourceName,
			builderNsName:  testResourceNamespace,
			schemeAttacher: config.SchemeAttacher,
			assertError:    isErrorNil,
		},
		{
			name:           "nil client returns error",
			clientNil:      true,
			builderName:    testResourceName,
			builderNsName:  testResourceNamespace,
			schemeAttacher: config.SchemeAttacher,
			assertError:    commonerrors.IsAPIClientNil,
		},
		{
			name:           "empty name returns error",
			clientNil:      false,
			builderName:    "",
			builderNsName:  testResourceNamespace,
			schemeAttacher: config.SchemeAttacher,
			assertError:    commonerrors.IsBuilderNameEmpty,
		},
	}

	if config.ResourceScope.IsNamespaced() {
		testCases = append(testCases, testCase{
			name:           "empty namespace returns error",
			clientNil:      false,
			builderName:    testResourceName,
			builderNsName:  "",
			schemeAttacher: config.SchemeAttacher,
			assertError:    commonerrors.IsBuilderNamespaceEmpty,
		})
	}

	if config.testSchemeAttacher {
		testCases = append(testCases, testCase{
			name:           "scheme attachment failure returns error",
			clientNil:      false,
			builderName:    testResourceName,
			builderNsName:  testResourceNamespace,
			schemeAttacher: testFailingSchemeAttacher,
			assertError:    commonerrors.IsSchemeAttacherFailed,
		})
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var client *clients.Settings

			if !testCase.clientNil {
				client = clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: []clients.SchemeAttacher{config.SchemeAttacher},
				})
			}

			builder := config.newBuilderFunc(client, testCase.schemeAttacher, testCase.builderName, testCase.builderNsName)

			require.NotNil(t, builder)
			require.Truef(t, testCase.assertError(builder.GetError()), "unexpected error, got: %v", builder.GetError())

			if builder.GetError() == nil {
				require.NotNil(t, builder.GetDefinition())
				assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testCase.builderNsName, builder.GetDefinition().GetNamespace())
				}

				assert.Equal(t, config.ExpectedGVK, builder.GetGVK())
			}
		})
	}
}
