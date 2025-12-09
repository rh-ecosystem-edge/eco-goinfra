package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultBGPSessionStateName = "bgpsessionstate-0"
)

func TestBGPSessionStateGet(t *testing.T) {
	var runtimeObjects []runtime.Object

	testCases := []struct {
		testBGPSessionState *BGPSessionStateBuilder
		addToRuntimeObjects bool
		expectedError       string
		client              bool
	}{
		{
			testBGPSessionState: buildValidBGPSessionStateTestBuilder(
				buildTestBGPSessionStateClientWithDummyState(defaultBGPSessionStateName)),
			expectedError: "",
		},
		{
			testBGPSessionState: buildValidBGPSessionStateTestBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: frrTestSchemes,
			})),
			expectedError: "bgpsessionstates.frrk8s.metallb.io \"bgpsessionstate-0\" not found",
		},
	}

	for _, testCase := range testCases {
		bgpSessionState, err := testCase.testBGPSessionState.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, bgpSessionState.Name, testCase.testBGPSessionState.Definition.Name, bgpSessionState.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestBGPSessionStateExist(t *testing.T) {
	testCases := []struct {
		testBGPSessionState *BGPSessionStateBuilder
		exist               bool
	}{
		{
			testBGPSessionState: buildValidBGPSessionStateTestBuilder(
				buildTestBGPSessionStateClientWithDummyState("test-state")),
			exist: false,
		},
		{
			testBGPSessionState: buildValidBGPSessionStateTestBuilder(
				buildTestBGPSessionStateClientWithDummyState(defaultBGPSessionStateName)),
			exist: true,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testBGPSessionState.Exists()
		assert.Equal(t, testCase.exist, exist)
	}
}

func TestPullBGPSessionState(t *testing.T) {
	generateBGPSessionState := func(name string) *frrtypes.BGPSessionState {
		return &frrtypes.BGPSessionState{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Status: frrtypes.BGPSessionStateStatus{},
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
			expectedErrorText:   "BGPSessionState 'name' cannot be empty",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test1",
			expectedError:       true,
			expectedErrorText:   "BGPSessionState object test1 does not exist",
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

		testBGPSessionState := generateBGPSessionState(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testBGPSessionState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: frrTestSchemes,
			})
		}

		builderResult, err := PullBGPSessionState(testSettings, testCase.name)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testBGPSessionState.Name, builderResult.Object.Name)
		}
	}
}

func TestPullBGPSessionStateByNodeAndPeer(t *testing.T) {
	generateBGPSessionState := func(name, nodeName, peerIP string) *frrtypes.BGPSessionState {
		return &frrtypes.BGPSessionState{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Status: frrtypes.BGPSessionStateStatus{
				Node: nodeName,
				Peer: peerIP,
			},
		}
	}

	testCases := getPullByNodeAndPeerTestCases()

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testBGPSessionState := generateBGPSessionState(
			testCase.name, testCase.nodeName, testCase.peerIP)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testBGPSessionState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: frrTestSchemes,
			})
		}

		builderResult, err := PullBGPSessionStateByNodeAndPeer(
			testSettings, testCase.nodeName, testCase.peerIP)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testBGPSessionState.Name, builderResult.Object.Name)
			assert.Equal(t, testCase.nodeName, builderResult.Object.Status.Node)
			assert.Equal(t, testCase.peerIP, builderResult.Object.Status.Peer)
		}
	}
}

// buildDummyBGPSessionState returns a BGPSessionState with the provided name.
func buildDummyBGPSessionState(name string) *frrtypes.BGPSessionState {
	return &frrtypes.BGPSessionState{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestBGPSessionStateClientWithDummyState returns a client with a dummy BGPSessionState.
func buildTestBGPSessionStateClientWithDummyState(stateName string) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyBGPSessionState(stateName)},
		SchemeAttachers: frrTestSchemes,
	})
}

func buildValidBGPSessionStateTestBuilder(apiClient *clients.Settings) *BGPSessionStateBuilder {
	return newBGPSessionStateBuilder(apiClient, defaultBGPSessionStateName)
}

func newBGPSessionStateBuilder(apiClient *clients.Settings, name string) *BGPSessionStateBuilder {
	if apiClient == nil {
		return nil
	}

	builder := BGPSessionStateBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummyBGPSessionState(name),
	}

	return &builder
}
func getPullByNodeAndPeerTestCases() []struct {
	name                string
	nodeName            string
	peerIP              string
	expectedError       bool
	addToRuntimeObjects bool
	expectedErrorText   string
	client              bool
} {
	return []struct {
		name                string
		nodeName            string
		peerIP              string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			name:                defaultBGPSessionStateName,
			nodeName:            defaultNodeName,
			peerIP:              "192.168.1.1",
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                defaultBGPSessionStateName,
			nodeName:            "",
			peerIP:              "192.168.1.1",
			expectedError:       true,
			expectedErrorText:   "node name cannot be empty",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                defaultBGPSessionStateName,
			nodeName:            defaultNodeName,
			peerIP:              "",
			expectedError:       true,
			expectedErrorText:   "peer IP cannot be empty",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                defaultBGPSessionStateName,
			nodeName:            defaultNodeName,
			peerIP:              "192.168.1.1",
			expectedError:       true,
			expectedErrorText:   "BGPSessionState for node worker-0 and peer 192.168.1.1 not found",
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                defaultBGPSessionStateName,
			nodeName:            defaultNodeName,
			peerIP:              "192.168.1.1",
			expectedError:       true,
			expectedErrorText:   "the apiClient cannot be nil",
			addToRuntimeObjects: true,
			client:              false,
		},
	}
}
