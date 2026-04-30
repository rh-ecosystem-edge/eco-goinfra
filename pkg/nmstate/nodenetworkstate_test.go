package nmstate

import (
	"fmt"
	"testing"

	"github.com/nmstate/kubernetes-nmstate/api/shared"
	nmstateV1beta1 "github.com/nmstate/kubernetes-nmstate/api/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	nmstateV1beta1TestSchemes = []clients.SchemeAttacher{
		nmstateV1beta1.AddToScheme,
	}
	sriovExistingInterface = "ensf0"
	sriovTestMACAddress    = "52:54:00:12:34:56"
	sriovTestPCIAddress    = "0000:00:03.0"
	sriovTestAltnames      = []string{"eth-alt1", "eth-alt2"}
	vfNumber               = 10
)

func TestPullNodeNetworkState(t *testing.T) {
	generateNodeNetState := func(name string) *nmstateV1beta1.NodeNetworkState {
		return &nmstateV1beta1.NodeNetworkState{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Status: shared.NodeNetworkStateStatus{},
		}
	}

	testCases := []struct {
		name                string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			name:                "test1",
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			expectedError:       true,
			expectedErrorText:   "nodeNetworkState 'name' cannot be empty",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test1",
			expectedError:       true,
			expectedErrorText:   "nodeNetworkState object test1 does not exist",
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "test1",
			expectedError:       true,
			expectedErrorText:   "the apiClient cannot be nil",
			addToRuntimeObjects: true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testNodeNetState := generateNodeNetState(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNodeNetState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: nmstateV1beta1TestSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := PullNodeNetworkState(testSettings, testCase.name)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testNodeNetState.Name, builderResult.Object.Name)
			assert.Equal(t, testNodeNetState.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestStateBuilderGet(t *testing.T) {
	testCases := []struct {
		nodeNetStateBuilder *StateBuilder
		expectedError       error
	}{
		{
			nodeNetStateBuilder: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			expectedError:       nil,
		},
		{
			nodeNetStateBuilder: buildValidNodeNetworkStateTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:       fmt.Errorf("nodenetworkstates.nmstate.io \"nmstatename\" not found"),
		},
	}

	for _, testCase := range testCases {
		nodeNetState, err := testCase.nodeNetStateBuilder.Get()
		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Equal(t, nodeNetState.Name, testCase.nodeNetStateBuilder.Object.Name)
		}
	}
}

func TestStateBuilderExist(t *testing.T) {
	testCases := []struct {
		testNodeNetState *StateBuilder
		expectedStatus   bool
	}{
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			expectedStatus:   true,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:   false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testNodeNetState.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

func TestStateBuilderGetTotalVFs(t *testing.T) {
	testCases := []struct {
		testNodeNetState   *StateBuilder
		sriovInterfaceName string
		vfsNumber          int
		expectedError      error
	}{
		{
			testNodeNetState:   buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			sriovInterfaceName: sriovExistingInterface,
			vfsNumber:          vfNumber,
			expectedError:      nil,
		},
		{
			testNodeNetState:   buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			sriovInterfaceName: "invalidname",
			vfsNumber:          0,
			expectedError:      fmt.Errorf("failed to find interface invalidname"),
		},
		{
			testNodeNetState:   buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			sriovInterfaceName: "",
			vfsNumber:          0,
			expectedError:      fmt.Errorf("the sriovInterfaceName is empty sting"),
		},
	}

	for _, testCase := range testCases {
		vfsNumber, err := testCase.testNodeNetState.GetTotalVFs(testCase.sriovInterfaceName)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, vfsNumber, testCase.vfsNumber)
	}
}

func TestStateBuilderGetInterfaceType(t *testing.T) {
	testCases := []struct {
		testNodeNetState *StateBuilder
		interfaceName    string
		interfaceType    string
		expectedError    error
	}{
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    sriovExistingInterface,
			interfaceType:    "ethernet",
			expectedError:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "",
			interfaceType:    "ethernet",
			expectedError:    fmt.Errorf("the interfaceName is empty sting"),
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    sriovExistingInterface,
			interfaceType:    "",
			expectedError:    fmt.Errorf("invalid interfaceType parameter"),
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "test",
			interfaceType:    "ethernet",
			expectedError:    fmt.Errorf("failed to find interface test or it is not a ethernet type"),
		},
	}

	for _, testCase := range testCases {
		networkInterface, err := testCase.testNodeNetState.GetInterfaceType(testCase.interfaceName, testCase.interfaceType)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, networkInterface.Name, testCase.interfaceName)
			assert.Equal(t, networkInterface.Type, testCase.interfaceType)
		}
	}
}

func TestStateBuilderGetInterfaceMACAddress(t *testing.T) {
	testCases := []struct {
		testNodeNetState *StateBuilder
		interfaceName    string
		expectedMAC      string
		expectedError    error
		mutateBuilder    func(*StateBuilder)
	}{
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    sriovExistingInterface,
			expectedMAC:      sriovTestMACAddress,
			expectedError:    nil,
			mutateBuilder:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "unknown",
			expectedMAC:      "",
			expectedError:    fmt.Errorf("failed to find interface unknown"),
			mutateBuilder:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "",
			expectedMAC:      "",
			expectedError:    fmt.Errorf("the interfaceName is empty sting"),
			mutateBuilder:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    sriovExistingInterface,
			expectedMAC:      "",
			expectedError:    fmt.Errorf("failed to find mac-address for interface %s", sriovExistingInterface),
			mutateBuilder: func(builder *StateBuilder) {
				currentState := DesiredState{}
				_ = yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &currentState)
				currentState.Interfaces[0].MacAddress = ""
				builder.Object.Status.CurrentState.Raw, _ = yaml.Marshal(currentState)
			},
		},
	}

	for _, testCase := range testCases {
		if testCase.mutateBuilder != nil {
			testCase.mutateBuilder(testCase.testNodeNetState)
		}

		mac, err := testCase.testNodeNetState.GetInterfaceMACAddress(testCase.interfaceName)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedMAC, mac)
	}
}

func TestStateBuilderGetInterfacePCIAddress(t *testing.T) {
	testCases := []struct {
		testNodeNetState *StateBuilder
		interfaceName    string
		expectedPCI      string
		expectedError    error
		mutateBuilder    func(*StateBuilder)
	}{
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    sriovExistingInterface,
			expectedPCI:      sriovTestPCIAddress,
			expectedError:    nil,
			mutateBuilder:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "unknown",
			expectedPCI:      "",
			expectedError:    fmt.Errorf("failed to find interface unknown"),
			mutateBuilder:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "",
			expectedPCI:      "",
			expectedError:    fmt.Errorf("the interfaceName is empty sting"),
			mutateBuilder:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    sriovExistingInterface,
			expectedPCI:      "",
			expectedError:    fmt.Errorf("failed to find pci-address for interface %s", sriovExistingInterface),
			mutateBuilder: func(builder *StateBuilder) {
				currentState := DesiredState{}
				_ = yaml.Unmarshal(builder.Object.Status.CurrentState.Raw, &currentState)
				currentState.Interfaces[0].PciAddress = ""
				builder.Object.Status.CurrentState.Raw, _ = yaml.Marshal(currentState)
			},
		},
	}

	for _, testCase := range testCases {
		if testCase.mutateBuilder != nil {
			testCase.mutateBuilder(testCase.testNodeNetState)
		}

		pci, err := testCase.testNodeNetState.GetInterfacePCIAddress(testCase.interfaceName)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedPCI, pci)
	}
}

func TestStateBuilderGetInterfaceAltnames(t *testing.T) {
	testCases := []struct {
		testNodeNetState *StateBuilder
		interfaceName    string
		expectedAltnames []string
		expectedError    error
	}{
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    sriovExistingInterface,
			expectedAltnames: sriovTestAltnames,
			expectedError:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "unknown",
			expectedAltnames: nil,
			expectedError:    fmt.Errorf("failed to find interface unknown"),
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "",
			expectedAltnames: nil,
			expectedError:    fmt.Errorf("the interfaceName is empty sting"),
		},
	}

	for _, testCase := range testCases {
		names, err := testCase.testNodeNetState.GetInterfaceAltnames(testCase.interfaceName)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedAltnames, names)
	}
}

func TestStateBuilderGetSriovVfs(t *testing.T) {
	testCases := []struct {
		testNodeNetState *StateBuilder
		interfaceName    string
		expectedError    error
		vfsInUse         int
	}{
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    sriovExistingInterface,
			vfsInUse:         1,
			expectedError:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "",
			expectedError:    fmt.Errorf("the sriovInterfaceName is empty sting"),
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "test",
			expectedError:    fmt.Errorf("failed to find interface test or SR-IOV VFs are not configured on it"),
		},
	}

	for _, testCase := range testCases {
		vfsList, err := testCase.testNodeNetState.GetSriovVfs(testCase.interfaceName)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.vfsInUse, len(vfsList))
		}
	}
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidNodeNetworkStateTestBuilder(apiClient *clients.Settings) *StateBuilder {
	return newNodeNetworkStateBuilder(apiClient, defaultNMStateName)
}

func buildTestClientWithDummyNodeNetworkStateObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyNodeNetworkStateObject(),
		SchemeAttachers: nmstateV1beta1TestSchemes,
	})
}

func buildDummyNodeNetworkStateObject() []runtime.Object {
	return append([]runtime.Object{}, &nmstateV1beta1.NodeNetworkState{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultNMStateName,
		},
	})
}

// newNodeNetworkStateBuilder creates a new instance of NodeNetworkStateBuilder Builder.
func newNodeNetworkStateBuilder(apiClient *clients.Settings, name string) *StateBuilder {
	desiredState := DesiredState{
		Interfaces: []NetworkInterface{
			{
				Name:       sriovExistingInterface,
				Type:       "ethernet",
				State:      "up",
				MacAddress: sriovTestMACAddress,
				PciAddress: sriovTestPCIAddress,
				AltNames: []InterfaceAltName{
					{Name: sriovTestAltnames[0]},
					{Name: sriovTestAltnames[1]},
				},
				Ethernet: Ethernet{
					Sriov: Sriov{
						TotalVfs: &vfNumber,
						Vfs: []Vf{{
							ID: 123,
						},
						},
					},
				},
			},
		},
	}
	byteDesiredState, _ := yaml.Marshal(desiredState)

	err := apiClient.AttachScheme(nmstateV1beta1.AddToScheme)
	if err != nil {
		return nil
	}

	builder := StateBuilder{
		apiClient: apiClient.Client,
		Object: &nmstateV1beta1.NodeNetworkState{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},

			Status: shared.NodeNetworkStateStatus{
				CurrentState: shared.State{
					Raw: byteDesiredState,
				},
			},
		},
	}

	return &builder
}
