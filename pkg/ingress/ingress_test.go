package ingress

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	networkingv1 "k8s.io/api/networking/v1"
)

func TestNewIngressBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			NewIngressBuilder,
			networkingv1.AddToScheme,
			ingressGVK,
		).ExecuteTests(t)
	})

	t.Run("sets default ingress class name", func(t *testing.T) {
		t.Parallel()

		builder := NewIngressBuilder(
			clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: []clients.SchemeAttacher{networkingv1.AddToScheme},
			}),
			"test-ingress",
			"test-namespace",
		)

		require.NoError(t, builder.GetError())
		require.NotNil(t, builder.Definition.Spec.IngressClassName)
		assert.Equal(t, DefaultIngressClassName, *builder.Definition.Spec.IngressClassName)
	})
}

func TestPullIngress(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullIngress,
		networkingv1.AddToScheme,
		ingressGVK,
	).ExecuteTests(t)
}

func TestIngressBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[networkingv1.Ingress, IngressBuilder](
		networkingv1.AddToScheme,
		ingressGVK,
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
