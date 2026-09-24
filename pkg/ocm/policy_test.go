package ocm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
)

const (
	defaultPolicyName            = "policy-test"
	defaultPolicyNsName          = "test-ns"
	defaultPolicyMessage         = "wrong type for value; expected string; got int"
	defaultPolicyExpectedMessage = "wrong type for value"
)

func TestNewPolicyBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *PolicyBuilder {
				return NewPolicyBuilder(apiClient, name, nsname, &policiesv1.PolicyTemplate{})
			},
			policiesv1.AddToScheme,
			policyGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		template      *policiesv1.PolicyTemplate
		expectedError error
	}{
		{
			name:          "valid template",
			template:      &policiesv1.PolicyTemplate{},
			expectedError: nil,
		},
		{
			name:          "nil template",
			template:      nil,
			expectedError: fmt.Errorf("policy 'template' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := NewPolicyBuilder(
				clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: []clients.SchemeAttacher{policiesv1.AddToScheme},
				}),
				defaultPolicyName,
				defaultPolicyNsName,
				testCase.template,
			)

			assert.Equal(t, testCase.expectedError, testBuilder.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.template, testBuilder.Definition.Spec.PolicyTemplates[0])
			}
		})
	}
}

func TestPullPolicy(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPolicy, policiesv1.AddToScheme, policyGVK).ExecuteTests(t)
}

func TestPolicyBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[policiesv1.Policy, PolicyBuilder](
		policiesv1.AddToScheme,
		policyGVK,
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

func TestWithRemediationAction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		action        policiesv1.RemediationAction
		valid         bool
		expectedError error
	}{
		{
			name:   "inform action",
			action: policiesv1.Inform,
			valid:  true,
		},
		{
			name:          "invalid action",
			action:        "",
			valid:         true,
			expectedError: fmt.Errorf("remediation action in policy spec must be either 'Inform' or 'Enforce'"),
		},
		{
			name:   "invalid builder",
			action: policiesv1.Inform,
			valid:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var policyBuilder *PolicyBuilder
			if testCase.valid {
				policyBuilder = buildValidPolicyTestBuilder(buildTestClientWithPolicyScheme())
			} else {
				policyBuilder = buildInvalidPolicyTestBuilder(buildTestClientWithPolicyScheme())
			}

			policyBuilder = policyBuilder.WithRemediationAction(testCase.action)

			switch testCase.name {
			case "invalid builder":
				require.Error(t, policyBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(policyBuilder.GetError()))
			default:
				assert.Equal(t, testCase.expectedError, policyBuilder.GetError())

				if testCase.expectedError == nil {
					assert.Equal(t, testCase.action, policyBuilder.Definition.Spec.RemediationAction)
				}
			}
		})
	}
}

func TestWithAdditionalPolicyTemplate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		policyTemplate *policiesv1.PolicyTemplate
		valid          bool
		expectedError  error
	}{
		{
			name:           "valid template",
			policyTemplate: &policiesv1.PolicyTemplate{},
			valid:          true,
		},
		{
			name:           "nil template",
			policyTemplate: nil,
			valid:          true,
			expectedError:  fmt.Errorf("policy template in policy policytemplates cannot be nil"),
		},
		{
			name:           "invalid builder",
			policyTemplate: &policiesv1.PolicyTemplate{},
			valid:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var policyBuilder *PolicyBuilder
			if testCase.valid {
				policyBuilder = buildValidPolicyTestBuilder(buildTestClientWithPolicyScheme())
			} else {
				policyBuilder = buildInvalidPolicyTestBuilder(buildTestClientWithPolicyScheme())
			}

			policyBuilder = policyBuilder.WithAdditionalPolicyTemplate(testCase.policyTemplate)

			switch testCase.name {
			case "invalid builder":
				require.Error(t, policyBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(policyBuilder.GetError()))
			default:
				assert.Equal(t, testCase.expectedError, policyBuilder.GetError())

				if testCase.expectedError == nil {
					assert.Equal(
						t,
						[]*policiesv1.PolicyTemplate{{}, testCase.policyTemplate},
						policyBuilder.Definition.Spec.PolicyTemplates,
					)
				}
			}
		})
	}
}

func TestPolicyWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		policyExists  bool
		expectedError error
	}{
		{
			name:         "deleted policy",
			valid:        true,
			policyExists: false,
		},
		{
			name:          "timeout waiting for deletion",
			valid:         true,
			policyExists:  true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:         "invalid builder",
			valid:        false,
			policyExists: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings
			if testCase.policyExists {
				testSettings = buildTestClientWithDummyPolicy()
			} else {
				testSettings = buildTestClientWithPolicyScheme()
			}

			var policyBuilder *PolicyBuilder
			if testCase.valid {
				policyBuilder = buildValidPolicyTestBuilder(testSettings)
			} else {
				policyBuilder = buildInvalidPolicyTestBuilder(testSettings)
			}

			err := policyBuilder.WaitUntilDeleted(time.Second)

			switch testCase.name {
			case "invalid builder":
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
			case "timeout waiting for deletion":
				assert.Equal(t, context.DeadlineExceeded, err)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

func TestPolicyWaitUntilComplianceState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		state policiesv1.ComplianceState
	}{
		{state: policiesv1.Compliant},
		{state: policiesv1.NonCompliant},
		{state: policiesv1.Pending},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.state), func(t *testing.T) {
			t.Parallel()

			dummyPolicy := buildDummyPolicy()
			dummyPolicy.Status.ComplianceState = testCase.state

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: []runtime.Object{dummyPolicy},
				SchemeAttachers: []clients.SchemeAttacher{
					policiesv1.AddToScheme,
				},
			})

			policyBuilder := buildValidPolicyTestBuilder(testSettings)
			err := policyBuilder.WaitUntilComplianceState(testCase.state, 5*time.Second)

			assert.NoError(t, err)
		})
	}
}

func TestPolicyWaitForStatusMessageToContain(t *testing.T) {
	t.Parallel()

	testCases := []policyWaitForStatusMessageTestCase{
		{
			name:            "message found",
			expectedMessage: defaultPolicyExpectedMessage,
			valid:           true,
			exists:          true,
			hasMessage:      true,
		},
		{
			name:            "empty expected message",
			expectedMessage: "",
			valid:           true,
			exists:          true,
			hasMessage:      true,
			expectedError:   fmt.Errorf("policy expectedMessage is empty"),
		},
		{
			name:            "invalid builder",
			expectedMessage: defaultPolicyExpectedMessage,
			valid:           false,
			exists:          true,
			hasMessage:      true,
		},
		{
			name:            "policy not exists",
			expectedMessage: defaultPolicyExpectedMessage,
			valid:           true,
			exists:          false,
			hasMessage:      true,
			expectedError: fmt.Errorf(
				"policy object %s does not exist in namespace %s", defaultPolicyName, defaultPolicyNsName),
		},
		{
			name:            "timeout waiting for message",
			expectedMessage: defaultPolicyExpectedMessage,
			valid:           true,
			exists:          true,
			hasMessage:      false,
			expectedError:   context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runPolicyWaitForStatusMessageTestCase(t, testCase)
		})
	}
}

type policyWaitForStatusMessageTestCase struct {
	name            string
	expectedMessage string
	valid           bool
	exists          bool
	hasMessage      bool
	expectedError   error
}

func runPolicyWaitForStatusMessageTestCase(t *testing.T, testCase policyWaitForStatusMessageTestCase) {
	t.Helper()

	var runtimeObjects []runtime.Object

	if testCase.exists {
		policy := buildDummyPolicy()

		if testCase.hasMessage {
			policy.Status.Details = []*policiesv1.DetailsPerTemplate{
				{History: []policiesv1.ComplianceHistory{{Message: defaultPolicyMessage}}},
			}
		}

		runtimeObjects = append(runtimeObjects, policy)
	}

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: runtimeObjects,
		SchemeAttachers: []clients.SchemeAttacher{
			policiesv1.AddToScheme,
		},
	})

	var policyBuilder *PolicyBuilder
	if testCase.valid {
		policyBuilder = buildValidPolicyTestBuilder(testSettings)
	} else {
		policyBuilder = buildInvalidPolicyTestBuilder(testSettings)
	}

	_, err := policyBuilder.WaitForStatusMessageToContain(testCase.expectedMessage, time.Second)

	switch testCase.name {
	case "invalid builder":
		require.Error(t, err)
		assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
	case "empty expected message", "policy not exists", "timeout waiting for message":
		assert.Equal(t, testCase.expectedError, err)
	default:
		assert.NoError(t, err)
	}
}

// buildDummyPolicy returns a Policy with the default test name and namespace.
func buildDummyPolicy() *policiesv1.Policy {
	return &policiesv1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPolicyName,
			Namespace: defaultPolicyNsName,
		},
		Spec: policiesv1.PolicySpec{
			PolicyTemplates: []*policiesv1.PolicyTemplate{{}},
		},
	}
}

// buildTestClientWithDummyPolicy returns a client with a mock dummy policy.
func buildTestClientWithDummyPolicy() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPolicy(),
		},
		SchemeAttachers: []clients.SchemeAttacher{
			policiesv1.AddToScheme,
		},
	})
}

// buildTestClientWithPolicyScheme returns a client with no objects but the Policy scheme attached.
func buildTestClientWithPolicyScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{
			policiesv1.AddToScheme,
		},
	})
}

// buildValidPolicyTestBuilder returns a valid PolicyBuilder for testing.
func buildValidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(apiClient, defaultPolicyName, defaultPolicyNsName, &policiesv1.PolicyTemplate{})
}

// buildInvalidPolicyTestBuilder returns an invalid PolicyBuilder for testing.
func buildInvalidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(apiClient, defaultPolicyName, "", &policiesv1.PolicyTemplate{})
}
