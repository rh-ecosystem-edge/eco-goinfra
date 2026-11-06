package ibgu

import (
	"context"
	"fmt"
	"strings"
	"time"

	lcav1 "github.com/openshift-kni/lifecycle-agent/api/imagebasedupgrade/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/imagebasedgroupupgrades/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var conditionComplete = metav1.Condition{Type: "Progressing", Status: metav1.ConditionFalse, Reason: "Completed"}

var ibguGVK = v1alpha1.SchemeGroupVersion.WithKind("ImageBasedGroupUpgrade")

// IbguBuilder provides struct for the ibgu object containing connection to
// the cluster and the ibgu definitions.
type IbguBuilder struct {
	common.EmbeddableBuilder[v1alpha1.ImageBasedGroupUpgrade, *v1alpha1.ImageBasedGroupUpgrade]
	common.EmbeddableCreator[
		v1alpha1.ImageBasedGroupUpgrade, IbguBuilder, *v1alpha1.ImageBasedGroupUpgrade, *IbguBuilder]
	common.EmbeddableDeleter[v1alpha1.ImageBasedGroupUpgrade, *v1alpha1.ImageBasedGroupUpgrade]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *IbguBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the ImageBasedGroupUpgrade GVK for this builder.
func (builder *IbguBuilder) GetGVK() schema.GroupVersionKind {
	return ibguGVK
}

// NewIbguBuilder creates a new instance of IbguBuilder.
func NewIbguBuilder(
	apiClient *clients.Settings,
	name string,
	nsname string) *IbguBuilder {
	builder := common.NewNamespacedBuilder[v1alpha1.ImageBasedGroupUpgrade, IbguBuilder](
		apiClient, v1alpha1.AddToScheme, name, nsname)
	builder.Definition.Spec.Plan = []v1alpha1.PlanItem{}

	return builder
}

// WithClusterLabelSelectors appends labels to the ibgu clusterLabelSelectors.
func (builder *IbguBuilder) WithClusterLabelSelectors(labels map[string]string) *IbguBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating IBGU with %v cluster label selector", labels)

	if len(labels) == 0 {
		klog.V(100).Info("The 'labels' of the IBGU is empty")

		builder.SetError(fmt.Errorf("can not apply empty cluster label selectors to the IBGU"))

		return builder
	}

	labelSelectors := []metav1.LabelSelector{
		{
			MatchLabels: labels,
		},
	}

	builder.Definition.Spec.ClusterLabelSelectors = labelSelectors

	return builder
}

// WithAutoRollbackOnFailure appends the AutoRollbackOnFailure InitMonitorTimeout to the ibuSpec.
func (builder *IbguBuilder) WithAutoRollbackOnFailure(initMonitorTimeout int) *IbguBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating IBGU with AutoRollbackOnFailure and InitMonitorTimeout set to %d seconds",
		initMonitorTimeout)

	if initMonitorTimeout < 0 {
		klog.V(100).Info("The 'initMonitorTimeout' parameter is undefined")

		builder.SetError(fmt.Errorf("initMonitorTimeout cannot be undefined"))

		return builder
	}

	autoRollbackOnFailure := lcav1.AutoRollbackOnFailure{
		InitMonitorTimeoutSeconds: initMonitorTimeout,
	}

	builder.Definition.Spec.IBUSpec.AutoRollbackOnFailure = &autoRollbackOnFailure

	return builder
}

// WithSeedImageRef appends the SeedImageRef to the ibuSpec.
func (builder *IbguBuilder) WithSeedImageRef(seedImage string, seedVersion string) *IbguBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating IBGU with %s seed image and %s seed version", seedImage, seedVersion)

	if seedImage == "" {
		klog.V(100).Info("The 'seedImage' parameter is empty")

		builder.SetError(fmt.Errorf("seedImage cannot be empty"))

		return builder
	}

	if seedVersion == "" {
		klog.V(100).Info("The 'seedVersion' parameter is empty")

		builder.SetError(fmt.Errorf("seedVersion cannot be empty"))

		return builder
	}

	ibuSpec := lcav1.ImageBasedUpgradeSpec{
		SeedImageRef: lcav1.SeedImageRef{
			Image:   seedImage,
			Version: seedVersion,
		},
	}

	builder.Definition.Spec.IBUSpec = ibuSpec

	return builder
}

// WithOadpContent appends the oadpContent to the ibuSpec.
func (builder *IbguBuilder) WithOadpContent(name string, namespace string) *IbguBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Creating IBGU with OADP configmap %s in namespace %s", name, namespace)

	if name == "" {
		klog.V(100).Info("The 'name' parameter for OADP content is empty")

		builder.SetError(fmt.Errorf("oadp content name cannot be empty"))

		return builder
	}

	if namespace == "" {
		klog.V(100).Info("The 'namespace' parameter for OADP content is empty")

		builder.SetError(fmt.Errorf("oadp content namespace cannot be empty"))

		return builder
	}

	oadpContent := lcav1.ConfigMapRef{
		Name:      name,
		Namespace: namespace,
	}

	builder.Definition.Spec.IBUSpec.OADPContent = append(builder.Definition.Spec.IBUSpec.OADPContent, oadpContent)

	return builder
}

// WithPlan appends the plan to the ibgu.
func (builder *IbguBuilder) WithPlan(actions []string, maxConcurrency int, timeout int) *IbguBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating IBGU with plan actions %v, maxConcurrency %d and timeout %d",
		actions,
		maxConcurrency,
		timeout,
	)

	if len(actions) == 0 {
		klog.V(100).Info("The 'actions' slice is empty")

		builder.SetError(fmt.Errorf("plan actions cannot be empty"))

		return builder
	}

	if maxConcurrency <= 0 {
		klog.V(100).Infof("Invalid maxConcurrency value: %d", maxConcurrency)

		builder.SetError(fmt.Errorf("maxConcurrency must be greater than 0"))

		return builder
	}

	if timeout <= 0 {
		klog.V(100).Infof("Invalid timeout value: %d", timeout)

		builder.SetError(fmt.Errorf("timeout must be greater than 0"))

		return builder
	}

	plan := v1alpha1.PlanItem{
		Actions: actions,
		RolloutStrategy: v1alpha1.RolloutStrategy{
			MaxConcurrency: maxConcurrency,
			Timeout:        timeout,
		},
	}

	builder.Definition.Spec.Plan = append(builder.Definition.Spec.Plan, plan)

	return builder
}

// PullIbgu pulls existing ibgu into IbguBuilder struct.
func PullIbgu(apiClient *clients.Settings, name, nsname string) (*IbguBuilder, error) {
	return common.PullNamespacedBuilder[v1alpha1.ImageBasedGroupUpgrade, IbguBuilder](
		context.TODO(), apiClient, v1alpha1.AddToScheme, name, nsname)
}

// DeleteAndWait deletes the ibgu object and waits until the ibgu is deleted.
func (builder *IbguBuilder) DeleteAndWait(timeout time.Duration) (*IbguBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Deleting ibgu %s in namespace %s and waiting for the defined period until it is removed",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.Delete()
	if err != nil {
		return builder, err
	}

	err = builder.WaitUntilDeleted(timeout)

	return builder, err
}

// WaitUntilDeleted waits for the duration of the defined timeout or until the ibgu is deleted.
func (builder *IbguBuilder) WaitUntilDeleted(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof(
		"Waiting for the defined period until ibgu %s in namespace %s is deleted",
		builder.Definition.Name, builder.Definition.Namespace)

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			_, err := builder.Get()
			if err == nil {
				klog.V(100).Infof("ibgu %s/%s still present", builder.Definition.Namespace, builder.Definition.Name)

				return false, nil
			}

			if k8serrors.IsNotFound(err) {
				klog.V(100).Infof("ibgu %s/%s is gone", builder.Definition.Namespace, builder.Definition.Name)

				return true, nil
			}

			klog.V(100).Infof("failed to get ibgu %s/%s: %v", builder.Definition.Namespace, builder.Definition.Name, err)

			return false, err
		})
}

// WaitForCondition waits until the IBGU has a condition that matches the expected, checking only the Type, Status,
// Reason, and Message fields. For the message field, it matches if the message contains the expected. Zero fields in
// the expected condition are ignored.
func (builder *IbguBuilder) WaitForCondition(expected metav1.Condition, timeout time.Duration) (*IbguBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	if !builder.Exists() {
		klog.V(100).Info("The IBGU does not exist on the cluster")

		return builder, fmt.Errorf(
			"ibgu object %s does not exist in namespace %s", builder.Definition.Name, builder.Definition.Namespace)
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), 10*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("failed to get ibgu %s/%s: %v", builder.Definition.Namespace, builder.Definition.Name, err)

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

	return builder, err
}

// WaitUntilComplete waits the specified timeout for the IBGU to complete.
func (builder *IbguBuilder) WaitUntilComplete(timeout time.Duration) (*IbguBuilder, error) {
	return builder.WaitForCondition(conditionComplete, timeout)
}
