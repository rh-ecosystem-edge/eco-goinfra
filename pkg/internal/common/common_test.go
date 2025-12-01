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
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
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

func TestCreate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		builderValid     bool
		objectExists     bool
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
	}{
		{
			name:         "valid create new resource",
			builderValid: true,
			objectExists: false,
			assertError:  isErrorNil,
		},
		{
			name:         "invalid builder",
			builderValid: false,
			objectExists: false,
			assertError:  isInvalidBuilder,
		},
		{
			name:         "resource already exists",
			builderValid: true,
			objectExists: true,
			assertError:  isErrorNil,
		},
		{
			name:             "failed creation",
			builderValid:     true,
			objectExists:     false,
			interceptorFuncs: interceptor.Funcs{Create: testFailingCreate},
			assertError:      isAPICallFailedWithVerb("create"),
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

			err := Create(t.Context(), builder)

			assert.Truef(t, testCase.assertError(err), "got error %v", err)

			if err == nil {
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
		assertError      func(error) bool
	}{
		{
			name:         "valid update existing resource",
			builderValid: true,
			objectExists: true,
			force:        false,
			assertError:  isErrorNil,
		},
		{
			name:         "invalid builder",
			builderValid: false,
			objectExists: false,
			force:        false,
			assertError:  isInvalidBuilder,
		},
		{
			name:         "resource does not exist",
			builderValid: true,
			objectExists: false,
			force:        false,
			assertError:  k8serrors.IsNotFound,
		},
		{
			name:         "valid force update existing resource",
			builderValid: true,
			objectExists: true,
			force:        true,
			assertError:  isErrorNil,
		},
		{
			name:             "force update with initial error",
			builderValid:     true,
			objectExists:     true,
			force:            true,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate},
			assertError:      isErrorNil,
		},
		{
			name:             "non-force update with error should fail",
			builderValid:     true,
			objectExists:     true,
			force:            false,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate},
			assertError:      isAPICallFailedWithVerb("update"),
		},
		{
			name:             "force update with delete failure",
			builderValid:     true,
			objectExists:     true,
			force:            true,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate, Delete: testFailingDelete},
			assertError:      isAPICallFailedWithVerb("delete"),
		},
		{
			name:             "force update with create failure",
			builderValid:     true,
			objectExists:     true,
			force:            true,
			interceptorFuncs: interceptor.Funcs{Update: testFailingUpdate, Create: testFailingCreate},
			assertError:      isAPICallFailedWithVerb("create"),
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

			err := Update(t.Context(), builder, testCase.force)

			assert.Truef(t, testCase.assertError(err), "got error %v", err)

			if err == nil {
				assert.NotNil(t, builder.GetObject())
				assert.Equal(t, defaultName, builder.GetObject().GetName())
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
		assertError      func(error) bool
	}{
		{
			name:         "valid delete existing resource",
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
			assertError:  isErrorNil,
		},
		{
			name:             "failed deletion",
			builderValid:     true,
			objectExists:     true,
			interceptorFuncs: interceptor.Funcs{Delete: testFailingDelete},
			assertError:      isAPICallFailedWithVerb("delete"),
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

			err := Delete(t.Context(), builder)

			assert.Truef(t, testCase.assertError(err), "got error %v", err)

			if err == nil {
				assert.Nil(t, builder.GetObject())
			}
		})
	}
}

//nolint:funlen // This function is only long because of the number of test cases.
func TestList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		clientNil        bool
		schemeAttacher   clients.SchemeAttacher
		objectsExist     bool
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
		expectedCount    int
	}{
		{
			name:           "valid list with resources",
			clientNil:      false,
			schemeAttacher: testSchemeAttacher,
			objectsExist:   true,
			assertError:    isErrorNil,
			expectedCount:  2,
		},
		{
			name:           "valid list empty",
			clientNil:      false,
			schemeAttacher: testSchemeAttacher,
			objectsExist:   false,
			assertError:    isErrorNil,
			expectedCount:  0,
		},
		{
			name:           "nil client",
			clientNil:      true,
			schemeAttacher: testSchemeAttacher,
			objectsExist:   false,
			assertError:    errors.IsAPIClientNil,
			expectedCount:  0,
		},
		{
			name:           "scheme attachment failure",
			clientNil:      false,
			schemeAttacher: testFailingSchemeAttacher,
			objectsExist:   false,
			assertError:    errors.IsSchemeAttacherFailed,
			expectedCount:  0,
		},
		{
			name:             "failed list call",
			clientNil:        false,
			schemeAttacher:   testSchemeAttacher,
			objectsExist:     false,
			interceptorFuncs: interceptor.Funcs{List: testFailingList},
			assertError:      isAPICallFailedWithVerb("list"),
			expectedCount:    0,
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
				if testCase.objectsExist {
					objects = append(objects,
						buildDummyNamespacedResource("resource-1", defaultNamespace),
						buildDummyNamespacedResource("resource-2", defaultNamespace),
					)
				}

				client = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:   objects,
					SchemeAttachers:  []clients.SchemeAttacher{testSchemeAttacher},
					InterceptorFuncs: testCase.interceptorFuncs,
				})
			}

			builders, err := List[corev1.ConfigMap, corev1.ConfigMapList, mockNamespacedBuilder](
				t.Context(), client, testCase.schemeAttacher)

			assert.Truef(t, testCase.assertError(err), "got error %v", err)

			if err == nil {
				assert.Len(t, builders, testCase.expectedCount)

				for _, builder := range builders {
					assert.NotNil(t, builder.GetDefinition())
					assert.NotNil(t, builder.GetObject())
					assert.NotNil(t, builder.GetClient())
				}
			} else {
				assert.Empty(t, builders)
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
