package pod

import (
	"bytes"
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
	defaultPodName     = "test-pod"
	defaultPodNsName   = "test-ns"
	defaultPodImage    = "test-image"
	defaultPodNodeName = "test-node"
	defaultVolumeName  = "test-volume"
	defaultMountPath   = "/test"
)

var podRunningErrorMsg = fmt.Sprintf("can not redefine running pod. pod already running on node %s", defaultPodNodeName)

func newPodCommonTestConfig() testhelper.CommonTestConfig[corev1.Pod, Builder, *corev1.Pod, *Builder] {
	return testhelper.NewCommonTestConfig[corev1.Pod, Builder](
		corev1.AddToScheme,
		podGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

func assertPodBuilderError(t *testing.T, expected string, err error) {
	t.Helper()

	if expected == "" {
		assert.NoError(t, err)

		return
	}

	assert.EqualError(t, err, expected)
}

func TestPodNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *Builder {
				return NewBuilder(apiClient, name, nsname, defaultPodImage)
			},
			corev1.AddToScheme,
			podGVK,
		).ExecuteTests(t)
	})

	t.Run("valid builder gets default container", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		require.NoError(t, testBuilder.GetError())
		require.Len(t, testBuilder.Definition.Spec.Containers, 1)
		assert.Equal(t, defaultPodImage, testBuilder.Definition.Spec.Containers[0].Image)
	})

	t.Run("empty image returns error", func(t *testing.T) {
		t.Parallel()

		testBuilder := NewBuilder(clients.GetTestClients(clients.TestClientParams{}), defaultPodName, defaultPodNsName, "")

		require.EqualError(t, testBuilder.GetError(), "pod 'image' cannot be empty")
	})
}

func TestPodPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(Pull, corev1.AddToScheme, podGVK).ExecuteTests(t)

	t.Run("pulled pod retains containers", func(t *testing.T) {
		t.Parallel()

		testBuilder, err := Pull(buildTestClientWithDummyPod(), defaultPodName, defaultPodNsName)

		require.NoError(t, err)
		require.NotNil(t, testBuilder.Object)
		assert.NotEmpty(t, testBuilder.Definition.Spec.Containers)
	})
}

func TestPodDefineOnNode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		nodeName      string
		hasObject     bool
		expectedError string
	}{
		{
			nodeName:      defaultPodNodeName,
			hasObject:     false,
			expectedError: "",
		},
		{
			nodeName:      "",
			hasObject:     false,
			expectedError: "can not define pod on empty node",
		},
		{
			nodeName:      defaultPodNodeName,
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.DefineOnNode(testCase.nodeName)
		assertPodBuilderError(t, testCase.expectedError, testBuilder.GetError())

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.nodeName, testBuilder.Definition.Spec.NodeName)
		}
	}
}

func TestPodBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newPodCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		Run(t)
}

func TestPodWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newPodCommonTestConfig()).ExecuteTests(t)
}

func TestPodModifiersRejectInvalidBuilder(t *testing.T) {
	t.Parallel()

	modifiers := map[string]func(*Builder) *Builder{
		"DefineOnNode":                 func(builder *Builder) *Builder { return builder.DefineOnNode(defaultPodNodeName) },
		"RedefineDefaultCMD":           func(builder *Builder) *Builder { return builder.RedefineDefaultCMD([]string{"test"}) },
		"WithRestartPolicy":            func(builder *Builder) *Builder { return builder.WithRestartPolicy(corev1.RestartPolicyAlways) },
		"WithTolerationToMaster":       func(builder *Builder) *Builder { return builder.WithTolerationToMaster() },
		"WithTolerationToControlPlane": func(builder *Builder) *Builder { return builder.WithTolerationToControlPlane() },
		"WithToleration":               func(builder *Builder) *Builder { return builder.WithToleration(corev1.Toleration{}) },
		"WithNodeSelector": func(builder *Builder) *Builder {
			return builder.WithNodeSelector(map[string]string{"test": "test"})
		},
		"WithPrivilegedFlag": func(builder *Builder) *Builder { return builder.WithPrivilegedFlag() },
		"WithVolume": func(builder *Builder) *Builder {
			return builder.WithVolume(corev1.Volume{Name: defaultVolumeName})
		},
		"WithLocalVolume": func(builder *Builder) *Builder {
			return builder.WithLocalVolume(defaultVolumeName, defaultMountPath)
		},
		"WithAdditionalContainer":           func(builder *Builder) *Builder { return builder.WithAdditionalContainer(&corev1.Container{}) },
		"WithAdditionalInitContainer":       func(builder *Builder) *Builder { return builder.WithAdditionalInitContainer(&corev1.Container{}) },
		"WithSecondaryNetwork":              func(builder *Builder) *Builder { return builder.WithSecondaryNetwork(nil) },
		"WithHostNetwork":                   func(builder *Builder) *Builder { return builder.WithHostNetwork() },
		"WithHostPid":                       func(builder *Builder) *Builder { return builder.WithHostPid(true) },
		"RedefineDefaultContainer":          func(builder *Builder) *Builder { return builder.RedefineDefaultContainer(corev1.Container{}) },
		"WithHugePages":                     func(builder *Builder) *Builder { return builder.WithHugePages() },
		"WithSecurityContext":               func(builder *Builder) *Builder { return builder.WithSecurityContext(&corev1.PodSecurityContext{}) },
		"WithLabel":                         func(builder *Builder) *Builder { return builder.WithLabel("test", "test") },
		"WithLabels":                        func(builder *Builder) *Builder { return builder.WithLabels(map[string]string{"test": "test"}) },
		"WithTerminationGracePeriodSeconds": func(builder *Builder) *Builder { return builder.WithTerminationGracePeriodSeconds(1) },
		"WithCommand":                       func(builder *Builder) *Builder { return builder.WithCommand([]string{"test"}) },
		"WithResourceClaim":                 func(builder *Builder) *Builder { return builder.WithResourceClaim("claim", "template") },
	}

	for name, modifier := range modifiers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildInvalidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			result := modifier(testBuilder)

			assert.Same(t, testBuilder, result)
			assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))
		})
	}
}

func TestPodDeleteAndWait(t *testing.T) {
	t.Parallel()

	testPodDeleteHelper(t, func(builder *Builder) (*Builder, error) {
		return builder.DeleteAndWait(5 * time.Second)
	})
}

func TestPodDeleteImmediate(t *testing.T) {
	t.Parallel()

	testPodDeleteHelper(t, func(builder *Builder) (*Builder, error) {
		return builder.DeleteImmediate()
	})
}

func TestPodWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testBuilder   *Builder
		expectedError error
		invalid       bool
	}{
		{
			testBuilder:   buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder: buildInvalidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			invalid:     true,
		},
		{
			testBuilder:   buildValidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.WaitUntilDeleted(2 * time.Second)
		if testCase.invalid {
			assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}
	}
}

func TestPodWaitUntilReady(t *testing.T) {
	t.Parallel()

	testPodWaitUntilConditionHelper(t, func(builder *Builder) error {
		return builder.WaitUntilReady(time.Second)
	})
}

func TestPodWaitUntilCondition(t *testing.T) {
	t.Parallel()

	testPodWaitUntilConditionHelper(t, func(builder *Builder) error {
		return builder.WaitUntilCondition(corev1.PodReady, time.Second)
	})
}

func TestPodRedefineDefaultCMD(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		command       []string
		hasObject     bool
		expectedError string
	}{
		{
			command:       []string{"test"},
			hasObject:     false,
			expectedError: "",
		},
		{
			command:       []string{},
			hasObject:     false,
			expectedError: "",
		},
		{
			command:       []string{"test"},
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.RedefineDefaultCMD(testCase.command)
		assertPodBuilderError(t, testCase.expectedError, testBuilder.GetError())

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.command, testBuilder.Definition.Spec.Containers[0].Command)
		}
	}
}

func TestPodWithRestartPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		restartPolicy corev1.RestartPolicy
		hasObject     bool
		expectedError string
	}{
		{
			restartPolicy: corev1.RestartPolicyAlways,
			hasObject:     false,
			expectedError: "",
		},
		{
			restartPolicy: "",
			hasObject:     false,
			expectedError: "can not define pod with empty restart policy",
		},
		{
			restartPolicy: corev1.RestartPolicyAlways,
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.WithRestartPolicy(testCase.restartPolicy)
		assertPodBuilderError(t, testCase.expectedError, testBuilder.GetError())

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.restartPolicy, testBuilder.Definition.Spec.RestartPolicy)
		}
	}
}

func TestPodWithTolerationToMaster(t *testing.T) {
	t.Parallel()

	toleration := corev1.Toleration{
		Key:    "node-role.kubernetes.io/master",
		Effect: taintEffectNoSchedule,
	}

	testPodWithTolerationHelper(t, toleration, func(builder *Builder, toleration corev1.Toleration) *Builder {
		return builder.WithTolerationToMaster()
	})
}

func TestPodWithTolerationToControlPlane(t *testing.T) {
	t.Parallel()

	toleration := corev1.Toleration{
		Key:    nodeRoleKubernetesIoControlPlane,
		Effect: taintEffectNoSchedule,
	}

	testPodWithTolerationHelper(t, toleration, func(builder *Builder, toleration corev1.Toleration) *Builder {
		return builder.WithTolerationToControlPlane()
	})
}

func TestPodWithToleration(t *testing.T) {
	t.Parallel()

	toleration := corev1.Toleration{
		Key:    nodeRoleKubernetesIoControlPlane,
		Effect: taintEffectNoSchedule,
	}

	testPodWithTolerationHelper(t, toleration, func(builder *Builder, toleration corev1.Toleration) *Builder {
		return builder.WithToleration(toleration)
	})
}

func TestPodWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		nodeSelector  map[string]string
		hasObject     bool
		expectedError string
	}{
		{
			nodeSelector:  map[string]string{"test": "test"},
			hasObject:     false,
			expectedError: "",
		},
		{
			nodeSelector:  map[string]string{},
			hasObject:     false,
			expectedError: "can not define pod with empty nodeSelector",
		},
		{
			nodeSelector:  map[string]string{"test": "test"},
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.WithNodeSelector(testCase.nodeSelector)
		assertPodBuilderError(t, testCase.expectedError, testBuilder.GetError())

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.nodeSelector, testBuilder.Definition.Spec.NodeSelector)
		}
	}
}

func TestPodWithPrivilegedFlag(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		hasObject     bool
		expectedError string
	}{
		{
			hasObject:     false,
			expectedError: "",
		},
		{
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.WithPrivilegedFlag()
		assertPodBuilderError(t, testCase.expectedError, testBuilder.GetError())

		if testCase.expectedError == "" {
			assert.True(t, *testBuilder.Definition.Spec.Containers[0].SecurityContext.Privileged)
		}
	}
}

func TestPodWithVolume(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		volume        corev1.Volume
		expectedError string
	}{
		{
			volume:        corev1.Volume{Name: defaultVolumeName},
			expectedError: "",
		},
		{
			volume:        corev1.Volume{},
			expectedError: "the volume's name cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())
		testBuilder = testBuilder.WithVolume(testCase.volume)
		assertPodBuilderError(t, testCase.expectedError, testBuilder.GetError())

		if testCase.expectedError == "" {
			assert.Equal(t, []corev1.Volume{testCase.volume}, testBuilder.Definition.Spec.Volumes)
		}
	}
}

func TestPodWithLocalVolume(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		volumeName    string
		mountPath     string
		alreadyUsed   bool
		hasObject     bool
		expectedError string
	}{
		{
			volumeName:    defaultVolumeName,
			mountPath:     defaultMountPath,
			alreadyUsed:   false,
			hasObject:     false,
			expectedError: "",
		},
		{
			volumeName:    "",
			mountPath:     defaultMountPath,
			alreadyUsed:   false,
			hasObject:     false,
			expectedError: "'volumeName' parameter is empty",
		},
		{
			volumeName:    defaultVolumeName,
			mountPath:     "",
			alreadyUsed:   false,
			hasObject:     false,
			expectedError: "'mountPath' parameter is empty",
		},
		{
			volumeName:    defaultVolumeName,
			mountPath:     defaultMountPath,
			alreadyUsed:   true,
			hasObject:     false,
			expectedError: fmt.Sprintf("given mount %s already mounted to pod's container test", defaultVolumeName),
		},
		{
			volumeName:    defaultVolumeName,
			mountPath:     defaultMountPath,
			alreadyUsed:   false,
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		if testCase.alreadyUsed {
			testBuilder.Definition.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{
				Name:      testCase.volumeName,
				MountPath: testCase.mountPath,
			}}
		}

		testBuilder = testBuilder.WithLocalVolume(testCase.volumeName, testCase.mountPath)
		assertPodBuilderError(t, testCase.expectedError, testBuilder.GetError())

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.volumeName, testBuilder.Definition.Spec.Containers[0].VolumeMounts[0].Name)
			assert.Equal(t, testCase.mountPath, testBuilder.Definition.Spec.Containers[0].VolumeMounts[0].MountPath)
			assert.Equal(t, testCase.volumeName, testBuilder.Definition.Spec.Volumes[0].Name)
		}
	}
}

func TestPodIsHealthy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		exists          bool
		phase           corev1.PodPhase
		condition       corev1.PodConditionType
		expectedHealthy bool
	}{
		{
			exists:          true,
			phase:           corev1.PodSucceeded,
			condition:       "",
			expectedHealthy: true,
		},
		{
			exists:          true,
			phase:           corev1.PodRunning,
			condition:       corev1.PodReady,
			expectedHealthy: true,
		},
		{
			exists:          false,
			phase:           "",
			condition:       "",
			expectedHealthy: false,
		},
		{
			exists:          true,
			phase:           corev1.PodRunning,
			condition:       "",
			expectedHealthy: false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			testPod := buildDummyPodWithPhaseAndCondition(testCase.phase, testCase.condition, false)
			runtimeObjects = append(runtimeObjects, testPod)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
		})
		testBuilder := buildValidPodTestBuilder(testSettings)

		assert.Equal(t, testCase.expectedHealthy, testBuilder.IsHealthy())
	}
}

// TestPodExecCommandWithTimeout tests the ExecCommandWithTimeout method validation.
func TestPodExecCommandWithTimeout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		command       []string
		timeout       time.Duration
		containerName []string
		testBuilder   *Builder
		expectedError string
		invalid       bool
	}{
		{
			name:          "empty command",
			command:       []string{},
			timeout:       5 * time.Second,
			containerName: []string{},
			testBuilder:   buildValidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: "command must be provided",
		},
		{
			name:          "zero timeout",
			command:       []string{"echo", "test"},
			timeout:       0,
			containerName: []string{},
			testBuilder:   buildValidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: "timeout must be greater than 0",
		},
		{
			name:          "negative timeout",
			command:       []string{"echo", "test"},
			timeout:       -1 * time.Second,
			containerName: []string{},
			testBuilder:   buildValidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: "timeout must be greater than 0",
		},
		{
			name:          "invalid pod builder",
			command:       []string{"echo", "test"},
			timeout:       5 * time.Second,
			containerName: []string{},
			testBuilder:   buildInvalidPodTestBuilder(buildTestClientWithDummyPod()),
			invalid:       true,
		},
		{
			name:          "pod does not exist",
			command:       []string{"echo", "test"},
			timeout:       5 * time.Second,
			containerName: []string{},
			testBuilder:   buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "does not exist in namespace",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				buffer bytes.Buffer
				err    error
			)

			if len(testCase.containerName) > 0 {
				buffer, err = testCase.testBuilder.ExecCommandWithTimeout(
					testCase.command, testCase.timeout, testCase.containerName[0])
			} else {
				buffer, err = testCase.testBuilder.ExecCommandWithTimeout(
					testCase.command, testCase.timeout)
			}

			assert.Error(t, err)

			if testCase.invalid {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
			} else {
				assert.Contains(t, err.Error(), testCase.expectedError)
			}

			assert.Empty(t, buffer.String())
		})
	}
}

// TestPodExecCommandWithTimeoutContainerSelection tests container name parameter handling.
func TestPodExecCommandWithTimeoutContainerSelection(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		containerName []string
	}{
		{
			name:          "no container specified - uses default",
			containerName: []string{},
		},
		{
			name:          "container specified",
			containerName: []string{"custom-container"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Use a builder with non-existent pod to trigger the "does not exist" error
			// This allows us to test parameter handling without hitting REST client issues
			builder := buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

			var err error
			if len(testCase.containerName) > 0 {
				_, err = builder.ExecCommandWithTimeout(
					[]string{"echo", "test"}, 5*time.Second, testCase.containerName...)
			} else {
				_, err = builder.ExecCommandWithTimeout(
					[]string{"echo", "test"}, 5*time.Second)
			}

			// Should get "does not exist" error, not a parameter validation error
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "does not exist in namespace")
		})
	}
}

func testPodDeleteHelper(t *testing.T, deleteFunc func(builder *Builder) (*Builder, error)) {
	t.Helper()

	testCases := []struct {
		testBuilder   *Builder
		expectedError error
		invalid       bool
	}{
		{
			testBuilder:   buildValidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: nil,
		},
		{
			testBuilder: buildInvalidPodTestBuilder(buildTestClientWithDummyPod()),
			invalid:     true,
		},
		{
			testBuilder:   buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := deleteFunc(testCase.testBuilder)
		if testCase.invalid {
			assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}

// testPodWaitUntilConditionHelper handles the test cases where a function waits for a condition. This helper uses
// corev1.PodReady so that it can also be used for testing WaitUntilReady.
func testPodWaitUntilConditionHelper(t *testing.T, waitFunc func(builder *Builder) error) {
	t.Helper()

	testCases := []struct {
		valid         bool
		ready         bool
		expectedError error
		invalid       bool
	}{
		{
			valid:         true,
			ready:         true,
			expectedError: nil,
		},
		{
			valid:   false,
			ready:   true,
			invalid: true,
		},
		{
			valid:         true,
			ready:         false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var testBuilder *Builder

		if testCase.valid {
			pod := buildDummyPodWithPhaseAndCondition(corev1.PodRunning, corev1.PodReady, false)

			if !testCase.ready {
				pod.Status.Conditions[0].Status = corev1.ConditionFalse
			}

			testBuilder = buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{pod},
				SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
			}))
		} else {
			testBuilder = buildInvalidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		}

		err := waitFunc(testBuilder)
		if testCase.invalid {
			assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}
	}
}

// testPodWithTolerationHelper handles the test cases where a function applies the specified toleration to a pod.
func testPodWithTolerationHelper(
	t *testing.T, toleration corev1.Toleration, testFunc func(builder *Builder, toleration corev1.Toleration) *Builder) {
	t.Helper()

	testCases := []struct {
		hasObject     bool
		expectedError string
	}{
		{
			hasObject:     false,
			expectedError: "",
		},
		{
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testFunc(testBuilder, toleration)
		assertPodBuilderError(t, testCase.expectedError, testBuilder.GetError())

		if testCase.expectedError == "" {
			assert.Equal(t, []corev1.Toleration{toleration}, testBuilder.Definition.Spec.Tolerations)
		}
	}
}

// buildDummyPod returns a Pod with the provided name, nsname, and container image.
func buildDummyPod(name, nsname, image string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "test",
					Image:   image,
					Command: []string{defaultShellBinBash, "-c", "sleep INF"},
				},
			},
		},
	}
}

// buildTestClientWithDummyPod returns a client with a dummy Pod.
func buildTestClientWithDummyPod() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPod(defaultPodName, defaultPodNsName, defaultPodImage),
		},
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
}

func TestPodWithCommand(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		command       []string
		hasObject     bool
		expectedError string
	}{
		{
			command:       []string{"sleep", "infinity"},
			hasObject:     false,
			expectedError: "",
		},
		{
			command:       []string{},
			hasObject:     false,
			expectedError: "pod command cannot be empty",
		},
		{
			command:       []string{"sleep", "infinity"},
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		result := testBuilder.WithCommand(testCase.command)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.command, result.Definition.Spec.Containers[0].Command)
		} else {
			assertPodBuilderError(t, testCase.expectedError, result.GetError())
		}
	}
}

func TestPodWithCommandNoContainers(t *testing.T) {
	t.Parallel()

	testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())
	testBuilder.Definition.Spec.Containers = nil

	result := testBuilder.WithCommand([]string{"sleep", "infinity"})
	assertPodBuilderError(t, "pod has no containers to set command on", result.GetError())
}

func TestPodWithResourceClaim(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		claimName         string
		claimTemplateName string
		hasObject         bool
		expectedError     string
	}{
		{
			claimName:         "test-claim",
			claimTemplateName: "test-template",
			hasObject:         false,
			expectedError:     "",
		},
		{
			claimName:         "",
			claimTemplateName: "test-template",
			hasObject:         false,
			expectedError:     "ResourceClaim name cannot be empty",
		},
		{
			claimName:         "test-claim",
			claimTemplateName: "",
			hasObject:         false,
			expectedError:     "ResourceClaimTemplate name cannot be empty",
		},
		{
			claimName:         "test-claim",
			claimTemplateName: "test-template",
			hasObject:         true,
			expectedError:     podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		result := testBuilder.WithResourceClaim(testCase.claimName, testCase.claimTemplateName)

		if testCase.expectedError == "" {
			assert.Len(t, result.Definition.Spec.ResourceClaims, 1)
			assert.Equal(t, testCase.claimName, result.Definition.Spec.ResourceClaims[0].Name)
			assert.Equal(t, testCase.claimTemplateName,
				*result.Definition.Spec.ResourceClaims[0].ResourceClaimTemplateName)
			assert.Len(t, result.Definition.Spec.Containers[0].Resources.Claims, 1)
			assert.Equal(t, testCase.claimName,
				result.Definition.Spec.Containers[0].Resources.Claims[0].Name)
		} else {
			assertPodBuilderError(t, testCase.expectedError, result.GetError())
		}
	}
}

func TestPodWithResourceClaimDuplicate(t *testing.T) {
	t.Parallel()

	testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

	result := testBuilder.
		WithResourceClaim("claim-a", "template-a").
		WithResourceClaim("claim-a", "template-a")

	assert.Len(t, result.Definition.Spec.ResourceClaims, 1,
		"Duplicate claim name should not create a second entry")
	assert.Len(t, result.Definition.Spec.Containers[0].Resources.Claims, 1)
}

func TestPodWithResourceClaimDistinct(t *testing.T) {
	t.Parallel()

	testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

	result := testBuilder.
		WithResourceClaim("claim-a", "template-a").
		WithResourceClaim("claim-b", "template-b")

	assert.Len(t, result.Definition.Spec.ResourceClaims, 2,
		"Distinct claim names should create separate entries")
	assert.Equal(t, "claim-a", result.Definition.Spec.ResourceClaims[0].Name)
	assert.Equal(t, "claim-b", result.Definition.Spec.ResourceClaims[1].Name)
	assert.Len(t, result.Definition.Spec.Containers[0].Resources.Claims, 2)
}

// buildValidPodTestBuilder returns a valid Pod builder for testing.
func buildValidPodTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultPodName, defaultPodNsName, defaultPodImage)
}

// buildInvalidPodTestBuilder returns an invalid Pod builder for testing.
func buildInvalidPodTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultPodName, "", defaultPodImage)
}

// buildDummyPodWithPhaseAndCondition returns a Pod object with the specified phase, condition, and restart policy.
func buildDummyPodWithPhaseAndCondition(
	phase corev1.PodPhase, conditionType corev1.PodConditionType, neverRestart bool) *corev1.Pod {
	pod := buildDummyPod(defaultPodName, defaultPodNsName, defaultPodImage)
	pod.Status.Phase = phase
	pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
		Type:   conditionType,
		Status: corev1.ConditionTrue,
	})

	if neverRestart {
		pod.Spec.RestartPolicy = corev1.RestartPolicyNever
	}

	return pod
}
