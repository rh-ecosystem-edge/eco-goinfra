package nvidiagpu

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	nvidiagpuv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/nvidiagpu/nvidiagputypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
)

var clusterPolicyGVK = nvidiagpuv1.SchemeGroupVersion.WithKind(nvidiagpuv1.ClusterPolicyCRDName)

// Builder provides a struct for ClusterPolicy object
// from the cluster and a ClusterPolicy definition.
type Builder struct {
	common.EmbeddableBuilder[nvidiagpuv1.ClusterPolicy, *nvidiagpuv1.ClusterPolicy]
	common.EmbeddableCreator[
		nvidiagpuv1.ClusterPolicy, Builder, *nvidiagpuv1.ClusterPolicy, *Builder]
	common.EmbeddableDeleteReturner[
		nvidiagpuv1.ClusterPolicy, Builder, *nvidiagpuv1.ClusterPolicy, *Builder]
	common.EmbeddableForceUpdater[
		nvidiagpuv1.ClusterPolicy, Builder, *nvidiagpuv1.ClusterPolicy, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the ClusterPolicy GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return clusterPolicyGVK
}

// NewBuilderFromObjectString creates a Builder object from CSV alm-examples.
func NewBuilderFromObjectString(apiClient *clients.Settings, almExample string) *Builder {
	klog.V(100).Infof("Initializing new Builder structure from almExample string")

	if apiClient == nil {
		klog.V(100).Info("The apiClient of the ClusterPolicy is nil")

		return nil
	}

	if err := apiClient.AttachScheme(nvidiagpuv1.AddToScheme); err != nil {
		klog.V(100).Infof("Failed to add nvidiagpuv1 scheme to client schemes: %v", err)

		return nil
	}

	builder := &Builder{}
	builder.AttachMixins()
	builder.SetGVK(builder.GetGVK())
	builder.SetClient(apiClient)

	clusterPolicy, err := getClusterPolicyFromAlmExample(almExample)
	if err != nil {
		klog.V(100).Infof("error initializing ClusterPolicy from alm-examples: %s", err.Error())

		builder.SetError(fmt.Errorf("error initializing ClusterPolicy from alm-examples: %w", err))

		return builder
	}

	builder.SetDefinition(clusterPolicy)

	klog.V(100).Infof(
		"Initializing new Builder structure from almExample string with clusterPolicy name: %s",
		clusterPolicy.Name)

	if builder.Definition == nil {
		klog.V(100).Info("The ClusterPolicy object definition is nil")

		builder.SetError(fmt.Errorf("ClusterPolicy 'Object.Definition' is nil"))
	}

	return builder
}

// Pull loads an existing clusterPolicy into Builder struct.
func Pull(apiClient *clients.Settings, name string) (*Builder, error) {
	return common.PullClusterScopedBuilder[nvidiagpuv1.ClusterPolicy, Builder](
		context.TODO(), apiClient, nvidiagpuv1.AddToScheme, name)
}

// getClusterPolicyFromAlmExample extracts the ClusterPolicy from the alm-examples block.
func getClusterPolicyFromAlmExample(almExample string) (*nvidiagpuv1.ClusterPolicy, error) {
	clusterPolicyList := &nvidiagpuv1.ClusterPolicyList{}

	if almExample == "" {
		return nil, fmt.Errorf("almExample is an empty string")
	}

	err := json.Unmarshal([]byte(almExample), &clusterPolicyList.Items)
	if err != nil {
		return nil, err
	}

	if len(clusterPolicyList.Items) == 0 {
		return nil, fmt.Errorf("failed to get alm examples")
	}

	return &clusterPolicyList.Items[0], nil
}
