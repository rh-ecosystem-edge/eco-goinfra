package nmstate

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/nmstate/kubernetes-nmstate/api/shared"
	nmstatev1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultPolicyName = "policyname"
)

func TestNewPolicyBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common cluster-scoped builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewClusterScopedBuilderTestConfig(
			func(apiClient *clients.Settings, name string) *PolicyBuilder {
				return NewPolicyBuilder(apiClient, name, map[string]string{"test": "test"})
			},
			nmstatev1.AddToScheme,
			policyGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		nodeSelector  map[string]string
		expectedError error
	}{
		{
			name:         "valid node selector",
			nodeSelector: map[string]string{"test": "test"},
		},
		{
			name:          "nil node selector",
			expectedError: fmt.Errorf(nodeNetConfPolSelectorEmpty),
		},
		{
			name:          "empty node selector",
			nodeSelector:  map[string]string{},
			expectedError: fmt.Errorf(nodeNetConfPolSelectorEmpty),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testPolicy := NewPolicyBuilder(
				clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: []clients.SchemeAttacher{nmstatev1.AddToScheme},
				}),
				defaultPolicyName,
				testCase.nodeSelector,
			)

			require.NotNil(t, testPolicy)
			assert.Equal(t, testCase.expectedError, testPolicy.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.nodeSelector, testPolicy.Definition.Spec.NodeSelector)
			}
		})
	}
}

func TestPolicyBuilderMethods(t *testing.T) {
	t.Parallel()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(newPolicyCommonTestConfig())).
		With(testhelper.NewExistsTestConfig(newPolicyCommonTestConfig())).
		With(testhelper.NewCreateTestConfig(newPolicyCommonTestConfig())).
		With(testhelper.NewDeleteReturnerTestConfig(newPolicyCommonTestConfig())).
		With(testhelper.NewForceUpdateTestConfig(newPolicyCommonTestConfig())).
		Run(t)
}

func TestPolicyWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newPolicyCommonTestConfig()).ExecuteTests(t)
}

func TestPolicyWithInterfaceAndVFs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		sriovInterface    string
		numberOfVF        uint8
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
			numberOfVF:        10,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "The sriovInterface is empty string",
			sriovInterface:    "",
			numberOfVF:        10,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
			numberOfVF:        0,
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolSelectorEmpty,
			sriovInterface:    "ens1",
			numberOfVF:        10,
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithInterfaceAndVFs(testCase.sriovInterface, testCase.numberOfVF)
		assertPolicyError(t, testCase.expectedError, testPolicy)
		numberOfVFs := int(testCase.numberOfVF)
		desireState := &DesiredState{}

		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, desireState, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  testCase.sriovInterface,
						Type:  interfaceTypeEthernet,
						State: "up",
						Ethernet: Ethernet{
							Sriov: Sriov{
								TotalVfs: &numberOfVFs,
							},
						},
					},
				},
			})
		}
	}
}

//nolint:funlen // The table includes package validation and invalid-builder cases.
func TestPolicyWithWithBondInterface(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		slavePorts        []string
		bondName          string
		mode              string
		miimon            int
		lacpRate          string
		minLinks          int
		expectedError     string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			slavePorts:        []string{"ens1", "ens2"},
			bondName:          "bd1",
			mode:              bondModeActiveBackup,
			miimon:            100,
			lacpRate:          "fast",
			minLinks:          1,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "The bondName is empty sting",
			slavePorts:        []string{"ens1", "ens2"},
			bondName:          "",
			mode:              bondModeActiveBackup,
			miimon:            100,
			lacpRate:          "fast",
			minLinks:          1,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "invalid Bond mode parameter",
			slavePorts:        []string{"ens1", "ens2"},
			bondName:          "bd1",
			mode:              "",
			miimon:            100,
			lacpRate:          "fast",
			minLinks:          1,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			slavePorts:        []string{"ens1", "ens2"},
			bondName:          "bd1",
			mode:              bondModeActiveBackup,
			miimon:            0,
			lacpRate:          "",
			minLinks:          0,
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolSelectorEmpty,
			slavePorts:        []string{"ens1", "ens2"},
			bondName:          "bd1",
			mode:              bondModeActiveBackup,
		},
	}
	for _, testCase := range testCases {
		options := OptionsLinkAggregation{
			Miimon:   testCase.miimon,
			LacpRate: testCase.lacpRate,
			MinLinks: testCase.minLinks,
		}
		testPolicy := testCase.testNMStatePolicy.WithBondInterface(testCase.slavePorts, testCase.bondName,
			testCase.mode, options)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, desireState, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  testCase.bondName,
						Type:  interfaceTypeBond,
						State: "up",
						LinkAggregation: LinkAggregation{
							Mode: testCase.mode,
							Port: testCase.slavePorts,
							Options: OptionsLinkAggregation{
								Miimon:   testCase.miimon,
								LacpRate: testCase.lacpRate,
								MinLinks: testCase.minLinks,
							},
						},
					},
				},
			})
		}
	}
}

func TestPolicyWithVlanInterface(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		sriovInterface    string
		vlanID            uint16
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
			vlanID:            10,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     errEmptyBaseInterface,
			sriovInterface:    "",
			vlanID:            10,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     errInvalidVLANID,
			sriovInterface:    "ens1",
			vlanID:            4099,
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolSelectorEmpty,
			sriovInterface:    "ens1",
			vlanID:            10,
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithVlanInterface(testCase.sriovInterface, testCase.vlanID)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, desireState, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  fmt.Sprintf("%s.%d", testCase.sriovInterface, testCase.vlanID),
						Type:  interfaceTypeVlan,
						State: "up",
						Vlan: Vlan{
							BaseIface: testCase.sriovInterface,
							ID:        int(testCase.vlanID),
						},
					},
				},
			})
		}
	}
}

//nolint:funlen // The table includes package validation and invalid-builder cases.
func TestPolicyWithVlanInterfaceIP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		sriovInterface    string
		vlanID            uint16
		ipv4              string
		ipv6              string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
			vlanID:            10,
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     errEmptyBaseInterface,
			sriovInterface:    "",
			vlanID:            10,
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     errInvalidVLANID,
			sriovInterface:    "ens1",
			vlanID:            4099,
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "vlanInterfaceIP 'ipv4Addresses' is an invalid ipv4 address",
			sriovInterface:    "ens1",
			vlanID:            10,
			ipv4:              "",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "vlanInterfaceIP 'ipv6Addresses' is an invalid ipv6 address",
			sriovInterface:    "ens1",
			vlanID:            10,
			ipv4:              "10.10.10.10",
			ipv6:              "",
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolSelectorEmpty,
			sriovInterface:    "ens1",
			vlanID:            10,
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithVlanInterfaceIP(testCase.sriovInterface,
			testCase.ipv4, testCase.ipv6, testCase.vlanID)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  fmt.Sprintf("%s.%d", testCase.sriovInterface, testCase.vlanID),
						Type:  interfaceTypeVlan,
						State: "up",
						Vlan: Vlan{
							BaseIface: testCase.sriovInterface,
							ID:        int(testCase.vlanID),
						},
						Ipv4: InterfaceIpv4{
							Enabled: true,
							Address: []InterfaceIPAddress{{
								PrefixLen: 24,
								IP:        net.ParseIP(testCase.ipv4),
							}},
						},
						Ipv6: InterfaceIpv6{Enabled: true,
							Address: []InterfaceIPAddress{{
								PrefixLen: 64,
								IP:        net.ParseIP(testCase.ipv6),
							}},
						},
					},
				},
			}, desireState)
		}
	}
}

func TestPolicyWithInterfaceAltnames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		interfaceName     string
		altnames          []string
		expectedError     string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens1",
			altnames:          []string{"alt-a", "alt-b"},
			expectedError:     "",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "",
			altnames:          []string{"alt-a"},
			expectedError:     nodeNetConfPolIntError,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens1",
			altnames:          []string{},
			expectedError:     nodeNetConfPolAltnamesEmpty,
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens1",
			altnames:          []string{"alt-a"},
			expectedError:     nodeNetConfPolSelectorEmpty,
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithInterfaceAltnames(testCase.interfaceName, testCase.altnames)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		if testCase.expectedError != "" {
			continue
		}

		desireState := &DesiredState{}
		_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)

		wantAlts := make([]InterfaceAltName, len(testCase.altnames))
		for i, n := range testCase.altnames {
			wantAlts[i] = InterfaceAltName{Name: n}
		}

		assert.Equal(t, &DesiredState{
			Interfaces: []NetworkInterface{{
				Name:     testCase.interfaceName,
				Type:     interfaceTypeEthernet,
				State:    "up",
				AltNames: wantAlts,
			}},
		}, desireState)
	}
}

func TestPolicyWithMACAddressAltnames(t *testing.T) {
	t.Parallel()

	const testMAC = "52:54:00:ab:cd:ef"

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		interfaceName     string
		mac               string
		altnames          []string
		expectedError     string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens9",
			mac:               testMAC,
			altnames:          []string{"vf0"},
			expectedError:     "",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens9",
			mac:               "",
			altnames:          []string{"vf0"},
			expectedError:     "macaddress cannot be empty string",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "",
			mac:               testMAC,
			altnames:          []string{"vf0"},
			expectedError:     nodeNetConfPolIntError,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens9",
			mac:               "invalid",
			altnames:          []string{"vf0"},
			expectedError:     "macaddress is invalid",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens9",
			mac:               testMAC,
			altnames:          []string{},
			expectedError:     nodeNetConfPolAltnamesEmpty,
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens9",
			mac:               testMAC,
			altnames:          []string{"vf0"},
			expectedError:     nodeNetConfPolSelectorEmpty,
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithMACAddressAltnames(
			testCase.interfaceName, testCase.mac, testCase.altnames)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		if testCase.expectedError != "" {
			continue
		}

		desireState := &DesiredState{}
		_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)

		wantAlts := make([]InterfaceAltName, len(testCase.altnames))
		for i, n := range testCase.altnames {
			wantAlts[i] = InterfaceAltName{Name: n}
		}

		assert.Equal(t, &DesiredState{
			Interfaces: []NetworkInterface{{
				Name:       testCase.interfaceName,
				Type:       interfaceTypeEthernet,
				State:      "up",
				MacAddress: testCase.mac,
				Identifier: "mac-address",
				AltNames:   wantAlts,
			}},
		}, desireState)
	}
}

func TestPolicyWithPCIAddressAltnames(t *testing.T) {
	t.Parallel()

	const testPCI = "0000:03:00.0"

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		interfaceName     string
		pci               string
		altnames          []string
		expectedError     string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens8",
			pci:               testPCI,
			altnames:          []string{"nic-pf"},
			expectedError:     "",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens8",
			pci:               "",
			altnames:          []string{"nic-pf"},
			expectedError:     "pciAddress cannot be empty string",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "",
			pci:               testPCI,
			altnames:          []string{"nic-pf"},
			expectedError:     nodeNetConfPolIntError,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens8",
			pci:               "0000:03:00",
			altnames:          []string{"nic-pf"},
			expectedError:     "pciAddress is invalid",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens8",
			pci:               testPCI,
			altnames:          []string{},
			expectedError:     nodeNetConfPolAltnamesEmpty,
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens8",
			pci:               testPCI,
			altnames:          []string{"nic-pf"},
			expectedError:     nodeNetConfPolSelectorEmpty,
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithPCIAddressAltnames(
			testCase.interfaceName, testCase.pci, testCase.altnames)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		if testCase.expectedError != "" {
			continue
		}

		desireState := &DesiredState{}
		_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)

		wantAlts := make([]InterfaceAltName, len(testCase.altnames))
		for i, n := range testCase.altnames {
			wantAlts[i] = InterfaceAltName{Name: n}
		}

		assert.Equal(t, &DesiredState{
			Interfaces: []NetworkInterface{{
				Name:       testCase.interfaceName,
				Type:       interfaceTypeEthernet,
				State:      "up",
				PciAddress: testCase.pci,
				Identifier: "pci-address",
				AltNames:   wantAlts,
			}},
		}, desireState)
	}
}

func TestPolicyRemoveInterfaceAltname(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		interfaceName     string
		altnames          []string
		expectedError     string
		wantAltnames      []InterfaceAltName
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens1",
			altnames:          []string{"old-alt"},
			expectedError:     "",
			wantAltnames:      []InterfaceAltName{{Name: "old-alt", State: interfaceStateAbsent}},
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "",
			altnames:          []string{"x"},
			expectedError:     nodeNetConfPolIntError,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens1",
			altnames:          nil,
			expectedError:     "",
			wantAltnames:      nil,
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			interfaceName:     "ens1",
			altnames:          []string{"old-alt"},
			expectedError:     nodeNetConfPolSelectorEmpty,
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.RemoveInterfaceAltname(testCase.interfaceName, testCase.altnames)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		if testCase.expectedError != "" {
			continue
		}

		desireState := &DesiredState{}
		_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
		assert.Equal(t, &DesiredState{
			Interfaces: []NetworkInterface{{
				Name:     testCase.interfaceName,
				Type:     interfaceTypeEthernet,
				State:    "up",
				AltNames: testCase.wantAltnames,
			}},
		}, desireState)
	}
}

func TestPolicyWithEthernetInterfaceIP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		sriovInterface    string
		ipv4              string
		ipv6              string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolIntError,
			sriovInterface:    "",
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "ethernet interface 'ipv4Addresses' is an invalid ipv4 address",
			sriovInterface:    "ens1",
			ipv4:              "",
			ipv6:              "2001:db8::68",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "ethernet interface 'ipv6Addresses' is an invalid ipv6 address",
			sriovInterface:    "ens1",
			ipv4:              "10.10.10.10",
			ipv6:              "",
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolSelectorEmpty,
			sriovInterface:    "ens1",
			ipv4:              "10.10.10.10",
			ipv6:              "2001:db8::68",
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithEthernetInterface(testCase.sriovInterface,
			testCase.ipv4, testCase.ipv6)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  testCase.sriovInterface,
						Type:  interfaceTypeEthernet,
						State: "up",
						Ipv4: InterfaceIpv4{
							Enabled: true,
							Address: []InterfaceIPAddress{{
								PrefixLen: 24,
								IP:        net.ParseIP(testCase.ipv4),
							}},
						},
						Ipv6: InterfaceIpv6{Enabled: true,
							Address: []InterfaceIPAddress{{
								PrefixLen: 64,
								IP:        net.ParseIP(testCase.ipv6),
							}},
						},
					},
				},
			}, desireState)
		}
	}
}

func TestPolicyWithEthernetIPv6LinkLocalInterface(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		sriovInterface    string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			sriovInterface:    "ens1",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolIntError,
			sriovInterface:    "",
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolSelectorEmpty,
			sriovInterface:    "ens1",
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithEthernetIPv6LinkLocalInterface(testCase.sriovInterface)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  testCase.sriovInterface,
						Type:  interfaceTypeEthernet,
						State: "up",
						Ipv4: InterfaceIpv4{
							Enabled: false,
						},
						Ipv6: InterfaceIpv6{
							Enabled: true,
						},
					},
				},
			}, desireState)
		}
	}
}

func TestPolicyWithEthernetDualStackInterface(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		interfaceName     string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			interfaceName:     "ens1",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolIntError,
			interfaceName:     "",
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "nodeNetworkConfigurationPolicy 'nodeSelector' cannot be empty map",
			interfaceName:     "ens1",
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithEthernetDualStackInterface(testCase.interfaceName)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, &DesiredState{
				Interfaces: []NetworkInterface{
					{
						Name:  testCase.interfaceName,
						Type:  interfaceTypeEthernet,
						State: "up",
						Ipv4: InterfaceIpv4{
							Enabled: true,
						},
						Ipv6: InterfaceIpv6{
							Enabled: true,
						},
					},
				},
			}, desireState)
		}
	}
}

func TestPolicyWithAbsentInterface(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		baseInterface     string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			baseInterface:     "ens1",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolIntError,
			baseInterface:     "",
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolSelectorEmpty,
			baseInterface:     "ens1",
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithAbsentInterface(testCase.baseInterface)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, desireState, &DesiredState{
				Interfaces: []NetworkInterface{{Name: testCase.baseInterface, State: interfaceStateAbsent}}})
		}
	}
}

func TestPolicyWithInterfaceUp(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		baseInterface     string
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			baseInterface:     "ens1",
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolIntError,
			baseInterface:     "",
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolSelectorEmpty,
			baseInterface:     "ens1",
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithInterfaceUp(testCase.baseInterface)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		desireState := &DesiredState{}
		if testCase.expectedError == "" {
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.Equal(t, desireState, &DesiredState{
				Interfaces: []NetworkInterface{{Name: testCase.baseInterface, Type: interfaceTypeEthernet, State: "up"}}})
		}
	}
}

func TestPolicyWithStaticRoute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     string
		destination       string
		nextHopAddress    string
		nextHopInterface  string
		metric            int
		tableID           int
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			destination:       "192.168.1.0/24",
			nextHopAddress:    "10.10.10.1",
			nextHopInterface:  "ens1",
			metric:            100,
			tableID:           254,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			destination:       "192.168.2.0/24",
			nextHopAddress:    "10.10.10.1",
			nextHopInterface:  "",
			metric:            0,
			tableID:           0,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "",
			destination:       "192.168.3.0/24",
			nextHopAddress:    "",
			nextHopInterface:  "ens1",
			metric:            50,
			tableID:           0,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "route 'destination' cannot be empty",
			destination:       "",
			nextHopAddress:    "10.10.10.1",
			nextHopInterface:  "ens1",
			metric:            100,
			tableID:           254,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     "route must have either 'nextHopAddress' or 'nextHopInterface'",
			destination:       "192.168.1.0/24",
			nextHopAddress:    "",
			nextHopInterface:  "",
			metric:            100,
			tableID:           254,
		},
		{
			testNMStatePolicy: buildInValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nodeNetConfPolSelectorEmpty,
			destination:       "192.168.1.0/24",
			nextHopAddress:    "10.10.10.1",
			nextHopInterface:  "ens1",
			metric:            100,
			tableID:           254,
		},
	}
	for _, testCase := range testCases {
		testPolicy := testCase.testNMStatePolicy.WithStaticRoute(
			testCase.destination, testCase.nextHopAddress, testCase.nextHopInterface,
			testCase.metric, testCase.tableID)
		assertPolicyError(t, testCase.expectedError, testPolicy)

		if testCase.expectedError == "" {
			desireState := &DesiredState{}
			_ = yaml.Unmarshal(testPolicy.Definition.Spec.DesiredState.Raw, desireState)
			assert.NotNil(t, desireState.Routes)
			assert.Len(t, desireState.Routes.Config, 1)
			assert.Equal(t, RouteConfig{
				Destination:      testCase.destination,
				NextHopAddress:   testCase.nextHopAddress,
				NextHopInterface: testCase.nextHopInterface,
				Metric:           testCase.metric,
				TableID:          testCase.tableID,
			}, desireState.Routes.Config[0])
		}
	}
}

func TestPolicyWaitUntilCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testNMStatePolicy *PolicyBuilder
		expectedError     error
		condition         shared.ConditionType
	}{
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     nil,
			condition:         shared.NodeNetworkStateConditionAvailable,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject()),
			expectedError:     context.DeadlineExceeded,
			condition:         shared.NodeNetworkConfigurationEnactmentConditionFailing,
		},
		{
			testNMStatePolicy: buildValidPolicyTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     fmt.Errorf("cannot wait for NodeNetworkConfigurationPolicy condition because it does not exist"),
			condition:         shared.NodeNetworkStateConditionAvailable,
		},
	}
	for _, testCase := range testCases {
		err := testCase.testNMStatePolicy.WaitUntilCondition(testCase.condition, 2*time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

// assertPolicyError verifies the expected package-specific builder error, if any.
func assertPolicyError(t *testing.T, expectedError string, builder *PolicyBuilder) {
	t.Helper()

	if expectedError == "" {
		assert.NoError(t, builder.GetError(), "expected the Policy builder to remain valid")

		return
	}

	assert.EqualError(t, builder.GetError(), expectedError, "unexpected Policy builder error")
}

// newPolicyCommonTestConfig returns the common test configuration for NodeNetworkConfigurationPolicy.
func newPolicyCommonTestConfig() testhelper.CommonTestConfig[
	nmstatev1.NodeNetworkConfigurationPolicy,
	PolicyBuilder,
	*nmstatev1.NodeNetworkConfigurationPolicy,
	*PolicyBuilder,
] {
	return testhelper.NewCommonTestConfig[nmstatev1.NodeNetworkConfigurationPolicy, PolicyBuilder](
		nmstatev1.AddToScheme,
		policyGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

// buildValidPolicyTestBuilder returns a valid PolicyBuilder for testing purposes.
func buildValidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(apiClient, defaultPolicyName, map[string]string{"test": "test"})
}

// buildInValidPolicyTestBuilder returns an invalid PolicyBuilder for testing purposes.
func buildInValidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(apiClient, "test", map[string]string{})
}

func buildTestClientWithDummyPolicyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPolicyObject(),
		SchemeAttachers: v1TestSchemes,
	})
}

func buildDummyPolicyObject() []runtime.Object {
	return append([]runtime.Object{}, &nmstatev1.NodeNetworkConfigurationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultPolicyName,
		},
		Spec: shared.NodeNetworkConfigurationPolicySpec{},
		Status: shared.NodeNetworkConfigurationPolicyStatus{
			Conditions: shared.ConditionList{
				shared.Condition{
					Type:   shared.NodeNetworkStateConditionAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		},
	})
}
