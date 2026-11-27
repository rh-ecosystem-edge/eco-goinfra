package events

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Builder provides struct for Event object which contains connection to cluster.
type Builder struct {
	common.EmbeddableBuilder[corev1.Event, *corev1.Event]
}

// GetGVK returns the core Event GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("Event")
}

// Pull pulls existing Event from cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	return common.PullNamespacedBuilder[corev1.Event, Builder](
		context.TODO(), apiClient, corev1.AddToScheme, name, nsname)
}
