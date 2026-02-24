package lca

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	lcaipcv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ipchange/api/ipconfig/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	lcaipcv1TestSchemes = []clients.SchemeAttacher{
		lcaipcv1.AddToScheme,
	}
)

func TestIPConfigWithOptions(t *testing.T) {
	testBuilder := buildTestIPConfigBuilderWithFakeObjects()

	testBuilder.WithOptions(func(builder *IPConfigBuilder) (*IPConfigBuilder, error) {
		return builder, nil
	})

	assert.Equal(t, "", testBuilder.errorMsg)

	testBuilder.WithOptions(func(builder *IPConfigBuilder) (*IPConfigBuilder, error) {
		return builder, fmt.Errorf("error")
	})

	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestIPConfigPull(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
		client              bool
	}{

		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			expectedError:       fmt.Errorf("the apiClient is nil"),
			addToRuntimeObjects: true,
			client:              false,
		},
		{
			expectedError:       fmt.Errorf("ipconfig object ipconfig does not exist"),
			addToRuntimeObjects: false,
			client:              true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIPConfig := generateIPConfig(ipConfigName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIPConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: lcaipcv1TestSchemes,
			})
		}

		builderResult, err := PullIPConfig(testSettings)

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestIPConfigDelete(t *testing.T) {
	testCases := []struct {
		ipconfig      *IPConfigBuilder
		expectedError error
	}{
		{
			ipconfig:      buildValidIPConfigBuilder(buildIPConfigTestClientWithDummyObject([]runtime.Object{})),
			expectedError: nil,
		},
		{
			ipconfig: buildValidIPConfigBuilder(
				buildIPConfigTestClientWithDummyObject(buildDummyIPConfigRuntime())),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testIPConfigBuilder, err := testCase.ipconfig.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testIPConfigBuilder.Object)
		}
	}
}

func TestIPConfigUpdate(t *testing.T) {
	testCases := []struct {
		ipconfig      *IPConfigBuilder
		expectedError string
	}{
		{
			ipconfig: buildValidIPConfigBuilderWithStacks(
				buildIPConfigTestClientWithDummyObject(buildIPConfigRuntimeWithResourceVersion())),
			expectedError: "",
		},
		{
			ipconfig: buildValidIPConfigBuilderWithStacks(
				buildIPConfigTestClientWithDummyObject([]runtime.Object{})),
			expectedError: "cannot update non-existing ipconfig",
		},
	}

	for _, testCase := range testCases {
		testCase.ipconfig.WithIPv4Address("192.168.1.10")

		updatedBuilder, err := testCase.ipconfig.Update()
		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, updatedBuilder.Object)
			assert.Equal(t, updatedBuilder.Definition.Spec.IPv4.Address, updatedBuilder.Object.Spec.IPv4.Address)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestIPConfigGet(t *testing.T) {
	testCases := []struct {
		ipconfig      *IPConfigBuilder
		expectedError error
	}{
		{
			ipconfig: buildValidIPConfigBuilder(
				buildIPConfigTestClientWithDummyObject(buildDummyIPConfigRuntime())),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testIPConfig, err := testCase.ipconfig.Get()
		if err != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, testIPConfig.Name, testCase.ipconfig.Definition.Name)
		}
	}
}

func TestIPConfigExists(t *testing.T) {
	testCases := []struct {
		ipconfig      *IPConfigBuilder
		expectedExist bool
	}{
		{
			ipconfig:      buildValidIPConfigBuilder(buildIPConfigTestClientWithDummyObject([]runtime.Object{})),
			expectedExist: false,
		},
		{
			ipconfig:      buildInValidIPConfigBuilder(buildIPConfigTestClientWithDummyObject([]runtime.Object{})),
			expectedExist: false,
		},
		{
			ipconfig: buildValidIPConfigBuilder(
				buildIPConfigTestClientWithDummyObject(buildDummyIPConfigRuntime())),
			expectedExist: true,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.ipconfig.Exists()
		assert.Equal(t, testCase.expectedExist, exist)
	}
}

func TestIPConfigWithIPv4Address(t *testing.T) {
	testCases := []struct {
		ipv4Address string
		expectedErr string
	}{
		{
			ipv4Address: "192.168.1.10",
			expectedErr: "",
		},
		{
			ipv4Address: "invalid",
			expectedErr: "invalid IPv4 address argument invalid",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidIPConfigBuilderWithStacks(
			buildIPConfigTestClientWithDummyObject([]runtime.Object{}))

		testBuilder.WithIPv4Address(testCase.ipv4Address)
		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.ipv4Address, testBuilder.Definition.Spec.IPv4.Address)
		}
	}
}

func TestIPConfigWithIPv6Address(t *testing.T) {
	testCases := []struct {
		ipv6Address string
		expectedErr string
	}{
		{
			ipv6Address: "2001:db8::1",
			expectedErr: "",
		},
		{
			ipv6Address: "invalid",
			expectedErr: "invalid IPv6 argument invalid",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidIPConfigBuilderWithStacks(
			buildIPConfigTestClientWithDummyObject([]runtime.Object{}))

		testBuilder.WithIPv6Address(testCase.ipv6Address)
		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.ipv6Address, testBuilder.Definition.Spec.IPv6.Address)
		}
	}
}

func TestIPConfigWithIPv4Gateway(t *testing.T) {
	testCases := []struct {
		ipv4Address string
		expectedErr string
	}{
		{
			ipv4Address: "192.168.1.1",
			expectedErr: "",
		},
		{
			ipv4Address: "invalid",
			expectedErr: "invalid IPv4 address argument invalid",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidIPConfigBuilderWithStacks(
			buildIPConfigTestClientWithDummyObject([]runtime.Object{}))

		testBuilder.WithIPv4Gateway(testCase.ipv4Address)
		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.ipv4Address, testBuilder.Definition.Spec.IPv4.Gateway)
		}
	}
}

func TestIPConfigWithIPv6Gateway(t *testing.T) {
	testCases := []struct {
		ipv6Address string
		expectedErr string
	}{
		{
			ipv6Address: "2001:db8::1",
			expectedErr: "",
		},
		{
			ipv6Address: "invalid",
			expectedErr: "invalid IPv6 argument invalid",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidIPConfigBuilderWithStacks(
			buildIPConfigTestClientWithDummyObject([]runtime.Object{}))

		testBuilder.WithIPv6Gateway(testCase.ipv6Address)
		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.ipv6Address, testBuilder.Definition.Spec.IPv6.Gateway)
		}
	}
}

func TestIPConfigWithIPv4MachineNetwork(t *testing.T) {
	testCases := []struct {
		ipv4MachineNetwork string
		expectedErr        string
	}{
		{
			ipv4MachineNetwork: "192.168.1.0/24",
			expectedErr:        "",
		},
		{
			ipv4MachineNetwork: "invalid",
			expectedErr:        "invalid CIDR argument invalid",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidIPConfigBuilderWithStacks(
			buildIPConfigTestClientWithDummyObject([]runtime.Object{}))

		testBuilder.WithIPv4MachineNetwork(testCase.ipv4MachineNetwork)
		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.ipv4MachineNetwork, testBuilder.Definition.Spec.IPv4.MachineNetwork)
		}
	}
}

func TestIPConfigWithIPv6MachineNetwork(t *testing.T) {
	testCases := []struct {
		ipv6MachineNetwork string
		expectedErr        string
	}{
		{
			ipv6MachineNetwork: "2001:db8::/64",
			expectedErr:        "",
		},
		{
			ipv6MachineNetwork: "invalid",
			expectedErr:        "invalid CIDR argument invalid",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidIPConfigBuilderWithStacks(
			buildIPConfigTestClientWithDummyObject([]runtime.Object{}))

		testBuilder.WithIPv6MachineNetwork(testCase.ipv6MachineNetwork)
		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.ipv6MachineNetwork, testBuilder.Definition.Spec.IPv6.MachineNetwork)
		}
	}
}

func TestIPConfigWithStage(t *testing.T) {
	testBuilder := buildValidIPConfigBuilder(buildIPConfigTestClientWithDummyObject([]runtime.Object{}))

	testBuilder.WithStage(string(lcaipcv1.IPStages.Config))

	assert.Equal(t, lcaipcv1.IPStages.Config, testBuilder.Definition.Spec.Stage)
}

func TestIPConfigWithVlanID(t *testing.T) {
	testCases := []struct {
		vlanID int
	}{
		{
			vlanID: 0,
		},
		{
			vlanID: 123,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidIPConfigBuilder(buildIPConfigTestClientWithDummyObject([]runtime.Object{}))

		testBuilder.WithVlanID(testCase.vlanID)

		assert.Equal(t, testCase.vlanID, testBuilder.Definition.Spec.VLANID)
	}
}

func TestIPConfigWithAutoRollbackOnFailure(t *testing.T) {
	testCases := []struct {
		seconds int
	}{
		{seconds: 10},
		{seconds: -1},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidIPConfigBuilder(buildIPConfigTestClientWithDummyObject([]runtime.Object{}))
		testBuilder.WithAutoRollbackOnFailure(testCase.seconds)
		assert.NotNil(t, testBuilder.Definition.Spec.AutoRollbackOnFailure)
		assert.Equal(t, testCase.seconds, testBuilder.Definition.Spec.AutoRollbackOnFailure.InitMonitorTimeoutSeconds)
	}
}

func TestIPConfigWithDNS(t *testing.T) {
	testCases := []struct {
		dnsServers         []string
		expectedErr        string
		expectedDNSServers []lcaipcv1.IPAddress
	}{
		{
			dnsServers:         []string{"8.8.8.8", "1.1.1.1"},
			expectedErr:        "",
			expectedDNSServers: []lcaipcv1.IPAddress{"8.8.8.8", "1.1.1.1"},
		},
		{
			dnsServers:  []string{},
			expectedErr: "dns servers list cannot be empty",
		},
		{
			dnsServers:  []string{" "},
			expectedErr: "dns server cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidIPConfigBuilder(buildIPConfigTestClientWithDummyObject([]runtime.Object{}))

		testBuilder.WithDNS(testCase.dnsServers)
		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.Equal(t, testCase.expectedDNSServers, testBuilder.Definition.Spec.DNSServers)
		}
	}
}

func TestIPConfigWaitUntilComplete(t *testing.T) {
	testCases := []struct {
		expectedError error
		status        lcaipcv1.IPConfigStatus
	}{
		{
			expectedError: context.DeadlineExceeded,
			status: lcaipcv1.IPConfigStatus{
				Conditions: []metav1.Condition{{Status: "True1", Type: "ConfigCompleted", Reason: "Completed"}},
			},
		},
		{
			expectedError: nil,
			status: lcaipcv1.IPConfigStatus{
				Conditions: []metav1.Condition{{Status: "True", Type: "ConfigCompleted", Reason: "Completed"}},
			},
		},
	}

	for _, testCase := range testCases {
		testIPConfig := generateIPConfig(ipConfigName)
		testIPConfig.Status = testCase.status

		var runtimeObjects []runtime.Object

		runtimeObjects = append(runtimeObjects, testIPConfig)

		testIPConfigBuilder := buildValidIPConfigBuilder(
			buildIPConfigTestClientWithDummyObject(runtimeObjects))
		_, err := testIPConfigBuilder.WaitUntilComplete(time.Millisecond * 100)

		assert.Equal(t, testCase.expectedError, err)
	}
}

func buildTestIPConfigBuilderWithFakeObjects() *IPConfigBuilder {
	apiClient := buildIPConfigTestClientWithDummyObject(buildDummyIPConfigRuntime())

	return &IPConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &lcaipcv1.IPConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: ipConfigName,
			},
		},
	}
}

func generateIPConfig(name string) *lcaipcv1.IPConfig {
	return &lcaipcv1.IPConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: lcaipcv1.IPConfigSpec{},
	}
}

func buildValidIPConfigBuilder(apiClient *clients.Settings) *IPConfigBuilder {
	return &IPConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &lcaipcv1.IPConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: ipConfigName,
			},
		},
	}
}

func buildValidIPConfigBuilderWithStacks(apiClient *clients.Settings) *IPConfigBuilder {
	builder := &IPConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &lcaipcv1.IPConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: ipConfigName,
			},
		},
	}

	builder.Definition.Spec.IPv4 = &lcaipcv1.IPv4Config{}
	builder.Definition.Spec.IPv6 = &lcaipcv1.IPv6Config{}

	return builder
}

func buildInValidIPConfigBuilder(apiClient *clients.Settings) *IPConfigBuilder {
	return &IPConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &lcaipcv1.IPConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: "",
			},
		}}
}

func buildIPConfigTestClientWithDummyObject(objects []runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: lcaipcv1TestSchemes,
	})
}

func buildDummyIPConfigRuntime() []runtime.Object {
	return append([]runtime.Object{}, &lcaipcv1.IPConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: ipConfigName,
		},
		Spec: lcaipcv1.IPConfigSpec{},
	})
}

func buildIPConfigRuntimeWithResourceVersion() []runtime.Object {
	return append([]runtime.Object{}, &lcaipcv1.IPConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ipConfigName,
			ResourceVersion: "1",
		},
		Spec: lcaipcv1.IPConfigSpec{},
	})
}
