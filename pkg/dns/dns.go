package dns

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const clusterDNSName = "cluster"

// Builder provides a struct for the cluster DNS object and a DNS definition.
type Builder struct {
	common.EmbeddableBuilder[configv1.DNS, *configv1.DNS]
	common.EmbeddableUpdater[configv1.DNS, Builder, *configv1.DNS, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the DNS GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return configv1.GroupVersion.WithKind("DNS")
}

// Pull loads the existing cluster DNS into Builder.
func Pull(apiClient *clients.Settings) (*Builder, error) {
	return common.PullClusterScopedBuilder[configv1.DNS, Builder](
		context.TODO(), apiClient, configv1.Install, clusterDNSName)
}
