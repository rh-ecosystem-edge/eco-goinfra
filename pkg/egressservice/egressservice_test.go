package egressservice

import (
	"testing"

	egresssvcv1 "github.com/ovn-kubernetes/ovn-kubernetes/go-controller/pkg/crd/egressservice/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultEgressServiceName      = "demo-egress-svc"
	defaultEgressServiceNamespace = "demo-ns"
)

func TestNewEgressServiceBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig[egresssvcv1.EgressService, EgressServiceBuilder](
			func(apiClient *clients.Settings, name, nsname string) *EgressServiceBuilder {
				return NewEgressServiceBuilder(apiClient, name, nsname, string(egresssvcv1.SourceIPLoadBalancer))
			},
			egresssvcv1.AddToScheme,
			egressServiceGVK,
		).ExecuteTests(t)
	})

	t.Run("source IP mode validation", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name          string
			sourceIPBy    string
			expectedValue string
			expectedError string
		}{
			{
				name:          "load balancer source IP mode is set",
				sourceIPBy:    string(egresssvcv1.SourceIPLoadBalancer),
				expectedValue: string(egresssvcv1.SourceIPLoadBalancer),
			},
			{
				name:          "network source IP mode is set",
				sourceIPBy:    string(egresssvcv1.SourceIPNetwork),
				expectedValue: string(egresssvcv1.SourceIPNetwork),
			},
			{
				name:          "case-insensitive source IP mode is preserved",
				sourceIPBy:    "  loadbalancerip  ",
				expectedValue: "  loadbalancerip  ",
			},
			{
				name:          "empty source IP mode returns error",
				sourceIPBy:    "",
				expectedError: errInvalidSourceIPBy,
			},
			{
				name:          "invalid source IP mode returns error",
				sourceIPBy:    "DemoByIp",
				expectedError: errInvalidSourceIPBy,
			},
		}

		for _, testCase := range testCases {
			testCase := testCase

			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				builder := NewEgressServiceBuilder(
					getTestEgressServiceAPIClient(), defaultEgressServiceName, defaultEgressServiceNamespace, testCase.sourceIPBy)

				if testCase.expectedError != "" {
					require.EqualError(t, builder.GetError(), testCase.expectedError)

					return
				}

				require.NoError(t, builder.GetError())
				assert.Equal(t, egresssvcv1.SourceIPMode(testCase.expectedValue), builder.Definition.Spec.SourceIPBy)
			})
		}
	})
}

func TestEgressServicePull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[egresssvcv1.EgressService, EgressServiceBuilder](
		Pull,
		egresssvcv1.AddToScheme,
		egressServiceGVK,
	).ExecuteTests(t)
}

func TestEgressServiceBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newEgressServiceCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestWithNodeLabelSelector(t *testing.T) {
	t.Parallel()

	selector := map[string]string{"egress-svc": "true", "prod": ""}

	testEgressServiceModifier(t, func(builder *EgressServiceBuilder) *EgressServiceBuilder {
		return builder.WithNodeLabelSelector(selector)
	}, func(t *testing.T, builder *EgressServiceBuilder) {
		t.Helper()

		assert.Equal(t, metav1.LabelSelector{MatchLabels: selector}, builder.Definition.Spec.NodeSelector)
	})

	t.Run("empty selector returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidEgressServiceTestBuilder()
		result := builder.WithNodeLabelSelector(map[string]string{})

		require.Same(t, builder, result)
		require.EqualError(t, result.GetError(), "cannot accept empty map as nodeSelector")
		assert.Empty(t, result.Definition.Spec.NodeSelector)
	})
}

func TestWithVRFNetwork(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		vrfNetwork string
	}{
		{
			name:       "numeric VRF network is set",
			vrfNetwork: "1001",
		},
		{
			name:       "named VRF network is set",
			vrfNetwork: "vrfName",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testEgressServiceModifier(t, func(builder *EgressServiceBuilder) *EgressServiceBuilder {
				return builder.WithVRFNetwork(testCase.vrfNetwork)
			}, func(t *testing.T, builder *EgressServiceBuilder) {
				t.Helper()

				assert.Equal(t, testCase.vrfNetwork, builder.Definition.Spec.Network)
			})
		})
	}

	t.Run("empty VRF network returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidEgressServiceTestBuilder()
		result := builder.WithVRFNetwork(" ")

		require.Same(t, builder, result)
		require.EqualError(t, result.GetError(), "cannot use empty VRF network")
		assert.Empty(t, result.Definition.Spec.Network)
	})
}

func testEgressServiceModifier(
	t *testing.T,
	modify func(*EgressServiceBuilder) *EgressServiceBuilder,
	verify func(*testing.T, *EgressServiceBuilder),
) {
	t.Helper()

	t.Run("applies to valid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildValidEgressServiceTestBuilder()
		result := modify(builder)

		require.Same(t, builder, result)
		require.NoError(t, result.GetError())
		verify(t, result)
	})

	t.Run("invalid builder short circuits", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidEgressServiceTestBuilder()
		before := builder.Definition.DeepCopy()
		result := modify(builder)

		require.Same(t, builder, result)
		require.Truef(t, commonerrors.IsBuilderNameEmpty(result.GetError()), "unexpected error: %v", result.GetError())
		assert.Equal(t, before, result.Definition)
	})
}

func newEgressServiceCommonTestConfig() testhelper.CommonTestConfig[
	egresssvcv1.EgressService,
	EgressServiceBuilder,
	*egresssvcv1.EgressService,
	*EgressServiceBuilder,
] {
	return testhelper.NewCommonTestConfig[egresssvcv1.EgressService, EgressServiceBuilder](
		egresssvcv1.AddToScheme,
		egressServiceGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

func buildValidEgressServiceTestBuilder() *EgressServiceBuilder {
	return NewEgressServiceBuilder(
		getTestEgressServiceAPIClient(),
		defaultEgressServiceName,
		defaultEgressServiceNamespace,
		string(egresssvcv1.SourceIPLoadBalancer),
	)
}

func buildInvalidEgressServiceTestBuilder() *EgressServiceBuilder {
	return NewEgressServiceBuilder(
		getTestEgressServiceAPIClient(),
		"",
		defaultEgressServiceNamespace,
		string(egresssvcv1.SourceIPLoadBalancer),
	)
}

func getTestEgressServiceAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}
