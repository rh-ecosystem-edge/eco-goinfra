package metallb

import (
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypesv1beta2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	defaultBGPPeerName     = "default-bgp-peer"
	defaultBGPPeerNsName   = "bgppeer-test-namespace"
	defaultBGPPeerPeerIP   = "172.16.1.100"
	defaultBGPPeerASN      = uint32(1000)
	defaultBGPPeerRemoteAS = uint32(2000)
)

func TestNewBPGPeerBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *BGPPeerBuilder {
				return NewBPGPeerBuilder(
					apiClient, name, nsname, defaultBGPPeerPeerIP, defaultBGPPeerASN, defaultBGPPeerRemoteAS)
			},
			mlbtypesv1beta2.AddToScheme,
			bgpPeerGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		builderName   string
		peerIP        string
		expectedError string
		builderErr    func(error) bool
	}{
		{
			name:        "valid builder",
			builderName: defaultBGPPeerName,
			peerIP:      defaultBGPPeerPeerIP,
		},
		{
			name:          "empty peer IP returns error",
			builderName:   defaultBGPPeerName,
			peerIP:        "",
			expectedError: bgppeerPeeripOfTheBgppeerContains,
		},
		{
			name:          "invalid peer IP returns error",
			builderName:   defaultBGPPeerName,
			peerIP:        "192.168.1.1000",
			expectedError: bgppeerPeeripOfTheBgppeerContains,
		},
		{
			name:        "common validation happens before peer IP validation",
			builderName: "",
			peerIP:      "invalid",
			builderErr:  commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewBPGPeerBuilder(
				newBGPPeerTestClient(),
				testCase.builderName,
				defaultBGPPeerNsName,
				testCase.peerIP,
				defaultBGPPeerASN,
				defaultBGPPeerRemoteAS,
			)

			require.NotNil(t, builder)

			if testCase.builderErr != nil {
				assert.True(t, testCase.builderErr(builder.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.builderName, builder.Definition.Name)
			assert.Equal(t, defaultBGPPeerNsName, builder.Definition.Namespace)
			assert.Equal(t, testCase.peerIP, builder.Definition.Spec.Address)
			assert.Equal(t, defaultBGPPeerASN, builder.Definition.Spec.MyASN)
			assert.Equal(t, defaultBGPPeerRemoteAS, builder.Definition.Spec.ASN)
		})
	}
}

func TestNewBGPPeerBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		func(apiClient *clients.Settings, name, nsname string) *BGPPeerBuilder {
			return NewBGPPeerBuilder(
				apiClient, name, nsname, defaultBGPPeerASN, defaultBGPPeerRemoteAS)
		},
		mlbtypesv1beta2.AddToScheme,
		bgpPeerGVK,
	).ExecuteTests(t)
}

func TestPullBGPPeer(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullBGPPeer, mlbtypesv1beta2.AddToScheme, bgpPeerGVK).ExecuteTests(t)
}

func TestBGPPeerBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newBGPPeerCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestBGPPeerWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newBGPPeerCommonTestConfig()).ExecuteTests(t)
}

func TestBGPPeerWithDynamicASN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		dynamicASN    mlbtypesv1beta2.DynamicASNMode
		expectedError string
	}{
		{name: "internal ASN", valid: true, dynamicASN: mlbtypesv1beta2.InternalASNMode},
		{name: "external ASN", valid: true, dynamicASN: mlbtypesv1beta2.ExternalASNMode},
		{
			name:          "invalid ASN mode",
			valid:         true,
			dynamicASN:    "ebgp",
			expectedError: "bgpPeer 'dynamicASN' must be either internal or external",
		},
		{name: "invalid builder short circuits", valid: false, dynamicASN: mlbtypesv1beta2.InternalASNMode},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				testCase.expectedError,
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithDynamicASN(testCase.dynamicASN)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.dynamicASN, builder.Definition.Spec.DynamicASN)
					assert.Zero(t, builder.Definition.Spec.ASN)
				},
			)
		})
	}
}

func TestBGPPeerWithBGPPeerIP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		peerIP        string
		expectedError string
	}{
		{name: "valid peer IP", valid: true, peerIP: "172.16.100.1"},
		{
			name:          "invalid peer IP",
			valid:         true,
			peerIP:        "test",
			expectedError: bgppeerBgppeeripOfTheBgppeerContains,
		},
		{
			name:          "empty peer IP",
			valid:         true,
			peerIP:        "",
			expectedError: bgppeerBgppeeripOfTheBgppeerContains,
		},
		{name: "invalid builder short circuits", valid: false, peerIP: "172.16.100.1"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				testCase.expectedError,
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithBGPPeerIP(testCase.peerIP)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.peerIP, builder.Definition.Spec.Address)
				},
			)
		})
	}
}

func TestBGPPeerWithIPUnnumbered(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		interfaceName string
		expectedError string
	}{
		{name: "valid interface", valid: true, interfaceName: "net1"},
		{
			name:          "empty interface",
			valid:         true,
			interfaceName: "",
			expectedError: "interface can not be empty string",
		},
		{name: "invalid builder short circuits", valid: false, interfaceName: "net1"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				testCase.expectedError,
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithIPUnnumbered(testCase.interfaceName)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.interfaceName, builder.Definition.Spec.Interface)
				},
			)
		})
	}
}

func TestBGPPeerWithRouterID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		routerID      string
		expectedError string
	}{
		{name: "valid router ID", valid: true, routerID: "1.1.1.1"},
		{
			name:          "empty router ID",
			valid:         true,
			routerID:      "",
			expectedError: "the routerID of the BGPPeer contains invalid ip address ",
		},
		{
			name:          "invalid router ID",
			valid:         true,
			routerID:      "1.1.1.300",
			expectedError: "the routerID of the BGPPeer contains invalid ip address 1.1.1.300",
		},
		{name: "invalid builder short circuits", valid: false, routerID: "1.1.1.1"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				testCase.expectedError,
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithRouterID(testCase.routerID)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.routerID, builder.Definition.Spec.RouterID)
				},
			)
		})
	}
}

func TestBGPPeerWithBFDProfile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		bfdProfile    string
		expectedError string
	}{
		{name: "valid BFD profile", valid: true, bfdProfile: "testprofile"},
		{
			name:          "empty BFD profile",
			valid:         true,
			bfdProfile:    "",
			expectedError: "The bfdProfile is empty string",
		},
		{name: "invalid builder short circuits", valid: false, bfdProfile: "testprofile"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				testCase.expectedError,
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithBFDProfile(testCase.bfdProfile)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.bfdProfile, builder.Definition.Spec.BFDProfile)
				},
			)
		})
	}
}

func TestBGPPeerWithSRCAddress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		srcAddress    string
		expectedError string
	}{
		{name: "valid source address", valid: true, srcAddress: "1.1.1.1"},
		{
			name:          "empty source address",
			valid:         true,
			srcAddress:    "",
			expectedError: "the srcAddress of the BGPPeer contains invalid ip address ",
		},
		{
			name:          "invalid source address",
			valid:         true,
			srcAddress:    "172.16.100.300",
			expectedError: "the srcAddress of the BGPPeer contains invalid ip address 172.16.100.300",
		},
		{name: "invalid builder short circuits", valid: false, srcAddress: "1.1.1.1"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				testCase.expectedError,
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithSRCAddress(testCase.srcAddress)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.srcAddress, builder.Definition.Spec.SrcAddress)
				},
			)
		})
	}
}

func TestBGPPeerWithPort(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		valid bool
		port  uint16
	}{
		{name: "sets port", valid: true, port: 179},
		{name: "invalid builder short circuits", valid: false, port: 179},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				"",
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithPort(testCase.port)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.port, builder.Definition.Spec.Port)
				},
			)
		})
	}
}

func TestBGPPeerWithHoldTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		valid    bool
		holdTime metav1.Duration
	}{
		{name: "sets hold time", valid: true, holdTime: metav1.Duration{Duration: 5 * time.Minute}},
		{name: "invalid builder short circuits", valid: false, holdTime: metav1.Duration{Duration: 5 * time.Minute}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				"",
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithHoldTime(testCase.holdTime)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					require.NotNil(t, builder.Definition.Spec.HoldTime)
					assert.Equal(t, testCase.holdTime, *builder.Definition.Spec.HoldTime)
				},
			)
		})
	}
}

func TestBGPPeerWithKeepalive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		valid     bool
		keepalive metav1.Duration
	}{
		{name: "sets keepalive", valid: true, keepalive: metav1.Duration{Duration: 5 * time.Minute}},
		{name: "invalid builder short circuits", valid: false, keepalive: metav1.Duration{Duration: 5 * time.Minute}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				"",
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithKeepalive(testCase.keepalive)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					require.NotNil(t, builder.Definition.Spec.KeepaliveTime)
					assert.Equal(t, testCase.keepalive, *builder.Definition.Spec.KeepaliveTime)
				},
			)
		})
	}
}

func TestBGPPeerWithConnectTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		connectTime   metav1.Duration
		expectedError string
	}{
		{name: "valid connect time", valid: true, connectTime: metav1.Duration{Duration: 10 * time.Second}},
		{name: "minimum connect time", valid: true, connectTime: metav1.Duration{Duration: time.Second}},
		{name: "maximum connect time", valid: true, connectTime: metav1.Duration{Duration: 65535 * time.Second}},
		{
			name:          "zero connect time",
			valid:         true,
			connectTime:   metav1.Duration{},
			expectedError: bgppeerConnecttimeValueIsNotValid,
		},
		{
			name:          "connect time above maximum",
			valid:         true,
			connectTime:   metav1.Duration{Duration: 65536 * time.Second},
			expectedError: bgppeerConnecttimeValueIsNotValid,
		},
		{name: "invalid builder short circuits", valid: false, connectTime: metav1.Duration{Duration: 10 * time.Second}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				testCase.expectedError,
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithConnectTime(testCase.connectTime)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					require.NotNil(t, builder.Definition.Spec.ConnectTime)
					assert.Equal(t, testCase.connectTime, *builder.Definition.Spec.ConnectTime)
				},
			)
		})
	}
}

func TestBGPPeerWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		nodeSelector  map[string]string
		expectedError string
	}{
		{name: "valid node selector", valid: true, nodeSelector: map[string]string{"test": "test1"}},
		{
			name:          "empty node selector",
			valid:         true,
			nodeSelector:  map[string]string{},
			expectedError: "BGPPeer 'nodeSelector' cannot be empty map",
		},
		{name: "invalid builder short circuits", valid: false, nodeSelector: map[string]string{"test": "test1"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				testCase.expectedError,
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithNodeSelector(testCase.nodeSelector)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					require.Len(t, builder.Definition.Spec.NodeSelectors, 1)
					assert.Equal(t, testCase.nodeSelector, builder.Definition.Spec.NodeSelectors[0].MatchLabels)
				},
			)
		})
	}
}

func TestBGPPeerWithPassword(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		password      string
		expectedError string
	}{
		{name: "valid password", valid: true, password: "test"},
		{
			name:          "empty password",
			valid:         true,
			password:      "",
			expectedError: "password can not be empty string",
		},
		{name: "invalid builder short circuits", valid: false, password: "test"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				testCase.expectedError,
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithPassword(testCase.password)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.password, builder.Definition.Spec.Password)
				},
			)
		})
	}
}

func TestBGPPeerWithEBGPMultiHop(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		valid        bool
		ebgpMultiHop bool
	}{
		{name: "enables multihop", valid: true, ebgpMultiHop: true},
		{name: "disables multihop", valid: true, ebgpMultiHop: false},
		{name: "invalid builder short circuits", valid: false, ebgpMultiHop: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				"",
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithEBGPMultiHop(testCase.ebgpMultiHop)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.ebgpMultiHop, builder.Definition.Spec.EBGPMultiHop)
				},
			)
		})
	}
}

func TestBGPPeerWithGracefulRestart(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		valid           bool
		gracefulRestart bool
	}{
		{name: "enables graceful restart", valid: true, gracefulRestart: true},
		{name: "disables graceful restart", valid: true, gracefulRestart: false},
		{name: "invalid builder short circuits", valid: false, gracefulRestart: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				"",
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithGracefulRestart(testCase.gracefulRestart)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.gracefulRestart, builder.Definition.Spec.EnableGracefulRestart)
				},
			)
		})
	}
}

func TestBGPPeerWithDisableMP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		valid     bool
		disableMP bool
	}{
		{name: "enables disable MP", valid: true, disableMP: true},
		{name: "disables disable MP", valid: true, disableMP: false},
		{name: "invalid builder short circuits", valid: false, disableMP: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				"",
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithDisableMP(testCase.disableMP)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					//nolint:staticcheck // DisableMP is deprecated, keep for backward compatibility.
					assert.Equal(t, testCase.disableMP, builder.Definition.Spec.DisableMP)
				},
			)
		})
	}
}

func TestBGPPeerWithDualStackAddressFamily(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		valid                  bool
		dualStackAddressFamily bool
	}{
		{name: "enables dual stack address family", valid: true, dualStackAddressFamily: true},
		{name: "disables dual stack address family", valid: true, dualStackAddressFamily: false},
		{name: "invalid builder short circuits", valid: false, dualStackAddressFamily: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runBGPPeerModifierTest(
				t,
				testCase.valid,
				"",
				func(builder *BGPPeerBuilder) *BGPPeerBuilder {
					return builder.WithDualStackAddressFamily(testCase.dualStackAddressFamily)
				},
				func(t *testing.T, builder *BGPPeerBuilder) {
					t.Helper()
					assert.Equal(t, testCase.dualStackAddressFamily, builder.Definition.Spec.DualStackAddressFamily)
				},
			)
		})
	}
}

func TestBGPPeerGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "bgppeers",
	}, GetBGPPeerGVR())
}

func newBGPPeerCommonTestConfig() testhelper.CommonTestConfig[
	mlbtypesv1beta2.BGPPeer, BGPPeerBuilder, *mlbtypesv1beta2.BGPPeer, *BGPPeerBuilder,
] {
	return testhelper.NewCommonTestConfig[mlbtypesv1beta2.BGPPeer, BGPPeerBuilder](
		mlbtypesv1beta2.AddToScheme, bgpPeerGVK, testhelper.ResourceScopeNamespaced)
}

func newBGPPeerTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{mlbtypesv1beta2.AddToScheme},
	})
}

func buildValidBGPPeerBuilder(apiClient *clients.Settings) *BGPPeerBuilder {
	return NewBGPPeerBuilder(
		apiClient, defaultBGPPeerName, defaultBGPPeerNsName, defaultBGPPeerASN, defaultBGPPeerRemoteAS).
		WithBGPPeerIP(defaultBGPPeerPeerIP)
}

func buildInvalidBGPPeerBuilder(apiClient *clients.Settings) *BGPPeerBuilder {
	return NewBGPPeerBuilder(
		apiClient, "", defaultBGPPeerNsName, defaultBGPPeerASN, defaultBGPPeerRemoteAS)
}

func runBGPPeerModifierTest(
	t *testing.T,
	valid bool,
	expectedError string,
	modify func(*BGPPeerBuilder) *BGPPeerBuilder,
	assertSuccess func(*testing.T, *BGPPeerBuilder),
) {
	t.Helper()

	apiClient := newBGPPeerTestClient()

	builder := buildValidBGPPeerBuilder(apiClient)
	if !valid {
		builder = buildInvalidBGPPeerBuilder(apiClient)
	}

	result := modify(builder)
	require.Same(t, builder, result)

	if !valid {
		assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

		return
	}

	if expectedError != "" {
		assert.EqualError(t, result.GetError(), expectedError)

		return
	}

	require.NoError(t, result.GetError())
	assertSuccess(t, result)
}
