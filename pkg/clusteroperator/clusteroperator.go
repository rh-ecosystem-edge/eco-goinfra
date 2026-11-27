package clusteroperator

import (
	"context"
	"fmt"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	fiveScds time.Duration = 5 * time.Second
	isTrue                 = "True"
)

// Builder provides struct for clusterOperator object.
type Builder struct {
	common.EmbeddableBuilder[configv1.ClusterOperator, *configv1.ClusterOperator]
}

// GetGVK returns the ClusterOperator GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return configv1.GroupVersion.WithKind(APIKind)
}

// Pull loads an existing clusterOperator into Builder struct.
func Pull(apiClient *clients.Settings, clusterOperatorName string) (*Builder, error) {
	return common.PullClusterScopedBuilder[configv1.ClusterOperator, Builder](
		context.TODO(), apiClient, configv1.Install, clusterOperatorName)
}

// IsAvailable check if the clusterOperator is available.
func (builder *Builder) IsAvailable() bool {
	if !builder.Exists() {
		return false
	}

	klog.V(100).Infof("Verify the availability of %s clusterOperator", builder.Definition.Name)

	for _, condition := range builder.Object.Status.Conditions {
		if condition.Type == "Available" {
			return condition.Status == isTrue
		}
	}

	return false
}

// IsDegraded checks if the clusterOperator is degraded.
func (builder *Builder) IsDegraded() bool {
	if !builder.Exists() {
		return false
	}

	klog.V(100).Infof("Check if %s clusterOperator is degraded", builder.Definition.Name)

	for _, condition := range builder.Object.Status.Conditions {
		if condition.Type == "Degraded" {
			return condition.Status == isTrue
		}
	}

	return false
}

// IsProgressing checks if the clusterOperator is progressing.
func (builder *Builder) IsProgressing() bool {
	if !builder.Exists() {
		return false
	}

	klog.V(100).Infof("Check if %s clusterOperator is progressing", builder.Definition.Name)

	for _, condition := range builder.Object.Status.Conditions {
		if condition.Type == "Progressing" {
			return condition.Status == isTrue
		}
	}

	return false
}

// GetConditionReason returns the specific condition type's reason value or an empty string if it does not exist.
func (builder *Builder) GetConditionReason(conditionType configv1.ClusterStatusConditionType) string {
	if err := common.Validate(builder); err != nil {
		return ""
	}

	klog.V(100).Infof("Get %s clusterOperator %v condition reason if exists",
		builder.Definition.Name, conditionType)

	err := builder.WaitUntilConditionTrue(conditionType, time.Second)
	if err != nil {
		return ""
	}

	for _, condition := range builder.Object.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Reason
		}
	}

	return ""
}

// WaitUntilAvailable waits for timeout duration or until clusterOperator is Available.
func (builder *Builder) WaitUntilAvailable(timeout time.Duration) error {
	return builder.WaitUntilConditionTrue("Available", timeout)
}

// WaitUntilProgressing waits for timeout duration or until clusterOperator is Progressing.
func (builder *Builder) WaitUntilProgressing(timeout time.Duration) error {
	return builder.WaitUntilConditionTrue("Progressing", timeout)
}

// WaitUntilConditionTrue waits for timeout duration or until clusterOperator gets to a specific status.
func (builder *Builder) WaitUntilConditionTrue(
	conditionType configv1.ClusterStatusConditionType, timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	if !builder.Exists() {
		return fmt.Errorf("%s clusterOperator not found", builder.Definition.Name)
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Type == conditionType {
					return condition.Status == isTrue, nil
				}
			}

			return false, err
		})
}

// HasDesiredVersion checks if an operator has a desiredVersion.
func (builder *Builder) HasDesiredVersion(desiredVersion string) (bool, error) {
	if err := common.Validate(builder); err != nil {
		return false, err
	}

	if desiredVersion == "" {
		return false, fmt.Errorf("desiredVersion can't be empty")
	}

	if builder.Object.Status.Versions == nil {
		return false, fmt.Errorf("undefined cluster operator status versions")
	}

	for _, operatorVersion := range builder.Object.Status.Versions {
		klog.V(100).Infof("Testing %s and %s", operatorVersion.Version, desiredVersion)

		if operatorVersion.Version == desiredVersion {
			klog.V(100).Infof("The clusterOperator %s version matches the desired version %s",
				builder.Definition.Name, desiredVersion)

			return true, nil
		}

		klog.V(100).Infof(
			"The clusterOperator %s doesn't have the desired version %s, but has version %s. Continue lookup",
			builder.Definition.Name, desiredVersion, operatorVersion.Version)
	}

	klog.V(100).Infof("lookup failed.The clusterOperator %s doesn't have the desired version %s",
		builder.Definition.Name, desiredVersion)

	return false, nil
}
