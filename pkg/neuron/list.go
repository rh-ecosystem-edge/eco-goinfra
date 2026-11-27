package neuron

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	neuronv1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/neuron/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListDeviceConfigs returns a list of DeviceConfig builders in the given namespace. An empty namespace lists
// DeviceConfigs in all namespaces.
func ListDeviceConfigs(
	apiClient *clients.Settings,
	nsname string,
	options ...runtimeclient.ListOptions) ([]*Builder, error) {
	var listOptions []runtimeclient.ListOption

	if len(options) > 0 {
		option := options[0]
		if nsname != "" {
			option.Namespace = nsname
		}

		listOptions = append(listOptions, &option)
	} else if nsname != "" {
		listOptions = append(listOptions, runtimeclient.InNamespace(nsname))
	}

	builders, err := common.List[neuronv1beta1.DeviceConfig, neuronv1beta1.DeviceConfigList, Builder](
		context.TODO(), apiClient, neuronv1beta1.AddToScheme, listOptions...)
	if err == nil && builders == nil {
		builders = []*Builder{}
	}

	return builders, err
}
