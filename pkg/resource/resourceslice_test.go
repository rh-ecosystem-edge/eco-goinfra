package resource

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestListResourceSlices(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
		expectedCount       int
		client              bool
		expectedError       error
	}{
		{
			addToRuntimeObjects: true,
			expectedCount:       1,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: false,
			expectedCount:       0,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: true,
			expectedCount:       0,
			client:              false,
			expectedError:       fmt.Errorf("resourceSlice 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects,
				generateResourceSlice("test-slice", "test-driver"))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testResourceSchemes,
			})
		}

		slices, err := ListResourceSlices(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Len(t, slices, testCase.expectedCount)
		}
	}
}

func TestListResourceSlicesByDriver(t *testing.T) {
	testCases := []struct {
		driverName    string
		addSlices     bool
		client        bool
		expectedCount int
		expectedError error
	}{
		{
			driverName:    "test-driver",
			addSlices:     true,
			client:        true,
			expectedCount: 1,
			expectedError: nil,
		},
		{
			driverName:    "other-driver",
			addSlices:     true,
			client:        true,
			expectedCount: 0,
			expectedError: nil,
		},
		{
			driverName:    "",
			addSlices:     false,
			client:        true,
			expectedCount: 0,
			expectedError: fmt.Errorf("resourceSlice 'driverName' cannot be empty"),
		},
		{
			driverName:    "test-driver",
			addSlices:     false,
			client:        false,
			expectedCount: 0,
			expectedError: fmt.Errorf("resourceSlice 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addSlices {
			runtimeObjects = append(runtimeObjects,
				generateResourceSlice("test-slice", "test-driver"))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testResourceSchemes,
			})
		}

		slices, err := ListResourceSlicesByDriver(testSettings, testCase.driverName)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Len(t, slices, testCase.expectedCount)
		}
	}
}

func generateResourceSlice(name, driverName string) *resourcev1.ResourceSlice {
	return &resourcev1.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: resourcev1.ResourceSliceSpec{
			Driver: driverName,
		},
	}
}
