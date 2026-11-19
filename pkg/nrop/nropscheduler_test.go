package nrop

import (
	"testing"

	nropv1 "github.com/openshift-kni/numaresources-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultNROSchedulerName      = "numaresourcesscheduler"
	defaultNROSchedulerNamespace = "openshift-numaresources"
	defaultImageSpec             = "test-registry.com/noderesourcetopology-scheduler-image:latest"
	defaultSchedulerName         = "topo-aware-scheduler"
)

func TestNewSchedulerBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewSchedulerBuilder,
		nropv1.AddToScheme,
		schedulerGVK,
	).ExecuteTests(t)
}

func TestPullScheduler(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(PullScheduler, nropv1.AddToScheme, schedulerGVK).ExecuteTests(t)
}

func TestSchedulerBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[nropv1.NUMAResourcesScheduler, SchedulerBuilder](
		nropv1.AddToScheme,
		schedulerGVK,
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

func TestWithImageSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		imageSpec      string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:      "valid image spec",
			imageSpec: defaultImageSpec,
		},
		{
			name:          "empty image spec",
			expectedError: "can not apply a NUMAResourcesScheduler with an empty imageSpec",
		},
		{
			name:           "invalid builder",
			imageSpec:      defaultImageSpec,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testBuilder *SchedulerBuilder
			if testCase.invalidBuilder {
				testBuilder = buildInvalidSchedulerBuilder(getTestSchedulerAPIClient())
			} else {
				testBuilder = buildValidSchedulerBuilder(getTestSchedulerAPIClient())
			}

			testBuilder.WithImageSpec(testCase.imageSpec)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.SchedulerImage)

				return
			}

			if testCase.expectedError != "" {
				require.Error(t, testBuilder.GetError())
				assert.Equal(t, testCase.expectedError, testBuilder.GetError().Error())
				assert.Empty(t, testBuilder.Definition.Spec.SchedulerImage)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.imageSpec, testBuilder.Definition.Spec.SchedulerImage)
		})
	}
}

func TestWithSchedulerName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		schedulerName  string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:          "valid scheduler name",
			schedulerName: defaultSchedulerName,
		},
		{
			name:          "empty scheduler name",
			expectedError: "can not apply a NUMAResourcesScheduler with an empty schedulerName",
		},
		{
			name:           "invalid builder",
			schedulerName:  defaultSchedulerName,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testBuilder *SchedulerBuilder
			if testCase.invalidBuilder {
				testBuilder = buildInvalidSchedulerBuilder(getTestSchedulerAPIClient())
			} else {
				testBuilder = buildValidSchedulerBuilder(getTestSchedulerAPIClient())
			}

			testBuilder.WithSchedulerName(testCase.schedulerName)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNameEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.SchedulerName)

				return
			}

			if testCase.expectedError != "" {
				require.Error(t, testBuilder.GetError())
				assert.Equal(t, testCase.expectedError, testBuilder.GetError().Error())
				assert.Empty(t, testBuilder.Definition.Spec.SchedulerName)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.schedulerName, testBuilder.Definition.Spec.SchedulerName)
		})
	}
}

func buildValidSchedulerBuilder(apiClient *clients.Settings) *SchedulerBuilder {
	return NewSchedulerBuilder(apiClient, defaultNROSchedulerName, defaultNROSchedulerNamespace)
}

func buildInvalidSchedulerBuilder(apiClient *clients.Settings) *SchedulerBuilder {
	return NewSchedulerBuilder(apiClient, "", defaultNROSchedulerNamespace)
}

func getTestSchedulerAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}
