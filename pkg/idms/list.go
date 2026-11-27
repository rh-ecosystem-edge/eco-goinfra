package idms

import (
	"context"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/klog/v2"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListImageDigestMirrorSets returns a cluster-wide imagedigestmirrorset inventory.
func ListImageDigestMirrorSets(
	apiClient *clients.Settings,
	options ...runtimeClient.ListOptions) ([]*Builder, error) {
	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	return common.List[configv1.ImageDigestMirrorSet, configv1.ImageDigestMirrorSetList, Builder](
		context.TODO(), apiClient, configv1.AddToScheme, common.ConvertListOptionsToOptions(options)...)
}
