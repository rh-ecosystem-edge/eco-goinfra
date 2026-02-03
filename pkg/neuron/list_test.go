package neuron

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/neuron/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func buildTestDeviceConfig(name, namespace string) *v1beta1.DeviceConfig {
	return &v1beta1.DeviceConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: v1beta1.DeviceConfigSpec{
			DriversImage:      defaultDriversImage,
			DriverVersion:     defaultDriverVersion,
			DevicePluginImage: defaultDevicePluginImage,
		},
	}
}

func TestListDeviceConfigs(t *testing.T) {
	t.Parallel()

	t.Run("list multiple DeviceConfigs", func(t *testing.T) {
		runtimeObjects := []runtime.Object{
			buildTestDeviceConfig("neuron-1", "ai-operator-on-aws"),
			buildTestDeviceConfig("neuron-2", "ai-operator-on-aws"),
		}
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes,
		})

		builders, err := ListDeviceConfigs(testSettings, "ai-operator-on-aws")
		assert.Nil(t, err)
		assert.NotNil(t, builders)
		assert.Equal(t, 2, len(builders))
	})

	t.Run("list with namespace filter returns empty", func(t *testing.T) {
		runtimeObjects := []runtime.Object{buildTestDeviceConfig("neuron", "ai-operator-on-aws")}
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes,
		})

		builders, err := ListDeviceConfigs(testSettings, "other-namespace")
		assert.Nil(t, err)
		assert.NotNil(t, builders)
		assert.Equal(t, 0, len(builders))
	})

	t.Run("list empty namespace returns empty slice", func(t *testing.T) {
		testSettings := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})

		builders, err := ListDeviceConfigs(testSettings, "ai-operator-on-aws")
		assert.Nil(t, err)
		assert.NotNil(t, builders)
		assert.Equal(t, 0, len(builders))
	})

	t.Run("nil client returns error", func(t *testing.T) {
		builders, err := ListDeviceConfigs(nil, "")
		assert.Nil(t, builders)
		assert.Equal(t, fmt.Errorf("failed to list deviceConfigs, 'apiClient' parameter is empty").Error(), err.Error())
	})
}
