package ocm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPoliciesInAllNamespaces(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		listOptions   []runtimeclient.ListOptions
		expectedError error
		client        bool
	}{
		{
			name:          "list all",
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			name:          "list with options",
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: nil,
			client:        true,
		},
		{
			name: "too many list options",
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			name:          "nil client",
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: fmt.Errorf("apiClient for Policy is nil"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				testSettings = buildTestClientWithDummyPolicy()
			}

			builders, err := ListPoliciesInAllNamespaces(testSettings, testCase.listOptions...)

			switch testCase.name {
			case "nil client":
				require.Error(t, err)
				assert.True(t, commonerrors.IsAPIClientNil(err))
				assert.Nil(t, builders)
			case "too many list options":
				assert.Equal(t, testCase.expectedError, err)
			default:
				assert.NoError(t, err)

				if len(testCase.listOptions) == 0 {
					assert.Len(t, builders, 1)
				}
			}
		})
	}
}

func TestWaitForAllPoliciesComplianceState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		compliant     bool
		client        bool
		listOptions   []runtimeclient.ListOptions
		expectedError error
	}{
		{
			name:      "all compliant",
			compliant: true,
			client:    true,
		},
		{
			name:          "not compliant",
			compliant:     false,
			client:        true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "nil client",
			client:        false,
			expectedError: fmt.Errorf("apiClient for Policy is nil"),
		},
		{
			name:      "too many list options",
			compliant: true,
			client:    true,
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				policy := buildDummyPolicy()

				if testCase.compliant {
					policy.Status.ComplianceState = policiesv1.Compliant
				}

				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects: []runtime.Object{policy},
					SchemeAttachers: []clients.SchemeAttacher{
						policiesv1.AddToScheme,
					},
				})
			}

			err := WaitForAllPoliciesComplianceState(
				testSettings, policiesv1.Compliant, time.Second, testCase.listOptions...)

			switch testCase.name {
			case "nil client":
				require.Error(t, err)
				assert.True(t, commonerrors.IsAPIClientNil(err))
			case "not compliant", "too many list options":
				assert.Equal(t, testCase.expectedError, err)
			default:
				assert.NoError(t, err)
			}
		})
	}
}
