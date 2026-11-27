package poddisruptionbudget

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	policyv1 "k8s.io/api/policy/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

const (
	defaultPDBName      = "poddisruptionbudget-test"
	defaultPDBNamespace = "test-namespace"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(NewBuilder, policyv1.AddToScheme, pdbGVK).ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(Pull, policyv1.AddToScheme, pdbGVK).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[policyv1.PodDisruptionBudget, Builder](
		policyv1.AddToScheme,
		pdbGVK,
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

func TestWithPDBSpec(t *testing.T) {
	t.Parallel()

	minAvailable := intstr.FromInt(1)
	spec := policyv1.PodDisruptionBudgetSpec{MinAvailable: &minAvailable}

	testCases := []struct {
		name          string
		builder       *Builder
		expectedNil   bool
		definitionNil bool
		expectedError func(error) bool
	}{
		{
			name:    "valid builder",
			builder: buildValidPDBTestBuilder(getPDBTestClient()),
		},
		{
			name:        "nil builder",
			builder:     nil,
			expectedNil: true,
		},
		{
			name:          "nil definition",
			definitionNil: true,
			builder: func() *Builder {
				builder := buildValidPDBTestBuilder(getPDBTestClient())
				builder.Definition = nil

				return builder
			}(),
		},
		{
			name:          "invalid builder",
			builder:       buildInvalidPDBTestBuilder(getPDBTestClient()),
			expectedError: commonerrors.IsBuilderNamespaceEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithPDBSpec(spec)

			if testCase.expectedNil {
				assert.Nil(t, result)

				return
			}

			if testCase.definitionNil {
				require.NotNil(t, result)
				assert.Nil(t, result.Definition)

				return
			}

			if testCase.expectedError != nil {
				require.NotNil(t, result)
				assert.True(t, testCase.expectedError(result.GetError()))

				return
			}

			require.NotNil(t, result)
			require.NoError(t, result.GetError())
			assert.Equal(t, spec, result.Definition.Spec)
		})
	}
}

// buildValidPDBTestBuilder returns a valid PodDisruptionBudget builder for testing.
func buildValidPDBTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultPDBName, defaultPDBNamespace)
}

// buildInvalidPDBTestBuilder returns an invalid PodDisruptionBudget builder for testing.
func buildInvalidPDBTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultPDBName, "")
}
