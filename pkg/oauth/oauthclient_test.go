package oauth

import (
	"testing"

	oauthv1 "github.com/openshift/api/oauth/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

func TestPullOAuthClient(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig[oauthv1.OAuthClient, OAuthClientBuilder](
		PullOAuthClient, oauthv1.AddToScheme, oauthClientGVK).ExecuteTests(t)
}

func TestOAuthClientBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[oauthv1.OAuthClient, OAuthClientBuilder](
		oauthv1.AddToScheme, oauthClientGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}
