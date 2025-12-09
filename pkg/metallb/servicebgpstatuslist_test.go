package metallb

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestServiceBGPStatusList(t *testing.T) {
	testCases := []struct {
		Definition    *mlbtypes.ServiceBGPStatus
		statuses      []*ServiceBGPStatusBuilder
		listOptions   []metav1.ListOptions
		client        bool
		expectedError error
	}{
		{
			statuses: []*ServiceBGPStatusBuilder{buildValidServiceBGPStatusTestBuilder(
				buildTestServiceBGPStatusClientWithDummyStatus(defaultServiceBGPStatusName))},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
		},
		{
			statuses: []*ServiceBGPStatusBuilder{buildValidServiceBGPStatusTestBuilder(
				buildTestServiceBGPStatusClientWithDummyStatus(defaultServiceBGPStatusName))},
			listOptions:   []metav1.ListOptions{{LabelSelector: "test"}},
			client:        true,
			expectedError: nil,
		},
		{
			statuses: []*ServiceBGPStatusBuilder{buildValidServiceBGPStatusTestBuilder(
				buildTestServiceBGPStatusClientWithDummyStatus(defaultServiceBGPStatusName))},
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("failed to list ServiceBGPStatuses, 'apiClient' parameter is empty"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestServiceBGPStatusClientWithDummyStatus(defaultServiceBGPStatusName)
		}

		var (
			serviceBGPStatuses []*ServiceBGPStatusBuilder
			err                error
		)

		if len(testCase.listOptions) > 0 {
			clientOptions := client.ListOptions{
				Raw: &testCase.listOptions[0],
			}
			serviceBGPStatuses, err = ListServiceBGPStatus(testSettings, clientOptions)
		} else {
			serviceBGPStatuses, err = ListServiceBGPStatus(testSettings)
		}

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.statuses), len(serviceBGPStatuses))
		}
	}
}
