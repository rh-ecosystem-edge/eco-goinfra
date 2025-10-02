package ingress

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

var testSchemes = []clients.SchemeAttacher{
	networkingv1.AddToScheme,
}

const (
	defaultIngressName      = "test-ingress"
	defaultIngressNamespace = "test-namespace"
)

func TestNewIngressBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		ingressName   string
		namespace     string
		client        bool
		expectedError string
	}{
		{
			name:          "valid parameters",
			ingressName:   defaultIngressName,
			namespace:     defaultIngressNamespace,
			client:        true,
			expectedError: "",
		},
		{
			name:          "empty ingress name",
			ingressName:   "",
			namespace:     defaultIngressNamespace,
			client:        true,
			expectedError: "ingress 'name' cannot be empty",
		},
		{
			name:          "empty namespace",
			ingressName:   defaultIngressName,
			namespace:     "",
			client:        true,
			expectedError: "ingress 'namespace' cannot be empty",
		},
		{
			name:          "nil client",
			ingressName:   defaultIngressName,
			namespace:     defaultIngressNamespace,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: testSchemes,
				})
			}

			ingressBuilder := NewIngressBuilder(testSettings, testCase.ingressName, testCase.namespace)

			if testCase.client {
				if testCase.expectedError == "" {
					assert.NotNil(t, ingressBuilder)
					assert.Equal(t, testCase.ingressName, ingressBuilder.Definition.Name)
					assert.Equal(t, testCase.namespace, ingressBuilder.Definition.Namespace)
					assert.Empty(t, ingressBuilder.errorMsg)
				} else {
					assert.NotNil(t, ingressBuilder)
					assert.Equal(t, testCase.expectedError, ingressBuilder.errorMsg)
				}
			} else {
				assert.Nil(t, ingressBuilder)
			}
		})
	}
}

func TestPullIngress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		ingressName         string
		namespace           string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                "valid ingress exists",
			ingressName:         defaultIngressName,
			namespace:           defaultIngressNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "empty ingress name",
			ingressName:         "",
			namespace:           defaultIngressNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("ingress name cannot be empty"),
		},
		{
			name:                "empty namespace",
			ingressName:         defaultIngressName,
			namespace:           "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("ingress namespace cannot be empty"),
		},
		{
			name:                "ingress does not exist",
			ingressName:         defaultIngressName,
			namespace:           defaultIngressNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf("could not find ingress %s in namespace %s",
				defaultIngressName, defaultIngressNamespace),
		},
		{
			name:                "nil client",
			ingressName:         defaultIngressName,
			namespace:           defaultIngressNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("ingress 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				runtimeObjects []runtime.Object
				testSettings   *clients.Settings
			)

			if testCase.addToRuntimeObjects {
				runtimeObjects = append(runtimeObjects, buildDummyIngress(testCase.ingressName, testCase.namespace))
			}

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  runtimeObjects,
					SchemeAttachers: testSchemes,
				})
			}

			ingressBuilder, err := PullIngress(testSettings, testCase.ingressName, testCase.namespace)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.ingressName, ingressBuilder.Object.Name)
				assert.Equal(t, testCase.namespace, ingressBuilder.Object.Namespace)
			}
		})
	}
}

func TestIngressGet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		testBuilder     *IngressBuilder
		expectedIngress *networkingv1.Ingress
	}{
		{
			name:            "ingress exists",
			testBuilder:     buildValidIngressTestBuilder(buildTestClientWithDummyIngress()),
			expectedIngress: buildDummyIngress(defaultIngressName, defaultIngressNamespace),
		},
		{
			name:            "ingress does not exist",
			testBuilder:     buildValidIngressTestBuilder(buildTestClientWithIngressScheme()),
			expectedIngress: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ingress, err := testCase.testBuilder.Get()

			if testCase.expectedIngress == nil {
				assert.Nil(t, ingress)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, testCase.expectedIngress.Name, ingress.Name)
				assert.Equal(t, testCase.expectedIngress.Namespace, ingress.Namespace)
			}
		})
	}
}

func TestIngressExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		testBuilder *IngressBuilder
		exists      bool
	}{
		{
			name:        "ingress exists",
			testBuilder: buildValidIngressTestBuilder(buildTestClientWithDummyIngress()),
			exists:      true,
		},
		{
			name:        "invalid builder",
			testBuilder: buildInvalidIngressTestBuilder(buildTestClientWithDummyIngress()),
			exists:      false,
		},
		{
			name:        "ingress does not exist",
			testBuilder: buildValidIngressTestBuilder(buildTestClientWithIngressScheme()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			exists := testCase.testBuilder.Exists()
			assert.Equal(t, testCase.exists, exists)
		})
	}
}

func TestIngressCreate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		testBuilder   *IngressBuilder
		expectedError error
	}{
		{
			name:          "create new ingress",
			testBuilder:   buildValidIngressTestBuilder(buildTestClientWithIngressScheme()),
			expectedError: nil,
		},
		{
			name:          "create ingress with invalid namespace",
			testBuilder:   buildInvalidIngressTestBuilder(buildTestClientWithIngressScheme()),
			expectedError: fmt.Errorf("ingress 'namespace' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ingressBuilder, err := testCase.testBuilder.Create()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, ingressBuilder.Definition, ingressBuilder.Object)
			}
		})
	}
}

func TestIngressUpdate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		testBuilder   *IngressBuilder
		expectedError error
	}{
		{
			name:          "successfully update ingress",
			testBuilder:   buildValidIngressTestBuilder(buildTestClientWithDummyIngress()),
			expectedError: nil,
		},
		{
			name:        "update non-existent ingress",
			testBuilder: buildValidIngressTestBuilder(buildTestClientWithIngressScheme()),
			expectedError: fmt.Errorf("ingress object %s does not exist in namespace %s",
				defaultIngressName, defaultIngressNamespace),
		},
		{
			name:          "update invalid ingress",
			testBuilder:   buildInvalidIngressTestBuilder(buildTestClientWithDummyIngress()),
			expectedError: fmt.Errorf("ingress 'namespace' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testCase.testBuilder.Definition.Spec.IngressClassName = ptr.To("test-ingress-class")
			ingressBuilder, err := testCase.testBuilder.Update()

			if testCase.expectedError == nil {
				assert.Nil(t, err)
				assert.NotNil(t, ingressBuilder)
				assert.Equal(t, "test-ingress-class", *ingressBuilder.Object.Spec.IngressClassName)
			} else {
				assert.Equal(t, testCase.expectedError, err)
			}
		})
	}
}

func TestIngressDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		testBuilder   *IngressBuilder
		expectedError error
	}{
		{
			name:          "delete existing ingress",
			testBuilder:   buildValidIngressTestBuilder(buildTestClientWithDummyIngress()),
			expectedError: nil,
		},
		{
			name:          "delete non-existent ingress",
			testBuilder:   buildValidIngressTestBuilder(buildTestClientWithIngressScheme()),
			expectedError: nil,
		},
		{
			name:          "delete invalid ingress",
			testBuilder:   buildInvalidIngressTestBuilder(buildTestClientWithDummyIngress()),
			expectedError: fmt.Errorf("ingress 'namespace' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.testBuilder.Delete()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Nil(t, testCase.testBuilder.Object)
			}
		})
	}
}

func TestIngressValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		builderNil      bool
		definitionNil   bool
		apiClientNil    bool
		builderErrorMsg string
		expectedError   error
	}{
		{
			name:            "valid builder",
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   nil,
		},
		{
			name:            "nil builder",
			builderNil:      true,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("error: received nil ingress builder"),
		},
		{
			name:            "nil definition",
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined ingress"),
		},
		{
			name:            "nil apiClient",
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("ingress builder cannot have nil apiClient"),
		},
		{
			name:            "builder error message set",
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "test error",
			expectedError:   fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ingressBuilder := buildValidIngressTestBuilder(buildTestClientWithIngressScheme())

			if testCase.builderNil {
				ingressBuilder = nil
			}

			if testCase.definitionNil {
				ingressBuilder.Definition = nil
			}

			if testCase.apiClientNil {
				ingressBuilder.apiClient = nil
			}

			if testCase.builderErrorMsg != "" {
				ingressBuilder.errorMsg = testCase.builderErrorMsg
			}

			valid, err := ingressBuilder.validate()

			if testCase.expectedError != nil {
				assert.False(t, valid)
				assert.Equal(t, testCase.expectedError, err)
			} else {
				assert.True(t, valid)
				assert.Nil(t, err)
			}
		})
	}
}

// buildDummyIngress returns an Ingress with the provided name and namespace.
func buildDummyIngress(name, namespace string) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// buildTestClientWithDummyIngress returns a client with a mock dummy ingress.
func buildTestClientWithDummyIngress() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyIngress(defaultIngressName, defaultIngressNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildTestClientWithIngressScheme returns a client with no objects but the Ingress scheme attached.
func buildTestClientWithIngressScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

// buildValidIngressTestBuilder returns a valid IngressBuilder for testing.
func buildValidIngressTestBuilder(apiClient *clients.Settings) *IngressBuilder {
	return NewIngressBuilder(apiClient, defaultIngressName, defaultIngressNamespace)
}

// buildInvalidIngressTestBuilder returns an invalid IngressBuilder for testing.
func buildInvalidIngressTestBuilder(apiClient *clients.Settings) *IngressBuilder {
	return NewIngressBuilder(apiClient, defaultIngressName, "")
}
