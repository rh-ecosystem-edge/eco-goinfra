package networkpolicy

import (
	"errors"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
)

const (
	defaultNetworkPolicyName      = "test-name"
	defaultNetworkPolicyNamespace = "test-namespace"
)

var networkPolicyGVK = netv1.SchemeGroupVersion.WithKind("NetworkPolicy")

func TestNewNetworkPolicyBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewNetworkPolicyBuilder,
		netv1.AddToScheme,
		networkPolicyGVK,
	).ExecuteTests(t)
}

func TestNetworkPolicyPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull,
		netv1.AddToScheme,
		networkPolicyGVK,
	).ExecuteTests(t)
}

func TestNetworkPolicyBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[netv1.NetworkPolicy, NetworkPolicyBuilder](
		netv1.AddToScheme,
		networkPolicyGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestNetworkPolicyWithNamespaceIngressRule(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		testNamespaceIngress map[string]string
		testPodIngress       map[string]string
		valid                bool
		expectedError        error
	}{
		{
			name:                 "pod selector only",
			testNamespaceIngress: map[string]string{},
			testPodIngress:       map[string]string{"test": "test"},
			valid:                true,
		},
		{
			name:                 "namespace selector only",
			testNamespaceIngress: map[string]string{"test": "test"},
			testPodIngress:       map[string]string{},
			valid:                true,
		},
		{
			name:                 "both selectors",
			testNamespaceIngress: map[string]string{"namespace": "test"},
			testPodIngress:       map[string]string{"pod": "test"},
			valid:                true,
		},
		{
			name:                 "empty selectors",
			testNamespaceIngress: map[string]string{},
			testPodIngress:       map[string]string{},
			valid:                true,
			expectedError: errors.New(
				"Both namespaceIngressMatchLabels and podIngressMatchLabels parameters are empty maps"),
		},
		{
			name:                 "invalid builder short circuits",
			testNamespaceIngress: map[string]string{"namespace": "test"},
			testPodIngress:       map[string]string{"pod": "test"},
			valid:                false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildNetworkPolicyTestBuilder(testCase.valid)

			result := testBuilder.WithNamespaceIngressRule(
				testCase.testNamespaceIngress,
				testCase.testPodIngress,
			)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.Ingress)

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError != nil {
				assert.Empty(t, result.Definition.Spec.Ingress)

				return
			}

			assert.Len(t, result.Definition.Spec.Ingress, 1)
			assert.Len(t, result.Definition.Spec.Ingress[0].From, 1)

			peer := result.Definition.Spec.Ingress[0].From[0]
			if len(testCase.testNamespaceIngress) != 0 {
				assert.Equal(t, testCase.testNamespaceIngress, peer.NamespaceSelector.MatchLabels)
			} else {
				assert.Nil(t, peer.NamespaceSelector)
			}

			if len(testCase.testPodIngress) != 0 {
				assert.Equal(t, testCase.testPodIngress, peer.PodSelector.MatchLabels)
			} else {
				assert.Nil(t, peer.PodSelector)
			}
		})
	}
}

func TestNetworkPolicyWithPolicyType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		testPolicyType netv1.PolicyType
		valid          bool
		expectedError  error
	}{
		{
			name:           "valid policy type",
			testPolicyType: netv1.PolicyTypeIngress,
			valid:          true,
		},
		{
			name:           "empty policy type",
			testPolicyType: "",
			valid:          true,
			expectedError:  errors.New("The policyType is an empty string"),
		},
		{
			name:           "invalid builder short circuits",
			testPolicyType: netv1.PolicyTypeIngress,
			valid:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildNetworkPolicyTestBuilder(testCase.valid)

			result := testBuilder.WithPolicyType(testCase.testPolicyType)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.PolicyTypes)

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, []netv1.PolicyType{testCase.testPolicyType}, result.Definition.Spec.PolicyTypes)
			} else {
				assert.Empty(t, result.Definition.Spec.PolicyTypes)
			}
		})
	}
}

func TestNetworkPolicyWithPodSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		testPodSelector map[string]string
		valid           bool
		expectedError   error
	}{
		{
			name:            "valid pod selector",
			testPodSelector: map[string]string{"test": "test"},
			valid:           true,
		},
		{
			name:            "empty pod selector",
			testPodSelector: map[string]string{},
			valid:           true,
			expectedError:   errors.New("The podSelector is an empty string"),
		},
		{
			name:            "invalid builder short circuits",
			testPodSelector: map[string]string{"test": "test"},
			valid:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildNetworkPolicyTestBuilder(testCase.valid)

			result := testBuilder.WithPodSelector(testCase.testPodSelector)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.PodSelector.MatchLabels)

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.testPodSelector, result.Definition.Spec.PodSelector.MatchLabels)
			} else {
				assert.Nil(t, result.Definition.Spec.PodSelector.MatchLabels)
			}
		})
	}
}

func TestNetworkPolicyGetGVR(t *testing.T) {
	t.Parallel()

	result := GetGVR()

	assert.Equal(t, "networking.k8s.io", result.Group)
	assert.Equal(t, "v1", result.Version)
	assert.Equal(t, "networkpolicies", result.Resource)
}

func buildNetworkPolicyTestBuilder(valid bool) *NetworkPolicyBuilder {
	if !valid {
		return NewNetworkPolicyBuilder(newNetworkPolicyTestClient(), "", defaultNetworkPolicyNamespace)
	}

	return NewNetworkPolicyBuilder(
		newNetworkPolicyTestClient(),
		defaultNetworkPolicyName,
		defaultNetworkPolicyNamespace,
	)
}

func newNetworkPolicyTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{netv1.AddToScheme},
	})
}
