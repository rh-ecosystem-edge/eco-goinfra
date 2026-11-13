package webhook

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	admregv1 "k8s.io/api/admissionregistration/v1"
)

func TestPullMutatingConfiguration(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullMutatingConfiguration, admregv1.AddToScheme, mutatingConfigurationGVK).ExecuteTests(t)
}

func TestMutatingConfigurationMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[admregv1.MutatingWebhookConfiguration, MutatingConfigurationBuilder](
		admregv1.AddToScheme,
		mutatingConfigurationGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}
