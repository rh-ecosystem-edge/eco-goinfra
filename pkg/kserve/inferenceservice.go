package kserve

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	kservev1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kserve/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var inferenceServiceGVK = kservev1beta1.GroupVersion.WithKind("InferenceService")

// InferenceServiceBuilder provides struct for the InferenceService object
// containing connection to the cluster and the InferenceService definition.
type InferenceServiceBuilder struct {
	common.EmbeddableBuilder[kservev1beta1.InferenceService, *kservev1beta1.InferenceService]
	common.EmbeddableCreator[
		kservev1beta1.InferenceService,
		InferenceServiceBuilder,
		*kservev1beta1.InferenceService,
		*InferenceServiceBuilder,
	]
	common.EmbeddableDeleteReturner[
		kservev1beta1.InferenceService,
		InferenceServiceBuilder,
		*kservev1beta1.InferenceService,
		*InferenceServiceBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *InferenceServiceBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
}

// GetGVK returns the InferenceService GVK for this builder.
func (builder *InferenceServiceBuilder) GetGVK() schema.GroupVersionKind {
	return inferenceServiceGVK
}

// NewInferenceServiceBuilder creates a new instance of InferenceServiceBuilder.
func NewInferenceServiceBuilder(
	apiClient *clients.Settings,
	name, namespace string) *InferenceServiceBuilder {
	return common.NewNamespacedBuilder[kservev1beta1.InferenceService, InferenceServiceBuilder](
		apiClient, kservev1beta1.AddToScheme, name, namespace)
}

// WithModelFormat sets the model format name.
func (builder *InferenceServiceBuilder) WithModelFormat(name string) *InferenceServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting InferenceService %s model format: %s",
		builder.Definition.Name, name)

	if name == "" {
		builder.SetError(fmt.Errorf("InferenceService 'modelFormat' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Predictor.Model.ModelFormat = kservev1beta1.ModelFormat{Name: name}

	return builder
}

// WithRuntime sets the serving runtime name.
func (builder *InferenceServiceBuilder) WithRuntime(runtimeName string) *InferenceServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting InferenceService %s runtime: %s",
		builder.Definition.Name, runtimeName)

	if runtimeName == "" {
		builder.SetError(fmt.Errorf("InferenceService 'runtime' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Predictor.Model.Runtime = &runtimeName

	return builder
}

// WithStorageURI sets the model storage URI.
func (builder *InferenceServiceBuilder) WithStorageURI(uri string) *InferenceServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting InferenceService %s storageUri: %s",
		builder.Definition.Name, uri)

	if uri == "" {
		builder.SetError(fmt.Errorf("InferenceService 'storageUri' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Predictor.Model.StorageURI = &uri

	return builder
}

// WithServiceAccountName sets the predictor service account.
func (builder *InferenceServiceBuilder) WithServiceAccountName(serviceAccountName string) *InferenceServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting InferenceService %s serviceAccountName: %s",
		builder.Definition.Name, serviceAccountName)

	builder.Definition.Spec.Predictor.ServiceAccountName = serviceAccountName

	return builder
}

// WithResources sets the resource requests and limits.
func (builder *InferenceServiceBuilder) WithResources(
	requests, limits corev1.ResourceList) *InferenceServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting InferenceService %s resources",
		builder.Definition.Name)

	builder.Definition.Spec.Predictor.Model.Resources = corev1.ResourceRequirements{
		Requests: requests,
		Limits:   limits,
	}

	return builder
}

// WithNeuronResources sets Neuron device count and memory resources.
func (builder *InferenceServiceBuilder) WithNeuronResources(
	neuronDevices int64, memoryRequest, memoryLimit string) *InferenceServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting InferenceService %s Neuron resources: devices=%d",
		builder.Definition.Name, neuronDevices)

	neuronQuantity, err := resource.ParseQuantity(fmt.Sprintf("%d", neuronDevices))
	if err != nil {
		builder.SetError(fmt.Errorf("failed to parse neuron device quantity: %w", err))

		return builder
	}

	memReq, err := resource.ParseQuantity(memoryRequest)
	if err != nil {
		builder.SetError(fmt.Errorf("failed to parse memory request %q: %w", memoryRequest, err))

		return builder
	}

	memLim, err := resource.ParseQuantity(memoryLimit)
	if err != nil {
		builder.SetError(fmt.Errorf("failed to parse memory limit %q: %w", memoryLimit, err))

		return builder
	}

	builder.Definition.Spec.Predictor.Model.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"aws.amazon.com/neuron": neuronQuantity,
			corev1.ResourceMemory:   memReq,
		},
		Limits: corev1.ResourceList{
			"aws.amazon.com/neuron": neuronQuantity,
			corev1.ResourceMemory:   memLim,
		},
	}

	return builder
}

// WithAnnotation adds an annotation to the InferenceService.
func (builder *InferenceServiceBuilder) WithAnnotation(
	key, value string) *InferenceServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting InferenceService %s annotation %s=%s",
		builder.Definition.Name, key, value)

	if builder.Definition.Annotations == nil {
		builder.Definition.Annotations = make(map[string]string)
	}

	builder.Definition.Annotations[key] = value

	return builder
}

// WithEnv adds environment variables to the model spec.
func (builder *InferenceServiceBuilder) WithEnv(
	envVars []corev1.EnvVar) *InferenceServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting InferenceService %s env vars",
		builder.Definition.Name)

	builder.Definition.Spec.Predictor.Model.Env = append(
		builder.Definition.Spec.Predictor.Model.Env, envVars...)

	return builder
}

// PullInferenceService retrieves an existing InferenceService from the cluster.
func PullInferenceService(
	apiClient *clients.Settings, name, namespace string) (*InferenceServiceBuilder, error) {
	return common.PullNamespacedBuilder[kservev1beta1.InferenceService, InferenceServiceBuilder](
		context.TODO(), apiClient, kservev1beta1.AddToScheme, name, namespace)
}

// GetURL returns the inference URL from the InferenceService status.
func (builder *InferenceServiceBuilder) GetURL() (string, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	isvc, err := builder.Get()
	if err != nil {
		return "", err
	}

	if isvc.Status.URL == "" {
		return "", fmt.Errorf("inferenceService %s has no URL in status",
			builder.Definition.Name)
	}

	return isvc.Status.URL, nil
}

// IsReady checks if the InferenceService has a Ready condition with status True.
func (builder *InferenceServiceBuilder) IsReady() (bool, error) {
	if err := common.Validate(builder); err != nil {
		return false, err
	}

	isvc, err := builder.Get()
	if err != nil {
		return false, err
	}

	for _, condition := range isvc.Status.Conditions {
		if condition.Type == "Ready" && condition.Status == kservev1beta1.ConditionTrue {
			return true, nil
		}
	}

	return false, nil
}

// WaitUntilReady waits until the InferenceService reaches Ready state.
func (builder *InferenceServiceBuilder) WaitUntilReady(
	timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Waiting for InferenceService %s to become Ready (timeout: %s)",
		builder.Definition.Name, timeout)

	return wait.PollUntilContextTimeout(
		context.TODO(), 15*time.Second, timeout, true,
		func(ctx context.Context) (bool, error) {
			ready, err := builder.IsReady()
			if err != nil {
				klog.V(100).Infof("Error checking InferenceService readiness: %v", err)

				return false, nil
			}

			return ready, nil
		})
}
