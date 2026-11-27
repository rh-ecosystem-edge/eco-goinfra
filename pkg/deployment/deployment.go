package deployment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	multus "gopkg.in/k8snetworkplumbingwg/multus-cni.v4/pkg/types"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder provides struct for deployment object containing connection to the cluster and the deployment definitions.
type Builder struct {
	common.EmbeddableBuilder[appsv1.Deployment, *appsv1.Deployment]
	common.EmbeddableWithOptions[appsv1.Deployment, Builder, *appsv1.Deployment, *Builder, AdditionalOptions]
	common.EmbeddableCreator[appsv1.Deployment, Builder, *appsv1.Deployment, *Builder]
	common.EmbeddableDeleter[appsv1.Deployment, *appsv1.Deployment]
	common.EmbeddableUpdater[appsv1.Deployment, Builder, *appsv1.Deployment, *Builder]
}

// AdditionalOptions additional options for deployment object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableWithOptions.SetBase(builder)
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the Deployment GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return appsv1.SchemeGroupVersion.WithKind("Deployment")
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(
	apiClient *clients.Settings, name, nsname string, labels map[string]string, containerSpec corev1.Container) *Builder {
	klog.V(100).Infof(
		"Initializing new deployment structure with the following params: "+
			"name: %s, namespace: %s, labels: %s, containerSpec %v",
		name, nsname, labels, containerSpec)

	builder := common.NewNamespacedBuilder[appsv1.Deployment, Builder](apiClient, appsv1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	builder.Definition.Spec.Template.Labels = labels

	builder.WithAdditionalContainerSpecs([]corev1.Container{containerSpec})

	if len(labels) == 0 {
		klog.V(100).Info("There are no labels for the deployment")

		builder.SetError(fmt.Errorf("deployment 'labels' cannot be empty"))

		return builder
	}

	return builder
}

// Pull loads an existing deployment into Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	klog.V(100).Infof("Pulling existing deployment name: %s under namespace: %s", name, nsname)

	return common.PullNamespacedBuilder[appsv1.Deployment, Builder](
		context.TODO(), apiClient, appsv1.AddToScheme, name, nsname)
}

// WithNodeSelector applies a nodeSelector to the deployment definition.
func (builder *Builder) WithNodeSelector(selector map[string]string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying nodeSelector %s to deployment %s in namespace %s",
		selector, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.NodeSelector = selector

	return builder
}

// WithReplicas sets the desired number of replicas in the deployment definition.
func (builder *Builder) WithReplicas(replicas int32) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting %d replicas in deployment %s in namespace %s",
		replicas, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Replicas = &replicas

	return builder
}

// WithAdditionalContainerSpecs appends a list of container specs to the deployment definition.
func (builder *Builder) WithAdditionalContainerSpecs(specs []corev1.Container) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Appending a list of container specs %v to deployment %s in namespace %s",
		specs, builder.Definition.Name, builder.Definition.Namespace)

	if len(specs) == 0 {
		klog.V(100).Info("The container specs are empty")

		builder.SetError(fmt.Errorf("cannot accept empty list as container specs"))

		return builder
	}

	builder.Definition.Spec.Template.Spec.Containers = append(builder.Definition.Spec.Template.Spec.Containers, specs...)

	return builder
}

// WithSecondaryNetwork applies Multus secondary network configuration on deployment definition.
func (builder *Builder) WithSecondaryNetwork(networks []*multus.NetworkSelectionElement) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying secondary networks %v to deployment %s", networks, builder.Definition.Name)

	if len(networks) == 0 {
		builder.SetError(fmt.Errorf("can not apply empty networks list"))

		return builder
	}

	netAnnotation, err := json.Marshal(networks)
	if err != nil {
		builder.SetError(fmt.Errorf("error to unmarshal networks annotation due to: %s", err.Error()))

		return builder
	}

	builder.Definition.Spec.Template.Annotations = map[string]string{
		"k8s.v1.cni.cncf.io/networks": string(netAnnotation)}

	return builder
}

// WithHugePages sets hugePages on all containers inside the deployment.
func (builder *Builder) WithHugePages() *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying hugePages configuration to all containers in deployment: %s",
		builder.Definition.Name)

	// Append hugepages volume to the deployment.
	builder.Definition.Spec.Template.Spec.Volumes = append(builder.Definition.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "hugepages", VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{Medium: "HugePages"}}})

	for idx := range builder.Definition.Spec.Template.Spec.Containers {
		// Append hugepages volume mount to the deployment.
		builder.Definition.Spec.Template.Spec.Containers[idx].VolumeMounts = append(
			builder.Definition.Spec.Template.Spec.Containers[idx].VolumeMounts,
			corev1.VolumeMount{Name: "hugepages", MountPath: "/mnt/huge"})
	}

	return builder
}

// WithSecurityContext sets SecurityContext on deployment definition.
func (builder *Builder) WithSecurityContext(securityContext *corev1.PodSecurityContext) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying SecurityContext configuration on deployment %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if securityContext == nil {
		klog.V(100).Info("The 'securityContext' of the deployment is empty")

		builder.SetError(fmt.Errorf("'securityContext' parameter is empty"))

		return builder
	}

	builder.Definition.Spec.Template.Spec.SecurityContext = securityContext

	return builder
}

// WithLabel applies label to deployment's definition.
func (builder *Builder) WithLabel(labelKey, labelValue string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%v", fmt.Sprintf("Defining deployment's label to %s:%s", labelKey, labelValue))

	if labelKey == "" {
		klog.V(100).Info("The 'labelKey' of the deployment is empty")

		builder.SetError(fmt.Errorf("can not apply empty labelKey"))

		return builder
	}

	if builder.Definition.Spec.Template.Labels == nil {
		builder.Definition.Spec.Template.Labels = map[string]string{}
	}

	builder.Definition.Spec.Template.Labels[labelKey] = labelValue

	return builder
}

// WithServiceAccountName sets the ServiceAccountName on deployment definition.
func (builder *Builder) WithServiceAccountName(serviceAccountName string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting ServiceAccount %s on deployment %s in namespace %s",
		serviceAccountName, builder.Definition.Name, builder.Definition.Namespace)

	if serviceAccountName == "" {
		klog.V(100).Info("The 'serviceAccount' of the deployment is empty")

		builder.SetError(fmt.Errorf("can not apply empty serviceAccount"))

		return builder
	}

	builder.Definition.Spec.Template.Spec.ServiceAccountName = serviceAccountName

	return builder
}

// WithVolume attaches given volume to the deployment.
func (builder *Builder) WithVolume(deployVolume corev1.Volume) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if deployVolume.Name == "" {
		klog.V(100).Info("The volume's name cannot be empty")

		builder.SetError(fmt.Errorf("the volume's name cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Adding volume %s to deployment %s in namespace %s",
		deployVolume.Name, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.Volumes = append(
		builder.Definition.Spec.Template.Spec.Volumes,
		deployVolume)

	return builder
}

// WithSchedulerName configures a scheduler to process pod's scheduling.
func (builder *Builder) WithSchedulerName(schedulerName string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if schedulerName == "" {
		klog.V(100).Info("Scheduler's name cannot be empty")

		builder.SetError(fmt.Errorf("scheduler's name cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting scheduler %s for deployment %s in namespace %s",
		schedulerName, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.SchedulerName = schedulerName

	return builder
}

// WithAffinity applies Affinity to the deployment definition.
func (builder *Builder) WithAffinity(affinity *corev1.Affinity) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if affinity == nil {
		klog.V(100).Info("The Affinity parameter is empty")

		builder.SetError(fmt.Errorf("affinity parameter is empty"))

		return builder
	}

	klog.V(100).Infof("Adding affinity to deployment %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.Affinity = affinity

	return builder
}

// WithHostNetwork applies a hostnetwork state to the deployment definition.
func (builder *Builder) WithHostNetwork(enableHostnetwork bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting hostnetwork %v to deployment %s in namespace %s",
		enableHostnetwork, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.HostNetwork = enableHostnetwork

	return builder
}

// DeleteGraceful removes a deployment while waiting for specified duration(in seconds)
// the object should be deleted.
func (builder *Builder) DeleteGraceful(gracePeriod *int64) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	switch {
	case gracePeriod == nil:
		klog.V(100).Info("gracePeriod cannot be nil")

		return fmt.Errorf("gracePeriod cannot be nil")
	case *gracePeriod < int64(0):
		klog.V(100).Infof("gracePeriod(%v) must be non-negative integer", gracePeriod)

		return fmt.Errorf("gracePeriod must be non-negative integer")
	}

	klog.V(100).Infof("Deleting deployment %s in namespace %s with %v seconds grace period",
		builder.Definition.Name, builder.Definition.Namespace, *gracePeriod)

	if !builder.Exists() {
		klog.V(100).Infof("Deployment %s in namespace %s does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.SetObject(nil)

		return nil
	}

	err := builder.GetClient().Delete(
		logging.DiscardContext(), builder.Definition, &runtimeclient.DeleteOptions{GracePeriodSeconds: gracePeriod})
	if err != nil {
		return err
	}

	builder.SetObject(nil)

	return nil
}

// CreateAndWaitUntilReady creates a deployment in the cluster and waits until the deployment is available.
func (builder *Builder) CreateAndWaitUntilReady(timeout time.Duration) (*Builder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Creating deployment %s in namespace %s and waiting for the defined period until it is ready",
		builder.Definition.Name, builder.Definition.Namespace)

	if _, err := builder.Create(); err != nil {
		klog.V(100).Infof("Failed to create deployment. Error is: '%s'", err.Error())

		return nil, err
	}

	if builder.IsReady(timeout) {
		return builder, nil
	}

	return nil, fmt.Errorf("deployment %s in namespace %s is not ready",
		builder.Definition.Name, builder.Definition.Namespace,
	)
}

// IsReady periodically checks if deployment is in ready status.
func (builder *Builder) IsReady(timeout time.Duration) bool {
	if err := common.Validate(builder); err != nil {
		return false
	}

	klog.V(100).Infof("Running periodic check until deployment %s in namespace %s is ready",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return false
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			// Add Discard logger to suppress verbose Kubernetes client logging.
			// This preserves the behavior from logging.DiscardContext() while allowing
			// us to add timeout and respect parent cancellation.
			ctx = logr.NewContext(ctx, logr.Discard())

			// Use context with 120s timeout for individual GET requests.
			// This prevents failures when API server is slow (e.g., post-reboot)
			// while still respecting the overall timeout from PollUntilContextTimeout.
			getCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
			defer cancel()

			deployment := &appsv1.Deployment{}

			err := builder.GetClient().Get(getCtx, runtimeclient.ObjectKeyFromObject(builder.Definition), deployment)
			if err != nil {
				klog.V(100).Infof("Failed to get deployment from cluster. Error is: '%s'", err.Error())

				// Retry on timeout errors instead of failing immediately.
				// This handles transient API server performance degradation and context timeouts.
				if k8serrors.IsTimeout(err) || k8serrors.IsServerTimeout(err) || errors.Is(err, context.DeadlineExceeded) {
					klog.V(100).Infof("API server timeout when getting deployment %s in namespace %s, will retry",
						builder.Definition.Name, builder.Definition.Namespace)

					return false, nil
				}

				return false, err
			}

			builder.SetObject(deployment)

			if deployment.Status.ReadyReplicas > 0 && deployment.Status.Replicas == deployment.Status.ReadyReplicas {
				return true, nil
			}

			return false, nil
		})

	return err == nil
}

// DeleteAndWait deletes a deployment and waits until it is removed from the cluster.
func (builder *Builder) DeleteAndWait(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Deleting deployment %s in namespace %s and waiting for the defined period until it is removed",
		builder.Definition.Name, builder.Definition.Namespace)

	if err := builder.Delete(); err != nil {
		return err
	}

	// Polls the deployment every second until it is removed.
	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			err := builder.GetClient().Get(logging.DiscardContext(), runtimeclient.ObjectKeyFromObject(builder.Definition),
				&appsv1.Deployment{})
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			return false, nil
		})
}

// WaitUntilCondition waits for the duration of the defined timeout or until the
// deployment gets to a specific condition.
func (builder *Builder) WaitUntilCondition(condition appsv1.DeploymentConditionType, timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Waiting for the defined period until deployment %s in namespace %s has condition %v",
		builder.Definition.Name, builder.Definition.Namespace, condition)

	if !builder.Exists() {
		return fmt.Errorf("cannot wait for deployment condition because it does not exist")
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			updatedDeployment := &appsv1.Deployment{}

			err := builder.GetClient().Get(logging.DiscardContext(),
				runtimeclient.ObjectKeyFromObject(builder.Definition), updatedDeployment)
			if err != nil {
				return false, nil
			}

			for _, cond := range updatedDeployment.Status.Conditions {
				if cond.Type == condition && cond.Status == corev1.ConditionTrue {
					return true, nil
				}
			}

			return false, nil
		})
}

// WaitUntilDeleted waits for the duration of the defined timeout or until the deployment is deleted.
func (builder *Builder) WaitUntilDeleted(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Waiting for the defined period until deployment %s in namespace %s is deleted",
		builder.Definition.Name, builder.Definition.Namespace)

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			err := builder.GetClient().Get(logging.DiscardContext(), runtimeclient.ObjectKeyFromObject(builder.Definition),
				&appsv1.Deployment{})

			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			return false, nil
		})
}

// GetGVR returns deployment's GroupVersionResource which could be used for Clean function.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
}

// WithToleration applies a toleration to the deployment's definition.
func (builder *Builder) WithToleration(toleration corev1.Toleration) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if toleration == (corev1.Toleration{}) {
		klog.V(100).Info("The toleration cannot be empty")

		builder.SetError(fmt.Errorf("the toleration cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Adding TaintToleration %v to deployment %s in namespace %s",
		toleration, builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.Tolerations = append(
		builder.Definition.Spec.Template.Spec.Tolerations,
		toleration)

	return builder
}

// WithTerminationGracePeriodSeconds configures TerminationGracePeriodSeconds on the pod template of the deployment.
func (builder *Builder) WithTerminationGracePeriodSeconds(terminationGracePeriodSeconds int64) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Applying terminationGracePeriodSeconds flag to the configuration of pod template of deployment: %s in namespace: %s",
		builder.Definition.Name, builder.Definition.Namespace)

	builder.Definition.Spec.Template.Spec.TerminationGracePeriodSeconds = &terminationGracePeriodSeconds

	return builder
}
