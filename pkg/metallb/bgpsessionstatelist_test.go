package metallb

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBGPSessionStateList(t *testing.T) {
	testCases := []struct {
		Definition    *frrtypes.BGPSessionState
		states        []*BGPSessionStateBuilder
		listOptions   []metav1.ListOptions
		client        bool
		expectedError error
	}{
		{
			states: []*BGPSessionStateBuilder{buildValidBGPSessionStateTestBuilder(
				buildTestBGPSessionStateClientWithDummyState(defaultBGPSessionStateName))},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
		},
		{
			states: []*BGPSessionStateBuilder{buildValidBGPSessionStateTestBuilder(
				buildTestBGPSessionStateClientWithDummyState(defaultBGPSessionStateName))},
			listOptions:   []metav1.ListOptions{{LabelSelector: "test"}},
			client:        true,
			expectedError: nil,
		},
		{
			states: []*BGPSessionStateBuilder{buildValidBGPSessionStateTestBuilder(
				buildTestBGPSessionStateClientWithDummyState(defaultBGPSessionStateName))},
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("failed to list BGPSessionStates, 'apiClient' parameter is empty"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestBGPSessionStateClientWithDummyState(defaultBGPSessionStateName)
		}

		var (
			bgpSessionStates []*BGPSessionStateBuilder
			err              error
		)

		if len(testCase.listOptions) > 0 {
			clientOptions := client.ListOptions{
				Raw: &testCase.listOptions[0],
			}
			bgpSessionStates, err = ListBGPSessionState(testSettings, clientOptions)
		} else {
			bgpSessionStates, err = ListBGPSessionState(testSettings)
		}

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.states), len(bgpSessionStates))
		}
	}
}
