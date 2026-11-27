package resourcequotas

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[corev1.ResourceQuota, Builder](
		NewBuilder, corev1.AddToScheme, resourceQuotaGVK).ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[corev1.ResourceQuota, Builder](
		Pull, corev1.AddToScheme, resourceQuotaGVK).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[corev1.ResourceQuota, Builder](
		corev1.AddToScheme, resourceQuotaGVK, testhelper.ResourceScopeNamespaced)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestWithQuotaSpec(t *testing.T) {
	t.Parallel()

	quotaSpec := corev1.ResourceQuotaSpec{
		Hard: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
	}

	testCases := []struct {
		name          string
		builder       func() *Builder
		assertError   func(error) bool
		expectApplied bool
		expectNil     bool
	}{
		{
			name: "valid builder",
			builder: func() *Builder {
				return NewBuilder(newResourceQuotaTestClient(), "test-rq", "test-namespace")
			},
			assertError:   func(err error) bool { return err == nil },
			expectApplied: true,
		},
		{
			name: "invalid builder",
			builder: func() *Builder {
				return NewBuilder(newResourceQuotaTestClient(), "", "test-namespace")
			},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
		{
			name:        "nil builder",
			builder:     func() *Builder { return nil },
			assertError: commonerrors.IsBuilderNil,
			expectNil:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithQuotaSpec(quotaSpec)

			if testCase.expectNil {
				assert.Nil(t, result)

				return
			}

			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if testCase.expectApplied {
				assert.Equal(t, quotaSpec, result.Definition.Spec)
			} else {
				assert.Empty(t, result.Definition.Spec.Hard)
			}
		})
	}
}

func TestGetGVR(t *testing.T) {
	t.Parallel()

	testGVR := GetGVR()
	assert.Equal(t, "", testGVR.Group)
	assert.Equal(t, "v1", testGVR.Version)
	assert.Equal(t, "resourcequotas", testGVR.Resource)
}

func newResourceQuotaTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
}
