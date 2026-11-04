package apiservers

import (
	"fmt"
	"testing"
	"time"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var openshiftAPIServerGVK = operatorv1.GroupVersion.WithKind("OpenShiftAPIServer")

func TestPullOpenshiftAPIServer(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		PullOpenshiftAPIServer,
		operatorv1.Install,
		openshiftAPIServerGVK,
		openshiftAPIServerObjName,
	).ExecuteTests(t)
}

func TestOpenshiftAPIServerBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[operatorv1.OpenShiftAPIServer, OpenshiftAPIServerBuilder](
		operatorv1.Install, openshiftAPIServerGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		Run(t)
}

func TestOpenshiftAPIServerGetCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testOpenshiftAPIServerBuilder *OpenshiftAPIServerBuilder
		condition                     string
		conditionStatus               operatorv1.ConditionStatus
		expectedError                 error
	}{
		{
			condition:                     conditionTypeAPIServerDeploymentProgressing,
			conditionStatus:               operatorv1.ConditionTrue,
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 nil,
		},
		{
			condition:                     "Unavailable",
			conditionStatus:               "",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 fmt.Errorf("the cluster openshiftAPIServer Unavailable condition not found"),
		},
		{
			condition:                     "",
			conditionStatus:               "",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 fmt.Errorf("openshiftAPIServer 'conditionType' cannot be empty"),
		},
		{
			condition:       conditionTypeAPIServerDeploymentProgressing,
			conditionStatus: operatorv1.ConditionTrue,
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				newOpenshiftAPIServerTestClient()),
			expectedError: fmt.Errorf("cluster openshiftAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.condition, func(t *testing.T) {
			t.Parallel()

			status, msg, err := testCase.testOpenshiftAPIServerBuilder.GetCondition(testCase.condition)
			assert.Equal(t, testCase.expectedError, err)

			if err == nil {
				assert.Equal(t, *status, testCase.conditionStatus)
			} else {
				assert.Nil(t, status)
				assert.Equal(t, "", msg)
			}
		})
	}
}

func TestOpenshiftAPIServerWaitUntilConditionTrue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                          string
		testOpenshiftAPIServerBuilder *OpenshiftAPIServerBuilder
		condition                     string
		expectedError                 error
	}{
		{
			name:      "condition becomes true",
			condition: conditionTypeAPIServerDeploymentProgressing,
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError: nil,
		},
		{
			name:      "unknown condition times out",
			condition: "Unavailable",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError: fmt.Errorf("the Unavailable condition not found exists: context deadline exceeded"),
		},
		{
			name:      "empty condition type",
			condition: "",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError: fmt.Errorf("openshiftAPIServer 'conditionType' cannot be empty"),
		},
		{
			name:      "resource not found",
			condition: conditionTypeAPIServerDeploymentProgressing,
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				newOpenshiftAPIServerTestClient()),
			expectedError: fmt.Errorf("cluster openshiftAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.testOpenshiftAPIServerBuilder.WaitUntilConditionTrue(testCase.condition, 1*time.Second)
			if testCase.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOpenshiftAPIServerWaitAllPodsAtTheLatestGeneration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                          string
		testOpenshiftAPIServerBuilder *OpenshiftAPIServerBuilder
		expectedError                 error
	}{
		{
			name:                          "all pods at latest generation",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 nil,
		},
		{
			name: "openshift apiserver not found",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				newOpenshiftAPIServerTestClient()),
			expectedError: fmt.Errorf("cluster openshiftAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.testOpenshiftAPIServerBuilder.WaitAllPodsAtTheLatestGeneration(1 * time.Second)
			if testCase.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func newOpenshiftAPIServerTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{operatorv1.Install},
	})
}

func buildValidOpenshiftAPIServerBuilder(apiClient *clients.Settings) *OpenshiftAPIServerBuilder {
	return common.NewClusterScopedBuilder[operatorv1.OpenShiftAPIServer, OpenshiftAPIServerBuilder](
		apiClient, operatorv1.Install, openshiftAPIServerObjName)
}

func buildOpenshiftAPIServerBuilderWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: append([]runtime.Object{}, &operatorv1.OpenShiftAPIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: openshiftAPIServerObjName,
			},
			Spec: operatorv1.OpenShiftAPIServerSpec{},
			Status: operatorv1.OpenShiftAPIServerStatus{
				OperatorStatus: operatorv1.OperatorStatus{
					Conditions: []operatorv1.OperatorCondition{
						{Type: conditionTypeAPIServerDeploymentProgressing, Status: operatorv1.ConditionTrue, Reason: conditionReasonAsExpected}},
				},
			},
		}),
		SchemeAttachers: []clients.SchemeAttacher{operatorv1.Install},
	})
}
