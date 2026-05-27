package v2alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// HoldoverParameters defines the holdover timing thresholds for a DPLL subsystem.
// Field names match ptp.openshift.io/v2alpha1 HardwareConfig CRD exactly.
type HoldoverParameters struct {
	// MaxInSpecOffset is the holdover specification threshold in nanoseconds.
	MaxInSpecOffset uint64 `json:"maxInSpecOffset,omitempty"`
	// LocalMaxHoldoverOffset is the maximum holdover offset in nanoseconds.
	LocalMaxHoldoverOffset uint64 `json:"localMaxHoldoverOffset,omitempty"`
	// LocalHoldoverTimeout is the time the clock stays in holdover before
	// reaching LocalMaxHoldoverOffset, in seconds.
	LocalHoldoverTimeout uint64 `json:"localHoldoverTimeout,omitempty"`
}

// DPLL holds the minimal DPLL configuration required for holdover test management.
type DPLL struct {
	// HoldoverParameters defines the combination of the DPLL complex hardware
	// parameters and the holdover specification threshold.
	HoldoverParameters *HoldoverParameters `json:"holdoverParameters,omitempty"`
}

// Subsystem represents one synchronization subsystem in the clock chain.
type Subsystem struct {
	// Name is a human-readable identifier for this subsystem.
	Name string `json:"name"`
	// DPLL contains the DPLL configuration for this subsystem.
	DPLL DPLL `json:"dpll,omitempty"`
}

// ClockChain is the root clock chain configuration.
type ClockChain struct {
	// Structure defines the system structure as a list of atomic synchronization
	// subsystems. Must contain at least one subsystem.
	Structure []Subsystem `json:"structure"`
}

// HardwareProfile defines a hardware configuration profile.
type HardwareProfile struct {
	// ClockChain contains the complete clock chain configuration for this profile.
	ClockChain *ClockChain `json:"clockChain"`
}

// HardwareConfigSpec defines the desired state of HardwareConfig.
type HardwareConfigSpec struct {
	// Profile contains the hardware profile with its configuration.
	Profile HardwareProfile `json:"profile"`
	// RelatedPtpProfileName specifies the name of the related PTP profile.
	RelatedPtpProfileName string `json:"relatedPtpProfileName,omitempty"`
}

// HardwareConfig is the Schema for the hardwareconfigs API.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type HardwareConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HardwareConfigSpec `json:"spec,omitempty"`
}

// HardwareConfigList contains a list of HardwareConfig.
//
// +kubebuilder:object:root=true
type HardwareConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HardwareConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HardwareConfig{}, &HardwareConfigList{})
}
