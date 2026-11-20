package lca

import (
	"context"
	"fmt"
	"time"

	lcasgv1 "github.com/openshift-kni/lifecycle-agent/api/seedgenerator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	seedImageName                 = "seedimage"
	conditionTypeSeedGenCompleted = "SeedGenCompleted"
)

// SeedGeneratorBuilder provides struct for the seedgenerator object containing connection to
// the cluster and the seedgenerator definitions.
type SeedGeneratorBuilder struct {
	common.EmbeddableBuilder[lcasgv1.SeedGenerator, *lcasgv1.SeedGenerator]
	common.EmbeddableCreator[
		lcasgv1.SeedGenerator, SeedGeneratorBuilder, *lcasgv1.SeedGenerator, *SeedGeneratorBuilder]
	common.EmbeddableDeleteReturner[
		lcasgv1.SeedGenerator, SeedGeneratorBuilder, *lcasgv1.SeedGenerator, *SeedGeneratorBuilder]
	common.EmbeddableWithOptions[
		lcasgv1.SeedGenerator, SeedGeneratorBuilder, *lcasgv1.SeedGenerator, *SeedGeneratorBuilder,
		SeedGeneratorAdditionalOptions]
}

// SeedGeneratorAdditionalOptions additional options for seedgenerator object.
type SeedGeneratorAdditionalOptions func(builder *SeedGeneratorBuilder) (*SeedGeneratorBuilder, error)

// AttachMixins wires the embedded mixins to this builder instance.
func (builder *SeedGeneratorBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleteReturner.SetBase(builder)
	builder.EmbeddableWithOptions.SetBase(builder)
}

// GetGVK returns the SeedGenerator GVK for this builder.
func (builder *SeedGeneratorBuilder) GetGVK() schema.GroupVersionKind {
	return lcasgv1.GroupVersion.WithKind("SeedGenerator")
}

// NewSeedGeneratorBuilder creates a new instance of SeedGenerator.
func NewSeedGeneratorBuilder(
	apiClient *clients.Settings,
	name string,
) *SeedGeneratorBuilder {
	builder := common.NewClusterScopedBuilder[lcasgv1.SeedGenerator, SeedGeneratorBuilder](
		apiClient, lcasgv1.AddToScheme, seedImageName)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.Name = name

	if name != seedImageName {
		klog.V(100).Infof("The name of the seedgenerator must be " + seedImageName)

		builder.SetError(fmt.Errorf("SeedGenerator name must be %s", seedImageName))

		return builder
	}

	return builder
}

// PullSeedGenerator pulls existing seedgenerator from cluster.
func PullSeedGenerator(apiClient *clients.Settings, name string) (*SeedGeneratorBuilder, error) {
	return common.PullClusterScopedBuilder[lcasgv1.SeedGenerator, SeedGeneratorBuilder](
		context.TODO(), apiClient, lcasgv1.AddToScheme, name)
}

// WithSeedImage sets the seed image used by the seedgenerator.
func (builder *SeedGeneratorBuilder) WithSeedImage(
	seedImage string) *SeedGeneratorBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if seedImage == "" {
		klog.V(100).Info("The name of the seedImage is empty")

		builder.SetError(fmt.Errorf("seedImage 'name' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting seed image %s in seedgenerator", seedImage)

	builder.Definition.Spec.SeedImage = seedImage

	return builder
}

// WithRecertImage sets the recert image used by the seedgenerator.
func (builder *SeedGeneratorBuilder) WithRecertImage(
	recertImage string) *SeedGeneratorBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	if recertImage == "" {
		klog.V(100).Info("The name of the recertImage is empty")

		builder.SetError(fmt.Errorf("recertImage 'name' cannot be empty"))

		return builder
	}

	klog.V(100).Infof("Setting recert image %s in seedgenerator", recertImage)

	builder.Definition.Spec.RecertImage = recertImage

	return builder
}

// WaitUntilComplete waits the specified timeout for the seedgenerator to complete
// actions.
func (builder *SeedGeneratorBuilder) WaitUntilComplete(timeout time.Duration) (*SeedGeneratorBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return builder, err
	}

	klog.V(100).Infof("Waiting for seedgenerator %s to complete actions",
		builder.Definition.Name)

	if !builder.Exists() {
		klog.V(100).Info("The seedgenerator does not exist on the cluster")

		return builder, fmt.Errorf("seedgenerator object %s does not exist", builder.Definition.Name)
	}

	// Polls periodically to determine if seedgenerator is in desired state.
	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second*3, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Status == isTrue && condition.Type == conditionTypeSeedGenCompleted &&
					condition.Reason == isComplete {
					return true, nil
				}
			}

			return false, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}
