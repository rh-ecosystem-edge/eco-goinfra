package configmap

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	corev1 "k8s.io/api/core/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*Builder, error) {
			return List(apiClient, nsname)
		},
		corev1.AddToScheme,
		configMapGVK,
	).ExecuteTests(t)
}

func TestListInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListInAllNamespaces,
		corev1.AddToScheme,
		configMapGVK,
	).ExecuteTests(t)
}
