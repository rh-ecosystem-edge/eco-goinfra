package metallb

import (
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	defaultFrrConfigurationName = "frrconfiguration"
	defaultFrrBGPPeerIP         = "10.46.73.131"
	invalidRouterIndexError     = "invalid routerIndex: 1"
	invalidNeighborIndexError   = "invalid neighborIndex: 1"
)

var (
	frrTestSchemes                = []clients.SchemeAttacher{frrtypes.AddToScheme}
	defaultFrrConfigurationNsName = "test-namespace"
)

func TestNewFrrConfigurationBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewFrrConfigurationBuilder,
		frrtypes.AddToScheme,
		frrConfigurationGVK,
	).ExecuteTests(t)
}

func TestFrrConfigurationMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[frrtypes.FRRConfiguration, FrrConfigurationBuilder](
		frrtypes.AddToScheme,
		frrConfigurationGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestGetFrrConfigurationGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, schema.GroupVersionResource{
		Group: FRRAPIGroup, Version: APIVersion, Resource: "frrconfigurations",
	}, GetFrrConfigurationGVR())
}

func TestFrrConfigurationWithBGPRouter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		localASN      uint32
		expectedError string
	}{
		{
			name:     "adds BGP router",
			valid:    true,
			localASN: 64500,
		},
		{
			name:     "invalid builder short circuits",
			valid:    false,
			localASN: 64500,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationBuilderForTest(testCase.valid)
			result := builder.WithBGPRouter(testCase.localASN)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			require.Len(t, result.Definition.Spec.BGP.Routers, 1)
			assert.Equal(t, testCase.localASN, result.Definition.Spec.BGP.Routers[0].ASN)
		})
	}
}

func TestFrrConfigurationBGPNeighbor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		bgpPeer       string
		remoteAS      uint32
		routerIndex   uint
		expectedError string
	}{
		{
			name:        "adds BGP neighbor",
			valid:       true,
			bgpPeer:     defaultFrrBGPPeerIP,
			remoteAS:    64500,
			routerIndex: 0,
		},
		{
			name:          "invalid peer IP",
			valid:         true,
			bgpPeer:       "10.46.71.",
			remoteAS:      64500,
			routerIndex:   0,
			expectedError: "frrConfiguration 'peerIP' of the BGPPeer contains invalid ip address",
		},
		{
			name:          "invalid router index",
			valid:         true,
			bgpPeer:       defaultFrrBGPPeerIP,
			remoteAS:      64500,
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:     "invalid builder short circuits",
			valid:    false,
			bgpPeer:  defaultFrrBGPPeerIP,
			remoteAS: 64500,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithRouterForTest(testCase.valid)
			result := builder.WithBGPNeighbor(testCase.bgpPeer, testCase.remoteAS, testCase.routerIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			require.Len(t, result.Definition.Spec.BGP.Routers[0].Neighbors, 1)
			assert.Equal(t, testCase.bgpPeer, result.Definition.Spec.BGP.Routers[0].Neighbors[0].Address)
			assert.Equal(t, testCase.remoteAS, result.Definition.Spec.BGP.Routers[0].Neighbors[0].ASN)
		})
	}
}

func TestFrrConfigurationWithBGPPassword(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		bgpPassword   string
		routerIndex   uint
		neighborIndex uint
		expectedError string
	}{
		{
			name:        "sets password",
			valid:       true,
			bgpPassword: "bgpPassword",
		},
		{
			name:          "empty password",
			valid:         true,
			bgpPassword:   "",
			expectedError: "the bgpPassword  is an empty string",
		},
		{
			name:          "invalid router index",
			valid:         true,
			bgpPassword:   "bgpPassword",
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:          "invalid neighbor index",
			valid:         true,
			bgpPassword:   "bgpPassword",
			neighborIndex: 1,
			expectedError: invalidNeighborIndexError,
		},
		{
			name:        "invalid builder short circuits",
			valid:       false,
			bgpPassword: "bgpPassword",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithBGPPassword(testCase.bgpPassword, testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			assert.Equal(t, testCase.bgpPassword, result.Definition.Spec.BGP.Routers[0].Neighbors[0].Password)
		})
	}
}

func TestFrrConfigurationWithHoldTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		holdTime      metav1.Duration
		routerIndex   uint
		neighborIndex uint
		expectedError string
	}{
		{
			name:     "sets hold time",
			valid:    true,
			holdTime: metav1.Duration{Duration: 90 * time.Second},
		},
		{
			name:          "hold time below minimum",
			valid:         true,
			holdTime:      metav1.Duration{},
			expectedError: "frrConfiguration 'holdtime' value is not valid",
		},
		{
			name:          "hold time above maximum",
			valid:         true,
			holdTime:      metav1.Duration{Duration: 65536 * time.Second},
			expectedError: "frrConfiguration 'holdtime' value is not valid",
		},
		{
			name:          "invalid router index",
			valid:         true,
			holdTime:      metav1.Duration{Duration: 90 * time.Second},
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:          "invalid neighbor index",
			valid:         true,
			holdTime:      metav1.Duration{Duration: 90 * time.Second},
			neighborIndex: 1,
			expectedError: invalidNeighborIndexError,
		},
		{
			name:     "invalid builder short circuits",
			valid:    false,
			holdTime: metav1.Duration{Duration: 90 * time.Second},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithHoldTime(testCase.holdTime, testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			assert.Equal(t, &testCase.holdTime, result.Definition.Spec.BGP.Routers[0].Neighbors[0].HoldTime)
		})
	}
}

func TestFrrConfigurationWithKeepalive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		keepalive     metav1.Duration
		routerIndex   uint
		neighborIndex uint
		expectedError string
	}{
		{
			name:      "sets keepalive",
			valid:     true,
			keepalive: metav1.Duration{Duration: 30 * time.Second},
		},
		{
			name:          "keepalive below minimum",
			valid:         true,
			keepalive:     metav1.Duration{},
			expectedError: "frrConfiguration 'keepAlive' value is not valid",
		},
		{
			name:          "keepalive above maximum",
			valid:         true,
			keepalive:     metav1.Duration{Duration: 65536 * time.Second},
			expectedError: "frrConfiguration 'keepAlive' value is not valid",
		},
		{
			name:          "invalid router index",
			valid:         true,
			keepalive:     metav1.Duration{Duration: 30 * time.Second},
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:          "invalid neighbor index",
			valid:         true,
			keepalive:     metav1.Duration{Duration: 30 * time.Second},
			neighborIndex: 1,
			expectedError: invalidNeighborIndexError,
		},
		{
			name:      "invalid builder short circuits",
			valid:     false,
			keepalive: metav1.Duration{Duration: 30 * time.Second},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithKeepalive(testCase.keepalive, testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			assert.Equal(t, &testCase.keepalive, result.Definition.Spec.BGP.Routers[0].Neighbors[0].KeepaliveTime)
		})
	}
}

func TestFrrConfigurationWithConnectTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		connectTime   metav1.Duration
		routerIndex   uint
		neighborIndex uint
		expectedError string
	}{
		{
			name:        "sets connect time",
			valid:       true,
			connectTime: metav1.Duration{Duration: 10 * time.Second},
		},
		{
			name:          "connect time below minimum",
			valid:         true,
			connectTime:   metav1.Duration{},
			expectedError: frrconfigurationConnecttimeValueIsNotValid,
		},
		{
			name:          "connect time above maximum",
			valid:         true,
			connectTime:   metav1.Duration{Duration: 65536 * time.Second},
			expectedError: frrconfigurationConnecttimeValueIsNotValid,
		},
		{
			name:          "invalid router index",
			valid:         true,
			connectTime:   metav1.Duration{Duration: 10 * time.Second},
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:          "invalid neighbor index",
			valid:         true,
			connectTime:   metav1.Duration{Duration: 10 * time.Second},
			neighborIndex: 1,
			expectedError: invalidNeighborIndexError,
		},
		{
			name:        "invalid builder short circuits",
			valid:       false,
			connectTime: metav1.Duration{Duration: 10 * time.Second},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithConnectTime(testCase.connectTime, testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			assert.Equal(t, &testCase.connectTime, result.Definition.Spec.BGP.Routers[0].Neighbors[0].ConnectTime)
		})
	}
}

func TestFrrConfigurationWithEBGPMultiHop(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		routerIndex   uint
		neighborIndex uint
		expectedError string
	}{
		{
			name:  "enables multihop",
			valid: true,
		},
		{
			name:          "invalid router index",
			valid:         true,
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:          "invalid neighbor index",
			valid:         true,
			neighborIndex: 1,
			expectedError: invalidNeighborIndexError,
		},
		{
			name:  "invalid builder short circuits",
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithEBGPMultiHop(testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			assert.True(t, result.Definition.Spec.BGP.Routers[0].Neighbors[0].EBGPMultiHop)
		})
	}
}

func TestFrrConfigurationWithBGPNeighborDisableMP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		disableMP     bool
		routerIndex   uint
		neighborIndex uint
		expectedError string
	}{
		{
			name:      "sets disable MP true",
			valid:     true,
			disableMP: true,
		},
		{
			name:      "sets disable MP false",
			valid:     true,
			disableMP: false,
		},
		{
			name:          "invalid router index",
			valid:         true,
			disableMP:     true,
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:          "invalid neighbor index",
			valid:         true,
			disableMP:     true,
			neighborIndex: 1,
			expectedError: invalidNeighborIndexError,
		},
		{
			name:      "invalid builder short circuits",
			valid:     false,
			disableMP: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithBGPNeighborDisableMP(
				testCase.disableMP, testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			//nolint:staticcheck // DisableMP is deprecated, keep for backwards compatibility.
			assert.Equal(t, testCase.disableMP, result.Definition.Spec.BGP.Routers[0].Neighbors[0].DisableMP)
		})
	}
}

func TestFrrConfigurationWithBGPNeighborDualStackAddressFamily(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		valid                  bool
		dualStackAddressFamily bool
		routerIndex            uint
		neighborIndex          uint
		expectedError          string
	}{
		{
			name:                   "enables dual stack address family",
			valid:                  true,
			dualStackAddressFamily: true,
		},
		{
			name:                   "disables dual stack address family",
			valid:                  true,
			dualStackAddressFamily: false,
		},
		{
			name:                   "invalid router index",
			valid:                  true,
			dualStackAddressFamily: true,
			routerIndex:            1,
			expectedError:          invalidRouterIndexError,
		},
		{
			name:                   "invalid neighbor index",
			valid:                  true,
			dualStackAddressFamily: true,
			neighborIndex:          1,
			expectedError:          invalidNeighborIndexError,
		},
		{
			name:                   "invalid builder short circuits",
			valid:                  false,
			dualStackAddressFamily: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithBGPNeighborDualStackAddressFamily(
				testCase.dualStackAddressFamily, testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			assert.Equal(t, testCase.dualStackAddressFamily,
				result.Definition.Spec.BGP.Routers[0].Neighbors[0].DualStackAddressFamily)
		})
	}
}

func TestFrrConfigurationWithPort(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		port          uint16
		routerIndex   uint
		neighborIndex uint
		expectedError string
	}{
		{
			name:  "sets port",
			valid: true,
			port:  179,
		},
		{
			name:          "port above maximum",
			valid:         true,
			port:          16385,
			expectedError: "invalid port number: 16385",
		},
		{
			name:          "invalid router index",
			valid:         true,
			port:          179,
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:          "invalid neighbor index",
			valid:         true,
			port:          179,
			neighborIndex: 1,
			expectedError: invalidNeighborIndexError,
		},
		{
			name:  "invalid builder short circuits",
			valid: false,
			port:  179,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithPort(testCase.port, testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			assert.Equal(t, &testCase.port, result.Definition.Spec.BGP.Routers[0].Neighbors[0].Port)
		})
	}
}

func TestFrrConfigurationWithToReceiveModeAll(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		routerIndex   uint
		neighborIndex uint
		expectedError string
	}{
		{
			name:  "allows all routes",
			valid: true,
		},
		{
			name:          "invalid router index",
			valid:         true,
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:          "invalid neighbor index",
			valid:         true,
			neighborIndex: 1,
			expectedError: invalidNeighborIndexError,
		},
		{
			name:  "invalid builder short circuits",
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithToReceiveModeAll(testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			assert.Equal(t, frrtypes.AllowAll,
				result.Definition.Spec.BGP.Routers[0].Neighbors[0].ToReceive.Allowed.Mode)
		})
	}
}

func TestFrrConfigurationWithToReceiveModeFiltered(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		prefixes      []string
		routerIndex   uint
		neighborIndex uint
		expectedError string
	}{
		{
			name:     "allows filtered routes",
			valid:    true,
			prefixes: []string{"192.168.1.0/24", "10.0.0.0/24"},
		},
		{
			name:          "invalid CIDR prefix",
			valid:         true,
			prefixes:      []string{"192.168.1."},
			expectedError: "the prefix 192.168.1. is not a valid CIDR",
		},
		{
			name:          "invalid router index",
			valid:         true,
			prefixes:      []string{"192.168.1.0/24"},
			routerIndex:   1,
			expectedError: invalidRouterIndexError,
		},
		{
			name:          "invalid neighbor index",
			valid:         true,
			prefixes:      []string{"192.168.1.0/24"},
			neighborIndex: 1,
			expectedError: invalidNeighborIndexError,
		},
		{
			name:     "invalid builder short circuits",
			valid:    false,
			prefixes: []string{"192.168.1.0/24"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildFrrConfigurationWithNeighborForTest(testCase.valid)
			result := builder.WithToReceiveModeFiltered(
				testCase.prefixes, testCase.routerIndex, testCase.neighborIndex)

			require.Same(t, builder, result)

			if !requireFrrConfigurationModifierResult(t, result, testCase.valid, testCase.expectedError) {
				return
			}

			allowedPrefixes := result.Definition.Spec.BGP.Routers[0].Neighbors[0].ToReceive.Allowed.Prefixes
			require.Len(t, allowedPrefixes, len(testCase.prefixes))

			for index, prefix := range testCase.prefixes {
				assert.Equal(t, prefix, allowedPrefixes[index].Prefix)
			}
		})
	}
}

// requireFrrConfigurationModifierResult checks the expected builder error state and reports whether field assertions
// should continue.
func requireFrrConfigurationModifierResult(
	t *testing.T, builder *FrrConfigurationBuilder, valid bool, expectedError string) bool {
	t.Helper()

	if !valid {
		require.Error(t, builder.GetError())
		assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

		return false
	}

	if expectedError != "" {
		assert.EqualError(t, builder.GetError(), expectedError)

		return false
	}

	require.NoError(t, builder.GetError())

	return true
}

// buildFrrConfigurationBuilderForTest returns a valid or builder-error-state FRRConfiguration builder.
func buildFrrConfigurationBuilderForTest(valid bool) *FrrConfigurationBuilder {
	apiClient := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: frrTestSchemes})

	if valid {
		return buildValidFRRConfigurationBuilder(apiClient)
	}

	return buildInValidFRRConfigurationBuilder(apiClient)
}

// buildFrrConfigurationWithRouterForTest returns a valid builder with one BGP router, or an invalid builder.
func buildFrrConfigurationWithRouterForTest(valid bool) *FrrConfigurationBuilder {
	builder := buildFrrConfigurationBuilderForTest(valid)

	if !valid {
		return builder
	}

	return builder.WithBGPRouter(64550)
}

// buildFrrConfigurationWithNeighborForTest returns a valid builder with one BGP neighbor, or an invalid builder.
func buildFrrConfigurationWithNeighborForTest(valid bool) *FrrConfigurationBuilder {
	builder := buildFrrConfigurationWithRouterForTest(valid)

	if !valid {
		return builder
	}

	return builder.WithBGPNeighbor(defaultFrrBGPPeerIP, 64500, 0)
}

// buildValidFRRConfigurationBuilder returns a valid FRRConfiguration builder for testing.
func buildValidFRRConfigurationBuilder(apiClient *clients.Settings) *FrrConfigurationBuilder {
	return NewFrrConfigurationBuilder(apiClient, defaultFrrConfigurationName, defaultFrrConfigurationNsName)
}

// buildInValidFRRConfigurationBuilder returns an invalid FRRConfiguration builder for testing.
func buildInValidFRRConfigurationBuilder(apiClient *clients.Settings) *FrrConfigurationBuilder {
	return NewFrrConfigurationBuilder(apiClient, "", defaultFrrConfigurationNsName)
}
