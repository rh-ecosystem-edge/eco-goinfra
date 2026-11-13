package clusterlogging

import (
	"context"
	"fmt"
	"time"

	lokiv1 "github.com/grafana/loki/operator/apis/loki/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var lokiStackGVK = lokiv1.GroupVersion.WithKind("LokiStack")

// LokiStackBuilder provides struct for the lokiStack object.
type LokiStackBuilder struct {
	common.EmbeddableBuilder[lokiv1.LokiStack, *lokiv1.LokiStack]
	common.EmbeddableCreator[lokiv1.LokiStack, LokiStackBuilder, *lokiv1.LokiStack, *LokiStackBuilder]
	common.EmbeddableDeleteReturner[lokiv1.LokiStack, LokiStackBuilder, *lokiv1.LokiStack, *LokiStackBuilder]
	common.EmbeddableUpdater[lokiv1.LokiStack, LokiStackBuilder, *lokiv1.LokiStack, *LokiStackBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *LokiStackBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the LokiStack GVK for this builder.
func (builder *LokiStackBuilder) GetGVK() schema.GroupVersionKind {
	return lokiStackGVK
}

// NewLokiStackBuilder creates new instance of builder.
func NewLokiStackBuilder(apiClient *clients.Settings, name, nsname string) *LokiStackBuilder {
	return common.NewNamespacedBuilder[lokiv1.LokiStack, LokiStackBuilder](
		apiClient, lokiv1.AddToScheme, name, nsname)
}

// PullLokiStack retrieves an existing lokiStack object from the cluster.
func PullLokiStack(apiClient *clients.Settings, name, nsname string) (*LokiStackBuilder, error) {
	return common.PullNamespacedBuilder[lokiv1.LokiStack, LokiStackBuilder](
		context.TODO(), apiClient, lokiv1.AddToScheme, name, nsname)
}

// WithSize sets the lokiStack operator's size.
func (builder *LokiStackBuilder) WithSize(
	size lokiv1.LokiStackSizeType) *LokiStackBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting lokiStack %s in namespace %s with the size: %v",
		builder.Definition.Name, builder.Definition.Namespace, size)

	if size == "" {
		klog.V(100).Info("'size' argument cannot be empty")

		builder.SetError(fmt.Errorf("'size' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Size = size

	return builder
}

// WithStorage sets the lokiStack operator's storage configuration.
func (builder *LokiStackBuilder) WithStorage(
	storage lokiv1.ObjectStorageSpec) *LokiStackBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting lokiStack %s in namespace %s with the storage config: %v",
		builder.Definition.Name, builder.Definition.Namespace, storage)

	builder.Definition.Spec.Storage = storage

	return builder
}

// WithStorageClassName sets the lokiStack operator's storage class name configuration.
func (builder *LokiStackBuilder) WithStorageClassName(
	storageClassName string) *LokiStackBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting lokiStack %s in namespace %s with the storage class name config: %v",
		builder.Definition.Name, builder.Definition.Namespace, storageClassName)

	if storageClassName == "" {
		klog.V(100).Info("'storageClassName' argument cannot be empty")

		builder.SetError(fmt.Errorf("'storageClassName' argument cannot be empty"))

		return builder
	}

	builder.Definition.Spec.StorageClassName = storageClassName

	return builder
}

// WithTenants sets the lokiStack operator's tenants configuration.
func (builder *LokiStackBuilder) WithTenants(
	tenants lokiv1.TenantsSpec) *LokiStackBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting lokiStack %s in namespace %s with the tenants config: %v",
		builder.Definition.Name, builder.Definition.Namespace, tenants)

	builder.Definition.Spec.Tenants = &tenants

	return builder
}

// WithRules sets the lokiStack operator's rules configuration.
func (builder *LokiStackBuilder) WithRules(
	rules lokiv1.RulesSpec) *LokiStackBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting lokiStack %s in namespace %s with the rules config: %v",
		builder.Definition.Name, builder.Definition.Namespace, rules)

	builder.Definition.Spec.Rules = &rules

	return builder
}

// WithManagementState sets the lokiStack operator's rules configuration.
func (builder *LokiStackBuilder) WithManagementState(
	managementState lokiv1.ManagementStateType) *LokiStackBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting lokiStack %s in namespace %s with the managementState config: %v",
		builder.Definition.Name, builder.Definition.Namespace, managementState)

	builder.Definition.Spec.ManagementState = managementState

	return builder
}

// WithLimits sets the lokiStack operator's limits configuration.
func (builder *LokiStackBuilder) WithLimits(
	limits lokiv1.LimitsSpec) *LokiStackBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting lokiStack %s in namespace %s with the limits config: %v",
		builder.Definition.Name, builder.Definition.Namespace, limits)

	builder.Definition.Spec.Limits = &limits

	return builder
}

// WithTemplate sets the lokiStack operator's template configuration.
func (builder *LokiStackBuilder) WithTemplate(
	template lokiv1.LokiTemplateSpec) *LokiStackBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting lokiStack %s in namespace %s with the template config: %v",
		builder.Definition.Name, builder.Definition.Namespace, template)

	builder.Definition.Spec.Template = &template

	return builder
}

// IsReady checks for the duration of timeout if the lokiStack state is Ready.
func (builder *LokiStackBuilder) IsReady(timeout time.Duration) bool {
	if err := common.Validate(builder); err != nil {
		return false
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			if !builder.Exists() {
				return false, nil
			}

			for _, condition := range builder.Definition.Status.Conditions {
				if condition.Type == "Ready" {
					if condition.Status == metav1.ConditionTrue {
						return true, nil
					}
				}
			}

			return false, nil
		})

	return err == nil
}
