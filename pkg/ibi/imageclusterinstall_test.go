package ibi

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	ibiv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/imagebasedinstall/api/hiveextensions/v1alpha1"
	hivev1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/imagebasedinstall/hive/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testImageClusterInstall = "test-image-cluster-install"
	testImageSet            = "4.16"
)

func TestNewImageClusterInstallBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		func(apiClient *clients.Settings, name, nsname string) *ImageClusterInstallBuilder {
			return NewImageClusterInstallBuilder(apiClient, name, nsname, testImageSet)
		},
		ibiv1alpha1.AddToScheme,
		ibiv1alpha1.GroupVersion.WithKind("ImageClusterInstall"),
	).ExecuteTests(t)

	t.Run("imageset validation and assignment", func(t *testing.T) {
		t.Parallel()

		apiClient := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: []clients.SchemeAttacher{ibiv1alpha1.AddToScheme},
		})

		builder := NewImageClusterInstallBuilder(apiClient, testImageClusterInstall, testImageClusterInstall, "")
		require.EqualError(t, builder.GetError(), "imageclusterinstall 'imageset' cannot be empty")

		builder = NewImageClusterInstallBuilder(apiClient, testImageClusterInstall, testImageClusterInstall, testImageSet)
		require.NoError(t, builder.GetError())
		assert.Equal(t, testImageSet, builder.Definition.Spec.ImageSetRef.Name)
	})
}

func TestImageClusterInstallPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullImageClusterInstall,
		ibiv1alpha1.AddToScheme,
		ibiv1alpha1.GroupVersion.WithKind("ImageClusterInstall"),
	).ExecuteTests(t)
}

func TestImageClusterInstallMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[ibiv1alpha1.ImageClusterInstall, ImageClusterInstallBuilder](
		ibiv1alpha1.AddToScheme,
		ibiv1alpha1.GroupVersion.WithKind("ImageClusterInstall"),
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestImageClusterInstallWithHostname(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		hostname       string
		invalidBuilder bool
		expectedError  string
	}{
		{name: "valid hostname", hostname: "sno-0-0"},
		{name: "empty hostname", expectedError: "imageclusterinstall hostname cannot be empty"},
		{name: "invalid builder", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := modifierTestBuilder(testCase.invalidBuilder)
			builder.WithHostname(testCase.hostname)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assertModifierResult(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.hostname, builder.Definition.Spec.Hostname)
			}
		})
	}
}

func TestImageClusterInstallWithClusterDeployment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		clusterDeploymentName string
		invalidBuilder        bool
		expectedError         string
	}{
		{name: "valid cluster deployment", clusterDeploymentName: "ibi-cluster-deployment"},
		{
			name:          "empty cluster deployment",
			expectedError: "imageclusterinstall clusterdeployment cannot be empty",
		},
		{name: "invalid builder", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := modifierTestBuilder(testCase.invalidBuilder)
			builder.WithClusterDeployment(testCase.clusterDeploymentName)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assertModifierResult(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				require.NotNil(t, builder.Definition.Spec.ClusterDeploymentRef)
				assert.Equal(t, testCase.clusterDeploymentName, builder.Definition.Spec.ClusterDeploymentRef.Name)
			}
		})
	}
}

func TestImageClusterInstallWithExtraManifests(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		extraManifest  string
		invalidBuilder bool
		expectedError  string
	}{
		{name: "valid extra manifest", extraManifest: "ibi-extra-manifest"},
		{name: "empty extra manifest", expectedError: "imageclusterinstall extramanifest cannot be empty"},
		{name: "invalid builder", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := modifierTestBuilder(testCase.invalidBuilder)
			builder.WithExtraManifests(testCase.extraManifest)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assertModifierResult(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				require.Len(t, builder.Definition.Spec.ExtraManifestsRefs, 1)
				assert.Equal(t, testCase.extraManifest, builder.Definition.Spec.ExtraManifestsRefs[0].Name)
			}
		})
	}
}

func TestImageClusterInstallWithCABundle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		caBundle       string
		invalidBuilder bool
		expectedError  string
	}{
		{name: "valid CA bundle", caBundle: "ibi-ca-bundle"},
		{name: "empty CA bundle", expectedError: "imageclusterinstall cabundle cannot be empty"},
		{name: "invalid builder", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := modifierTestBuilder(testCase.invalidBuilder)
			builder.WithCABundle(testCase.caBundle)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assertModifierResult(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				require.NotNil(t, builder.Definition.Spec.CABundleRef)
				assert.Equal(t, testCase.caBundle, builder.Definition.Spec.CABundleRef.Name)
			}
		})
	}
}

func TestImageClusterInstallWithMachineNetwork(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		machineNetwork string
		invalidBuilder bool
		expectedError  string
	}{
		{name: "valid IPv4 network", machineNetwork: "192.168.0.0/24"},
		{name: "valid IPv6 network", machineNetwork: "fd2e:6f44:5dd8::/64"},
		{
			name:           "invalid IPv4 mask",
			machineNetwork: "192.168.0.0/255.255.255.0",
			expectedError:  errInvalidMachineNetworkFormat,
		},
		{
			name:           "invalid IPv6 network",
			machineNetwork: "fd2e:6f44:5dd8::",
			expectedError:  errInvalidMachineNetworkFormat,
		},
		{name: "empty network", expectedError: errInvalidMachineNetworkFormat},
		{name: "invalid builder", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := modifierTestBuilder(testCase.invalidBuilder)
			builder.WithMachineNetwork(testCase.machineNetwork)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assertModifierResult(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.machineNetwork, builder.Definition.Spec.MachineNetwork)
			}
		})
	}
}

func TestImageClusterInstallWithSSHKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		sshKey         string
		invalidBuilder bool
		expectedError  string
	}{
		{name: "valid SSH key", sshKey: "mysecretsshkey"},
		{name: "empty SSH key", expectedError: "imageclusterinstall sshkey cannot be empty"},
		{name: "invalid builder", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := modifierTestBuilder(testCase.invalidBuilder)
			builder.WithSSHKey(testCase.sshKey)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assertModifierResult(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.sshKey, builder.Definition.Spec.SSHKey)
			}
		})
	}
}

func TestImageClusterInstallConditions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		conditionType hivev1.ClusterInstallConditionType
		getCondition  func(*ImageClusterInstallBuilder) (*hivev1.ClusterInstallCondition, error)
	}{
		{"Completed", hivev1.ClusterInstallCompleted, (*ImageClusterInstallBuilder).GetCompletedCondition},
		{"Failed", hivev1.ClusterInstallFailed, (*ImageClusterInstallBuilder).GetFailedCondition},
		{
			"RequirementsMet",
			hivev1.ClusterInstallRequirementsMet,
			(*ImageClusterInstallBuilder).GetRequirementsMetCondition,
		},
		{"Stopped", hivev1.ClusterInstallStopped, (*ImageClusterInstallBuilder).GetStoppedCondition},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			object := generateImageClusterInstall()
			object.Status.Conditions = []hivev1.ClusterInstallCondition{{
				Type:    testCase.conditionType,
				Message: "expected condition",
			}}

			builder := generateImageClusterInstallBuilderWithFakeObjects([]runtime.Object{object})
			condition, err := testCase.getCondition(builder)
			require.NoError(t, err)
			require.NotNil(t, condition)
			assert.Equal(t, "expected condition", condition.Message)

			object.Status.Conditions = nil
			builder = generateImageClusterInstallBuilderWithFakeObjects([]runtime.Object{object})
			condition, err = testCase.getCondition(builder)
			require.EqualError(t, err, "cannot find "+testCase.name+" condition in imageclusterinstall status")
			assert.Nil(t, condition)

			builder = generateImageClusterInstallBuilderWithFakeObjects(nil)
			condition, err = testCase.getCondition(builder)
			require.EqualError(t, err, "cannot get condition from non-existent imageclusterinstall")
			assert.Nil(t, condition)
		})
	}
}

func assertModifierResult(t *testing.T, err error, expectedError string) {
	t.Helper()

	if expectedError == "" {
		require.NoError(t, err)

		return
	}

	require.EqualError(t, err, expectedError)
}

func modifierTestBuilder(invalid bool) *ImageClusterInstallBuilder {
	apiClient := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{ibiv1alpha1.AddToScheme},
	})

	if invalid {
		return NewImageClusterInstallBuilder(apiClient, "", testImageClusterInstall, testImageSet)
	}

	return NewImageClusterInstallBuilder(
		apiClient, testImageClusterInstall, testImageClusterInstall, testImageSet)
}

func generateImageClusterInstallBuilderWithFakeObjects(objects []runtime.Object) *ImageClusterInstallBuilder {
	return NewImageClusterInstallBuilder(
		clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  objects,
			SchemeAttachers: []clients.SchemeAttacher{ibiv1alpha1.AddToScheme},
		}),
		testImageClusterInstall,
		testImageClusterInstall,
		testImageSet,
	)
}

func generateImageClusterInstall() *ibiv1alpha1.ImageClusterInstall {
	return &ibiv1alpha1.ImageClusterInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testImageClusterInstall,
			Namespace: testImageClusterInstall,
		},
		Spec: ibiv1alpha1.ImageClusterInstallSpec{
			ImageSetRef: hivev1.ClusterImageSetReference{Name: testImageSet},
		},
	}
}
