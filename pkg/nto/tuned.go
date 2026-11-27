package nto //nolint:misspell

import (
	"context"

	tunedv1 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/tuned/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var tunedGVK = tunedv1.SchemeGroupVersion.WithKind("Tuned")

// TunedBuilder provides a struct for Tuned object from the cluster and a Tuned definition.
type TunedBuilder struct {
	common.EmbeddableBuilder[tunedv1.Tuned, *tunedv1.Tuned]
	common.EmbeddableCreator[tunedv1.Tuned, TunedBuilder, *tunedv1.Tuned, *TunedBuilder]
	common.EmbeddableDeleteReturner[tunedv1.Tuned, TunedBuilder, *tunedv1.Tuned, *TunedBuilder]
	common.EmbeddableUpdater[tunedv1.Tuned, TunedBuilder, *tunedv1.Tuned, *TunedBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *TunedBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the Tuned GVK for this builder.
func (builder *TunedBuilder) GetGVK() schema.GroupVersionKind {
	return tunedGVK
}

// NewTunedBuilder creates a new instance of TunedBuilder.
func NewTunedBuilder(apiClient *clients.Settings, name, nsname string) *TunedBuilder {
	klog.V(100).Infof(
		"Initializing new Tuned structure with the following params: "+
			"name: %s, namespace: %s", name, nsname)

	return common.NewNamespacedBuilder[tunedv1.Tuned, TunedBuilder](
		apiClient, tunedv1.AddToScheme, name, nsname)
}

// PullTuned pulls existing Tuned from cluster.
func PullTuned(apiClient *clients.Settings, name, nsname string) (*TunedBuilder, error) {
	return common.PullNamespacedBuilder[tunedv1.Tuned, TunedBuilder](
		context.TODO(), apiClient, tunedv1.AddToScheme, name, nsname)
}

// WithProfile sets the tuned operator's profile.
func (builder *TunedBuilder) WithProfile(profile tunedv1.TunedProfile) *TunedBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting tuned %s in namespace %s with the Profile: %v",
		builder.Definition.Name, builder.Definition.Namespace, profile)

	builder.Definition.Spec.Profile = append(builder.Definition.Spec.Profile, profile)

	return builder
}

// WithRecommend sets the tuned operator's recommend.
func (builder *TunedBuilder) WithRecommend(recommend tunedv1.TunedRecommend) *TunedBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting tuned %s in namespace %s with the Recommend: %v",
		builder.Definition.Name, builder.Definition.Namespace, recommend)

	builder.Definition.Spec.Recommend = append(builder.Definition.Spec.Recommend, recommend)

	return builder
}
