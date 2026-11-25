package storage

import (
	"context"
	"fmt"

	noobaav1alpha1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var objectBucketClaimGVK = noobaav1alpha1.ObjectBucketClaimGVK()

// ObjectBucketClaimBuilder provides struct for the objectBucketClaim object.
type ObjectBucketClaimBuilder struct {
	common.EmbeddableBuilder[
		noobaav1alpha1.ObjectBucketClaim,
		*noobaav1alpha1.ObjectBucketClaim,
	]
	common.EmbeddableCreator[
		noobaav1alpha1.ObjectBucketClaim,
		ObjectBucketClaimBuilder,
		*noobaav1alpha1.ObjectBucketClaim,
		*ObjectBucketClaimBuilder,
	]
	common.EmbeddableDeleter[
		noobaav1alpha1.ObjectBucketClaim,
		*noobaav1alpha1.ObjectBucketClaim,
	]
	common.EmbeddableUpdater[
		noobaav1alpha1.ObjectBucketClaim,
		ObjectBucketClaimBuilder,
		*noobaav1alpha1.ObjectBucketClaim,
		*ObjectBucketClaimBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ObjectBucketClaimBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the ObjectBucketClaim GVK for this builder.
func (builder *ObjectBucketClaimBuilder) GetGVK() schema.GroupVersionKind {
	return objectBucketClaimGVK
}

// NewObjectBucketClaimBuilder creates new instance of builder.
func NewObjectBucketClaimBuilder(
	apiClient *clients.Settings, name, nsname string) *ObjectBucketClaimBuilder {
	return common.NewNamespacedBuilder[noobaav1alpha1.ObjectBucketClaim, ObjectBucketClaimBuilder](
		apiClient, noobaav1alpha1.AddToScheme, name, nsname)
}

// PullObjectBucketClaim retrieves an existing objectBucketClaim object from the cluster.
func PullObjectBucketClaim(apiClient *clients.Settings, name, nsname string) (*ObjectBucketClaimBuilder, error) {
	return common.PullNamespacedBuilder[noobaav1alpha1.ObjectBucketClaim, ObjectBucketClaimBuilder](
		context.TODO(), apiClient, noobaav1alpha1.AddToScheme, name, nsname)
}

// WithStorageClassName sets the objectBucketClaim operator's storageClassName configuration.
func (builder *ObjectBucketClaimBuilder) WithStorageClassName(
	storageClassName string) *ObjectBucketClaimBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting objectBucketClaim %s in namespace %s with the storageClassName config: %v",
		builder.Definition.Name, builder.Definition.Namespace, storageClassName)

	if storageClassName == "" {
		builder.SetError(fmt.Errorf("'storageClassName' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.StorageClassName = storageClassName

	return builder
}

// WithGenerateBucketName sets the objectBucketClaim operator's generateBucketName configuration.
func (builder *ObjectBucketClaimBuilder) WithGenerateBucketName(
	generateBucketName string) *ObjectBucketClaimBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting objectBucketClaim %s in namespace %s with the generateBucketName config: %v",
		builder.Definition.Name, builder.Definition.Namespace, generateBucketName)

	if generateBucketName == "" {
		builder.SetError(fmt.Errorf("'generateBucketName' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.GenerateBucketName = generateBucketName

	return builder
}
