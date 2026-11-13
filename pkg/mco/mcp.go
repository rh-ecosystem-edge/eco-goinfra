package mco

import (
	"context"
	"fmt"
	"time"

	mcv1 "github.com/openshift/api/machineconfiguration/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const fiveScds = 5 * time.Second

var mcpGVK = mcv1.GroupVersion.WithKind("MachineConfigPool")

// MCPBuilder provides struct for MachineConfigPool object which contains connection to cluster
// and MachineConfigPool definitions.
type MCPBuilder struct {
	common.EmbeddableBuilder[mcv1.MachineConfigPool, *mcv1.MachineConfigPool]
	common.EmbeddableCreator[
		mcv1.MachineConfigPool, MCPBuilder, *mcv1.MachineConfigPool, *MCPBuilder]
	common.EmbeddableDeleter[mcv1.MachineConfigPool, *mcv1.MachineConfigPool]
	common.EmbeddableWithOptions[
		mcv1.MachineConfigPool, MCPBuilder, *mcv1.MachineConfigPool, *MCPBuilder, MCPAdditionalOptions]
}

// MCPAdditionalOptions additional options for mcp object.
type MCPAdditionalOptions func(builder *MCPBuilder) (*MCPBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *MCPBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the MachineConfigPool GVK for this builder.
func (builder *MCPBuilder) GetGVK() schema.GroupVersionKind {
	return mcpGVK
}

// NewMCPBuilder method creates new instance of builder.
func NewMCPBuilder(apiClient *clients.Settings, mcpName string) *MCPBuilder {
	return common.NewClusterScopedBuilder[mcv1.MachineConfigPool, MCPBuilder](
		apiClient, mcv1.Install, mcpName)
}

// Pull pulls existing machineconfigpool from cluster.
func Pull(apiClient *clients.Settings, name string) (*MCPBuilder, error) {
	return common.PullClusterScopedBuilder[mcv1.MachineConfigPool, MCPBuilder](
		context.TODO(), apiClient, mcv1.Install, name)
}

// WithMcSelector defines the machineConfigSelector in the machine config pool.
func (builder *MCPBuilder) WithMcSelector(mcSelector map[string]string) *MCPBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("WithMcSelector updates builder object with "+
		"machineConfigSelector label: %v", mcSelector)

	if len(mcSelector) == 0 {
		builder.SetError(fmt.Errorf("machineConfigSelector 'MatchLabels' field cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.MachineConfigSelector == nil {
		builder.Definition.Spec.MachineConfigSelector = &metav1.LabelSelector{}
	}

	builder.Definition.Spec.MachineConfigSelector.MatchLabels = mcSelector

	return builder
}

// WaitToBeInCondition waits for a specific time duration until the MachineConfigPool will have a
// specified condition type with the expected status.
func (builder *MCPBuilder) WaitToBeInCondition(
	conditionType mcv1.MachineConfigPoolConditionType,
	conditionStatus corev1.ConditionStatus,
	timeout time.Duration,
) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("WaitToBeInCondition waits up to specified time duration %v until "+
		"MachineConfigPool condition %v is met", timeout, conditionType)

	return wait.PollUntilContextTimeout(
		context.TODO(), fiveScds, timeout, true, func(ctx context.Context) (bool, error) {
			mcp, err := builder.Get()
			if err != nil {
				return false, nil
			}

			for _, condition := range mcp.Status.Conditions {
				if condition.Type == conditionType && condition.Status == conditionStatus {
					return true, nil
				}
			}

			return false, nil
		})
}

// WaitForUpdate waits for a MachineConfigPool to be updating and then updated.
func (builder *MCPBuilder) WaitForUpdate(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("WaitForUpdate waits up to specified time %v until updating"+
		" machineConfigPool object is updated", timeout)

	mcpUpdating, err := builder.Get()
	if err != nil {
		return err
	}

	for _, condition := range mcpUpdating.Status.Conditions {
		if condition.Type == "Updating" && condition.Status == corev1.ConditionTrue {
			err := wait.PollUntilContextTimeout(
				context.TODO(), fiveScds, timeout, true, func(ctx context.Context) (bool, error) {
					mcpUpdated, err := builder.Get()
					if err != nil {
						return false, nil
					}

					for _, condition := range mcpUpdated.Status.Conditions {
						if condition.Type == "Updated" && condition.Status == corev1.ConditionTrue {
							return true, nil
						}
					}

					return false, nil
				})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// WaitToBeStableFor waits on MachineConfigPool to stable for a time duration or until timeout.
func (builder *MCPBuilder) WaitToBeStableFor(stableDuration time.Duration, timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("WaitToBeStableFor waits up to duration of %v for "+
		"MachineConfigPool to be stable for %v", timeout, stableDuration)

	isMcpStable := true

	// Wait 5 secs in each iteration before condition function () returns true or errors
	// or times out after stableDuration
	err := wait.PollUntilContextTimeout(
		context.TODO(), fiveScds, timeout, true, func(ctx context.Context) (bool, error) {
			isMcpStable = true

			_ = wait.PollUntilContextTimeout(
				context.TODO(), fiveScds, stableDuration, true, func(ctx2 context.Context) (done bool, err error) {
					if !builder.Exists() {
						return false, nil
					}

					if builder.Object.Status.ReadyMachineCount != builder.Object.Status.MachineCount ||
						builder.Object.Status.MachineCount != builder.Object.Status.UpdatedMachineCount ||
						builder.Object.Status.DegradedMachineCount != 0 {
						klog.V(100).Infof("MachineConfigPool: %v degraded and has a mismatch in "+
							"machineCount: %v "+"vs machineCountUpdated: "+"%v vs readyMachineCount: %v and "+
							"degradedMachineCount is : %v \n", builder.Object.Name,
							builder.Object.Status.MachineCount, builder.Object.Status.UpdatedMachineCount,
							builder.Object.Status.ReadyMachineCount, builder.Object.Status.DegradedMachineCount)

						isMcpStable = false

						return true, nil
					}

					return false, nil
				})

			if isMcpStable {
				klog.V(100).Infof("MachineConfigPool was stable during during stableDuration: %v",
					stableDuration)

				// this will exit the outer wait.PollUntilContextTimeout block since the mcp was stable during stableDuration
				return true, nil
			}

			klog.V(100).Infof("MachineConfigPool was not stable during stableDuration: %v, retrying ...",
				stableDuration)

			// keep iterating in the outer wait.PollUntilContextTimeout waiting for cluster to be stable
			return false, nil
		})

	// After the timout in outer wait.PollUntilContextTimeout.
	if err == nil {
		klog.V(100).Infof("Cluster was stable during stableDuration: %v", stableDuration)
	} else {
		// Here err is "timed out waiting for the condition"
		klog.V(100).Infof("Cluster was Un-stable during stableDuration: %v", stableDuration)
	}

	return err
}

// IsInCondition parses MachineConfigPool conditions.
// Returns true if given MachineConfigPool is in given condition, otherwise false.
func (builder *MCPBuilder) IsInCondition(mcpConditionType mcv1.MachineConfigPoolConditionType) bool {
	if err := common.Validate(builder); err != nil {
		return false
	}

	klog.V(100).Infof("IsInCondition returns true"+
		" if MachineConfigPool object is in a given condition %v, otherwise false", mcpConditionType)

	if builder.Exists() {
		for _, condition := range builder.Object.Status.Conditions {
			if condition.Type == mcpConditionType && condition.Status == corev1.ConditionTrue {
				return true
			}
		}
	}

	return false
}
