package nrop

import (
	"testing"

	nropv1 "github.com/openshift-kni/numaresources-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultNROPName = "numaresourcesoperator"

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedBuilderTestConfig(NewBuilder, nropv1.AddToScheme, nropGVK).ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(Pull, nropv1.AddToScheme, nropGVK).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[nropv1.NUMAResourcesOperator, Builder](
		nropv1.AddToScheme,
		nropGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

//nolint:funlen // The selector test table covers each refresh mode plus package-specific error paths.
func TestWithMCPSelector(t *testing.T) {
	t.Parallel()

	periodic := nropv1.InfoRefreshPeriodic
	events := nropv1.InfoRefreshEvents
	periodicAndEvents := nropv1.InfoRefreshPeriodicAndEvents

	predefinedNodeGroup := nropv1.NodeGroup{
		Config: &nropv1.NodeGroupConfig{InfoRefreshMode: &periodic},
		MachineConfigPoolSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"existing-selector": "existing-value"},
		},
	}

	testCases := []struct {
		name               string
		config             nropv1.NodeGroupConfig
		mcpSelector        metav1.LabelSelector
		existingNodeGroups []nropv1.NodeGroup
		invalidBuilder     bool
		expectedError      string
	}{
		{
			name:        "periodic refresh mode",
			config:      nropv1.NodeGroupConfig{InfoRefreshMode: &periodic},
			mcpSelector: metav1.LabelSelector{MatchLabels: map[string]string{"role": "worker"}},
		},
		{
			name:        "events refresh mode",
			config:      nropv1.NodeGroupConfig{InfoRefreshMode: &events},
			mcpSelector: metav1.LabelSelector{MatchLabels: map[string]string{"role": "worker"}},
		},
		{
			name:        "periodic and events refresh mode",
			config:      nropv1.NodeGroupConfig{InfoRefreshMode: &periodicAndEvents},
			mcpSelector: metav1.LabelSelector{MatchLabels: map[string]string{"role": "worker"}},
		},
		{
			name:        "empty refresh mode",
			config:      nropv1.NodeGroupConfig{},
			mcpSelector: metav1.LabelSelector{MatchLabels: map[string]string{"role": "worker"}},
		},
		{
			name:   "match expression selector",
			config: nropv1.NodeGroupConfig{InfoRefreshMode: &periodic},
			mcpSelector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "role",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"worker"},
				}},
			},
		},
		{
			name:          "empty selector",
			config:        nropv1.NodeGroupConfig{InfoRefreshMode: &periodic},
			expectedError: "NUMAResourcesOperator 'machineConfigPoolSelector' cannot be empty",
		},
		{
			name:               "appends to existing node groups",
			config:             nropv1.NodeGroupConfig{InfoRefreshMode: &periodic},
			mcpSelector:        metav1.LabelSelector{MatchLabels: map[string]string{"role": "worker"}},
			existingNodeGroups: []nropv1.NodeGroup{predefinedNodeGroup},
		},
		{
			name:           "invalid builder",
			config:         nropv1.NodeGroupConfig{InfoRefreshMode: &periodic},
			mcpSelector:    metav1.LabelSelector{MatchLabels: map[string]string{"role": "worker"}},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testBuilder *Builder
			if testCase.invalidBuilder {
				testBuilder = buildInvalidNROPBuilder(getTestNROPAPIClient())
			} else {
				testBuilder = buildValidNROPBuilder(getTestNROPAPIClient())
			}

			testBuilder.Definition.Spec.NodeGroups = testCase.existingNodeGroups

			testBuilder.WithMCPSelector(testCase.config, testCase.mcpSelector)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(testBuilder.GetError()))
				assert.Equal(t, testCase.existingNodeGroups, testBuilder.Definition.Spec.NodeGroups)

				return
			}

			if testCase.expectedError != "" {
				require.Error(t, testBuilder.GetError())
				assert.Equal(t, testCase.expectedError, testBuilder.GetError().Error())
				assert.Equal(t, testCase.existingNodeGroups, testBuilder.Definition.Spec.NodeGroups)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.Len(t, testBuilder.Definition.Spec.NodeGroups, len(testCase.existingNodeGroups)+1)

			nodeGroup := testBuilder.Definition.Spec.NodeGroups[len(testCase.existingNodeGroups)]
			require.NotNil(t, nodeGroup.Config)
			require.NotNil(t, nodeGroup.MachineConfigPoolSelector)
			assert.Equal(t, testCase.config, *nodeGroup.Config)
			assert.Equal(t, testCase.mcpSelector, *nodeGroup.MachineConfigPoolSelector)

			if len(testCase.existingNodeGroups) > 0 {
				assert.Equal(t, testCase.existingNodeGroups, testBuilder.Definition.Spec.NodeGroups[:len(testCase.existingNodeGroups)])
			}
		})
	}
}

func buildValidNROPBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultNROPName)
}

func buildInvalidNROPBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "")
}

func getTestNROPAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}
