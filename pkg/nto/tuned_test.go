package nto //nolint:misspell

import (
	"testing"

	tunedv1 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/tuned/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultTunedName        = "default"
	defaultTunedNamespace   = "openshift-cluster-node-tuning-operator"
	defaultTunedProfileName = "openshift"
	defaultTunedProfileData = "[main]\nsummary=Optimize systems running OpenShift (provider specific parent profile)" +
		"\ninclude=-provider-${f:exec:cat:/var/lib/ocp-tuned/provider},openshift\n"
	tunedTestSchemes = []clients.SchemeAttacher{
		tunedv1.AddToScheme,
	}
)

func TestPullTuned(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[tunedv1.Tuned, TunedBuilder](
		PullTuned,
		tunedv1.AddToScheme,
		tunedGVK,
	).ExecuteTests(t)
}

func TestNewTunedBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[tunedv1.Tuned, TunedBuilder](
		NewTunedBuilder,
		tunedv1.AddToScheme,
		tunedGVK,
	).ExecuteTests(t)
}

func TestTunedBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[tunedv1.Tuned, TunedBuilder](
		tunedv1.AddToScheme,
		tunedGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestTunedWithProfile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		profile      tunedv1.TunedProfile
		validBuilder bool
	}{
		{
			name: "valid profile",
			profile: tunedv1.TunedProfile{
				Name: &defaultTunedName,
				Data: &defaultTunedProfileData,
			},
			validBuilder: true,
		},
		{
			name:         "empty profile is accepted",
			profile:      tunedv1.TunedProfile{},
			validBuilder: true,
		},
		{
			name: "invalid builder",
			profile: tunedv1.TunedProfile{
				Name: &defaultTunedName,
				Data: &defaultTunedProfileData,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildTunedBuilderForTest(testCase.validBuilder)
			result := builder.WithProfile(testCase.profile)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.Profile)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, []tunedv1.TunedProfile{testCase.profile}, result.Definition.Spec.Profile)
		})
	}
}

func TestTunedWithRecommend(t *testing.T) {
	t.Parallel()

	recommend := tunedv1.TunedRecommend{Profile: &defaultTunedProfileName}

	testCases := []struct {
		name         string
		validBuilder bool
	}{
		{
			name:         "valid recommend",
			validBuilder: true,
		},
		{
			name: "invalid builder",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildTunedBuilderForTest(testCase.validBuilder)
			result := builder.WithRecommend(recommend)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Empty(t, result.Definition.Spec.Recommend)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, []tunedv1.TunedRecommend{recommend}, result.Definition.Spec.Recommend)
		})
	}
}

func buildTunedBuilderForTest(valid bool) *TunedBuilder {
	apiClient := clients.GetTestClients(clients.TestClientParams{SchemeAttachers: tunedTestSchemes})

	if valid {
		return buildValidTunedBuilder(apiClient)
	}

	return buildInvalidTunedBuilder(apiClient)
}

func buildValidTunedBuilder(apiClient *clients.Settings) *TunedBuilder {
	return NewTunedBuilder(apiClient, defaultTunedName, defaultTunedNamespace)
}

func buildInvalidTunedBuilder(apiClient *clients.Settings) *TunedBuilder {
	return NewTunedBuilder(apiClient, "", defaultTunedNamespace)
}
