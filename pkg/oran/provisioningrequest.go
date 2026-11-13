package oran

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errEmptyTemplateName = "provisioningRequest 'templateName' cannot be empty"
)

// ProvisioningRequestBuilder provides a struct to inferface with ProvisioningRequest resources on a specific cluster.
type ProvisioningRequestBuilder struct {
	common.EmbeddableBuilder[
		provisioningv1alpha1.ProvisioningRequest, *provisioningv1alpha1.ProvisioningRequest]
	common.EmbeddableCreator[
		provisioningv1alpha1.ProvisioningRequest, ProvisioningRequestBuilder,
		*provisioningv1alpha1.ProvisioningRequest, *ProvisioningRequestBuilder]
	common.EmbeddableDeleter[provisioningv1alpha1.ProvisioningRequest, *provisioningv1alpha1.ProvisioningRequest]
	common.EmbeddableUpdater[
		provisioningv1alpha1.ProvisioningRequest, ProvisioningRequestBuilder,
		*provisioningv1alpha1.ProvisioningRequest, *ProvisioningRequestBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ProvisioningRequestBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the ProvisioningRequest GVK for this builder.
func (builder *ProvisioningRequestBuilder) GetGVK() schema.GroupVersionKind {
	return provisioningv1alpha1.GroupVersion.WithKind("ProvisioningRequest")
}

// NewPRBuilder creates a new instance of a ProvisioningRequest builder.
func NewPRBuilder(
	apiClient runtimeclient.Client, name, templateName, templateVersion string) *ProvisioningRequestBuilder {
	klog.V(100).Infof(
		"Initializing new ProvisioningRequest structure with the following params: "+
			"name: %s, templateName: %s, templateVersion: %s",
		name, templateName, templateVersion)

	builder := common.NewClusterScopedBuilder[
		provisioningv1alpha1.ProvisioningRequest, ProvisioningRequestBuilder](
		apiClient, provisioningv1alpha1.AddToScheme, name)
	if builder.GetError() != nil {
		return builder
	}

	if err := uuid.Validate(name); err != nil {
		klog.V(100).Infof("The name of the ProvisioningRequest is not a valid UUID: %v", err)

		builder.SetError(fmt.Errorf("provisioningRequest 'name' must be a valid UUID"))

		return builder
	}

	if templateName == "" {
		klog.V(100).Info("The template name of the ProvisioningRequest is empty")

		builder.SetError(fmt.Errorf("%s", errEmptyTemplateName))

		return builder
	}

	if templateVersion == "" {
		klog.V(100).Info("The template version of the ProvisioningRequest is empty")

		builder.SetError(fmt.Errorf("provisioningRequest 'templateVersion' cannot be empty"))

		return builder
	}

	builder.Definition.Spec = provisioningv1alpha1.ProvisioningRequestSpec{
		// Name is supposed to be human readable, but is required for the O2IMS API. We default
		// to using the ID and consumers can override it if desired.
		Name:            name,
		TemplateName:    templateName,
		TemplateVersion: templateVersion,
	}

	return builder
}

// WithTemplateParameter sets key to value in the TemplateParameters field.
func (builder *ProvisioningRequestBuilder) WithTemplateParameter(key string, value any) *ProvisioningRequestBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting ProvisioningRequest TemplateParameter %s to %v", key, value)

	if key == "" {
		klog.V(100).Info("ProvisioningRequest TemplateParameter key is empty")

		builder.SetError(fmt.Errorf("provisioningRequest TemplateParameter 'key' cannot be empty"))

		return builder
	}

	templateParameters, err := builder.GetTemplateParameters()
	if err != nil {
		klog.V(100).Infof("Failed to unmarshal ProvisioningRequest TemplateParameters: %v", err)

		builder.SetError(fmt.Errorf("failed to unmarshal TemplateParameters: %w", err))

		return builder
	}

	templateParameters[key] = value
	builder = builder.WithTemplateParameters(templateParameters)

	return builder
}

// GetTemplateParameters unmarshals the raw JSON stored in the TemplateParameters in the Definition, returning an empty
// rather than nil map if TemplateParameters is empty.
func (builder *ProvisioningRequestBuilder) GetTemplateParameters() (map[string]any, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting the TemplateParameters map for ProvisioningRequest %s", builder.Definition.Name)

	templateParameters := make(map[string]any)

	if len(builder.Definition.Spec.TemplateParameters.Raw) == 0 {
		return templateParameters, nil
	}

	err := json.Unmarshal(builder.Definition.Spec.TemplateParameters.Raw, &templateParameters)
	if err != nil {
		return nil, err
	}

	return templateParameters, nil
}

// WithTemplateParameters marshals the provided map into JSON and stores it in the builder Definition. Nil maps are
// converted to empty maps before marshaling.
func (builder *ProvisioningRequestBuilder) WithTemplateParameters(
	templateParameters map[string]any) *ProvisioningRequestBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting the TemplateParameters map for ProvisioningRequest %s", builder.Definition.Name)

	if templateParameters == nil {
		templateParameters = make(map[string]any)
	}

	marshaled, err := json.Marshal(templateParameters)
	if err != nil {
		klog.V(100).Infof("Failed to marshal TemplateParameters for ProvisioningRequest %s: %v", builder.Definition.Name, err)

		builder.SetError(fmt.Errorf("failed to marshal TemplateParameters: %w", err))

		return builder
	}

	builder.Definition.Spec.TemplateParameters = runtime.RawExtension{Raw: marshaled}

	return builder
}

// PullPR pulls an existing ProvisioningRequest into a Builder struct.
func PullPR(apiClient runtimeclient.Client, name string) (*ProvisioningRequestBuilder, error) {
	return common.PullClusterScopedBuilder[
		provisioningv1alpha1.ProvisioningRequest, ProvisioningRequestBuilder](
		context.TODO(), apiClient, provisioningv1alpha1.AddToScheme, name)
}

// DeleteAndWait deletes the ProvisioningRequest then waits up to timeout until the ProvisioningRequest no longer
// exists.
func (builder *ProvisioningRequestBuilder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Deleting ProvisioningRequest %s and waiting up to %s until it is deleted",
		builder.Definition.Name, timeout)

	err := builder.Delete()
	if err != nil {
		return err
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			return !builder.Exists(), nil
		})
}

// WaitForCondition waits up to the provided timeout for a condition matching expected. It checks only the Type, Status,
// Reason, and Message fields. For the message, it matches if the message contains the expected. Zero fields in the
// expected condition are ignored.
func (builder *ProvisioningRequestBuilder) WaitForCondition(
	expected metav1.Condition, timeout time.Duration) (*ProvisioningRequestBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Waiting up to %s until ProvisioningRequest %s has condition %v",
		timeout, builder.Definition.Name, expected)

	if !builder.Exists() {
		klog.V(100).Infof("ProvisioningRequest %s does not exist", builder.Definition.Name)

		return nil, fmt.Errorf("cannot wait for non-existent ProvisioningRequest")
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("Failed to get ProvisioningRequest %s: %v", builder.Definition.Name, err)

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
	if err != nil {
		return nil, err
	}

	return builder, nil
}

// WaitUntilFulfilled waits up to timeout until the ProvisioningRequest is in the fulfilled phase. Changes to the
// template will not be accepted until it is fulfilled.
//
// This function is equivalent to WaitUntilFulfilledAfter with a zero-valued start time.
func (builder *ProvisioningRequestBuilder) WaitUntilFulfilled(
	timeout time.Duration) (*ProvisioningRequestBuilder, error) {
	err := builder.WaitForPhaseAfter(provisioningv1alpha1.StateFulfilled, time.Time{}, timeout)
	if err != nil {
		return nil, err
	}

	return builder, nil
}

// WaitForPhaseAfter waits up to timeout until the ProvisioningRequest is in the provided phase and was updated after
// the provided start time. This can be used to ensure the ProvisioningRequest is not only in the provided phase, but
// also has been updated after changes are made to the ProvisioningRequest.
//
// A zero-valued start time will cause the start time to be ignored and the function will wait until the
// ProvisioningRequest is in the provided phase regardless of the update time.
func (builder *ProvisioningRequestBuilder) WaitForPhaseAfter(
	phase provisioningv1alpha1.ProvisioningPhase, start time.Time, timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Waiting up to %s until ProvisioningRequest %s is fulfilled after %s",
		timeout, builder.Definition.Name, start)

	if !builder.Exists() {
		klog.V(100).Infof("ProvisioningRequest %s does not exist", builder.Definition.Name)

		return fmt.Errorf("cannot wait for non-existent ProvisioningRequest")
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("Failed to get ProvisioningRequest %s: %v", builder.Definition.Name, err)

				return false, nil
			}

			builder.Definition = builder.Object

			inPhase := builder.Definition.Status.ProvisioningStatus.ProvisioningPhase == phase
			updatedAfterStart := builder.Definition.Status.ProvisioningStatus.UpdateTime.After(start)

			return inPhase && (updatedAfterStart || start.IsZero()), nil
		})
	if err != nil {
		return err
	}

	return nil
}
