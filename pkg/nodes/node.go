package nodes

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/drain"
)

const (
	isTrue = "True"
)

// AdditionalOptions additional options for node object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// Builder provides struct for Node object containing connection to the cluster and the list of Node definitions.
type Builder struct {
	common.EmbeddableBuilder[corev1.Node, *corev1.Node]
	common.EmbeddableUpdater[corev1.Node, Builder, *corev1.Node, *Builder]
	common.EmbeddableDeleter[corev1.Node, *corev1.Node]
	common.EmbeddableWithOptions[corev1.Node, Builder, *corev1.Node, *Builder, AdditionalOptions]

	kubeClient  kubernetes.Interface
	drainHelper *drain.Helper
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableUpdater.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the Node GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("Node")
}

// NewBuilder creates a new instance of a Node builder.
func NewBuilder(apiClient *clients.Settings, nodeName string) *Builder {
	builder := common.NewClusterScopedBuilder[corev1.Node, Builder](apiClient, corev1.AddToScheme, nodeName)
	builder.setKubeClient(apiClient)

	return builder
}

// Pull gathers existing node from cluster.
func Pull(apiClient *clients.Settings, nodeName string) (*Builder, error) {
	builder, err := common.PullClusterScopedBuilder[corev1.Node, Builder](
		context.TODO(), apiClient, corev1.AddToScheme, nodeName)
	if err != nil {
		return nil, err
	}

	builder.setKubeClient(apiClient)

	return builder, nil
}

// SetDrainHelper builds drain Helper that contains parameters to control the behaviour of drain.
func (builder *Builder) SetDrainHelper(
	force bool,
	ignoreDaemonsets bool,
	deleteLocalData bool,
	gracePeriod int,
	skipWaitForDeleteTimeoutSeconds int,
	timeout time.Duration,
) {
	klog.V(100).Info("Creating new DrainOptions config")

	msg := fmt.Sprintf("Node draining configuration: 'force': %v,", force)
	msg += fmt.Sprintf(" 'gracePeriod': %d seconds,", gracePeriod)
	msg += fmt.Sprintf(" 'skipWaitForDeletionTimeout': %d seconds,", skipWaitForDeleteTimeoutSeconds)
	msg += fmt.Sprintf(" 'ignoreAllDaemonSets': %v,", ignoreDaemonsets)
	msg += fmt.Sprintf(" 'timeout': %v,", timeout)
	msg += fmt.Sprintf(" 'deleteEmptyDir': %v", deleteLocalData)

	klog.V(100).Infof("%v", msg)

	builder.drainHelper = &drain.Helper{
		Ctx:    logging.DiscardContext(),
		Client: builder.kubeClient,
		// Delete pods that do not declare a controller.
		Force: force,
		// GracePeriodSeconds is how long to wait for a pod to terminate.
		GracePeriodSeconds: gracePeriod,
		// Ignore DaemonSet-managed pods
		IgnoreAllDaemonSets: ignoreDaemonsets,
		// The length of time to wait before giving up
		Timeout: timeout,
		// Local data from emptyDir volumes will be deleted
		// when the node is drained
		DeleteEmptyDirData: deleteLocalData,
		Out:                os.Stdout,
		ErrOut:             os.Stderr,
		// If pod DeletionTimestamp older than N seconds, skip waiting for the pod.
		SkipWaitForDeleteTimeoutSeconds: skipWaitForDeleteTimeoutSeconds,
	}
}

// Drain evicts or deletes all pods.
func (builder *Builder) Drain() error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	builder.ensureDrainHelperIsSet()
	klog.V(100).Infof("Draining node %s", builder.Definition.Name)

	// Workaround for kubernetes/kubernetes#135512. Set logging verbosity to 0 while draining to avoid logging the
	// DaemonSet get requests at maximum verbosity.
	if verbosity := flag.Lookup("v"); verbosity != nil {
		old := verbosity.Value.String()

		_ = verbosity.Value.Set("0")

		defer func() {
			_ = verbosity.Value.Set(old)
		}()
	}

	return drain.RunNodeDrain(builder.drainHelper, builder.Definition.Name)
}

// Cordon marks node as unschedulable.
func (builder *Builder) Cordon() error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	builder.ensureDrainHelperIsSet()
	klog.V(100).Infof("Cordoning node %s", builder.Definition.Name)

	return drain.RunCordonOrUncordon(builder.drainHelper, builder.Definition, true)
}

// Uncordon marks node as schedulable.
func (builder *Builder) Uncordon() error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	builder.ensureDrainHelperIsSet()
	klog.V(100).Infof("Uncordoning node %s", builder.Definition.Name)

	return drain.RunCordonOrUncordon(builder.drainHelper, builder.Definition, false)
}

// WithNewLabel defines the new label placed in the Node metadata.
func (builder *Builder) WithNewLabel(key, value string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding label %s=%s to node %s ", key, value, builder.Definition.Name)

	if key == "" {
		klog.V(100).Infof("Failed to apply label with an empty key to node %s", builder.Definition.Name)
		builder.SetError(fmt.Errorf("error to set empty key to node"))

		return builder
	}

	if builder.Definition.Labels == nil {
		builder.Definition.Labels = map[string]string{key: value}
	} else {
		_, labelExist := builder.Definition.Labels[key]
		if !labelExist {
			builder.Definition.Labels[key] = value
		} else {
			builder.SetError(fmt.Errorf("cannot overwrite existing node label: %s", key))

			return builder
		}
	}

	return builder
}

// RemoveLabel removes given label from Node metadata.
func (builder *Builder) RemoveLabel(key, value string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Removing label %s=%s from node %s", key, value, builder.Definition.Name)

	if key == "" {
		klog.V(100).Infof("Failed to remove empty label's key from node %s", builder.Definition.Name)
		builder.SetError(fmt.Errorf("error to remove empty key from node"))

		return builder
	}

	delete(builder.Definition.Labels, key)

	return builder
}

// ExternalIPv4Network returns nodes external ip address.
func (builder *Builder) ExternalIPv4Network() (string, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Info("Collecting node's external ipv4 addresses")

	if builder.Object == nil {
		return "", fmt.Errorf("cannot collect external networks when node object is nil")
	}

	if _, ok := builder.Object.Annotations[ovnExternalAddresses]; !ok {
		return "", fmt.Errorf("node %s does not have external addresses annotation", builder.Definition.Name)
	}

	var extNetwork ExternalNetworks

	err := json.Unmarshal([]byte(builder.Object.Annotations[ovnExternalAddresses]), &extNetwork)
	if err != nil {
		return "",
			fmt.Errorf("error to unmarshal node %s, annotation %s due to %w", builder.Object.Name, ovnExternalAddresses, err)
	}

	return extNetwork.IPv4, nil
}

// ExternalIPv6Network returns nodes external ipv6 address.
func (builder *Builder) ExternalIPv6Network() (string, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Info("Collecting node's external ipv6 addresses")

	if builder.Object == nil {
		return "", fmt.Errorf("cannot collect external networks when node object is nil")
	}

	if _, ok := builder.Object.Annotations[ovnExternalAddresses]; !ok {
		return "", fmt.Errorf("node %s does not have external addresses annotation", builder.Definition.Name)
	}

	var extNetwork ExternalNetworks

	err := json.Unmarshal([]byte(builder.Object.Annotations[ovnExternalAddresses]), &extNetwork)
	if err != nil {
		return "",
			fmt.Errorf("error to unmarshal node %s, annotation %s due to %w", builder.Object.Name, ovnExternalAddresses, err)
	}

	if extNetwork.IPv6 == "" {
		return "", fmt.Errorf("node %s does not have external ipv6 addresses", builder.Definition.Name)
	}

	return extNetwork.IPv6, nil
}

// IsReady check if the Node is Ready.
func (builder *Builder) IsReady() (bool, error) {
	if err := common.Validate(builder); err != nil {
		return false, err
	}

	klog.V(100).Infof("Verify %s node availability", builder.Definition.Name)

	if !builder.Exists() {
		return false, fmt.Errorf("node object %s does not exist", builder.Definition.Name)
	}

	for _, condition := range builder.Object.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == isTrue, nil
		}
	}

	return false, fmt.Errorf("the Ready condition could not be found for node %s", builder.Definition.Name)
}

// WaitForCondition waits until the node has a condition that matches the expected, checking only the Type, Status,
// Reason, and Message fields. For the message field, it matches if the message contains the expected. Zero fields in
// the expected condition are ignored.
func (builder *Builder) WaitForCondition(expected corev1.NodeCondition, timeout time.Duration) (*Builder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("failed to get node %q, retrying: %v", builder.Definition.Name, err)

				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if (expected.Type == "" || condition.Type == expected.Type) &&
					(expected.Status == "" || condition.Status == expected.Status) &&
					(expected.Reason == "" || condition.Reason == expected.Reason) &&
					(expected.Message == "" || strings.Contains(condition.Message, expected.Message)) {
					return true, nil
				}
			}

			return false, nil
		})

	return builder, err
}

// WaitUntilConditionTrue waits for timeout duration or until node gets to a specific status.
func (builder *Builder) WaitUntilConditionTrue(conditionType corev1.NodeConditionType, timeout time.Duration) error {
	_, err := builder.WaitForCondition(corev1.NodeCondition{Type: conditionType, Status: corev1.ConditionTrue}, timeout)

	return err
}

// WaitUntilReady waits for up to timeout until the node's Ready condition is True.
func (builder *Builder) WaitUntilReady(timeout time.Duration) error {
	_, err := builder.WaitForCondition(corev1.NodeCondition{Type: corev1.NodeReady, Status: corev1.ConditionTrue}, timeout)

	return err
}

// WaitUntilNotReady waits for up to timeout until the node's Ready condition is not True (i.e., False or Unknown).
func (builder *Builder) WaitUntilNotReady(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("failed to get node %q, retrying: %v", builder.Definition.Name, err)

				return false, nil
			}

			builder.Definition = builder.Object

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Type == corev1.NodeReady {
					return condition.Status != corev1.ConditionTrue, nil
				}
			}

			return false, nil
		})
}

// ensureDrainHelperIsSet ensures that drainHelper is always set.
func (builder *Builder) ensureDrainHelperIsSet() {
	if builder.drainHelper == nil {
		klog.V(100).Infof(
			"DrainHelper is not initialized for node %s. Init DrainHelper with defaul parameters",
			builder.Definition.Name)
		builder.SetDrainHelper(true, true, true, 300, 180, 10*time.Minute)
	}
}

func (builder *Builder) setKubeClient(apiClient *clients.Settings) {
	if apiClient != nil {
		builder.kubeClient = apiClient.K8sClient
	}
}
