package keda

import (
	"testing"

	kedav2v1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultTriggerAuthName      = "keda-trigger-auth-prometheus"
	defaultTriggerAuthNamespace = "test-appspace"
)

func TestNewTriggerAuthenticationBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[kedav2v1alpha1.TriggerAuthentication, TriggerAuthenticationBuilder](
		NewTriggerAuthenticationBuilder,
		kedav2v1alpha1.AddToScheme,
		triggerAuthenticationGVK,
	).ExecuteTests(t)
}

func TestPullTriggerAuthentication(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[kedav2v1alpha1.TriggerAuthentication, TriggerAuthenticationBuilder](
		PullTriggerAuthentication,
		kedav2v1alpha1.AddToScheme,
		triggerAuthenticationGVK,
	).ExecuteTests(t)
}

func TestTriggerAuthenticationBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[
		kedav2v1alpha1.TriggerAuthentication,
		TriggerAuthenticationBuilder,
	](
		kedav2v1alpha1.AddToScheme,
		triggerAuthenticationGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestTriggerAuthenticationWithSecretTargetRef(t *testing.T) {
	t.Parallel()

	tokenReference := kedav2v1alpha1.AuthSecretTargetRef{Name: "token-name", Key: "token"}
	certificateReference := kedav2v1alpha1.AuthSecretTargetRef{Name: "cert-name", Key: "ca.crt"}

	testCases := []struct {
		name             string
		invalidBuilder   bool
		secretTargetRefs []kedav2v1alpha1.AuthSecretTargetRef
		expectedError    string
		expectedRefs     []kedav2v1alpha1.AuthSecretTargetRef
	}{
		{
			name:             "sets token reference",
			secretTargetRefs: []kedav2v1alpha1.AuthSecretTargetRef{tokenReference},
			expectedRefs:     []kedav2v1alpha1.AuthSecretTargetRef{tokenReference},
		},
		{
			name:             "sets certificate reference",
			secretTargetRefs: []kedav2v1alpha1.AuthSecretTargetRef{certificateReference},
			expectedRefs:     []kedav2v1alpha1.AuthSecretTargetRef{certificateReference},
		},
		{
			name:             "sets multiple references",
			secretTargetRefs: []kedav2v1alpha1.AuthSecretTargetRef{tokenReference, certificateReference},
			expectedRefs:     []kedav2v1alpha1.AuthSecretTargetRef{tokenReference, certificateReference},
		},
		{
			name:             "rejects empty references",
			secretTargetRefs: []kedav2v1alpha1.AuthSecretTargetRef{},
			expectedError:    "'secretTargetRef' argument cannot be empty",
		},
		{
			name:             "invalid builder short circuits",
			invalidBuilder:   true,
			secretTargetRefs: []kedav2v1alpha1.AuthSecretTargetRef{tokenReference},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			triggerAuthenticationName := defaultTriggerAuthName
			if testCase.invalidBuilder {
				triggerAuthenticationName = ""
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{})
			builder := NewTriggerAuthenticationBuilder(
				testSettings, triggerAuthenticationName, defaultTriggerAuthNamespace)
			require.NotNil(t, builder)

			result := builder.WithSecretTargetRef(testCase.secretTargetRefs)

			require.Same(t, builder, result)

			switch {
			case testCase.invalidBuilder:
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
			case testCase.expectedError != "":
				require.EqualError(t, result.GetError(), testCase.expectedError)
			default:
				require.NoError(t, result.GetError())
			}

			assert.Equal(t, testCase.expectedRefs, result.Definition.Spec.SecretTargetRef)
		})
	}
}
