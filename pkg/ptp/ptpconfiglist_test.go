package ptp

import (
	"testing"

	ptpv1 "github.com/k8snetworkplumbingwg/ptp-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPtpConfigsOptions(t *testing.T) {
	t.Parallel()

	t.Run("rejects more than one list option", func(t *testing.T) {
		t.Parallel()

		builders, err := ListPtpConfigs(
			buildTestClientWithDummyPtpConfig(),
			runtimeclient.ListOptions{LabelSelector: labels.NewSelector()},
			runtimeclient.ListOptions{LabelSelector: labels.NewSelector()},
		)
		require.EqualError(t, err, "error: more than one ListOptions was passed")
		require.Nil(t, builders)
	})
}

func TestListPtpConfigs(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListPtpConfigs,
		ptpv1.AddToScheme,
		ptpConfigGVK,
	).ExecuteTests(t)
}
