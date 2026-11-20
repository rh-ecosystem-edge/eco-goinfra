package olm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	oplmV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	annotationAlmExamples = "alm-examples"
)

var clusterServiceVersionGVK = oplmV1alpha1.SchemeGroupVersion.WithKind("ClusterServiceVersion")

// ClusterServiceVersionBuilder provides a struct for clusterserviceversion object
// from the cluster and a clusterserviceversion definition.
type ClusterServiceVersionBuilder struct {
	common.EmbeddableBuilder[oplmV1alpha1.ClusterServiceVersion, *oplmV1alpha1.ClusterServiceVersion]
	common.EmbeddableDeleter[oplmV1alpha1.ClusterServiceVersion, *oplmV1alpha1.ClusterServiceVersion]
	common.EmbeddableUpdater[
		oplmV1alpha1.ClusterServiceVersion,
		ClusterServiceVersionBuilder,
		*oplmV1alpha1.ClusterServiceVersion,
		*ClusterServiceVersionBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ClusterServiceVersionBuilder) AttachMixins() {
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the ClusterServiceVersion GVK for this builder.
func (builder *ClusterServiceVersionBuilder) GetGVK() schema.GroupVersionKind {
	return clusterServiceVersionGVK
}

// PullClusterServiceVersion loads an existing clusterserviceversion into Builder struct.
func PullClusterServiceVersion(apiClient *clients.Settings, name, namespace string) (*ClusterServiceVersionBuilder,
	error) {
	return common.PullNamespacedBuilder[oplmV1alpha1.ClusterServiceVersion, ClusterServiceVersionBuilder](
		context.TODO(), apiClient, oplmV1alpha1.AddToScheme, name, namespace)
}

// GetAlmExamples extracts and returns the alm-examples block from the clusterserviceversion.
func (builder *ClusterServiceVersionBuilder) GetAlmExamples() (string, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Infof("Extracting the 'alm-examples' section from clusterserviceversion %s in "+
		"namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	almExamples := annotationAlmExamples

	if builder.Exists() {
		annotations := builder.Object.GetAnnotations()

		if example, ok := annotations[almExamples]; ok {
			return example, nil
		}
	}

	return "", fmt.Errorf("%s not found in given clusterserviceversion named %v",
		almExamples, builder.Definition.Name)
}

// IsSuccessful checks if the clusterserviceversion is Successful.
func (builder *ClusterServiceVersionBuilder) IsSuccessful() (bool, error) {
	if err := common.Validate(builder); err != nil {
		return false, err
	}

	klog.V(100).Infof("Verify clusterserviceversion %s in namespace %s is Successful",
		builder.Definition.Name, builder.Definition.Namespace)

	phase, err := builder.GetPhase()
	if err != nil {
		return false, fmt.Errorf("failed to get phase value for %s clusterserviceversion in %s namespace due to %w",
			builder.Definition.Name, builder.Definition.Namespace, err)
	}

	return phase == "Succeeded", nil
}

// GetPhase gets current clusterserviceversion phase.
func (builder *ClusterServiceVersionBuilder) GetPhase() (oplmV1alpha1.ClusterServiceVersionPhase, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Infof("Get clusterserviceversion %s phase in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return "", fmt.Errorf("%s clusterserviceversion not found in %s namespace",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return builder.Object.Status.Phase, nil
}
