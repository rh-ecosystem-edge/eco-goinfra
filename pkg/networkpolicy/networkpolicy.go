package networkpolicy

import (
	"context"
	"errors"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errEmptyNamespaceIngressSelectors = "Both namespaceIngressMatchLabels and podIngressMatchLabels parameters are empty maps"
	errEmptyPolicyType                = "The policyType is an empty string"
	errEmptyPodSelector               = "The podSelector is an empty string"
)

// NetworkPolicyBuilder provides struct for networkPolicy object.
type NetworkPolicyBuilder struct {
	common.EmbeddableBuilder[netv1.NetworkPolicy, *netv1.NetworkPolicy]
	common.EmbeddableCreator[netv1.NetworkPolicy, NetworkPolicyBuilder, *netv1.NetworkPolicy, *NetworkPolicyBuilder]
	common.EmbeddableDeleter[netv1.NetworkPolicy, *netv1.NetworkPolicy]
	common.EmbeddableUpdater[netv1.NetworkPolicy, NetworkPolicyBuilder, *netv1.NetworkPolicy, *NetworkPolicyBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *NetworkPolicyBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the NetworkPolicy GVK for this builder.
func (builder *NetworkPolicyBuilder) GetGVK() schema.GroupVersionKind {
	return netv1.SchemeGroupVersion.WithKind("NetworkPolicy")
}

// NewNetworkPolicyBuilder method creates new instance of builder.
func NewNetworkPolicyBuilder(apiClient *clients.Settings, name, nsname string) *NetworkPolicyBuilder {
	return common.NewNamespacedBuilder[netv1.NetworkPolicy, NetworkPolicyBuilder](
		apiClient, netv1.AddToScheme, name, nsname)
}

// WithNamespaceIngressRule applies ingress rule for the networkPolicy.
func (builder *NetworkPolicyBuilder) WithNamespaceIngressRule(
	namespaceIngressMatchLabels map[string]string,
	podIngressMatchLabels map[string]string) *NetworkPolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Applying Ingress rule to networkPolicy %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if len(namespaceIngressMatchLabels) == 0 && len(podIngressMatchLabels) == 0 {
		klog.V(100).Info("At least one type of the selector for NetworkPolicy ingress rule should be defined")

		//nolint:staticcheck // Preserve the existing caller-visible error text.
		builder.SetError(errors.New(errEmptyNamespaceIngressSelectors))

		return builder
	}

	var peerRule netv1.NetworkPolicyPeer

	if len(namespaceIngressMatchLabels) != 0 {
		klog.V(100).Infof(
			"Applying Ingress rule with namespaceIngressMatchLabels %v parameter to networkPolicy %s in namespace %s",
			namespaceIngressMatchLabels, builder.Definition.Name, builder.Definition.Namespace)

		peerRule.NamespaceSelector = &metav1.LabelSelector{
			MatchLabels: namespaceIngressMatchLabels,
		}
	}

	if len(podIngressMatchLabels) != 0 {
		klog.V(100).Infof(
			"Applying Ingress rule with podIngressMatchLabels %v parameter to networkPolicy %s in namespace %s",
			podIngressMatchLabels, builder.Definition.Name, builder.Definition.Namespace)

		peerRule.PodSelector = &metav1.LabelSelector{
			MatchLabels: podIngressMatchLabels,
		}
	}

	builder.Definition.Spec.Ingress = append(builder.Definition.Spec.Ingress, netv1.NetworkPolicyIngressRule{
		From: []netv1.NetworkPolicyPeer{peerRule},
	})

	return builder
}

// WithPolicyType add policyType to the networkPolicy.
func (builder *NetworkPolicyBuilder) WithPolicyType(policyType netv1.PolicyType) *NetworkPolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating networkPolicy %s in %s namespace with the policyType defined: %v",
		builder.Definition.Name, builder.Definition.Namespace, policyType)

	if policyType == "" {
		klog.V(100).Info("The policyType value has to be provided")

		//nolint:staticcheck // Preserve the existing caller-visible error text.
		builder.SetError(errors.New(errEmptyPolicyType))

		return builder
	}

	builder.Definition.Spec.PolicyTypes = append(builder.Definition.Spec.PolicyTypes, policyType)

	return builder
}

// WithPodSelector add podSelector to the networkPolicy.
func (builder *NetworkPolicyBuilder) WithPodSelector(podSelectorMatchLabels map[string]string) *NetworkPolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating networkPolicy %s in %s namespace with podSelector defined: %v",
		builder.Definition.Name, builder.Definition.Namespace, podSelectorMatchLabels)

	if len(podSelectorMatchLabels) == 0 {
		klog.V(100).Info("The podSelector could not be empty")

		//nolint:staticcheck // Preserve the existing caller-visible error text.
		builder.SetError(errors.New(errEmptyPodSelector))

		return builder
	}

	builder.Definition.Spec.PodSelector = metav1.LabelSelector{MatchLabels: podSelectorMatchLabels}

	return builder
}

// Pull loads an existing networkPolicy into the Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*NetworkPolicyBuilder, error) {
	return common.PullNamespacedBuilder[netv1.NetworkPolicy, NetworkPolicyBuilder](
		context.TODO(), apiClient, netv1.AddToScheme, name, nsname)
}

// GetGVR returns networkPolicy's GroupVersionResource which could be used for Clean function.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"}
}
