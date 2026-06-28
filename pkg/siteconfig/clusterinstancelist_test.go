package siteconfig

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListClusterInstances(t *testing.T) {
	testCases := []struct {
		clusterInstances []*CIBuilder
		listOptions      []runtimeclient.ListOptions
		client           bool
		expectedError    error
	}{
		{
			clusterInstances: []*CIBuilder{buildValidClusterInstanceTestBuilder(
				buildClusterInstanceClientWithDummyObject())},
			client:        true,
			expectedError: nil,
		},
		{
			clusterInstances: []*CIBuilder{buildValidClusterInstanceTestBuilder(
				buildClusterInstanceClientWithDummyObject())},
			listOptions:   []runtimeclient.ListOptions{{Continue: "test"}},
			client:        true,
			expectedError: nil,
		},
		{
			clusterInstances: []*CIBuilder{buildValidClusterInstanceTestBuilder(
				buildClusterInstanceClientWithDummyObject())},
			listOptions: []runtimeclient.ListOptions{
				{Namespace: "test"},
				{Continue: "true"},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			clusterInstances: []*CIBuilder{buildValidClusterInstanceTestBuilder(
				buildClusterInstanceClientWithDummyObject())},
			expectedError: fmt.Errorf("failed to list ClusterInstances, 'apiClient' parameter is nil"),
			client:        false,
		},
		{
			clusterInstances: nil,
			client:           true,
			expectedError:    nil,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			if testCase.clusterInstances == nil {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: testSchemes,
				})
			} else {
				testSettings = buildClusterInstanceClientWithDummyObject()
			}
		}

		builders, err := ListClusterInstances(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			expectedCount := 0
			if testCase.clusterInstances != nil {
				expectedCount = len(testCase.clusterInstances)
			}

			assert.Equal(t, expectedCount, len(builders))
		}
	}
}

func buildClusterInstanceClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{generateClusterInstance()},
		SchemeAttachers: testSchemes,
	})
}
