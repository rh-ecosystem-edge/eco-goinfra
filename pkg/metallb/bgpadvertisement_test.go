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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	defaultBGPAdvertisementName   = "default-bgpadvert"
	defaultBGPAdvertisementNsName = "test-namespace"
)

func TestNewBGPAdvertisementBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewBGPAdvertisementBuilder,
		mlbtypes.AddToScheme,
		bgpAdvertisementGVK,
	).ExecuteTests(t)
}

func TestPullBGPAdvertisement(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullBGPAdvertisement,
		mlbtypes.AddToScheme,
		bgpAdvertisementGVK,
	).ExecuteTests(t)
}

func TestBGPAdvertisementBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[mlbtypes.BGPAdvertisement, BGPAdvertisementBuilder](
		mlbtypes.AddToScheme,
		bgpAdvertisementGVK,
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

func TestBGPAdvertisementWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[mlbtypes.BGPAdvertisement, BGPAdvertisementBuilder](
			mlbtypes.AddToScheme,
			bgpAdvertisementGVK,
			testhelper.ResourceScopeNamespaced,
		)).ExecuteTests(t)
}

func TestBGPAdvertisementWithAggregationLength4(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		aggregationLength int32
		validBuilder      bool
		expectedError     string
	}{
		{
			name:              "sets aggregation length",
			aggregationLength: 32,
			validBuilder:      true,
		},
		{
			name:              "accepts lower boundary",
			aggregationLength: 0,
			validBuilder:      true,
		},
		{
			name:              "rejects aggregation length above upper boundary",
			aggregationLength: 33,
			validBuilder:      true,
			expectedError:     "AggregationLength 33 is invalid, the value shoud be in range 0...32",
		},
		{
			name:              "rejects aggregation length below lower boundary",
			aggregationLength: -1,
			validBuilder:      true,
			expectedError:     "AggregationLength -1 is invalid, the value shoud be in range 0...32",
		},
		{
			name:              "invalid builder short-circuits",
			aggregationLength: 32,
			validBuilder:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBGPAdvertisementTestBuilder(testCase.validBuilder)
			testBuilder.WithAggregationLength4(testCase.aggregationLength)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.AggregationLength)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.AggregationLength)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.NotNil(t, testBuilder.Definition.Spec.AggregationLength)
			assert.Equal(t, testCase.aggregationLength, *testBuilder.Definition.Spec.AggregationLength)
		})
	}
}

func TestBGPAdvertisementWithAggregationLength6(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		aggregationLength int32
		validBuilder      bool
		expectedError     string
	}{
		{
			name:              "sets aggregation length v6",
			aggregationLength: 128,
			validBuilder:      true,
		},
		{
			name:              "accepts lower boundary",
			aggregationLength: 0,
			validBuilder:      true,
		},
		{
			name:              "rejects aggregation length above upper boundary",
			aggregationLength: 129,
			validBuilder:      true,
			expectedError:     "AggregationLength 129 is invalid, the value shoud be in range 0...128",
		},
		{
			name:              "rejects aggregation length below lower boundary",
			aggregationLength: -1,
			validBuilder:      true,
			expectedError:     "AggregationLength -1 is invalid, the value shoud be in range 0...128",
		},
		{
			name:              "invalid builder short-circuits",
			aggregationLength: 128,
			validBuilder:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBGPAdvertisementTestBuilder(testCase.validBuilder)
			testBuilder.WithAggregationLength6(testCase.aggregationLength)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.AggregationLengthV6)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.AggregationLengthV6)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.NotNil(t, testBuilder.Definition.Spec.AggregationLengthV6)
			assert.Equal(t, testCase.aggregationLength, *testBuilder.Definition.Spec.AggregationLengthV6)
		})
	}
}

func TestBGPAdvertisementWithLocalPref(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		localPref    uint32
		validBuilder bool
	}{
		{
			name:         "sets local preference",
			localPref:    100,
			validBuilder: true,
		},
		{
			name:         "accepts zero local preference",
			localPref:    0,
			validBuilder: true,
		},
		{
			name:         "invalid builder short-circuits",
			localPref:    100,
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBGPAdvertisementTestBuilder(testCase.validBuilder)
			testBuilder.WithLocalPref(testCase.localPref)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Zero(t, testBuilder.Definition.Spec.LocalPref)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.localPref, testBuilder.Definition.Spec.LocalPref)
		})
	}
}

func TestBGPAdvertisementWithCommunities(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		communities   []string
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "sets communities",
			communities:  []string{"5252", "65000:1"},
			validBuilder: true,
		},
		{
			name:          "rejects empty communities",
			communities:   []string{},
			validBuilder:  true,
			expectedError: "error: community setting is empty list, the list should contain at least one element",
		},
		{
			name:         "invalid builder short-circuits",
			communities:  []string{"5252"},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBGPAdvertisementTestBuilder(testCase.validBuilder)
			testBuilder.WithCommunities(testCase.communities)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.Communities)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.Communities)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.communities, testBuilder.Definition.Spec.Communities)
		})
	}
}

func TestBGPAdvertisementWithIPAddressPools(t *testing.T) {
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

			testBuilder := buildBGPAdvertisementTestBuilder(testCase.validBuilder)
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

func TestBGPAdvertisementWithIPAddressPoolsSelectors(t *testing.T) {
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

			testBuilder := buildBGPAdvertisementTestBuilder(testCase.validBuilder)
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

func TestBGPAdvertisementWithNodeSelector(t *testing.T) {
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

			testBuilder := buildBGPAdvertisementTestBuilder(testCase.validBuilder)
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

func TestBGPAdvertisementWithPeers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		peers         []string
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "sets peers",
			peers:        []string{"peer-1", "peer-2"},
			validBuilder: true,
		},
		{
			name:          "rejects empty peers",
			peers:         []string{},
			validBuilder:  true,
			expectedError: "error: peers setting is empty list, the list should contain at least one element",
		},
		{
			name:         "invalid builder short-circuits",
			peers:        []string{"peer-1"},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildBGPAdvertisementTestBuilder(testCase.validBuilder)
			testBuilder.WithPeers(testCase.peers)

			if !testCase.validBuilder {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.Peers)

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.Peers)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.peers, testBuilder.Definition.Spec.Peers)
		})
	}
}

func TestBGPAdvertisementGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "bgpadvertisements",
	}, GetBGPAdvertisementGVR())
}

func buildBGPAdvertisementTestBuilder(valid bool) *BGPAdvertisementBuilder {
	namespace := defaultBGPAdvertisementNsName
	if !valid {
		namespace = ""
	}

	return NewBGPAdvertisementBuilder(getTestBGPAdvertisementAPIClient(), defaultBGPAdvertisementName, namespace)
}

func getTestBGPAdvertisementAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{mlbtypes.AddToScheme},
	})
}
