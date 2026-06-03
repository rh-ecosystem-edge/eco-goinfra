package oran

import (
	"testing"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

var hardwareProfileGVK = hardwaremanagementv1alpha1.GroupVersion.WithKind("HardwareProfile")

func TestPullHardwareProfile(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullHardwareProfile,
		hardwaremanagementv1alpha1.AddToScheme,
		hardwareProfileGVK,
	).ExecuteTests(t)
}

func TestListHardwareProfiles(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListHardwareProfiles,
		hardwaremanagementv1alpha1.AddToScheme,
		hardwareProfileGVK,
	).ExecuteTests(t)
}

func TestHardwareProfileMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[hardwaremanagementv1alpha1.HardwareProfile, HardwareProfileBuilder](
		hardwaremanagementv1alpha1.AddToScheme,
		hardwareProfileGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}
