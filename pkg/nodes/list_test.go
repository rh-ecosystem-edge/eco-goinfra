package nodes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNodesList(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(List, corev1.AddToScheme, nodeGVK).ExecuteTests(t)
}

func TestNodesListWithOptions(t *testing.T) {
	t.Parallel()

	workerNode := buildDummyNode("worker-node")
	workerNode.Labels = map[string]string{"role": "worker"}
	controlPlaneNode := buildDummyNode("control-plane-node")
	controlPlaneNode.Labels = map[string]string{"role": "control-plane"}

	testCases := []struct {
		name          string
		options       []metav1.ListOptions
		expectedCount int
		expectedError string
	}{
		{
			name:          "applies label selector",
			options:       []metav1.ListOptions{{LabelSelector: "role=worker"}},
			expectedCount: 1,
		},
		{
			name:          "more than one option returns error",
			options:       []metav1.ListOptions{{}, {}},
			expectedError: "error: more than one ListOptions was passed",
		},
		{
			name:          "invalid selector returns error",
			options:       []metav1.ListOptions{{LabelSelector: "role in ("}},
			expectedError: "failed to parse label selector",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{workerNode, controlPlaneNode},
				SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
			})

			nodeBuilders, err := List(testSettings, testCase.options...)
			if testCase.expectedError != "" {
				require.ErrorContains(t, err, testCase.expectedError)

				return
			}

			require.NoError(t, err)
			assert.Len(t, nodeBuilders, testCase.expectedCount)
		})
	}
}

func TestNodesListExternalIPv4Networks(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		client      bool
		assertError func(error) bool
	}{
		{
			name:        "returns external IPv4 networks",
			client:      true,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "nil client returns error",
			assertError: commonerrors.IsAPIClientNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				node := buildDummyNode(defaultNodeName)
				node.Annotations = map[string]string{ovnExternalAddresses: defaultExternalNetworks}
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  []runtime.Object{node},
					SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
				})
			}

			ipv4Networks, err := ListExternalIPv4Networks(testSettings)
			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)

			if err == nil {
				assert.Equal(t, []string{defaultExternalIPv4}, ipv4Networks)
			}
		})
	}
}

func TestNodesListExternalIPv6Networks(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		client      bool
		assertError func(error) bool
	}{
		{
			name:        "returns external IPv6 networks",
			client:      true,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "nil client returns error",
			assertError: commonerrors.IsAPIClientNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				node := buildDummyNode(defaultNodeName)
				node.Annotations = map[string]string{ovnExternalAddresses: defaultExternalNetworks}
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  []runtime.Object{node},
					SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
				})
			}

			ipv6Networks, err := ListExternalIPv6Networks(testSettings)
			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)

			if err == nil {
				assert.Equal(t, []string{defaultExternalIPv6}, ipv6Networks)
			}
		})
	}
}

func TestNodesWaitForAllNodesAreReady(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		client      bool
		ready       bool
		assertError func(error) bool
	}{
		{
			name:        "all nodes ready",
			client:      true,
			ready:       true,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "nil client returns error",
			assertError: commonerrors.IsAPIClientNil,
		},
		{
			name:        "node not ready times out",
			client:      true,
			ready:       false,
			assertError: func(err error) bool { return errors.Is(err, context.DeadlineExceeded) },
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  []runtime.Object{buildDummyNodeWithReadiness(defaultNodeName, testCase.ready)},
					SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
				})
			}

			ready, err := WaitForAllNodesAreReady(testSettings, time.Second)
			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)
			assert.Equal(t, err == nil, ready)
		})
	}
}

func TestNodesWaitForAllNodesToReboot(t *testing.T) {
	t.Parallel()

	// There's no way to test for success without editing the node while the function is running so only failures
	// are covered here.
	testCases := []struct {
		name        string
		client      bool
		rebooted    bool
		assertError func(error) bool
	}{
		{
			name:        "nil client returns error",
			rebooted:    true,
			assertError: commonerrors.IsAPIClientNil,
		},
		{
			name:        "node does not reboot times out",
			client:      true,
			rebooted:    false,
			assertError: func(err error) bool { return errors.Is(err, context.DeadlineExceeded) },
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  []runtime.Object{buildDummyNodeWithReadiness(defaultNodeName, testCase.rebooted)},
					SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
				})
			}

			ready, err := WaitForAllNodesToReboot(testSettings, time.Second)
			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)
			assert.Equal(t, err == nil, ready)
		})
	}
}

// buildDummyNodeWithReadiness returns a dummy node with the specified value for the ready condition.
func buildDummyNodeWithReadiness(name string, ready bool) *corev1.Node {
	node := buildDummyNode(name)

	readyStatus := corev1.ConditionFalse
	if ready {
		readyStatus = corev1.ConditionTrue
	}

	node.Status.Conditions = []corev1.NodeCondition{{
		Type:   corev1.NodeReady,
		Status: readyStatus,
	}}

	return node
}
