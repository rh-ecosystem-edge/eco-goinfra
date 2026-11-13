package lca

import (
	"context"
	"fmt"
	"slices"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	lcav1 "github.com/openshift-kni/lifecycle-agent/api/imagebasedupgrade/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
)

const (
	isTrue     = "True"
	isFalse    = "False"
	isComplete = "Completed"
	ibuName    = "upgrade"
	idle       = "Idle"
	disabled   = "Disabled"
)

// ImageBasedUpgradeBuilder provides struct for the imagebasedupgrade object containing connection to
// the cluster and the imagebasedupgrade definitions.
type ImageBasedUpgradeBuilder struct {
	common.EmbeddableBuilder[lcav1.ImageBasedUpgrade, *lcav1.ImageBasedUpgrade]
	common.EmbeddableDeleteReturner[
		lcav1.ImageBasedUpgrade, ImageBasedUpgradeBuilder, *lcav1.ImageBasedUpgrade, *ImageBasedUpgradeBuilder]
	common.EmbeddableWithOptions[
		lcav1.ImageBasedUpgrade, ImageBasedUpgradeBuilder, *lcav1.ImageBasedUpgrade, *ImageBasedUpgradeBuilder, AdditionalOptions]
}

// AdditionalOptions additional options for imagebasedupgrade object.
type AdditionalOptions func(builder *ImageBasedUpgradeBuilder) (*ImageBasedUpgradeBuilder, error)

// AttachMixins wires the embedded mixins to this builder instance.
func (builder *ImageBasedUpgradeBuilder) AttachMixins() {
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the ImageBasedUpgrade GVK for this builder.
func (builder *ImageBasedUpgradeBuilder) GetGVK() schema.GroupVersionKind {
	return lcav1.GroupVersion.WithKind("ImageBasedUpgrade")
}

// PullImageBasedUpgrade pulls existing imagebasedupgrade from cluster.
func PullImageBasedUpgrade(apiClient *clients.Settings) (*ImageBasedUpgradeBuilder, error) {
	return common.PullClusterScopedBuilder[lcav1.ImageBasedUpgrade, ImageBasedUpgradeBuilder](
		context.TODO(), apiClient, lcav1.AddToScheme, ibuName)
}

// Update modifies the imagebasedupgrade resource on the cluster
// to match what is defined in the local definition of the builder.
func (builder *ImageBasedUpgradeBuilder) Update() (*ImageBasedUpgradeBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	if err := common.Update(context.TODO(), builder, false); err != nil {
		return builder, err
	}

	// Wait for the IBU to reconcile after it is updated.
	err := wait.PollUntilContextTimeout(
		context.TODO(), time.Second*2, time.Second*10, true, func(ctx context.Context) (bool, error) {
			klog.V(100).Infof("Waiting for imagebasedupgrade %s to finish reconciling",
				builder.Definition.Name)

			ibu, err := builder.Get()
			if err != nil {
				return false, err
			}

			if ibu.Generation == ibu.Status.ObservedGeneration {
				builder.Object = ibu
				builder.Definition = ibu

				return true, nil
			}

			return false, nil
		})
	if err != nil {
		return builder, err
	}

	return builder, nil
}

// WithSeedImage sets the seed image used by the imagebasedupgrade.
func (builder *ImageBasedUpgradeBuilder) WithSeedImage(
	seedImage string) *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting image %s in imagebasedupgrade", seedImage)

	builder.Definition.Spec.SeedImageRef.Image = seedImage

	return builder
}

// WithExtraManifests adds extraManifests to be used by the imagebasedupgrade.
// This is used to create/configure resources during upgrade.
func (builder *ImageBasedUpgradeBuilder) WithExtraManifests(
	extraManifestsConfigMapName, extraManifestsConfigMapNamespace string) *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Appending extraManifests's configmap name %s in namespace %s to the imagebasedupgrade",
		extraManifestsConfigMapName, extraManifestsConfigMapNamespace)

	builder.Definition.Spec.ExtraManifests = append(builder.Definition.Spec.ExtraManifests,
		lcav1.ConfigMapRef{Name: extraManifestsConfigMapName, Namespace: extraManifestsConfigMapNamespace})

	return builder
}

// WithOadpContent adds oadpContent to be used by the imagebasedupgrade.
// This is used for backup/restore during upgrade.
func (builder *ImageBasedUpgradeBuilder) WithOadpContent(
	oadpContentConfigMapName, oadpContentConfigMapNamespace string) *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Appending oadpContent's configmap name %s in namespace %s to the imagebasedupgrade",
		oadpContentConfigMapName, oadpContentConfigMapNamespace)

	builder.Definition.Spec.OADPContent = append(builder.Definition.Spec.OADPContent,
		lcav1.ConfigMapRef{Name: oadpContentConfigMapName, Namespace: oadpContentConfigMapNamespace})

	return builder
}

// AutoRollbackOnFailureInitMonitorTimeoutSeconds allows controlling
// the timeout for the upgrade to complete before the rollback.
// Set to 1800 seconds by default.
func (builder *ImageBasedUpgradeBuilder) AutoRollbackOnFailureInitMonitorTimeoutSeconds(
	seconds uint) *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting timeout for InitMonitor to %d seconds in imagebasedupgrade", seconds)

	builder.Definition.Spec.AutoRollbackOnFailure.InitMonitorTimeoutSeconds = int(seconds)

	return builder
}

// AutoRollbackOnFailureDisabledInitMonitor allows disabling the watchdog
// triggering a rollback upon upgrade failure within the set timeout.
func (builder *ImageBasedUpgradeBuilder) AutoRollbackOnFailureDisabledInitMonitor() *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Info("Setting the Init Monitor for Auto Rollback on failure to Disabled in imagebasedupgrade")

	if builder.Definition.Annotations == nil {
		builder.Definition.Annotations = make(map[string]string)
	}

	builder.Definition.Annotations["auto-rollback-on-failure.lca.openshift.io/init-monitor"] = disabled

	return builder
}

// AutoRollbackOnFailureDisableForPostReboot allows controlling
// AutoRollback on failure for post reboot stage.
func (builder *ImageBasedUpgradeBuilder) AutoRollbackOnFailureDisableForPostReboot() *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Info("Setting Auto Rollback on failure for post reboot to Disabled in imagebasedupgrade")

	if builder.Definition.Annotations == nil {
		builder.Definition.Annotations = make(map[string]string)
	}

	builder.Definition.Annotations["auto-rollback-on-failure.lca.openshift.io/post-reboot-config"] = disabled

	return builder
}

// AutoRollbackOnFailureDisableForUpgradeCompletion allows controlling
// AutoRollback on failure for upgrade completion stage. Enabled by default.
func (builder *ImageBasedUpgradeBuilder) AutoRollbackOnFailureDisableForUpgradeCompletion() *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Info("Setting Auto Rollback on failure for upgrade completion to Disabled in imagebasedupgrade")

	if builder.Definition.Annotations == nil {
		builder.Definition.Annotations = make(map[string]string)
	}

	builder.Definition.Annotations["auto-rollback-on-failure.lca.openshift.io/upgrade-completion"] = disabled

	return builder
}

// WithSeedImageVersion sets the seed image version used by the imagebasedupgrade.
func (builder *ImageBasedUpgradeBuilder) WithSeedImageVersion(
	seedImageVersion string) *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting seed image version %s in imagebasedupgrade", seedImageVersion)

	builder.Definition.Spec.SeedImageRef.Version = seedImageVersion

	return builder
}

// WithSeedImagePullSecretRef sets the imagebasedupgrade with reference to the pull-secret
// for pulling the seed image.
func (builder *ImageBasedUpgradeBuilder) WithSeedImagePullSecretRef(
	pullSecretName string) *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting pull-secret %s in imagebasedupgrade for pulling the seed image", pullSecretName)

	builder.Definition.Spec.SeedImageRef.PullSecretRef = &lcav1.PullSecretRef{Name: pullSecretName}

	return builder
}

// WaitUntilStageComplete waits the specified timeout for the imagebasedupgrade to complete
// actions for the provided stage.
//
//nolint:gocognit
func (builder *ImageBasedUpgradeBuilder) WaitUntilStageComplete(stage string) (*ImageBasedUpgradeBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Waiting for imagebasedupgrade %s to set stage %s",
		builder.Definition.Name,
		stage)

	if !builder.Exists() {
		klog.V(100).Info("The imagebasedupgrade does not exist on the cluster")

		return builder, fmt.Errorf("imagebasedupgrade object %s does not exist", builder.Definition.Name)
	}

	if !slices.Contains([]string{idle, "Prep", "Upgrade", "Rollback"}, stage) {
		return builder, fmt.Errorf("wrong stage selected for imagebasedupgrade")
	}

	// Polls periodically to determine if imagebasedupgrade is in desired state.
	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second*3, time.Minute*30, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			builder.Definition = builder.Object

			for _, condition := range builder.Object.Status.Conditions {
				switch stage {
				case idle:
					if condition.Status == isTrue && condition.Type == idle {
						return true, nil
					}

				case "Prep":
					if condition.Status == isFalse && condition.Type == "PrepInProgress" &&
						condition.Message == "Prep completed" && condition.Reason == isComplete {
						return true, nil
					}
				case "Upgrade":
					if condition.Status == isFalse && condition.Type == "UpgradeInProgress" &&
						condition.Message == "Upgrade completed" && condition.Reason == isComplete {
						return true, nil
					}

				case "Rollback":
					if condition.Status == isFalse && condition.Type == "RollbackInProgress" &&
						condition.Message == "Rollback completed" && condition.Reason == isComplete {
						return true, nil
					}

				default:
					return false, fmt.Errorf("wrong stage selected for imagebasedupgrade")
				}
			}

			return false, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}

// WithStage sets the stage used by the imagebasedupgrade.
func (builder *ImageBasedUpgradeBuilder) WithStage(
	stage string) *ImageBasedUpgradeBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Setting stage %s in imagebasedupgrade", stage)
	builder.Definition.Spec.Stage = lcav1.ImageBasedUpgradeStage(stage)

	return builder
}
