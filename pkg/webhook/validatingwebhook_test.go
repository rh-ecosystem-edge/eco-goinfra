package webhook

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	admregv1 "k8s.io/api/admissionregistration/v1"
)

func TestPullValidatingConfiguration(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullValidatingConfiguration, admregv1.AddToScheme, validatingConfigurationGVK).ExecuteTests(t)
}

func TestValidatingConfigurationMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[admregv1.ValidatingWebhookConfiguration, ValidatingConfigurationBuilder](
		admregv1.AddToScheme,
		validatingConfigurationGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}
