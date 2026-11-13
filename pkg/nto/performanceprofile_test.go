package nto //nolint:misspell

import (
	"testing"

	performanceprofilev2 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/performanceprofile/v2"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultPerformanceProfileName = "default"
	defaultIsolatedCPU            = "2-27,30-55"
	defaultReservedCPU            = "0-1,28-29"
	defaultMCPSelector            = map[string]string{"machineconfiguration.openshift.io/role": "worker"}
	defaultNodeSelector           = map[string]string{"node-role.kubernetes.io/worker": ""}
	defaultNumaTopology           = "single-numa-node"
	defaultHugePagesNodeOne       = int32(0)
	defaultHugePagesNodeTwo       = int32(1)
	defaultHugepageSize           = "2M"
	defaultHugepages              = []performanceprofilev2.HugePage{
		{
			Size:  performanceprofilev2.HugePageSize(defaultHugepageSize),
			Count: 32768,
			Node:  &defaultHugePagesNodeOne,
		},
	}
	defaultHugepagesTwoNumaNodes = []performanceprofilev2.HugePage{
		{
			Size:  performanceprofilev2.HugePageSize(defaultHugepageSize),
			Count: 32768,
			Node:  &defaultHugePagesNodeOne,
		},
		{
			Size:  performanceprofilev2.HugePageSize(defaultHugepageSize),
			Count: 32768,
			Node:  &defaultHugePagesNodeTwo,
		},
	}
	defaultNetInterfaceNameOne = "ens2f(0|1)"
	defaultNetInterfaceNameTwo = "ens3f(0|1)"
	defaultVendorID            = "123456"
	defaultDeviceID            = "7654321"
	emptyString                = ""
	paoTestSchemes             = []clients.SchemeAttacher{
		performanceprofilev2.AddToScheme,
	}
)

func TestPullPerformanceProfile(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig[performanceprofilev2.PerformanceProfile, Builder](
		Pull,
		performanceprofilev2.AddToScheme,
		performanceProfileGVK,
	).ExecuteTests(t)
}

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common cluster scoped builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewClusterScopedBuilderTestConfig[performanceprofilev2.PerformanceProfile, Builder](
			func(apiClient *clients.Settings, name string) *Builder {
				return NewBuilder(apiClient, name, defaultIsolatedCPU, defaultReservedCPU, defaultNodeSelector)
			},
			performanceprofilev2.AddToScheme,
			performanceProfileGVK,
		).ExecuteTests(t)
	})

	t.Run("performance profile specific inputs", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name          string
			cpuIsolated   string
			cpuReserved   string
			nodeSelector  map[string]string
			expectedError string
		}{
			{
				name:         "valid inputs",
				cpuIsolated:  defaultIsolatedCPU,
				cpuReserved:  defaultReservedCPU,
				nodeSelector: defaultNodeSelector,
			},
			{
				name:          "empty isolated CPU",
				cpuIsolated:   "",
				cpuReserved:   defaultReservedCPU,
				nodeSelector:  defaultNodeSelector,
				expectedError: "PerformanceProfile's 'cpuIsolated' is empty",
			},
			{
				name:          "empty reserved CPU",
				cpuIsolated:   defaultIsolatedCPU,
				cpuReserved:   "",
				nodeSelector:  defaultNodeSelector,
				expectedError: "PerformanceProfile's 'cpuReserved' is empty",
			},
			{
				name:          "empty node selector",
				cpuIsolated:   defaultIsolatedCPU,
				cpuReserved:   defaultReservedCPU,
				nodeSelector:  map[string]string{},
				expectedError: "PerformanceProfile's 'nodeSelector' is empty",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				builder := NewBuilder(
					clients.GetTestClients(clients.TestClientParams{SchemeAttachers: paoTestSchemes}),
					defaultPerformanceProfileName,
					testCase.cpuIsolated,
					testCase.cpuReserved,
					testCase.nodeSelector,
				)

				require.NotNil(t, builder)
				require.NotNil(t, builder.Definition)
				require.NotNil(t, builder.Definition.Spec.CPU)
				assert.Equal(t, performanceprofilev2.CPUSet(testCase.cpuIsolated),
					*builder.Definition.Spec.CPU.Isolated)
				assert.Equal(t, performanceprofilev2.CPUSet(testCase.cpuReserved),
					*builder.Definition.Spec.CPU.Reserved)
				assert.Equal(t, testCase.nodeSelector, builder.Definition.Spec.NodeSelector)

				if testCase.expectedError != "" {
					assert.EqualError(t, builder.GetError(), testCase.expectedError)

					return
				}

				require.NoError(t, builder.GetError())
			})
		}
	})
}

func TestPerformanceProfileBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[performanceprofilev2.PerformanceProfile, Builder](
		performanceprofilev2.AddToScheme,
		performanceProfileGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

//nolint:funlen // Table-driven test cases are intentionally colocated.
func TestPerformanceProfileWithHugePages(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		hugePageSize      string
		pages             []performanceprofilev2.HugePage
		validBuilder      bool
		expectedErrorText string
	}{
		{
			name:         "valid 2M huge pages",
			hugePageSize: defaultHugepageSize,
			pages:        defaultHugepages,
			validBuilder: true,
		},
		{
			name:         "valid 1G huge pages",
			hugePageSize: "1G",
			pages: []performanceprofilev2.HugePage{
				{
					Size:  "1G",
					Count: 32768,
					Node:  &defaultHugePagesNodeOne,
				},
			},
			validBuilder: true,
		},
		{
			name:         "valid pages on two NUMA nodes",
			hugePageSize: defaultHugepageSize,
			pages:        defaultHugepagesTwoNumaNodes,
			validBuilder: true,
		},
		{
			name:         "valid page without NUMA node",
			hugePageSize: defaultHugepageSize,
			pages: []performanceprofilev2.HugePage{
				{
					Size:  performanceprofilev2.HugePageSize(defaultHugepageSize),
					Count: 32768,
				},
			},
			validBuilder: true,
		},
		{
			name:              "invalid huge page size",
			hugePageSize:      "2G",
			pages:             defaultHugepages,
			validBuilder:      true,
			expectedErrorText: "'hugePageSize' argument is not in allowed list: [2M 1G]",
		},
		{
			name:              "empty huge page size",
			pages:             defaultHugepages,
			validBuilder:      true,
			expectedErrorText: "'hugePageSize' argument cannot be empty",
		},
		{
			name:              "empty huge pages",
			hugePageSize:      defaultHugepageSize,
			pages:             []performanceprofilev2.HugePage{},
			validBuilder:      true,
			expectedErrorText: "'hugePages' argument cannot be empty",
		},
		{
			name:         "invalid builder",
			hugePageSize: defaultHugepageSize,
			pages:        defaultHugepages,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPerformanceProfileBuilderForTest(testCase.validBuilder)
			result := builder.WithHugePages(testCase.hugePageSize, testCase.pages)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.HugePages)

				return
			}

			if testCase.expectedErrorText != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedErrorText)
				assert.Nil(t, result.Definition.Spec.HugePages)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.HugePages)
			assert.Equal(t, performanceprofilev2.HugePageSize(testCase.hugePageSize),
				*result.Definition.Spec.HugePages.DefaultHugePagesSize)
			assert.Equal(t, testCase.pages, result.Definition.Spec.HugePages.Pages)
		})
	}
}

func TestPerformanceProfileWithMachineConfigPoolSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		selector      map[string]string
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "valid selector",
			selector:     defaultMCPSelector,
			validBuilder: true,
		},
		{
			name:          "empty selector",
			selector:      map[string]string{},
			validBuilder:  true,
			expectedError: "'machineConfigPoolSelector' argument cannot be empty",
		},
		{
			name:     "invalid builder",
			selector: defaultMCPSelector,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPerformanceProfileBuilderForTest(testCase.validBuilder)
			result := builder.WithMachineConfigPoolSelector(testCase.selector)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.MachineConfigPoolSelector)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Spec.MachineConfigPoolSelector)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.selector, result.Definition.Spec.MachineConfigPoolSelector)
		})
	}
}

func TestPerformanceProfileWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		nodeSelector  map[string]string
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "valid node selector",
			nodeSelector: defaultNodeSelector,
			validBuilder: true,
		},
		{
			name:         "invalid builder",
			nodeSelector: map[string]string{"node-role.kubernetes.io/master": ""},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPerformanceProfileBuilderForTest(testCase.validBuilder)
			result := builder.WithNodeSelector(testCase.nodeSelector)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Equal(t, defaultNodeSelector, result.Definition.Spec.NodeSelector)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.nodeSelector, result.Definition.Spec.NodeSelector)
		})
	}

	t.Run("empty node selector returns error", func(t *testing.T) {
		t.Parallel()

		result := buildValidPerformanceProfileBuilder(
			clients.GetTestClients(clients.TestClientParams{SchemeAttachers: paoTestSchemes}),
		).WithNodeSelector(map[string]string{})

		assert.EqualError(t, result.GetError(), "'nodeSelector' argument cannot be empty")
	})
}

func TestPerformanceProfileWithNumaTopology(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		topologyPolicy string
		validBuilder   bool
		expectedError  string
	}{
		{
			name:           "best effort",
			topologyPolicy: "best-effort",
			validBuilder:   true,
		},
		{
			name:           "restricted",
			topologyPolicy: "restricted",
			validBuilder:   true,
		},
		{
			name:           "single NUMA node",
			topologyPolicy: "single-numa-node",
			validBuilder:   true,
		},
		{
			name:           "unsupported policy",
			topologyPolicy: "some-not-supported-policy",
			validBuilder:   true,
			expectedError: "'allowedTopologyPolicies' argument is not in allowed list " +
				"[best-effort restricted single-numa-node]",
		},
		{
			name:          "empty policy",
			validBuilder:  true,
			expectedError: "'topologyPolicy' argument cannot be empty",
		},
		{
			name:           "invalid builder",
			topologyPolicy: defaultNumaTopology,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPerformanceProfileBuilderForTest(testCase.validBuilder)
			result := builder.WithNumaTopology(testCase.topologyPolicy)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.NUMA)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Spec.NUMA)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.topologyPolicy, *result.Definition.Spec.NUMA.TopologyPolicy)
		})
	}
}

func TestPerformanceProfileWithRTKernel(t *testing.T) {
	t.Parallel()

	t.Run("valid builder", func(t *testing.T) {
		t.Parallel()

		result := buildPerformanceProfileBuilderForTest(true).WithRTKernel()

		require.NoError(t, result.GetError())
		require.NotNil(t, result.Definition.Spec.RealTimeKernel)
		assert.True(t, *result.Definition.Spec.RealTimeKernel.Enabled)
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		result := buildPerformanceProfileBuilderForTest(false).WithRTKernel()

		require.Error(t, result.GetError())
		assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
		assert.Nil(t, result.Definition.Spec.RealTimeKernel)
	})
}

func TestPerformanceProfileWithGloballyDisableIrqLoadBalancing(t *testing.T) {
	t.Parallel()

	t.Run("valid builder", func(t *testing.T) {
		t.Parallel()

		result := buildPerformanceProfileBuilderForTest(true).WithGloballyDisableIrqLoadBalancing()

		require.NoError(t, result.GetError())
		require.NotNil(t, result.Definition.Spec.GloballyDisableIrqLoadBalancing)
		assert.True(t, *result.Definition.Spec.GloballyDisableIrqLoadBalancing)
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		result := buildPerformanceProfileBuilderForTest(false).WithGloballyDisableIrqLoadBalancing()

		require.Error(t, result.GetError())
		assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
		assert.Nil(t, result.Definition.Spec.GloballyDisableIrqLoadBalancing)
	})
}

func TestPerformanceProfileWithWorkloadHints(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		rtHint              bool
		perPodPowerMgmtHint bool
		highPowerHint       bool
		validBuilder        bool
	}{
		{
			name:                "all enabled",
			rtHint:              true,
			perPodPowerMgmtHint: true,
			highPowerHint:       true,
			validBuilder:        true,
		},
		{
			name:                "real time disabled",
			perPodPowerMgmtHint: true,
			highPowerHint:       true,
			validBuilder:        true,
		},
		{
			name:          "per pod power management disabled",
			rtHint:        true,
			highPowerHint: true,
			validBuilder:  true,
		},
		{
			name:                "high power consumption disabled",
			rtHint:              true,
			perPodPowerMgmtHint: true,
			validBuilder:        true,
		},
		{
			name:         "all disabled",
			validBuilder: true,
		},
		{
			name:                "invalid builder",
			rtHint:              true,
			perPodPowerMgmtHint: true,
			highPowerHint:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPerformanceProfileBuilderForTest(testCase.validBuilder)
			result := builder.WithWorkloadHints(
				testCase.rtHint, testCase.perPodPowerMgmtHint, testCase.highPowerHint)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.WorkloadHints)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.WorkloadHints)
			assert.Equal(t, testCase.rtHint, *result.Definition.Spec.WorkloadHints.RealTime)
			assert.Equal(t, testCase.perPodPowerMgmtHint,
				*result.Definition.Spec.WorkloadHints.PerPodPowerManagement)
			assert.Equal(t, testCase.highPowerHint, *result.Definition.Spec.WorkloadHints.HighPowerConsumption)
		})
	}
}

func TestPerformanceProfileWithAnnotations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		annotations   map[string]string
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "single annotation",
			annotations:  map[string]string{"performance.openshift.io/ignore-cgroups-version": "true"},
			validBuilder: true,
		},
		{
			name: "experimental kubelet config annotation",
			annotations: map[string]string{
				"kubeletconfig.experimental": "{'systemReserved':{'cpu':'500m','memory':'28Gi'}}",
			},
			validBuilder: true,
		},
		{
			name: "multiple annotations",
			annotations: map[string]string{
				"performance.openshift.io/ignore-cgroups-version": "true",
				"kubeletconfig.experimental":                      "{\"systemReserved\":{\"cpu\":\"500m\",\"memory\":\"28Gi\"}}",
			},
			validBuilder: true,
		},
		{
			name:          "empty annotations",
			annotations:   map[string]string{},
			validBuilder:  true,
			expectedError: "'annotations' argument cannot be empty",
		},
		{
			name:         "invalid builder",
			annotations:  map[string]string{"test": "value"},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPerformanceProfileBuilderForTest(testCase.validBuilder)
			result := builder.WithAnnotations(testCase.annotations)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Annotations)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Annotations)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.annotations, result.Definition.Annotations)
		})
	}
}

//nolint:funlen // Table-driven test cases are intentionally colocated.
func TestPerformanceProfileWithNet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		userLevelNet  bool
		devices       []performanceprofilev2.Device
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "user level networking enabled",
			userLevelNet: true,
			devices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &defaultVendorID,
				DeviceID:      &defaultDeviceID,
			}},
			validBuilder: true,
		},
		{
			name: "user level networking disabled",
			devices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &defaultVendorID,
				DeviceID:      &defaultDeviceID,
			}},
			validBuilder: true,
		},
		{
			name:         "empty device ID",
			userLevelNet: true,
			devices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &defaultVendorID,
				DeviceID:      &emptyString,
			}},
			validBuilder: true,
		},
		{
			name:         "empty vendor ID",
			userLevelNet: true,
			devices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &emptyString,
				DeviceID:      &defaultDeviceID,
			}},
			validBuilder: true,
		},
		{
			name:         "interface only",
			userLevelNet: true,
			devices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
			}},
			validBuilder: true,
		},
		{
			name:         "empty interface name",
			userLevelNet: true,
			devices: []performanceprofilev2.Device{{
				InterfaceName: &emptyString,
				VendorID:      &defaultVendorID,
				DeviceID:      &emptyString,
			}},
			validBuilder: true,
		},
		{
			name:         "empty interface and vendor IDs",
			userLevelNet: true,
			devices: []performanceprofilev2.Device{{
				InterfaceName: &emptyString,
				VendorID:      &emptyString,
				DeviceID:      &defaultDeviceID,
			}},
			validBuilder: true,
		},
		{
			name:         "empty vendor and device IDs",
			userLevelNet: true,
			devices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
				VendorID:      &emptyString,
				DeviceID:      &emptyString,
			}},
			validBuilder: true,
		},
		{
			name:         "multiple interfaces",
			userLevelNet: true,
			devices: []performanceprofilev2.Device{
				{InterfaceName: &defaultNetInterfaceNameOne},
				{InterfaceName: &defaultNetInterfaceNameTwo},
			},
			validBuilder: true,
		},
		{
			name:          "empty devices",
			userLevelNet:  true,
			devices:       []performanceprofilev2.Device{},
			validBuilder:  true,
			expectedError: "'net' argument cannot be empty",
		},
		{
			name:         "invalid builder",
			userLevelNet: true,
			devices: []performanceprofilev2.Device{{
				InterfaceName: &defaultNetInterfaceNameOne,
			}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPerformanceProfileBuilderForTest(testCase.validBuilder)
			result := builder.WithNet(testCase.userLevelNet, testCase.devices)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.Net)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Spec.Net)

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Net)
			assert.Equal(t, testCase.userLevelNet, *result.Definition.Spec.Net.UserLevelNetworking)
			assert.Equal(t, testCase.devices, result.Definition.Spec.Net.Devices)
		})
	}
}

func TestPerformanceProfileWithAdditionalKernelArgs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		kernelArgs    []string
		validBuilder  bool
		expectedError string
	}{
		{
			name:         "single kernel argument",
			kernelArgs:   []string{"nohz_full=2-27,30-55"},
			validBuilder: true,
		},
		{
			name:         "multiple kernel arguments",
			kernelArgs:   []string{"nohz_full=2-27,30-55", "intel_idle.max_cstate=0", "audit=0"},
			validBuilder: true,
		},
		{
			name:          "empty kernel arguments",
			kernelArgs:    []string{},
			validBuilder:  true,
			expectedError: "'additionalKernelArgs' argument cannot be empty",
		},
		{
			name:         "invalid builder",
			kernelArgs:   []string{"audit=0"},
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildPerformanceProfileBuilderForTest(testCase.validBuilder)
			result := builder.WithAdditionalKernelArgs(testCase.kernelArgs)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Nil(t, result.Definition.Spec.AdditionalKernelArgs)

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
				assert.Nil(t, result.Definition.Spec.AdditionalKernelArgs)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.kernelArgs, result.Definition.Spec.AdditionalKernelArgs)
		})
	}
}

func buildPerformanceProfileBuilderForTest(valid bool) *Builder {
	if valid {
		return buildValidPerformanceProfileBuilder(
			clients.GetTestClients(clients.TestClientParams{SchemeAttachers: paoTestSchemes}))
	}

	return buildInvalidPerformanceProfileBuilder(
		clients.GetTestClients(clients.TestClientParams{SchemeAttachers: paoTestSchemes}))
}

func buildValidPerformanceProfileBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(
		apiClient,
		defaultPerformanceProfileName,
		defaultIsolatedCPU,
		defaultReservedCPU,
		defaultNodeSelector,
	)
}

func buildInvalidPerformanceProfileBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(
		apiClient,
		"",
		defaultIsolatedCPU,
		defaultReservedCPU,
		defaultNodeSelector,
	)
}

func buildPerformanceProfileWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPerformanceProfile(),
		SchemeAttachers: paoTestSchemes,
	})
}

func buildDummyPerformanceProfile() []runtime.Object {
	return append([]runtime.Object{}, &performanceprofilev2.PerformanceProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultPerformanceProfileName,
		},
	})
}
