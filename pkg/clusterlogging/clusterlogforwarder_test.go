package clusterlogging

import (
	"testing"

	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultClusterLogForwarderName      = "instance"
	defaultClusterLogForwarderNamespace = "openshift-logging"
	metaDataNameErrorMgs                = "metadata.name: Required value: name is required"
	newOutputs                          = observabilityv1.OutputSpec{
		Name: "elasticsearch-external",
		Type: "elasticsearch",
		Kafka: &observabilityv1.Kafka{
			URL: "https://dummy-domain.amazonaws.com:443",
		},
	}
	newPipelines = observabilityv1.PipelineSpec{
		OutputRefs: []string{"elasticsearch-external"},
		InputRefs:  []string{"application", "infra"},
	}
)

func TestNewClusterLogForwarderBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewClusterLogForwarderBuilder,
		observabilityv1.AddToScheme,
		clusterLogForwarderGVK,
	).ExecuteTests(t)
}

func TestPullClusterLogForwarder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullClusterLogForwarder,
		observabilityv1.AddToScheme,
		clusterLogForwarderGVK,
	).ExecuteTests(t)
}

func TestClusterLogForwarderBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[observabilityv1.ClusterLogForwarder, ClusterLogForwarderBuilder](
		observabilityv1.AddToScheme,
		clusterLogForwarderGVK,
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

func TestWithManagementState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		managementState observabilityv1.ManagementState
		valid           bool
		expectedError   string
	}{
		{
			name:            "managed",
			managementState: observabilityv1.ManagementStateManaged,
			valid:           true,
		},
		{
			name:            "unmanaged",
			managementState: observabilityv1.ManagementStateUnmanaged,
			valid:           true,
		},
		{
			name:            "unsupported state",
			managementState: "",
			valid:           true,
			expectedError: "the management state of the clusterlogforwarder is unsupported: \"\";" +
				"accepted only Managed or Unmanaged states",
		},
		{
			name:            "invalid builder",
			managementState: observabilityv1.ManagementStateManaged,
			valid:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildClusterLogForwarderBuilder(testCase.valid)
			result := builder.WithManagementState(testCase.managementState)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.managementState, result.Definition.Spec.ManagementState)
		})
	}
}

func TestWithServiceAccount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		serviceAccount string
		valid          bool
		expectedError  string
	}{
		{
			name:           "valid service account",
			serviceAccount: "dummy",
			valid:          true,
		},
		{
			name:           "empty service account",
			serviceAccount: "",
			valid:          true,
			expectedError:  "clusterlogforwarder 'serviceAccount' cannot be empty",
		},
		{
			name:           "invalid builder",
			serviceAccount: "dummy",
			valid:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildClusterLogForwarderBuilder(testCase.valid)
			result := builder.WithServiceAccount(testCase.serviceAccount)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.serviceAccount, result.Definition.Spec.ServiceAccount.Name)
		})
	}
}

func TestWithOutput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		output        *observabilityv1.OutputSpec
		valid         bool
		expectedError string
	}{
		{
			name:   "valid output",
			output: &newOutputs,
			valid:  true,
		},
		{
			name:          "nil output",
			valid:         true,
			expectedError: "'outputSpec' parameter is empty",
		},
		{
			name:   "invalid builder",
			output: &newOutputs,
			valid:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildClusterLogForwarderBuilder(testCase.valid)
			result := builder.WithOutput(testCase.output)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, []observabilityv1.OutputSpec{*testCase.output}, result.Definition.Spec.Outputs)
		})
	}
}

func TestWithPipeline(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		pipeline      *observabilityv1.PipelineSpec
		valid         bool
		expectedError string
	}{
		{
			name:     "valid pipeline",
			pipeline: &newPipelines,
			valid:    true,
		},
		{
			name:          "nil pipeline",
			valid:         true,
			expectedError: "'pipelineSpec' parameter is empty",
		},
		{
			name:     "invalid builder",
			pipeline: &newPipelines,
			valid:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildClusterLogForwarderBuilder(testCase.valid)
			result := builder.WithPipeline(testCase.pipeline)

			if !testCase.valid {
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, []observabilityv1.PipelineSpec{*testCase.pipeline}, result.Definition.Spec.Pipelines)
		})
	}
}

func buildClusterLogForwarderBuilder(valid bool) *ClusterLogForwarderBuilder {
	name := defaultClusterLogForwarderName
	if !valid {
		name = ""
	}

	return common.NewNamespacedBuilder[observabilityv1.ClusterLogForwarder, ClusterLogForwarderBuilder](
		clients.GetTestClients(clients.TestClientParams{}),
		observabilityv1.AddToScheme,
		name,
		defaultClusterLogForwarderNamespace,
	)
}
