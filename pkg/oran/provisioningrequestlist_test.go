package oran

import (
	"testing"

	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListProvisioningRequests(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListProvisioningRequests,
		provisioningv1alpha1.AddToScheme,
		provisioningRequestGVK,
	).ExecuteTests(t)

	t.Run("multiple list options return error", func(t *testing.T) {
		t.Parallel()

		_, err := ListProvisioningRequests(nil, runtimeclient.ListOptions{}, runtimeclient.ListOptions{})
		assert.EqualError(t, err, "error: more than one ListOptions was passed")
	})
}
