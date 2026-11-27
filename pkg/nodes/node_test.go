package nodes

import (
	"context"
	"fmt"
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

const (
	defaultNodeName         = "test-node"
	defaultNodeLabel        = "node-role.kubernetes.io/control-plane"
	defaultExternalNetworks = `{"ipv4":"10.0.0.0/8","ipv6":"fd00::/8"}`
	defaultExternalIPv4     = "10.0.0.0/8"
	defaultExternalIPv6     = "fd00::/8"
)

var nodeGVK = corev1.SchemeGroupVersion.WithKind("Node")

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig[corev1.Node, Builder](
		NewBuilder, corev1.AddToScheme, nodeGVK).ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig[corev1.Node, Builder](
		Pull, corev1.AddToScheme, nodeGVK).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newNodeCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newNodeCommonTestConfig()).ExecuteTests(t)
}

func TestNodeWithNewLabel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		key            string
		value          string
		keyExists      bool
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:  "adds label",
			key:   defaultNodeLabel,
			value: "test-value",
		},
		{
			name:          "empty key sets builder error",
			key:           "",
			value:         "test-value",
			expectedError: "error to set empty key to node",
		},
		{
			name:          "existing label sets builder error",
			key:           defaultNodeLabel,
			value:         "test-value",
			keyExists:     true,
			expectedError: fmt.Sprintf("cannot overwrite existing node label: %s", defaultNodeLabel),
		},
		{
			name:           "invalid builder short circuits",
			key:            defaultNodeLabel,
			value:          "test-value",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			if testCase.keyExists {
				testBuilder.Definition.Labels = map[string]string{testCase.key: "existing-value"}
			}

			if testCase.invalidBuilder {
				testBuilder = buildInvalidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			}

			result := testBuilder.WithNewLabel(testCase.key, testCase.value)
			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.value, result.Definition.Labels[testCase.key])
		})
	}
}

func TestNodeRemoveLabel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		key            string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name: "removes label",
			key:  defaultNodeLabel,
		},
		{
			name:          "empty key sets builder error",
			key:           "",
			expectedError: "error to remove empty key from node",
		},
		{
			name:           "invalid builder short circuits",
			key:            defaultNodeLabel,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			testBuilder.Definition.Labels = map[string]string{testCase.key: "test-value"}

			if testCase.invalidBuilder {
				testBuilder = buildInvalidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			}

			result := testBuilder.RemoveLabel(testCase.key, "test-value")
			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.NotContains(t, result.Definition.Labels, testCase.key)
		})
	}
}

func TestNodeExternalIPv4Network(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		objectNil         bool
		externalAddresses string
		expectedError     error
	}{
		{
			name:              "returns external IPv4 network",
			externalAddresses: defaultExternalNetworks,
		},
		{
			name:          "nil object returns error",
			objectNil:     true,
			expectedError: fmt.Errorf("cannot collect external networks when node object is nil"),
		},
		{
			name:          "missing annotation returns error",
			expectedError: fmt.Errorf("node %s does not have external addresses annotation", defaultNodeName),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidNodeTestBuilder(buildTestClientWithDummyNode())

			if !testCase.objectNil {
				testBuilder.Object = testBuilder.Definition

				if testCase.externalAddresses != "" {
					testBuilder.Object.Annotations = map[string]string{ovnExternalAddresses: testCase.externalAddresses}
				}
			}

			externalIPv4, err := testBuilder.ExternalIPv4Network()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, defaultExternalIPv4, externalIPv4)
			}
		})
	}
}

func TestNodeExternalIPv6Network(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		objectNil         bool
		externalAddresses string
		expectedError     error
	}{
		{
			name:              "returns external IPv6 network",
			externalAddresses: defaultExternalNetworks,
		},
		{
			name:          "nil object returns error",
			objectNil:     true,
			expectedError: fmt.Errorf("cannot collect external networks when node object is nil"),
		},
		{
			name:          "missing annotation returns error",
			expectedError: fmt.Errorf("node %s does not have external addresses annotation", defaultNodeName),
		},
		{
			name:              "missing IPv6 address returns error",
			externalAddresses: `{"ipv4":"10.0.0.0/8"}`,
			expectedError:     fmt.Errorf("node %s does not have external ipv6 addresses", defaultNodeName),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidNodeTestBuilder(buildTestClientWithDummyNode())

			if !testCase.objectNil {
				testBuilder.Object = testBuilder.Definition

				if testCase.externalAddresses != "" {
					testBuilder.Object.Annotations = map[string]string{ovnExternalAddresses: testCase.externalAddresses}
				}
			}

			externalIPv6, err := testBuilder.ExternalIPv6Network()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, defaultExternalIPv6, externalIPv6)
			}
		})
	}
}

func TestNodeIsReady(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		exists         bool
		readyCondition bool
		ready          bool
		expectedError  error
	}{
		{
			name:           "ready node",
			exists:         true,
			readyCondition: true,
			ready:          true,
		},
		{
			name:          "missing node returns error",
			expectedError: fmt.Errorf("node object %s does not exist", defaultNodeName),
		},
		{
			name:           "missing Ready condition returns error",
			exists:         true,
			readyCondition: false,
			expectedError:  fmt.Errorf("the Ready condition could not be found for node %s", defaultNodeName),
		},
		{
			name:           "not ready node",
			exists:         true,
			readyCondition: true,
			ready:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				if testCase.readyCondition {
					runtimeObjects = append(runtimeObjects, buildDummyNodeWithReadiness(defaultNodeName, testCase.ready))
				} else {
					runtimeObjects = append(runtimeObjects, buildDummyNode(defaultNodeName))
				}
			}

			testBuilder := buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
			}))

			ready, err := testBuilder.IsReady()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.ready, ready)
			}
		})
	}
}

func TestNodeWaitForCondition(t *testing.T) {
	t.Parallel()

	expectedCondition := corev1.NodeCondition{Type: corev1.NodeReady, Status: corev1.ConditionTrue}
	testNodeWaitUntilConditionHelper(t, expectedCondition, func(testBuilder *Builder, timeout time.Duration) error {
		_, err := testBuilder.WaitForCondition(expectedCondition, timeout)

		return err
	})
}

func TestNodeWaitUntilConditionTrue(t *testing.T) {
	t.Parallel()

	expectedCondition := corev1.NodeCondition{Type: corev1.NodeReady, Status: corev1.ConditionTrue}
	testNodeWaitUntilConditionHelper(t, expectedCondition, func(testBuilder *Builder, timeout time.Duration) error {
		return testBuilder.WaitUntilConditionTrue(expectedCondition.Type, timeout)
	})
}

func TestNodeWaitUntilReady(t *testing.T) {
	t.Parallel()

	expectedCondition := corev1.NodeCondition{Type: corev1.NodeReady, Status: corev1.ConditionTrue}
	testNodeWaitUntilConditionHelper(t, expectedCondition, func(testBuilder *Builder, timeout time.Duration) error {
		return testBuilder.WaitUntilReady(timeout)
	})
}

func TestNodeWaitUntilNotReady(t *testing.T) {
	t.Parallel()

	builderError := fmt.Errorf("test error")

	testCases := []struct {
		name          string
		exists        bool
		hasCondition  bool
		conditionVal  corev1.ConditionStatus
		builderError  error
		expectedError error
	}{
		{
			name:         "exists and Ready is False",
			exists:       true,
			hasCondition: true,
			conditionVal: corev1.ConditionFalse,
		},
		{
			name:         "exists and Ready is Unknown",
			exists:       true,
			hasCondition: true,
			conditionVal: corev1.ConditionUnknown,
		},
		{
			name:          "exists and Ready is True",
			exists:        true,
			hasCondition:  true,
			conditionVal:  corev1.ConditionTrue,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "does not exist",
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "exists but has no Ready condition",
			exists:        true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:         "invalid builder with error message",
			exists:       true,
			hasCondition: true,
			conditionVal: corev1.ConditionFalse,
			builderError: builderError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				if testCase.hasCondition {
					runtimeObjects = append(runtimeObjects,
						buildDummyNodeWithCondition(defaultNodeName, corev1.NodeReady, testCase.conditionVal))
				} else {
					runtimeObjects = append(runtimeObjects, buildDummyNode(defaultNodeName))
				}
			}

			testBuilder := buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
			}))
			testBuilder.SetError(testCase.builderError)

			err := testBuilder.WaitUntilNotReady(time.Second)

			if testCase.builderError != nil {
				assert.ErrorIs(t, err, testCase.builderError)

				return
			}

			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

// testNodeWaitUntilConditionHelper tests methods that wait for a specific condition status. Depending on the test case,
// the condition may be set to expectedCondition.
func testNodeWaitUntilConditionHelper(
	t *testing.T,
	expectedCondition corev1.NodeCondition,
	testedFunc func(testBuilder *Builder, timeout time.Duration) error) {
	t.Helper()

	testCases := []struct {
		name           string
		exists         bool
		hasCondition   bool
		conditionUnmet bool
		expectedError  error
	}{
		{
			name:         "exists and has condition met",
			exists:       true,
			hasCondition: true,
		},
		{
			name:          "does not exist",
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "exists and does not have condition",
			exists:        true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:           "exists and has condition but is unmet",
			exists:         true,
			hasCondition:   true,
			conditionUnmet: true,
			expectedError:  context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				if testCase.hasCondition {
					status := expectedCondition.Status
					if testCase.conditionUnmet {
						status = "invalid"
					}

					runtimeObjects = append(runtimeObjects,
						buildDummyNodeWithCondition(defaultNodeName, expectedCondition.Type, status))
				} else {
					runtimeObjects = append(runtimeObjects, buildDummyNode(defaultNodeName))
				}
			}

			testBuilder := buildValidNodeTestBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
			}))

			err := testedFunc(testBuilder, time.Second)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

// buildDummyNode returns a Node with the provided name.
func buildDummyNode(name string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildDummyNodeWithCondition returns a Node with one provided condition.
func buildDummyNodeWithCondition(
	name string, conditionType corev1.NodeConditionType, status corev1.ConditionStatus) *corev1.Node {
	node := buildDummyNode(name)

	node.Status.Conditions = []corev1.NodeCondition{{
		Type:   conditionType,
		Status: status,
	}}

	return node
}

// buildTestClientWithDummyNode returns a client with a dummy node.
func buildTestClientWithDummyNode() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyNode(defaultNodeName)},
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
}

func buildValidNodeTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultNodeName)
}

func buildInvalidNodeTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "")
}

func newNodeCommonTestConfig() testhelper.CommonTestConfig[corev1.Node, Builder, *corev1.Node, *Builder] {
	return testhelper.NewCommonTestConfig[corev1.Node, Builder](
		corev1.AddToScheme, nodeGVK, testhelper.ResourceScopeClusterScoped)
}
