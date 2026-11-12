package sriov

import (
	"context"
	"testing"
	"time"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const errEmptyNetworkResourceName = "SrIovNetwork 'resName' cannot be empty"

var (
	defaultNetName         = "sriovnet"
	defaultNetNsName       = "testnamespace"
	defaultNetTargetNsName = "targetns"
	defaultNetResName      = "resname"
)

var testSchemes = []clients.SchemeAttacher{
	srIovV1.AddToScheme,
}

func TestNewNetworkBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *NetworkBuilder {
				return NewNetworkBuilder(
					apiClient, name, nsname, defaultNetTargetNsName, defaultNetResName)
			},
			srIovV1.AddToScheme,
			sriovNetworkGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name              string
		builderName       string
		targetNamespace   string
		resourceName      string
		expectedErrorText string
		assertError       func(error) bool
	}{
		{
			name:            "sets required fields",
			builderName:     defaultNetName,
			targetNamespace: defaultNetTargetNsName,
			resourceName:    defaultNetResName,
		},
		{
			name:              "empty resource name returns error",
			builderName:       defaultNetName,
			targetNamespace:   defaultNetTargetNsName,
			expectedErrorText: errEmptyNetworkResourceName,
		},
		{
			name:            "empty target namespace is allowed",
			builderName:     defaultNetName,
			targetNamespace: "",
			resourceName:    defaultNetResName,
		},
		{
			name:            "common name validation takes precedence",
			builderName:     "",
			targetNamespace: defaultNetTargetNsName,
			resourceName:    "",
			assertError:     commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewNetworkBuilder(
				buildNetworkSchemeTestClient(),
				testCase.builderName,
				defaultNetNsName,
				testCase.targetNamespace,
				testCase.resourceName,
			)

			require.NotNil(t, builder)

			if testCase.assertError != nil {
				assert.True(t, testCase.assertError(builder.GetError()), "unexpected error: %v", builder.GetError())

				return
			}

			if testCase.expectedErrorText != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedErrorText)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.targetNamespace, builder.Definition.Spec.NetworkNamespace)
			assert.Equal(t, testCase.resourceName, builder.Definition.Spec.ResourceName)
		})
	}
}

func TestPullNetwork(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullNetwork, srIovV1.AddToScheme, sriovNetworkGVK).ExecuteTests(t)
}

func TestNetworkBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newNetworkCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestNetworkWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newNetworkCommonTestConfig()).ExecuteTests(t)
}

func TestNetworkList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		List, srIovV1.AddToScheme, sriovNetworkGVK).ExecuteTests(t)

	t.Run("rejects multiple options", func(t *testing.T) {
		t.Parallel()

		_, err := List(
			buildNetworkSchemeTestClient(),
			defaultNetNsName,
			client.ListOptions{Namespace: defaultNetNsName},
			client.ListOptions{Continue: "true"},
		)

		assert.EqualError(t, err, "error: more than one ListOptions was passed")
	})

	t.Run("lists only the requested namespace", func(t *testing.T) {
		t.Parallel()

		otherNamespaceNetwork := &srIovV1.SriovNetwork{
			ObjectMeta: metav1.ObjectMeta{Name: "other-network", Namespace: "other-namespace"},
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  append(buildDummySrIovNetworkObject(), otherNamespaceNetwork),
			SchemeAttachers: testSchemes,
		})

		builders, err := List(testSettings, defaultNetNsName)

		require.NoError(t, err)
		require.Len(t, builders, 1)
		assert.Equal(t, defaultNetName, builders[0].Definition.Name)
		assert.Equal(t, defaultNetNsName, builders[0].Definition.Namespace)
	})
}

func TestNetworkWithTargetNamespace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		validBuilder      bool
		targetNamespace   string
		expectedErrorText string
	}{
		{
			name:            "sets target namespace",
			validBuilder:    true,
			targetNamespace: "new-target-namespace",
		},
		{
			name:              "empty target namespace returns error",
			validBuilder:      true,
			expectedErrorText: errEmptyTargetNsname,
		},
		{
			name:              "invalid builder short circuits",
			targetNamespace:   "new-target-namespace",
			expectedErrorText: errEmptyNetworkResourceName,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := networkBuilderForModifierTest(testCase.validBuilder)
			result := builder.WithTargetNamespace(testCase.targetNamespace)

			require.Same(t, builder, result)

			if testCase.expectedErrorText == "" {
				require.NoError(t, result.GetError())
				assert.Equal(t, testCase.targetNamespace, result.Definition.Spec.NetworkNamespace)

				return
			}

			assert.EqualError(t, result.GetError(), testCase.expectedErrorText)
		})
	}
}

func TestNetworkWithVLAN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		validBuilder      bool
		vlanID            uint16
		expectedErrorText string
	}{
		{name: "sets VLAN", validBuilder: true, vlanID: 100},
		{name: "accepts maximum VLAN", validBuilder: true, vlanID: 4094},
		{
			name: "rejects out of range VLAN", validBuilder: true, vlanID: 4095,
			expectedErrorText: "invalid vlanID, allowed vlanID values are between 0-4094",
		},
		{
			name: "invalid builder short circuits", vlanID: 100,
			expectedErrorText: errEmptyNetworkResourceName,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithVLAN(testCase.vlanID)

			if testCase.expectedErrorText == "" {
				require.NoError(t, result.GetError())
				assert.Equal(t, int(testCase.vlanID), result.Definition.Spec.Vlan)

				return
			}

			assert.EqualError(t, result.GetError(), testCase.expectedErrorText)
		})
	}
}

func TestNetworkWithVlanProto(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		validBuilder      bool
		vlanProtocol      string
		expectedErrorText string
	}{
		{name: "sets 802.1q", validBuilder: true, vlanProtocol: "802.1q"},
		{name: "sets 802.1Q", validBuilder: true, vlanProtocol: "802.1Q"},
		{name: "sets 802.1ad", validBuilder: true, vlanProtocol: "802.1ad"},
		{name: "sets 802.1AD", validBuilder: true, vlanProtocol: "802.1AD"},
		{
			name: "rejects invalid protocol", validBuilder: true, vlanProtocol: "802.1",
			expectedErrorText: "invalid 'vlanProtocol' parameters",
		},
		{
			name: "invalid builder short circuits", vlanProtocol: "802.1q",
			expectedErrorText: errEmptyNetworkResourceName,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithVlanProto(testCase.vlanProtocol)

			if testCase.expectedErrorText == "" {
				require.NoError(t, result.GetError())
				assert.Equal(t, testCase.vlanProtocol, result.Definition.Spec.VlanProto)

				return
			}

			assert.EqualError(t, result.GetError(), testCase.expectedErrorText)
		})
	}
}

func TestNetworkWithSpoof(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		validBuilder bool
		enabled      bool
		expected     string
	}{
		{name: "enables spoof check", validBuilder: true, enabled: true, expected: "on"},
		{name: "disables spoof check", validBuilder: true, enabled: false, expected: spoofChkOff},
		{name: "invalid builder short circuits", enabled: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithSpoof(testCase.enabled)

			if !testCase.validBuilder {
				assert.EqualError(t, result.GetError(), errEmptyNetworkResourceName)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.expected, result.Definition.Spec.SpoofChk)
		})
	}
}

func TestNetworkWithMetaPluginAllMultiFlag(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		validBuilder bool
		allMulti     bool
		expected     string
	}{
		{
			name: "enables allmulti", validBuilder: true, allMulti: true,
			expected: `{ "type": "tuning", "allmulti": true }`,
		},
		{
			name: "disables allmulti", validBuilder: true, allMulti: false,
			expected: `{ "type": "tuning", "allmulti": false }`,
		},
		{name: "invalid builder short circuits", allMulti: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithMetaPluginAllMultiFlag(testCase.allMulti)

			if !testCase.validBuilder {
				assert.EqualError(t, result.GetError(), errEmptyNetworkResourceName)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.expected, result.Definition.Spec.MetaPluginsConfig)
		})
	}
}

func TestNetworkWithMetaPluginRdma(t *testing.T) {
	t.Parallel()

	t.Run("sets RDMA meta plugin", func(t *testing.T) {
		t.Parallel()

		result := networkBuilderForModifierTest(true).WithMetaPluginRdma()

		require.NoError(t, result.GetError())
		assert.Equal(t, `{ "type": "rdma" }`, result.Definition.Spec.MetaPluginsConfig)
	})

	t.Run("invalid builder short circuits", func(t *testing.T) {
		t.Parallel()

		result := networkBuilderForModifierTest(false).WithMetaPluginRdma()

		assert.EqualError(t, result.GetError(), errEmptyNetworkResourceName)
	})
}

func TestNetworkWithLinkState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		validBuilder      bool
		linkState         string
		expectedErrorText string
	}{
		{name: "sets enable", validBuilder: true, linkState: "enable"},
		{name: "sets disable", validBuilder: true, linkState: "disable"},
		{name: "sets auto", validBuilder: true, linkState: "auto"},
		{name: "rejects invalid state", validBuilder: true, linkState: "invalid", expectedErrorText: errInvalidLinkState},
		{name: "rejects empty state", validBuilder: true, expectedErrorText: errInvalidLinkState},
		{name: "invalid builder short circuits", linkState: "auto", expectedErrorText: errEmptyNetworkResourceName},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithLinkState(testCase.linkState)

			if testCase.expectedErrorText == "" {
				require.NoError(t, result.GetError())
				assert.Equal(t, testCase.linkState, result.Definition.Spec.LinkState)

				return
			}

			assert.EqualError(t, result.GetError(), testCase.expectedErrorText)
		})
	}
}

func TestNetworkWithMaxTxRate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		validBuilder bool
		maxTxRate    uint16
	}{
		{name: "sets zero", validBuilder: true, maxTxRate: 0},
		{name: "sets rate", validBuilder: true, maxTxRate: 100},
		{name: "sets maximum rate", validBuilder: true, maxTxRate: 10000},
		{name: "invalid builder short circuits", maxTxRate: 100},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithMaxTxRate(testCase.maxTxRate)

			if !testCase.validBuilder {
				assert.EqualError(t, result.GetError(), errEmptyNetworkResourceName)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.MaxTxRate)
			assert.Equal(t, int(testCase.maxTxRate), *result.Definition.Spec.MaxTxRate)
		})
	}
}

func TestNetworkWithMinTxRate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		validBuilder bool
		minTxRate    uint16
	}{
		{name: "sets zero", validBuilder: true, minTxRate: 0},
		{name: "sets rate", validBuilder: true, minTxRate: 100},
		{name: "sets maximum rate", validBuilder: true, minTxRate: 10000},
		{name: "invalid builder short circuits", minTxRate: 100},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithMinTxRate(testCase.minTxRate)

			if !testCase.validBuilder {
				assert.EqualError(t, result.GetError(), errEmptyNetworkResourceName)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.MaxTxRate)
			assert.Equal(t, int(testCase.minTxRate), *result.Definition.Spec.MaxTxRate)
		})
	}
}

func TestNetworkWithTrustFlag(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		validBuilder bool
		enabled      bool
		expected     string
	}{
		{name: "enables trust", validBuilder: true, enabled: true, expected: "on"},
		{name: "disables trust", validBuilder: true, enabled: false, expected: spoofChkOff},
		{name: "invalid builder short circuits", enabled: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithTrustFlag(testCase.enabled)

			if !testCase.validBuilder {
				assert.EqualError(t, result.GetError(), errEmptyNetworkResourceName)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.expected, result.Definition.Spec.Trust)
		})
	}
}

func TestNetworkWithVlanQoS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		validBuilder      bool
		vlanQoS           uint16
		expectedErrorText string
	}{
		{name: "sets zero", validBuilder: true, vlanQoS: 0},
		{name: "sets maximum", validBuilder: true, vlanQoS: 7},
		{
			name: "rejects out of range QoS", validBuilder: true, vlanQoS: 8,
			expectedErrorText: "Invalid QoS class. Supported vlan QoS class values are between 0...7",
		},
		{name: "invalid builder short circuits", vlanQoS: 0, expectedErrorText: errEmptyNetworkResourceName},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithVlanQoS(testCase.vlanQoS)

			if testCase.expectedErrorText == "" {
				require.NoError(t, result.GetError())
				assert.Equal(t, int(testCase.vlanQoS), result.Definition.Spec.VlanQoS)

				return
			}

			assert.EqualError(t, result.GetError(), testCase.expectedErrorText)
		})
	}
}

func TestNetworkWithLogLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		validBuilder      bool
		logLevel          string
		expectedErrorText string
	}{
		{name: "sets panic", validBuilder: true, logLevel: "panic"},
		{name: "sets error", validBuilder: true, logLevel: "error"},
		{name: "sets warning", validBuilder: true, logLevel: "warning"},
		{name: "sets info", validBuilder: true, logLevel: "info"},
		{name: "sets debug", validBuilder: true, logLevel: "debug"},
		{name: "sets empty", validBuilder: true, logLevel: ""},
		{
			name: "rejects invalid log level", validBuilder: true, logLevel: "invalid",
			expectedErrorText: "invalid logLevel value, allowed logLevel values are:" +
				" panic, error, warning, info, debug or empty",
		},
		{name: "invalid builder short circuits", logLevel: "info", expectedErrorText: errEmptyNetworkResourceName},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithLogLevel(testCase.logLevel)

			if testCase.expectedErrorText == "" {
				require.NoError(t, result.GetError())
				assert.Equal(t, testCase.logLevel, result.Definition.Spec.LogLevel)

				return
			}

			assert.EqualError(t, result.GetError(), testCase.expectedErrorText)
		})
	}
}

func TestNetworkCapabilities(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		validBuilder bool
		modifier     func(*NetworkBuilder) *NetworkBuilder
		expected     string
		field        func(*NetworkBuilder) string
	}{
		{
			name: "adds IP address support", validBuilder: true,
			modifier: func(builder *NetworkBuilder) *NetworkBuilder { return builder.WithIPAddressSupport() },
			expected: `{ "ips": true }`,
			field:    func(builder *NetworkBuilder) string { return builder.Definition.Spec.Capabilities },
		},
		{
			name: "adds MAC address support", validBuilder: true,
			modifier: func(builder *NetworkBuilder) *NetworkBuilder { return builder.WithMacAddressSupport() },
			expected: `{ "mac": true }`,
			field:    func(builder *NetworkBuilder) string { return builder.Definition.Spec.Capabilities },
		},
		{
			name: "adds static IPAM", validBuilder: true,
			modifier: func(builder *NetworkBuilder) *NetworkBuilder { return builder.WithStaticIpam() },
			expected: `{ "type": "static" }`,
			field:    func(builder *NetworkBuilder) string { return builder.Definition.Spec.IPAM },
		},
		{name: "IP address support invalid builder short circuits", modifier: (*NetworkBuilder).WithIPAddressSupport},
		{name: "MAC address support invalid builder short circuits", modifier: (*NetworkBuilder).WithMacAddressSupport},
		{name: "static IPAM invalid builder short circuits", modifier: (*NetworkBuilder).WithStaticIpam},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.modifier(networkBuilderForModifierTest(testCase.validBuilder))

			if !testCase.validBuilder {
				assert.EqualError(t, result.GetError(), errEmptyNetworkResourceName)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.expected, testCase.field(result))
		})
	}
}

func TestNetworkWithWhereaboutsIPAM(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		validBuilder      bool
		ipRange           string
		gateway           string
		exclude           string
		networkName       string
		expectedErrorText string
		expectedIPAM      string
	}{
		{
			name: "sets all fields", validBuilder: true, ipRange: "192.168.100.0/24", gateway: "192.168.100.1",
			exclude: "192.168.100.1", networkName: "test-network",
			expectedIPAM: `{"exclude":["192.168.100.1"],"gateway":"192.168.100.1",` +
				`"network_name":"test-network","range":"192.168.100.0/24","type":"whereabouts"}`,
		},
		{
			name: "sets required fields", validBuilder: true, ipRange: "10.0.0.0/16", gateway: "10.0.0.1",
			expectedIPAM: `{"gateway":"10.0.0.1","range":"10.0.0.0/16","type":"whereabouts"}`,
		},
		{
			name: "rejects empty range", validBuilder: true, gateway: "192.168.100.1",
			expectedErrorText: "failed to configure IPAM, 'ipRange' parameter is empty",
		},
		{
			name: "rejects empty gateway", validBuilder: true, ipRange: "192.168.100.0/24",
			expectedErrorText: "failed to configure IPAM, 'gateway' parameter is empty",
		},
		{name: "invalid builder short circuits", ipRange: "192.168.100.0/24", gateway: "192.168.100.1",
			expectedErrorText: errEmptyNetworkResourceName},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := networkBuilderForModifierTest(testCase.validBuilder).WithWhereaboutsIPAM(
				testCase.ipRange, testCase.gateway, testCase.exclude, testCase.networkName)

			if testCase.expectedErrorText == "" {
				require.NoError(t, result.GetError())
				assert.Equal(t, testCase.expectedIPAM, result.Definition.Spec.IPAM)

				return
			}

			assert.EqualError(t, result.GetError(), testCase.expectedErrorText)
		})
	}
}

func TestNetworkDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		validBuilder      bool
		objectExists      bool
		expectedErrorText string
	}{
		{name: "deletes existing network", validBuilder: true, objectExists: true},
		{name: "network already absent", validBuilder: true},
		{name: "invalid builder", expectedErrorText: "failed to validate: " + errEmptyNetworkResourceName},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *NetworkBuilder

			switch {
			case testCase.objectExists:
				builder = buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())
			case testCase.validBuilder:
				builder = buildValidSriovNetworkTestBuilder(buildNetworkSchemeTestClient())
			default:
				builder = buildInvalidSrIovNetworkTestBuilder(buildNetworkSchemeTestClient())
			}

			err := builder.DeleteAndWait(time.Second)

			if testCase.expectedErrorText == "" {
				require.NoError(t, err)
				assert.Nil(t, builder.Object)

				return
			}

			assert.EqualError(t, err, testCase.expectedErrorText)
		})
	}
}

func TestNetworkWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		validBuilder      bool
		objectExists      bool
		expectedError     error
		expectedErrorText string
	}{
		{name: "network already absent", validBuilder: true},
		{name: "times out while network remains", validBuilder: true, objectExists: true, expectedError: context.DeadlineExceeded},
		{name: "invalid builder", expectedErrorText: "failed to validate: " + errEmptyNetworkResourceName},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *NetworkBuilder

			switch {
			case testCase.objectExists:
				builder = buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())
			case testCase.validBuilder:
				builder = buildValidSriovNetworkTestBuilder(buildNetworkSchemeTestClient())
			default:
				builder = buildInvalidSrIovNetworkTestBuilder(buildNetworkSchemeTestClient())
			}

			err := builder.WaitUntilDeleted(time.Second)

			if testCase.expectedErrorText != "" {
				assert.EqualError(t, err, testCase.expectedErrorText)

				return
			}

			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestGetSriovNetworksGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, GetSriovNetworksGVR(), schema.GroupVersionResource{
		Group: "sriovnetwork.openshift.io", Version: "v1", Resource: "sriovnetworks",
	})
}

func newNetworkCommonTestConfig() testhelper.CommonTestConfig[
	srIovV1.SriovNetwork, NetworkBuilder, *srIovV1.SriovNetwork, *NetworkBuilder,
] {
	return testhelper.NewCommonTestConfig[srIovV1.SriovNetwork, NetworkBuilder](
		srIovV1.AddToScheme,
		sriovNetworkGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

func buildValidSriovNetworkTestBuilder(apiClient *clients.Settings) *NetworkBuilder {
	return NewNetworkBuilder(
		apiClient, defaultNetName, defaultNetNsName, defaultNetTargetNsName, defaultNetResName)
}

func buildInvalidSrIovNetworkTestBuilder(apiClient *clients.Settings) *NetworkBuilder {
	return NewNetworkBuilder(apiClient, defaultNetName, defaultNetNsName, defaultNetTargetNsName, "")
}

func networkBuilderForModifierTest(valid bool) *NetworkBuilder {
	if valid {
		return buildValidSriovNetworkTestBuilder(buildNetworkSchemeTestClient())
	}

	return buildInvalidSrIovNetworkTestBuilder(buildNetworkSchemeTestClient())
}

func buildNetworkSchemeTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummySrIovNetworkObject(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummySrIovNetworkObject() []runtime.Object {
	return append([]runtime.Object{}, &srIovV1.SriovNetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultNetName,
			Namespace: defaultNetNsName,
		},
		Spec: srIovV1.SriovNetworkSpec{
			ResourceName:     defaultNetResName,
			NetworkNamespace: defaultNetTargetNsName,
		},
	})
}
