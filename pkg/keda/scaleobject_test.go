package keda

import (
	"fmt"
	"testing"

	kedav2v1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultScaledObjectName      = "prometheus-scaledobject"
	defaultScaledObjectNamespace = "test-appspace"
	pollingInterval              = int32(5)
	cooldownPeriod               = int32(10)
	minReplicaCount              = int32(1)
	maxReplicaCount              = int32(8)
	zeroValue                    = int32(0)
	kedav2v1alpha1TestSchemes    = []clients.SchemeAttacher{
		kedav2v1alpha1.AddToScheme,
	}
)

func TestNewScaledObjectBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewScaledObjectBuilder, kedav2v1alpha1.AddToScheme, scaledObjectGVK).ExecuteTests(t)
}

func TestPullScaledObject(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullScaledObject, kedav2v1alpha1.AddToScheme, scaledObjectGVK).ExecuteTests(t)
}

func TestScaledObjectBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[kedav2v1alpha1.ScaledObject, ScaledObjectBuilder](
		kedav2v1alpha1.AddToScheme,
		scaledObjectGVK,
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

func TestScaledObjectWithScaleTargetRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		scaleTargetRef kedav2v1alpha1.ScaleTarget
		valid          bool
	}{
		{
			name: "valid scale target ref",
			scaleTargetRef: kedav2v1alpha1.ScaleTarget{
				Name: "test-app",
			},
			valid: true,
		},
		{
			name:           "empty scale target ref",
			scaleTargetRef: kedav2v1alpha1.ScaleTarget{},
			valid:          true,
		},
		{
			name: "invalid builder",
			scaleTargetRef: kedav2v1alpha1.ScaleTarget{
				Name: "test-app",
			},
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidScaledObjectBuilder()
			if !testCase.valid {
				builder = buildInvalidScaledObjectBuilder()
			}

			builder = builder.WithScaleTargetRef(testCase.scaleTargetRef)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assert.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.ScaleTargetRef)
			assert.Equal(t, testCase.scaleTargetRef, *builder.Definition.Spec.ScaleTargetRef)
		})
	}
}

func TestScaledObjectWithPollingInterval(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		pollingInterval int32
		valid           bool
	}{
		{
			name:            "valid polling interval",
			pollingInterval: pollingInterval,
			valid:           true,
		},
		{
			name:            "zero polling interval",
			pollingInterval: zeroValue,
			valid:           true,
		},
		{
			name:            "invalid builder",
			pollingInterval: pollingInterval,
			valid:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidScaledObjectBuilder()
			if !testCase.valid {
				builder = buildInvalidScaledObjectBuilder()
			}

			builder = builder.WithPollingInterval(testCase.pollingInterval)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assert.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.PollingInterval)
			assert.Equal(t, testCase.pollingInterval, *builder.Definition.Spec.PollingInterval)
		})
	}
}

func TestScaledObjectWithCooldownPeriod(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		cooldownPeriod int32
		valid          bool
	}{
		{
			name:           "valid cooldown period",
			cooldownPeriod: cooldownPeriod,
			valid:          true,
		},
		{
			name:           "zero cooldown period",
			cooldownPeriod: zeroValue,
			valid:          true,
		},
		{
			name:           "invalid builder",
			cooldownPeriod: cooldownPeriod,
			valid:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidScaledObjectBuilder()
			if !testCase.valid {
				builder = buildInvalidScaledObjectBuilder()
			}

			builder = builder.WithCooldownPeriod(testCase.cooldownPeriod)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assert.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.CooldownPeriod)
			assert.Equal(t, testCase.cooldownPeriod, *builder.Definition.Spec.CooldownPeriod)
		})
	}
}

func TestScaledObjectWithMinReplicaCount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		minReplicaCount int32
		valid           bool
	}{
		{
			name:            "valid min replica count",
			minReplicaCount: minReplicaCount,
			valid:           true,
		},
		{
			name:            "zero min replica count",
			minReplicaCount: zeroValue,
			valid:           true,
		},
		{
			name:            "invalid builder",
			minReplicaCount: minReplicaCount,
			valid:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidScaledObjectBuilder()
			if !testCase.valid {
				builder = buildInvalidScaledObjectBuilder()
			}

			builder = builder.WithMinReplicaCount(testCase.minReplicaCount)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assert.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.MinReplicaCount)
			assert.Equal(t, testCase.minReplicaCount, *builder.Definition.Spec.MinReplicaCount)
		})
	}
}

func TestScaledObjectWithMaxReplicaCount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		maxReplicaCount int32
		valid           bool
	}{
		{
			name:            "valid max replica count",
			maxReplicaCount: maxReplicaCount,
			valid:           true,
		},
		{
			name:            "zero max replica count",
			maxReplicaCount: zeroValue,
			valid:           true,
		},
		{
			name:            "invalid builder",
			maxReplicaCount: maxReplicaCount,
			valid:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidScaledObjectBuilder()
			if !testCase.valid {
				builder = buildInvalidScaledObjectBuilder()
			}

			builder = builder.WithMaxReplicaCount(testCase.maxReplicaCount)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assert.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.MaxReplicaCount)
			assert.Equal(t, testCase.maxReplicaCount, *builder.Definition.Spec.MaxReplicaCount)
		})
	}
}

func TestScaledObjectWithTriggers(t *testing.T) {
	t.Parallel()

	validTriggers := []kedav2v1alpha1.ScaleTriggers{{
		Type: "prometheus",
		Metadata: map[string]string{
			"serverAddress": "https://thanos-querier.openshift-monitoring.svc.cluster.local:9092",
			"namespace":     defaultScaledObjectNamespace,
			"metricName":    "http_requests_total",
			"threshold":     "5",
			"query":         "sum(rate(http_requests_total{job=\"test-app\"}[1m]))",
			"authModes":     "bearer",
		},
		AuthenticationRef: &kedav2v1alpha1.AuthenticationRef{
			Name: defaultScaledObjectName,
		},
	}}

	testCases := []struct {
		name          string
		triggers      []kedav2v1alpha1.ScaleTriggers
		valid         bool
		expectedError error
	}{
		{
			name:     "valid triggers",
			triggers: validTriggers,
			valid:    true,
		},
		{
			name: "trigger without metadata",
			triggers: []kedav2v1alpha1.ScaleTriggers{{
				Type: "prometheus",
			}},
			valid: true,
		},
		{
			name:          "empty triggers",
			triggers:      []kedav2v1alpha1.ScaleTriggers{},
			valid:         true,
			expectedError: fmt.Errorf("'triggers' argument cannot be empty"),
		},
		{
			name:     "invalid builder",
			triggers: validTriggers,
			valid:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidScaledObjectBuilder()
			if !testCase.valid {
				builder = buildInvalidScaledObjectBuilder()
			}

			builder = builder.WithTriggers(testCase.triggers)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			assert.Equal(t, testCase.expectedError, builder.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.triggers, builder.Definition.Spec.Triggers)
			}
		})
	}
}

func buildValidScaledObjectBuilder() *ScaledObjectBuilder {
	return NewScaledObjectBuilder(
		buildScaledObjectTestClient(), defaultScaledObjectName, defaultScaledObjectNamespace)
}

func buildInvalidScaledObjectBuilder() *ScaledObjectBuilder {
	return NewScaledObjectBuilder(
		buildScaledObjectTestClient(), "", defaultScaledObjectNamespace)
}

func buildScaledObjectTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: kedav2v1alpha1TestSchemes,
	})
}
