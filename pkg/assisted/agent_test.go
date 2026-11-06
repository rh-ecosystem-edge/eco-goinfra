package assisted

import (
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPullAgent(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullAgent, agentInstallV1Beta1.AddToScheme, agentGVK).ExecuteTests(t)
}

func TestAgentBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[agentInstallV1Beta1.Agent, agentBuilder](
		agentInstallV1Beta1.AddToScheme,
		agentGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		With(testhelper.NewWithOptionsTestConfig(commonTestConfig)).
		Run(t)
}

func TestAgentModifiers(t *testing.T) {
	t.Parallel()

	for _, modifier := range agentModifierTestCases() {
		modifier := modifier

		t.Run(modifier.name+"/success", func(t *testing.T) {
			t.Parallel()

			builder := modifier.apply(buildValidAgentTestBuilder(buildTestClientWithDummyAgent()))

			require.NoError(t, builder.GetError())
			modifier.verify(t, builder)
		})

		t.Run(modifier.name+"/invalid builder", func(t *testing.T) {
			t.Parallel()

			builder := modifier.apply(buildInvalidAgentTestBuilder())

			require.Error(t, builder.GetError())
			assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
		})

		if modifier.checksExists {
			t.Run(modifier.name+"/non-existent agent", func(t *testing.T) {
				t.Parallel()

				builder := modifier.apply(buildMissingAgentTestBuilder())

				assert.EqualError(t, builder.GetError(), nonExistentMsg)
			})
		}
	}
}

type agentModifierTestCase struct {
	name         string
	checksExists bool
	apply        func(*agentBuilder) *agentBuilder
	verify       func(*testing.T, *agentBuilder)
}

func agentModifierTestCases() []agentModifierTestCase {
	return []agentModifierTestCase{
		{
			name:         "hostname",
			checksExists: true,
			apply: func(builder *agentBuilder) *agentBuilder {
				return builder.WithHostName("worker-0")
			},
			verify: func(t *testing.T, builder *agentBuilder) {
				t.Helper()
				assert.Equal(t, "worker-0", builder.Definition.Spec.Hostname)
			},
		},
		{
			name:         "role",
			checksExists: true,
			apply: func(builder *agentBuilder) *agentBuilder {
				return builder.WithRole("worker")
			},
			verify: func(t *testing.T, builder *agentBuilder) {
				t.Helper()
				assert.Equal(t, models.HostRole("worker"), builder.Definition.Spec.Role)
			},
		},
		{
			name: "installation disk",
			apply: func(builder *agentBuilder) *agentBuilder {
				return builder.WithInstallationDisk("disk-1")
			},
			verify: func(t *testing.T, builder *agentBuilder) {
				t.Helper()
				assert.Equal(t, "disk-1", builder.Definition.Spec.InstallationDiskID)
			},
		},
		{
			name: "ignition config override",
			apply: func(builder *agentBuilder) *agentBuilder {
				return builder.WithIgnitionConfigOverride("override")
			},
			verify: func(t *testing.T, builder *agentBuilder) {
				t.Helper()
				assert.Equal(t, "override", builder.Definition.Spec.IgnitionConfigOverrides)
			},
		},
		{
			name: "approval",
			apply: func(builder *agentBuilder) *agentBuilder {
				return builder.WithApproval(true)
			},
			verify: func(t *testing.T, builder *agentBuilder) {
				t.Helper()
				assert.True(t, builder.Definition.Spec.Approved)
			},
		},
	}
}

func TestAgentWaitForState(t *testing.T) {
	t.Parallel()

	agent := buildDummyAgent(testAgentName)
	agent.Status.DebugInfo.State = "installed"

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{agent},
		SchemeAttachers: agentTestSchemes,
	})

	builder := newAgentBuilder(testSettings.Client, buildDummyAgent(testAgentName))

	result, err := builder.WaitForState("installed", time.Second)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "installed", result.Object.Status.DebugInfo.State)

	_, err = buildInvalidAgentTestBuilder().WaitForState("installed", time.Second)

	require.Error(t, err)
}

func TestAgentWaitForStateInfo(t *testing.T) {
	t.Parallel()

	agent := buildDummyAgent(testAgentName)
	agent.Status.DebugInfo.StateInfo = "done"

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{agent},
		SchemeAttachers: agentTestSchemes,
	})

	builder := newAgentBuilder(testSettings.Client, buildDummyAgent(testAgentName))

	result, err := builder.WaitForStateInfo("done", time.Second)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "done", result.Object.Status.DebugInfo.StateInfo)

	_, err = buildInvalidAgentTestBuilder().WaitForStateInfo("done", time.Second)

	require.Error(t, err)
}

func buildInvalidAgentTestBuilder() *agentBuilder {
	return newAgentBuilder(buildTestClientWithDummyAgent().Client, buildDummyAgent(""))
}

func buildMissingAgentTestBuilder() *agentBuilder {
	return newAgentBuilder(buildTestClientWithDummyAgent().Client, buildDummyAgent("missing-agent"))
}
