package ocm

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
	"k8s.io/apimachinery/pkg/runtime"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
)

func TestListPoliciesInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListPoliciesInAllNamespaces,
		policiesv1.AddToScheme,
		policyGVK,
	).ExecuteTests(t)
}

func TestWaitForAllPoliciesComplianceState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		compliant     bool
		client        bool
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
				testSettings, policiesv1.Compliant, time.Second)

			switch testCase.name {
			case "nil client":
				require.Error(t, err)
				assert.True(t, commonerrors.IsAPIClientNil(err))
			case "not compliant":
				assert.Equal(t, testCase.expectedError, err)
			default:
				assert.NoError(t, err)
			}
		})
	}
}
