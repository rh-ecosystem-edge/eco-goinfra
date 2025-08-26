package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestNewClusterScopedBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		clientNil      bool
		builderName    string
		schemeAttacher clients.SchemeAttacher
		expectedNil    bool
		expectedError  string
	}{
		{
			name:           "valid builder creation",
			clientNil:      false,
			builderName:    defaultName,
			schemeAttacher: testSchemeAttacher,
			expectedNil:    false,
			expectedError:  "",
		},
		{
			name:           "nil client",
			clientNil:      true,
			builderName:    defaultName,
			schemeAttacher: testSchemeAttacher,
			expectedNil:    true,
			expectedError:  "",
		},
		{
			name:           "empty name",
			clientNil:      false,
			builderName:    "",
			schemeAttacher: testSchemeAttacher,
			expectedNil:    false,
			expectedError:  getBuilderNameEmptyError(clusterScopedGVK.Kind).Error(),
		},
		{
			name:           "scheme attachment failure",
			clientNil:      false,
			builderName:    defaultName,
			schemeAttacher: testFailingSchemeAttacher,
			expectedNil:    true,
			expectedError:  "",
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

			if testCase.expectedNil {
				assert.Nil(t, builder)

				return
			}

			assert.NotNil(t, builder)
			assert.Equal(t, testCase.expectedError, builder.GetErrorMessage())

			if testCase.expectedError == "" {
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
		expectedNil    bool
		expectedError  string
	}{
		{
			name:           "valid builder creation",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			expectedNil:    false,
			expectedError:  "",
		},
		{
			name:           "nil client",
			clientNil:      true,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			expectedNil:    true,
			expectedError:  "",
		},
		{
			name:           "empty name",
			clientNil:      false,
			builderName:    "",
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			expectedNil:    false,
			expectedError:  getBuilderNameEmptyError(namespacedGVK.Kind).Error(),
		},
		{
			name:           "empty namespace",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  "",
			schemeAttacher: testSchemeAttacher,
			expectedNil:    false,
			expectedError:  getBuilderNamespaceEmptyError(namespacedGVK.Kind).Error(),
		},
		{
			name:           "scheme attachment failure",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testFailingSchemeAttacher,
			expectedNil:    true,
			expectedError:  "",
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

			if testCase.expectedNil {
				assert.Nil(t, builder)

				return
			}

			assert.NotNil(t, builder)
			assert.Equal(t, testCase.expectedError, builder.GetErrorMessage())

			if testCase.expectedError == "" {
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
		expectedError  error
	}{
		{
			name:           "valid pull existing resource",
			clientNil:      false,
			builderName:    defaultName,
			schemeAttacher: testSchemeAttacher,
			objectExists:   true,
			expectedError:  nil,
		},
		{
			name:           "nil client",
			clientNil:      true,
			builderName:    defaultName,
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			expectedError:  getAPIClientNilError(clusterScopedGVK.Kind),
		},
		{
			name:           "empty name",
			clientNil:      false,
			builderName:    "",
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			expectedError:  getBuilderNameEmptyError(clusterScopedGVK.Kind),
		},
		{
			name:           "scheme attachment failure",
			clientNil:      false,
			builderName:    defaultName,
			schemeAttacher: testFailingSchemeAttacher,
			objectExists:   false,
			expectedError:  wrapSchemeAttacherError(clusterScopedGVK.Kind, errSchemeAttachment),
		},
		{
			name:           "resource does not exist",
			clientNil:      false,
			builderName:    "non-existent-namespace",
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			expectedError:  getBuilderNotFoundError(clusterScopedGVK.Kind),
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
				client, testCase.schemeAttacher, testCase.builderName)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName())
			} else {
				assert.Nil(t, builder)
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
		expectedError  error
	}{
		{
			name:           "valid pull existing resource",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			objectExists:   true,
			expectedError:  nil,
		},
		{
			name:           "nil client",
			clientNil:      true,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			expectedError:  getAPIClientNilError(namespacedGVK.Kind),
		},
		{
			name:           "empty name",
			clientNil:      false,
			builderName:    "",
			builderNsName:  defaultNamespace,
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			expectedError:  getBuilderNameEmptyError(namespacedGVK.Kind),
		},
		{
			name:           "empty namespace",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  "",
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			expectedError:  getBuilderNamespaceEmptyError(namespacedGVK.Kind),
		},
		{
			name:           "scheme attachment failure",
			clientNil:      false,
			builderName:    defaultName,
			builderNsName:  defaultNamespace,
			schemeAttacher: testFailingSchemeAttacher,
			objectExists:   false,
			expectedError:  wrapSchemeAttacherError(namespacedGVK.Kind, errSchemeAttachment),
		},
		{
			name:           "resource does not exist",
			clientNil:      false,
			builderName:    "non-existent-resource",
			builderNsName:  "non-existent-namespace",
			schemeAttacher: testSchemeAttacher,
			objectExists:   false,
			expectedError:  getBuilderNotFoundError(namespacedGVK.Kind),
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
				client, testCase.schemeAttacher, testCase.builderName, testCase.builderNsName)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
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
		name          string
		builderValid  bool
		objectExists  bool
		expectedError string
	}{
		{
			name:          "valid get existing resource",
			builderValid:  true,
			objectExists:  true,
			expectedError: "",
		},
		{
			name:          "invalid builder",
			builderValid:  false,
			objectExists:  false,
			expectedError: errInvalidBuilder.Error(),
		},
		{
			name:         "resource does not exist",
			builderValid: true,
			objectExists: false,
			// Ideally we would reuse the wrap function here so the test is less coupled, but that function
			// requires the builder and specifics of the error being wrapped.
			expectedError: "failed to get the Namespace builder test-resource: namespaces \"test-resource\" not found",
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

			result, err := Get(builder)

			if testCase.expectedError == "" {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, defaultName, result.GetName())
			} else {
				assert.EqualError(t, err, testCase.expectedError)
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

			result := Exists(builder)
			assert.Equal(t, testCase.expectedResult, result)

			if testCase.expectedResult {
				assert.NotNil(t, builder.GetObject())
				assert.Equal(t, defaultName, builder.GetObject().GetName())
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		builderValid     bool
		objectExists     bool
		interceptorFuncs interceptor.Funcs
		expectedError    error
	}{
		{
			name:          "valid create new resource",
			builderValid:  true,
			objectExists:  false,
			expectedError: nil,
		},
		{
			name:          "invalid builder",
			builderValid:  false,
			objectExists:  false,
			expectedError: errInvalidBuilder,
		},
		{
			name:          "resource already exists",
			builderValid:  true,
			objectExists:  true,
			expectedError: nil,
		},
		{
			name:             "failed creation",
			builderValid:     true,
			objectExists:     false,
			interceptorFuncs: interceptor.Funcs{Create: testFailingCreate},
			expectedError:    fmt.Errorf("failed to create the Namespace builder test-resource: %w", errCreateFailure),
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
				K8sMockObjects:   objects,
				SchemeAttachers:  []clients.SchemeAttacher{testSchemeAttacher},
				InterceptorFuncs: testCase.interceptorFuncs,
			})

			builder := buildValidMockClusterScopedBuilder(client)
			if !testCase.builderValid {
				builder = buildInvalidMockClusterScopedBuilder(client)
			}

			err := Create(builder)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.NotNil(t, builder.GetObject())
				assert.Equal(t, defaultName, builder.GetObject().GetName())
			}
		})
	}
}

//nolint:funlen // This function is only long because of the number of test cases.
func TestUpdate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		builderValid     bool
		objectExists     bool
		force            bool
		interceptorFuncs interceptor.Funcs
		expectedError    error
	}{
		{
			name:          "valid update existing resource",
			builderValid:  true,
			objectExists:  true,
			force:         false,
			expectedError: nil,
		},
		{
			name:          "invalid builder",
			builderValid:  false,
			objectExists:  false,
			force:         false,
			expectedError: errInvalidBuilder,
		},
		{
			name:          "resource does not exist",
			builderValid:  true,
			objectExists:  false,
			force:         false,
			expectedError: getBuilderNotFoundError(clusterScopedGVK.Kind),
		},
		{
			name:          "valid force update existing resource",
			builderValid:  true,
			objectExists:  true,
			force:         true,
			expectedError: nil,
		},
		{
			name:         "force update with initial update conflict",
			builderValid: true,
			objectExists: true,
			force:        true,
			interceptorFuncs: interceptor.Funcs{
				Update: testFailingUpdate,
			},
			expectedError: nil,
		},
		{
			name:             "non-force update with conflict should fail",
			builderValid:     true,
			objectExists:     true,
			force:            false,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate},
			expectedError:    fmt.Errorf("failed to update the Namespace builder test-resource: %w", errUpdateConflict),
		},
		{
			name:             "force update with delete failure",
			builderValid:     true,
			objectExists:     true,
			force:            true,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate, Delete: testFailingDelete},
			expectedError: fmt.Errorf("failed to delete the Namespace builder test-resource during force update: "+
				"failed to delete the Namespace builder test-resource: %w", errDeleteFailure),
		},
		{
			name:             "force update with create failure",
			builderValid:     true,
			objectExists:     true,
			force:            true,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate, Create: testFailingCreate},
			expectedError: fmt.Errorf("failed to recreate the Namespace builder test-resource during force update: "+
				"failed to create the Namespace builder test-resource: %w", errCreateFailure),
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
				K8sMockObjects:   objects,
				SchemeAttachers:  []clients.SchemeAttacher{testSchemeAttacher},
				InterceptorFuncs: testCase.interceptorFuncs,
			})

			builder := buildValidMockClusterScopedBuilder(client)
			if !testCase.builderValid {
				builder = buildInvalidMockClusterScopedBuilder(client)
			}

			err := Update(builder, testCase.force)
			if testCase.expectedError == nil {
				assert.NoError(t, err)
			} else {
				// The expected error for the force update failures will not match because of how it is
				// wrapped, so we instead compare the error text.
				assert.EqualError(t, err, testCase.expectedError.Error())
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		builderValid     bool
		objectExists     bool
		interceptorFuncs interceptor.Funcs
		expectedError    error
	}{
		{
			name:          "valid delete existing resource",
			builderValid:  true,
			objectExists:  true,
			expectedError: nil,
		},
		{
			name:          "invalid builder",
			builderValid:  false,
			objectExists:  false,
			expectedError: errInvalidBuilder,
		},
		{
			name:          "resource does not exist",
			builderValid:  true,
			objectExists:  false,
			expectedError: nil,
		},
		{
			name:             "failed deletion",
			builderValid:     true,
			objectExists:     true,
			interceptorFuncs: interceptor.Funcs{Delete: testFailingDelete},
			expectedError:    fmt.Errorf("failed to delete the Namespace builder test-resource: %w", errDeleteFailure),
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
				K8sMockObjects:   objects,
				SchemeAttachers:  []clients.SchemeAttacher{testSchemeAttacher},
				InterceptorFuncs: testCase.interceptorFuncs,
			})

			builder := buildValidMockClusterScopedBuilder(client)
			if !testCase.builderValid {
				builder = buildInvalidMockClusterScopedBuilder(client)
			}

			err := Delete(builder)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Nil(t, builder.GetObject())
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		builderNil      bool
		definitionNil   bool
		apiClientNil    bool
		builderErrorMsg string
		expectedError   error
	}{
		{
			name:            "valid builder",
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   nil,
		},
		{
			name:            "nil builder",
			builderNil:      true,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   getBuilderUninitializedError(),
		},
		{
			name:            "nil definition",
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   getBuilderDefinitionNilError(clusterScopedGVK.Kind),
		},
		{
			name:            "nil apiClient",
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   getBuilderAPIClientNilError(clusterScopedGVK.Kind),
		},
		{
			name:            "error message set",
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: defaultErrorMessage,
			expectedError:   errors.New(defaultErrorMessage),
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

				if testCase.builderErrorMsg != "" {
					builder.SetErrorMessage(testCase.builderErrorMsg)
				}
			}

			err := Validate(builder)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}
