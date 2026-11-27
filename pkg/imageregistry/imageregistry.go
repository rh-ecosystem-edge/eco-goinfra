package imageregistry

import (
	"context"
	"fmt"
	"strings"
	"time"

	imageregistryv1 "github.com/openshift/api/imageregistry/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var imageRegistryGVK = schema.GroupVersion{Group: APIGroup, Version: APIVersion}.WithKind(APIKind)

// Builder provides a struct for imageRegistry object from the cluster and a imageRegistry definition.
type Builder struct {
	common.EmbeddableBuilder[imageregistryv1.Config, *imageregistryv1.Config]
	common.EmbeddableUpdater[imageregistryv1.Config, Builder, *imageregistryv1.Config, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the imageRegistry Config GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return imageRegistryGVK
}

// Pull retrieves an existing imageRegistry object from the cluster.
func Pull(apiClient *clients.Settings, imageRegistryObjName string) (*Builder, error) {
	return common.PullClusterScopedBuilder[imageregistryv1.Config, Builder](
		context.TODO(), apiClient, imageregistryv1.Install, imageRegistryObjName)
}

// GetManagementState fetches imageRegistry ManagementState.
func (builder *Builder) GetManagementState() (*operatorv1.ManagementState, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Info("Getting imageRegistry ManagementState configuration")

	if !builder.Exists() {
		return nil, fmt.Errorf("imageRegistry object does not exist")
	}

	return &builder.Object.Spec.ManagementState, nil
}

// GetStorageConfig fetches imageRegistry Storage configuration.
func (builder *Builder) GetStorageConfig() (*imageregistryv1.ImageRegistryConfigStorage, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Info("Getting imageRegistry Storage configuration")

	if !builder.Exists() {
		return nil, fmt.Errorf("imageRegistry object does not exist")
	}

	return &builder.Object.Spec.Storage, nil
}

// WithManagementState sets the imageRegistry operator's management state.
func (builder *Builder) WithManagementState(expectedManagementState operatorv1.ManagementState) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting imageRegistry %s with ManagementState: %v",
		builder.Definition.Name, expectedManagementState)

	builder.Definition.Spec.ManagementState = expectedManagementState

	return builder
}

// WithStorage sets the imageRegistry operator's storage.
func (builder *Builder) WithStorage(expectedStorage imageregistryv1.ImageRegistryConfigStorage) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting imageRegistry %s with Storage: %v",
		builder.Definition.Name, expectedStorage)

	builder.Definition.Spec.Storage = expectedStorage

	return builder
}

// WaitForCondition waits until the imageRegistry has a condition that matches the expected, checking only the Type,
// Status, Reason, and Message fields. For the messages field, it matches if the message contains the expected. Zero
// value fields in the expected condition are ignored.
func (builder *Builder) WaitForCondition(
	expected operatorv1.OperatorCondition, timeout time.Duration) (*Builder, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Waiting until condition of imageRegistry %s matches %v", builder.Definition.Name, expected)

	if !builder.Exists() {
		return nil, fmt.Errorf("imageRegistry object %s does not exist", builder.Definition.Name)
	}

	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if expected.Type != "" && condition.Type != expected.Type {
					continue
				}

				if expected.Status != "" && condition.Status != expected.Status {
					continue
				}

				if expected.Reason != "" && condition.Reason != expected.Reason {
					continue
				}

				if expected.Message != "" && !strings.Contains(condition.Message, expected.Message) {
					continue
				}

				return true, nil
			}

			return false, nil
		})
	if err != nil {
		return nil, err
	}

	return builder, nil
}
