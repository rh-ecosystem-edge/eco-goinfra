package egressservice

import (
	"context"
	"fmt"
	"strings"

	egresssvcv1 "github.com/ovn-kubernetes/ovn-kubernetes/go-controller/pkg/crd/egressservice/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errInvalidSourceIPBy = "invalid sourceIPBy parameter for the EgressService"
)

var egressServiceGVK = egresssvcv1.SchemeGroupVersion.WithKind("EgressService")

// EgressServiceBuilder provides a struct for EgressService object.
type EgressServiceBuilder struct {
	common.EmbeddableBuilder[egresssvcv1.EgressService, *egresssvcv1.EgressService]
	common.EmbeddableCreator[egresssvcv1.EgressService, EgressServiceBuilder, *egresssvcv1.EgressService, *EgressServiceBuilder]
	common.EmbeddableDeleteReturner[egresssvcv1.EgressService, EgressServiceBuilder, *egresssvcv1.EgressService, *EgressServiceBuilder]
	common.EmbeddableUpdater[egresssvcv1.EgressService, EgressServiceBuilder, *egresssvcv1.EgressService, *EgressServiceBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *EgressServiceBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the EgressService GVK for this builder.
func (builder *EgressServiceBuilder) GetGVK() schema.GroupVersionKind {
	return egressServiceGVK
}

// NewEgressServiceBuilder creates a new instance of EgressService builder.
func NewEgressServiceBuilder(
	apiClient *clients.Settings, name, nsname, sourceIPBy string) *EgressServiceBuilder {
	klog.V(100).Infof(
		"Initializing new EgressService structure with the following params: "+
			"name: %s; namespace: %s; sourceIPBy: %s",
		name, nsname, sourceIPBy)

	builder := common.NewNamespacedBuilder[egresssvcv1.EgressService, EgressServiceBuilder](
		apiClient, egresssvcv1.AddToScheme, name, nsname)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Spec.SourceIPBy = egresssvcv1.SourceIPMode(sourceIPBy)

	if _, err := validateSourceIPMode(sourceIPBy); err != nil {
		klog.V(100).Info("Invalid sourceIPBy parameter for the EgressService")

		builder.SetError(fmt.Errorf("%s", errInvalidSourceIPBy))

		return builder
	}

	return builder
}

// WithNodeLabelSelector applies nodeSelector to the EgressService definition,
// which uses key:value pairs for nodes matching.
func (builder *EgressServiceBuilder) WithNodeLabelSelector(selector map[string]string) *EgressServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Applying nodeSelector %s to EgressService %q in namespace %q",
		selector, builder.Definition.Name, builder.Definition.Namespace)

	if len(selector) == 0 {
		klog.V(100).Info("The nodeselector is empty")

		builder.SetError(fmt.Errorf("cannot accept empty map as nodeSelector"))

		return builder
	}

	builder.Definition.Spec.NodeSelector = metav1.LabelSelector{MatchLabels: selector}

	return builder
}

// WithVRFNetwork sets the network to be used for sending egress and corresponding ingress replies to.
func (builder *EgressServiceBuilder) WithVRFNetwork(vrfnet string) *EgressServiceBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting VRF network %q to EgressService %q in namespace %q",
		vrfnet, builder.Definition.Name, builder.Definition.Namespace)

	if strings.TrimSpace(vrfnet) == "" {
		klog.V(100).Info("Cannot use empty VRF network")

		builder.SetError(fmt.Errorf("cannot use empty VRF network"))

		return builder
	}

	builder.Definition.Spec.Network = vrfnet

	return builder
}

// Pull fetches existing EgressService from the cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*EgressServiceBuilder, error) {
	klog.V(100).Infof("Pulling existing EgressService %q in namespace %q from cluster",
		name, nsname)

	return common.PullNamespacedBuilder[egresssvcv1.EgressService, EgressServiceBuilder](
		context.TODO(), apiClient, egresssvcv1.AddToScheme, name, nsname)
}

func validateSourceIPMode(sourceIPMode string) (string, error) {
	klog.V(100).Infof("Validating requested sourceIPMode %q", sourceIPMode)

	switch {
	case strings.EqualFold(strings.TrimSpace(sourceIPMode), ""):
		klog.V(100).Infof("Invalid sourceIPMode %q(empty)", sourceIPMode)

		return "", fmt.Errorf("invalid empty sourceIPMode %q", sourceIPMode)
	case strings.EqualFold(strings.TrimSpace(sourceIPMode), string(egresssvcv1.SourceIPLoadBalancer)):
		klog.V(100).Infof("Valid sourceIPMode %q", sourceIPMode)

		return string(egresssvcv1.SourceIPLoadBalancer), nil
	case strings.EqualFold(strings.TrimSpace(sourceIPMode), string(egresssvcv1.SourceIPNetwork)):
		klog.V(100).Infof("Valid sourceIPMode %q", sourceIPMode)

		return string(egresssvcv1.SourceIPNetwork), nil
	default:
		klog.V(100).Infof("Invalid sourceIPMode %q", sourceIPMode)

		return "", fmt.Errorf("invalid sourceIPMode %s", sourceIPMode)
	}
}
