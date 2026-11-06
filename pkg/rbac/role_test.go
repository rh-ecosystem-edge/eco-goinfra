package rbac

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultRoleName   = "test"
	defaultRoleNsName = "testns"
)

func TestNewRoleBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *RoleBuilder {
				return NewRoleBuilder(apiClient, name, nsname, validRoleRule())
			},
			rbacv1.AddToScheme,
			roleGVK,
		).ExecuteTests(t)
	})

	t.Run("applies initial rule", func(t *testing.T) {
		t.Parallel()

		rule := validRoleRule()
		builder := NewRoleBuilder(getRoleTestClient(), defaultRoleName, defaultRoleNsName, rule)

		require.NotNil(t, builder)
		require.NoError(t, builder.GetError())
		assert.Equal(t, []rbacv1.PolicyRule{rule}, builder.Definition.Rules)
	})

	testCases := []struct {
		name          string
		rule          rbacv1.PolicyRule
		expectedError string
	}{
		{
			name:          "rejects rule without verbs",
			rule:          rbacv1.PolicyRule{Resources: []string{"pods"}, APIGroups: []string{"v1"}},
			expectedError: roleMustContainAtLeastOne,
		},
		{
			name:          "rejects rule without resources",
			rule:          rbacv1.PolicyRule{APIGroups: []string{"v1"}, Verbs: []string{"get"}},
			expectedError: "role must contain at least one Resource",
		},
		{
			name:          "rejects rule without API groups",
			rule:          rbacv1.PolicyRule{Resources: []string{"pods"}, Verbs: []string{"get"}},
			expectedError: "role must contain at least one APIGroup",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewRoleBuilder(
				getRoleTestClient(), defaultRoleName, defaultRoleNsName, testCase.rule)

			require.NotNil(t, builder)
			require.EqualError(t, builder.GetError(), testCase.expectedError)
			assert.Empty(t, builder.Definition.Rules)
		})
	}
}

func TestPullRole(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullRole,
		rbacv1.AddToScheme,
		roleGVK,
	).ExecuteTests(t)
}

func TestRoleBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newRoleCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestRoleWithRules(t *testing.T) {
	t.Parallel()

	additionalRules := []rbacv1.PolicyRule{
		{Resources: []string{"pods"}, APIGroups: []string{"v1"}, Verbs: []string{"get"}},
		{Resources: []string{"services"}, APIGroups: []string{"v1"}, Verbs: []string{"list"}},
	}

	testCases := []struct {
		name           string
		invalidBuilder bool
		rules          []rbacv1.PolicyRule
		expectedRules  []rbacv1.PolicyRule
		expectedError  string
	}{
		{
			name:  "appends valid rules",
			rules: additionalRules,
			expectedRules: append(
				[]rbacv1.PolicyRule{validRoleRule()},
				additionalRules...,
			),
		},
		{
			name:          "rejects empty rules",
			rules:         []rbacv1.PolicyRule{},
			expectedRules: []rbacv1.PolicyRule{validRoleRule()},
			expectedError: "cannot create role with empty rule",
		},
		{
			name:          "rejects rule without verbs",
			rules:         []rbacv1.PolicyRule{{Resources: []string{"pods"}, APIGroups: []string{"v1"}}},
			expectedRules: []rbacv1.PolicyRule{validRoleRule()},
			expectedError: roleMustContainAtLeastOne,
		},
		{
			name:          "rejects rule without resources",
			rules:         []rbacv1.PolicyRule{{APIGroups: []string{"v1"}, Verbs: []string{"get"}}},
			expectedRules: []rbacv1.PolicyRule{validRoleRule()},
			expectedError: "role must contain at least one Resource",
		},
		{
			name:          "rejects rule without API groups",
			rules:         []rbacv1.PolicyRule{{Resources: []string{"pods"}, Verbs: []string{"get"}}},
			expectedRules: []rbacv1.PolicyRule{validRoleRule()},
			expectedError: "role must contain at least one APIGroup",
		},
		{
			name:           "invalid builder short circuits",
			invalidBuilder: true,
			rules:          additionalRules,
			expectedRules:  nil,
			expectedError:  roleMustContainAtLeastOne,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidRoleBuilder(getRoleTestClient())
			if testCase.invalidBuilder {
				builder = buildInvalidRoleTestBuilder(getRoleTestClient())
			}

			require.NotNil(t, builder)

			result := builder.WithRules(testCase.rules)

			require.Same(t, builder, result)

			if testCase.expectedError == "" {
				require.NoError(t, result.GetError())
			} else {
				require.EqualError(t, result.GetError(), testCase.expectedError)
			}

			assert.Equal(t, testCase.expectedRules, result.Definition.Rules)
		})
	}
}

func TestRoleWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newRoleCommonTestConfig()).ExecuteTests(t)
}

func newRoleCommonTestConfig() testhelper.CommonTestConfig[
	rbacv1.Role,
	RoleBuilder,
	*rbacv1.Role,
	*RoleBuilder,
] {
	return testhelper.NewCommonTestConfig[rbacv1.Role, RoleBuilder](
		rbacv1.AddToScheme,
		roleGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

func getRoleTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{rbacv1.AddToScheme},
	})
}

func validRoleRule() rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		Resources: []string{"pods"},
		APIGroups: []string{"v1"},
		Verbs:     []string{"get"},
	}
}

// buildValidRoleBuilder returns a valid RoleBuilder for testing purposes.
func buildValidRoleBuilder(apiClient *clients.Settings) *RoleBuilder {
	return NewRoleBuilder(apiClient, defaultRoleName, defaultRoleNsName, validRoleRule())
}

// buildInvalidRoleTestBuilder returns a RoleBuilder that is already in an error state.
func buildInvalidRoleTestBuilder(apiClient *clients.Settings) *RoleBuilder {
	return NewRoleBuilder(apiClient, defaultRoleName, defaultRoleNsName, rbacv1.PolicyRule{})
}

// buildTestClientWithDummyObject is retained for the baseline ClusterRole tests that use a Role fixture.
func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyRoleObject(),
	})
}

func buildDummyRoleObject() []runtime.Object {
	return append([]runtime.Object{}, &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultRoleName,
			Namespace: defaultRoleNsName,
		},
	})
}
