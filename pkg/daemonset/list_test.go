package daemonset

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestList(t *testing.T) {
	apiClient := clients.GetTestClients(clients.TestClientParams{})
	fakeClient := k8sfake.NewClientset(&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{
		Name: "selected", Namespace: "operator", Labels: map[string]string{"app": "dra"}}},
		&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{
			Name: "other", Namespace: "operator"}})
	apiClient.AppsV1Interface = fakeClient.AppsV1()
	result, err := List(apiClient, "operator", metav1.ListOptions{LabelSelector: "app=dra"})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "selected", result[0].Object.Name)
	fakeClient.PrependReactor("list", "daemonsets", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Group: "apps", Resource: "daemonsets"}, "", nil)
	})

	_, err = List(apiClient, "operator")
	require.True(t, apierrors.IsForbidden(err))
	_, err = List(nil, "operator")
	require.Error(t, err)
	_, err = List(apiClient, "")
	require.Error(t, err)
	_, err = List(apiClient, "operator", metav1.ListOptions{}, metav1.ListOptions{})
	require.Error(t, err)
}
