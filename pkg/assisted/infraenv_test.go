package assisted

import (
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	infraEnvBuilderName      = "infraenv-builder-test"
	infraEnvBuilderNamespace = "infraenv-builder-test-namespace"
	infraEnvBuilderSecret    = "infraenv-pull-secret"
)

func TestNewInfraEnvBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[agentInstallV1Beta1.InfraEnv, InfraEnvBuilder](
		func(apiClient *clients.Settings, name, nsname string) *InfraEnvBuilder {
			return NewInfraEnvBuilder(apiClient, name, nsname, infraEnvBuilderSecret)
		},
		agentInstallV1Beta1.AddToScheme,
		infraEnvGVK,
	).ExecuteTests(t)
}

func TestNewInfraEnvBuilderPullSecret(t *testing.T) {
	t.Parallel()

	t.Run("sets pull secret reference", func(t *testing.T) {
		t.Parallel()

		builder := buildValidInfraEnvTestBuilder(getInfraEnvTestAPIClient())

		require.NoError(t, builder.GetError())
		require.NotNil(t, builder.Definition.Spec.PullSecretRef)
		assert.Equal(t, infraEnvBuilderSecret, builder.Definition.Spec.PullSecretRef.Name)
	})

	t.Run("empty pull secret returns error", func(t *testing.T) {
		t.Parallel()

		builder := NewInfraEnvBuilder(
			getInfraEnvTestAPIClient(), infraEnvBuilderName, infraEnvBuilderNamespace, "")

		require.Error(t, builder.GetError())
		assert.EqualError(t, builder.GetError(), "infraenv 'pull-secret' cannot be empty")
		assert.Nil(t, builder.Definition.Spec.PullSecretRef)
	})
}

func TestPullInfraEnvInstall(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[agentInstallV1Beta1.InfraEnv, InfraEnvBuilder](
		PullInfraEnvInstall,
		agentInstallV1Beta1.AddToScheme,
		infraEnvGVK,
	).ExecuteTests(t)
}

func TestInfraEnvBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newInfraEnvCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestInfraEnvWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newInfraEnvCommonTestConfig()).ExecuteTests(t)
}

func TestWithClusterRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		clusterName   string
		clusterNSName string
		invalid       bool
		expectedError string
	}{
		{
			name:          "valid cluster reference",
			clusterName:   "cluster",
			clusterNSName: "cluster-namespace",
		},
		{
			name:          "empty cluster name",
			clusterNSName: "cluster-namespace",
			expectedError: "infraenv clusterRef 'name' cannot be empty",
		},
		{
			name:          "empty cluster namespace",
			clusterName:   "cluster",
			expectedError: "infraenv clusterRef 'namespace' cannot be empty",
		},
		{
			name:          "invalid builder",
			clusterName:   "cluster",
			clusterNSName: "cluster-namespace",
			invalid:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *InfraEnvBuilder
			if testCase.invalid {
				builder = buildInvalidInfraEnvTestBuilder(getInfraEnvTestAPIClient())
			} else {
				builder = buildValidInfraEnvTestBuilder(getInfraEnvTestAPIClient())
			}

			builder = builder.WithClusterRef(testCase.clusterName, testCase.clusterNSName)

			if testCase.invalid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)
				assert.Nil(t, builder.Definition.Spec.ClusterRef)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.ClusterRef)
			assert.Equal(t, testCase.clusterName, builder.Definition.Spec.ClusterRef.Name)
			assert.Equal(t, testCase.clusterNSName, builder.Definition.Spec.ClusterRef.Namespace)
		})
	}
}

func TestWithAdditionalNTPSource(t *testing.T) {
	t.Parallel()

	assertInfraEnvModifier(t, "additional NTP sources", func(builder *InfraEnvBuilder) *InfraEnvBuilder {
		return builder.WithAdditionalNTPSource("ntp-1.example.com").WithAdditionalNTPSource("ntp-2.example.com")
	}, func(t *testing.T, builder *InfraEnvBuilder) {
		t.Helper()

		assert.Equal(t, []string{"ntp-1.example.com", "ntp-2.example.com"}, builder.Definition.Spec.AdditionalNTPSources)
	})
}

func TestWithSSHAuthorizedKey(t *testing.T) {
	t.Parallel()

	assertInfraEnvModifier(t, "SSH authorized key", func(builder *InfraEnvBuilder) *InfraEnvBuilder {
		return builder.WithSSHAuthorizedKey("ssh-rsa test")
	}, func(t *testing.T, builder *InfraEnvBuilder) {
		t.Helper()

		assert.Equal(t, "ssh-rsa test", builder.Definition.Spec.SSHAuthorizedKey)
	})
}

func TestWithAgentLabel(t *testing.T) {
	t.Parallel()

	assertInfraEnvModifier(t, "agent labels", func(builder *InfraEnvBuilder) *InfraEnvBuilder {
		return builder.WithAgentLabel("location", "east").WithAgentLabel("tier", "control-plane")
	}, func(t *testing.T, builder *InfraEnvBuilder) {
		t.Helper()

		assert.Equal(t, map[string]string{"location": "east", "tier": "control-plane"}, builder.Definition.Spec.AgentLabels)
	})
}

func TestWithProxy(t *testing.T) {
	t.Parallel()

	proxy := agentInstallV1Beta1.Proxy{HTTPProxy: "http://proxy.example.com", NoProxy: "localhost"}

	assertInfraEnvModifier(t, "proxy", func(builder *InfraEnvBuilder) *InfraEnvBuilder {
		return builder.WithProxy(proxy)
	}, func(t *testing.T, builder *InfraEnvBuilder) {
		t.Helper()

		assert.Equal(t, &proxy, builder.Definition.Spec.Proxy)
	})
}

func TestWithNmstateConfigLabelSelector(t *testing.T) {
	t.Parallel()

	selector := metav1.LabelSelector{MatchLabels: map[string]string{"network": "static"}}

	assertInfraEnvModifier(t, "NMStateConfig selector", func(builder *InfraEnvBuilder) *InfraEnvBuilder {
		return builder.WithNmstateConfigLabelSelector(selector)
	}, func(t *testing.T, builder *InfraEnvBuilder) {
		t.Helper()

		assert.Equal(t, selector, builder.Definition.Spec.NMStateConfigLabelSelector)
	})
}

func TestWithCPUType(t *testing.T) {
	t.Parallel()

	assertInfraEnvModifier(t, "CPU type", func(builder *InfraEnvBuilder) *InfraEnvBuilder {
		return builder.WithCPUType("arm64")
	}, func(t *testing.T, builder *InfraEnvBuilder) {
		t.Helper()

		assert.Equal(t, "arm64", builder.Definition.Spec.CpuArchitecture)
	})
}

func TestWithIgnitionConfigOverride(t *testing.T) {
	t.Parallel()

	assertInfraEnvModifier(t, "ignition config override", func(builder *InfraEnvBuilder) *InfraEnvBuilder {
		return builder.WithIgnitionConfigOverride(`{"ignition":{"version":"3.2.0"}}`)
	}, func(t *testing.T, builder *InfraEnvBuilder) {
		t.Helper()

		assert.JSONEq(t, `{"ignition":{"version":"3.2.0"}}`, builder.Definition.Spec.IgnitionConfigOverride)
	})
}

func TestWithIPXEScriptType(t *testing.T) {
	t.Parallel()

	assertInfraEnvModifier(t, "iPXE script type", func(builder *InfraEnvBuilder) *InfraEnvBuilder {
		return builder.WithIPXEScriptType(agentInstallV1Beta1.BootOrderControl)
	}, func(t *testing.T, builder *InfraEnvBuilder) {
		t.Helper()

		assert.Equal(t, agentInstallV1Beta1.BootOrderControl, builder.Definition.Spec.IPXEScriptType)
	})
}

func TestWithKernelArgument(t *testing.T) {
	t.Parallel()

	kernelArgs := []agentInstallV1Beta1.KernelArgument{
		{Operation: "append", Value: "console=ttyS0"},
		{Operation: "append", Value: "quiet"},
	}

	assertInfraEnvModifier(t, "kernel arguments", func(builder *InfraEnvBuilder) *InfraEnvBuilder {
		return builder.WithKernelArgument(kernelArgs[0]).WithKernelArgument(kernelArgs[1])
	}, func(t *testing.T, builder *InfraEnvBuilder) {
		t.Helper()

		assert.Equal(t, kernelArgs, builder.Definition.Spec.KernelArguments)
	})
}

func TestInfraEnvWaitForDiscoveryISOCreation(t *testing.T) {
	t.Parallel()

	t.Run("created discovery ISO", func(t *testing.T) {
		t.Parallel()

		infraEnv := buildDummyInfraEnv(infraEnvBuilderName, infraEnvBuilderNamespace)
		createdTime := metav1.Now()
		infraEnv.Status.CreatedTime = &createdTime
		settings := buildTestClientWithInfraEnvs(interceptor.Funcs{}, infraEnv)
		builder := buildValidInfraEnvTestBuilder(settings)

		result, err := builder.WaitForDiscoveryISOCreation(time.Second)

		require.NoError(t, err)
		assert.Same(t, builder, result)
		require.NotNil(t, result.Object)
		assert.NotNil(t, result.Object.Status.CreatedTime)
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidInfraEnvTestBuilder(getInfraEnvTestAPIClient())

		result, err := builder.WaitForDiscoveryISOCreation(time.Second)

		require.Error(t, err)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
		assert.Same(t, builder, result)
	})
}

func TestInfraEnvDeleteAndWait(t *testing.T) {
	t.Parallel()

	t.Run("existing InfraEnv", func(t *testing.T) {
		t.Parallel()

		settings := buildTestClientWithInfraEnvs(
			interceptor.Funcs{}, buildDummyInfraEnv(infraEnvBuilderName, infraEnvBuilderNamespace))
		builder := buildValidInfraEnvTestBuilder(settings)

		err := builder.DeleteAndWait(2 * time.Second)

		require.NoError(t, err)
		assert.Nil(t, builder.Object)
	})

	t.Run("missing InfraEnv", func(t *testing.T) {
		t.Parallel()

		builder := buildValidInfraEnvTestBuilder(getInfraEnvTestAPIClient())

		err := builder.DeleteAndWait(time.Second)

		require.NoError(t, err)
		assert.Nil(t, builder.Object)
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidInfraEnvTestBuilder(getInfraEnvTestAPIClient())

		err := builder.DeleteAndWait(time.Second)

		require.Error(t, err)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
	})
}

func TestListInfraEnvsInAllNamespacesCommon(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig[runtimeclient.ListOption, agentInstallV1Beta1.InfraEnv, InfraEnvBuilder](
		ListInfraEnvsInAllNamespaces,
		agentInstallV1Beta1.AddToScheme,
		infraEnvGVK,
	).ExecuteTests(t)
}

func newInfraEnvCommonTestConfig() testhelper.CommonTestConfig[
	agentInstallV1Beta1.InfraEnv, InfraEnvBuilder, *agentInstallV1Beta1.InfraEnv, *InfraEnvBuilder] {
	return testhelper.NewCommonTestConfig[agentInstallV1Beta1.InfraEnv, InfraEnvBuilder](
		agentInstallV1Beta1.AddToScheme,
		infraEnvGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

func buildValidInfraEnvTestBuilder(apiClient *clients.Settings) *InfraEnvBuilder {
	return NewInfraEnvBuilder(apiClient, infraEnvBuilderName, infraEnvBuilderNamespace, infraEnvBuilderSecret)
}

func buildInvalidInfraEnvTestBuilder(apiClient *clients.Settings) *InfraEnvBuilder {
	return NewInfraEnvBuilder(apiClient, "", infraEnvBuilderNamespace, infraEnvBuilderSecret)
}

func getInfraEnvTestAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{agentInstallV1Beta1.AddToScheme},
	})
}

func assertInfraEnvModifier(
	t *testing.T,
	name string,
	modifier func(*InfraEnvBuilder) *InfraEnvBuilder,
	assertApplied func(*testing.T, *InfraEnvBuilder),
) {
	t.Helper()

	t.Run(name+" applies to valid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildValidInfraEnvTestBuilder(getInfraEnvTestAPIClient())
		result := modifier(builder)

		require.NoError(t, result.GetError())
		assertApplied(t, result)
	})

	t.Run(name+" preserves invalid builder error", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidInfraEnvTestBuilder(getInfraEnvTestAPIClient())
		originalError := builder.GetError()
		originalSpec := builder.Definition.DeepCopy().Spec

		require.Error(t, originalError)

		result := modifier(builder)

		require.Error(t, result.GetError())
		assert.Equal(t, originalError, result.GetError())
		assert.Equal(t, originalSpec, result.Definition.Spec)
	})
}
