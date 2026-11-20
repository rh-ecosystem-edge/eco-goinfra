package olm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	operatorsv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var operatorGroupGVK = operatorsv1.SchemeGroupVersion.WithKind("OperatorGroup")

func TestNewOperatorGroupBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewOperatorGroupBuilder,
		operatorsv1.AddToScheme,
		operatorGroupGVK,
	).ExecuteTests(t)

	t.Run("initializes OperatorGroup fields", func(t *testing.T) {
		t.Parallel()

		apiClient := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: []clients.SchemeAttacher{operatorsv1.AddToScheme},
		})
		builder := NewOperatorGroupBuilder(apiClient, "operatorgroup", "test-namespace")

		require.NoError(t, builder.GetError())
		assert.Equal(t, "operatorgroup-", builder.Definition.GenerateName)
		assert.Equal(t, []string{"test-namespace"}, builder.Definition.Spec.TargetNamespaces)
	})
}

func TestPullOperatorGroup(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullOperatorGroup,
		operatorsv1.AddToScheme,
		operatorGroupGVK,
	).ExecuteTests(t)
}

func TestOperatorGroupBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[operatorsv1.OperatorGroup, OperatorGroupBuilder](
		operatorsv1.AddToScheme,
		operatorGroupGVK,
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
