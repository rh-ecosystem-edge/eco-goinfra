package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultServiceBGPStatusName = "servicebgpstatus-0"
	serviceBGPStatusGVK         = mlbtypes.GroupVersion.WithKind("ServiceBGPStatus")
	serviceBGPStatusTestSchemes = []clients.SchemeAttacher{mlbtypes.AddToScheme}
)

func TestPullServiceBGPStatus(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullServiceBGPStatus,
		mlbtypes.AddToScheme,
		serviceBGPStatusGVK,
	).ExecuteTests(t)
}

func TestServiceBGPStatusMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newServiceBGPStatusCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}

func newServiceBGPStatusCommonTestConfig() testhelper.CommonTestConfig[
	mlbtypes.ServiceBGPStatus, ServiceBGPStatusBuilder, *mlbtypes.ServiceBGPStatus, *ServiceBGPStatusBuilder] {
	return testhelper.NewCommonTestConfig[mlbtypes.ServiceBGPStatus, ServiceBGPStatusBuilder](
		mlbtypes.AddToScheme,
		serviceBGPStatusGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

// buildDummyServiceBGPStatus returns a ServiceBGPStatus with the provided name.
func buildDummyServiceBGPStatus(name string) *mlbtypes.ServiceBGPStatus {
	return &mlbtypes.ServiceBGPStatus{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestServiceBGPStatusClientWithDummyStatus returns a client with a dummy ServiceBGPStatus.
func buildTestServiceBGPStatusClientWithDummyStatus(statusName string) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyServiceBGPStatus(statusName)},
		SchemeAttachers: serviceBGPStatusTestSchemes,
	})
}
