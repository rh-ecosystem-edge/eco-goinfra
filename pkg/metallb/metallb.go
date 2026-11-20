package metallb

import (
	"context"
	"fmt"
	"net"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	mlbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlboperator"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// Builder provides struct for the MetalLb object containing connection to
// the cluster and the MetalLb definitions.
type Builder struct {
	common.EmbeddableBuilder[mlbtypes.MetalLB, *mlbtypes.MetalLB]
	common.EmbeddableCreator[mlbtypes.MetalLB, Builder, *mlbtypes.MetalLB, *Builder]
	common.EmbeddableDeleteReturner[mlbtypes.MetalLB, Builder, *mlbtypes.MetalLB, *Builder]
	common.EmbeddableForceUpdater[mlbtypes.MetalLB, Builder, *mlbtypes.MetalLB, *Builder]
	common.EmbeddableWithOptions[mlbtypes.MetalLB, Builder, *mlbtypes.MetalLB, *Builder, AdditionalOptions]
}

// AdditionalOptions additional options for metallb object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the MetalLB GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return mlbtypes.GroupVersion.WithKind("MetalLB")
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(apiClient *clients.Settings, name, nsname string, nodeSelector map[string]string) *Builder {
	klog.V(100).Infof(
		"Initializing new metallb structure with the following params: %s, %s, %v",
		name, nsname, nodeSelector)

	builder := common.NewNamespacedBuilder[mlbtypes.MetalLB, Builder](
		apiClient, mlbtypes.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	if len(nodeSelector) < 1 {
		klog.V(100).Info("The SpeakerNodeSelector of the metallb is empty")

		builder.SetError(fmt.Errorf("metallb 'nodeSelector' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.SpeakerNodeSelector = nodeSelector

	return builder
}

// Pull retrieves an existing metallb.io object from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	klog.V(100).Infof(
		"Pulling metallb.io object name:%s in namespace: %s", name, nsname)

	return common.PullNamespacedBuilder[mlbtypes.MetalLB, Builder](
		context.TODO(), apiClient, mlbtypes.AddToScheme, name, nsname)
}

// RemoveLabel removes given label from metallb metadata.
func (builder *Builder) RemoveLabel(key string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Removing label %s from metalLbIo %s", key, builder.Definition.Name)

	if key == "" {
		klog.V(100).Infof("Failed to remove empty label's key from metalLbIo %s", builder.Definition.Name)
		builder.SetError(fmt.Errorf("error to remove empty key from metalLbIo"))

		return builder
	}

	delete(builder.Definition.Spec.SpeakerNodeSelector, key)

	return builder
}

// WithSpeakerNodeSelector adds the specified label to the MetalLbIo SpeakerNodeSelector.
func (builder *Builder) WithSpeakerNodeSelector(label map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding label selector %v to metallb.io object %s",
		label, builder.Definition.Name,
	)

	if len(label) < 1 {
		builder.SetError(fmt.Errorf("can not accept empty label and redefine metallb NodeSelector"))

		return builder
	}

	builder.Definition.Spec.SpeakerNodeSelector = label

	return builder
}

// WithFRRConfigAlwaysBlock adds specific routes to block from being advertised to the FRR nodes.
func (builder *Builder) WithFRRConfigAlwaysBlock(prefixes []string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding prefixes to block in the metallb.io object %s", builder.Definition.Name)

	if len(prefixes) < 1 {
		klog.V(100).Info("the frrConfiguration prefixes list is empty")

		builder.SetError(fmt.Errorf("can not accept empty prefix list for the metallb alwaysBlock mode"))

		return builder
	}

	// Validate CIDR prefixes
	for _, prefix := range prefixes {
		if _, _, err := net.ParseCIDR(prefix); err != nil {
			klog.V(100).Infof("the frrConfiguration prefix %s is not a valid CIDR", prefix)
			builder.SetError(fmt.Errorf("the prefix %s is not a valid CIDR", prefix))

			return builder
		}
	}

	builder.Definition.Spec.FRRK8SConfig = &mlbtypes.FRRK8SConfig{AlwaysBlock: prefixes}

	return builder
}

// GetMetalLbIoGVR returns metalLb's GroupVersionResource which could be used for Clean function.
func GetMetalLbIoGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: "metallb.io", Version: "v1beta1", Resource: "metallbs",
	}
}
