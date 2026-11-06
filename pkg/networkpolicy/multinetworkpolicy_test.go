package networkpolicy

import (
	"errors"
	"testing"

	"github.com/k8snetworkplumbingwg/multi-networkpolicy/pkg/apis/k8s.cni.cncf.io/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultMultiNetworkPolicyName      = "test-name"
	defaultMultiNetworkPolicyNamespace = "test-namespace"
)

var multiNetworkPolicyGVK = v1beta1.SchemeGroupVersion.WithKind("MultiNetworkPolicy")

func TestNewMultiNetworkPolicyBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewMultiNetworkPolicyBuilder,
		v1beta1.AddToScheme,
		multiNetworkPolicyGVK,
	).ExecuteTests(t)
}

func TestMultiNetworkPolicyPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullMultiNetworkPolicy,
		v1beta1.AddToScheme,
		multiNetworkPolicyGVK,
	).ExecuteTests(t)
}

func TestMultiNetworkPolicyBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[v1beta1.MultiNetworkPolicy, MultiNetworkPolicyBuilder](
		v1beta1.AddToScheme,
		multiNetworkPolicyGVK,
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

func TestMultiNetworkPolicyWithPodSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		podSelector metav1.LabelSelector
		valid       bool
	}{
		{
			name:        "valid pod selector",
			podSelector: metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			valid:       true,
		},
		{
			name:        "empty pod selector",
			podSelector: metav1.LabelSelector{},
			valid:       true,
		},
		{
			name:        "invalid builder short circuits",
			podSelector: metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			valid:       false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildMultiNetworkPolicyTestBuilder(testCase.valid)

			result := testBuilder.WithPodSelector(testCase.podSelector)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.PodSelector.MatchLabels)

				return
			}

			assert.NoError(t, result.GetError())
			assert.Equal(t, testCase.podSelector, result.Definition.Spec.PodSelector)
		})
	}
}

func TestMultiNetworkPolicyWithNetwork(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		networkName   string
		valid         bool
		expectedError error
	}{
		{
			name:        "valid network",
			networkName: "test-network",
			valid:       true,
		},
		{
			name:          "empty network",
			networkName:   "",
			valid:         true,
			expectedError: errors.New("The networkName is an empty string"),
		},
		{
			name:        "invalid builder short circuits",
			networkName: "test-network",
			valid:       false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildMultiNetworkPolicyTestBuilder(testCase.valid)

			result := testBuilder.WithNetwork(testCase.networkName)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Annotations)

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(
					t,
					testCase.networkName,
					result.Definition.Annotations["k8s.v1.cni.cncf.io/policy-for"],
				)
			} else {
				assert.Nil(t, result.Definition.Annotations)
			}
		})
	}
}

func TestMultiNetworkPolicyWithEmptyIngress(t *testing.T) {
	t.Parallel()

	t.Run("valid builder sets empty ingress", func(t *testing.T) {
		t.Parallel()

		result := buildMultiNetworkPolicyTestBuilder(true).WithEmptyIngress()

		assert.NoError(t, result.GetError())
		assert.NotNil(t, result.Definition.Spec.Ingress)
		assert.Empty(t, result.Definition.Spec.Ingress)
	})

	t.Run("invalid builder short circuits", func(t *testing.T) {
		t.Parallel()

		result := buildMultiNetworkPolicyTestBuilder(false).WithEmptyIngress()

		assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
		assert.Nil(t, result.Definition.Spec.Ingress)
	})
}

func TestMultiNetworkPolicyWithIngressRule(t *testing.T) {
	t.Parallel()

	peerRule := v1beta1.MultiNetworkPolicyIngressRule{
		From: []v1beta1.MultiNetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"test": "test"},
				},
			},
		},
	}

	t.Run("valid builder appends ingress rule", func(t *testing.T) {
		t.Parallel()

		result := buildMultiNetworkPolicyTestBuilder(true).WithIngressRule(peerRule)

		assert.NoError(t, result.GetError())
		assert.Equal(t, []v1beta1.MultiNetworkPolicyIngressRule{peerRule}, result.Definition.Spec.Ingress)
	})

	t.Run("invalid builder short circuits", func(t *testing.T) {
		t.Parallel()

		result := buildMultiNetworkPolicyTestBuilder(false).WithIngressRule(peerRule)

		assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
		assert.Empty(t, result.Definition.Spec.Ingress)
	})
}

func TestMultiNetworkPolicyWithEgressRule(t *testing.T) {
	t.Parallel()

	peerRule := v1beta1.MultiNetworkPolicyEgressRule{
		To: []v1beta1.MultiNetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"test": "test"},
				},
			},
		},
	}

	t.Run("valid builder appends egress rule", func(t *testing.T) {
		t.Parallel()

		result := buildMultiNetworkPolicyTestBuilder(true).WithEgressRule(peerRule)

		assert.NoError(t, result.GetError())
		assert.Equal(t, []v1beta1.MultiNetworkPolicyEgressRule{peerRule}, result.Definition.Spec.Egress)
	})

	t.Run("invalid builder short circuits", func(t *testing.T) {
		t.Parallel()

		result := buildMultiNetworkPolicyTestBuilder(false).WithEgressRule(peerRule)

		assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
		assert.Empty(t, result.Definition.Spec.Egress)
	})
}

func TestMultiNetworkPolicyWithPolicyType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		policyType    v1beta1.MultiPolicyType
		valid         bool
		expectedError error
	}{
		{
			name:       "valid policy type",
			policyType: v1beta1.PolicyTypeEgress,
			valid:      true,
		},
		{
			name:          "empty policy type",
			policyType:    "",
			valid:         true,
			expectedError: errors.New("The policy Type is an empty string"),
		},
		{
			name:       "invalid builder short circuits",
			policyType: v1beta1.PolicyTypeEgress,
			valid:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := buildMultiNetworkPolicyTestBuilder(testCase.valid).WithPolicyType(testCase.policyType)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.PolicyTypes)

				return
			}

			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, []v1beta1.MultiPolicyType{testCase.policyType}, result.Definition.Spec.PolicyTypes)
			} else {
				assert.Empty(t, result.Definition.Spec.PolicyTypes)
			}
		})
	}
}

func TestMultiNetworkPolicyGetMultiNetworkGVR(t *testing.T) {
	t.Parallel()

	result := GetMultiNetworkGVR()

	assert.Equal(t, "k8s.cni.cncf.io", result.Group)
	assert.Equal(t, "v1beta1", result.Version)
	assert.Equal(t, "multi-networkpolicies", result.Resource)
}

func buildMultiNetworkPolicyTestBuilder(valid bool) *MultiNetworkPolicyBuilder {
	if !valid {
		return NewMultiNetworkPolicyBuilder(newMultiNetworkPolicyTestClient(), "", defaultMultiNetworkPolicyNamespace)
	}

	return NewMultiNetworkPolicyBuilder(
		newMultiNetworkPolicyTestClient(),
		defaultMultiNetworkPolicyName,
		defaultMultiNetworkPolicyNamespace,
	)
}

func newMultiNetworkPolicyTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{v1beta1.AddToScheme},
	})
}
