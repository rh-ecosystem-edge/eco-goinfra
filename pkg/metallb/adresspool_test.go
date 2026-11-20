package metallb

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	testSchemes = []clients.SchemeAttacher{
		mlbtypes.AddToScheme,
	}
	defaultIPAddressPoolName = "default-pool"
	defaultNsName            = "test-namespace"
	defaultIPPoolRange       = []string{"1.1.1.1", "1.1.1.20"}
	ipAddressPoolGVK         = mlbtypes.GroupVersion.WithKind("IPAddressPool")
	errEmptyIPAddressPool    = fmt.Errorf("IPAddressPool 'addrPool' cannot be empty list")
)

func TestNewIPAddressPoolBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *IPAddressPoolBuilder {
				return NewIPAddressPoolBuilder(apiClient, name, nsname, defaultIPPoolRange)
			},
			mlbtypes.AddToScheme,
			ipAddressPoolGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		addresses     []string
		expectedError error
	}{
		{
			name:      "valid address pool",
			addresses: defaultIPPoolRange,
		},
		{
			name:          "empty address pool",
			addresses:     []string{},
			expectedError: errEmptyIPAddressPool,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewIPAddressPoolBuilder(
				newIPAddressPoolTestClient(), defaultIPAddressPoolName, defaultNsName, testCase.addresses)

			require.NotNil(t, builder, "NewIPAddressPoolBuilder returned a nil builder")
			assert.Equal(t, testCase.expectedError, builder.GetError(), "unexpected builder error")

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.addresses, builder.Definition.Spec.Addresses,
					"builder definition should contain the requested addresses")
			}
		})
	}
}

func TestPullAddressPool(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullAddressPool,
		mlbtypes.AddToScheme,
		ipAddressPoolGVK,
	).ExecuteTests(t)
}

func TestIPAddressPoolMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newIPAddressPoolCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestIPAddressPoolWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newIPAddressPoolCommonTestConfig()).ExecuteTests(t)
}

func TestIPAddressPoolWithAutoAssign(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		autoAssign    bool
		expectedError error
	}{
		{
			name:       "sets auto assign true on valid builder",
			valid:      true,
			autoAssign: true,
		},
		{
			name:       "sets auto assign false on valid builder",
			valid:      true,
			autoAssign: false,
		},
		{
			name:          "invalid builder short circuits",
			valid:         false,
			autoAssign:    true,
			expectedError: errEmptyIPAddressPool,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := newIPAddressPoolBuilderForModifierTest(testCase.valid)

			result := builder.WithAutoAssign(testCase.autoAssign)

			require.Same(t, builder, result, "WithAutoAssign should return the original builder")
			assert.Equal(t, testCase.expectedError, result.GetError(), "unexpected builder error")

			if testCase.expectedError == nil {
				require.NotNil(t, result.Definition.Spec.AutoAssign, "AutoAssign should be set")
				assert.Equal(t, testCase.autoAssign, *result.Definition.Spec.AutoAssign,
					"AutoAssign should match the requested value")
			} else {
				assert.Nil(t, result.Definition.Spec.AutoAssign, "AutoAssign should not change on an invalid builder")
			}
		})
	}
}

func TestIPAddressPoolWithAvoidBuggyIPs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		avoidBuggyIPs bool
		expectedError error
	}{
		{
			name:          "sets avoid buggy IPs true on valid builder",
			valid:         true,
			avoidBuggyIPs: true,
		},
		{
			name:          "sets avoid buggy IPs false on valid builder",
			valid:         true,
			avoidBuggyIPs: false,
		},
		{
			name:          "invalid builder short circuits",
			valid:         false,
			avoidBuggyIPs: true,
			expectedError: errEmptyIPAddressPool,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := newIPAddressPoolBuilderForModifierTest(testCase.valid)

			result := builder.WithAvoidBuggyIPs(testCase.avoidBuggyIPs)

			require.Same(t, builder, result, "WithAvoidBuggyIPs should return the original builder")
			assert.Equal(t, testCase.expectedError, result.GetError(), "unexpected builder error")

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.avoidBuggyIPs, result.Definition.Spec.AvoidBuggyIPs,
					"AvoidBuggyIPs should match the requested value")
			} else {
				assert.False(t, result.Definition.Spec.AvoidBuggyIPs,
					"AvoidBuggyIPs should not change on an invalid builder")
			}
		})
	}
}

func TestGetIPAddressPoolGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, GetIPAddressPoolGVR(), schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: resourceIPAddressPools,
	}, "GVR should match the documented IPAddressPool resource")
}

// newIPAddressPoolCommonTestConfig returns the shared test configuration for IPAddressPool builder methods.
func newIPAddressPoolCommonTestConfig() testhelper.CommonTestConfig[
	mlbtypes.IPAddressPool, IPAddressPoolBuilder, *mlbtypes.IPAddressPool, *IPAddressPoolBuilder] {
	return testhelper.NewCommonTestConfig[mlbtypes.IPAddressPool, IPAddressPoolBuilder](
		mlbtypes.AddToScheme,
		ipAddressPoolGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

// newIPAddressPoolTestClient returns a fake client configured with the MetalLB scheme.
func newIPAddressPoolTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

// newIPAddressPoolBuilderForModifierTest returns a valid or builder-error-state IPAddressPool builder.
func newIPAddressPoolBuilderForModifierTest(valid bool) *IPAddressPoolBuilder {
	if valid {
		return buildValidIPAddressPoolBuilder(newIPAddressPoolTestClient())
	}

	return buildInValidIPAddressPoolBuilder(newIPAddressPoolTestClient())
}

// buildValidIPAddressPoolBuilder returns a valid IPAddressPool builder for testing.
func buildValidIPAddressPoolBuilder(apiClient *clients.Settings) *IPAddressPoolBuilder {
	return NewIPAddressPoolBuilder(
		apiClient, defaultIPAddressPoolName, defaultNsName, defaultIPPoolRange)
}

// buildInValidIPAddressPoolBuilder returns an IPAddressPool builder with an empty address list.
func buildInValidIPAddressPoolBuilder(apiClient *clients.Settings) *IPAddressPoolBuilder {
	return NewIPAddressPoolBuilder(
		apiClient, defaultIPAddressPoolName, defaultNsName, []string{})
}

// buildTestClientWithDummyObject returns a fake client containing the shared dummy IPAddressPool.
func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyIPAddressPool(),
		SchemeAttachers: testSchemes,
	})
}

// buildDummyIPAddressPool returns the shared dummy IPAddressPool objects used by MetalLB tests.
func buildDummyIPAddressPool() []runtime.Object {
	return append([]runtime.Object{}, &mlbtypes.IPAddressPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultIPAddressPoolName,
			Namespace: defaultNsName,
		},
		Spec: mlbtypes.IPAddressPoolSpec{
			Addresses: defaultIPPoolRange,
		},
	})
}
