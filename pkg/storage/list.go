package storage

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	commonkey "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPV returns a list of builders for persistentVolume.
func ListPV(apiClient *clients.Settings, options ...metav1.ListOptions) ([]*PVBuilder, error) {
	if apiClient == nil {
		klog.V(100).Info("persistentVolume 'apiClient' can not be empty")

		return nil, fmt.Errorf("failed to list persistentVolume, 'apiClient' parameter is empty")
	}

	passedOptions := metav1.ListOptions{}
	logMessage := "Listing all PV resources"

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	klog.V(100).Infof("%v", logMessage)

	pvList, err := apiClient.PersistentVolumes().List(logging.DiscardContext(), passedOptions)
	if err != nil {
		klog.V(100).Infof("Failed to list PV objects due to %s", err.Error())

		return nil, err
	}

	var pvObjects []*PVBuilder

	for _, pv := range pvList.Items {
		copiedPV := pv
		pvBuilder := &PVBuilder{
			apiClient:  apiClient,
			Object:     &copiedPV,
			Definition: &copiedPV,
		}

		pvObjects = append(pvObjects, pvBuilder)
	}

	return pvObjects, nil
}

// ListPVC returns a list of builders for persistentVolumeClaim.
func ListPVC(apiClient *clients.Settings, nsname string, options ...metav1.ListOptions) ([]*PVCBuilder, error) {
	if nsname == "" {
		klog.V(100).Info("PVC namespace is empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			commonkey.NewResourceKey("PersistentVolumeClaim", "", ""), commonerrors.BuilderFieldNamespace)
	}

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	convertedOptions, err := common.ConvertMetaListOptionsToListOptions(options)
	if err != nil {
		return nil, err
	}

	allOptions := append([]runtimeclient.ListOption{runtimeclient.InNamespace(nsname)}, convertedOptions...)

	return common.List[corev1.PersistentVolumeClaim, corev1.PersistentVolumeClaimList, PVCBuilder](
		context.TODO(), apiClient, corev1.AddToScheme, allOptions...)
}
