package certificate

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	certificatesv1 "k8s.io/api/certificates/v1"
)

var signingRequestGVK = certificatesv1.SchemeGroupVersion.WithKind("CertificateSigningRequest")

func TestNewSigningRequestBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(
		NewSigningRequestBuilder, certificatesv1.AddToScheme, signingRequestGVK).
		ExecuteTests(t)
}

func TestPullSigningRequest(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullSigningRequest, certificatesv1.AddToScheme, signingRequestGVK).
		ExecuteTests(t)
}

func TestSigningRequestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newSigningRequestCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		Run(t)
}

func newSigningRequestCommonTestConfig() testhelper.CommonTestConfig[
	certificatesv1.CertificateSigningRequest, SigningRequestBuilder,
	*certificatesv1.CertificateSigningRequest, *SigningRequestBuilder,
] {
	return testhelper.NewCommonTestConfig[certificatesv1.CertificateSigningRequest, SigningRequestBuilder](
		certificatesv1.AddToScheme, signingRequestGVK, testhelper.ResourceScopeClusterScoped)
}
