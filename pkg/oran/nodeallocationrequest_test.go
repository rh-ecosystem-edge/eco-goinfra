package oran

import (
	"testing"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var nodeAllocationRequestGVK = hardwaremanagementv1alpha1.GroupVersion.WithKind("NodeAllocationRequest")

func TestPullNodeAllocationRequest(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullNodeAllocationRequest,
		hardwaremanagementv1alpha1.AddToScheme,
		nodeAllocationRequestGVK,
	).ExecuteTests(t)
}

func TestNodeAllocationRequestMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[hardwaremanagementv1alpha1.NodeAllocationRequest, NARBuilder](
		hardwaremanagementv1alpha1.AddToScheme,
		nodeAllocationRequestGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}
