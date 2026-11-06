package resourcequotas

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var resourceQuotaGVK = corev1.SchemeGroupVersion.WithKind("ResourceQuota")

// Builder provides struct for resource quotas containing connection to the cluster.
type Builder struct {
	common.EmbeddableBuilder[corev1.ResourceQuota, *corev1.ResourceQuota]
	common.EmbeddableCreator[corev1.ResourceQuota, Builder, *corev1.ResourceQuota, *Builder]
	common.EmbeddableDeleter[corev1.ResourceQuota, *corev1.ResourceQuota]
	common.EmbeddableForceUpdater[corev1.ResourceQuota, Builder, *corev1.ResourceQuota, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the ResourceQuota GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return resourceQuotaGVK
}

// NewBuilder creates a new resource quota builder.
func NewBuilder(apiClient *clients.Settings, name, nsname string) *Builder {
	klog.V(100).Infof("Initializing new resource quota structure with the following params: "+
		"name=%s, namespace=%s", name, nsname)

	return common.NewNamespacedBuilder[corev1.ResourceQuota, Builder](
		apiClient, corev1.AddToScheme, name, nsname)
}

// Pull retrieves the resource quota from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	klog.V(100).Infof("Pulling resource quota with the following params: name=%s, namespace=%s", name, nsname)

	return common.PullNamespacedBuilder[corev1.ResourceQuota, Builder](
		context.TODO(), apiClient, corev1.AddToScheme, name, nsname)
}

// WithQuotaSpec sets the resource quota spec.
func (builder *Builder) WithQuotaSpec(quotaSpec corev1.ResourceQuotaSpec) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting resource quota spec for %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec = quotaSpec

	return builder
}

// GetGVR returns resource quotas's GroupVersionResource which could be used for Clean function.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "resourcequotas",
	}
}
