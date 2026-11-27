package siteconfig

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	aiv1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	siteconfigv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/siteconfig/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/util/wait"
)

var clusterInstanceGVK = siteconfigv1alpha1.GroupVersion.WithKind(siteconfigv1alpha1.ClusterInstanceKind)

// CIBuilder provides struct for the ClusterInstance object.
type CIBuilder struct {
	common.EmbeddableBuilder[siteconfigv1alpha1.ClusterInstance, *siteconfigv1alpha1.ClusterInstance]
	common.EmbeddableCreator[
		siteconfigv1alpha1.ClusterInstance, CIBuilder, *siteconfigv1alpha1.ClusterInstance, *CIBuilder]
	common.EmbeddableDeleter[siteconfigv1alpha1.ClusterInstance, *siteconfigv1alpha1.ClusterInstance]
	common.EmbeddableForceUpdater[
		siteconfigv1alpha1.ClusterInstance, CIBuilder, *siteconfigv1alpha1.ClusterInstance, *CIBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *CIBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the ClusterInstance GVK for this builder.
func (builder *CIBuilder) GetGVK() schema.GroupVersionKind {
	return clusterInstanceGVK
}

// NewCIBuilder creates a new instance of CIBuilder.
func NewCIBuilder(apiClient *clients.Settings, name, nsname string) *CIBuilder {
	klog.V(100).Infof(
		"Initializing new ClusterInstance structure with the following params: name: %s, nsname: %s",
		name, nsname)

	return common.NewNamespacedBuilder[siteconfigv1alpha1.ClusterInstance, CIBuilder](
		apiClient, siteconfigv1alpha1.AddToScheme, name, nsname)
}

// PullClusterInstance retrieves an existing ClusterInstance from the cluster.
func PullClusterInstance(apiClient *clients.Settings, name, nsname string) (*CIBuilder, error) {
	klog.V(100).Infof(
		"Pulling existing clusterinstance with name %s from namespace %s", name, nsname)

	return common.PullNamespacedBuilder[siteconfigv1alpha1.ClusterInstance, CIBuilder](
		context.TODO(), apiClient, siteconfigv1alpha1.AddToScheme, name, nsname)
}

// WithPullSecretRef sets a reference to the pull-secret to be used by the clusterinstance.
func (builder *CIBuilder) WithPullSecretRef(secretRef string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding pullSecretRef %s to clusterinstance %s in namespace %s",
		secretRef, builder.Definition.Name, builder.Definition.Namespace)

	if secretRef == "" {
		klog.V(100).Info("The clusterinstance secretRef is empty")

		builder.SetError(fmt.Errorf("clusterinstance secretRef cannot be empty"))

		return builder
	}

	builder.Definition.Spec.PullSecretRef = corev1.LocalObjectReference{
		Name: secretRef,
	}

	return builder
}

// WithClusterTemplateRef specifies the cluster template to use for the clusterinstance.
func (builder *CIBuilder) WithClusterTemplateRef(clusterTemplateName, clusterTemplateNamespace string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding clusterTemplateRef %s in namespace %s to clusterinstance %s in namespace %s",
		clusterTemplateName, clusterTemplateNamespace, builder.Definition.Name, builder.Definition.Namespace)

	if clusterTemplateName == "" {
		klog.V(100).Info("The clusterinstance clusterTemplateName is empty")

		builder.SetError(fmt.Errorf("clusterinstance clusterTemplateName cannot be empty"))

		return builder
	}

	if clusterTemplateNamespace == "" {
		klog.V(100).Info("The clusterinstance clusterTemplateNamespace is empty")

		builder.SetError(fmt.Errorf("clusterinstance clusterTemplateNamespace cannot be empty"))

		return builder
	}

	builder.Definition.Spec.TemplateRefs = append(builder.Definition.Spec.TemplateRefs, siteconfigv1alpha1.TemplateRef{
		Name:      clusterTemplateName,
		Namespace: clusterTemplateNamespace,
	})

	return builder
}

// WithBaseDomain sets the base domain for the clusterinstance.
func (builder *CIBuilder) WithBaseDomain(baseDomain string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding baseDomain %s to clusterinstance %s in namespace %s",
		baseDomain, builder.Definition.Name, builder.Definition.Namespace)

	if baseDomain == "" {
		klog.V(100).Info("The clusterinstance baseDomain is empty")

		builder.SetError(fmt.Errorf("clusterinstance baseDomain cannot be empty"))

		return builder
	}

	builder.Definition.Spec.BaseDomain = baseDomain

	return builder
}

// WithClusterImageSetRef sets the clusterimageset used for installation in the clusterinstance.
func (builder *CIBuilder) WithClusterImageSetRef(imageSet string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding imageSet %s to clusterinstance %s in namespace %s",
		imageSet, builder.Definition.Name, builder.Definition.Namespace)

	if imageSet == "" {
		klog.V(100).Info("The clusterinstance imageSet is empty")

		builder.SetError(fmt.Errorf("clusterinstance imageSet cannot be empty"))

		return builder
	}

	builder.Definition.Spec.ClusterImageSetNameRef = imageSet

	return builder
}

// WithClusterName adds a cluster name to the clusterinstance.
func (builder *CIBuilder) WithClusterName(clusterName string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding clusterName %s to clusterinstance %s in namespace %s",
		clusterName, builder.Definition.Name, builder.Definition.Namespace)

	if clusterName == "" {
		klog.V(100).Info("The clusterinstance clusterName is empty")

		builder.SetError(fmt.Errorf("clusterinstance clusterName cannot be empty"))

		return builder
	}

	builder.Definition.Spec.ClusterName = clusterName

	return builder
}

// WithSSHPubKey adds the provided public SSH key for accessing the nodes.
func (builder *CIBuilder) WithSSHPubKey(sshPubKey string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding sshPubKey %s to clusterinstance %s in namespace %s",
		sshPubKey, builder.Definition.Name, builder.Definition.Namespace)

	if sshPubKey == "" {
		klog.V(100).Info("The clusterinstance sshPubKey is empty")

		builder.SetError(fmt.Errorf("clusterinstance sshPubKey cannot be empty"))

		return builder
	}

	builder.Definition.Spec.SSHPublicKey = sshPubKey

	return builder
}

// WithMachineNetwork adds the machineNetwork belonging to the node(s) to the clusterinstance.
func (builder *CIBuilder) WithMachineNetwork(machineNetwork string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding machineNetwork %s to clusterinstance %s in namespace %s",
		machineNetwork, builder.Definition.Name, builder.Definition.Namespace)

	if _, _, err := net.ParseCIDR(machineNetwork); err != nil {
		klog.V(100).Infof("The clusterinstance machineNetwork cidr %s is invalid cidr", machineNetwork)

		builder.SetError(fmt.Errorf("clusterinstance contains invalid machineNetwork cidr"))

		return builder
	}

	builder.Definition.Spec.MachineNetwork =
		append(builder.Definition.Spec.MachineNetwork, siteconfigv1alpha1.MachineNetworkEntry{
			CIDR: machineNetwork,
		})

	return builder
}

// WithAdditionalNTPSources appends additional NTP sources to the clusterinstance.
func (builder *CIBuilder) WithAdditionalNTPSources(ntpSources []string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding NTP sources %v to clusterinstance %s in namespace %s",
		ntpSources, builder.Definition.Name, builder.Definition.Namespace)

	if len(ntpSources) == 0 {
		klog.V(100).Infof("The clusterinstance additional NTP sources list is empty")

		builder.SetError(fmt.Errorf("clusterinstance additional NTP sources cannot be empty"))

		return builder
	}

	for _, ntpSource := range ntpSources {
		ntpSourceTrimmed := strings.TrimSpace(ntpSource)

		if ntpSourceTrimmed == "" {
			klog.V(100).Infof("An additional NTP source is empty")

			builder.SetError(fmt.Errorf("clusterinstance additional NTP source cannot be empty"))

			return builder
		}

		builder.Definition.Spec.AdditionalNTPSources =
			append(builder.Definition.Spec.AdditionalNTPSources, ntpSourceTrimmed)
	}

	return builder
}

// WithProxy adds the specified proxy to the clusterinstance.
func (builder *CIBuilder) WithProxy(proxy *aiv1beta1.Proxy) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding proxy %v to clusterinstance %s in namespace %s",
		proxy, builder.Definition.Name, builder.Definition.Namespace)

	if proxy == nil {
		klog.V(100).Info("The clusterinstance proxy is nil")

		builder.SetError(fmt.Errorf("clusterinstance proxy cannot be nil"))

		return builder
	}

	builder.Definition.Spec.Proxy = proxy

	return builder
}

// WithNode adds the specified node spec to the clusterinstance.
func (builder *CIBuilder) WithNode(node *siteconfigv1alpha1.NodeSpec) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if node == nil {
		klog.V(100).Info("The clusterinstance node is nil")

		builder.SetError(fmt.Errorf("clusterinstance node cannot be nil"))

		return builder
	}

	klog.V(100).Infof("Adding node %s to clusterinstance %s in namespace %s",
		node.HostName, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Nodes = append(builder.Definition.Spec.Nodes, *node)

	return builder
}

// WithExtraManifests includes manifests via configmap name.
func (builder *CIBuilder) WithExtraManifests(extraManifestsName string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding extra manifest %s to ClusterInstance definition", extraManifestsName)

	if extraManifestsName == "" {
		klog.V(100).Info("checking the clusterinstance extramanifest is empty")

		builder.SetError(fmt.Errorf("clusterinstance extramanifest cannot be empty"))

		return builder
	}

	builder.Definition.Spec.ExtraManifestsRefs =
		append(builder.Definition.Spec.ExtraManifestsRefs, corev1.LocalObjectReference{
			Name: extraManifestsName,
		})

	return builder
}

// WithCABundle sets a CA bundle via configmap name.
func (builder *CIBuilder) WithCABundle(caBundleConfigMapName string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding CA bundle via configmap %s to ClusterInstance definition", caBundleConfigMapName)

	if caBundleConfigMapName == "" {
		klog.V(100).Info("The clusterinstance cabundle is empty")

		builder.SetError(fmt.Errorf("clusterinstance cabundle cannot be empty"))

		return builder
	}

	builder.Definition.Spec.CaBundleRef = &corev1.LocalObjectReference{
		Name: caBundleConfigMapName,
	}

	return builder
}

// WithExtraLabels applies extraLabels to ClusterInstance definition.
func (builder *CIBuilder) WithExtraLabels(key string, labels map[string]string) *CIBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Defining clusterinstance extraLabels to %s:%v", key, labels)

	if key == "" {
		klog.V(100).Info("checking the key is empty")

		builder.SetError(fmt.Errorf("can not apply empty key"))

		return builder
	}

	if len(labels) == 0 {
		klog.V(100).Info("checking the labels are empty")

		builder.SetError(fmt.Errorf("labels can not be empty"))

		return builder
	}

	for key := range labels {
		if key == "" {
			klog.V(100).Info("The 'labels' key cannot be empty")

			builder.SetError(fmt.Errorf("can not apply a labels with an empty key"))

			return builder
		}
	}

	if builder.Definition.Spec.ExtraLabels == nil {
		builder.Definition.Spec.ExtraLabels = map[string]map[string]string{}
	}

	builder.Definition.Spec.ExtraLabels[key] = labels

	return builder
}

// WaitForCondition waits until the ClusterInstance
// has a condition that matches the expected, checking only the Type, Status, Reason, and Message fields.
// For the message field, it matches if the message contains the expected.
// Zero fields in the expected condition are ignored.
func (builder *CIBuilder) WaitForCondition(expected metav1.Condition, timeout time.Duration) (*CIBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	if !builder.Exists() {
		klog.V(100).Info("The clusterinstance does not exist on the cluster")

		return builder, fmt.Errorf(
			"clusterinstance object %s does not exist in namespace %s", builder.Definition.Name, builder.Definition.Namespace)
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("failed to get clusterinstance %s/%s: %v",
					builder.Definition.Namespace, builder.Definition.Name, err)

				return false, nil
			}

			builder.Definition = builder.Object

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

	return builder, err
}

// WaitForReinstallCondition waits until the ClusterInstance
// has a reinstall condition that matches the expected, checking only the Type, Status, Reason, and Message fields.
// For the message field, it matches if the message contains the expected.
// Zero fields in the expected condition are ignored.
func (builder *CIBuilder) WaitForReinstallCondition(expected metav1.Condition,
	timeout time.Duration) (*CIBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	if !builder.Exists() {
		klog.V(100).Info("The clusterinstance does not exist on the cluster")

		return builder, fmt.Errorf(
			"clusterinstance object %s does not exist in namespace %s", builder.Definition.Name, builder.Definition.Namespace)
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("failed to get clusterinstance %s/%s: %v",
					builder.Definition.Namespace, builder.Definition.Name, err)

				return false, nil
			}

			if builder.Object.Status.Reinstall == nil || builder.Object.Status.Reinstall.Conditions == nil {
				klog.V(100).Infof("failed to get reinstall status %s/%s: %v",
					builder.Definition.Namespace, builder.Definition.Name, err)

				return false, nil
			}

			builder.Definition = builder.Object

			for _, condition := range builder.Object.Status.Reinstall.Conditions {
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

	return builder, err
}

// WaitForExtraLabel waits up to timeout until the ExtraLabel label exists for manifest of kind.
func (builder *CIBuilder) WaitForExtraLabel(kind, label string, timeout time.Duration) (*CIBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Waiting up to %s until ClusterInstance %s in namespace %s has extra label %s on kind %s",
		timeout, builder.Definition.Name, builder.Definition.Namespace, label, kind)

	if !builder.Exists() {
		klog.V(100).Infof("ClusterInstance %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, fmt.Errorf(
			"clusterinstance object %s does not exist in namespace %s", builder.Definition.Name, builder.Definition.Namespace)
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("Failed to get ClusterInstance %s in namespace %s: %v",
					builder.Definition.Name, builder.Definition.Namespace, err)

				return false, nil
			}

			builder.Definition = builder.Object

			if builder.Definition.Spec.ExtraLabels == nil {
				return false, nil
			}

			kindLabels, exists := builder.Definition.Spec.ExtraLabels[kind]
			if !exists {
				return false, nil
			}

			_, exists = kindLabels[label]

			return exists, nil
		})
	if err != nil {
		return nil, err
	}

	return builder, nil
}
