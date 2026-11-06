package serviceaccount

import (
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[corev1.ServiceAccount, Builder](
		NewBuilder, corev1.AddToScheme, serviceAccountGVK).ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[corev1.ServiceAccount, Builder](
		Pull, corev1.AddToScheme, serviceAccountGVK).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.ServiceAccount, Builder](
		corev1.AddToScheme, serviceAccountGVK, testhelper.ResourceScopeNamespaced)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		Run(t)
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.ServiceAccount, Builder](
		corev1.AddToScheme, serviceAccountGVK, testhelper.ResourceScopeNamespaced)

	testhelper.NewWithOptionsTestConfig(commonConfig).ExecuteTests(t)
}

func TestCreateToken(t *testing.T) {
	t.Parallel()

	existingServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "test-sa", Namespace: "test-ns"},
	}

	t.Run("existing serviceaccount", func(t *testing.T) {
		t.Parallel()

		testClient := newServiceAccountTestClient([]runtime.Object{existingServiceAccount})
		fakeClient, ok := testClient.K8sClient.(*k8sfake.Clientset)
		require.True(t, ok)
		fakeClient.PrependReactor("create", "serviceaccounts", func(_ k8stesting.Action) (bool, runtime.Object, error) {
			return true, &authenticationv1.TokenRequest{
				Status: authenticationv1.TokenRequestStatus{Token: "test-token"},
			}, nil
		})

		builder := NewBuilder(testClient, "test-sa", "test-ns")
		token, err := builder.CreateToken(time.Hour, "https://example.com")

		require.NoError(t, err)
		assert.Equal(t, "test-token", token)
	})

	t.Run("missing serviceaccount", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder(newServiceAccountTestClient(nil), "test-sa", "test-ns")
		_, err := builder.CreateToken(time.Hour)

		require.EqualError(t, err, "serviceaccount test-sa does not exist in namespace test-ns")
	})
}

func TestGetGVR(t *testing.T) {
	t.Parallel()

	testGVR := GetGVR()
	assert.Equal(t, "", testGVR.Group)
	assert.Equal(t, "v1", testGVR.Version)
	assert.Equal(t, "serviceaccounts", testGVR.Resource)
}

func newServiceAccountTestClient(objects []runtime.Object) *clients.Settings {
	testClient := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
	testClient.K8sClient = k8sfake.NewSimpleClientset(objects...)
	testClient.CoreV1Interface = testClient.K8sClient.CoreV1()

	return testClient
}
