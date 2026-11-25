package olm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	operatorsV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var subscriptionGVK = operatorsV1alpha1.SchemeGroupVersion.WithKind("Subscription")

// SubscriptionBuilder provides a struct for Subscription object containing connection to the
// cluster and the Subscription definition.
type SubscriptionBuilder struct {
	common.EmbeddableBuilder[operatorsV1alpha1.Subscription, *operatorsV1alpha1.Subscription]
	common.EmbeddableCreator[
		operatorsV1alpha1.Subscription, SubscriptionBuilder, *operatorsV1alpha1.Subscription, *SubscriptionBuilder]
	common.EmbeddableDeleter[operatorsV1alpha1.Subscription, *operatorsV1alpha1.Subscription]
	common.EmbeddableUpdater[
		operatorsV1alpha1.Subscription, SubscriptionBuilder, *operatorsV1alpha1.Subscription, *SubscriptionBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *SubscriptionBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the Subscription GVK for this builder.
func (builder *SubscriptionBuilder) GetGVK() schema.GroupVersionKind {
	return subscriptionGVK
}

// NewSubscriptionBuilder returns a SubscriptionBuilder.
func NewSubscriptionBuilder(apiClient *clients.Settings, subName, subNamespace, catalogSource, catalogSourceNamespace,
	packageName string) *SubscriptionBuilder {
	builder := common.NewNamespacedBuilder[operatorsV1alpha1.Subscription, SubscriptionBuilder](
		apiClient, operatorsV1alpha1.AddToScheme, subName, subNamespace)
	if builder.GetError() != nil {
		return builder
	}

	if catalogSource == "" {
		klog.V(100).Info("The Catalogsource of the Subscription is empty")

		builder.SetError(fmt.Errorf("subscription 'catalogSource' cannot be empty"))

		return builder
	}

	if catalogSourceNamespace == "" {
		klog.V(100).Info("The Catalogsource namespace of the Subscription is empty")

		builder.SetError(fmt.Errorf("subscription 'catalogSourceNamespace' cannot be empty"))

		return builder
	}

	if packageName == "" {
		klog.V(100).Info("The Package name of the Subscription is empty")

		builder.SetError(fmt.Errorf("subscription 'packageName' cannot be empty"))

		return builder
	}

	builder.Definition.Spec = &operatorsV1alpha1.SubscriptionSpec{
		CatalogSource:          catalogSource,
		CatalogSourceNamespace: catalogSourceNamespace,
		Package:                packageName,
	}

	return builder
}

// WithChannel adds the specific channel to the Subscription.
func (builder *SubscriptionBuilder) WithChannel(channel string) *SubscriptionBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Defining Subscription builder object with channel: %s", channel)

	if channel == "" {
		builder.SetError(fmt.Errorf("can not redefine subscription with empty channel"))

		return builder
	}

	builder.Definition.Spec.Channel = channel

	return builder
}

// WithStartingCSV adds the specific startingCSV to the Subscription.
func (builder *SubscriptionBuilder) WithStartingCSV(startingCSV string) *SubscriptionBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Defining Subscription builder object with startingCSV: %s",
		startingCSV)

	if startingCSV == "" {
		builder.SetError(fmt.Errorf("can not redefine subscription with empty startingCSV"))

		return builder
	}

	builder.Definition.Spec.StartingCSV = startingCSV

	return builder
}

// WithInstallPlanApproval adds the specific installPlanApproval to the Subscription.
func (builder *SubscriptionBuilder) WithInstallPlanApproval(
	installPlanApproval operatorsV1alpha1.Approval) *SubscriptionBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Defining Subscription builder object with "+
		"installPlanApproval: %s", installPlanApproval)

	if installPlanApproval != "Automatic" && installPlanApproval != "Manual" {
		klog.V(100).Infof("The InstallPlanApproval of the Subscription must be either \"Automatic\" " +
			"or \"Manual\"")

		//nolint:staticcheck // Preserve the existing public error string for compatibility.
		builder.SetError(fmt.Errorf("Subscription 'installPlanApproval' must be either \"Automatic\" or \"Manual\""))

		return builder
	}

	builder.Definition.Spec.InstallPlanApproval = installPlanApproval

	return builder
}

// PullSubscription loads existing Subscription from cluster into the SubscriptionBuilder struct.
func PullSubscription(apiClient *clients.Settings, subName, subNamespace string) (*SubscriptionBuilder, error) {
	return common.PullNamespacedBuilder[operatorsV1alpha1.Subscription, SubscriptionBuilder](
		context.TODO(), apiClient, operatorsV1alpha1.AddToScheme, subName, subNamespace)
}
