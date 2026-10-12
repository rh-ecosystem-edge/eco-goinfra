package oran

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListProvisioningRequests(t *testing.T) {
	testCases := []struct {
		provisioningRequests []*ProvisioningRequestBuilder
		listOptions          []runtimeclient.ListOptions
		client               bool
		expectedError        error
	}{
		{
			provisioningRequests: []*ProvisioningRequestBuilder{
				buildValidPRTestBuilder(buildTestClientWithDummyPR()),
			},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
		},
		{
			provisioningRequests: []*ProvisioningRequestBuilder{
				buildValidPRTestBuilder(buildTestClientWithDummyPR()),
			},
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:        true,
			expectedError: nil,
		},
		{
			provisioningRequests: []*ProvisioningRequestBuilder{
				buildValidPRTestBuilder(buildTestClientWithDummyPR()),
			},
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			client:        true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			provisioningRequests: []*ProvisioningRequestBuilder{
				buildValidPRTestBuilder(buildTestClientWithDummyPR()),
			},
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("failed to list provisioningRequests, 'apiClient' parameter is nil"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyPR()
		}

		builders, err := ListProvisioningRequests(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.provisioningRequests), len(builders))
		}
	}
}
