package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	defaultL2AdvertisementName   = "default-l2advertisement"
	defaultL2AdvertisementNsName = "test-namespace"
)

func TestNewL2AdvertisementBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewL2AdvertisementBuilder,
		mlbtypes.AddToScheme,
		l2AdvertisementGVK,
	).ExecuteTests(t)
}

func TestPullL2Advertisement(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullL2Advertisement,
		mlbtypes.AddToScheme,
		l2AdvertisementGVK,
	).ExecuteTests(t)
}

func TestL2AdvertisementBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[mlbtypes.L2Advertisement, L2AdvertisementBuilder](
		mlbtypes.AddToScheme,
		l2AdvertisementGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestL2AdvertisementWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[mlbtypes.L2Advertisement, L2AdvertisementBuilder](
			mlbtypes.AddToScheme,
			l2AdvertisementGVK,
			testhelper.ResourceScopeNamespaced,
		)).ExecuteTests(t)
}

func TestL2AdvertisementWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		nodeSelectors []metav1.LabelSelector
		validBuilder  bool
		expectedError string
	}{
		{
			name:          "sets node selectors",
			nodeSelectors: []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
			validBuilder:  true,
		},
		{
			name:          "rejects empty node selectors",
			nodeSelectors: []metav1.LabelSelector{},
			validBuilder:  true,
			expectedError: errEmptyNodeSelectors,
		},
		{
			name:          "invalid builder short-circuits",
			nodeSelectors: []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
			validBuilder:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildL2AdvertisementTestBuilder(testCase.validBuilder)
			testBuilder.WithNodeSelector(testCase.nodeSelectors)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.NodeSelectors)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.NodeSelectors)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.nodeSelectors, testBuilder.Definition.Spec.NodeSelectors)
		})
	}
}

func TestL2AdvertisementWithIPAddressPools(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		ipAddressPools []string
		validBuilder   bool
		expectedError  string
	}{
		{
			name:           "sets IP address pools",
			ipAddressPools: []string{"address-pool"},
			validBuilder:   true,
		},
		{
			name:           "rejects empty IP address pools",
			ipAddressPools: []string{},
			validBuilder:   true,
			expectedError:  errEmptyIPAddressPools,
		},
		{
			name:           "invalid builder short-circuits",
			ipAddressPools: []string{"address-pool"},
			validBuilder:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildL2AdvertisementTestBuilder(testCase.validBuilder)
			testBuilder.WithIPAddressPools(testCase.ipAddressPools)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.IPAddressPools)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.IPAddressPools)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.ipAddressPools, testBuilder.Definition.Spec.IPAddressPools)
		})
	}
}

func TestL2AdvertisementWithIPAddressPoolsSelectors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		poolSelector  []metav1.LabelSelector
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "sets IP address pool selectors",
			poolSelector: []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
			validBuilder: true,
		},
		{
			name:         "rejects empty IP address pool selectors",
			poolSelector: []metav1.LabelSelector{},
			validBuilder: true,
			expectedError: "error: IPAddressPoolSelectors setting is empty list, the list should contain at " +
				"least one element",
		},
		{
			name:         "invalid builder short-circuits",
			poolSelector: []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildL2AdvertisementTestBuilder(testCase.validBuilder)
			testBuilder.WithIPAddressPoolsSelectors(testCase.poolSelector)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.IPAddressPoolSelectors)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.IPAddressPoolSelectors)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.poolSelector, testBuilder.Definition.Spec.IPAddressPoolSelectors)
		})
	}
}

func TestL2AdvertisementWithInterfaces(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		interfaces    []string
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "sets interfaces",
			interfaces:   []string{"eno1"},
			validBuilder: true,
		},
		{
			name:          "rejects empty interfaces",
			interfaces:    []string{},
			validBuilder:  true,
			expectedError: "error: Interfaces setting is empty list, the list should contain at least one element",
		},
		{
			name:         "invalid builder short-circuits",
			interfaces:   []string{"eno1"},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildL2AdvertisementTestBuilder(testCase.validBuilder)
			testBuilder.WithInterfaces(testCase.interfaces)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.Interfaces)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.Interfaces)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.interfaces, testBuilder.Definition.Spec.Interfaces)
		})
	}
}

func TestGetL2AdvertisementGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "l2advertisements",
	}, GetL2AdvertisementGVR())
}

// buildL2AdvertisementTestBuilder returns a valid or builder-error-state L2Advertisement builder.
func buildL2AdvertisementTestBuilder(valid bool) *L2AdvertisementBuilder {
	namespace := defaultL2AdvertisementNsName
	if !valid {
		namespace = ""
	}

	return NewL2AdvertisementBuilder(getTestL2AdvertisementAPIClient(), defaultL2AdvertisementName, namespace)
}

// getTestL2AdvertisementAPIClient returns a fake client configured with the MetalLB scheme.
func getTestL2AdvertisementAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyL2Advertisement(),
		SchemeAttachers: testSchemes,
	})
}

// buildDummyL2Advertisement returns the shared dummy L2Advertisement used by MetalLB tests.
func buildDummyL2Advertisement() []runtime.Object {
	return append([]runtime.Object{}, &mlbtypes.L2Advertisement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultL2AdvertisementName,
			Namespace: defaultL2AdvertisementNsName,
		},
		Spec: mlbtypes.L2AdvertisementSpec{},
	})
}
