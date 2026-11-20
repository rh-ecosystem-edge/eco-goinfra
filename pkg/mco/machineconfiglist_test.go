package mco

import (
	"testing"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListMC(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListMC,
		mcv1.Install,
		machineConfigGVK,
	).ExecuteTests(t)
}

func TestListMCOptions(t *testing.T) {
	t.Parallel()

	t.Run("single option is accepted", func(t *testing.T) {
		t.Parallel()

		builders, err := ListMC(newMachineConfigTestClient(), runtimeclient.ListOptions{Continue: "test"})
		require.NoError(t, err)
		assert.Empty(t, builders)
	})

	t.Run("more than one option returns error", func(t *testing.T) {
		t.Parallel()

		builders, err := ListMC(
			newMachineConfigTestClient(),
			runtimeclient.ListOptions{Continue: "test"},
			runtimeclient.ListOptions{Continue: "test"},
		)
		require.EqualError(t, err, "error: more than one ListOptions was passed")
		assert.Empty(t, builders)
	})
}
