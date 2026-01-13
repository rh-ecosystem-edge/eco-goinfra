package neuron

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/neuron/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestListDeviceConfigs(t *testing.T) {
	testCases := []struct {
		deviceConfigs       []*v1alpha1.DeviceConfig
		namespace           string
		expectedCount       int
		expectedError       error
		addToRuntimeObjects bool
		client              bool
	}{
		{
			deviceConfigs: []*v1alpha1.DeviceConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "neuron-1",
						Namespace: "ai-operator-on-aws",
					},
					Spec: v1alpha1.DeviceConfigSpec{
						DriversImage:      defaultDriversImage,
						DriverVersion:     defaultDriverVersion,
						DevicePluginImage: defaultDevicePluginImage,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "neuron-2",
						Namespace: "ai-operator-on-aws",
					},
					Spec: v1alpha1.DeviceConfigSpec{
						DriversImage:      defaultDriversImage,
						DriverVersion:     defaultDriverVersion,
						DevicePluginImage: defaultDevicePluginImage,
					},
				},
			},
			namespace:           "ai-operator-on-aws",
			expectedCount:       2,
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			deviceConfigs: []*v1alpha1.DeviceConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "neuron",
						Namespace: "ai-operator-on-aws",
					},
					Spec: v1alpha1.DeviceConfigSpec{
						DriversImage:      defaultDriversImage,
						DriverVersion:     defaultDriverVersion,
						DevicePluginImage: defaultDevicePluginImage,
					},
				},
			},
			namespace:           "other-namespace",
			expectedCount:       0,
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			deviceConfigs:       []*v1alpha1.DeviceConfig{},
			namespace:           "ai-operator-on-aws",
			expectedCount:       0,
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			deviceConfigs:       []*v1alpha1.DeviceConfig{},
			namespace:           "",
			expectedCount:       0,
			expectedError:       fmt.Errorf("failed to list deviceConfigs, 'apiClient' parameter is empty"),
			addToRuntimeObjects: true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			for _, dc := range testCase.deviceConfigs {
				runtimeObjects = append(runtimeObjects, dc)
			}
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builders, err := ListDeviceConfigs(testSettings, testCase.namespace)

		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.NotNil(t, builders)
			assert.Equal(t, testCase.expectedCount, len(builders))
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}
