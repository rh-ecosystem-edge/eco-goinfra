package networkpolicy

import (
	"context"
	"errors"

	"github.com/k8snetworkplumbingwg/multi-networkpolicy/pkg/apis/k8s.cni.cncf.io/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errEmptyNetworkName     = "The networkName is an empty string"
	errEmptyMultiPolicyType = "The policy Type is an empty string"
)

// MultiNetworkPolicyBuilder provides struct for MultiNetworkPolicy object.
type MultiNetworkPolicyBuilder struct {
	common.EmbeddableBuilder[v1beta1.MultiNetworkPolicy, *v1beta1.MultiNetworkPolicy]
	common.EmbeddableCreator[
		v1beta1.MultiNetworkPolicy, MultiNetworkPolicyBuilder, *v1beta1.MultiNetworkPolicy, *MultiNetworkPolicyBuilder]
	common.EmbeddableDeleter[v1beta1.MultiNetworkPolicy, *v1beta1.MultiNetworkPolicy]
	common.EmbeddableUpdater[
		v1beta1.MultiNetworkPolicy, MultiNetworkPolicyBuilder, *v1beta1.MultiNetworkPolicy, *MultiNetworkPolicyBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *MultiNetworkPolicyBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the MultiNetworkPolicy GVK for this builder.
func (builder *MultiNetworkPolicyBuilder) GetGVK() schema.GroupVersionKind {
	return v1beta1.SchemeGroupVersion.WithKind("MultiNetworkPolicy")
}

// NewMultiNetworkPolicyBuilder method creates new instance of builder.
func NewMultiNetworkPolicyBuilder(apiClient *clients.Settings, name, nsname string) *MultiNetworkPolicyBuilder {
	return common.NewNamespacedBuilder[v1beta1.MultiNetworkPolicy, MultiNetworkPolicyBuilder](
		apiClient, v1beta1.AddToScheme, name, nsname)
}

// WithPodSelector adds podSelector to MultiNetworkPolicy.
func (builder *MultiNetworkPolicyBuilder) WithPodSelector(podSelector metav1.LabelSelector) *MultiNetworkPolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating MultiNetworkPolicy %s in %s namespace with the podSelector defined: %v",
		builder.Definition.Name, builder.Definition.Namespace, podSelector)

	builder.Definition.Spec.PodSelector = podSelector

	return builder
}

// WithNetwork adds network name to the MultiNetworkPolicy.
func (builder *MultiNetworkPolicyBuilder) WithNetwork(networkName string) *MultiNetworkPolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating MultiNetworkPolicy %s in %s namespace with the networkName defined: %v",
		builder.Definition.Name, builder.Definition.Namespace, networkName)

	if networkName == "" {
		klog.V(100).Info("The networkName can not be empty string")

		//nolint:staticcheck // Preserve the existing caller-visible error text.
		builder.SetError(errors.New(errEmptyNetworkName))

		return builder
	}

	if builder.Definition.Annotations == nil {
		builder.Definition.Annotations = make(map[string]string)
	}

	builder.Definition.Annotations["k8s.v1.cni.cncf.io/policy-for"] = networkName

	return builder
}

// WithEmptyIngress adds empty ingress rule to the MultiNetworkPolicy. Empty ingress denies all.
func (builder *MultiNetworkPolicyBuilder) WithEmptyIngress() *MultiNetworkPolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating MultiNetworkPolicy %s in %s namespace with the empty Ingress rule deny all",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Ingress = []v1beta1.MultiNetworkPolicyIngressRule{}

	return builder
}

// WithIngressRule adds Ingress rule to the MultiNetworkPolicy. Empty rule is allowed and works as allow all traffic.
func (builder *MultiNetworkPolicyBuilder) WithIngressRule(
	ingressRule v1beta1.MultiNetworkPolicyIngressRule) *MultiNetworkPolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating multiNetworkPolicy %s in %s namespace with the Ingress rule defined: %v",
		builder.Definition.Name, builder.Definition.Namespace, ingressRule)

	builder.Definition.Spec.Ingress = append(builder.Definition.Spec.Ingress, ingressRule)

	return builder
}

// WithEgressRule adds Egress rule to the MultiNetworkPolicy. Empty rule is allowed and works as allow all traffic.
func (builder *MultiNetworkPolicyBuilder) WithEgressRule(
	egressRule v1beta1.MultiNetworkPolicyEgressRule) *MultiNetworkPolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating multiNetworkPolicy %s in %s namespace with the Egress rule defined: %v",
		builder.Definition.Name, builder.Definition.Namespace, egressRule)

	builder.Definition.Spec.Egress = append(builder.Definition.Spec.Egress, egressRule)

	return builder
}

// WithPolicyType adds policyType to the MultiNetworkPolicy.
func (builder *MultiNetworkPolicyBuilder) WithPolicyType(
	policyType v1beta1.MultiPolicyType) *MultiNetworkPolicyBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating multiNetworkPolicy %s in %s namespace with the Policy type defined: %v",
		builder.Definition.Name, builder.Definition.Namespace, policyType)

	if policyType == "" {
		klog.V(100).Info("The policy type can not be an empty string")

		//nolint:staticcheck // Preserve the existing caller-visible error text.
		builder.SetError(errors.New(errEmptyMultiPolicyType))

		return builder
	}

	builder.Definition.Spec.PolicyTypes = append(builder.Definition.Spec.PolicyTypes, policyType)

	return builder
}

// PullMultiNetworkPolicy loads an existing MultiNetworkPolicy into the Builder struct.
func PullMultiNetworkPolicy(apiClient *clients.Settings, name, nsname string) (*MultiNetworkPolicyBuilder, error) {
	return common.PullNamespacedBuilder[v1beta1.MultiNetworkPolicy, MultiNetworkPolicyBuilder](
		context.TODO(), apiClient, v1beta1.AddToScheme, name, nsname)
}

// GetMultiNetworkGVR returns MultiNetworkPolicy's GroupVersionResource which could be used for Clean function.
func GetMultiNetworkGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "k8s.cni.cncf.io", Version: "v1beta1", Resource: "multi-networkpolicies"}
}
