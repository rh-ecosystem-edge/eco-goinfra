package oran

import (
	"testing"

	inventoryv1alpha1 "github.com/openshift-kni/oran-o2ims/api/inventory/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var inventoryGVK = inventoryv1alpha1.GroupVersion.WithKind("Inventory")

func TestPullInventory(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullInventory,
		inventoryv1alpha1.AddToScheme,
		inventoryGVK,
	).ExecuteTests(t)
}

func TestInventoryMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[inventoryv1alpha1.Inventory, InventoryBuilder](
		inventoryv1alpha1.AddToScheme,
		inventoryGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}
