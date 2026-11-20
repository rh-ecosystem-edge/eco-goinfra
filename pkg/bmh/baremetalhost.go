package bmh

import (
	"context"
	"fmt"
	"slices"
	"time"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	bootModeUEFISecureBoot = "UEFISecureBoot"
	errInvalidBootMode     = "not acceptable 'bootMode' value"
)

var bareMetalHostGVK = bmhv1alpha1.GroupVersion.WithKind("BareMetalHost")

// BmhBuilder provides struct for the bmh object containing connection to
// the cluster and the bmh definitions.
type BmhBuilder struct {
	common.EmbeddableBuilder[bmhv1alpha1.BareMetalHost, *bmhv1alpha1.BareMetalHost]
	common.EmbeddableCreator[bmhv1alpha1.BareMetalHost, BmhBuilder, *bmhv1alpha1.BareMetalHost, *BmhBuilder]
	common.EmbeddableDeleteReturner[bmhv1alpha1.BareMetalHost, BmhBuilder, *bmhv1alpha1.BareMetalHost, *BmhBuilder]
	common.EmbeddableWithOptions[
		bmhv1alpha1.BareMetalHost, BmhBuilder, *bmhv1alpha1.BareMetalHost, *BmhBuilder, AdditionalOptions]
}

// AdditionalOptions additional options for bmh object.
type AdditionalOptions func(builder *BmhBuilder) (*BmhBuilder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *BmhBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the BareMetalHost GVK for this builder.
func (builder *BmhBuilder) GetGVK() schema.GroupVersionKind {
	return bareMetalHostGVK
}

// NewBuilder creates a new instance of BmhBuilder.
func NewBuilder(
	apiClient *clients.Settings, name, nsname, bmcAddress, bmcSecretName, bootMacAddress, bootMode string) *BmhBuilder {
	builder := common.NewNamespacedBuilder[bmhv1alpha1.BareMetalHost, BmhBuilder](
		apiClient, bmhv1alpha1.AddToScheme, name, nsname)

	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Spec = bmhv1alpha1.BareMetalHostSpec{
		BMC: bmhv1alpha1.BMCDetails{
			Address:                        bmcAddress,
			CredentialsName:                bmcSecretName,
			DisableCertificateVerification: true,
		},
		BootMode:              bmhv1alpha1.BootMode(bootMode),
		BootMACAddress:        bootMacAddress,
		Online:                true,
		ExternallyProvisioned: false,
	}

	if bmcAddress == "" {
		klog.V(100).Info("The bootmacaddress of the baremetalhost is empty")

		builder.SetError(fmt.Errorf("BMH 'bmcAddress' cannot be empty"))

		return builder
	}

	if bmcSecretName == "" {
		klog.V(100).Info("The bmcsecret of the baremetalhost is empty")

		builder.SetError(fmt.Errorf("BMH 'bmcSecretName' cannot be empty"))

		return builder
	}

	bootModeAcceptable := []string{"UEFI", bootModeUEFISecureBoot, "legacy"}
	if !slices.Contains(bootModeAcceptable, bootMode) {
		builder.SetError(fmt.Errorf("%s", errInvalidBootMode))

		return builder
	}

	if bootMacAddress == "" {
		klog.V(100).Info("The bootmacaddress of the baremetalhost is empty")

		builder.SetError(fmt.Errorf("BMH 'bootMacAddress' cannot be empty"))

		return builder
	}

	return builder
}

// WithRootDeviceDeviceName sets rootDeviceHints DeviceName to specified value.
func (builder *BmhBuilder) WithRootDeviceDeviceName(deviceName string) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if deviceName == "" {
		klog.V(100).Info("The baremetalhost rootDeviceHint deviceName is empty")

		builder.SetError(fmt.Errorf("the baremetalhost rootDeviceHint deviceName cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.DeviceName = deviceName

	return builder
}

// WithRootDeviceHTCL sets rootDeviceHints HTCL to specified value.
func (builder *BmhBuilder) WithRootDeviceHTCL(hctl string) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if hctl == "" {
		klog.V(100).Info("The baremetalhost rootDeviceHint hctl is empty")

		builder.SetError(fmt.Errorf("the baremetalhost rootDeviceHint hctl cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.HCTL = hctl

	return builder
}

// WithRootDeviceModel sets rootDeviceHints Model to specified value.
func (builder *BmhBuilder) WithRootDeviceModel(model string) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if model == "" {
		klog.V(100).Info("The baremetalhost rootDeviceHint model is empty")

		builder.SetError(fmt.Errorf("the baremetalhost rootDeviceHint model cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.Model = model

	return builder
}

// WithRootDeviceVendor sets rootDeviceHints Vendor to specified value.
func (builder *BmhBuilder) WithRootDeviceVendor(vendor string) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if vendor == "" {
		klog.V(100).Info("The baremetalhost rootDeviceHint vendor is empty")

		builder.SetError(fmt.Errorf("the baremetalhost rootDeviceHint vendor cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.Model = vendor

	return builder
}

// WithRootDeviceSerialNumber sets rootDeviceHints serialNumber to specified value.
func (builder *BmhBuilder) WithRootDeviceSerialNumber(serialNumber string) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if serialNumber == "" {
		klog.V(100).Info("The baremetalhost rootDeviceHint serialNumber is empty")

		builder.SetError(fmt.Errorf("the baremetalhost rootDeviceHint serialNumber cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.SerialNumber = serialNumber

	return builder
}

// WithRootDeviceMinSizeGigabytes sets rootDeviceHints MinSizeGigabytes to specified value.
func (builder *BmhBuilder) WithRootDeviceMinSizeGigabytes(size int) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if size < 0 {
		klog.V(100).Info("The baremetalhost rootDeviceHint size is less than 0")

		builder.SetError(fmt.Errorf("the baremetalhost rootDeviceHint size cannot be less than 0"))

		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.MinSizeGigabytes = size

	return builder
}

// WithRootDeviceWWN sets rootDeviceHints WWN to specified value.
func (builder *BmhBuilder) WithRootDeviceWWN(wwn string) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if wwn == "" {
		klog.V(100).Info("The baremetalhost rootDeviceHint wwn is empty")

		builder.SetError(fmt.Errorf("the baremetalhost rootDeviceHint wwn cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.WWN = wwn

	return builder
}

// WithRootDeviceWWNWithExtension sets rootDeviceHints WWNWithExtension to specified value.
func (builder *BmhBuilder) WithRootDeviceWWNWithExtension(wwnWithExtension string) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if wwnWithExtension == "" {
		klog.V(100).Info("The baremetalhost rootDeviceHint wwnWithExtension is empty")

		builder.SetError(fmt.Errorf("the baremetalhost rootDeviceHint wwnWithExtension cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.WWNWithExtension = wwnWithExtension

	return builder
}

// WithRootDeviceWWNVendorExtension sets rootDeviceHint WWNVendorExtension to specified value.
func (builder *BmhBuilder) WithRootDeviceWWNVendorExtension(wwnVendorExtension string) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if wwnVendorExtension == "" {
		klog.V(100).Info("The baremetalhost rootDeviceHint wwnVendorExtension is empty")

		builder.SetError(fmt.Errorf("the baremetalhost rootDeviceHint wwnVendorExtension cannot be empty"))

		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.WWNVendorExtension = wwnVendorExtension

	return builder
}

// WithRootDeviceRotationalDisk sets rootDeviceHint Rotational to specified value.
func (builder *BmhBuilder) WithRootDeviceRotationalDisk(rotational bool) *BmhBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if builder.Definition.Spec.RootDeviceHints == nil {
		builder.Definition.Spec.RootDeviceHints = &bmhv1alpha1.RootDeviceHints{}
	}

	builder.Definition.Spec.RootDeviceHints.Rotational = &rotational

	return builder
}

// Pull pulls existing baremetalhost from cluster.
func Pull(apiClient *clients.Settings, name, nsname string) (*BmhBuilder, error) {
	return common.PullNamespacedBuilder[bmhv1alpha1.BareMetalHost, BmhBuilder](
		context.TODO(), apiClient, bmhv1alpha1.AddToScheme, name, nsname)
}

// GetBmhOperationalState returns the current OperationalStatus of the bmh.
func (builder *BmhBuilder) GetBmhOperationalState() bmhv1alpha1.OperationalStatus {
	if err := common.Validate(builder); err != nil {
		return ""
	}

	klog.V(100).Infof("Pull OperationalStatus value for %s baremetalhost within %s namespace",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return ""
	}

	return builder.Object.Status.OperationalStatus
}

// GetBmhPowerOnStatus checks BareMetalHost PowerOn status.
func (builder *BmhBuilder) GetBmhPowerOnStatus() bool {
	if err := common.Validate(builder); err != nil {
		return false
	}

	klog.V(100).Infof("Pull PoweredOn value for %s baremetalhost within %s namespace",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return false
	}

	return builder.Object.Status.PoweredOn
}

// CreateAndWaitUntilProvisioned creates bmh object and waits until bmh is provisioned.
func (builder *BmhBuilder) CreateAndWaitUntilProvisioned(timeout time.Duration) (*BmhBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof(`Creating the baremetalhost %s in namespace %s and
	waiting for the defined period until it is created`,
		builder.Definition.Name, builder.Definition.Namespace)

	builder, err := builder.Create()
	if err != nil {
		return nil, err
	}

	err = builder.WaitUntilProvisioned(timeout)

	return builder, err
}

// WaitUntilProvisioned waits for timeout duration or until bmh is provisioned.
func (builder *BmhBuilder) WaitUntilProvisioned(timeout time.Duration) error {
	return builder.WaitUntilInStatus(bmhv1alpha1.StateProvisioned, timeout)
}

// WaitUntilProvisioning waits for timeout duration or until bmh is provisioning.
func (builder *BmhBuilder) WaitUntilProvisioning(timeout time.Duration) error {
	return builder.WaitUntilInStatus(bmhv1alpha1.StateProvisioning, timeout)
}

// WaitUntilReady waits for timeout duration or until bmh is ready.
func (builder *BmhBuilder) WaitUntilReady(timeout time.Duration) error {
	return builder.WaitUntilInStatus(bmhv1alpha1.StateReady, timeout)
}

// WaitUntilAvailable waits for timeout duration or until bmh is available.
func (builder *BmhBuilder) WaitUntilAvailable(timeout time.Duration) error {
	return builder.WaitUntilInStatus(bmhv1alpha1.StateAvailable, timeout)
}

// WaitUntilInStatus waits for timeout duration or until bmh gets to a specific status.
func (builder *BmhBuilder) WaitUntilInStatus(status bmhv1alpha1.ProvisioningState, timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			if builder.Object.Status.Provisioning.State == status {
				return true, nil
			}

			return false, err
		})
}

// DeleteAndWaitUntilDeleted delete bmh object and waits until deleted.
func (builder *BmhBuilder) DeleteAndWaitUntilDeleted(timeout time.Duration) (*BmhBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof(`Deleting baremetalhost %s in namespace %s and
	waiting for the defined period until it is removed`,
		builder.Definition.Name, builder.Definition.Namespace)

	builder, err := builder.Delete()
	if err != nil {
		return builder, err
	}

	err = builder.WaitUntilDeleted(timeout)

	return nil, err
}

// WaitUntilDeleted waits for timeout duration or until bmh is deleted.
func (builder *BmhBuilder) WaitUntilDeleted(timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, false, func(ctx context.Context) (bool, error) {
			_, err := builder.Get()
			if err == nil {
				klog.V(100).Infof("bmh %s/%s still present",
					builder.Definition.Namespace,
					builder.Definition.Name)

				return false, nil
			}

			if k8serrors.IsNotFound(err) {
				klog.V(100).Infof("bmh %s/%s is gone",
					builder.Definition.Namespace,
					builder.Definition.Name)

				return true, nil
			}

			klog.V(100).Infof("failed to get bmh %s/%s: %v",
				builder.Definition.Namespace,
				builder.Definition.Name, err)

			return false, err
		})

	return err
}

// WaitUntilAnnotationExists waits up to the specified timeout until the annotation exists.
func (builder *BmhBuilder) WaitUntilAnnotationExists(annotation string, timeout time.Duration) (*BmhBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	if annotation == "" {
		klog.V(100).Info("BMH annotation key cannot be empty")

		return nil, fmt.Errorf("bmh annotation key cannot be empty")
	}

	klog.V(100).Infof(
		"Waiting until BMH %s in namespace %s has annotation %s",
		builder.Definition.Name, builder.Definition.Namespace, annotation)

	if !builder.Exists() {
		return nil, fmt.Errorf(
			"baremetalhost object %s does not exist in namespace %s", builder.Definition.Name, builder.Definition.Namespace)
	}

	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("failed to get bmh %s/%s: %v", builder.Definition.Namespace, builder.Definition.Name, err)

				return false, nil
			}

			if _, ok := builder.Object.Annotations[annotation]; !ok {
				return false, nil
			}

			return true, nil
		})
	if err != nil {
		return nil, err
	}

	return builder, nil
}
