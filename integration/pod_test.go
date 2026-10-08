//go:build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/namespace"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/pod"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
)

func TestPodCreate(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = CreateRandomNamespace()
		podName       = "create-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder, err := namespace.NewBuilder(client, testNamespace).Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"})
	containerDefinition, err := testContainerBuilder.GetContainerCfg()
	assert.Nil(t, err)

	podBuilder := pod.NewBuilder(client, podName, testNamespace, containerImage)
	podBuilder = podBuilder.RedefineDefaultContainer(*containerDefinition)

	// Create a pod in the namespace
	_, err = podBuilder.CreateAndWaitUntilRunning(timeoutDuration)
	assert.Nil(t, err)

	// Check if the pod was created
	podBuilder, err = pod.Pull(client, podName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, podBuilder.Object)
}

func TestPodDelete(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = CreateRandomNamespace()
		podName       = "delete-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder, err := namespace.NewBuilder(client, testNamespace).Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"})
	containerDefinition, err := testContainerBuilder.GetContainerCfg()
	assert.Nil(t, err)

	podBuilder := pod.NewBuilder(client, podName, testNamespace, containerImage)
	podBuilder = podBuilder.RedefineDefaultContainer(*containerDefinition)

	// Create a pod in the namespace
	_, err = podBuilder.CreateAndWaitUntilRunning(timeoutDuration)
	assert.Nil(t, err)

	defer func() {
		_, err = podBuilder.DeleteAndWait(timeoutDuration)
		assert.Nil(t, err)

		// Check if the pod was deleted
		podBuilder, err = pod.Pull(client, podName, testNamespace)
		assert.EqualError(t, err, fmt.Sprintf("pod object %s does not exist in namespace %s", podName, testNamespace))
	}()

	// Check if the pod was created
	podBuilder, err = pod.Pull(client, podName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, podBuilder.Object)
}

func TestPodExecCommand(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	var (
		testNamespace = CreateRandomNamespace()
		podName       = "exec-test"
	)

	// Create a namespace in the cluster using the namespaces package
	namespaceBuilder, err := namespace.NewBuilder(client, testNamespace).Create()
	assert.Nil(t, err)

	// Defer the deletion of the namespace
	defer func() {
		// Delete the namespace
		err := namespaceBuilder.Delete()
		assert.Nil(t, err)
	}()

	testContainerBuilder := pod.NewContainerBuilder("test", containerImage, []string{"sleep", "3600"})
	containerDefinition, err := testContainerBuilder.GetContainerCfg()
	assert.Nil(t, err)

	podBuilder := pod.NewBuilder(client, podName, testNamespace, containerImage)
	podBuilder = podBuilder.RedefineDefaultContainer(*containerDefinition)

	// Create a pod in the namespace
	podBuilder, err = podBuilder.CreateAndWaitUntilRunning(timeoutDuration)
	assert.Nil(t, err)

	// Check if the pod was created
	podBuilder, err = pod.Pull(client, podName, testNamespace)
	assert.Nil(t, err)
	assert.NotNil(t, podBuilder.Object)

	// Execute a command in the pod
	buffer, err := podBuilder.ExecCommand([]string{"sh", "-c", "echo f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2"})
	assert.Nil(t, err)
	assert.Equal(t, "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2\r\n", buffer.String())
}

func TestPodExecCommandWithContext(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	testNamespace := CreateRandomNamespace()
	namespaceBuilder, err := namespace.NewBuilder(client, testNamespace).Create()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, namespaceBuilder.DeleteAndWait(timeoutDuration))
	}()

	containerDefinition, err := CreateTestContainerDefinition(
		"test-container", containerImage, []string{"sleep", "3600"})
	assert.NoError(t, err)

	podBuilder := pod.NewBuilder(client, "exec-context-test", testNamespace, containerImage).
		RedefineDefaultContainer(*containerDefinition)
	podBuilder, err = podBuilder.CreateAndWaitUntilRunning(timeoutDuration)
	assert.NoError(t, err)
	defer func() {
		_, err := podBuilder.DeleteAndWait(timeoutDuration)
		assert.NoError(t, err)
	}()

	t.Run("streams stdout and stderr separately", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := podBuilder.ExecCommandWithContext(context.Background(), corev1.PodExecOptions{
			Container: "test-container",
			Command:   []string{"/bin/sh", "-c", "printf stdout-marker; printf stderr-marker >&2"},
			Stdout:    true,
			Stderr:    true,
		}, remotecommand.StreamOptions{
			Stdout: &stdout,
			Stderr: &stderr,
		})

		assert.NoError(t, err)
		assert.Equal(t, "stdout-marker", stdout.String())
		assert.Equal(t, "stderr-marker", stderr.String())
	})

	t.Run("forwards stdin", func(t *testing.T) {
		var stdout bytes.Buffer
		err := podBuilder.ExecCommandWithContext(context.Background(), corev1.PodExecOptions{
			Container: "test-container",
			Command:   []string{"/bin/sh", "-c", "read value; printf 'received:%s' \"$value\""},
			Stdin:     true,
			Stdout:    true,
		}, remotecommand.StreamOptions{
			Stdin:  strings.NewReader("input-marker\n"),
			Stdout: &stdout,
		})

		assert.NoError(t, err)
		assert.Equal(t, "received:input-marker", stdout.String())
	})

	t.Run("cancels a running command when its context deadline expires", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := podBuilder.ExecCommandWithContext(ctx, corev1.PodExecOptions{
			Container: "test-container",
			Command:   []string{"/bin/sh", "-c", "sleep 30"},
			Stdout:    true,
			Stderr:    true,
		}, remotecommand.StreamOptions{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})

		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

// TestExecCommandWithTimeoutEnforcement tests that ExecCommandWithTimeout
// properly enforces the timeout parameter for long-running commands.
//
// This test verifies the fix for the timeout bug where commands would hang
// for 15+ minutes instead of timing out after the specified duration.
//
// Test Cases:
//  1. Fast command with long timeout - should succeed
//  2. Long command with short timeout - should timeout
//  3. Verify timeout occurs at expected time (not command duration)
func TestExecCommandWithTimeoutEnforcement(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client, "Failed to create Kubernetes client")

	testNamespace := CreateRandomNamespace()
	podName := "timeout-enforcement-test"

	// Create namespace
	nsBuilder, err := namespace.NewBuilder(client, testNamespace).Create()
	assert.Nil(t, err, "Failed to create test namespace")

	// Cleanup namespace after test
	defer func() {
		err := nsBuilder.DeleteAndWait(timeoutDuration)
		assert.Nil(t, err, "Failed to delete test namespace")
	}()

	// Create pod with sleep-capable container
	containerDef, err := CreateTestContainerDefinition("test-container",
		containerImage, []string{"sleep", "3600"})
	assert.Nil(t, err, "Failed to create container definition")

	podBuilder := pod.NewBuilder(client, podName, testNamespace, containerImage)
	podBuilder = podBuilder.RedefineDefaultContainer(*containerDef)

	// Create and wait for pod to be running
	podBuilder, err = podBuilder.CreateAndWaitUntilRunning(timeoutDuration)
	assert.Nil(t, err, "Failed to create and start pod")

	// Ensure pod is deleted at the end
	defer func() {
		_, err := podBuilder.DeleteAndWait(timeoutDuration)
		assert.Nil(t, err, "Failed to delete test pod")
	}()

	// Test Case 1: Fast command with long timeout - should succeed
	t.Run("fast command with long timeout succeeds", func(t *testing.T) {
		buffer, err := podBuilder.ExecCommandWithTimeout(
			[]string{"/bin/sh", "-c", "echo 'success'"},
			10*time.Second,
		)
		assert.Nil(t, err, "Fast command should succeed")
		assert.Contains(t, buffer.String(), "success",
			"Command output should contain expected string")
	})

	// Test Case 2: Long command with short timeout - should timeout
	t.Run("long command with short timeout fails", func(t *testing.T) {
		start := time.Now()
		_, err := podBuilder.ExecCommandWithTimeout(
			[]string{"/bin/sh", "-c", "sleep 30"},
			3*time.Second,
		)
		elapsed := time.Since(start)

		// Should return an error
		assert.Error(t, err, "Long command should timeout")

		// Error should indicate timeout
		assert.ErrorIs(t, err, context.DeadlineExceeded)

		// Should timeout close to 3 seconds, NOT 30 seconds
		assert.Less(t, elapsed, 5*time.Second,
			"Command should timeout after ~3s, but took %v", elapsed)
		assert.Greater(t, elapsed, 2*time.Second,
			"Timeout should be enforced, took %v (expected ~3s)", elapsed)
	})

	// Test Case 3: Verify timeout occurs at expected time
	t.Run("timeout occurs at expected time", func(t *testing.T) {
		testCases := []struct {
			name            string
			sleepDuration   int
			timeout         time.Duration
			shouldTimeout   bool
			maxElapsedTime  time.Duration
			minElapsedTime  time.Duration
			expectedInError string
		}{
			{
				name:            "2s sleep with 5s timeout - succeeds",
				sleepDuration:   2,
				timeout:         5 * time.Second,
				shouldTimeout:   false,
				maxElapsedTime:  4 * time.Second,
				minElapsedTime:  1 * time.Second,
				expectedInError: "",
			},
			{
				name:            "10s sleep with 2s timeout - times out",
				sleepDuration:   10,
				timeout:         2 * time.Second,
				shouldTimeout:   true,
				maxElapsedTime:  4 * time.Second,
				minElapsedTime:  1 * time.Second,
				expectedInError: "context deadline exceeded",
			},
			{
				name:            "20s sleep with 4s timeout - times out",
				sleepDuration:   20,
				timeout:         4 * time.Second,
				shouldTimeout:   true,
				maxElapsedTime:  6 * time.Second,
				minElapsedTime:  3 * time.Second,
				expectedInError: "context deadline exceeded",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				start := time.Now()
				_, err := podBuilder.ExecCommandWithTimeout(
					[]string{"/bin/sh", "-c", fmt.Sprintf("sleep %d", tc.sleepDuration)},
					tc.timeout,
				)
				elapsed := time.Since(start)

				if tc.shouldTimeout {
					assert.Error(t, err, "Command should timeout")
					assert.Contains(t, err.Error(), tc.expectedInError,
						"Error should contain: %s", tc.expectedInError)
					assert.Less(t, elapsed, tc.maxElapsedTime,
						"Timeout took too long: %v (expected < %v)", elapsed, tc.maxElapsedTime)
					assert.Greater(t, elapsed, tc.minElapsedTime,
						"Timeout too fast: %v (expected > %v)", elapsed, tc.minElapsedTime)
				} else {
					assert.Nil(t, err, "Command should succeed")
					assert.Less(t, elapsed, tc.maxElapsedTime,
						"Command took too long: %v (expected < %v)", elapsed, tc.maxElapsedTime)
				}
			})
		}
	})
}
