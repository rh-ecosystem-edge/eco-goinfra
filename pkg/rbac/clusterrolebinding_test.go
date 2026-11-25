package rbac

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	defaultClusterRoleBindingName      = "clusterrolebinding-test"
	defaultClusterRoleBindingRoleName  = "clusterrole-test"
	defaultClusterRoleBindingSubjectID = "test-user"
)

func TestNewClusterRoleBindingBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common cluster-scoped builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewClusterScopedBuilderTestConfig(
			func(apiClient *clients.Settings, name string) *ClusterRoleBindingBuilder {
				return NewClusterRoleBindingBuilder(
					apiClient, name, defaultClusterRoleBindingRoleName, validClusterRoleBindingSubject())
			},
			rbacv1.AddToScheme,
			clusterRoleBindingGVK,
		).ExecuteTests(t)
	})

	t.Run("applies role reference and subject", func(t *testing.T) {
		t.Parallel()

		subject := validClusterRoleBindingSubject()
		builder := NewClusterRoleBindingBuilder(
			getClusterRoleBindingTestClient(), defaultClusterRoleBindingName, defaultClusterRoleBindingRoleName, subject)

		require.NotNil(t, builder)
		require.NoError(t, builder.GetError())
		assert.Equal(t, rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     defaultClusterRoleBindingRoleName,
			Kind:     "ClusterRole",
		}, builder.Definition.RoleRef)
		assert.Equal(t, []rbacv1.Subject{subject}, builder.Definition.Subjects)
	})

	t.Run("rejects invalid initial subject", func(t *testing.T) {
		t.Parallel()

		builder := NewClusterRoleBindingBuilder(
			getClusterRoleBindingTestClient(),
			defaultClusterRoleBindingName,
			defaultClusterRoleBindingRoleName,
			rbacv1.Subject{Name: defaultClusterRoleBindingSubjectID},
		)

		require.NotNil(t, builder)
		require.EqualError(
			t, builder.GetError(), "clusterrolebinding subject kind must be one of 'ServiceAccount', 'User', or 'Group'")
		assert.Empty(t, builder.Definition.Subjects)
	})
}

func TestPullClusterRoleBinding(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullClusterRoleBinding,
		rbacv1.AddToScheme,
		clusterRoleBindingGVK,
	).ExecuteTests(t)
}

func TestClusterRoleBindingBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newClusterRoleBindingCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestClusterRoleBindingWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newClusterRoleBindingCommonTestConfig()).ExecuteTests(t)
}

func TestClusterRoleBindingWithSubjects(t *testing.T) {
	t.Parallel()

	additionalSubjects := []rbacv1.Subject{
		{Kind: "ServiceAccount", Name: "test-service-account", Namespace: "test-namespace"},
		{Kind: "Group", Name: "test-group"},
	}

	testCases := []struct {
		name             string
		builder          func() *ClusterRoleBindingBuilder
		subjects         []rbacv1.Subject
		expectedSubjects []rbacv1.Subject
		expectedError    string
		assertError      func(error) bool
	}{
		{
			name: "appends valid subjects",
			builder: func() *ClusterRoleBindingBuilder {
				return buildValidClusterRoleBindingTestBuilder(getClusterRoleBindingTestClient())
			},
			subjects: additionalSubjects,
			expectedSubjects: append(
				[]rbacv1.Subject{validClusterRoleBindingSubject()},
				additionalSubjects...,
			),
			assertError: clusterRoleBindingErrorIsNil,
		},
		{
			name: "rejects empty subjects",
			builder: func() *ClusterRoleBindingBuilder {
				return buildValidClusterRoleBindingTestBuilder(getClusterRoleBindingTestClient())
			},
			subjects:         []rbacv1.Subject{},
			expectedSubjects: []rbacv1.Subject{validClusterRoleBindingSubject()},
			expectedError:    "cannot accept nil or empty slice as subjects",
		},
		{
			name: "rejects invalid subject kind",
			builder: func() *ClusterRoleBindingBuilder {
				return buildValidClusterRoleBindingTestBuilder(getClusterRoleBindingTestClient())
			},
			subjects:         []rbacv1.Subject{{Kind: "Invalid", Name: "test"}},
			expectedSubjects: []rbacv1.Subject{validClusterRoleBindingSubject()},
			expectedError:    "clusterrolebinding subject kind must be one of 'ServiceAccount', 'User', or 'Group'",
		},
		{
			name: "rejects empty subject name",
			builder: func() *ClusterRoleBindingBuilder {
				return buildValidClusterRoleBindingTestBuilder(getClusterRoleBindingTestClient())
			},
			subjects:         []rbacv1.Subject{{Kind: "User"}},
			expectedSubjects: []rbacv1.Subject{validClusterRoleBindingSubject()},
			expectedError:    "clusterrolebinding subject name cannot be empty",
		},
		{
			name: "invalid builder short circuits",
			builder: func() *ClusterRoleBindingBuilder {
				return NewClusterRoleBindingBuilder(
					getClusterRoleBindingTestClient(), "", defaultClusterRoleBindingRoleName, validClusterRoleBindingSubject())
			},
			subjects:    additionalSubjects,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithSubjects(testCase.subjects)
			require.Same(t, builder, result)

			switch {
			case testCase.assertError != nil:
				require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())
			case testCase.expectedError == "":
				require.NoError(t, result.GetError())
			default:
				require.EqualError(t, result.GetError(), testCase.expectedError)
			}

			assert.Equal(t, testCase.expectedSubjects, result.Definition.Subjects)
		})
	}
}

func newClusterRoleBindingCommonTestConfig() testhelper.CommonTestConfig[
	rbacv1.ClusterRoleBinding,
	ClusterRoleBindingBuilder,
	*rbacv1.ClusterRoleBinding,
	*ClusterRoleBindingBuilder,
] {
	return testhelper.NewCommonTestConfig[rbacv1.ClusterRoleBinding, ClusterRoleBindingBuilder](
		rbacv1.AddToScheme,
		clusterRoleBindingGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

func getClusterRoleBindingTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{rbacv1.AddToScheme},
	})
}

func buildValidClusterRoleBindingTestBuilder(apiClient *clients.Settings) *ClusterRoleBindingBuilder {
	return NewClusterRoleBindingBuilder(
		apiClient,
		defaultClusterRoleBindingName,
		defaultClusterRoleBindingRoleName,
		validClusterRoleBindingSubject(),
	)
}

func validClusterRoleBindingSubject() rbacv1.Subject {
	return rbacv1.Subject{Kind: "User", Name: defaultClusterRoleBindingSubjectID}
}

func clusterRoleBindingErrorIsNil(err error) bool {
	return err == nil
}
