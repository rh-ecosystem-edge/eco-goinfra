package nmstate

import (
	"fmt"
	"testing"

	"github.com/nmstate/kubernetes-nmstate/api/shared"
	nmstateV1beta1 "github.com/nmstate/kubernetes-nmstate/api/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
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
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullNodeNetworkState, nmstateV1beta1.AddToScheme, nodeNetworkStateGVK).ExecuteTests(t)
}

func TestStateBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := testhelper.NewCommonTestConfig[
		nmstateV1beta1.NodeNetworkState, StateBuilder](
		nmstateV1beta1.AddToScheme, nodeNetworkStateGVK, testhelper.ResourceScopeClusterScoped)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		Run(t)
}

func TestStateBuilderGetTotalVFs(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	testCases := []struct {
		testNodeNetState *StateBuilder
		interfaceName    string
		interfaceType    string
		expectedError    error
	}{
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    sriovExistingInterface,
			interfaceType:    interfaceTypeEthernet,
			expectedError:    nil,
		},
		{
			testNodeNetState: buildValidNodeNetworkStateTestBuilder(buildTestClientWithDummyNodeNetworkStateObject()),
			interfaceName:    "",
			interfaceType:    interfaceTypeEthernet,
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
			interfaceType:    interfaceTypeEthernet,
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
	t.Parallel()

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
			interfaceName:    unknown,
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
	t.Parallel()

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
			interfaceName:    unknown,
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
	t.Parallel()

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
			interfaceName:    unknown,
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
	t.Parallel()

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
				Type:       interfaceTypeEthernet,
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

	builder := common.NewClusterScopedBuilder[nmstateV1beta1.NodeNetworkState, StateBuilder](
		apiClient, nmstateV1beta1.AddToScheme, name)
	if builder.GetError() != nil {
		return builder
	}

	builder.Object = &nmstateV1beta1.NodeNetworkState{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: shared.NodeNetworkStateStatus{
			CurrentState: shared.State{
				Raw: byteDesiredState,
			},
		},
	}

	return builder
}
