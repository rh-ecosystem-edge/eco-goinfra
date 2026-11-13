package nodesconfig

import (
	"context"
	"fmt"

	configV1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// Builder provides a struct for nodesConfig object from the cluster and a nodesConfig definition.
type Builder struct {
	common.EmbeddableBuilder[configV1.Node, *configV1.Node]
	common.EmbeddableUpdater[configV1.Node, Builder, *configV1.Node, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the Node GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return configV1.GroupVersion.WithKind(APIKind)
}

// Pull retrieves an existing nodesConfig object from the cluster.
func Pull(apiClient *clients.Settings, nodesConfigObjName string) (*Builder, error) {
	return common.PullClusterScopedBuilder[configV1.Node, Builder](
		context.TODO(), apiClient, configV1.Install, nodesConfigObjName)
}

// GetCGroupMode fetches nodesConfig cgroupMode.
func (builder *Builder) GetCGroupMode() (configV1.CgroupMode, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Info("Getting nodesConfig cGroupMode configuration")

	if !builder.Exists() {
		return "", fmt.Errorf("nodesConfig object does not exist")
	}

	return builder.Object.Spec.CgroupMode, nil
}

// WithCGroupMode sets the nodesConfig operator's cgroupMode.
func (builder *Builder) WithCGroupMode(expectedCGroupMode configV1.CgroupMode) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting nodesConfig %s with cGroupMode: %v",
		builder.Definition.Name, expectedCGroupMode)

	if expectedCGroupMode == configV1.CgroupModeEmpty {
		klog.V(100).Info("the cGroupMode value can not be empty")

		builder.SetError(fmt.Errorf("the cGroupMode value can not be empty"))

		return builder
	}

	builder.Definition.Spec.CgroupMode = expectedCGroupMode

	return builder
}

// GetNodesConfigIoGVR returns nodesConfig's GroupVersionResource which could be used for Clean function.
func GetNodesConfigIoGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: APIKind,
	}
}
