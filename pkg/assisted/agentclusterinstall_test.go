package assisted

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	hiveextV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/hiveextension/v1beta1"
	hivev1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/hive/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	aciTestName              = "aci-test-name"
	aciTestNamespace         = "aci-test-namespace"
	aciTestClusterDeployment = "aci-test-clusterdeployment"
)

func TestNewAgentClusterInstallBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *AgentClusterInstallBuilder {
				return NewAgentClusterInstallBuilder(
					apiClient, name, nsname, aciTestClusterDeployment, 3, 2, dummyTestNetwork())
			},
			hiveextV1Beta1.AddToScheme,
			agentClusterInstallGVK,
		).ExecuteTests(t)
	})

	t.Run("constructor populates spec", func(t *testing.T) {
		t.Parallel()

		network := dummyTestNetwork()
		builder := NewAgentClusterInstallBuilder(
			newAgentClusterInstallTestClient(nil), aciTestName, aciTestNamespace,
			aciTestClusterDeployment, 3, 2, network)

		require.NotNil(t, builder)
		require.NoError(t, builder.GetError())
		assert.Equal(t, aciTestName, builder.Definition.Name)
		assert.Equal(t, aciTestNamespace, builder.Definition.Namespace)
		assert.Equal(t, aciTestClusterDeployment, builder.Definition.Spec.ClusterDeploymentRef.Name)
		assert.Equal(t, 3, builder.Definition.Spec.ProvisionRequirements.ControlPlaneAgents)
		assert.Equal(t, 2, builder.Definition.Spec.ProvisionRequirements.WorkerAgents)
		assert.Equal(t, network, builder.Definition.Spec.Networking)
	})

	t.Run("cluster deployment is required", func(t *testing.T) {
		t.Parallel()

		builder := NewAgentClusterInstallBuilder(
			newAgentClusterInstallTestClient(nil), aciTestName, aciTestNamespace, "", 3, 2, dummyTestNetwork())

		require.NotNil(t, builder)
		assert.EqualError(t, builder.GetError(), "agentclusterinstall 'clusterDeployment' cannot be empty")
	})
}

func TestPullAgentClusterInstall(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullAgentClusterInstall, hiveextV1Beta1.AddToScheme, agentClusterInstallGVK).ExecuteTests(t)
}

func TestAgentClusterInstallBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[
		hiveextV1Beta1.AgentClusterInstall, AgentClusterInstallBuilder](
		hiveextV1Beta1.AddToScheme,
		agentClusterInstallGVK,
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

func TestAgentClusterInstallWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[hiveextV1Beta1.AgentClusterInstall, AgentClusterInstallBuilder](
			hiveextV1Beta1.AddToScheme,
			agentClusterInstallGVK,
			testhelper.ResourceScopeNamespaced,
		)).ExecuteTests(t)
}

func TestAgentClusterInstallWithOptionsMutations(t *testing.T) {
	t.Parallel()

	t.Run("applies package-specific options", func(t *testing.T) {
		t.Parallel()

		builder := generateAgentClusterInstallTestBuilder().WithOptions(
			func(builder *AgentClusterInstallBuilder) (*AgentClusterInstallBuilder, error) {
				builder.Definition.Spec.HoldInstallation = true

				return builder, nil
			},
			func(builder *AgentClusterInstallBuilder) (*AgentClusterInstallBuilder, error) {
				builder.Definition.Spec.IgnitionEndpoint = &hiveextV1Beta1.IgnitionEndpoint{
					Url: "https://some.ign.endpoint.com",
				}

				return builder, nil
			},
		)

		require.NoError(t, builder.GetError())
		assert.True(t, builder.Definition.Spec.HoldInstallation)
		require.NotNil(t, builder.Definition.Spec.IgnitionEndpoint)
		assert.Equal(t, "https://some.ign.endpoint.com", builder.Definition.Spec.IgnitionEndpoint.Url)
	})
}

func TestAgentClusterInstallDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		exists bool
		valid  bool
	}{
		{
			name:   "deletes existing resource",
			exists: true,
			valid:  true,
		},
		{
			name:   "resource already absent",
			exists: false,
			valid:  true,
		},
		{
			name:   "invalid builder",
			exists: true,
			valid:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.exists {
				objects = append(objects, generateAgentClusterInstall())
			}

			var builder *AgentClusterInstallBuilder
			if testCase.valid {
				builder = buildTestBuilderWithFakeObjects(objects)
			} else {
				builder = buildInvalidAgentClusterInstallTestBuilder(objects)
			}

			err := builder.DeleteAndWait(time.Second)

			if testCase.valid {
				require.NoError(t, err)
				assert.Nil(t, builder.Object)
			} else {
				assert.True(t, commonerrors.IsBuilderNameEmpty(err))
			}
		})
	}
}

func TestAgentClusterInstallWithAPIVip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		ip            string
		expectedError string
	}{
		{name: "IPv4", ip: "192.168.1.5"},
		{name: "IPv6", ip: "fd2e::5"},
		{name: "invalid IP", ip: "notanip", expectedError: errIncorrectAPIVIPFormat},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithAPIVip(testCase.ip)

			assertBuilderError(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.ip, builder.Definition.Spec.APIVIP)
			}
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithAPIVip("192.168.1.5")
		})
	})
}

func TestAgentClusterInstallWithAdditionalAPIVip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		ip            string
		expectedError string
	}{
		{name: "IPv4", ip: "192.168.1.5"},
		{name: "IPv6", ip: "fd2e::5"},
		{name: "invalid IP", ip: "notanip", expectedError: errIncorrectAPIVIPFormat},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithAdditionalAPIVip(testCase.ip)

			assertBuilderError(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				assert.Contains(t, builder.Definition.Spec.APIVIPs, testCase.ip)
			}
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithAdditionalAPIVip("192.168.1.5")
		})
	})
}

func TestAgentClusterInstallWithIngressVip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		ip            string
		expectedError string
	}{
		{name: "IPv4", ip: "192.168.1.10"},
		{name: "IPv6", ip: "fd2e::10"},
		{name: "invalid IP", ip: "notanip", expectedError: errIncorrectIngressVIPFormat},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithIngressVip(testCase.ip)

			assertBuilderError(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.ip, builder.Definition.Spec.IngressVIP)
			}
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithIngressVip("192.168.1.10")
		})
	})
}

func TestAgentClusterInstallWithAdditionalIngressVip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		ip            string
		expectedError string
	}{
		{name: "IPv4", ip: "192.168.1.10"},
		{name: "IPv6", ip: "fd2e::10"},
		{name: "invalid IP", ip: "notanip", expectedError: errIncorrectIngressVIPFormat},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithAdditionalIngressVip(testCase.ip)

			assertBuilderError(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				assert.Contains(t, builder.Definition.Spec.IngressVIPs, testCase.ip)
			}
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithAdditionalIngressVip("192.168.1.10")
		})
	})
}

func TestAgentClusterInstallWithUserManagedNetworking(t *testing.T) {
	t.Parallel()

	for _, enabled := range []bool{true, false} {
		t.Run(fmt.Sprintf("enabled %t", enabled), func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithUserManagedNetworking(enabled)

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.Networking.UserManagedNetworking)
			assert.Equal(t, enabled, *builder.Definition.Spec.Networking.UserManagedNetworking)
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithUserManagedNetworking(true)
		})
	})
}

func TestAgentClusterInstallWithPlatformType(t *testing.T) {
	t.Parallel()

	for _, platform := range []hiveextV1Beta1.PlatformType{"BareMetal", "None", "VSphere", "Nutanix"} {
		t.Run(string(platform), func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithPlatformType(platform)

			require.NoError(t, builder.GetError())
			assert.Equal(t, platform, builder.Definition.Spec.PlatformType)
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithPlatformType(hiveextV1Beta1.BareMetalPlatformType)
		})
	})
}

func TestAgentClusterInstallWithControlPlaneAgents(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		count         int
		expectedError string
	}{
		{name: "three agents", count: 3},
		{name: "one agent", count: 1},
		{name: "zero agents", count: 0, expectedError: errInvalidControlplaneAgentCount},
		{name: "negative agents", count: -1, expectedError: errInvalidControlplaneAgentCount},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithControlPlaneAgents(testCase.count)

			assertBuilderError(t, builder.GetError(), testCase.expectedError)
			assert.Equal(t, testCase.count, builder.Definition.Spec.ProvisionRequirements.ControlPlaneAgents)
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithControlPlaneAgents(3)
		})
	})
}

func TestAgentClusterInstallWithImageSet(t *testing.T) {
	t.Parallel()

	builder := generateAgentClusterInstallTestBuilder().WithImageSet("4.16")

	require.NoError(t, builder.GetError())
	require.NotNil(t, builder.Definition.Spec.ImageSetRef)
	assert.Equal(t, "4.16", builder.Definition.Spec.ImageSetRef.Name)

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithImageSet("4.16")
		})
	})
}

func TestAgentClusterInstallWithWorkerAgents(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		count         int
		expectedError string
	}{
		{name: "three workers", count: 3},
		{name: "one worker", count: 1},
		{name: "zero workers", count: 0},
		{
			name:          "negative workers",
			count:         -1,
			expectedError: "agentclusterinstall worker agents cannot be less that 0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithWorkerAgents(testCase.count)

			assertBuilderError(t, builder.GetError(), testCase.expectedError)
			assert.Equal(t, testCase.count, builder.Definition.Spec.ProvisionRequirements.WorkerAgents)
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithWorkerAgents(3)
		})
	})
}

//nolint:lll
func TestAgentClusterInstallWithSSHPublicKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		sshKey string
	}{
		{
			name:   "non-empty key",
			sshKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCaX4UsUhbLbeTGx6wob/6Y2qQzBDi2VhjFaFGugNk9zuNgPB4Um4VNa8TgRr/nrJxLWduuM0gbBgAA9719cCV6FsTq6e0TTrtFAy498vNrE1gyCp8gX2/+60NWwjeK3iglVym0WxmXhWpZCPk6EBsBQgclt6iQhVH20f60lWAD6+8I5iXCm5u63HNEzuTmJUVaVCEU62zamwwJUR+6yddnu+rIrLqw7xYuuATKMBuL+AUnB3zR08qv65Tyhf/YpyjlW/w7kCFwTsE0Xrbtumd1irF49DiY4Xb5WGPRTjrhVhzpAvtvEhfrMGGnbwnQ0P1pUTAQz8eYKSN4M0oxaGPObtjkEM4EAjBo5vyzN1xi9H/tekOQCmVDY2XOCVvIuxUPgq/8Uc3oGkoh3Eay+KSRIhQC3+kqXX889H//SmENoejIAxnETXPnVygKMaoJbh2d44RjIC5LkJsNSwcK24+hv4VU0uJjwXMPPE5YTUxcVPukATd4i22FOSiIL/ULuos=",
		},
		{name: "empty key", sshKey: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithSSHPublicKey(testCase.sshKey)

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.sshKey, builder.Definition.Spec.SSHPublicKey)
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithSSHPublicKey("key")
		})
	})
}

func TestAgentClusterInstallWithNetworkType(t *testing.T) {
	t.Parallel()

	for _, networkType := range []string{"OpenShiftSDN", "OVNKubernetes"} {
		t.Run(networkType, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithNetworkType(networkType)

			require.NoError(t, builder.GetError())
			assert.Equal(t, networkType, builder.Definition.Spec.Networking.NetworkType)
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithNetworkType("OVNKubernetes")
		})
	})
}

func TestAgentClusterInstallWithAdditionalClusterNetwork(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		cidr          string
		prefix        int32
		expectedError string
	}{
		{name: "valid network", cidr: "10.128.0.0/14", prefix: 23},
		{
			name:          "invalid CIDR",
			cidr:          "10.128.0.0",
			prefix:        23,
			expectedError: "agentclusterinstall contains invalid clusterNetwork cidr",
		},
		{name: "zero prefix", cidr: "10.128.0.0/14", prefix: 0, expectedError: errInvalidClusterNetworkPrefix},
		{name: "negative prefix", cidr: "10.128.0.0/14", prefix: -1, expectedError: errInvalidClusterNetworkPrefix},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().
				WithAdditionalClusterNetwork(testCase.cidr, testCase.prefix)

			assertBuilderError(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				require.Len(t, builder.Definition.Spec.Networking.ClusterNetwork, 2)
				assert.Equal(t, testCase.cidr, builder.Definition.Spec.Networking.ClusterNetwork[1].CIDR)
				assert.Equal(t, testCase.prefix, builder.Definition.Spec.Networking.ClusterNetwork[1].HostPrefix)
			}
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithAdditionalClusterNetwork("10.128.0.0/14", 23)
		})
	})
}

func TestAgentClusterInstallWithAdditionalServiceNetwork(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		cidr          string
		expectedError string
	}{
		{name: "valid network", cidr: "172.30.0.0/16"},
		{name: "empty CIDR", cidr: "", expectedError: errInvalidServiceNetworkCIDR},
		{name: "invalid CIDR", cidr: "172.30.0.0/162", expectedError: errInvalidServiceNetworkCIDR},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := generateAgentClusterInstallTestBuilder().WithAdditionalServiceNetwork(testCase.cidr)

			assertBuilderError(t, builder.GetError(), testCase.expectedError)

			if testCase.expectedError == "" {
				require.Len(t, builder.Definition.Spec.Networking.ServiceNetwork, 2)
				assert.Equal(t, testCase.cidr, builder.Definition.Spec.Networking.ServiceNetwork[1])
			}
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		assertInvalidModifierShortCircuit(t, func(builder *AgentClusterInstallBuilder) *AgentClusterInstallBuilder {
			return builder.WithAdditionalServiceNetwork("172.30.0.0/16")
		})
	})
}

func TestAgentClusterInstallWaitForState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		state string
	}{
		{name: "adding hosts", state: "adding-hosts"},
		{name: "installing", state: "installing"},
		{name: "finalizing", state: "finalizing"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			aci := generateAgentClusterInstall()
			aci.Status.DebugInfo.State = testCase.state
			builder := buildTestBuilderWithFakeObjects([]runtime.Object{aci})

			result, err := builder.WaitForState(testCase.state, 2*time.Second)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, testCase.state, result.Object.Status.DebugInfo.State)
		})
	}
}

func TestAgentClusterInstallWaitForStateInfo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		info string
	}{
		{name: "installed", info: "Cluster is installed"},
		{name: "finalizing", info: "Cluster is finalizing"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			aci := generateAgentClusterInstall()
			aci.Status.DebugInfo.StateInfo = testCase.info
			builder := buildTestBuilderWithFakeObjects([]runtime.Object{aci})

			result, err := builder.WaitForStateInfo(testCase.info, 2*time.Second)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, testCase.info, result.Object.Status.DebugInfo.StateInfo)
		})
	}
}

func TestAgentClusterInstallWaitForConditionMessage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		conditionType hivev1.ClusterInstallConditionType
		message       string
	}{
		{
			name:          "stopped",
			conditionType: "Stopped",
			message:       "The installation has stopped because it completed successfully",
		},
		{name: "spec synced", conditionType: "SpecSynced", message: "SyncOK"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			aci := generateAgentClusterInstall()
			aci.Status.Conditions = []hivev1.ClusterInstallCondition{{
				Type:    testCase.conditionType,
				Status:  corev1.ConditionTrue,
				Reason:  "Reason",
				Message: testCase.message,
			}}
			builder := buildTestBuilderWithFakeObjects([]runtime.Object{aci})

			assert.NoError(t, builder.WaitForConditionMessage(testCase.conditionType, testCase.message, 5*time.Second))
		})
	}
}

func TestAgentClusterInstallWaitForConditionStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		conditionType   hivev1.ClusterInstallConditionType
		conditionStatus corev1.ConditionStatus
	}{
		{name: "true condition", conditionType: "Stopped", conditionStatus: corev1.ConditionTrue},
		{name: "false condition", conditionType: "Failed", conditionStatus: corev1.ConditionFalse},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			aci := generateAgentClusterInstall()
			aci.Status.Conditions = []hivev1.ClusterInstallCondition{{
				Type:    testCase.conditionType,
				Status:  testCase.conditionStatus,
				Reason:  "Reason",
				Message: "Message",
			}}
			builder := buildTestBuilderWithFakeObjects([]runtime.Object{aci})

			assert.NoError(t, builder.WaitForConditionStatus(testCase.conditionType, testCase.conditionStatus, 2*time.Second))
		})
	}
}

func TestAgentClusterInstallWaitForConditionReason(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		conditionType hivev1.ClusterInstallConditionType
		reason        string
	}{
		{name: "installation completed", conditionType: "Stopped", reason: "InstallationCompleted"},
		{name: "installation not failed", conditionType: "Failed", reason: "InstallationNotFailed"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			aci := generateAgentClusterInstall()
			aci.Status.Conditions = []hivev1.ClusterInstallCondition{{
				Type:    testCase.conditionType,
				Status:  corev1.ConditionTrue,
				Reason:  testCase.reason,
				Message: "Message",
			}}
			builder := buildTestBuilderWithFakeObjects([]runtime.Object{aci})

			assert.NoError(t, builder.WaitForConditionReason(testCase.conditionType, testCase.reason, 2*time.Second))
		})
	}
}

func TestAgentClusterInstallGetEvents(t *testing.T) {
	t.Parallel()

	path, err := os.Getwd()
	require.NoError(t, err)

	transport := &http.Transport{}
	transport.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	eventsTransport = transport

	aci := generateAgentClusterInstall()
	aci.Status.DebugInfo.EventsURL = fmt.Sprintf("file://%s/testdata/aci_events.json", path)
	builder := buildTestBuilderWithFakeObjects([]runtime.Object{aci})

	events, err := builder.GetEvents(true)

	require.NoError(t, err)
	assert.Len(t, events, 63)
}

func assertInvalidModifierShortCircuit(
	t *testing.T,
	modifier func(*AgentClusterInstallBuilder) *AgentClusterInstallBuilder,
) {
	t.Helper()

	builder := buildInvalidAgentClusterInstallTestBuilder(nil)
	result := modifier(builder)

	require.Same(t, builder, result)
	assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
}

func assertBuilderError(t *testing.T, err error, expected string) {
	t.Helper()

	if expected == "" {
		assert.NoError(t, err)

		return
	}

	assert.EqualError(t, err, expected)
}

func dummyTestNetwork() hiveextV1Beta1.Networking {
	return hiveextV1Beta1.Networking{
		ClusterNetwork: []hiveextV1Beta1.ClusterNetworkEntry{{
			CIDR:       "10.128.0.0/14",
			HostPrefix: 23,
		}},
		ServiceNetwork: []string{"172.30.0.0/16"},
	}
}

func newAgentClusterInstallTestClient(objects []runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{hiveextV1Beta1.AddToScheme},
	})
}

func buildTestBuilderWithFakeObjects(objects []runtime.Object) *AgentClusterInstallBuilder {
	return NewAgentClusterInstallBuilder(
		newAgentClusterInstallTestClient(objects),
		aciTestName,
		aciTestNamespace,
		aciTestClusterDeployment,
		3,
		2,
		dummyTestNetwork())
}

func buildInvalidAgentClusterInstallTestBuilder(objects []runtime.Object) *AgentClusterInstallBuilder {
	return NewAgentClusterInstallBuilder(
		newAgentClusterInstallTestClient(objects),
		"",
		aciTestNamespace,
		aciTestClusterDeployment,
		3,
		2,
		dummyTestNetwork())
}

func generateAgentClusterInstallTestBuilder() *AgentClusterInstallBuilder {
	return buildTestBuilderWithFakeObjects(nil)
}

func generateAgentClusterInstall() *hiveextV1Beta1.AgentClusterInstall {
	return &hiveextV1Beta1.AgentClusterInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      aciTestName,
			Namespace: aciTestNamespace,
		},
		Spec: hiveextV1Beta1.AgentClusterInstallSpec{
			ClusterDeploymentRef: corev1.LocalObjectReference{Name: aciTestName},
			Networking:           dummyTestNetwork(),
			ProvisionRequirements: hiveextV1Beta1.ProvisionRequirements{
				ControlPlaneAgents: 3,
				WorkerAgents:       2,
			},
		},
	}
}
