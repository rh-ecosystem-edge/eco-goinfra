package sriovdra

import (
	"testing"

	drav1 "github.com/k8snetworkplumbingwg/dra-driver-sriov/pkg/api/sriovdra/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestPullDeviceAttributes(t *testing.T) {
	t.Parallel()
	testhelper.NewNamespacedPullTestConfig(PullDeviceAttributes, drav1.AddToScheme,
		drav1.SchemeGroupVersion.WithKind("DeviceAttributes")).ExecuteTests(t)
}

func TestListDeviceAttributes(t *testing.T) {
	t.Parallel()
	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, ns string, _ ...runtimeclient.ListOptions) ([]*DeviceAttributesBuilder, error) {
			return ListDeviceAttributes(apiClient, ns)
		}, drav1.AddToScheme, drav1.SchemeGroupVersion.WithKind("DeviceAttributes")).ExecuteTests(t)
}

func TestDeviceAttributesReadMethods(t *testing.T) {
	t.Parallel()

	config := testhelper.NewCommonTestConfig[drav1.DeviceAttributes, DeviceAttributesBuilder](drav1.AddToScheme,
		drav1.SchemeGroupVersion.WithKind("DeviceAttributes"), testhelper.ResourceScopeNamespaced)
	testhelper.NewTestSuite().With(testhelper.NewGetTestConfig(config)).
		With(testhelper.NewExistsTestConfig(config)).Run(t)
}

func TestListDeviceAttributesFilters(t *testing.T) {
	apiClient := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			&drav1.DeviceAttributes{ObjectMeta: metav1.ObjectMeta{
				Name: "selected", Namespace: "operator", Labels: map[string]string{"owner": "test"}}},
			&drav1.DeviceAttributes{ObjectMeta: metav1.ObjectMeta{
				Name: "unrelated", Namespace: "operator"}},
			&drav1.DeviceAttributes{ObjectMeta: metav1.ObjectMeta{
				Name: "other-namespace", Namespace: "other", Labels: map[string]string{"owner": "test"}}},
		}, SchemeAttachers: []clients.SchemeAttacher{drav1.AddToScheme},
	})
	builders, err := ListDeviceAttributes(
		apiClient, "operator", runtimeclient.InNamespace("other"), runtimeclient.MatchingLabels{"owner": "test"})
	require.NoError(t, err)
	require.Len(t, builders, 1)
	require.Equal(t, "selected", builders[0].Object.Name)
}
