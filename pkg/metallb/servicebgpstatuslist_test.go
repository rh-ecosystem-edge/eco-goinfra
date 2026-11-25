package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestServiceBGPStatusList(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListServiceBGPStatus,
		mlbtypes.AddToScheme,
		serviceBGPStatusGVK,
	).ExecuteTests(t)
}

func TestServiceBGPStatusListOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies a single list option", func(t *testing.T) {
		t.Parallel()

		builders, err := ListServiceBGPStatus(
			buildTestServiceBGPStatusClientWithDummyStatus(defaultServiceBGPStatusName),
			client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{"missing": "true"})},
		)

		require.NoError(t, err)
		assert.Empty(t, builders)
	})

	t.Run("rejects multiple list options", func(t *testing.T) {
		t.Parallel()

		builders, err := ListServiceBGPStatus(
			buildTestServiceBGPStatusClientWithDummyStatus(defaultServiceBGPStatusName),
			client.ListOptions{},
			client.ListOptions{},
		)

		assert.EqualError(t, err, "error: more than one ListOptions was passed")
		assert.Nil(t, builders)
	})
}
