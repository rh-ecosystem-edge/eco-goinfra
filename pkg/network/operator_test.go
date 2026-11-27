package network

import (
	"context"
	"fmt"
	"testing"
	"time"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var operatorNetworkGVK = operatorv1.GroupVersion.WithKind("Network")

func TestPullOperator(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		PullOperator,
		operatorv1.Install,
		operatorNetworkGVK,
		clusterNetworkName,
	).ExecuteTests(t)
}

func TestOperatorBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[operatorv1.Network, OperatorBuilder](
		operatorv1.Install, operatorNetworkGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestOperatorWaitUntilInCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		inCondition   bool
		expectedError error
	}{
		{
			name:        "condition found",
			exists:      true,
			inCondition: true,
		},
		{
			name:          "network operator does not exist",
			exists:        false,
			inCondition:   true,
			expectedError: fmt.Errorf("network.operator object %s does not exist", clusterNetworkName),
		},
		{
			name:          "condition not found before timeout",
			exists:        true,
			inCondition:   false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				networkOperator := buildDummyNetworkOperator()

				if testCase.inCondition {
					networkOperator.Status.Conditions = []operatorv1.OperatorCondition{{
						Type:   operatorv1.OperatorStatusTypeAvailable,
						Status: operatorv1.ConditionTrue,
					}}
				}

				runtimeObjects = append(runtimeObjects, networkOperator)
			}

			testBuilder := newOperatorBuilder(buildOperatorTestClient(runtimeObjects))

			err := testBuilder.WaitUntilInCondition(
				operatorv1.OperatorStatusTypeAvailable, time.Second, operatorv1.ConditionTrue)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestOperatorSetRouteAdvertisements(t *testing.T) {
	t.Parallel()

	// These tests exercise the idempotent path where the desired state already matches. The update path requires
	// mocked status changes and is covered by TestOperatorWaitUntilInCondition and the shared Update tests.
	testCases := []struct {
		name          string
		testBuilder   *OperatorBuilder
		state         operatorv1.RouteAdvertisementsEnablement
		ovnConfigNil  bool
		alreadySet    bool
		expectedError bool
	}{
		{
			name:         "state already set with nil OVN config",
			testBuilder:  newOperatorBuilder(buildOperatorTestClientWithDummyNetwork()),
			state:        operatorv1.RouteAdvertisementsEnabled,
			ovnConfigNil: true,
			alreadySet:   true,
		},
		{
			name:         "state already set with existing OVN config",
			testBuilder:  newOperatorBuilder(buildOperatorTestClientWithDummyNetwork()),
			state:        operatorv1.RouteAdvertisementsDisabled,
			ovnConfigNil: false,
			alreadySet:   true,
		},
		{
			name:          "network operator does not exist",
			testBuilder:   newOperatorBuilder(buildOperatorTestClient(nil)),
			state:         operatorv1.RouteAdvertisementsEnabled,
			ovnConfigNil:  true,
			alreadySet:    false,
			expectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if !testCase.ovnConfigNil {
				testCase.testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig = &operatorv1.OVNKubernetesConfig{}
			}

			if testCase.alreadySet {
				if testCase.testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig == nil {
					testCase.testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig = &operatorv1.OVNKubernetesConfig{}
				}

				testCase.testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.RouteAdvertisements = testCase.state
			}

			testBuilder, err := testCase.testBuilder.SetRouteAdvertisements(testCase.state, time.Second)

			if testCase.expectedError {
				require.Error(t, err)
				assert.True(t, commonerrors.IsAPICallFailedWithVerb(err, "get"))

				return
			}

			require.NoError(t, err)
			require.NotNil(t, testBuilder)
			require.NotNil(t, testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig)
			assert.Equal(t, testCase.state,
				testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.RouteAdvertisements)
		})
	}
}

func TestOperatorSetIPForwarding(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		testBuilder   *OperatorBuilder
		mode          operatorv1.IPForwardingMode
		alreadySet    bool
		expectedError bool
	}{
		{
			name:        "restricted mode already set",
			testBuilder: newOperatorBuilder(buildOperatorTestClientWithDummyNetwork()),
			mode:        operatorv1.IPForwardingRestricted,
			alreadySet:  true,
		},
		{
			name:        "global mode already set",
			testBuilder: newOperatorBuilder(buildOperatorTestClientWithDummyNetwork()),
			mode:        operatorv1.IPForwardingGlobal,
			alreadySet:  true,
		},
		{
			name:          "network operator does not exist",
			testBuilder:   newOperatorBuilder(buildOperatorTestClient(nil)),
			mode:          operatorv1.IPForwardingRestricted,
			alreadySet:    false,
			expectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.alreadySet {
				testCase.testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig = &operatorv1.OVNKubernetesConfig{
					GatewayConfig: &operatorv1.GatewayConfig{
						IPForwarding: testCase.mode,
					},
				}
			}

			testBuilder, err := testCase.testBuilder.SetIPForwarding(testCase.mode, time.Second)

			if testCase.expectedError {
				require.Error(t, err)
				assert.True(t, commonerrors.IsAPICallFailedWithVerb(err, "get"))

				return
			}

			require.NoError(t, err)
			require.NotNil(t, testBuilder)
			require.NotNil(t, testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig)
			require.NotNil(t, testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.GatewayConfig)
			assert.Equal(t, testCase.mode,
				testBuilder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.GatewayConfig.IPForwarding)
		})
	}
}

// buildDummyNetworkOperator builds a dummy network.operator object. It uses the clusterNetworkName.
func buildDummyNetworkOperator() *operatorv1.Network {
	return &operatorv1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterNetworkName,
		},
	}
}

func buildOperatorTestClient(objects []runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{operatorv1.Install},
	})
}

func buildOperatorTestClientWithDummyNetwork() *clients.Settings {
	return buildOperatorTestClient([]runtime.Object{buildDummyNetworkOperator()})
}

func newOperatorBuilder(apiClient *clients.Settings) *OperatorBuilder {
	return common.NewClusterScopedBuilder[operatorv1.Network, OperatorBuilder](
		apiClient, operatorv1.Install, clusterNetworkName)
}
