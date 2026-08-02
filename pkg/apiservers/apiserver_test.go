package apiservers

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var apiServerGVK = configv1.GroupVersion.WithKind("APIServer")

func TestPullAPIServer(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		PullAPIServer,
		configv1.Install,
		apiServerGVK,
		apiServerName,
	).ExecuteTests(t)
}

func TestAPIServerBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[configv1.APIServer, APIServerBuilder](
		configv1.Install, apiServerGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestAPIServerWithTLSAdherence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		policy   configv1.TLSAdherencePolicy
		expected configv1.TLSAdherencePolicy
		builder  func() *APIServerBuilder
	}{
		{
			name:     "sets strict policy on valid builder",
			policy:   configv1.TLSAdherencePolicyStrictAllComponents,
			expected: configv1.TLSAdherencePolicyStrictAllComponents,
			builder: func() *APIServerBuilder {
				return buildValidAPIServerBuilder(buildTestClientWithDummyAPIServer())
			},
		},
		{
			name:     "sets legacy policy on valid builder",
			policy:   configv1.TLSAdherencePolicyLegacyAdheringComponentsOnly,
			expected: configv1.TLSAdherencePolicyLegacyAdheringComponentsOnly,
			builder: func() *APIServerBuilder {
				return buildValidAPIServerBuilder(buildTestClientWithDummyAPIServer())
			},
		},
		{
			name:   "invalid builder short circuits",
			policy: configv1.TLSAdherencePolicyStrictAllComponents,
			builder: func() *APIServerBuilder {
				return buildInvalidAPIServerBuilder(newAPIServerTestClient())
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := testCase.builder()
			require.NotNil(t, testBuilder)

			result := testBuilder.WithTLSAdherence(testCase.policy)
			require.Same(t, testBuilder, result)

			if testCase.expected != "" {
				require.Nil(t, result.GetError())
				assert.Equal(t, testCase.expected, result.Definition.Spec.TLSAdherence)
			} else {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
			}
		})
	}
}

func TestAPIServerWithTLSSecurityProfile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		profile     *configv1.TLSSecurityProfile
		assertError func(error) bool
		expectNil   bool
		builder     func() *APIServerBuilder
	}{
		{
			name: "sets profile on valid builder",
			profile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			},
			assertError: func(err error) bool { return err == nil },
			builder: func() *APIServerBuilder {
				return buildValidAPIServerBuilder(buildTestClientWithDummyAPIServer())
			},
		},
		{
			name:    "nil profile sets builder error",
			profile: nil,
			assertError: func(err error) bool {
				return err != nil && err.Error() == "apiserver TLS security profile cannot be nil"
			},
			expectNil: true,
			builder: func() *APIServerBuilder {
				return buildValidAPIServerBuilder(buildTestClientWithDummyAPIServer())
			},
		},
		{
			name: "invalid builder short circuits",
			profile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			},
			assertError: commonerrors.IsBuilderNameEmpty,
			expectNil:   true,
			builder: func() *APIServerBuilder {
				return buildInvalidAPIServerBuilder(newAPIServerTestClient())
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := testCase.builder()
			require.NotNil(t, testBuilder)

			result := testBuilder.WithTLSSecurityProfile(testCase.profile)
			require.Same(t, testBuilder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if testCase.expectNil {
				assert.Nil(t, result.Definition.Spec.TLSSecurityProfile)
			} else {
				assert.Equal(t, testCase.profile, result.Definition.Spec.TLSSecurityProfile)
			}
		})
	}
}

func newAPIServerTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{configv1.Install},
	})
}

func buildValidAPIServerBuilder(apiClient *clients.Settings) *APIServerBuilder {
	return common.NewClusterScopedBuilder[configv1.APIServer, APIServerBuilder](
		apiClient, configv1.Install, apiServerName)
}

func buildInvalidAPIServerBuilder(apiClient *clients.Settings) *APIServerBuilder {
	return common.NewClusterScopedBuilder[configv1.APIServer, APIServerBuilder](
		apiClient, configv1.Install, "")
}

func buildTestClientWithDummyAPIServer() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			&configv1.APIServer{
				ObjectMeta: metav1.ObjectMeta{
					Name: apiServerName,
				},
			},
		},
		SchemeAttachers: []clients.SchemeAttacher{configv1.Install},
	})
}
