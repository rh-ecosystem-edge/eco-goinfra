package imagestream

import (
	"context"
	"fmt"

	imagev1 "github.com/openshift/api/image/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// Builder provides a struct for an imageStream object and its definition.
type Builder struct {
	common.EmbeddableBuilder[imagev1.ImageStream, *imagev1.ImageStream]
	common.EmbeddableDeleter[imagev1.ImageStream, *imagev1.ImageStream]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the ImageStream GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return imagev1.GroupVersion.WithKind("ImageStream")
}

// Pull retrieves an existing imageStream object from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	return common.PullNamespacedBuilder[imagev1.ImageStream, Builder](
		context.TODO(), apiClient, imagev1.AddToScheme, name, nsname)
}

// GetDockerImage fetches imageStream DockerImage value.
func (builder *Builder) GetDockerImage(imageTag string) (string, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Info("Getting imageStream DockerImage value")

	if imageTag == "" {
		klog.V(100).Info("The imageTag of the imageStream is empty")

		return "", fmt.Errorf("imageStream 'imageTag' cannot be empty")
	}

	if !builder.Exists() {
		return "", fmt.Errorf("imageStream object does not exist")
	}

	if len(builder.Object.Spec.Tags) == 0 {
		return "", fmt.Errorf("imageStream object %s in namespace %s has no tags",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	for _, tag := range builder.Object.Spec.Tags {
		if tag.From == nil {
			return "", fmt.Errorf("imageStream object %s in namespace %s has no DockerImage value",
				builder.Definition.Name, builder.Definition.Namespace)
		}

		if tag.Name == "" {
			return "", fmt.Errorf("imageStream object %s in namespace %s has no DockerImage tag value",
				builder.Definition.Name, builder.Definition.Namespace)
		}

		if tag.Name == imageTag {
			return tag.From.Name, nil
		}
	}

	return "", fmt.Errorf("image tag %s not found for imageStream object %s in namespace %s",
		imageTag, builder.Definition.Name, builder.Definition.Namespace)
}

// GetStatusTags returns the status tags from the imageStream as an array of tag names.
func (builder *Builder) GetStatusTags() ([]string, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	if !builder.Exists() {
		klog.V(100).Infof("ImageStream %s does not exist", builder.Definition.Name)

		return nil, fmt.Errorf("cannot get status tags from non-existent imageStream")
	}

	if len(builder.Object.Status.Tags) == 0 {
		return nil, fmt.Errorf("imageStream object %s in namespace %s has no status tags",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	var tagNames []string
	for _, tag := range builder.Object.Status.Tags {
		tagNames = append(tagNames, tag.Tag)
	}

	klog.V(100).Infof("Retrieved %d status tags from imageStream %s/%s: %v",
		len(tagNames), builder.Definition.Namespace, builder.Definition.Name, tagNames)

	return tagNames, nil
}
