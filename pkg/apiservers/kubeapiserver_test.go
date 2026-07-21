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

var kubeAPIServerGVK = operatorv1.GroupVersion.WithKind("KubeAPIServer")

func TestPullKubeAPIServer(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		PullKubeAPIServer,
		operatorv1.Install,
		kubeAPIServerGVK,
		kubeAPIServerObjName,
	).ExecuteTests(t)
}

func TestKubeAPIServerBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[operatorv1.KubeAPIServer, KubeAPIServerBuilder](
		operatorv1.Install, kubeAPIServerGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		Run(t)
}

func TestKubeAPIServerGetCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testKubeAPIServerBuilder *KubeAPIServerBuilder
		condition                string
		conditionStatus          operatorv1.ConditionStatus
		expectedError            error
	}{
		{
			condition:                conditionTypeNodeInstallerProgressing,
			conditionStatus:          operatorv1.ConditionTrue,
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            nil,
		},
		{
			condition:                "Unavailable",
			conditionStatus:          "",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("the cluster kubeAPIServer Unavailable condition not found"),
		},
		{
			condition:                "",
			conditionStatus:          "",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("kubeAPIServer 'conditionType' cannot be empty"),
		},
		{
			condition:                conditionTypeNodeInstallerProgressing,
			conditionStatus:          operatorv1.ConditionTrue,
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(newKubeAPIServerTestClient()),
			expectedError:            fmt.Errorf("cluster kubeAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.condition, func(t *testing.T) {
			t.Parallel()

			status, msg, err := testCase.testKubeAPIServerBuilder.GetCondition(testCase.condition)
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

func TestKubeAPIServerWaitUntilConditionTrue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                     string
		testKubeAPIServerBuilder *KubeAPIServerBuilder
		condition                string
		expectedError            error
	}{
		{
			name:                     "condition becomes true",
			condition:                conditionTypeNodeInstallerProgressing,
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            nil,
		},
		{
			name:                     "unknown condition times out",
			condition:                "unavailable",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("the unavailable condition not found exists: context deadline exceeded"),
		},
		{
			name:                     "empty condition type",
			condition:                "",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("kubeAPIServer 'conditionType' cannot be empty"),
		},
		{
			name:                     "resource not found",
			condition:                conditionTypeNodeInstallerProgressing,
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(newKubeAPIServerTestClient()),
			expectedError:            fmt.Errorf("cluster kubeAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.testKubeAPIServerBuilder.WaitUntilConditionTrue(testCase.condition, 1*time.Second)
			if testCase.expectedError != nil {
				require.EqualError(t, err, testCase.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKubeAPIServerWaitAllNodesAtTheLatestRevision(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                     string
		testKubeAPIServerBuilder *KubeAPIServerBuilder
		expectedError            error
	}{
		{
			name:                     "all nodes at latest revision",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            nil,
		},
		{
			name:                     "resource not found",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(newKubeAPIServerTestClient()),
			expectedError:            fmt.Errorf("cluster kubeAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.testKubeAPIServerBuilder.WaitAllNodesAtTheLatestRevision(1 * time.Second)
			if testCase.expectedError != nil {
				require.EqualError(t, err, testCase.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newKubeAPIServerTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{operatorv1.Install},
	})
}

func buildValidKubeAPIServerBuilder(apiClient *clients.Settings) *KubeAPIServerBuilder {
	return common.NewClusterScopedBuilder[operatorv1.KubeAPIServer, KubeAPIServerBuilder](
		apiClient, operatorv1.Install, kubeAPIServerObjName)
}

func buildKubeAPIServerWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: append([]runtime.Object{}, &operatorv1.KubeAPIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:            kubeAPIServerObjName,
				ResourceVersion: "999",
			},
			Spec: operatorv1.KubeAPIServerSpec{},
			Status: operatorv1.KubeAPIServerStatus{
				StaticPodOperatorStatus: operatorv1.StaticPodOperatorStatus{
					OperatorStatus: operatorv1.OperatorStatus{
						Conditions: []operatorv1.OperatorCondition{
							{Type: conditionTypeNodeInstallerProgressing, Status: operatorv1.ConditionTrue, Reason: conditionReasonAllNodesAtLatestRevision}},
					},
				},
			},
		}),
		SchemeAttachers: []clients.SchemeAttacher{operatorv1.Install},
	})
}
