package ingress

import (
	"context"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var ingressControllerGVK = operatorv1.GroupVersion.WithKind("IngressController")

// Builder provides a struct for an ingresscontroller object from the cluster and an ingresscontroller definition.
type Builder struct {
	common.EmbeddableBuilder[operatorv1.IngressController, *operatorv1.IngressController]
	common.EmbeddableCreator[operatorv1.IngressController, Builder, *operatorv1.IngressController, *Builder]
	common.EmbeddableDeleter[operatorv1.IngressController, *operatorv1.IngressController]
	common.EmbeddableUpdater[operatorv1.IngressController, Builder, *operatorv1.IngressController, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the IngressController GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return ingressControllerGVK
}

// Pull loads an existing ingresscontroller into Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	return common.PullNamespacedBuilder[operatorv1.IngressController, Builder](
		context.TODO(), apiClient, operatorv1.Install, name, nsname)
}
