package lso

import (
	"context"
	"fmt"
	"time"

	lsov1alpha1 "github.com/openshift/local-storage-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	errEmptyTolerations = "'tolerations' argument cannot be empty"
)

var localVolumeDiscoveryGVK = lsov1alpha1.GroupVersion.WithKind("LocalVolumeDiscovery")

// LocalVolumeDiscoveryBuilder provides a struct for localVolumeDiscovery object from the cluster
// and a localVolumeDiscovery definition.
type LocalVolumeDiscoveryBuilder struct {
	common.EmbeddableBuilder[lsov1alpha1.LocalVolumeDiscovery, *lsov1alpha1.LocalVolumeDiscovery]
	common.EmbeddableCreator[
		lsov1alpha1.LocalVolumeDiscovery, LocalVolumeDiscoveryBuilder,
		*lsov1alpha1.LocalVolumeDiscovery, *LocalVolumeDiscoveryBuilder,
	]
	common.EmbeddableDeleter[lsov1alpha1.LocalVolumeDiscovery, *lsov1alpha1.LocalVolumeDiscovery]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *LocalVolumeDiscoveryBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the LocalVolumeDiscovery GVK for this builder.
func (builder *LocalVolumeDiscoveryBuilder) GetGVK() schema.GroupVersionKind {
	return localVolumeDiscoveryGVK
}

// NewLocalVolumeDiscoveryBuilder creates new instance of LocalVolumeDiscoveryBuilder.
func NewLocalVolumeDiscoveryBuilder(apiClient *clients.Settings, name, nsname string) *LocalVolumeDiscoveryBuilder {
	return common.NewNamespacedBuilder[lsov1alpha1.LocalVolumeDiscovery, LocalVolumeDiscoveryBuilder](
		apiClient, lsov1alpha1.AddToScheme, name, nsname)
}

// PullLocalVolumeDiscovery retrieves an existing localVolumeDiscovery object from the cluster.
func PullLocalVolumeDiscovery(apiClient *clients.Settings, name, nsname string) (*LocalVolumeDiscoveryBuilder, error) {
	return common.PullNamespacedBuilder[lsov1alpha1.LocalVolumeDiscovery, LocalVolumeDiscoveryBuilder](
		context.TODO(), apiClient, lsov1alpha1.AddToScheme, name, nsname)
}

// IsDiscovering check if the localVolumeDiscovery is Discovering.
func (builder *LocalVolumeDiscoveryBuilder) IsDiscovering(timeout time.Duration) bool {
	if err := common.Validate(builder); err != nil {
		return false
	}

	klog.V(100).Infof("Verify localVolumeDiscovery %s in namespace %s is in Discovering phase",
		builder.Definition.Name, builder.Definition.Namespace)

	err := wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			phase, err := builder.GetPhase()
			if err != nil {
				klog.V(100).Infof("failed to get phase value for localVolumeDiscovery %s in namespace %s due to %v",
					builder.Definition.Name, builder.Definition.Namespace, err)

				return false, nil
			}

			return phase == lsov1alpha1.Discovering, nil
		})
	if err != nil {
		klog.V(100).Infof("localVolumeDiscovery %s in namespace %s is found not in the discovering state; %v",
			builder.Definition.Name, builder.Definition.Namespace, err)

		return false
	}

	return true
}

// GetPhase get current localVolumeDiscovery phase.
func (builder *LocalVolumeDiscoveryBuilder) GetPhase() (lsov1alpha1.DiscoveryPhase, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Infof("Get %s localVolumeDiscovery in %s namespace phase",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return "", fmt.Errorf("%s localVolumeDiscovery not found in %s namespace",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return builder.Object.Status.Phase, nil
}

// WithNodeSelector sets the localVolumeDiscovery's nodeSelector.
func (builder *LocalVolumeDiscoveryBuilder) WithNodeSelector(
	nodeSelector corev1.NodeSelector) *LocalVolumeDiscoveryBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding nodeSelector %v to localVolumeDiscovery %s in namespace %s",
		nodeSelector, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.NodeSelector = &nodeSelector

	return builder
}

// WithTolerations sets the localVolumeDiscovery's generation.
func (builder *LocalVolumeDiscoveryBuilder) WithTolerations(
	tolerations []corev1.Toleration) *LocalVolumeDiscoveryBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding tolerations %v to localVolumeDiscovery %s in namespace %s",
		tolerations, builder.Definition.Name, builder.Definition.Namespace)

	if len(tolerations) == 0 {
		klog.V(100).Info("The tolerations list is empty")

		builder.SetError(fmt.Errorf(errEmptyTolerations))

		return builder
	}

	builder.Definition.Spec.Tolerations = tolerations

	return builder
}
