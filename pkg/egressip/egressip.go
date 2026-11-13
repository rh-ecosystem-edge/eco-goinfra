package egressip

import (
	"context"
	"fmt"

	egressipv1 "github.com/ovn-kubernetes/ovn-kubernetes/go-controller/pkg/crd/egressip/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var egressIPGVK = egressipv1.SchemeGroupVersion.WithKind("EgressIP")

// EgressIPBuilder provides a struct for EgressIP object.
type EgressIPBuilder struct {
	common.EmbeddableBuilder[egressipv1.EgressIP, *egressipv1.EgressIP]
	common.EmbeddableCreator[egressipv1.EgressIP, EgressIPBuilder, *egressipv1.EgressIP, *EgressIPBuilder]
	common.EmbeddableDeleteReturner[egressipv1.EgressIP, EgressIPBuilder, *egressipv1.EgressIP, *EgressIPBuilder]
	common.EmbeddableUpdater[egressipv1.EgressIP, EgressIPBuilder, *egressipv1.EgressIP, *EgressIPBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *EgressIPBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the EgressIP GVK for this builder.
func (builder *EgressIPBuilder) GetGVK() schema.GroupVersionKind {
	return egressIPGVK
}

// NewEgressIPBuilder creates a new instance of EgressIP builder.
func NewEgressIPBuilder(apiClient *clients.Settings, name string) *EgressIPBuilder {
	klog.V(100).Infof(
		"Initializing new EgressIP structure with the following params: name: %s", name)

	return common.NewClusterScopedBuilder[egressipv1.EgressIP, EgressIPBuilder](
		apiClient, egressipv1.AddToScheme, name)
}

// WithEgressIPs applies egressIPs to the EgressIP spec definition.
func (builder *EgressIPBuilder) WithEgressIPs(egressIPs []string) *EgressIPBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying egressIPs %s to EgressIP %q", egressIPs, builder.Definition.Name)

	if len(egressIPs) == 0 {
		klog.V(100).Info("The egressIP is empty")

		builder.SetError(fmt.Errorf("cannot accept empty list as egressIPs value"))

		return builder
	}

	builder.Definition.Spec.EgressIPs = egressIPs

	return builder
}

// WithNamespaceSelector applies namespaceSelector to the EgressIP spec definition.
func (builder *EgressIPBuilder) WithNamespaceSelector(namespaceSelector metav1.LabelSelector) *EgressIPBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying namespaceSelector %s to EgressIP %q",
		namespaceSelector, builder.Definition.Name)

	if len(namespaceSelector.MatchLabels) == 0 && len(namespaceSelector.MatchExpressions) == 0 {
		klog.V(100).Info("There are no labels for the namespaceSelector")

		builder.SetError(fmt.Errorf("EgressIP 'namespaceSelector' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.NamespaceSelector = namespaceSelector

	return builder
}

// WithPodSelector applies podSelector to the EgressIP spec definition.
func (builder *EgressIPBuilder) WithPodSelector(podSelector metav1.LabelSelector) *EgressIPBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying podSelector %s to EgressIP %q", podSelector, builder.Definition.Name)

	if len(podSelector.MatchLabels) == 0 && len(podSelector.MatchExpressions) == 0 {
		klog.V(100).Info("There are no labels for the podSelector")

		builder.SetError(fmt.Errorf("EgressIP 'podSelector' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.PodSelector = podSelector

	return builder
}

// Pull fetches existing egressIP from the cluster.
func Pull(apiClient *clients.Settings, name string) (*EgressIPBuilder, error) {
	klog.V(100).Infof("Pulling existing egressIP %q from cluster", name)

	return common.PullClusterScopedBuilder[egressipv1.EgressIP, EgressIPBuilder](
		context.TODO(), apiClient, egressipv1.AddToScheme, name)
}

// GetAssignedEgressIPMap fetches the next recommended or conditional update for the cluster.
func (builder *EgressIPBuilder) GetAssignedEgressIPMap() (map[string]string, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Pull assigned egressIPs map for egressIP %q", builder.Definition.Name)

	if !builder.Exists() {
		return nil, fmt.Errorf("egressIP %q object does not exist", builder.Definition.Name)
	}

	if len(builder.Object.Status.Items) == 0 {
		return nil, fmt.Errorf("egressIP %q nodes assignment does not exist", builder.Definition.Name)
	}

	egressIPMap := make(map[string]string)
	for _, item := range builder.Object.Status.Items {
		egressIPMap[item.Node] = item.EgressIP
	}

	return egressIPMap, nil
}
