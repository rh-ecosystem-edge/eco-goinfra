package lca

import (
	"context"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	lcaipcv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ipchange/api/ipconfig/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	lcaipcv1TestSchemes = []clients.SchemeAttacher{
		lcaipcv1.AddToScheme,
	}
)

func TestIPConfigWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newIPConfigCommonTestConfig()).ExecuteTests(t)
}

func TestIPConfigPull(t *testing.T) {
	t.Parallel()

	testhelper.NewSingletonClusterScopedPullTestConfig(
		PullIPConfig,
		lcaipcv1.AddToScheme,
		ipConfigGVK,
		ipConfigName,
	).ExecuteTests(t)
}

func TestIPConfigDelete(t *testing.T) {
	t.Parallel()

	testhelper.NewDeleteReturnerTestConfig(newIPConfigCommonTestConfig()).ExecuteTests(t)
}

func TestIPConfigUpdate(t *testing.T) {
	t.Parallel()

	testhelper.NewUpdateTestConfig(newIPConfigCommonTestConfig()).ExecuteTests(t)
}

func TestIPConfigGet(t *testing.T) {
	t.Parallel()

	testhelper.NewGetTestConfig(newIPConfigCommonTestConfig()).ExecuteTests(t)
}

func TestIPConfigExists(t *testing.T) {
	t.Parallel()

	testhelper.NewExistsTestConfig(newIPConfigCommonTestConfig()).ExecuteTests(t)
}

func TestIPConfigWithIPv4Address(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *IPConfigBuilder
		ipv4Address    string
		expectedError  string
		invalidBuilder bool
	}{
		{
			name:        "sets valid IPv4 address",
			builder:     buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv4Address: "192.168.1.10",
		},
		{
			name:          "rejects invalid IPv4 address",
			builder:       buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv4Address:   "invalid",
			expectedError: "invalid IPv4 address argument invalid",
		},
		{
			name:           "invalid builder short circuits",
			builder:        buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			ipv4Address:    "192.168.1.10",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithIPv4Address(testCase.ipv4Address)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.ipv4Address, result.Definition.Spec.IPv4.Address)
		})
	}
}

func TestIPConfigWithIPv6Address(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *IPConfigBuilder
		ipv6Address    string
		expectedError  string
		invalidBuilder bool
	}{
		{
			name:        "sets valid IPv6 address",
			builder:     buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv6Address: "2001:db8::1",
		},
		{
			name:          "rejects invalid IPv6 address",
			builder:       buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv6Address:   "invalid",
			expectedError: "invalid IPv6 argument invalid",
		},
		{
			name:           "invalid builder short circuits",
			builder:        buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			ipv6Address:    "2001:db8::1",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithIPv6Address(testCase.ipv6Address)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.ipv6Address, result.Definition.Spec.IPv6.Address)
		})
	}
}

func TestIPConfigWithIPv4Gateway(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *IPConfigBuilder
		ipv4Address    string
		expectedError  string
		invalidBuilder bool
	}{
		{
			name:        "sets valid IPv4 gateway",
			builder:     buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv4Address: "192.168.1.1",
		},
		{
			name:          "rejects invalid IPv4 gateway",
			builder:       buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv4Address:   "invalid",
			expectedError: "invalid IPv4 address argument invalid",
		},
		{
			name:           "invalid builder short circuits",
			builder:        buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			ipv4Address:    "192.168.1.1",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithIPv4Gateway(testCase.ipv4Address)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.ipv4Address, result.Definition.Spec.IPv4.Gateway)
		})
	}
}

func TestIPConfigWithIPv6Gateway(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *IPConfigBuilder
		ipv6Address    string
		expectedError  string
		invalidBuilder bool
	}{
		{
			name:        "sets valid IPv6 gateway",
			builder:     buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv6Address: "2001:db8::1",
		},
		{
			name:          "rejects invalid IPv6 gateway",
			builder:       buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv6Address:   "invalid",
			expectedError: "invalid IPv6 argument invalid",
		},
		{
			name:           "invalid builder short circuits",
			builder:        buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			ipv6Address:    "2001:db8::1",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithIPv6Gateway(testCase.ipv6Address)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.ipv6Address, result.Definition.Spec.IPv6.Gateway)
		})
	}
}

func TestIPConfigWithIPv4MachineNetwork(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		builder            *IPConfigBuilder
		ipv4MachineNetwork string
		expectedError      string
		invalidBuilder     bool
	}{
		{
			name:               "sets valid IPv4 machine network",
			builder:            buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv4MachineNetwork: "192.168.1.0/24",
		},
		{
			name:               "rejects invalid IPv4 machine network",
			builder:            buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv4MachineNetwork: "invalid",
			expectedError:      "invalid CIDR argument invalid",
		},
		{
			name:               "invalid builder short circuits",
			builder:            buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			ipv4MachineNetwork: "192.168.1.0/24",
			invalidBuilder:     true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithIPv4MachineNetwork(testCase.ipv4MachineNetwork)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.ipv4MachineNetwork, result.Definition.Spec.IPv4.MachineNetwork)
		})
	}
}

func TestIPConfigWithIPv6MachineNetwork(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		builder            *IPConfigBuilder
		ipv6MachineNetwork string
		expectedError      string
		invalidBuilder     bool
	}{
		{
			name:               "sets valid IPv6 machine network",
			builder:            buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv6MachineNetwork: "2001:db8::/64",
		},
		{
			name:               "rejects invalid IPv6 machine network",
			builder:            buildValidIPConfigBuilderWithStacks(newIPConfigTestClient()),
			ipv6MachineNetwork: "invalid",
			expectedError:      "invalid CIDR argument invalid",
		},
		{
			name:               "invalid builder short circuits",
			builder:            buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			ipv6MachineNetwork: "2001:db8::/64",
			invalidBuilder:     true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithIPv6MachineNetwork(testCase.ipv6MachineNetwork)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.ipv6MachineNetwork, result.Definition.Spec.IPv6.MachineNetwork)
		})
	}
}

func TestIPConfigWithStage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *IPConfigBuilder
		stage          string
		invalidBuilder bool
	}{
		{
			name:    "sets stage",
			builder: buildValidIPConfigBuilder(newIPConfigTestClient()),
			stage:   string(lcaipcv1.IPStages.Config),
		},
		{
			name:           "invalid builder short circuits",
			builder:        buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			stage:          string(lcaipcv1.IPStages.Config),
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithStage(testCase.stage)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, lcaipcv1.IPConfigStage(testCase.stage), result.Definition.Spec.Stage)
		})
	}
}

func TestIPConfigWithVlanID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *IPConfigBuilder
		vlanID         int
		invalidBuilder bool
	}{
		{
			name:    "sets zero VLAN ID",
			builder: buildValidIPConfigBuilder(newIPConfigTestClient()),
			vlanID:  0,
		},
		{
			name:    "sets nonzero VLAN ID",
			builder: buildValidIPConfigBuilder(newIPConfigTestClient()),
			vlanID:  123,
		},
		{
			name:           "invalid builder short circuits",
			builder:        buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			vlanID:         123,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithVlanID(testCase.vlanID)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.vlanID, result.Definition.Spec.VLANID)
		})
	}
}

func TestIPConfigWithAutoRollbackOnFailure(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *IPConfigBuilder
		seconds        int
		invalidBuilder bool
	}{
		{
			name:    "sets positive timeout",
			builder: buildValidIPConfigBuilder(newIPConfigTestClient()),
			seconds: 10,
		},
		{
			name:    "sets negative timeout",
			builder: buildValidIPConfigBuilder(newIPConfigTestClient()),
			seconds: -1,
		},
		{
			name:           "invalid builder short circuits",
			builder:        buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			seconds:        10,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithAutoRollbackOnFailure(testCase.seconds)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.AutoRollbackOnFailure)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.AutoRollbackOnFailure)
			assert.Equal(t, testCase.seconds, result.Definition.Spec.AutoRollbackOnFailure.InitMonitorTimeoutSeconds)
		})
	}
}

func TestIPConfigWithDNS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		builder            *IPConfigBuilder
		dnsServers         []string
		expectedError      string
		expectedDNSServers []lcaipcv1.IPAddress
		invalidBuilder     bool
	}{
		{
			name:               "sets DNS servers",
			builder:            buildValidIPConfigBuilder(newIPConfigTestClient()),
			dnsServers:         []string{"8.8.8.8", "1.1.1.1"},
			expectedDNSServers: []lcaipcv1.IPAddress{"8.8.8.8", "1.1.1.1"},
		},
		{
			name:          "rejects empty DNS server list",
			builder:       buildValidIPConfigBuilder(newIPConfigTestClient()),
			dnsServers:    []string{},
			expectedError: "dns servers list cannot be empty",
		},
		{
			name:          "rejects blank DNS server entry",
			builder:       buildValidIPConfigBuilder(newIPConfigTestClient()),
			dnsServers:    []string{" "},
			expectedError: "dns server cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			builder:        buildInvalidIPConfigBuilder(newIPConfigTestClient()),
			dnsServers:     []string{"8.8.8.8"},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.builder.WithDNS(testCase.dnsServers)
			require.Same(t, testCase.builder, result)

			if testCase.invalidBuilder {
				require.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.expectedDNSServers, result.Definition.Spec.DNSServers)
		})
	}
}

func TestIPConfigWaitUntilComplete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		expectedError error
		status        lcaipcv1.IPConfigStatus
	}{
		{
			name:          "times out before config completes",
			expectedError: context.DeadlineExceeded,
			status: lcaipcv1.IPConfigStatus{
				Conditions: []metav1.Condition{{Status: "True1", Type: conditionTypeConfigCompleted, Reason: isComplete}},
			},
		},
		{
			name: "returns when config completes",
			status: lcaipcv1.IPConfigStatus{
				Conditions: []metav1.Condition{{Status: isTrue, Type: conditionTypeConfigCompleted, Reason: isComplete}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testIPConfig := generateIPConfig()
			testIPConfig.Status = testCase.status

			testIPConfigBuilder := buildValidIPConfigBuilder(
				buildIPConfigTestClientWithDummyObject([]runtime.Object{testIPConfig}))
			_, err := testIPConfigBuilder.WaitUntilComplete(time.Millisecond * 100)

			if testCase.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestIPConfigWaitUntilFailed(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		expectedError error
		status        lcaipcv1.IPConfigStatus
	}{
		{
			name:          "times out before config fails",
			expectedError: context.DeadlineExceeded,
			status: lcaipcv1.IPConfigStatus{
				Conditions: []metav1.Condition{{Status: isFalse, Type: conditionTypeConfigCompleted, Reason: isComplete}},
			},
		},
		{
			name: "returns when config fails",
			status: lcaipcv1.IPConfigStatus{
				Conditions: []metav1.Condition{{Status: isFalse, Type: conditionTypeConfigCompleted, Reason: "Failed"}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testIPConfig := generateIPConfig()
			testIPConfig.Status = testCase.status

			testIPConfigBuilder := buildValidIPConfigBuilder(
				buildIPConfigTestClientWithDummyObject([]runtime.Object{testIPConfig}))
			_, err := testIPConfigBuilder.WaitUntilFailed(time.Millisecond * 100)

			if testCase.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestIPConfigWaitUntilIdle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		expectedError error
		status        lcaipcv1.IPConfigStatus
	}{
		{
			name:          "times out before idle",
			expectedError: context.DeadlineExceeded,
			status: lcaipcv1.IPConfigStatus{
				Conditions: []metav1.Condition{{Status: "True1", Type: idle, Reason: idle}},
			},
		},
		{
			name: "returns when idle",
			status: lcaipcv1.IPConfigStatus{
				Conditions: []metav1.Condition{{Status: isTrue, Type: idle, Reason: idle}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testIPConfig := generateIPConfig()
			testIPConfig.Status = testCase.status

			testIPConfigBuilder := buildValidIPConfigBuilder(
				buildIPConfigTestClientWithDummyObject([]runtime.Object{testIPConfig}))
			_, err := testIPConfigBuilder.WaitUntilIdle(time.Millisecond * 100)

			if testCase.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

// newIPConfigCommonTestConfig returns the shared testhelper configuration for IPConfig builder tests.
func newIPConfigCommonTestConfig() testhelper.CommonTestConfig[
	lcaipcv1.IPConfig,
	IPConfigBuilder,
	*lcaipcv1.IPConfig,
	*IPConfigBuilder,
] {
	return testhelper.NewCommonTestConfig[lcaipcv1.IPConfig, IPConfigBuilder](
		lcaipcv1.AddToScheme,
		ipConfigGVK,
		testhelper.ResourceScopeClusterScoped,
	)
}

func generateIPConfig() *lcaipcv1.IPConfig {
	return &lcaipcv1.IPConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: ipConfigName,
		},
		Spec: lcaipcv1.IPConfigSpec{},
	}
}

func buildValidIPConfigBuilder(apiClient *clients.Settings) *IPConfigBuilder {
	return common.NewClusterScopedBuilder[lcaipcv1.IPConfig, IPConfigBuilder](
		apiClient,
		lcaipcv1.AddToScheme,
		ipConfigName,
	)
}

func buildValidIPConfigBuilderWithStacks(apiClient *clients.Settings) *IPConfigBuilder {
	builder := buildValidIPConfigBuilder(apiClient)

	builder.Definition.Spec.IPv4 = &lcaipcv1.IPv4Config{}
	builder.Definition.Spec.IPv6 = &lcaipcv1.IPv6Config{}

	return builder
}

func buildInvalidIPConfigBuilder(apiClient *clients.Settings) *IPConfigBuilder {
	return common.NewClusterScopedBuilder[lcaipcv1.IPConfig, IPConfigBuilder](
		apiClient,
		lcaipcv1.AddToScheme,
		"",
	)
}

func newIPConfigTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: lcaipcv1TestSchemes,
	})
}

func buildIPConfigTestClientWithDummyObject(objects []runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: lcaipcv1TestSchemes,
	})
}
