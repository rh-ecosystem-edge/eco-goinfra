package console

import (
	"context"
	"fmt"
	"slices"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// ConsoleOperatorBuilder provides a struct for consoleOperator object from the cluster and a console definition.
type ConsoleOperatorBuilder struct {
	common.EmbeddableBuilder[operatorv1.Console, *operatorv1.Console]
	common.EmbeddableUpdater[operatorv1.Console, ConsoleOperatorBuilder, *operatorv1.Console, *ConsoleOperatorBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ConsoleOperatorBuilder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the Console operator GVK for this builder.
func (builder *ConsoleOperatorBuilder) GetGVK() schema.GroupVersionKind {
	return operatorv1.GroupVersion.WithKind("Console")
}

// PullConsoleOperator loads an existing consoleOperator into the ConsoleOperatorBuilder struct.
func PullConsoleOperator(apiClient *clients.Settings, consoleOperatorName string) (*ConsoleOperatorBuilder, error) {
	return common.PullClusterScopedBuilder[operatorv1.Console, ConsoleOperatorBuilder](
		context.TODO(), apiClient, operatorv1.Install, consoleOperatorName)
}

// GetPlugins fetches consoleOperator plugins list.
func (builder *ConsoleOperatorBuilder) GetPlugins() (*[]string, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Info("Getting consoleOperator plugins list configuration")

	if !builder.Exists() {
		return nil, fmt.Errorf("consoleOperator %s object does not exist", builder.Definition.Name)
	}

	return &builder.Object.Spec.Plugins, nil
}

// WithPlugins adds to the consoleOperator operator's new plugins.
func (builder *ConsoleOperatorBuilder) WithPlugins(newPluginsList []string, redefine bool) *ConsoleOperatorBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting consoleOperator %s with new plugins: %v",
		builder.Definition.Name, newPluginsList)

	if len(newPluginsList) == 0 {
		klog.V(100).Info("the newPluginsList can not be empty")

		builder.SetError(fmt.Errorf("the newPluginsList can not be empty"))

		return builder
	}

	if redefine {
		klog.V(100).Infof("Redefining existing plugins list with %v", newPluginsList)

		builder.Definition.Spec.Plugins = newPluginsList
	} else {
		klog.V(100).Info("Existing plugins list will not be redefined")

		for _, newPlugin := range newPluginsList {
			if slices.Contains(builder.Definition.Spec.Plugins, newPlugin) {
				klog.V(100).Infof("the newPlugin %s was already defined in Plugins list %v",
					newPlugin, builder.Definition.Spec.Plugins)
			} else {
				builder.Definition.Spec.Plugins = append(builder.Definition.Spec.Plugins, newPlugin)
			}
		}
	}

	return builder
}
