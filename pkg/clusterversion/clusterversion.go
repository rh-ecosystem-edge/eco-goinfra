package clusterversion

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Masterminds/semver/v3"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const clusterVersionName = "version"

// Builder provides a struct for clusterversion object from the cluster and a clusterversion definition.
type Builder struct {
	common.EmbeddableBuilder[configv1.ClusterVersion, *configv1.ClusterVersion]
	common.EmbeddableUpdater[configv1.ClusterVersion, Builder, *configv1.ClusterVersion, *Builder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the ClusterVersion GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return configv1.GroupVersion.WithKind("ClusterVersion")
}

// Pull loads an existing clusterversion into Builder struct.
func Pull(apiClient *clients.Settings) (*Builder, error) {
	return common.PullClusterScopedBuilder[configv1.ClusterVersion, Builder](
		context.TODO(), apiClient, configv1.Install, clusterVersionName)
}

// WithDesiredUpdateImage adds the desired image to the clusterversion struct.
func (builder *Builder) WithDesiredUpdateImage(desiredUpdateImage string, force bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding the desired image %s to ClusterVersion %s",
		desiredUpdateImage, builder.Definition.Name)

	if desiredUpdateImage == "" {
		klog.V(100).Info("The desiredUpdateImage is empty")

		builder.SetError(fmt.Errorf("clusterversion 'desiredUpdateImage' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.DesiredUpdate = &configv1.Update{Image: desiredUpdateImage, Force: force}

	return builder
}

// WithDesiredUpdateChannel adds the desired channel to the clusterversion struct.
func (builder *Builder) WithDesiredUpdateChannel(updateChannel string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding the desired updateChannel %s to ClusterVersion %s",
		updateChannel, builder.Definition.Name)

	if updateChannel == "" {
		klog.V(100).Info("The updateChannel is empty")

		builder.SetError(fmt.Errorf("clusterversion 'updateChannel' cannot be empty"))

		return builder
	}

	builder.Definition.Spec.Channel = updateChannel

	return builder
}

// WaitUntilProgressing waits for timeout duration or until clusterversion is in Progressing state.
func (builder *Builder) WaitUntilProgressing(timeout time.Duration) error {
	return builder.WaitUntilConditionTrue(configv1.OperatorProgressing, timeout)
}

// WaitUntilAvailable waits for timeout duration or until clusterversion is in Available state.
func (builder *Builder) WaitUntilAvailable(timeout time.Duration) error {
	return builder.WaitUntilConditionTrue(configv1.OperatorAvailable, timeout)
}

// WaitUntilConditionTrue waits for timeout duration or until clusterversion gets to a specific status.
func (builder *Builder) WaitUntilConditionTrue(
	conditionType configv1.ClusterStatusConditionType, timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	if !builder.Exists() {
		return fmt.Errorf("clusterversion object %s does not exist", builder.Definition.Name)
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("Failed to get the ClusterVersion with error %s", err)

				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Type == conditionType {
					return condition.Status == configv1.ConditionTrue, nil
				}
			}

			return false, nil
		})
}

// WaitUntilUpdateIsStarted waits until there is a history entry indicating the update start.
func (builder *Builder) WaitUntilUpdateIsStarted(timeout time.Duration) error {
	return builder.WaitUntilUpdateHistoryStateTrue(configv1.PartialUpdate, timeout)
}

// WaitUntilUpdateIsCompleted waits until there is a history entry indicating the update completed.
func (builder *Builder) WaitUntilUpdateIsCompleted(timeout time.Duration) error {
	return builder.WaitUntilUpdateHistoryStateTrue(configv1.CompletedUpdate, timeout)
}

// WaitUntilUpdateHistoryStateTrue waits until there is a history entry indicating an updateHistoryState.
func (builder *Builder) WaitUntilUpdateHistoryStateTrue(
	updateHistoryState configv1.UpdateState, timeout time.Duration) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	if !builder.Exists() {
		return fmt.Errorf("clusterversion object %s does not exist", builder.Definition.Name)
	}

	return wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("Failed to get the ClusterVersion with error %s", err)

				return false, nil
			}

			updateImage := builder.Object.Status.Desired.Image

			for _, updateHistory := range builder.Object.Status.History {
				if updateHistory.Image == updateImage && updateHistory.State == updateHistoryState {
					return true, nil
				}
			}

			return false, nil
		})
}

// GetNextUpdateVersionImage fetches the next recommended or conditional update for the cluster.
func (builder *Builder) GetNextUpdateVersionImage(stream string, acceptConditionalVersions bool) (string, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Infof("Getting the update version image in stream %s for clusterversion %s",
		stream, builder.Definition.Name)

	if stream == "" {
		return "", fmt.Errorf("stream can not be empty")
	}

	if !builder.Exists() {
		return "", fmt.Errorf("clusterversion object %s does not exist", builder.Definition.Name)
	}

	currentVersion := builder.Object.Status.Desired.Version

	for _, availableUpdate := range builder.Object.Status.AvailableUpdates {
		isStreamUpdate, err := builder.isStreamUpdate(currentVersion, availableUpdate.Version, stream)
		if isStreamUpdate && err == nil {
			return availableUpdate.Image, nil
		}
	}

	if acceptConditionalVersions {
		for _, conditionalUpdate := range builder.Object.Status.ConditionalUpdates {
			isStreamUpdate, err := builder.isStreamUpdate(currentVersion, conditionalUpdate.Release.Version, stream)
			if isStreamUpdate && err == nil {
				return conditionalUpdate.Release.Image, nil
			}
		}
	}

	return "", fmt.Errorf("update version in %s stream not found", stream)
}

// isStreamUpdate checks if updateVersion is a 'stream' (X, Y or Z) update for version.
func (builder *Builder) isStreamUpdate(version, updateVersion, stream string) (isStreamUpdate bool, err error) {
	klog.V(100).Infof("Verify if updateVersion %s is a stream %s update for version %s", updateVersion, stream, version)

	if !slices.Contains([]string{X, Z, Y}, stream) {
		klog.V(100).Infof("invalid stream %s", stream)

		return false, fmt.Errorf("invalid stream %s", stream)
	}

	semVersion, semVersionError := semver.NewVersion(version)
	if semVersionError != nil {
		return false, fmt.Errorf("the version %s is invalid", version)
	}

	semUpdateVersion, updateVersionError := semver.NewVersion(updateVersion)
	klog.V(100).Infof("Testing %s and %s", semVersion.String(), semUpdateVersion.String())

	if updateVersionError != nil {
		return false, fmt.Errorf("the Update Version %s is invalid", updateVersion)
	}

	switch major := semVersion.Major(); {
	case major == semUpdateVersion.Major() && semVersion.Minor() == semUpdateVersion.Minor() &&
		semVersion.Patch() < semUpdateVersion.Patch() && stream == Z:
		klog.V(100).Infof("This version is a z update: %s", semUpdateVersion.String())

		return true, nil

	case major == semUpdateVersion.Major() &&
		semVersion.Minor() < semUpdateVersion.Minor() && stream == Y:
		klog.V(100).Infof("This version is a y update: %s", semUpdateVersion.String())

		return true, nil

	case semVersion.Major() < semUpdateVersion.Major() && stream == X:
		klog.V(100).Infof("This version is an x update: %s", semUpdateVersion.String())

		return true, nil

	default:
		klog.V(100).Infof("The version %s is not an update for %s ",
			semUpdateVersion.String(), semVersion.String())

		return false, nil
	}
}
