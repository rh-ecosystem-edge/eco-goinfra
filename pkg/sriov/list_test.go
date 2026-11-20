package sriov

import (
	"fmt"
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestNetworkList(t *testing.T) {
	testCases := []struct {
		testNetwork   []*NetworkBuilder
		nsName        string
		listOptions   []client.ListOptions
		expectedError error
		client        bool
	}{
		{
			testNetwork:   []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        "testnamespace",
			expectedError: nil,
			client:        true,
		},
		{
			testNetwork:   []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        "",
			expectedError: fmt.Errorf("failed to list sriov networks, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			testNetwork: []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:      "testnamespace",
			listOptions: []client.ListOptions{{Continue: "test"}},
			client:      true,
		},
		{
			testNetwork:   []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        "testnamespace",
			listOptions:   []client.ListOptions{{Namespace: "test"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testNetwork:   []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			nsName:        "testnamespace",
			expectedError: fmt.Errorf("failed to list sriov networks, 'apiClient' parameter is empty"),
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummySrIovNetworkObject(),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := List(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.testNetwork))
		}
	}
}

func TestNetworkCleanAllNetworksByTargetNamespace(t *testing.T) {
	testCases := []struct {
		testNetwork    []*NetworkBuilder
		operatorNsName string
		targetNsName   string
		listOptions    []client.ListOptions
		client         bool
		expectedError  error
	}{
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: "testnamespace",
			targetNsName:   "targetns",
			client:         true,
		},
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			targetNsName:   "testnamespace",
			expectedError:  fmt.Errorf("failed to clean up sriov networks, 'operatornsname' parameter is empty"),
			client:         true,
			operatorNsName: "",
		},
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: "targetns",
			expectedError:  fmt.Errorf("failed to clean up sriov networks, 'targetnsname' parameter is empty"),
			client:         true,
			targetNsName:   "",
		},
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: "testnamespace",
			targetNsName:   "targetns",
			listOptions:    []client.ListOptions{{Namespace: "test"}},
			client:         true,
		},
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: "testnamespace",
			targetNsName:   "targetns",
			listOptions:    []client.ListOptions{{Namespace: "test"}, {Continue: "true"}},
			expectedError:  fmt.Errorf("error: more than one ListOptions was passed"),
			client:         true,
		},
		{
			testNetwork:    []*NetworkBuilder{buildValidSriovNetworkTestBuilder(buildTestClientWithDummyObject())},
			operatorNsName: "testnamespace",
			targetNsName:   "targetns",
			expectedError:  fmt.Errorf("failed to list sriov networks, 'apiClient' parameter is empty"),
			client:         false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: buildDummySrIovNetworkObject(),
			})
		}

		err := CleanAllNetworksByTargetNamespace(
			testSettings, testCase.operatorNsName, testCase.targetNsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			netBuilders, err := List(testSettings, testCase.operatorNsName)
			assert.Nil(t, err)
			assert.Zero(t, len(netBuilders))
		}
	}
}

func TestListNetworkNodeState(t *testing.T) {
	testCases := []struct {
		testNetworkNodeStates []*srIovV1.SriovNetworkNodeState
		listOptions           []client.ListOptions
		nsName                string
		client                bool
		expectedError         error
	}{
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", "testnamespace"),
				buildNodeNetworkState("test2", "testnamespace")},
			nsName: "testnamespace",
			client: true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", "testnamespace"),
				buildNodeNetworkState("test2", "testnamespace")},
			expectedError: fmt.Errorf("failed to list SriovNetworkNodeStates, 'nsname' parameter is empty"),
			client:        true,
			nsName:        "",
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", "testnamespace"),
				buildNodeNetworkState("test2", "testnamespace")},
			nsName:      "testnamespace",
			listOptions: []client.ListOptions{{Continue: "true"}},
			client:      true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", "testnamespace"),
				buildNodeNetworkState("test2", "testnamespace")},
			nsName:        "testnamespace",
			listOptions:   []client.ListOptions{{Namespace: "testnamespace"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodeState{buildNodeNetworkState("test", "testnamespace"),
				buildNodeNetworkState("test2", "testnamespace")},
			nsName:        "testnamespace",
			expectedError: fmt.Errorf("failed to list SriovNetworkNodeStates, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		for _, networkNodeState := range testCase.testNetworkNodeStates {
			runtimeObjects = append(runtimeObjects, networkNodeState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builders, err := ListNetworkNodeState(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(testCase.testNetworkNodeStates), len(builders))
		}
	}
}

func TestListPolicy(t *testing.T) {
	testCases := []struct {
		testNetworkNodeStates []*srIovV1.SriovNetworkNodePolicy
		listOptions           []client.ListOptions
		nsName                string
		client                bool
		expectedError         error
	}{
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			nsName: "testnamespace",
			client: true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			expectedError: fmt.Errorf("failed to list SriovNetworkNodePolicies, 'nsname' parameter is empty"),
			client:        true,
			nsName:        "",
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			nsName:      "testnamespace",
			listOptions: []client.ListOptions{{Continue: "false"}},
			client:      true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			nsName:        "testnamespace",
			listOptions:   []client.ListOptions{{Namespace: "testnamespace"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testNetworkNodeStates: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			nsName:        "testnamespace",
			expectedError: fmt.Errorf("failed to list SriovNetworkNodePolicies, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		for _, networkNodeState := range testCase.testNetworkNodeStates {
			runtimeObjects = append(runtimeObjects, networkNodeState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builders, err := ListPolicy(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, len(testCase.testNetworkNodeStates), len(builders))
		}
	}
}

func TestCleanAllNetworkNodePolicies(t *testing.T) {
	testCases := []struct {
		testNetworkPolicy []*srIovV1.SriovNetworkNodePolicy
		listOptions       []client.ListOptions
		nsName            string
		client            bool
		expectedError     error
	}{
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			nsName: "testnamespace",
			client: true,
		},
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			nsName:        "",
			expectedError: fmt.Errorf("failed to clean up SriovNetworkNodePolicies, 'operatornsname' parameter is empty"),
			client:        true,
		},
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			nsName:      "testnamespace",
			listOptions: []client.ListOptions{{Continue: "false"}},
			client:      true,
		},
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			nsName:        "testnamespace",
			listOptions:   []client.ListOptions{{Namespace: "testnamespace"}, {Continue: "true"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			testNetworkPolicy: []*srIovV1.SriovNetworkNodePolicy{
				buildDummySrIovPolicy("test", "testnamespace"),
				buildDummySrIovPolicy("test1", "testnamespace")},
			nsName:        "testnamespace",
			expectedError: fmt.Errorf("failed to list SriovNetworkNodePolicies, 'apiClient' parameter is empty"),
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		for _, networkNodeState := range testCase.testNetworkPolicy {
			runtimeObjects = append(runtimeObjects, networkNodeState)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		err := CleanAllNetworkNodePolicies(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			netBuilders, err := ListPolicy(testSettings, testCase.nsName)
			assert.Nil(t, err)
			assert.Zero(t, len(netBuilders))
		}
	}
}

func TestListPoolConfigs(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, namespace string, _ ...client.ListOptions) ([]*PoolConfigBuilder, error) {
			return ListPoolConfigs(apiClient, namespace)
		},
		srIovV1.AddToScheme,
		poolConfigGVK,
	).ExecuteTests(t)
}

func TestCleanPoolConfigs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		operatorNamespace string
		expectedError     string
		errorMatches      func(error) bool
		client            bool
	}{
		{
			name:              "cleans resources",
			operatorNamespace: defaultNetNsName,
			client:            true,
		},
		{
			name:              "nil client returns error",
			operatorNamespace: defaultNetNsName,
			client:            false,
			errorMatches:      commonerrors.IsAPIClientNil,
		},
		{
			name:              "empty operator namespace returns error",
			operatorNamespace: "",
			expectedError:     "failed to clean up SriovNetworkPoolConfigs, 'operatornsname' parameter is empty",
			client:            true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings
			if testCase.client {
				testSettings = newPoolConfigTestClient(buildDummyPoolConfigObject()...)
			}

			err := CleanAllPoolConfigs(testSettings, testCase.operatorNamespace)

			if testCase.errorMatches != nil {
				require.Truef(t, testCase.errorMatches(err), "unexpected error, got: %v", err)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)

				return
			}

			require.NoError(t, err)

			poolConfigBuilders, err := ListPoolConfigs(testSettings, testCase.operatorNamespace)
			require.NoError(t, err)
			assert.Empty(t, poolConfigBuilders)
		})
	}
}
