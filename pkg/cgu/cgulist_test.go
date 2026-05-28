package cgu

import (
	"testing"

	"github.com/openshift-kni/cluster-group-upgrades-operator/pkg/api/clustergroupupgrades/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

func TestListInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig[v1alpha1.ClusterGroupUpgrade, CguBuilder](
		ListInAllNamespaces,
		v1alpha1.AddToScheme,
		cguGVK,
	).ExecuteTests(t)
}
