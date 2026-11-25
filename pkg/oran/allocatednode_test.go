package oran

import (
	"testing"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

// pluginsTestSchemes is shared with the NodeAllocationRequest tests in this package.
var pluginsTestSchemes = []clients.SchemeAttacher{
	hardwaremanagementv1alpha1.AddToScheme,
}

var allocatedNodeGVK = hardwaremanagementv1alpha1.GroupVersion.WithKind("AllocatedNode")

func TestPullAllocatedNode(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullAllocatedNode,
		hardwaremanagementv1alpha1.AddToScheme,
		allocatedNodeGVK,
	).ExecuteTests(t)
}

func TestAllocatedNodeMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[hardwaremanagementv1alpha1.AllocatedNode, AllocatedNodeBuilder](
		hardwaremanagementv1alpha1.AddToScheme,
		allocatedNodeGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}
