package lso

import (
	"testing"

	lsov1 "github.com/openshift/local-storage-operator/api/v1"
	lsov1alpha1 "github.com/openshift/local-storage-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultLocalVolumeSetName      = "ocs-deviceset"
	defaultLocalVolumeSetNamespace = "test-lvsspace"
)

func TestNewLocalVolumeSetBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[
		lsov1alpha1.LocalVolumeSet, LocalVolumeSetBuilder,
	](NewLocalVolumeSetBuilder, lsov1alpha1.AddToScheme, localVolumeSetGVK).ExecuteTests(t)
}

func TestPullLocalVolumeSet(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[
		lsov1alpha1.LocalVolumeSet, LocalVolumeSetBuilder,
	](PullLocalVolumeSet, lsov1alpha1.AddToScheme, localVolumeSetGVK).ExecuteTests(t)
}

func TestLocalVolumeSetBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[
		lsov1alpha1.LocalVolumeSet, LocalVolumeSetBuilder,
	](lsov1alpha1.AddToScheme, localVolumeSetGVK, testhelper.ResourceScopeNamespaced)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestLocalVolumeSetWithTolerations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		tolerations    []corev1.Toleration
		invalidBuilder bool
		expectedError  string
	}{
		{
			name: "valid tolerations",
			tolerations: []corev1.Toleration{{
				Key:      "node.ocs.openshift.io/storage",
				Operator: corev1.TolerationOpEqual,
				Value:    "true",
				Effect:   corev1.TaintEffectNoSchedule,
			}},
		},
		{
			name:          "empty tolerations",
			tolerations:   []corev1.Toleration{},
			expectedError: errEmptyTolerations,
		},
		{
			name:           "invalid builder short circuits",
			tolerations:    []corev1.Toleration{},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *LocalVolumeSetBuilder
			if testCase.invalidBuilder {
				builder = buildInvalidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			} else {
				builder = buildValidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			}

			initialError := builder.GetError()
			result := builder.WithTolerations(testCase.tolerations)

			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Tolerations)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Spec.Tolerations)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.tolerations, result.Definition.Spec.Tolerations)
		})
	}
}

func TestLocalVolumeSetWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		nodeSelector   corev1.NodeSelector
		invalidBuilder bool
	}{
		{
			name: "node selector with matching expression",
			nodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{""},
					}},
				}},
			},
		},
		{
			name: "node selector with exists expression",
			nodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "cluster.ocs.openshift.io/openshift-storage",
						Operator: corev1.NodeSelectorOpExists,
					}},
				}},
			},
		},
		{
			name: "node selector with multiple expressions",
			nodeSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "cluster.ocs.openshift.io/openshift-storage",
							Operator: corev1.NodeSelectorOpExists,
						},
						{
							Key:      "machineconfiguration.openshift.io/role",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"customcnf", "worker"},
						},
					},
				}},
			},
		},
		{
			name:           "invalid builder short circuits",
			nodeSelector:   corev1.NodeSelector{},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *LocalVolumeSetBuilder
			if testCase.invalidBuilder {
				builder = buildInvalidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			} else {
				builder = buildValidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			}

			initialError := builder.GetError()
			result := builder.WithNodeSelector(testCase.nodeSelector)

			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.NodeSelector)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.NodeSelector)
			assert.Equal(t, testCase.nodeSelector, *result.Definition.Spec.NodeSelector)
		})
	}
}

func TestLocalVolumeSetWithStorageClassName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		storageClassName string
		invalidBuilder   bool
		expectedError    string
	}{
		{
			name:             "valid storage class name",
			storageClassName: "test-storage-class",
		},
		{
			name:          "empty storage class name",
			expectedError: "'storageClassName' argument cannot be empty",
		},
		{name: "invalid builder short circuits", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *LocalVolumeSetBuilder
			if testCase.invalidBuilder {
				builder = buildInvalidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			} else {
				builder = buildValidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			}

			initialError := builder.GetError()
			result := builder.WithStorageClassName(testCase.storageClassName)

			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.StorageClassName)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Empty(t, result.Definition.Spec.StorageClassName)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.storageClassName, result.Definition.Spec.StorageClassName)
		})
	}
}

func TestLocalVolumeSetWithVolumeMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		volumeMode     lsov1.PersistentVolumeMode
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:       "block volume mode",
			volumeMode: lsov1.PersistentVolumeBlock,
		},
		{
			name:       "filesystem volume mode",
			volumeMode: lsov1.PersistentVolumeFilesystem,
		},
		{
			name:          "empty volume mode",
			expectedError: "'volumeMode' argument cannot be empty",
		},
		{name: "invalid builder short circuits", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *LocalVolumeSetBuilder
			if testCase.invalidBuilder {
				builder = buildInvalidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			} else {
				builder = buildValidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			}

			initialError := builder.GetError()
			result := builder.WithVolumeMode(testCase.volumeMode)

			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.VolumeMode)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Empty(t, result.Definition.Spec.VolumeMode)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.volumeMode, result.Definition.Spec.VolumeMode)
		})
	}
}

func TestLocalVolumeSetWithFSType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		fstype         string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:   "valid filesystem type",
			fstype: "ext4",
		},
		{
			name:          "empty filesystem type",
			expectedError: "'fstype' argument cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *LocalVolumeSetBuilder
			if testCase.invalidBuilder {
				builder = buildInvalidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			} else {
				builder = buildValidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			}

			initialError := builder.GetError()
			result := builder.WithFSType(testCase.fstype)

			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.FSType)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Empty(t, result.Definition.Spec.FSType)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.fstype, result.Definition.Spec.FSType)
		})
	}
}

func TestLocalVolumeSetWithMaxDeviceCount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		maxDeviceCount int32
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:           "valid max device count",
			maxDeviceCount: 42,
		},
		{
			name:          "zero max device count",
			expectedError: "'maxDeviceCount' argument cannot be equal zero",
		},
		{
			name:           "invalid builder short circuits",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *LocalVolumeSetBuilder
			if testCase.invalidBuilder {
				builder = buildInvalidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			} else {
				builder = buildValidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			}

			initialError := builder.GetError()
			result := builder.WithMaxDeviceCount(testCase.maxDeviceCount)

			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.MaxDeviceCount)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Spec.MaxDeviceCount)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.MaxDeviceCount)
			assert.Equal(t, testCase.maxDeviceCount, *result.Definition.Spec.MaxDeviceCount)
		})
	}
}

func TestLocalVolumeSetWithDeviceInclusionSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		deviceInclusionSpec lsov1alpha1.DeviceInclusionSpec
		invalidBuilder      bool
	}{
		{
			name: "raw non-rotational devices",
			deviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes:                []lsov1alpha1.DeviceType{lsov1alpha1.RawDisk},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.NonRotational},
			},
		},
		{
			name: "rotational partitions",
			deviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes:                []lsov1alpha1.DeviceType{lsov1alpha1.Partition},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.Rotational},
			},
		},
		{
			name: "loop devices",
			deviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes:                []lsov1alpha1.DeviceType{lsov1alpha1.Loop},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.Rotational},
			},
		},
		{
			name: "multipath devices",
			deviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes:                []lsov1alpha1.DeviceType{lsov1alpha1.MultiPath},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.NonRotational},
			},
		},
		{
			name: "raw and multipath devices",
			deviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes: []lsov1alpha1.DeviceType{
					lsov1alpha1.RawDisk,
					lsov1alpha1.MultiPath,
				},
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.Rotational},
			},
		},
		{
			name: "device types only",
			deviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceTypes: []lsov1alpha1.DeviceType{lsov1alpha1.RawDisk},
			},
		},
		{
			name: "mechanical properties only",
			deviceInclusionSpec: lsov1alpha1.DeviceInclusionSpec{
				DeviceMechanicalProperties: []lsov1alpha1.DeviceMechanicalProperty{lsov1alpha1.NonRotational},
			},
		},
		{name: "invalid builder short circuits", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *LocalVolumeSetBuilder
			if testCase.invalidBuilder {
				builder = buildInvalidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			} else {
				builder = buildValidLocalVolumeSetTestBuilder(getTestLocalVolumeSetAPIClient())
			}

			initialError := builder.GetError()
			result := builder.WithDeviceInclusionSpec(testCase.deviceInclusionSpec)

			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				assert.Equal(t, initialError, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.DeviceInclusionSpec)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.DeviceInclusionSpec)
			assert.Equal(t, testCase.deviceInclusionSpec, *result.Definition.Spec.DeviceInclusionSpec)
		})
	}
}

func buildValidLocalVolumeSetTestBuilder(apiClient *clients.Settings) *LocalVolumeSetBuilder {
	return NewLocalVolumeSetBuilder(apiClient, defaultLocalVolumeSetName, defaultLocalVolumeSetNamespace)
}

func buildInvalidLocalVolumeSetTestBuilder(apiClient *clients.Settings) *LocalVolumeSetBuilder {
	return NewLocalVolumeSetBuilder(apiClient, "", defaultLocalVolumeSetNamespace)
}

func getTestLocalVolumeSetAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{lsov1alpha1.AddToScheme},
	})
}
