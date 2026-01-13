/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeviceConfigSpec defines the desired state of DeviceConfig
type DeviceConfigSpec struct {
	// DriversImage specifies the container image for Neuron drivers
	DriversImage string `json:"driversImage"`

	// DriverVersion specifies the version of the Neuron driver
	// A rolling upgrade is triggered when this field is updated
	// +kubebuilder:validation:Required
	DriverVersion string `json:"driverVersion"`

	// DevicePluginImage specifies the container image for the device plugin
	DevicePluginImage string `json:"devicePluginImage"`

	// CustomSchedulerImage specifies the container image for custom scheduler
	// +optional
	CustomSchedulerImage string `json:"customSchedulerImage,omitempty"`

	// SchedulerExtensionImage specifies the scheduler extension image
	// +optional
	SchedulerExtensionImage string `json:"schedulerExtensionImage,omitempty"`

	// ImageRepoSecret specifies the secret for pulling images from private registries
	// +optional
	ImageRepoSecret *ImageRepoSecret `json:"imageRepoSecret,omitempty"`

	// Selector defines which nodes should run Neuron components
	// +optional
	Selector map[string]string `json:"selector,omitempty"`
}

// ImageRepoSecret defines the secret reference for image repository
type ImageRepoSecret struct {
	// Name is the name of the secret
	Name string `json:"name"`
}

// DeviceConfigStatus defines the observed state of DeviceConfig
type DeviceConfigStatus struct {
	// Add status fields as needed by the operator
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=deviceconfigs,scope=Namespaced

// DeviceConfig is the Schema for AWS Neuron device configuration
type DeviceConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceConfigSpec   `json:"spec,omitempty"`
	Status DeviceConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeviceConfigList contains a list of DeviceConfig
type DeviceConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceConfig{}, &DeviceConfigList{})
}
