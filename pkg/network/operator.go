package network

import (
	"context"
	"errors"
	"fmt"
	"time"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

// OperatorBuilder provides a struct for network.operator object from the cluster and a network.operator definition.
type OperatorBuilder struct {
	common.EmbeddableBuilder[operatorv1.Network, *operatorv1.Network]
	common.EmbeddableUpdater[operatorv1.Network, OperatorBuilder, *operatorv1.Network, *OperatorBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *OperatorBuilder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the network.operator GVK for this builder.
func (builder *OperatorBuilder) GetGVK() schema.GroupVersionKind {
	return operatorv1.GroupVersion.WithKind("Network")
}

// PullOperator loads an existing network.operator into OperatorBuilder struct.
func PullOperator(apiClient *clients.Settings) (*OperatorBuilder, error) {
	return common.PullClusterScopedBuilder[operatorv1.Network, OperatorBuilder](
		context.TODO(), apiClient, operatorv1.Install, clusterNetworkName)
}

// SetLocalGWMode switches network.operator OVN mode from/to local mode.
func (builder *OperatorBuilder) SetLocalGWMode(state bool, timeout time.Duration) (*OperatorBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	var err error

	if builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.GatewayConfig.RoutingViaHost != state {
		builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.GatewayConfig.RoutingViaHost = state

		builder, err := builder.Update()
		if err != nil {
			return nil, err
		}

		err = builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeProgressing, 300*time.Second, operatorv1.ConditionTrue)
		if err != nil {
			return nil, err
		}

		err = builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeProgressing, timeout, operatorv1.ConditionFalse)
		if err != nil {
			return nil, err
		}

		return builder, builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeAvailable, 60*time.Second, operatorv1.ConditionTrue)
	}

	return builder, err
}

// SetIPForwarding sets the IPForwarding mode on the OVN-Kubernetes gateway configuration
// and waits for the network operator to stabilize.
func (builder *OperatorBuilder) SetIPForwarding(
	mode operatorv1.IPForwardingMode, timeout time.Duration) (*OperatorBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Setting IPForwarding mode %q on network.operator %s", mode, builder.Definition.Name)

	if builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig == nil {
		builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig = &operatorv1.OVNKubernetesConfig{}
	}

	if builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.GatewayConfig == nil {
		builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.GatewayConfig = &operatorv1.GatewayConfig{}
	}

	var err error

	if builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.GatewayConfig.IPForwarding != mode {
		builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.GatewayConfig.IPForwarding = mode

		builder, err := builder.Update()
		if err != nil {
			return nil, err
		}

		err = builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeProgressing, 300*time.Second, operatorv1.ConditionTrue)
		if err != nil {
			return nil, err
		}

		err = builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeProgressing, timeout, operatorv1.ConditionFalse)
		if err != nil {
			return nil, err
		}

		return builder, builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeAvailable, 60*time.Second, operatorv1.ConditionTrue)
	}

	return builder, err
}

// SetMultiNetworkPolicy sets network.operator spec.useMultiNetworkPolicy to state.
// No-op if unchanged; otherwise Update then wait until Available=True within timeout.
func (builder *OperatorBuilder) SetMultiNetworkPolicy(state bool, timeout time.Duration) (*OperatorBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Applying MultiNetworkPolicy flag %t to network.operator %s", state, builder.Definition.Name)

	var err error

	if *builder.Definition.Spec.UseMultiNetworkPolicy != state {
		builder.Definition.Spec.UseMultiNetworkPolicy = &state

		builder, err := builder.Update()
		if err != nil {
			return nil, err
		}

		return builder, builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeAvailable, timeout, operatorv1.ConditionTrue)
	}

	return builder, err
}

// SetRouteAdvertisements enables or disables RouteAdvertisements on the OVN-Kubernetes network configuration.
func (builder *OperatorBuilder) SetRouteAdvertisements(
	state operatorv1.RouteAdvertisementsEnablement, timeout time.Duration) (*OperatorBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Setting RouteAdvertisements to %s on network.operator %s", state, builder.Definition.Name)

	if builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig == nil {
		builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig = &operatorv1.OVNKubernetesConfig{}
	}

	if builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.RouteAdvertisements != state {
		builder.Definition.Spec.DefaultNetwork.OVNKubernetesConfig.RouteAdvertisements = state

		builder, err := builder.Update()
		if err != nil {
			return nil, err
		}

		err = builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeProgressing, 60*time.Second, operatorv1.ConditionTrue)
		if err != nil {
			return nil, err
		}

		err = builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeProgressing, timeout, operatorv1.ConditionFalse)
		if err != nil {
			return nil, err
		}

		return builder, builder.WaitUntilInCondition(
			operatorv1.OperatorStatusTypeAvailable, 60*time.Second, operatorv1.ConditionTrue)
	}

	return builder, nil
}

// WaitUntilInCondition waits for a specific time duration until the network.operator will have a
// specified condition type with the expected status.
func (builder *OperatorBuilder) WaitUntilInCondition(
	condition string, timeout time.Duration, status operatorv1.ConditionStatus) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Infof("Wait until network.operator object %s has condition %s=%s",
		builder.Definition.Name, condition, status)

	err := wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			obj, getErr := builder.Get()
			if getErr != nil {
				if k8serrors.IsNotFound(getErr) {
					return false, fmt.Errorf("network.operator object %s does not exist", builder.Definition.Name)
				}

				return false, nil
			}

			builder.Object = obj

			for _, c := range obj.Status.Conditions {
				if c.Type == condition && c.Status == status {
					return true, nil
				}
			}

			return false, nil
		})

	if err != nil && errors.Is(err, context.DeadlineExceeded) && builder.Object != nil {
		klog.V(100).Infof("timeout waiting for network.operator %s condition %s=%s; last status conditions: %#v",
			builder.Definition.Name, condition, status, builder.Object.Status.Conditions)
	}

	return err
}
