package ocm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	policiesv1beta1 "open-cluster-management.io/governance-policy-propagator/api/v1beta1"
)

func TestListPolicieSetsInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListPolicieSetsInAllNamespaces,
		policiesv1beta1.AddToScheme,
		policySetGVK,
	).ExecuteTests(t)
}
