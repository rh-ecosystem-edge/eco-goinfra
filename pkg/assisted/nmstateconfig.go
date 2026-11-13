package assisted

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	assistedv1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var nmStateConfigGVK = assistedv1beta1.GroupVersion.WithKind("NMStateConfig")

// NmStateConfigBuilder provides struct for the NMStateConfig object containing connection to
// the cluster and the NMStateConfig definitions.
type NmStateConfigBuilder struct {
	common.EmbeddableBuilder[assistedv1beta1.NMStateConfig, *assistedv1beta1.NMStateConfig]
	common.EmbeddableCreator[assistedv1beta1.NMStateConfig, NmStateConfigBuilder,
		*assistedv1beta1.NMStateConfig, *NmStateConfigBuilder]
	common.EmbeddableDeleter[assistedv1beta1.NMStateConfig, *assistedv1beta1.NMStateConfig]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *NmStateConfigBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the NMStateConfig GVK for this builder.
func (builder *NmStateConfigBuilder) GetGVK() schema.GroupVersionKind {
	return nmStateConfigGVK
}

// NewNmStateConfigBuilder creates a new instance of NMStateConfig Builder.
func NewNmStateConfigBuilder(apiClient *clients.Settings, name, namespace string) *NmStateConfigBuilder {
	klog.V(100).Infof("Initializing new nmstateconfig structure with the name: %s in namespace: %s", name, namespace)

	return common.NewNamespacedBuilder[assistedv1beta1.NMStateConfig, NmStateConfigBuilder](
		apiClient, assistedv1beta1.AddToScheme, name, namespace)
}

// ListNmStateConfigsInAllNamespaces returns a cluster-wide NMStateConfig list.
func ListNmStateConfigsInAllNamespaces(apiClient *clients.Settings) ([]*NmStateConfigBuilder, error) {
	return common.List[assistedv1beta1.NMStateConfig, assistedv1beta1.NMStateConfigList, NmStateConfigBuilder](
		context.TODO(), apiClient, assistedv1beta1.AddToScheme)
}

// ListNmStateConfigs returns a NMStateConfig list in a given namespace.
func ListNmStateConfigs(apiClient *clients.Settings, namespace string) ([]*NmStateConfigBuilder, error) {
	if namespace == "" {
		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("NMStateConfig", "", ""), commonerrors.BuilderFieldNamespace)
	}

	return common.List[assistedv1beta1.NMStateConfig, assistedv1beta1.NMStateConfigList, NmStateConfigBuilder](
		context.TODO(), apiClient, assistedv1beta1.AddToScheme, runtimeclient.InNamespace(namespace))
}
