package nfd

import (
	"context"
	"fmt"

	nfdv1 "github.com/openshift/cluster-nfd-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
)

var nodeFeatureDiscoveryGVK = nfdv1.GroupVersion.WithKind("NodeFeatureDiscovery")

// Builder provides a struct for NodeFeatureDiscovery object
// from the cluster and a NodeFeatureDiscovery definition.
type Builder struct {
	common.EmbeddableBuilder[nfdv1.NodeFeatureDiscovery, *nfdv1.NodeFeatureDiscovery]
	common.EmbeddableCreator[
		nfdv1.NodeFeatureDiscovery, Builder, *nfdv1.NodeFeatureDiscovery, *Builder]
	common.EmbeddableDeleteReturner[
		nfdv1.NodeFeatureDiscovery, Builder, *nfdv1.NodeFeatureDiscovery, *Builder]
	common.EmbeddableForceUpdater[
		nfdv1.NodeFeatureDiscovery, Builder, *nfdv1.NodeFeatureDiscovery, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the NodeFeatureDiscovery GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return nodeFeatureDiscoveryGVK
}

// NewBuilderFromObjectString creates a Builder object from CSV alm-examples.
func NewBuilderFromObjectString(apiClient *clients.Settings, almExample string) *Builder {
	klog.V(100).Infof("Initializing new Builder structure from almExample string")

	if apiClient == nil {
		klog.V(100).Info("The apiClient of the NodeFeatureDiscovery is nil")

		return nil
	}

	if err := apiClient.AttachScheme(nfdv1.AddToScheme); err != nil {
		klog.V(100).Infof("Failed to add nfd v1 scheme to client schemes: %v", err)

		return nil
	}

	builder := &Builder{}
	builder.AttachMixins()
	builder.SetGVK(builder.GetGVK())
	builder.SetClient(apiClient)

	nodeFeatureDiscovery, err := getNodeFeatureDiscoveryFromAlmExample(almExample)
	if err != nil {
		klog.V(100).Infof(
			"Error initializing NodeFeatureDiscovery from alm-examples: %s", err.Error())

		builder.SetError(fmt.Errorf("error initializing NodeFeatureDiscovery from alm-examples: %w", err))

		return builder
	}

	builder.SetDefinition(nodeFeatureDiscovery)

	klog.V(100).Infof("Initializing Builder definition to NodeFeatureDiscovery object")

	if builder.Definition == nil {
		klog.V(100).Info("The NodeFeatureDiscovery object definition is nil")

		builder.SetError(fmt.Errorf("nodeFeatureDiscovery definition is nil"))
	}

	return builder
}

// Pull loads an existing NodeFeatureDiscovery into Builder struct.
func Pull(apiClient *clients.Settings, name, namespace string) (*Builder, error) {
	return common.PullNamespacedBuilder[nfdv1.NodeFeatureDiscovery, Builder](
		context.TODO(), apiClient, nfdv1.AddToScheme, name, namespace)
}

// getNodeFeatureDiscoveryFromAlmExample extracts the NodeFeatureDiscovery from the alm-examples block.
func getNodeFeatureDiscoveryFromAlmExample(almExample string) (*nfdv1.NodeFeatureDiscovery, error) {
	nodeFeatureDiscoveryList := &nfdv1.NodeFeatureDiscoveryList{}

	if almExample == "" {
		return nil, fmt.Errorf("almExample is an empty string")
	}

	err := json.Unmarshal([]byte(almExample), &nodeFeatureDiscoveryList.Items)
	if err != nil {
		return nil, err
	}

	if len(nodeFeatureDiscoveryList.Items) == 0 {
		return nil, fmt.Errorf("failed to get alm examples")
	}

	return &nodeFeatureDiscoveryList.Items[0], nil
}
