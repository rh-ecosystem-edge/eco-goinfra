package pod

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestWaitForPodsInNamespacesHealthy(t *testing.T) {
	testCases := []struct {
		namespaces    []string
		listOptions   []metav1.ListOptions
		healthy       bool
		failed        bool
		client        bool
		expectedError error
	}{
		{
			namespaces:    nil,
			listOptions:   nil,
			healthy:       true,
			failed:        false,
			client:        true,
			expectedError: nil,
		},
		{
			namespaces:    nil,
			listOptions:   nil,
			healthy:       false,
			failed:        true,
			client:        true,
			expectedError: nil,
		},
		{
			namespaces:    nil,
			listOptions:   []metav1.ListOptions{{}, {}},
			healthy:       true,
			failed:        false,
			client:        true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			namespaces:    nil,
			listOptions:   nil,
			healthy:       true,
			failed:        false,
			client:        false,
			expectedError: fmt.Errorf("podList 'apiClient' cannot be empty"),
		},
		{
			namespaces:    nil,
			listOptions:   nil,
			healthy:       false,
			failed:        false,
			client:        true,
			expectedError: context.DeadlineExceeded,
		},
		{
			namespaces:    []string{defaultPodNsName + "-has-no-pods"},
			listOptions:   nil,
			healthy:       false,
			failed:        false,
			client:        true,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testPod := buildDummyPod(defaultPodName, defaultPodNsName, defaultPodImage)
			if testCase.healthy {
				testPod = buildDummyPodWithPhaseAndCondition(corev1.PodSucceeded, corev1.PodReady, false)
			} else if testCase.failed {
				testPod = buildDummyPodWithPhaseAndCondition(corev1.PodFailed, corev1.PodReady, true)
			}

			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: []runtime.Object{testPod},
			})
		}

		err := WaitForPodsInNamespacesHealthy(testSettings, testCase.namespaces, time.Second, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestWaitForPodsRemainingHealthy(t *testing.T) {
	testCases := []struct {
		name        string
		states      []string
		stable      time.Duration
		timeout     time.Duration
		elapsed     time.Duration
		wantTimeout bool
	}{
		{"healthy window", []string{"healthy"}, 2 * time.Minute, 10 * time.Minute, 2 * time.Minute, false},
		{"unhealthy resets", []string{"healthy", "unhealthy", "healthy"}, 10 * time.Second, time.Minute, 20 * time.Second, false},
		{"single list error tolerated", []string{"healthy", "error", "healthy"},
			10 * time.Second, time.Minute, 10 * time.Second, false},
		{"two list errors tolerated", []string{"healthy", "error", "error", "healthy"},
			10 * time.Second, time.Minute, 15 * time.Second, false},
		{"three list errors reset", []string{"healthy", "error", "error", "error", "healthy"},
			10 * time.Second, time.Minute, 30 * time.Second, false},
		{"errors cannot confirm health", []string{"healthy", "error"}, 10 * time.Second, 20 * time.Second, 20 * time.Second, true},
		{"errors cannot start window", []string{"error", "healthy"}, 10 * time.Second, time.Minute, 15 * time.Second, false},
		{"between polls", []string{"healthy"}, 7 * time.Second, time.Minute, 10 * time.Second, false},
		{"failed pod", []string{"failed"}, 0, time.Minute, 0, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				apiClient := clients.GetTestClients(clients.TestClientParams{})
				fakeClient, ok := apiClient.K8sClient.(*fake.Clientset)
				require.True(t, ok)

				calls := 0

				fakeClient.PrependReactor("list", "pods", func(action clienttesting.Action) (bool, runtime.Object, error) {
					state := testCase.states[min(calls, len(testCase.states)-1)]
					calls++

					if state == "error" {
						return true, nil, fmt.Errorf("temporary list failure")
					}

					pod := buildDummyPodWithPhaseAndCondition(corev1.PodRunning, corev1.PodReady, true)

					switch state {
					case "unhealthy":
						pod.Status.Phase = corev1.PodPending
					case "failed":
						pod.Status.Phase = corev1.PodFailed
						pod.Spec.RestartPolicy = corev1.RestartPolicyAlways
					}

					return true, &corev1.PodList{Items: []corev1.Pod{*pod}}, nil
				})

				start := time.Now()
				err := WaitForPodsRemainingHealthy(apiClient, nil, testCase.timeout, testCase.stable)

				if testCase.wantTimeout {
					assert.ErrorIs(t, err, context.DeadlineExceeded)
				} else {
					assert.NoError(t, err)
				}

				assert.Equal(t, testCase.elapsed, time.Since(start))
			})
		})
	}
}

func TestWaitForPodsRemainingHealthyValidation(t *testing.T) {
	apiClient := clients.GetTestClients(clients.TestClientParams{})

	testCases := []struct {
		name       string
		client     *clients.Settings
		namespaces []string
		timeout    time.Duration
		stable     time.Duration
		options    []metav1.ListOptions
	}{
		{name: "nil client", timeout: time.Minute},
		{name: "multiple options", client: apiClient, timeout: time.Minute, options: []metav1.ListOptions{{}, {}}},
		{name: "zero timeout", client: apiClient},
		{name: "negative timeout", client: apiClient, timeout: -time.Second},
		{name: "negative window", client: apiClient, timeout: time.Minute, stable: -time.Second},
		{name: "window equals timeout", client: apiClient, timeout: time.Second, stable: time.Second},
		{name: "window exceeds timeout", client: apiClient, timeout: time.Second, stable: time.Minute},
		{name: "empty namespace", client: apiClient, timeout: time.Minute, namespaces: []string{""}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := WaitForPodsRemainingHealthy(testCase.client, testCase.namespaces,
				testCase.timeout, testCase.stable, testCase.options...)
			assert.Error(t, err)
		})
	}
}
