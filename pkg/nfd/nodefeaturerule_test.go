package nfd

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	nfdv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/nfd/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nodeFeatureRuleExampleName = "test-node-feature-rule"
	nodeFeatureRuleNamespace   = "test-namespace"
)

var (
	nodeFeatureRuleAlmExample = fmt.Sprintf(`[
	{
		"apiVersion": "nfd.openshift.io/v1alpha1",
		"kind": "NodeFeatureRule",
		"metadata": {
			"name": "%s",
			"namespace": "%s"
		}
	}]`, nodeFeatureRuleExampleName, nodeFeatureRuleNamespace)

	nodeFeatureRuleMismatchedKindAlmExample = `[
	{
		"apiVersion": "nfd.openshift.io/v1alpha1",
		"kind": "Other",
		"metadata": {
			"name": "other",
			"namespace": "test-namespace"
		}
	}]`
)

func TestNewNodeFeatureRuleBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewNodeFeatureRuleBuilder,
		nfdv1.AddToScheme,
		nodeFeatureRuleGVK,
	).ExecuteTests(t)

	t.Run("initializes empty rules list", func(t *testing.T) {
		t.Parallel()

		builder := buildValidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme())

		require.NoError(t, builder.GetError())
		require.NotNil(t, builder.Definition.Spec.Rules)
		assert.Empty(t, builder.Definition.Spec.Rules)
	})
}

func TestNewNodeFeatureRuleBuilderFromObjectString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		almString         string
		client            bool
		expectedErrorText string
	}{
		{
			name:      "valid ALM example",
			almString: nodeFeatureRuleAlmExample,
			client:    true,
		},
		{
			name:              "empty ALM example",
			almString:         "",
			client:            true,
			expectedErrorText: "error initializing NodeFeatureRule from alm-examples: almExample is an empty string",
		},
		{
			name:      "invalid ALM example",
			almString: "{invalid}",
			client:    true,
			expectedErrorText: "error initializing NodeFeatureRule from alm-examples:" +
				" invalid character 'i' looking for beginning of object key string",
		},
		{
			name:              "empty ALM example list",
			almString:         "[]",
			client:            true,
			expectedErrorText: "error initializing NodeFeatureRule from alm-examples: failed to get alm examples",
		},
		{
			name:              "missing NodeFeatureRule in ALM example",
			almString:         nodeFeatureRuleMismatchedKindAlmExample,
			client:            true,
			expectedErrorText: "error initializing NodeFeatureRule from alm-examples: nodeFeatureRule is missing in alm-examples ",
		},
		{
			name:              "no client",
			almString:         nodeFeatureRuleAlmExample,
			client:            false,
			expectedErrorText: "apiClient for NodeFeatureRule is nil",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var client *clients.Settings
			if testCase.client {
				client = buildTestClientWithNFDRuleScheme()
			}

			builder := NewNodeFeatureRuleBuilderFromObjectString(client, testCase.almString)

			require.NotNil(t, builder)
			assert.Equal(t, nodeFeatureRuleGVK, builder.GetGVK())

			if testCase.expectedErrorText != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedErrorText)

				_, err := builder.Get()
				assert.ErrorContains(t, err, testCase.expectedErrorText)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition)
			assert.Equal(t, nodeFeatureRuleExampleName, builder.Definition.Name)
			assert.Equal(t, nodeFeatureRuleNamespace, builder.Definition.Namespace)
		})
	}
}

func TestPullNodeFeatureRule(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullFeatureRule, nfdv1.AddToScheme, nodeFeatureRuleGVK).ExecuteTests(t)
}

func TestNodeFeatureRuleBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[nfdv1.NodeFeatureRule, NodeFeatureRuleBuilder](
		nfdv1.AddToScheme,
		nodeFeatureRuleGVK,
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

func TestNodeFeatureRuleBuilderWithRule(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		rule          nfdv1.Rule
		validBuilder  bool
		expectedError string
	}{
		{
			name: "valid rule",
			rule: nfdv1.Rule{
				Name: "test-rule",
				Labels: map[string]string{
					"feature.node.kubernetes.io/test": "true",
				},
			},
			validBuilder: true,
		},
		{
			name:          "empty rule name",
			rule:          nfdv1.Rule{},
			validBuilder:  true,
			expectedError: ruleNameEmptyError,
		},
		{
			name:         "invalid builder",
			rule:         nfdv1.Rule{Name: "test-rule"},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme())
			if !testCase.validBuilder {
				builder = buildInvalidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme())
			}

			builder = builder.WithRule(testCase.rule)

			if !testCase.validBuilder {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.Rules)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)
				assert.Empty(t, builder.Definition.Spec.Rules)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, []nfdv1.Rule{testCase.rule}, builder.Definition.Spec.Rules)
		})
	}
}

func TestNodeFeatureRuleBuilderWithRules(t *testing.T) {
	t.Parallel()

	validRules := []nfdv1.Rule{
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
	}

	testCases := []struct {
		name          string
		rules         []nfdv1.Rule
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "valid rules",
			rules:        validRules,
			validBuilder: true,
		},
		{
			name:          "empty rules",
			rules:         []nfdv1.Rule{},
			validBuilder:  true,
			expectedError: "rules list cannot be empty",
		},
		{
			name:          "rule with empty name",
			rules:         []nfdv1.Rule{{Labels: map[string]string{"test": "true"}}},
			validBuilder:  true,
			expectedError: ruleNameEmptyError,
		},
		{
			name:         "invalid builder",
			rules:        validRules,
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme())
			if !testCase.validBuilder {
				builder = buildInvalidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme())
			}

			builder = builder.WithRules(testCase.rules)

			if !testCase.validBuilder {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
				assert.Empty(t, builder.Definition.Spec.Rules)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)
				assert.Empty(t, builder.Definition.Spec.Rules)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.rules, builder.Definition.Spec.Rules)
		})
	}
}

type simplePCIRuleTestCase struct {
	name         string
	ruleName     string
	labels       map[string]string
	vendorIDs    []string
	deviceIDs    []string
	validBuilder bool
	expectedErr  string
}

func TestNodeFeatureRuleBuilderWithSimplePCIRule(t *testing.T) {
	t.Parallel()

	testCases := []simplePCIRuleTestCase{
		{
			name:     "valid Neuron-like rule",
			ruleName: "neuron-device",
			labels: map[string]string{
				"feature.node.kubernetes.io/aws-neuron": "true",
			},
			vendorIDs:    []string{"1d0f"},
			deviceIDs:    []string{"7064", "7065", "7066", "7067"},
			validBuilder: true,
		},
		{
			name:     "valid rule without device IDs",
			ruleName: "vendor-only-rule",
			labels: map[string]string{
				"feature.node.kubernetes.io/vendor": "true",
			},
			vendorIDs:    []string{"1d0f"},
			deviceIDs:    []string{},
			validBuilder: true,
		},
		{
			name:     "empty rule name",
			ruleName: "",
			labels: map[string]string{
				"feature.node.kubernetes.io/test": "true",
			},
			vendorIDs:    []string{"1d0f"},
			deviceIDs:    []string{"7064"},
			validBuilder: true,
			expectedErr:  ruleNameEmptyError,
		},
		{
			name:         "empty labels",
			ruleName:     "test-rule",
			labels:       map[string]string{},
			vendorIDs:    []string{"1d0f"},
			deviceIDs:    []string{"7064"},
			validBuilder: true,
			expectedErr:  "rule 'labels' cannot be empty",
		},
		{
			name:     "empty vendor IDs",
			ruleName: "test-rule",
			labels: map[string]string{
				"feature.node.kubernetes.io/test": "true",
			},
			vendorIDs:    []string{},
			deviceIDs:    []string{"7064"},
			validBuilder: true,
			expectedErr:  "vendorIDs cannot be empty",
		},
		{
			name:         "invalid builder",
			ruleName:     "test-rule",
			labels:       map[string]string{"test": "true"},
			vendorIDs:    []string{"1d0f"},
			deviceIDs:    []string{"7064"},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runSimplePCIRuleTestCase(t, testCase)
		})
	}
}

func runSimplePCIRuleTestCase(t *testing.T, testCase simplePCIRuleTestCase) {
	t.Helper()

	builder := buildValidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme())
	if !testCase.validBuilder {
		builder = buildInvalidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme())
	}

	builder = builder.WithSimplePCIRule(
		testCase.ruleName,
		testCase.labels,
		testCase.vendorIDs,
		testCase.deviceIDs,
	)

	if !testCase.validBuilder {
		require.Error(t, builder.GetError())
		assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
		assert.Empty(t, builder.Definition.Spec.Rules)

		return
	}

	if testCase.expectedErr != "" {
		assert.EqualError(t, builder.GetError(), testCase.expectedErr)
		assert.Empty(t, builder.Definition.Spec.Rules)

		return
	}

	require.NoError(t, builder.GetError())
	require.Len(t, builder.Definition.Spec.Rules, 1)

	rule := builder.Definition.Spec.Rules[0]
	assert.Equal(t, testCase.ruleName, rule.Name)
	assert.Equal(t, testCase.labels, rule.Labels)
	require.Len(t, rule.MatchAny, 1)
	require.Len(t, rule.MatchAny[0].MatchFeatures, 1)
	assert.Equal(t, "pci.device", rule.MatchAny[0].MatchFeatures[0].Feature)

	require.NotNil(t, rule.MatchAny[0].MatchFeatures[0].MatchExpressions)
	matchExpressions := *rule.MatchAny[0].MatchFeatures[0].MatchExpressions
	assert.Equal(t, nfdv1.MatchIn, matchExpressions["vendor"].Op)
	assert.Equal(t, testCase.vendorIDs, []string(matchExpressions["vendor"].Value))

	if len(testCase.deviceIDs) == 0 {
		assert.NotContains(t, matchExpressions, "device")
	} else {
		assert.Equal(t, nfdv1.MatchIn, matchExpressions["device"].Op)
		assert.Equal(t, testCase.deviceIDs, []string(matchExpressions["device"].Value))
	}
}

func buildValidNFDRuleTestBuilder(apiClient *clients.Settings) *NodeFeatureRuleBuilder {
	return NewNodeFeatureRuleBuilder(apiClient, nodeFeatureRuleExampleName, nodeFeatureRuleNamespace)
}

func buildInvalidNFDRuleTestBuilder(apiClient *clients.Settings) *NodeFeatureRuleBuilder {
	return NewNodeFeatureRuleBuilder(apiClient, "", nodeFeatureRuleNamespace)
}

func buildTestClientWithNFDRuleScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{nfdv1.AddToScheme},
	})
}
