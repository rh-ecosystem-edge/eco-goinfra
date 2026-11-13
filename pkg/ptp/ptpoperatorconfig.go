package ptp

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	ptpv1 "github.com/k8snetworkplumbingwg/ptp-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	// PtpOperatorConfigName is the name of the PtpOperatorConfig singleton.
	PtpOperatorConfigName = "default"
	// PtpOperatorConfigNamespace is the namespace for the PtpOperatorConfig singleton.
	PtpOperatorConfigNamespace = "openshift-ptp"
)

// PtpOperatorConfigBuilder provides a struct for the PtpOperatorConfig resource.
type PtpOperatorConfigBuilder struct {
	common.EmbeddableBuilder[ptpv1.PtpOperatorConfig, *ptpv1.PtpOperatorConfig]
	common.EmbeddableUpdater[
		ptpv1.PtpOperatorConfig, PtpOperatorConfigBuilder, *ptpv1.PtpOperatorConfig, *PtpOperatorConfigBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *PtpOperatorConfigBuilder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the PtpOperatorConfig GVK for this builder.
func (builder *PtpOperatorConfigBuilder) GetGVK() schema.GroupVersionKind {
	return ptpv1.GroupVersion.WithKind("PtpOperatorConfig")
}

// PullPtpOperatorConfig pulls the existing PtpOperatorConfig singleton resource.
func PullPtpOperatorConfig(apiClient *clients.Settings) (*PtpOperatorConfigBuilder, error) {
	return common.PullNamespacedBuilder[ptpv1.PtpOperatorConfig, PtpOperatorConfigBuilder](
		context.TODO(), apiClient, ptpv1.AddToScheme, PtpOperatorConfigName, PtpOperatorConfigNamespace)
}

// WithEventConfig sets the PtpEventConfig for the PtpOperatorConfig. It validates that TransportHost is a valid URL and
// ApiVersion is either "1.0" or starts with "2." if provided.
func (builder *PtpOperatorConfigBuilder) WithEventConfig(eventConfig ptpv1.PtpEventConfig) *PtpOperatorConfigBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting PtpEventConfig for PtpOperatorConfig %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if eventConfig.TransportHost != "" {
		_, err := url.Parse(eventConfig.TransportHost)
		if err != nil {
			builder.SetError(fmt.Errorf("invalid TransportHost for PtpEventConfig: %w", err))

			return builder
		}
	}

	if eventConfig.ApiVersion != "" && eventConfig.ApiVersion != "1.0" &&
		!strings.HasPrefix(eventConfig.ApiVersion, "2.") {
		//nolint:staticcheck // Preserve the package's existing error message for API compatibility.
		builder.SetError(fmt.Errorf("invalid ApiVersion for PtpEventConfig: must be 1.0 or start with 2."))

		return builder
	}

	builder.Definition.Spec.EventConfig = &eventConfig

	return builder
}
