package nfd

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	nfdv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/nfd/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
)

const (
	// ruleNameEmptyError is the error message when rule name is empty.
	ruleNameEmptyError = "rule 'name' cannot be empty"
)

var nodeFeatureRuleGVK = nfdv1.GroupVersion.WithKind("NodeFeatureRule")

// NodeFeatureRuleBuilder provides a struct for NodeFeatureRule object
// from the cluster and a NodeFeatureRule definition.
type NodeFeatureRuleBuilder struct {
	common.EmbeddableBuilder[nfdv1.NodeFeatureRule, *nfdv1.NodeFeatureRule]
	common.EmbeddableCreator[
		nfdv1.NodeFeatureRule, NodeFeatureRuleBuilder, *nfdv1.NodeFeatureRule, *NodeFeatureRuleBuilder]
	common.EmbeddableDeleteReturner[
		nfdv1.NodeFeatureRule, NodeFeatureRuleBuilder, *nfdv1.NodeFeatureRule, *NodeFeatureRuleBuilder]
	common.EmbeddableForceUpdater[
		nfdv1.NodeFeatureRule, NodeFeatureRuleBuilder, *nfdv1.NodeFeatureRule, *NodeFeatureRuleBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *NodeFeatureRuleBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the NodeFeatureRule GVK for this builder.
func (builder *NodeFeatureRuleBuilder) GetGVK() schema.GroupVersionKind {
	return nodeFeatureRuleGVK
}

// NewNodeFeatureRuleBuilder creates a new instance of NodeFeatureRuleBuilder.
func NewNodeFeatureRuleBuilder(apiClient *clients.Settings, name, namespace string) *NodeFeatureRuleBuilder {
	builder := common.NewNamespacedBuilder[nfdv1.NodeFeatureRule, NodeFeatureRuleBuilder](
		apiClient, nfdv1.AddToScheme, name, namespace)
	builder.Definition.Spec.Rules = []nfdv1.Rule{}

	return builder
}

// NewNodeFeatureRuleBuilderFromObjectString creates a Builder object from CSV alm-examples.
func NewNodeFeatureRuleBuilderFromObjectString(apiClient *clients.Settings, almExample string) *NodeFeatureRuleBuilder {
	klog.V(100).Infof("Initializing new Builder structure from almExample string")

	if apiClient == nil {
		klog.V(100).Info("The apiClient of the NodeFeatureRule is nil")

		return nil
	}

	if err := apiClient.AttachScheme(nfdv1.AddToScheme); err != nil {
		klog.V(100).Infof("Failed to add nfd v1 scheme to client schemes: %v", err)

		return nil
	}

	builder := &NodeFeatureRuleBuilder{}
	builder.AttachMixins()
	builder.SetGVK(builder.GetGVK())
	builder.SetClient(apiClient)

	nodeFeatureRule, err := getNodeFeatureRuleFromAlmExample(almExample)
	if err != nil {
		klog.V(100).Infof(
			"Error initializing NodeFeatureRule from alm-examples: %s", err.Error())

		builder.SetError(fmt.Errorf("error initializing NodeFeatureRule from alm-examples: %w", err))

		return builder
	}

	klog.V(100).Infof("Initializing Builder definition to NodeFeatureRule object")

	builder.SetDefinition(nodeFeatureRule)

	if builder.Definition == nil {
		klog.V(100).Info("The NodeFeatureRule object definition is nil")

		builder.SetError(fmt.Errorf("nodeFeatureRule definition is nil"))
	}

	return builder
}

// PullFeatureRule loads an existing NodeFeatureRule into Builder struct.
func PullFeatureRule(apiClient *clients.Settings, name, namespace string) (*NodeFeatureRuleBuilder, error) {
	return common.PullNamespacedBuilder[nfdv1.NodeFeatureRule, NodeFeatureRuleBuilder](
		context.TODO(), apiClient, nfdv1.AddToScheme, name, namespace)
}

// getNodeFeatureRuleFromAlmExample extracts the NodeFeatureRule from the alm-examples block.
func getNodeFeatureRuleFromAlmExample(almExample string) (*nfdv1.NodeFeatureRule, error) {
	nodeFeatureRuleList := &nfdv1.NodeFeatureRuleList{}

	if almExample == "" {
		return nil, fmt.Errorf("almExample is an empty string")
	}

	err := json.Unmarshal([]byte(almExample), &nodeFeatureRuleList.Items)
	if err != nil {
		return nil, err
	}

	if len(nodeFeatureRuleList.Items) == 0 {
		return nil, fmt.Errorf("failed to get alm examples")
	}

	for i, item := range nodeFeatureRuleList.Items {
		if item.Kind == "NodeFeatureRule" {
			return &nodeFeatureRuleList.Items[i], nil
		}
	}

	return nil, fmt.Errorf("nodeFeatureRule is missing in alm-examples ")
}

// WithRule adds a rule to the NodeFeatureRule.
func (builder *NodeFeatureRuleBuilder) WithRule(rule nfdv1.Rule) *NodeFeatureRuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding rule %s to NodeFeatureRule %s in namespace %s",
		rule.Name, builder.Definition.Name, builder.Definition.Namespace)

	if rule.Name == "" {
		klog.V(100).Info("Rule 'name' cannot be empty")

		builder.SetError(fmt.Errorf("%s", ruleNameEmptyError))

		return builder
	}

	builder.Definition.Spec.Rules = append(builder.Definition.Spec.Rules, rule)

	return builder
}

// WithRules sets the rules for the NodeFeatureRule, replacing any existing rules.
func (builder *NodeFeatureRuleBuilder) WithRules(rules []nfdv1.Rule) *NodeFeatureRuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting %d rules on NodeFeatureRule %s in namespace %s",
		len(rules), builder.Definition.Name, builder.Definition.Namespace)

	if len(rules) == 0 {
		klog.V(100).Info("Rules list cannot be empty")

		builder.SetError(fmt.Errorf("rules list cannot be empty"))

		return builder
	}

	for _, rule := range rules {
		if rule.Name == "" {
			klog.V(100).Info("Rule 'name' cannot be empty")

			builder.SetError(fmt.Errorf("%s", ruleNameEmptyError))

			return builder
		}
	}

	builder.Definition.Spec.Rules = rules

	return builder
}

// WithSimplePCIRule adds a simple PCI device matching rule with labels.
// This is a convenience method for common PCI device detection patterns like the Neuron NFD rule.
func (builder *NodeFeatureRuleBuilder) WithSimplePCIRule(
	ruleName string,
	labels map[string]string,
	vendorIDs []string,
	deviceIDs []string,
) *NodeFeatureRuleBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Adding simple PCI rule %s to NodeFeatureRule %s in namespace %s",
		ruleName, builder.Definition.Name, builder.Definition.Namespace)

	if ruleName == "" {
		klog.V(100).Info("Rule 'name' cannot be empty")

		builder.SetError(fmt.Errorf("%s", ruleNameEmptyError))

		return builder
	}

	if len(labels) == 0 {
		klog.V(100).Info("Rule 'labels' cannot be empty")

		builder.SetError(fmt.Errorf("rule 'labels' cannot be empty"))

		return builder
	}

	if len(vendorIDs) == 0 {
		klog.V(100).Info("vendorIDs cannot be empty")

		builder.SetError(fmt.Errorf("vendorIDs cannot be empty"))

		return builder
	}

	matchExpressions := nfdv1.MatchExpressionSet{
		"vendor": &nfdv1.MatchExpression{
			Op:    nfdv1.MatchIn,
			Value: vendorIDs,
		},
	}

	if len(deviceIDs) > 0 {
		matchExpressions["device"] = &nfdv1.MatchExpression{
			Op:    nfdv1.MatchIn,
			Value: deviceIDs,
		}
	}

	rule := nfdv1.Rule{
		Name:   ruleName,
		Labels: labels,
		MatchAny: []nfdv1.MatchAnyElem{
			{
				MatchFeatures: nfdv1.FeatureMatcher{
					{
						Feature:          "pci.device",
						MatchExpressions: &matchExpressions,
					},
				},
			},
		},
	}

	builder.Definition.Spec.Rules = append(builder.Definition.Spec.Rules, rule)

	return builder
}
