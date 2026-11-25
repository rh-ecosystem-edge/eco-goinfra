package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultBGPSessionStateName     = "bgpsessionstate-0"
	defaultBGPSessionStateNodeName = "worker-0"
	defaultBGPSessionStatePeerIP   = "192.168.1.1"
	bgpSessionStateGVK             = frrtypes.GroupVersion.WithKind("BGPSessionState")
)

func TestPullBGPSessionState(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullBGPSessionState,
		frrtypes.AddToScheme,
		bgpSessionStateGVK,
	).ExecuteTests(t)
}

func TestBGPSessionStateMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newBGPSessionStateCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}

func TestPullBGPSessionStateByNodeAndPeer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		nodeName      string
		peerIP        string
		objects       []runtime.Object
		client        bool
		expectedError string
	}{
		{
			name:     "finds matching node and peer",
			nodeName: defaultBGPSessionStateNodeName,
			peerIP:   defaultBGPSessionStatePeerIP,
			objects: []runtime.Object{buildDummyBGPSessionState(
				defaultBGPSessionStateName, defaultBGPSessionStateNodeName, defaultBGPSessionStatePeerIP)},
			client: true,
		},
		{
			name:     "rejects empty node name",
			nodeName: "",
			peerIP:   defaultBGPSessionStatePeerIP,
			objects: []runtime.Object{buildDummyBGPSessionState(
				defaultBGPSessionStateName, defaultBGPSessionStateNodeName, defaultBGPSessionStatePeerIP)},
			client:        true,
			expectedError: "node name cannot be empty",
		},
		{
			name:     "rejects empty peer IP",
			nodeName: defaultBGPSessionStateNodeName,
			peerIP:   "",
			objects: []runtime.Object{buildDummyBGPSessionState(
				defaultBGPSessionStateName, defaultBGPSessionStateNodeName, defaultBGPSessionStatePeerIP)},
			client:        true,
			expectedError: "peer IP cannot be empty",
		},
		{
			name:     "returns error when no session matches",
			nodeName: defaultBGPSessionStateNodeName,
			peerIP:   defaultBGPSessionStatePeerIP,
			objects: []runtime.Object{buildDummyBGPSessionState(
				defaultBGPSessionStateName, "other-node", defaultBGPSessionStatePeerIP)},
			client:        true,
			expectedError: "BGPSessionState for node worker-0 and peer 192.168.1.1 not found",
		},
		{
			name:          "rejects nil client",
			nodeName:      defaultBGPSessionStateNodeName,
			peerIP:        defaultBGPSessionStatePeerIP,
			expectedError: "the apiClient cannot be nil",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings
			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  testCase.objects,
					SchemeAttachers: frrTestSchemes,
				})
			}

			builder, err := PullBGPSessionStateByNodeAndPeer(testSettings, testCase.nodeName, testCase.peerIP)

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, builder)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, builder)
			require.NotNil(t, builder.Object)
			assert.Equal(t, defaultBGPSessionStateName, builder.Object.Name)
			assert.Equal(t, testCase.nodeName, builder.Object.Status.Node)
			assert.Equal(t, testCase.peerIP, builder.Object.Status.Peer)
		})
	}
}

func newBGPSessionStateCommonTestConfig() testhelper.CommonTestConfig[
	frrtypes.BGPSessionState, BGPSessionStateBuilder, *frrtypes.BGPSessionState, *BGPSessionStateBuilder] {
	return testhelper.NewCommonTestConfig[frrtypes.BGPSessionState, BGPSessionStateBuilder](
		frrtypes.AddToScheme,
		bgpSessionStateGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

// buildDummyBGPSessionState returns a BGPSessionState with the provided identity.
func buildDummyBGPSessionState(name, nodeName, peerIP string) *frrtypes.BGPSessionState {
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

// buildTestBGPSessionStateClientWithDummyState returns a client with a dummy BGPSessionState.
func buildTestBGPSessionStateClientWithDummyState(stateName string) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{buildDummyBGPSessionState(
			stateName, defaultBGPSessionStateNodeName, defaultBGPSessionStatePeerIP)},
		SchemeAttachers: frrTestSchemes,
	})
}
