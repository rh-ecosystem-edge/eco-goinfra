package sriov

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

var poolConfigGVK = srIovV1.GroupVersion.WithKind("SriovNetworkPoolConfig")

// PoolConfigBuilder provides struct for SriovNetworkPoolConfig object containing connection to the cluster
// and the SriovNetworkPoolConfig definitions.
type PoolConfigBuilder struct {
	common.EmbeddableBuilder[srIovV1.SriovNetworkPoolConfig, *srIovV1.SriovNetworkPoolConfig]
	common.EmbeddableCreator[srIovV1.SriovNetworkPoolConfig, PoolConfigBuilder,
		*srIovV1.SriovNetworkPoolConfig, *PoolConfigBuilder]
	common.EmbeddableDeleter[srIovV1.SriovNetworkPoolConfig, *srIovV1.SriovNetworkPoolConfig]
	common.EmbeddableUpdater[srIovV1.SriovNetworkPoolConfig, PoolConfigBuilder,
		*srIovV1.SriovNetworkPoolConfig, *PoolConfigBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *PoolConfigBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the SriovNetworkPoolConfig GVK for this builder.
func (builder *PoolConfigBuilder) GetGVK() schema.GroupVersionKind {
	return poolConfigGVK
}

// NewPoolConfigBuilder creates a new instance of PoolConfigBuilder.
func NewPoolConfigBuilder(apiClient *clients.Settings, name, nsname string) *PoolConfigBuilder {
	return common.NewNamespacedBuilder[srIovV1.SriovNetworkPoolConfig, PoolConfigBuilder](
		apiClient, srIovV1.AddToScheme, name, nsname)
}

// WithNodeSelector sets nodeSelector in the SriovNetworkPoolConfig definition.
func (builder *PoolConfigBuilder) WithNodeSelector(nodeSelector map[string]string) *PoolConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating SriovNetworkPoolConfig %s in namespace %s with Node selector: %v", builder.Definition.Name,
		builder.Definition.Namespace, nodeSelector)

	if len(nodeSelector) == 0 {
		builder.SetError(fmt.Errorf("SriovNetworkPoolConfig 'nodeSelector' cannot be empty map"))

		return builder
	}

	builder.Definition.Spec.NodeSelector = &metav1.LabelSelector{MatchLabels: nodeSelector}

	return builder
}

// WithMaxUnavailable sets MaxUnavailable in the SriovNetworkPoolConfig definition.
func (builder *PoolConfigBuilder) WithMaxUnavailable(maxUnavailable intstrutil.IntOrString) *PoolConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Creating SriovNetworkPoolConfig %s in namespace %s with MaxUnavailable: %v", builder.Definition.Name,
		builder.Definition.Namespace, maxUnavailable)

	if maxUnavailable.Type == intstrutil.String {
		if strings.HasSuffix(maxUnavailable.StrVal, "%") {
			value, err := strconv.Atoi(strings.TrimSuffix(maxUnavailable.StrVal, "%"))
			if err != nil {
				builder.SetError(fmt.Errorf("invalid value %q: %w", maxUnavailable.StrVal, err))

				return builder
			}

			if value > 100 || value < 1 {
				builder.SetError(fmt.Errorf(
					"invalid value: percentage needs to be between 1 and 100: %v", maxUnavailable))

				return builder
			}
		} else {
			builder.SetError(fmt.Errorf("invalid type: strings needs to be a percentage: %v", maxUnavailable))

			return builder
		}
	} else if maxUnavailable.IntValue() < 0 {
		builder.SetError(fmt.Errorf("negative number is not allowed: %v", maxUnavailable))

		return builder
	}

	builder.Definition.Spec.MaxUnavailable = &maxUnavailable

	return builder
}

// WithRDMAMode sets rdmaMode to shared or exclusive.
func (builder *PoolConfigBuilder) WithRDMAMode(mode string) *PoolConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if mode == "" {
		klog.V(100).Info("PoolConfig RdmaMode cannot be empty")

		builder.SetError(fmt.Errorf("rdmaMode cannot be empty"))

		return builder
	}

	if mode != "shared" && mode != "exclusive" {
		klog.V(100).Info("Invalid RdmaMode. Acceptable values: shared or exclusive")

		builder.SetError(fmt.Errorf("invalid value for rdmaMode. It should be 'shared' or 'exclusive'"))

		return builder
	}

	builder.Definition.Spec.RdmaMode = mode

	return builder
}

// PullPoolConfig pulls existing SriovNetworkPoolConfig from cluster.
func PullPoolConfig(apiClient *clients.Settings, name, nsname string) (*PoolConfigBuilder, error) {
	return common.PullNamespacedBuilder[srIovV1.SriovNetworkPoolConfig, PoolConfigBuilder](
		context.TODO(), apiClient, srIovV1.AddToScheme, name, nsname)
}
