package sriov

import (
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	defaultPoolConfigName   = "poolconfig"
	defaultPoolConfigNsName = "testnamespace"
)

func TestNewPoolConfigBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewPoolConfigBuilder,
		srIovV1.AddToScheme,
		poolConfigGVK,
	).ExecuteTests(t)
}

func TestPullPoolConfig(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPoolConfig,
		srIovV1.AddToScheme,
		poolConfigGVK,
	).ExecuteTests(t)
}

func TestPoolConfigBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[srIovV1.SriovNetworkPoolConfig, PoolConfigBuilder](
		srIovV1.AddToScheme,
		poolConfigGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		nodeSelector     map[string]string
		invalidBuilder   bool
		expectedError    string
		expectedSelector map[string]string
	}{
		{
			name:             "valid selector",
			nodeSelector:     map[string]string{"test": "test"},
			expectedSelector: map[string]string{"test": "test"},
		},
		{
			name:          "empty selector returns error",
			nodeSelector:  map[string]string{},
			expectedError: "SriovNetworkPoolConfig 'nodeSelector' cannot be empty map",
		},
		{
			name:           "invalid builder short-circuits",
			nodeSelector:   map[string]string{"test": "test"},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidPoolConfigTestBuilder(newPoolConfigTestClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidPoolConfigTestBuilder(newPoolConfigTestClient())
			}

			testBuilder.WithNodeSelector(testCase.nodeSelector)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.NodeSelector)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.NodeSelector)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.NotNil(t, testBuilder.Definition.Spec.NodeSelector)
			assert.Equal(t, testCase.expectedSelector, testBuilder.Definition.Spec.NodeSelector.MatchLabels)
		})
	}
}

func TestWithMaxUnavailable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		maxUnavailable intstr.IntOrString
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:           "integer value",
			maxUnavailable: intstr.FromInt32(3),
		},
		{
			name:           "percentage value",
			maxUnavailable: intstr.FromString("50%"),
		},
		{
			name:           "string without percentage returns error",
			maxUnavailable: intstr.FromString("wrongValue"),
			expectedError:  "invalid type: strings needs to be a percentage: {1 0 wrongValue}",
		},
		{
			name:           "invalid percentage returns parse error",
			maxUnavailable: intstr.FromString("99s%"),
			expectedError:  "invalid value \"99s%\": strconv.Atoi: parsing \"99s\": invalid syntax",
		},
		{
			name:           "percentage above range returns error",
			maxUnavailable: intstr.FromString("101%"),
			expectedError:  "invalid value: percentage needs to be between 1 and 100: {1 0 101%}",
		},
		{
			name:           "negative integer returns error",
			maxUnavailable: intstr.FromInt32(-2),
			expectedError:  "negative number is not allowed: {0 -2 }",
		},
		{
			name:           "invalid builder short-circuits",
			maxUnavailable: intstr.FromInt32(3),
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidPoolConfigTestBuilder(newPoolConfigTestClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidPoolConfigTestBuilder(newPoolConfigTestClient())
			}

			testBuilder.WithMaxUnavailable(testCase.maxUnavailable)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.MaxUnavailable)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.MaxUnavailable)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.NotNil(t, testBuilder.Definition.Spec.MaxUnavailable)
			assert.Equal(t, testCase.maxUnavailable, *testBuilder.Definition.Spec.MaxUnavailable)
		})
	}
}

func TestWithRDMAMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		rdmaMode       string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:     "shared mode",
			rdmaMode: "shared",
		},
		{
			name:     "exclusive mode",
			rdmaMode: "exclusive",
		},
		{
			name:          "empty mode returns error",
			rdmaMode:      "",
			expectedError: "rdmaMode cannot be empty",
		},
		{
			name:          "invalid mode returns error",
			rdmaMode:      "none",
			expectedError: "invalid value for rdmaMode. It should be 'shared' or 'exclusive'",
		},
		{
			name:           "invalid builder short-circuits",
			rdmaMode:       "shared",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidPoolConfigTestBuilder(newPoolConfigTestClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidPoolConfigTestBuilder(newPoolConfigTestClient())
			}

			testBuilder.WithRDMAMode(testCase.rdmaMode)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.RdmaMode)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Empty(t, testBuilder.Definition.Spec.RdmaMode)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.rdmaMode, testBuilder.Definition.Spec.RdmaMode)
		})
	}
}

// newPoolConfigTestClient returns a test client with the SRIOV scheme and any provided objects.
func newPoolConfigTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: testSchemes,
	})
}

// buildValidPoolConfigTestBuilder returns a valid PoolConfigBuilder for testing purposes.
func buildValidPoolConfigTestBuilder(apiClient *clients.Settings) *PoolConfigBuilder {
	return NewPoolConfigBuilder(apiClient, defaultPoolConfigName, defaultPoolConfigNsName)
}

// buildInvalidPoolConfigTestBuilder returns an invalid PoolConfigBuilder for testing purposes.
func buildInvalidPoolConfigTestBuilder(apiClient *clients.Settings) *PoolConfigBuilder {
	return NewPoolConfigBuilder(apiClient, defaultPoolConfigName, "")
}

// buildDummyPoolConfigObject returns a SriovNetworkPoolConfig suitable for use as a mock API object.
func buildDummyPoolConfigObject() []runtime.Object {
	return []runtime.Object{&srIovV1.SriovNetworkPoolConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPoolConfigName,
			Namespace: defaultPoolConfigNsName,
		},
		Spec: srIovV1.SriovNetworkPoolConfigSpec{},
	}}
}
