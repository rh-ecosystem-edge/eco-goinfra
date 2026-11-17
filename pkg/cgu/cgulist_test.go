package cgu

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/cgu/clustergroupupgrades/v1alpha1"
)

func TestListInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListInAllNamespaces,
		v1alpha1.AddToScheme,
		cguGVK,
	).ExecuteTests(t)
}
