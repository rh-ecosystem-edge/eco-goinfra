package egressip

import (
	"fmt"
	"testing"

	egressipv1 "github.com/ovn-kubernetes/ovn-kubernetes/go-controller/pkg/crd/egressip/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultEgressIPName            = "egress-test"
	defaultEgressIPs               = []string{"1.1.1.2", "1.1.1.3"}
	defaultEgressNamespaceSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{"env": "qa"},
	}
	defaultEgressPodSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{"env": "qa"},
	}
)

func TestNewEgressIPBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig[egressipv1.EgressIP, EgressIPBuilder](
		NewEgressIPBuilder,
		egressipv1.AddToScheme,
		egressIPGVK,
	).ExecuteTests(t)
}

func TestEgressIPPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig[egressipv1.EgressIP, EgressIPBuilder](
		Pull,
		egressipv1.AddToScheme,
		egressIPGVK,
	).ExecuteTests(t)
}

func TestEgressIPBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newEgressIPCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestEgressIPWithEgressIPs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		builder                func() *EgressIPBuilder
		egressIPs              []string
		expectedEgressIPs      []string
		expectedError          string
		expectBuilderNameError bool
	}{
		{
			name:          "empty list returns error",
			builder:       buildValidEgressIPTestBuilder,
			egressIPs:     []string{},
			expectedError: "cannot accept empty list as egressIPs value",
		},
		{
			name:              "single egress IP is set",
			builder:           buildValidEgressIPTestBuilder,
			egressIPs:         []string{"10.10.11.2"},
			expectedEgressIPs: []string{"10.10.11.2"},
		},
		{
			name:              "multiple egress IPs are set",
			builder:           buildValidEgressIPTestBuilder,
			egressIPs:         []string{"10.10.11.2", "10.10.11.3", "10.10.11.4"},
			expectedEgressIPs: []string{"10.10.11.2", "10.10.11.3", "10.10.11.4"},
		},
		{
			name:                   "invalid builder short circuits",
			builder:                buildInvalidEgressIPTestBuilder,
			egressIPs:              defaultEgressIPs,
			expectBuilderNameError: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithEgressIPs(testCase.egressIPs)
			require.Same(t, builder, result)

			if testCase.expectBuilderNameError {
				require.Truef(t, commonerrors.IsBuilderNameEmpty(result.GetError()), "unexpected error: %v", result.GetError())
				assert.Nil(t, result.Definition.Spec.EgressIPs)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.expectedEgressIPs, result.Definition.Spec.EgressIPs)
		})
	}
}

func TestEgressIPWithNamespaceSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		builder                func() *EgressIPBuilder
		namespaceSelector      metav1.LabelSelector
		expectedSelector       metav1.LabelSelector
		expectedError          string
		expectBuilderNameError bool
	}{
		{
			name:              "namespace selector is set",
			builder:           buildValidEgressIPTestBuilder,
			namespaceSelector: defaultEgressNamespaceSelector,
			expectedSelector:  defaultEgressNamespaceSelector,
		},
		{
			name:    "match expressions are set",
			builder: buildValidEgressIPTestBuilder,
			namespaceSelector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{
				Key:      "environment",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"production"},
			}}},
			expectedSelector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{
				Key:      "environment",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"production"},
			}}},
		},
		{
			name:              "empty selector returns error",
			builder:           buildValidEgressIPTestBuilder,
			namespaceSelector: metav1.LabelSelector{},
			expectedError:     "EgressIP 'namespaceSelector' cannot be empty",
		},
		{
			name:                   "invalid builder short circuits",
			builder:                buildInvalidEgressIPTestBuilder,
			namespaceSelector:      defaultEgressNamespaceSelector,
			expectBuilderNameError: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithNamespaceSelector(testCase.namespaceSelector)
			require.Same(t, builder, result)

			if testCase.expectBuilderNameError {
				require.Truef(t, commonerrors.IsBuilderNameEmpty(result.GetError()), "unexpected error: %v", result.GetError())
				assert.Empty(t, result.Definition.Spec.NamespaceSelector)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.expectedSelector, result.Definition.Spec.NamespaceSelector)
		})
	}
}

func TestEgressIPWithPodSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		builder                func() *EgressIPBuilder
		podSelector            metav1.LabelSelector
		expectedSelector       metav1.LabelSelector
		expectedError          string
		expectBuilderNameError bool
	}{
		{
			name:             "pod selector is set",
			builder:          buildValidEgressIPTestBuilder,
			podSelector:      defaultEgressPodSelector,
			expectedSelector: defaultEgressPodSelector,
		},
		{
			name:          "empty selector returns error",
			builder:       buildValidEgressIPTestBuilder,
			podSelector:   metav1.LabelSelector{},
			expectedError: "EgressIP 'podSelector' cannot be empty",
		},
		{
			name:                   "invalid builder short circuits",
			builder:                buildInvalidEgressIPTestBuilder,
			podSelector:            defaultEgressPodSelector,
			expectBuilderNameError: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithPodSelector(testCase.podSelector)
			require.Same(t, builder, result)

			if testCase.expectBuilderNameError {
				require.Truef(t, commonerrors.IsBuilderNameEmpty(result.GetError()), "unexpected error: %v", result.GetError())
				assert.Empty(t, result.Definition.Spec.PodSelector)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.expectedSelector, result.Definition.Spec.PodSelector)
		})
	}
}

func TestEgressIPGetAssignedEgressIPMap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		object                 *egressipv1.EgressIP
		invalidBuilder         bool
		expectedMap            map[string]string
		expectedError          error
		expectBuilderNameError bool
	}{
		{
			name: "single assigned EgressIP is returned",
			object: buildEgressIPWithStatus([]egressipv1.EgressIPStatusItem{{
				Node:     "node-1",
				EgressIP: "1.1.1.2",
			}}),
			expectedMap: map[string]string{"node-1": "1.1.1.2"},
		},
		{
			name: "multiple assigned EgressIPs are returned",
			object: buildEgressIPWithStatus([]egressipv1.EgressIPStatusItem{
				{Node: "node-1", EgressIP: "1.1.1.2"},
				{Node: "node-2", EgressIP: "1.1.1.3"},
			}),
			expectedMap: map[string]string{"node-1": "1.1.1.2", "node-2": "1.1.1.3"},
		},
		{
			name:          "missing assignments returns error",
			object:        buildEgressIPWithStatus(nil),
			expectedError: fmt.Errorf("egressIP %q nodes assignment does not exist", defaultEgressIPName),
		},
		{
			name:          "missing EgressIP returns error",
			expectedError: fmt.Errorf("egressIP %q object does not exist", defaultEgressIPName),
		},
		{
			name:                   "invalid builder returns error",
			object:                 buildEgressIPWithStatus(nil),
			invalidBuilder:         true,
			expectBuilderNameError: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object
			if testCase.object != nil {
				runtimeObjects = []runtime.Object{testCase.object}
			}

			apiClient := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{egressipv1.AddToScheme},
			})

			builder := buildValidEgressIPTestBuilderWithClient(apiClient)
			if testCase.invalidBuilder {
				builder = buildInvalidEgressIPTestBuilderWithClient(apiClient)
			}

			result, err := builder.GetAssignedEgressIPMap()

			if testCase.expectBuilderNameError {
				require.Truef(t, commonerrors.IsBuilderNameEmpty(err), "unexpected error: %v", err)
				assert.Nil(t, result)

				return
			}

			if testCase.expectedError != nil {
				require.EqualError(t, err, testCase.expectedError.Error())
				assert.Nil(t, result)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.expectedMap, result)
		})
	}
}

func newEgressIPCommonTestConfig() testhelper.CommonTestConfig[
	egressipv1.EgressIP,
	EgressIPBuilder,
	*egressipv1.EgressIP,
	*EgressIPBuilder,
] {
	return testhelper.NewCommonTestConfig[egressipv1.EgressIP, EgressIPBuilder](
		egressipv1.AddToScheme,
		egressIPGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

func getTestEgressIPAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}

func buildValidEgressIPTestBuilder() *EgressIPBuilder {
	return buildValidEgressIPTestBuilderWithClient(getTestEgressIPAPIClient())
}

func buildValidEgressIPTestBuilderWithClient(apiClient *clients.Settings) *EgressIPBuilder {
	return NewEgressIPBuilder(apiClient, defaultEgressIPName)
}

func buildInvalidEgressIPTestBuilder() *EgressIPBuilder {
	return buildInvalidEgressIPTestBuilderWithClient(getTestEgressIPAPIClient())
}

func buildInvalidEgressIPTestBuilderWithClient(apiClient *clients.Settings) *EgressIPBuilder {
	return NewEgressIPBuilder(apiClient, "")
}

func buildEgressIPWithStatus(items []egressipv1.EgressIPStatusItem) *egressipv1.EgressIP {
	return &egressipv1.EgressIP{
		ObjectMeta: metav1.ObjectMeta{Name: defaultEgressIPName},
		Status:     egressipv1.EgressIPStatus{Items: items},
	}
}
