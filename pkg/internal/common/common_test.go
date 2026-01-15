package common

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestNewClusterScopedBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		clientNil      bool
		builderName    string
		schemeAttacher clients.SchemeAttacher
		assertError    func(error) bool
	}{
		{
			name:           "valid builder creation",
			clientNil:      false,
			builderName:    defaultName,
			schemeAttacher: testSchemeAttacher,
			assertError:    isErrorNil,
		},
		{
			name:           "nil client",
			clientNil:      true,
			builderName:    defaultName,
			schemeAttacher: testSchemeAttacher,
			assertError:    errors.IsAPIClientNil,
		},
		{
			name:           "empty name",
			clientNil:      false,
			builderName:    "",
			schemeAttacher: testSchemeAttacher,
			assertError:    errors.IsBuilderNameEmpty,
		},
		{
			name:           "scheme attachment failure",
			clientNil:      false,
			builderName:    defaultName,
			schemeAttacher: testFailingSchemeAttacher,
			assertError:    errors.IsSchemeAttacherFailed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var client runtimeclient.Client
			if !testCase.clientNil {
				client = clients.GetTestClients(clients.TestClientParams{})
			}

			builder := NewClusterScopedBuilder[corev1.Namespace, mockClusterScopedBuilder](
				client, testCase.schemeAttacher, testCase.builderName)

			assert.NotNil(t, builder)
			assert.Truef(t, testCase.assertError(builder.GetError()), "got error %v", builder.GetError())

			if builder.GetError() == nil {
				assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName())
			}
		})
	}
}

func TestNewNamespacedBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		clientNil      bool
		builderName    string
		builderNsName  string
		schemeAttacher clients.SchemeAttacher
		assertError    func(error) bool
	}{
		{
			name:           "valid builder creation",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			assertError:    isErrorNil,
		},
		{
			name:           "nil client",
			clientNil:      true,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			assertError:    errors.IsAPIClientNil,
		},
		{
			name:           "empty name",
			clientNil:      false,
			builderName:    "",
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			assertError:    errors.IsBuilderNameEmpty,
		},
		{
			name:           "empty namespace",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  "",
			schemeAttacher: testSchemeAttacher,
			assertError:    errors.IsBuilderNamespaceEmpty,
		},
		{
			name:           "scheme attachment failure",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testFailingSchemeAttacher,
			assertError:    errors.IsSchemeAttacherFailed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var client runtimeclient.Client
			if !testCase.clientNil {
				client = clients.GetTestClients(clients.TestClientParams{})
			}

			builder := NewNamespacedBuilder[corev1.ConfigMap, mockNamespacedBuilder](
				client, testCase.schemeAttacher, testCase.builderName, testCase.builderNsName)

			assert.NotNil(t, builder)
			assert.Truef(t, testCase.assertError(builder.GetError()), "got error %v", builder.GetError())

			if builder.GetError() == nil {
				assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName())
				assert.Equal(t, testCase.builderNsName, builder.GetDefinition().GetNamespace())
			}
		})
	}
}

func TestPullClusterScopedBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		clientNil      bool
		builderName    string
		schemeAttacher clients.SchemeAttacher
		objectExists   bool
		assertError    func(error) bool
	}{
		{
			name:           "valid pull existing resource",
			clientNil:      false,
			builderName:    defaultName,
			schemeAttacher: testSchemeAttacher,
			objectExists:   true,
			assertError:    isErrorNil,
		},
		{
			name:           "nil client",
			clientNil:      true,
			builderName:    defaultName,
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			assertError:    errors.IsAPIClientNil,
		},
		{
			name:           "empty name",
			clientNil:      false,
			builderName:    "",
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			assertError:    errors.IsBuilderNameEmpty,
		},
		{
			name:           "scheme attachment failure",
			clientNil:      false,
			builderName:    defaultName,
			schemeAttacher: testFailingSchemeAttacher,
			objectExists:   false,
			assertError:    errors.IsSchemeAttacherFailed,
		},
		{
			name:           "resource does not exist",
			clientNil:      false,
			builderName:    "non-existent-namespace",
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			assertError:    k8serrors.IsNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				client  runtimeclient.Client
				objects []runtime.Object
			)

			if !testCase.clientNil {
				if testCase.objectExists {
					objects = append(objects, buildDummyClusterScopedResource())
				}

				client = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  objects,
					SchemeAttachers: []clients.SchemeAttacher{testSchemeAttacher},
				})
			}

			builder, err := PullClusterScopedBuilder[corev1.Namespace, mockClusterScopedBuilder](
				t.Context(), client, testCase.schemeAttacher, testCase.builderName)

			assert.Truef(t, testCase.assertError(err), "got error %v", err)

			if err == nil {
				assert.NotNil(t, builder)
				assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName())
			}
		})
	}
}

//nolint:funlen // This function is only long because of the number of test cases.
func TestPullNamespacedBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		clientNil      bool
		builderName    string
		builderNsName  string
		schemeAttacher clients.SchemeAttacher
		objectExists   bool
		assertError    func(error) bool
	}{
		{
			name:           "valid pull existing resource",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			objectExists:   true,
			assertError:    isErrorNil,
		},
		{
			name:           "nil client",
			clientNil:      true,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			assertError:    errors.IsAPIClientNil,
		},
		{
			name:           "empty name",
			clientNil:      false,
			builderName:    "",
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			assertError:    errors.IsBuilderNameEmpty,
		},
		{
			name:           "empty namespace",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  "",
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			assertError:    errors.IsBuilderNamespaceEmpty,
		},
		{
			name:           "scheme attachment failure",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testFailingSchemeAttacher,
			objectExists:   false,
			assertError:    errors.IsSchemeAttacherFailed,
		},
		{
			name:           "resource does not exist",
			clientNil:      false,
			builderName:    "non-existent-resource",
			builderNsName:  "non-existent-namespace",
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			assertError:    k8serrors.IsNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				client  runtimeclient.Client
				objects []runtime.Object
			)

			if !testCase.clientNil {
				if testCase.objectExists {
					objects = append(objects, buildDummyNamespacedResource(defaultName, defaultNamespace))
				}

				client = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  objects,
					SchemeAttachers: []clients.SchemeAttacher{testSchemeAttacher},
				})
			}

			builder, err := PullNamespacedBuilder[corev1.ConfigMap, mockNamespacedBuilder](
				t.Context(), client, testCase.schemeAttacher, testCase.builderName, testCase.builderNsName)

			assert.Truef(t, testCase.assertError(err), "got error %v", err)

			if err == nil {
				assert.NotNil(t, builder)
				assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName())
				assert.Equal(t, testCase.builderNsName, builder.GetDefinition().GetNamespace())
			} else {
				assert.Nil(t, builder)
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		builderValid bool
		objectExists bool
		assertError  func(error) bool
	}{
		{
			name:         "valid get existing resource",
			builderValid: true,
			objectExists: true,
			assertError:  isErrorNil,
		},
		{
			name:         "invalid builder",
			builderValid: false,
			objectExists: false,
			assertError:  isInvalidBuilder,
		},
		{
			name:         "resource does not exist",
			builderValid: true,
			objectExists: false,
			assertError:  k8serrors.IsNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.objectExists {
				objects = append(objects, buildDummyClusterScopedResource())
			}

			client := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  objects,
				SchemeAttachers: []clients.SchemeAttacher{testSchemeAttacher},
			})

			builder := buildValidMockClusterScopedBuilder(client)
			if !testCase.builderValid {
				builder = buildInvalidMockClusterScopedBuilder(client)
			}

			result, err := Get(t.Context(), builder)

			assert.Truef(t, testCase.assertError(err), "got error %v", err)

			if err == nil {
				assert.NotNil(t, result)
				assert.Equal(t, defaultName, result.GetName())
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builderValid   bool
		objectExists   bool
		expectedResult bool
	}{
		{
			name:           "valid exists existing resource",
			builderValid:   true,
			objectExists:   true,
			expectedResult: true,
		},
		{
			name:           "invalid builder",
			builderValid:   false,
			objectExists:   false,
			expectedResult: false,
		},
		{
			name:           "resource does not exist",
			builderValid:   true,
			objectExists:   false,
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.objectExists {
				objects = append(objects, buildDummyClusterScopedResource())
			}

			client := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  objects,
				SchemeAttachers: []clients.SchemeAttacher{testSchemeAttacher},
			})

			builder := buildValidMockClusterScopedBuilder(client)
			if !testCase.builderValid {
				builder = buildInvalidMockClusterScopedBuilder(client)
			}

			result := Exists(t.Context(), builder)
			assert.Equal(t, testCase.expectedResult, result)

			if testCase.expectedResult {
				assert.NotNil(t, builder.GetObject())
				assert.Equal(t, defaultName, builder.GetObject().GetName())
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		builderError  error
		assertError   func(error) bool
	}{
		{
			name:          "valid builder",
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			builderError:  nil,
			assertError:   isErrorNil,
		},
		{
			name:          "nil builder",
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			builderError:  nil,
			assertError:   errors.IsBuilderNil,
		},
		{
			name:          "nil definition",
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			builderError:  nil,
			assertError:   errors.IsBuilderDefinitionNil,
		},
		{
			name:          "nil apiClient",
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			builderError:  nil,
			assertError:   errors.IsAPIClientNil,
		},
		{
			name:          "error message set",
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			builderError:  errInvalidBuilder,
			assertError:   isInvalidBuilder,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *mockClusterScopedBuilder

			if !testCase.builderNil {
				builder = buildValidMockClusterScopedBuilder(clients.GetTestClients(clients.TestClientParams{}))

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

			err := Validate(builder)

			assert.Truef(t, testCase.assertError(err), "got error %v", err)
		})
	}
}
