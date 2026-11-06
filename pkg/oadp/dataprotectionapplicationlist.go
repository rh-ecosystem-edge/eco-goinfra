package oadp

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	oadpv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/oadp/api/v1alpha1"
	"k8s.io/klog/v2"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListDataProtectionApplication returns dataprotectionapplication inventory in the given namespace.
func ListDataProtectionApplication(
	apiClient *clients.Settings, nsname string, options ...runtimeClient.ListOptions) ([]*DPABuilder, error) {
	klog.V(100).Infof("Listing dataprotectionapplications in namespace %s", nsname)

	if apiClient == nil {
		klog.V(100).Info("The apiClient cannot be nil")

		return nil, commonerrors.NewAPIClientNil(
			commonkey.NewResourceKey(dataProtectionApplicationGVK.Kind, "", ""))
	}

	if err := apiClient.AttachScheme(oadpv1alpha1.AddToScheme); err != nil {
		klog.V(100).Infof(
			"Failed to add oadpv1alpha1 scheme to client schemes")

		return nil, commonerrors.NewSchemeAttacherFailed(
			commonkey.NewResourceKey(dataProtectionApplicationGVK.Kind, "", ""), err)
	}

	if nsname == "" {
		klog.V(100).Info("dataprotectionapplication 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey(dataProtectionApplicationGVK.Kind, "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[
		oadpv1alpha1.DataProtectionApplication, oadpv1alpha1.DataProtectionApplicationList, DPABuilder](
		context.TODO(), apiClient, oadpv1alpha1.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
