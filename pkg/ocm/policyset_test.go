package ocm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	policiesv1beta1 "open-cluster-management.io/governance-policy-propagator/api/v1beta1"
)

const (
	defaultPolicySetName   = "policyset-test"
	defaultPolicySetNsName = "test-ns"
)

func TestNewPolicySetBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *PolicySetBuilder {
				return NewPolicySetBuilder(apiClient, name, nsname, policiesv1beta1.NonEmptyString(defaultPolicyName))
			},
			policiesv1beta1.AddToScheme,
			policySetGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		policyName    policiesv1beta1.NonEmptyString
		expectedError error
	}{
		{
			name:          "valid policy",
			policyName:    policiesv1beta1.NonEmptyString(defaultPolicyName),
			expectedError: nil,
		},
		{
			name:          "empty policy",
			policyName:    "",
			expectedError: fmt.Errorf("policyset's 'policy' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := NewPolicySetBuilder(
				clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: []clients.SchemeAttacher{policiesv1beta1.AddToScheme},
				}),
				defaultPolicySetName,
				defaultPolicySetNsName,
				testCase.policyName,
			)

			assert.Equal(t, testCase.expectedError, testBuilder.GetError())

			if testCase.expectedError == nil {
				assert.Equal(
					t,
					[]policiesv1beta1.NonEmptyString{testCase.policyName},
					testBuilder.Definition.Spec.Policies,
				)
			}
		})
	}
}

func TestPullPolicySet(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPolicySet, policiesv1beta1.AddToScheme, policySetGVK).ExecuteTests(t)
}

func TestPolicySetBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[policiesv1beta1.PolicySet, PolicySetBuilder](
		policiesv1beta1.AddToScheme,
		policySetGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestPolicySetWithAdditionalPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		policyName    policiesv1beta1.NonEmptyString
		valid         bool
		expectedError error
	}{
		{
			name:       "valid additional policy",
			policyName: policiesv1beta1.NonEmptyString(defaultPolicyName),
			valid:      true,
		},
		{
			name:          "empty additional policy",
			policyName:    "",
			valid:         true,
			expectedError: fmt.Errorf("policy in PolicySet Policies spec cannot be empty"),
		},
		{
			name:       "invalid builder",
			policyName: policiesv1beta1.NonEmptyString(defaultPolicyName),
			valid:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var policySetBuilder *PolicySetBuilder
			if testCase.valid {
				policySetBuilder = buildValidPolicySetTestBuilder(buildTestClientWithPolicySetScheme())
			} else {
				policySetBuilder = buildInvalidPolicySetTestBuilder(buildTestClientWithPolicySetScheme())
			}

			policySetBuilder = policySetBuilder.WithAdditionalPolicy(testCase.policyName)

			switch testCase.name {
			case "invalid builder":
				require.Error(t, policySetBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(policySetBuilder.GetError()))
			default:
				assert.Equal(t, testCase.expectedError, policySetBuilder.GetError())

				if testCase.expectedError == nil {
					assert.Equal(
						t,
						[]policiesv1beta1.NonEmptyString{
							policiesv1beta1.NonEmptyString(defaultPolicyName),
							testCase.policyName,
						},
						policySetBuilder.Definition.Spec.Policies,
					)
				}
			}
		})
	}
}

// buildDummyPolicySet returns a PolicySet with the provided name and namespace.
func buildDummyPolicySet(name, nsname string) *policiesv1beta1.PolicySet {
	return &policiesv1beta1.PolicySet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyPolicySet returns a client with a mock dummy PolicySet.
func buildTestClientWithDummyPolicySet() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPolicySet(defaultPolicySetName, defaultPolicySetNsName),
		},
		SchemeAttachers: []clients.SchemeAttacher{
			policiesv1beta1.AddToScheme,
		},
	})
}

// buildTestClientWithPolicySetScheme returns a client with no objects but the PolicySet scheme attached.
func buildTestClientWithPolicySetScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{
			policiesv1beta1.AddToScheme,
		},
	})
}

// buildValidPolicySetTestBuilder returns a valid PolicySetBuilder for testing.
func buildValidPolicySetTestBuilder(apiClient *clients.Settings) *PolicySetBuilder {
	return NewPolicySetBuilder(
		apiClient, defaultPolicySetName, defaultPolicySetNsName, policiesv1beta1.NonEmptyString(defaultPolicyName))
}

// buildInvalidPolicySetTestBuilder returns an invalid PolicySetBuilder for testing.
func buildInvalidPolicySetTestBuilder(apiClient *clients.Settings) *PolicySetBuilder {
	return NewPolicySetBuilder(apiClient, defaultPolicySetName, "", policiesv1beta1.NonEmptyString(defaultPolicyName))
}
