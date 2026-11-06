package nmstate

import (
	"testing"

	nmstatev1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	v1TestSchemes = []clients.SchemeAttacher{
		nmstatev1.AddToScheme,
	}
	defaultNMStateName = "nmstatename"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		NewBuilder, nmstatev1.AddToScheme, nmStateGVK).ExecuteTests(t)
}

func TestPullNMstate(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullNMstate, nmstatev1.AddToScheme, nmStateGVK).ExecuteTests(t)
}

func TestNMStateBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[nmstatev1.NMState, Builder](
		nmstatev1.AddToScheme,
		nmStateGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

// buildTestClientWithDummyObject returns a test client containing the default NMState object.
func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyNMStateObject(),
		SchemeAttachers: v1TestSchemes,
	})
}

// buildDummyNMStateObject returns the NMState objects used by the shared NMState package tests.
func buildDummyNMStateObject() []runtime.Object {
	return append([]runtime.Object{}, &nmstatev1.NMState{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultNMStateName,
		},
		Spec: nmstatev1.NMStateSpec{},
	})
}
