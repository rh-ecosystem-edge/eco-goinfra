package pfstatus

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	pfstatustypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/pfstatus/pfstatustypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const pfstatusconfigurationPollingintervalValueIsNotValid = "pfStatusConfiguration 'pollingInterval' value is not valid"

// PfStatusConfigurationBuilder provides struct for the PfStatusConfiguration object containing connection to
// the cluster and the PFLACPMonitor definitions.
type PfStatusConfigurationBuilder struct {
	common.EmbeddableBuilder[pfstatustypes.PFLACPMonitor, *pfstatustypes.PFLACPMonitor]
	common.EmbeddableCreator[
		pfstatustypes.PFLACPMonitor, PfStatusConfigurationBuilder, *pfstatustypes.PFLACPMonitor,
		*PfStatusConfigurationBuilder]
	common.EmbeddableDeleter[pfstatustypes.PFLACPMonitor, *pfstatustypes.PFLACPMonitor]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *PfStatusConfigurationBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the PFLACPMonitor GVK for this builder.
func (builder *PfStatusConfigurationBuilder) GetGVK() schema.GroupVersionKind {
	return pfstatustypes.GroupVersion.WithKind("PFLACPMonitor")
}

// NewPfStatusConfigurationBuilder creates a new instance of PfStatusConfiguration.
func NewPfStatusConfigurationBuilder(
	apiClient *clients.Settings, name, nsname string) *PfStatusConfigurationBuilder {
	klog.V(100).Infof(
		"Initializing new NewPfStatusConfiguration structure with the following params: %s, %s",
		name, nsname)

	return common.NewNamespacedBuilder[pfstatustypes.PFLACPMonitor, PfStatusConfigurationBuilder](
		apiClient, pfstatustypes.AddToScheme, name, nsname)
}

// PullPfStatusConfiguration pulls existing pfStatusConfiguration from cluster.
func PullPfStatusConfiguration(
	apiClient *clients.Settings, name, nsname string) (*PfStatusConfigurationBuilder, error) {
	klog.V(100).Infof(
		"Pulling existing pfStatusConfiguration name %s under namespace %s from cluster", name, nsname)

	return common.PullNamespacedBuilder[pfstatustypes.PFLACPMonitor, PfStatusConfigurationBuilder](
		context.TODO(), apiClient, pfstatustypes.AddToScheme, name, nsname)
}

// WithNodeSelector defines the nodeSelector placed in the PfStatusConfiguration spec.
func (builder *PfStatusConfigurationBuilder) WithNodeSelector(
	nodeSelector map[string]string) *PfStatusConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating pfStatusConfiguration %s in namespace %s with this nodeSelector: %s",
		builder.Definition.Name, builder.Definition.Namespace, nodeSelector)

	if len(nodeSelector) == 0 {
		klog.V(100).Info("Can not redefine pfStatusConfiguration with empty nodeSelector map")

		builder.SetError(fmt.Errorf("pfStatusConfiguration 'nodeSelector' cannot be empty map"))

		return builder
	}

	builder.Definition.Spec.NodeSelector = nodeSelector

	return builder
}

// WithInterface defines the interface to be used in the PfStatusConfiguration.
func (builder *PfStatusConfigurationBuilder) WithInterface(interfaceName string) *PfStatusConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating pfStatusConfiguration %s in namespace %s with interface: %s",
		builder.Definition.Name, builder.Definition.Namespace, interfaceName)

	if interfaceName == "" {
		klog.V(100).Info("Can not redefine pfStatusConfiguration with empty interface string")

		builder.SetError(fmt.Errorf("interface can not be empty string"))

		return builder
	}

	// Append the new unique interface
	builder.Definition.Spec.Interfaces = append(builder.Definition.Spec.Interfaces, interfaceName)

	return builder
}

// WithPollingInterval defines the polling time between LACP messages.
func (builder *PfStatusConfigurationBuilder) WithPollingInterval(pollingInterval int) *PfStatusConfigurationBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating pfStatusConfiguration %s in namespace %s with this pollingInterval: %d",
		builder.Definition.Name, builder.Definition.Namespace, pollingInterval)

	if pollingInterval < 100 || pollingInterval > 65535 {
		klog.V(100).Info("A valid polling interval is between 100-65535")

		builder.SetError(fmt.Errorf("%s", pfstatusconfigurationPollingintervalValueIsNotValid))

		return builder
	}

	builder.Definition.Spec.PollingInterval = pollingInterval

	return builder
}

// GetPfStatusConfigurationGVR returns PfStatusConfiguration GroupVersionResource.
func GetPfStatusConfigurationGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: "pfstatusrelay.openshift.io", Version: "v1alpha1", Resource: "pflacpmonitors",
	}
}
