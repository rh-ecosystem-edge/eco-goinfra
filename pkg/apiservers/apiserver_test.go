package apiservers

import (
	"fmt"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var apiServerTestSchemes = []clients.SchemeAttacher{
	configv1.Install,
}

func TestPullAPIServer(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("apiserver object %s does not exist", apiServerName),
		},
		{
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("apiserver 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testAPIServer := buildDummyAPIServer()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testAPIServer)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: apiServerTestSchemes,
			})
		}

		testBuilder, err := PullAPIServer(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, apiServerName, testBuilder.Definition.Name)
		}
	}
}

func TestAPIServerGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *APIServerBuilder
		expectedError string
	}{
		{
			testBuilder:   newAPIServerBuilder(buildTestClientWithDummyAPIServer()),
			expectedError: "",
		},
		{
			testBuilder: newAPIServerBuilder(clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: apiServerTestSchemes,
			})),
			expectedError: "apiservers.config.openshift.io \"cluster\" not found",
		},
	}

	for _, testCase := range testCases {
		apiServerObject, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, apiServerObject.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestAPIServerExists(t *testing.T) {
	testCases := []struct {
		testBuilder *APIServerBuilder
		exists      bool
	}{
		{
			testBuilder: newAPIServerBuilder(buildTestClientWithDummyAPIServer()),
			exists:      true,
		},
		{
			testBuilder: newAPIServerBuilder(clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: apiServerTestSchemes,
			})),
			exists: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestAPIServerUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *APIServerBuilder
		expectedError error
	}{
		{
			testBuilder:   newAPIServerBuilder(buildTestClientWithDummyAPIServer()),
			expectedError: nil,
		},
		{
			testBuilder: newAPIServerBuilder(clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: apiServerTestSchemes,
			})),
			expectedError: fmt.Errorf("apiserver object %s does not exist", apiServerName),
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, testCase.testBuilder.Definition.Spec.TLSSecurityProfile)

		testCase.testBuilder.Definition.Spec.TLSSecurityProfile = &configv1.TLSSecurityProfile{
			Type: configv1.TLSProfileOldType,
		}

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, configv1.TLSProfileOldType, testBuilder.Object.Spec.TLSSecurityProfile.Type)
		}
	}
}

func TestAPIServerWithTLSAdherence(t *testing.T) {
	testCases := []struct {
		policy   configv1.TLSAdherencePolicy
		expected configv1.TLSAdherencePolicy
	}{
		{
			policy:   configv1.TLSAdherencePolicyStrictAllComponents,
			expected: configv1.TLSAdherencePolicyStrictAllComponents,
		},
		{
			policy:   configv1.TLSAdherencePolicyLegacyAdheringComponentsOnly,
			expected: configv1.TLSAdherencePolicyLegacyAdheringComponentsOnly,
		},
	}

	for _, testCase := range testCases {
		testBuilder := newAPIServerBuilder(buildTestClientWithDummyAPIServer())
		result := testBuilder.WithTLSAdherence(testCase.policy)

		assert.Equal(t, testBuilder, result)
		assert.Equal(t, testCase.expected, result.Definition.Spec.TLSAdherence)
	}
}

func TestAPIServerWithTLSSecurityProfile(t *testing.T) {
	testCases := []struct {
		profile       *configv1.TLSSecurityProfile
		expectedError string
	}{
		{
			profile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			},
			expectedError: "",
		},
		{
			profile:       nil,
			expectedError: "apiserver TLS security profile cannot be nil",
		},
	}

	for _, testCase := range testCases {
		testBuilder := newAPIServerBuilder(buildTestClientWithDummyAPIServer())
		result := testBuilder.WithTLSSecurityProfile(testCase.profile)

		assert.Equal(t, testBuilder, result)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.profile, result.Definition.Spec.TLSSecurityProfile)
		} else {
			assert.Equal(t, testCase.expectedError, result.errorMsg)
		}
	}
}

func TestAPIServerBuilderValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		builderErrMsg string
		expectedError string
	}{
		{
			builderNil:    true,
			expectedError: "error: received nil apiservers.config.openshift.io builder",
		},
		{
			definitionNil: true,
			expectedError: "can not redefine the undefined apiservers.config.openshift.io",
		},
		{
			apiClientNil:  true,
			expectedError: "apiservers.config.openshift.io builder cannot have nil apiClient",
		},
		{
			expectedError: "",
		},
		{
			builderErrMsg: "test error",
			expectedError: "test error",
		},
	}

	for _, testCase := range testCases {
		testBuilder := newAPIServerBuilder(clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: apiServerTestSchemes,
		}))

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.builderErrMsg != "" {
			testBuilder.errorMsg = testCase.builderErrMsg
		}

		valid, err := testBuilder.validate()

		if testCase.expectedError != "" {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

func buildDummyAPIServer() *configv1.APIServer {
	return &configv1.APIServer{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiServerName,
		},
	}
}

func buildTestClientWithDummyAPIServer() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyAPIServer()},
		SchemeAttachers: apiServerTestSchemes,
	})
}

func newAPIServerBuilder(apiClient *clients.Settings) *APIServerBuilder {
	return &APIServerBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummyAPIServer(),
	}
}
