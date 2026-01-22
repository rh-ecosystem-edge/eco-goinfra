package nfd

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	nfdv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/nfd/v1alpha1"
	"github.com/stretchr/testify/assert"
)

var (
	nodeFeatureRuleExampleName = "test-node-feature-rule"
	nodeFeatureRuleNamespace   = "test-namespace"
	nodeFeatureRuleAlmExample  = fmt.Sprintf(`[
	{
		"apiVersion": "nfd.openshift.io/v1alpha1",
		"kind": "NodeFeatureRule",
		"metadata": {
			"name": "%s",
			"namespace": "%s"
		}
	}]`, nodeFeatureRuleExampleName, nodeFeatureRuleNamespace)

	nfdRuleTestSchemes = []clients.SchemeAttacher{
		nfdv1.AddToScheme,
	}
)

func TestNewNodeFeatureRuleBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		ruleName          string
		namespace         string
		client            bool
		expectedErrorText string
	}{
		{
			name:              "Valid builder with all parameters",
			ruleName:          nodeFeatureRuleExampleName,
			namespace:         nodeFeatureRuleNamespace,
			client:            true,
			expectedErrorText: "",
		},
		{
			name:              "Empty name",
			ruleName:          "",
			namespace:         nodeFeatureRuleNamespace,
			client:            true,
			expectedErrorText: "nodeFeatureRule 'name' cannot be empty",
		},
		{
			name:              "Empty namespace",
			ruleName:          nodeFeatureRuleExampleName,
			namespace:         "",
			client:            true,
			expectedErrorText: "nodeFeatureRule 'namespace' cannot be empty",
		},
		{
			name:              "No Client Provided",
			ruleName:          nodeFeatureRuleExampleName,
			namespace:         nodeFeatureRuleNamespace,
			client:            false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var client *clients.Settings
			if testCase.client {
				client = buildTestClientWithNFDRuleScheme()
			}

			builder := NewNodeFeatureRuleBuilder(client, testCase.ruleName, testCase.namespace)

			if testCase.client {
				if testCase.expectedErrorText != "" {
					assert.NotNil(t, builder)
					assert.Equal(t, testCase.expectedErrorText, builder.errorMsg)
				} else {
					assert.NotNil(t, builder)
					assert.Equal(t, testCase.ruleName, builder.Definition.Name)
					assert.Equal(t, testCase.namespace, builder.Definition.Namespace)
					assert.Equal(t, "", builder.errorMsg)
				}
			} else {
				assert.Nil(t, builder)
			}
		})
	}
}

func TestNewnodeFeatureRuleBuilderFromObjectString(t *testing.T) {
	testCases := []struct {
		name              string
		almString         string
		client            bool
		expectedErrorText string
	}{
		{
			name:              "Valid ALM Example with Client",
			almString:         nodeFeatureRuleAlmExample,
			client:            true,
			expectedErrorText: "",
		},
		{
			name:              "Empty ALM Example",
			almString:         "",
			client:            true,
			expectedErrorText: "error initializing NodeFeatureRule from alm-examples: almExample is an empty string",
		},
		{
			name:      "Invalid ALM Example",
			almString: "{invalid}",
			client:    true,
			expectedErrorText: "error initializing NodeFeatureRule from alm-examples:" +
				" invalid character 'i' looking for beginning of object key string",
		},
		{
			name:              "No Client Provided",
			almString:         nodeFeatureRuleAlmExample,
			client:            false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var client *clients.Settings
			if testCase.client {
				client = buildTestClientWithNFDRuleScheme()
			}

			builder := NewNodeFeatureRuleBuilderFromObjectString(client, testCase.almString)

			errormessage := ""

			if builder != nil {
				errormessage = builder.errorMsg
			}

			if testCase.client {
				assert.Equal(t, testCase.expectedErrorText, errormessage)

				if testCase.expectedErrorText == "" {
					assert.Equal(t, nodeFeatureRuleExampleName, builder.Definition.Name)
				}
			} else {
				assert.Nil(t, builder)
			}
		})
	}
}

func TestNodeFeatureRuleBuilderWithRule(t *testing.T) {
	testCases := []struct {
		name          string
		rule          nfdv1.Rule
		expectedError string
	}{
		{
			name: "Valid Rule",
			rule: nfdv1.Rule{
				Name: "test-rule",
				Labels: map[string]string{
					"feature.node.kubernetes.io/test": "true",
				},
			},
			expectedError: "",
		},
		{
			name: "Empty Rule Name",
			rule: nfdv1.Rule{
				Name: "",
				Labels: map[string]string{
					"feature.node.kubernetes.io/test": "true",
				},
			},
			expectedError: "rule 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			builder := NewNodeFeatureRuleBuilder(
				buildTestClientWithNFDRuleScheme(),
				nodeFeatureRuleExampleName,
				nodeFeatureRuleNamespace,
			)
			assert.NotNil(t, builder)

			builder = builder.WithRule(testCase.rule)

			if testCase.expectedError != "" {
				assert.Equal(t, testCase.expectedError, builder.errorMsg)
			} else {
				assert.Equal(t, "", builder.errorMsg)
				assert.Len(t, builder.Definition.Spec.Rules, 1)
				assert.Equal(t, testCase.rule.Name, builder.Definition.Spec.Rules[0].Name)
			}
		})
	}
}

func TestNodeFeatureRuleBuilderWithRules(t *testing.T) {
	testCases := []struct {
		name          string
		rules         []nfdv1.Rule
		expectedError string
	}{
		{
			name: "Valid Rules",
			rules: []nfdv1.Rule{
				{
					Name: "test-rule-1",
					Labels: map[string]string{
						"feature.node.kubernetes.io/test1": "true",
					},
				},
				{
					Name: "test-rule-2",
					Labels: map[string]string{
						"feature.node.kubernetes.io/test2": "true",
					},
				},
			},
			expectedError: "",
		},
		{
			name:          "Empty Rules",
			rules:         []nfdv1.Rule{},
			expectedError: "rules list cannot be empty",
		},
		{
			name: "Rule with Empty Name",
			rules: []nfdv1.Rule{
				{
					Name: "",
					Labels: map[string]string{
						"feature.node.kubernetes.io/test": "true",
					},
				},
			},
			expectedError: "rule 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			builder := NewNodeFeatureRuleBuilder(
				buildTestClientWithNFDRuleScheme(),
				nodeFeatureRuleExampleName,
				nodeFeatureRuleNamespace,
			)
			assert.NotNil(t, builder)

			builder = builder.WithRules(testCase.rules)

			if testCase.expectedError != "" {
				assert.Equal(t, testCase.expectedError, builder.errorMsg)
			} else {
				assert.Equal(t, "", builder.errorMsg)
				assert.Len(t, builder.Definition.Spec.Rules, len(testCase.rules))
			}
		})
	}
}

func TestNodeFeatureRuleBuilderWithSimplePCIRule(t *testing.T) {
	testCases := []struct {
		name          string
		ruleName      string
		labels        map[string]string
		vendorIDs     []string
		deviceIDs     []string
		expectedError string
	}{
		{
			name:     "Valid Neuron-like Rule",
			ruleName: "neuron-device",
			labels: map[string]string{
				"feature.node.kubernetes.io/aws-neuron": "true",
			},
			vendorIDs:     []string{"1d0f"},
			deviceIDs:     []string{"7064", "7065", "7066", "7067"},
			expectedError: "",
		},
		{
			name:     "Valid Rule without device IDs",
			ruleName: "vendor-only-rule",
			labels: map[string]string{
				"feature.node.kubernetes.io/vendor": "true",
			},
			vendorIDs:     []string{"1d0f"},
			deviceIDs:     []string{},
			expectedError: "",
		},
		{
			name:     "Empty Rule Name",
			ruleName: "",
			labels: map[string]string{
				"feature.node.kubernetes.io/test": "true",
			},
			vendorIDs:     []string{"1d0f"},
			deviceIDs:     []string{"7064"},
			expectedError: "rule 'name' cannot be empty",
		},
		{
			name:          "Empty Labels",
			ruleName:      "test-rule",
			labels:        map[string]string{},
			vendorIDs:     []string{"1d0f"},
			deviceIDs:     []string{"7064"},
			expectedError: "rule 'labels' cannot be empty",
		},
		{
			name:     "Empty Vendor IDs",
			ruleName: "test-rule",
			labels: map[string]string{
				"feature.node.kubernetes.io/test": "true",
			},
			vendorIDs:     []string{},
			deviceIDs:     []string{"7064"},
			expectedError: "vendorIDs cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			builder := NewNodeFeatureRuleBuilder(
				buildTestClientWithNFDRuleScheme(),
				nodeFeatureRuleExampleName,
				nodeFeatureRuleNamespace,
			)
			assert.NotNil(t, builder)

			builder = builder.WithSimplePCIRule(
				testCase.ruleName,
				testCase.labels,
				testCase.vendorIDs,
				testCase.deviceIDs,
			)

			if testCase.expectedError != "" {
				assert.Equal(t, testCase.expectedError, builder.errorMsg)
			} else {
				assert.Equal(t, "", builder.errorMsg)
				assert.Len(t, builder.Definition.Spec.Rules, 1)
				assert.Equal(t, testCase.ruleName, builder.Definition.Spec.Rules[0].Name)
				assert.Equal(t, testCase.labels, builder.Definition.Spec.Rules[0].Labels)
			}
		})
	}
}

func TestNodeFeatureRuleBuilderCreate(t *testing.T) {
	testCases := []struct {
		name          string
		builder       *NodeFeatureRuleBuilder
		expectedError error
	}{
		{
			name:          "Valid Create",
			builder:       buildValidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			expectedError: nil,
		},
		{
			name:          "Invalid Builder",
			builder:       buildInvalidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			expectedError: fmt.Errorf("can not redefine the undefined nodeFeatureRule"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			builder, err := testCase.builder.Create()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.NotNil(t, builder.Object)
				assert.Equal(t, builder.Definition.Name, builder.Object.Name)
			}
		})
	}
}

func TestNodeFeatureRuleBuilderDelete(t *testing.T) {
	testCases := []struct {
		name          string
		builder       *NodeFeatureRuleBuilder
		expectedError error
	}{
		{
			name:          "Delete Existing",
			builder:       buildValidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			expectedError: nil,
		},
		{
			name:          "Delete Non-Existing",
			builder:       buildValidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme()),
			expectedError: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			builder, err := testCase.builder.Delete()
			assert.Equal(t, testCase.expectedError, err)
			assert.Nil(t, builder.Object)
		})
	}
}

func TestNodeFeatureRuleBuilderUpdate(t *testing.T) {
	testCases := []struct {
		name          string
		builder       *NodeFeatureRuleBuilder
		force         bool
		expectedError string
	}{
		{
			name:          "Update Existing",
			builder:       buildValidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			force:         false,
			expectedError: "",
		},
		{
			name:          "Update Non-Existing",
			builder:       buildValidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme()),
			force:         false,
			expectedError: "cannot update non-existent NodeFeatureRule",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			builder, err := testCase.builder.Update(testCase.force)

			if testCase.expectedError != "" {
				assert.NotNil(t, err)
				assert.Equal(t, testCase.expectedError, err.Error())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, builder.Object)
			}
		})
	}
}

func TestNodeFeatureRuleBuilderExists(t *testing.T) {
	testCases := []struct {
		name           string
		builder        *NodeFeatureRuleBuilder
		expectedStatus bool
	}{
		{
			name:           "Existing Object",
			builder:        buildValidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			expectedStatus: true,
		},
		{
			name:           "Non-Existent Object",
			builder:        buildValidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			exists := testCase.builder.Exists()
			assert.Equal(t, testCase.expectedStatus, exists)
		})
	}
}

func TestNodeFeatureRuleBuilderGet(t *testing.T) {
	testCases := []struct {
		name          string
		builder       *NodeFeatureRuleBuilder
		expectedError error
	}{
		{
			name:          "Valid Get",
			builder:       buildValidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			expectedError: nil,
		},
		{
			name: "Invalid Get - Missing Object",
			builder: buildValidNFDRuleTestBuilder(
				buildTestClientWithNFDRuleScheme(),
			),
			expectedError: fmt.Errorf("nodefeaturerules.nfd.openshift.io \"%s\" not found", nodeFeatureRuleExampleName),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			obj, err := testCase.builder.Get()
			if testCase.expectedError == nil {
				assert.NotNil(t, obj)
			} else {
				assert.Equal(t, testCase.expectedError.Error(), err.Error())
			}

			if testCase.expectedError == nil {
				assert.NotNil(t, obj)
				assert.Equal(t, testCase.builder.Definition.Name, obj.Name)
			}
		})
	}
}

func TestPullNodeFeatureRule(t *testing.T) {
	testCases := []struct {
		name          string
		ruleName      string
		namespace     string
		client        *clients.Settings
		exists        bool
		expectedError string
	}{
		{
			name:          "Valid Pull",
			ruleName:      nodeFeatureRuleExampleName,
			namespace:     nodeFeatureRuleNamespace,
			client:        buildTestClientWithDummyNFDRule(),
			exists:        true,
			expectedError: "",
		},
		{
			name:          "Empty Name",
			ruleName:      "",
			namespace:     nodeFeatureRuleNamespace,
			client:        buildTestClientWithNFDRuleScheme(),
			exists:        false,
			expectedError: "nodeFeatureRule 'name' cannot be empty",
		},
		{
			name:          "Empty Namespace",
			ruleName:      nodeFeatureRuleExampleName,
			namespace:     "",
			client:        buildTestClientWithNFDRuleScheme(),
			exists:        false,
			expectedError: "nodeFeatureRule 'namespace' cannot be empty",
		},
		{
			name:      "Non-Existent",
			ruleName:  nodeFeatureRuleExampleName,
			namespace: nodeFeatureRuleNamespace,
			client:    buildTestClientWithNFDRuleScheme(),
			exists:    false,
			expectedError: fmt.Sprintf("nodeFeatureRule object %s does not exist in namespace %s",
				nodeFeatureRuleExampleName, nodeFeatureRuleNamespace),
		},
		{
			name:          "Nil Client",
			ruleName:      nodeFeatureRuleExampleName,
			namespace:     nodeFeatureRuleNamespace,
			client:        nil,
			exists:        false,
			expectedError: "the apiClient of the NodeFeatureRule is nil",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			builder, err := PullFeatureRule(testCase.client, testCase.ruleName, testCase.namespace)

			if testCase.expectedError != "" {
				assert.NotNil(t, err)
				assert.Equal(t, testCase.expectedError, err.Error())
				assert.Nil(t, builder)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, builder)
				assert.Equal(t, testCase.ruleName, builder.Definition.Name)
			}
		})
	}
}

// Helper Functions

func buildTestClientWithDummyNFDRule() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyNFDRule(nodeFeatureRuleExampleName, nodeFeatureRuleNamespace),
		},
		SchemeAttachers: nfdRuleTestSchemes,
	})
}

func buildDummyNFDRule(name, namespace string) *nfdv1.NodeFeatureRule {
	return &nfdv1.NodeFeatureRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: nfdv1.NodeFeatureRuleSpec{},
	}
}

func buildValidNFDRuleTestBuilder(apiClient *clients.Settings) *NodeFeatureRuleBuilder {
	return NewNodeFeatureRuleBuilderFromObjectString(apiClient, nodeFeatureRuleAlmExample)
}

func buildInvalidNFDRuleTestBuilder(apiClient *clients.Settings) *NodeFeatureRuleBuilder {
	return NewNodeFeatureRuleBuilderFromObjectString(apiClient, "{invalid}")
}

func buildTestClientWithNFDRuleScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: nfdRuleTestSchemes,
	})
}
