package ocm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
)

func TestListPlacementBindingsInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListPlacementBindingsInAllNamespaces,
		policiesv1.AddToScheme,
		placementBindingGVK,
	).ExecuteTests(t)
}
