package rbac

import (
	"errors"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	defaultRoleBindingName      = "rolebinding-test"
	defaultRoleBindingNamespace = "rolebinding-test-namespace"
	defaultRoleBindingRole      = "rolebinding-test-role"
)

var defaultRoleBindingSubject = rbacv1.Subject{
	Kind:     "ServiceAccount",
	Name:     "rolebinding-test-service-account",
	APIGroup: "",
}

func TestNewRoleBindingBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *RoleBindingBuilder {
				return NewRoleBindingBuilder(
					apiClient, name, nsname, defaultRoleBindingRole, defaultRoleBindingSubject)
			},
			rbacv1.AddToScheme,
			roleBindingGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		role          string
		subject       rbacv1.Subject
		expectedError error
	}{
		{
			name:    "valid role and subject",
			role:    defaultRoleBindingRole,
			subject: defaultRoleBindingSubject,
		},
		{
			name: "invalid subject kind",
			role: defaultRoleBindingRole,
			subject: rbacv1.Subject{
				Kind: "Invalid",
				Name: "subject",
			},
			expectedError: errors.New("rolebinding subject kind must be one of 'ServiceAccount', 'User', 'Group'"),
		},
		{
			name: "empty subject name",
			role: defaultRoleBindingRole,
			subject: rbacv1.Subject{
				Kind: "User",
			},
			expectedError: errors.New("rolebinding subject name cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewRoleBindingBuilder(
				buildRoleBindingTestClient(),
				defaultRoleBindingName,
				defaultRoleBindingNamespace,
				testCase.role,
				testCase.subject,
			)

			assert.Equal(t, testCase.expectedError, builder.GetError())

			if testCase.expectedError == nil {
				expectedRoleRef := rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Name:     testCase.role,
					Kind:     "Role",
				}

				assert.Equal(t, expectedRoleRef, builder.Definition.RoleRef)
				assert.Equal(t, []rbacv1.Subject{testCase.subject}, builder.Definition.Subjects)
			}
		})
	}
}

func TestPullRoleBinding(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullRoleBinding, rbacv1.AddToScheme, roleBindingGVK).ExecuteTests(t)
}

func TestRoleBindingBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newRoleBindingCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestRoleBindingWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newRoleBindingCommonTestConfig()).ExecuteTests(t)
}

func TestRoleBindingWithSubjects(t *testing.T) {
	t.Parallel()

	additionalSubject := rbacv1.Subject{
		Kind:     "Group",
		Name:     "rolebinding-test-group",
		APIGroup: "rbac.authorization.k8s.io",
	}

	testCases := []struct {
		name               string
		subjects           []rbacv1.Subject
		validBuilder       bool
		expectedError      error
		expectedSubjects   []rbacv1.Subject
		expectedBuilderErr func(error) bool
	}{
		{
			name:             "valid subjects",
			subjects:         []rbacv1.Subject{additionalSubject},
			validBuilder:     true,
			expectedSubjects: []rbacv1.Subject{defaultRoleBindingSubject, additionalSubject},
		},
		{
			name:             "empty subjects",
			subjects:         nil,
			validBuilder:     true,
			expectedError:    errors.New("cannot create rolebinding with empty subject"),
			expectedSubjects: []rbacv1.Subject{defaultRoleBindingSubject},
		},
		{
			name: "invalid subject kind",
			subjects: []rbacv1.Subject{{
				Kind: "Invalid",
				Name: "subject",
			}},
			validBuilder:     true,
			expectedError:    errors.New("rolebinding subject kind must be one of 'ServiceAccount', 'User', 'Group'"),
			expectedSubjects: []rbacv1.Subject{defaultRoleBindingSubject},
		},
		{
			name: "empty subject name",
			subjects: []rbacv1.Subject{{
				Kind: "User",
			}},
			validBuilder:     true,
			expectedError:    errors.New("rolebinding subject name cannot be empty"),
			expectedSubjects: []rbacv1.Subject{defaultRoleBindingSubject},
		},
		{
			name:               "invalid builder",
			subjects:           []rbacv1.Subject{additionalSubject},
			validBuilder:       false,
			expectedSubjects:   nil,
			expectedBuilderErr: commonerrors.IsBuilderNamespaceEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidRoleBindingTestBuilder(buildRoleBindingTestClient())
			if !testCase.validBuilder {
				builder = buildInvalidRoleBindingTestBuilder(buildRoleBindingTestClient())
			}

			result := builder.WithSubjects(testCase.subjects)

			if testCase.expectedBuilderErr != nil {
				require.Error(t, result.GetError())
				assert.True(t, testCase.expectedBuilderErr(result.GetError()))
			} else {
				assert.Equal(t, testCase.expectedError, result.GetError())
			}

			assert.Equal(t, testCase.expectedSubjects, result.Definition.Subjects)
		})
	}
}

func newRoleBindingCommonTestConfig() testhelper.CommonTestConfig[
	rbacv1.RoleBinding, RoleBindingBuilder, *rbacv1.RoleBinding, *RoleBindingBuilder] {
	return testhelper.NewCommonTestConfig[rbacv1.RoleBinding, RoleBindingBuilder](
		rbacv1.AddToScheme,
		roleBindingGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

func buildRoleBindingTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{rbacv1.AddToScheme},
	})
}

func buildValidRoleBindingTestBuilder(apiClient *clients.Settings) *RoleBindingBuilder {
	return NewRoleBindingBuilder(
		apiClient,
		defaultRoleBindingName,
		defaultRoleBindingNamespace,
		defaultRoleBindingRole,
		defaultRoleBindingSubject,
	)
}

func buildInvalidRoleBindingTestBuilder(apiClient *clients.Settings) *RoleBindingBuilder {
	return NewRoleBindingBuilder(
		apiClient,
		defaultRoleBindingName,
		"",
		defaultRoleBindingRole,
		defaultRoleBindingSubject,
	)
}
