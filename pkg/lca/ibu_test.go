package lca

import (
	"fmt"
	"testing"

	lcav1 "github.com/openshift-kni/lifecycle-agent/api/imagebasedupgrade/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var imageBasedUpgradeGVK = lcav1.GroupVersion.WithKind("ImageBasedUpgrade")

func TestPullImageBasedUpgrade(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		PullImageBasedUpgrade,
		lcav1.AddToScheme,
		imageBasedUpgradeGVK,
		ibuName,
	).ExecuteTests(t)
}

func TestImageBasedUpgradeBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[lcav1.ImageBasedUpgrade, ImageBasedUpgradeBuilder](
		lcav1.AddToScheme,
		imageBasedUpgradeGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestImageBasedUpgradeWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[lcav1.ImageBasedUpgrade, ImageBasedUpgradeBuilder](
			lcav1.AddToScheme,
			imageBasedUpgradeGVK,
			testhelper.ResourceScopeClusterScoped,
		)).ExecuteTests(t)
}

//nolint:funlen // Table-driven coverage for all ImageBasedUpgrade modifiers.
func TestImageBasedUpgradeModifiers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		prepare       func(*ImageBasedUpgradeBuilder)
		modify        func(*ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder
		assertApplied func(*testing.T, *ImageBasedUpgradeBuilder)
	}{
		{
			name: "with seed image",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithSeedImage("quay.io/foo")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, "quay.io/foo", builder.Definition.Spec.SeedImageRef.Image)
			},
		},
		{
			name: "with empty seed image",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithSeedImage("")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Empty(t, builder.Definition.Spec.SeedImageRef.Image)
			},
		},
		{
			name: "with extra manifests",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithExtraManifests("cmName", "nsName")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, []lcav1.ConfigMapRef{{Name: "cmName", Namespace: "nsName"}},
					builder.Definition.Spec.ExtraManifests)
			},
		},
		{
			name: "with empty extra manifests",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithExtraManifests("", "")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, []lcav1.ConfigMapRef{{}}, builder.Definition.Spec.ExtraManifests)
			},
		},
		{
			name: "with OADP content",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithOadpContent("cmName", "nsName")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, []lcav1.ConfigMapRef{{Name: "cmName", Namespace: "nsName"}},
					builder.Definition.Spec.OADPContent)
			},
		},
		{
			name: "with empty OADP content",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithOadpContent("", "")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, []lcav1.ConfigMapRef{{}}, builder.Definition.Spec.OADPContent)
			},
		},
		{
			name: "with auto rollback timeout",
			prepare: func(builder *ImageBasedUpgradeBuilder) {
				builder.Definition.Spec.AutoRollbackOnFailure = &lcav1.AutoRollbackOnFailure{}
			},
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.AutoRollbackOnFailureInitMonitorTimeoutSeconds(1800)
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, 1800, builder.Definition.Spec.AutoRollbackOnFailure.InitMonitorTimeoutSeconds)
			},
		},
		{
			name: "disable init monitor rollback",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.AutoRollbackOnFailureDisabledInitMonitor()
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, disabled,
					builder.Definition.Annotations["auto-rollback-on-failure.lca.openshift.io/init-monitor"])
			},
		},
		{
			name: "disable post reboot rollback",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.AutoRollbackOnFailureDisableForPostReboot()
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, disabled,
					builder.Definition.Annotations["auto-rollback-on-failure.lca.openshift.io/post-reboot-config"])
			},
		},
		{
			name: "disable upgrade completion rollback",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.AutoRollbackOnFailureDisableForUpgradeCompletion()
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, disabled,
					builder.Definition.Annotations["auto-rollback-on-failure.lca.openshift.io/upgrade-completion"])
			},
		},
		{
			name: "with seed image version",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithSeedImageVersion("seedversion")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, "seedversion", builder.Definition.Spec.SeedImageRef.Version)
			},
		},
		{
			name: "with empty seed image version",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithSeedImageVersion("")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Empty(t, builder.Definition.Spec.SeedImageRef.Version)
			},
		},
		{
			name: "with seed image pull secret",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithSeedImagePullSecretRef("pull-secret")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				require.NotNil(t, builder.Definition.Spec.SeedImageRef.PullSecretRef)
				assert.Equal(t, "pull-secret", builder.Definition.Spec.SeedImageRef.PullSecretRef.Name)
			},
		},
		{
			name: "with empty seed image pull secret",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithSeedImagePullSecretRef("")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				require.NotNil(t, builder.Definition.Spec.SeedImageRef.PullSecretRef)
				assert.Empty(t, builder.Definition.Spec.SeedImageRef.PullSecretRef.Name)
			},
		},
		{
			name: "with stage",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithStage(idle)
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, lcav1.ImageBasedUpgradeStage(idle), builder.Definition.Spec.Stage)
			},
		},
		{
			name: "with unrecognized stage",
			modify: func(builder *ImageBasedUpgradeBuilder) *ImageBasedUpgradeBuilder {
				return builder.WithStage("Wrong")
			},
			assertApplied: func(t *testing.T, builder *ImageBasedUpgradeBuilder) {
				t.Helper()
				assert.Equal(t, lcav1.ImageBasedUpgradeStage("Wrong"), builder.Definition.Spec.Stage)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidImageBasedUpgradeTestBuilder(newImageBasedUpgradeTestClient(nil))
			if testCase.prepare != nil {
				testCase.prepare(builder)
			}

			builder = testCase.modify(builder)
			require.NoError(t, builder.GetError())
			testCase.assertApplied(t, builder)

			invalidBuilder := testCase.modify(buildInvalidImageBasedUpgradeTestBuilder(newImageBasedUpgradeTestClient(nil)))
			require.Error(t, invalidBuilder.GetError())
			assert.True(t, commonerrors.IsBuilderNameEmpty(invalidBuilder.GetError()))
		})
	}
}

func TestImageBasedUpgradeWaitUntilStageComplete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		stage         string
		objects       []runtime.Object
		expectedError string
	}{
		{
			name:          "wrong stage",
			stage:         "wrong_stage",
			objects:       buildDummyImageBasedUpgradeRuntime(),
			expectedError: "wrong stage selected for imagebasedupgrade",
		},
		{
			name:    "idle stage completes",
			stage:   idle,
			objects: buildDummyImageBasedUpgradeRuntime(),
		},
		{
			name:          "resource does not exist",
			stage:         idle,
			expectedError: fmt.Sprintf("imagebasedupgrade object %s does not exist", ibuName),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidImageBasedUpgradeTestBuilder(newImageBasedUpgradeTestClient(testCase.objects))
			result, err := builder.WaitUntilStageComplete(testCase.stage)

			if testCase.expectedError == "" {
				require.NoError(t, err)
				require.NotNil(t, result)
			} else {
				require.EqualError(t, err, testCase.expectedError)
				require.NotNil(t, result)
			}
		})
	}
}

func buildValidImageBasedUpgradeTestBuilder(apiClient *clients.Settings) *ImageBasedUpgradeBuilder {
	return common.NewClusterScopedBuilder[lcav1.ImageBasedUpgrade, ImageBasedUpgradeBuilder](
		apiClient, lcav1.AddToScheme, ibuName)
}

func buildInvalidImageBasedUpgradeTestBuilder(apiClient *clients.Settings) *ImageBasedUpgradeBuilder {
	return common.NewClusterScopedBuilder[lcav1.ImageBasedUpgrade, ImageBasedUpgradeBuilder](
		apiClient, lcav1.AddToScheme, "")
}

func newImageBasedUpgradeTestClient(objects []runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{lcav1.AddToScheme},
	})
}

func buildDummyImageBasedUpgradeRuntime() []runtime.Object {
	return []runtime.Object{
		&lcav1.ImageBasedUpgrade{
			ObjectMeta: metav1.ObjectMeta{Name: ibuName},
			Spec:       lcav1.ImageBasedUpgradeSpec{Stage: idle},
			Status: lcav1.ImageBasedUpgradeStatus{
				Conditions: []metav1.Condition{{Type: idle, Status: isTrue}},
			},
		},
	}
}
