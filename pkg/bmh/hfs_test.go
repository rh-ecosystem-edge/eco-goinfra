package bmh

import (
	"testing"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultHFSName      = "hostfirmwaresettings-test"
	defaultHFSNamespace = "test-ns"
)

func TestPullHFS(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullHFS,
		bmhv1alpha1.AddToScheme,
		hfsGVK,
	).ExecuteTests(t)
}

func TestHFSBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[bmhv1alpha1.HostFirmwareSettings, HFSBuilder](
		bmhv1alpha1.AddToScheme,
		hfsGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestHFSWithSettings(t *testing.T) {
	t.Parallel()

	settings := map[string]string{
		"SecureBoot":        "Enabled",
		"SriovGlobalEnable": "Enabled",
	}

	testCases := []struct {
		name     string
		valid    bool
		settings map[string]string
	}{
		{
			name:     "sets firmware settings",
			valid:    true,
			settings: settings,
		},
		{
			name:     "initializes empty settings map",
			valid:    true,
			settings: nil,
		},
		{
			name:     "invalid builder short-circuits",
			valid:    false,
			settings: settings,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := clients.GetTestClients(clients.TestClientParams{})

			builder := buildValidHFSBuilder(apiClient)
			if !testCase.valid {
				builder = buildInvalidHFSBuilder(apiClient)
			}

			result := builder.WithSettings(testCase.settings)

			if !testCase.valid {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Settings)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Settings)

			for key, value := range testCase.settings {
				setting := result.Definition.Spec.Settings[key]
				assert.Equal(t, value, setting.String())
			}
		})
	}
}

func buildValidHFSBuilder(apiClient *clients.Settings) *HFSBuilder {
	return common.NewNamespacedBuilder[bmhv1alpha1.HostFirmwareSettings, HFSBuilder](
		apiClient, bmhv1alpha1.AddToScheme, defaultHFSName, defaultHFSNamespace)
}

func buildInvalidHFSBuilder(apiClient *clients.Settings) *HFSBuilder {
	return common.NewNamespacedBuilder[bmhv1alpha1.HostFirmwareSettings, HFSBuilder](
		apiClient, bmhv1alpha1.AddToScheme, "", defaultHFSNamespace)
}
