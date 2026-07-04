package oran

import (
	"context"

	hardwaremanagementv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// HardwareProfileBuilder provides a struct for the HardwareProfile resource containing
// a connection to the cluster and the HardwareProfile definition.
type HardwareProfileBuilder struct {
	common.EmbeddableBuilder[hardwaremanagementv1alpha1.HardwareProfile, *hardwaremanagementv1alpha1.HardwareProfile]
}

// GetGVK returns the HardwareProfile GVK for this builder.
func (builder *HardwareProfileBuilder) GetGVK() schema.GroupVersionKind {
	return hardwaremanagementv1alpha1.GroupVersion.WithKind("HardwareProfile")
}

// PullHardwareProfile fetches an existing HardwareProfile from the cluster by name and namespace.
func PullHardwareProfile(apiClient *clients.Settings, name, nsname string) (*HardwareProfileBuilder, error) {
	return common.PullNamespacedBuilder[hardwaremanagementv1alpha1.HardwareProfile, HardwareProfileBuilder](
		context.TODO(), apiClient, hardwaremanagementv1alpha1.AddToScheme, name, nsname)
}

// ListHardwareProfiles returns all HardwareProfile CRs across all namespaces.
func ListHardwareProfiles(apiClient *clients.Settings, options ...runtimeclient.ListOption) ([]*HardwareProfileBuilder, error) {
	return common.List[hardwaremanagementv1alpha1.HardwareProfile, hardwaremanagementv1alpha1.HardwareProfileList, HardwareProfileBuilder](
		context.TODO(), apiClient, hardwaremanagementv1alpha1.AddToScheme, options...)
}
