package storage

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// ListPV returns a list of builders for persistentVolume.
func ListPV(apiClient *clients.Settings, options ...metav1.ListOptions) ([]*PVBuilder, error) {
	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	convertedOptions, err := common.ConvertMetaListOptionsToListOptions(options)
	if err != nil {
		return nil, err
	}

	return common.List[corev1.PersistentVolume, corev1.PersistentVolumeList, PVBuilder](
		context.TODO(), apiClient, corev1.AddToScheme, convertedOptions...)
}

// ListPVC returns a list of builders for persistentVolumeClaim.
func ListPVC(apiClient *clients.Settings, nsname string, options ...metav1.ListOptions) ([]*PVCBuilder, error) {
	if apiClient == nil {
		klog.V(100).Info("persistentVolumeClaim 'apiClient' can not be empty")

		return nil, fmt.Errorf("failed to list persistentVolumeClaim, 'apiClient' parameter is empty")
	}

	if nsname == "" {
		klog.V(100).Info("PVC namespace is empty")

		return nil, fmt.Errorf("PVC namespace can not be empty")
	}

	passedOptions := metav1.ListOptions{}
	logMessage := "Listing all PVC resources"

	if len(options) > 1 {
		klog.V(100).Info("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	klog.V(100).Infof("%v", logMessage)

	pvcList, err := apiClient.PersistentVolumeClaims(nsname).List(logging.DiscardContext(), passedOptions)
	if err != nil {
		klog.V(100).Infof("Failed to list PVC objects due to %s", err.Error())

		return nil, err
	}

	var pvcObjects []*PVCBuilder

	for _, pvc := range pvcList.Items {
		copiedPVC := pvc
		pvcBuilder := &PVCBuilder{
			apiClient:  apiClient,
			Object:     &copiedPVC,
			Definition: &copiedPVC,
		}

		pvcObjects = append(pvcObjects, pvcBuilder)
	}

	return pvcObjects, nil
}
