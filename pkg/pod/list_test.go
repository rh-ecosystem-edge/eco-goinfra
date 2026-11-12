package pod

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//nolint:funlen // Since this is only long because of the number of test cases, we can ignore the length.
func TestWaitForPodsInNamespacesHealthy(t *testing.T) {
	testCases := []struct {
		name          string
		namespaces    []string
		listOptions   []metav1.ListOptions
		healthy       bool
		failed        bool
		neverRestart  bool
		client        bool
		expectedError error
	}{
		{
			name:          "healthy pod succeeds immediately",
			namespaces:    nil,
			listOptions:   nil,
			healthy:       true,
			failed:        false,
			client:        true,
			expectedError: nil,
		},
		{
			name:          "failed pod with never restart policy is skipped",
			namespaces:    nil,
			listOptions:   nil,
			healthy:       false,
			failed:        true,
			neverRestart:  true,
			client:        true,
			expectedError: nil,
		},
		{
			name:          "failed pod with default restart policy is skipped",
			namespaces:    nil,
			listOptions:   nil,
			healthy:       false,
			failed:        true,
			neverRestart:  false,
			client:        true,
			expectedError: nil,
		},
		{
			name:          "too many list options returns error",
			namespaces:    nil,
			listOptions:   []metav1.ListOptions{{}, {}},
			healthy:       true,
			failed:        false,
			client:        true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			name:          "nil api client returns error",
			namespaces:    nil,
			listOptions:   nil,
			healthy:       true,
			failed:        false,
			client:        false,
			expectedError: fmt.Errorf("podList 'apiClient' cannot be empty"),
		},
		{
			name:          "unhealthy pod times out",
			namespaces:    nil,
			listOptions:   nil,
			healthy:       false,
			failed:        false,
			client:        true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "namespace with no pods succeeds",
			namespaces:    []string{defaultPodNsName + "-has-no-pods"},
			listOptions:   nil,
			healthy:       false,
			failed:        false,
			client:        true,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var testSettings *clients.Settings

			if testCase.client {
				testPod := buildDummyPod(defaultPodName, defaultPodNsName, defaultPodImage)
				if testCase.healthy {
					testPod = buildDummyPodWithPhaseAndCondition(corev1.PodSucceeded, corev1.PodReady, false)
				} else if testCase.failed {
					testPod = buildDummyPodWithPhaseAndCondition(corev1.PodFailed, corev1.PodReady, testCase.neverRestart)
				}

				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects: []runtime.Object{testPod},
				})
			}

			err := WaitForPodsInNamespacesHealthy(testSettings, testCase.namespaces, time.Second, testCase.listOptions...)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}
