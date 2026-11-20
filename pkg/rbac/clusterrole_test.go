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

func TestNewClusterRoleBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common cluster-scoped builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewClusterScopedBuilderTestConfig(
			func(apiClient *clients.Settings, name string) *ClusterRoleBuilder {
				return NewClusterRoleBuilder(apiClient, name, validClusterRoleRule())
			},
			rbacv1.AddToScheme,
			clusterRoleGVK,
		).ExecuteTests(t)
	})

	t.Run("applies initial rule", func(t *testing.T) {
		t.Parallel()

		rule := validClusterRoleRule()
		builder := NewClusterRoleBuilder(getClusterRoleTestClient(), "clusterrole-test", rule)

		require.NotNil(t, builder)
		require.NoError(t, builder.GetError())
		assert.Equal(t, []rbacv1.PolicyRule{rule}, builder.Definition.Rules)
	})

	t.Run("rejects initial rule without verbs", func(t *testing.T) {
		t.Parallel()

		builder := NewClusterRoleBuilder(getClusterRoleTestClient(), "clusterrole-test", rbacv1.PolicyRule{})

		require.NotNil(t, builder)
		require.EqualError(t, builder.GetError(), clusterroleRuleMustContainAtLeast)
		assert.Empty(t, builder.Definition.Rules)
	})
}

func TestPullClusterRole(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullClusterRole,
		rbacv1.AddToScheme,
		clusterRoleGVK,
	).ExecuteTests(t)
}

func TestClusterRoleBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newClusterRoleCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestClusterRoleWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newClusterRoleCommonTestConfig()).ExecuteTests(t)
}

func TestClusterRoleWithRules(t *testing.T) {
	t.Parallel()

	additionalRules := []rbacv1.PolicyRule{
		{Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
		{Resources: []string{"services"}, APIGroups: []string{"v1"}, Verbs: []string{"list"}},
	}

	testCases := []struct {
		name          string
		builder       func() *ClusterRoleBuilder
		rules         []rbacv1.PolicyRule
		expectedRules []rbacv1.PolicyRule
		assertError   func(error) bool
	}{
		{
			name: "appends valid rules",
			builder: func() *ClusterRoleBuilder {
				return buildValidClusterRoleTestBuilder(getClusterRoleTestClient())
			},
			rules: additionalRules,
			expectedRules: append(
				[]rbacv1.PolicyRule{validClusterRoleRule()},
				additionalRules...,
			),
			assertError: isNilError,
		},
		{
			name: "rejects empty rules",
			builder: func() *ClusterRoleBuilder {
				return buildValidClusterRoleTestBuilder(getClusterRoleTestClient())
			},
			rules:         []rbacv1.PolicyRule{},
			expectedRules: []rbacv1.PolicyRule{validClusterRoleRule()},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "cannot accept nil or empty slice as rules"
			},
		},
		{
			name: "rejects rule without verbs",
			builder: func() *ClusterRoleBuilder {
				return buildValidClusterRoleTestBuilder(getClusterRoleTestClient())
			},
			rules: []rbacv1.PolicyRule{{
				Resources: []string{"pods"},
				APIGroups: []string{"v1"},
			}},
			expectedRules: []rbacv1.PolicyRule{validClusterRoleRule()},
			assertError: func(err error) bool {
				return err != nil && err.Error() == clusterroleRuleMustContainAtLeast
			},
		},
		{
			name: "invalid builder short circuits",
			builder: func() *ClusterRoleBuilder {
				return NewClusterRoleBuilder(getClusterRoleTestClient(), "", validClusterRoleRule())
			},
			rules:         additionalRules,
			expectedRules: nil,
			assertError:   commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithRules(testCase.rules)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())
			assert.Equal(t, testCase.expectedRules, result.Definition.Rules)
		})
	}
}

func newClusterRoleCommonTestConfig() testhelper.CommonTestConfig[
	rbacv1.ClusterRole,
	ClusterRoleBuilder,
	*rbacv1.ClusterRole,
	*ClusterRoleBuilder,
] {
	return testhelper.NewCommonTestConfig[rbacv1.ClusterRole, ClusterRoleBuilder](
		rbacv1.AddToScheme,
		clusterRoleGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

func getClusterRoleTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{rbacv1.AddToScheme},
	})
}

func buildValidClusterRoleTestBuilder(apiClient *clients.Settings) *ClusterRoleBuilder {
	return NewClusterRoleBuilder(apiClient, "clusterrole-test", validClusterRoleRule())
}

func validClusterRoleRule() rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		Resources: []string{"pods"},
		APIGroups: []string{"v1"},
		Verbs:     []string{"get"},
	}
}

func isNilError(err error) bool {
	return err == nil
}
