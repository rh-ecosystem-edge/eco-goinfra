package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultServiceBGPStatusName = "servicebgpstatus-0"
	serviceBGPStatusTestSchemes = []clients.SchemeAttacher{mlbtypes.AddToScheme}
)

func TestServiceBGPStatusGet(t *testing.T) {
	var runtimeObjects []runtime.Object

	testCases := []struct {
		testServiceBGPStatus *ServiceBGPStatusBuilder
		addToRuntimeObjects  bool
		expectedError        string
		client               bool
	}{
		{
			testServiceBGPStatus: buildValidServiceBGPStatusTestBuilder(
				buildTestServiceBGPStatusClientWithDummyStatus(defaultServiceBGPStatusName)),
			expectedError: "",
		},
		{
			testServiceBGPStatus: buildValidServiceBGPStatusTestBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: serviceBGPStatusTestSchemes,
			})),
			expectedError: "servicebgpstatuses.metallb.io \"servicebgpstatus-0\" not found",
		},
	}

	for _, testCase := range testCases {
		serviceBGPStatus, err := testCase.testServiceBGPStatus.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, serviceBGPStatus.Name, testCase.testServiceBGPStatus.Definition.Name, serviceBGPStatus.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestServiceBGPStatusExist(t *testing.T) {
	testCases := []struct {
		testServiceBGPStatus *ServiceBGPStatusBuilder
		exist                bool
	}{
		{
			testServiceBGPStatus: buildValidServiceBGPStatusTestBuilder(
				buildTestServiceBGPStatusClientWithDummyStatus("test-status")),
			exist: false,
		},
		{
			testServiceBGPStatus: buildValidServiceBGPStatusTestBuilder(
				buildTestServiceBGPStatusClientWithDummyStatus(defaultServiceBGPStatusName)),
			exist: true,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testServiceBGPStatus.Exists()
		assert.Equal(t, testCase.exist, exist)
	}
}

func TestPullServiceBGPStatus(t *testing.T) {
	generateServiceBGPStatus := func(name string) *mlbtypes.ServiceBGPStatus {
		return &mlbtypes.ServiceBGPStatus{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Status: mlbtypes.MetalLBServiceBGPStatus{},
		}
	}

	testCases := []struct {
		name                string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			name:                "test1",
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			expectedError:       true,
			expectedErrorText:   "serviceBGPStatus 'name' cannot be empty",
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test1",
			expectedError:       true,
			expectedErrorText:   "serviceBGPStatus object test1 does not exist",
			addToRuntimeObjects: false,
			client:              true,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testServiceBGPStatus := generateServiceBGPStatus(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testServiceBGPStatus)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: serviceBGPStatusTestSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := PullServiceBGPStatus(testSettings, testCase.name)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testServiceBGPStatus.Name, builderResult.Object.Name)
		}
	}
}

// buildDummyServiceBGPStatus returns a ServiceBGPStatus with the provided name.
func buildDummyServiceBGPStatus(name string) *mlbtypes.ServiceBGPStatus {
	return &mlbtypes.ServiceBGPStatus{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestServiceBGPStatusClientWithDummyStatus returns a client with a dummy servicebgpstatus.
func buildTestServiceBGPStatusClientWithDummyStatus(statusName string) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyServiceBGPStatus(statusName)},
		SchemeAttachers: serviceBGPStatusTestSchemes,
	})
}

func buildValidServiceBGPStatusTestBuilder(apiClient *clients.Settings) *ServiceBGPStatusBuilder {
	return newServiceBGPStatusBuilder(apiClient, defaultServiceBGPStatusName)
}

func newServiceBGPStatusBuilder(apiClient *clients.Settings, name string) *ServiceBGPStatusBuilder {
	if apiClient == nil {
		return nil
	}

	builder := ServiceBGPStatusBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummyServiceBGPStatus(name),
	}

	return &builder
}
