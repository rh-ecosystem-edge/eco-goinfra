package ibi

import (
	"context"
	"fmt"
	"net"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	ibiv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/imagebasedinstall/api/hiveextensions/v1alpha1"
	hivev1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/imagebasedinstall/hive/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errInvalidMachineNetworkFormat = "imageclusterinstall machinenetwork incorrectly formatted"
)

// ImageClusterInstallBuilder provides struct for the imageclusterinstall object containing connection to
// the cluster and the imageclusterinstall definitions.
type ImageClusterInstallBuilder struct {
	common.EmbeddableBuilder[ibiv1alpha1.ImageClusterInstall, *ibiv1alpha1.ImageClusterInstall]
	common.EmbeddableCreator[
		ibiv1alpha1.ImageClusterInstall, ImageClusterInstallBuilder,
		*ibiv1alpha1.ImageClusterInstall, *ImageClusterInstallBuilder]
	common.EmbeddableDeleter[ibiv1alpha1.ImageClusterInstall, *ibiv1alpha1.ImageClusterInstall]
	common.EmbeddableForceUpdater[
		ibiv1alpha1.ImageClusterInstall, ImageClusterInstallBuilder,
		*ibiv1alpha1.ImageClusterInstall, *ImageClusterInstallBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ImageClusterInstallBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the ImageClusterInstall GVK for this builder.
func (builder *ImageClusterInstallBuilder) GetGVK() schema.GroupVersionKind {
	return ibiv1alpha1.GroupVersion.WithKind("ImageClusterInstall")
}

// NewImageClusterInstallBuilder creates a new instance of ImageClusterInstallBuilder.
func NewImageClusterInstallBuilder(
	apiClient *clients.Settings, name, nsname, imageset string) *ImageClusterInstallBuilder {
	klog.V(100).Infof(
		"Initializing new imageclusterinstall structure with the following params: "+
			"name: %s, namespace: %s, imageset: %s",
		name, nsname, imageset)

	builder := common.NewNamespacedBuilder[ibiv1alpha1.ImageClusterInstall, ImageClusterInstallBuilder](
		apiClient, ibiv1alpha1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	if imageset == "" {
		klog.V(100).Info("The imageset of the imageclusterinstall is empty")

		builder.SetError(fmt.Errorf("imageclusterinstall 'imageset' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.ImageSetRef = hivev1.ClusterImageSetReference{Name: imageset}

	return builder
}

// PullImageClusterInstall retrieves an existing imageclusterinstall from the cluster.
func PullImageClusterInstall(apiClient *clients.Settings, name, nsname string) (*ImageClusterInstallBuilder, error) {
	return common.PullNamespacedBuilder[ibiv1alpha1.ImageClusterInstall, ImageClusterInstallBuilder](
		context.TODO(), apiClient, ibiv1alpha1.AddToScheme, name, nsname)
}

// WithHostname sets hostname of installed node.
func (builder *ImageClusterInstallBuilder) WithHostname(hostname string) *ImageClusterInstallBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if hostname == "" {
		klog.V(100).Info("The imageclusterinstall hostname is empty")

		builder.SetError(fmt.Errorf("imageclusterinstall hostname cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Hostname = hostname

	return builder
}

// WithClusterDeployment links imageclusterinstall to an existing cluster deployment.
func (builder *ImageClusterInstallBuilder) WithClusterDeployment(
	clusterDeploymentName string) *ImageClusterInstallBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if clusterDeploymentName == "" {
		klog.V(100).Info("The imageclusterinstall clusterdeployment is empty")

		builder.SetError(fmt.Errorf("imageclusterinstall clusterdeployment cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.ClusterDeploymentRef == nil {
		builder.Definition.Spec.ClusterDeploymentRef = &corev1.LocalObjectReference{}
	}

	builder.Definition.Spec.ClusterDeploymentRef.Name = clusterDeploymentName

	return builder
}

// WithExtraManifests includes manifests via configmap name.
func (builder *ImageClusterInstallBuilder) WithExtraManifests(extraManifestsName string) *ImageClusterInstallBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if extraManifestsName == "" {
		klog.V(100).Info("The imageclusterinstall extramanifest is empty")

		builder.SetError(fmt.Errorf("imageclusterinstall extramanifest cannot be empty"))

		return builder
	}

	builder.Definition.Spec.ExtraManifestsRefs =
		append(builder.Definition.Spec.ExtraManifestsRefs, corev1.LocalObjectReference{
			Name: extraManifestsName,
		})

	return builder
}

// WithCABundle sets a CA bundle via configmap name.
func (builder *ImageClusterInstallBuilder) WithCABundle(caBundleConfigMapName string) *ImageClusterInstallBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if caBundleConfigMapName == "" {
		klog.V(100).Info("The imageclusterinstall cabundle is empty")

		builder.SetError(fmt.Errorf("imageclusterinstall cabundle cannot be empty"))

		return builder
	}

	builder.Definition.Spec.CABundleRef = &corev1.LocalObjectReference{Name: caBundleConfigMapName}

	return builder
}

// WithMachineNetwork specifies the machine network where nodes will be installed.
func (builder *ImageClusterInstallBuilder) WithMachineNetwork(network string) *ImageClusterInstallBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if _, _, err := net.ParseCIDR(network); err != nil {
		klog.V(100).Info("The machinenetwork is not a properly formatted IP network address")

		builder.SetError(fmt.Errorf("%s", errInvalidMachineNetworkFormat))

		return builder
	}

	builder.Definition.Spec.MachineNetwork = network

	return builder
}

// WithSSHKey adds specified ssh key to authorized_keys of installed nodes.
func (builder *ImageClusterInstallBuilder) WithSSHKey(sshKey string) *ImageClusterInstallBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if sshKey == "" {
		klog.V(100).Info("The imageclusterinstall sshkey is empty")

		builder.SetError(fmt.Errorf("imageclusterinstall sshkey cannot be empty"))

		return builder
	}

	builder.Definition.Spec.SSHKey = sshKey

	return builder
}

// GetCompletedCondition returns Completed condition from imageclusterinstall.
func (builder *ImageClusterInstallBuilder) GetCompletedCondition() (*hivev1.ClusterInstallCondition, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting Completed condition from imageclusterinstall %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	return builder.getCondition(hivev1.ClusterInstallCompleted)
}

// GetFailedCondition returns Failed condition from imageclusterinstall.
func (builder *ImageClusterInstallBuilder) GetFailedCondition() (*hivev1.ClusterInstallCondition, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting Failed condition from imageclusterinstall %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	return builder.getCondition(hivev1.ClusterInstallFailed)
}

// GetRequirementsMetCondition returns RequirementsMet condition from imageclusterinstall.
func (builder *ImageClusterInstallBuilder) GetRequirementsMetCondition() (*hivev1.ClusterInstallCondition, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting RequirementsMet condition from imageclusterinstall %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	return builder.getCondition(hivev1.ClusterInstallRequirementsMet)
}

// GetStoppedCondition returns Stopped condition from imageclusterinstall.
func (builder *ImageClusterInstallBuilder) GetStoppedCondition() (*hivev1.ClusterInstallCondition, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting Stopped condition from imageclusterinstall %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	return builder.getCondition(hivev1.ClusterInstallStopped)
}

func (builder *ImageClusterInstallBuilder) getCondition(
	conditionType hivev1.ClusterInstallConditionType) (*hivev1.ClusterInstallCondition, error) {
	if !builder.Exists() {
		return nil, fmt.Errorf("cannot get condition from non-existent imageclusterinstall")
	}

	for _, condition := range builder.Object.Status.Conditions {
		if condition.Type == conditionType {
			return &condition, nil
		}
	}

	return nil, fmt.Errorf("cannot find %s condition in imageclusterinstall status", conditionType)
}
