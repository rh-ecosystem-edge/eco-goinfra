package sriov

import (
	"testing"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultPolicyName         = "sriovnet"
	defaultPolicyNsName       = "testnamespace"
	defaultPolicyResName      = "resname"
	defaultPolicyVFNum        = 1
	defaultPolicyNICs         = []string{"eth1"}
	defaultPolicyNodeSelector = map[string]string{"node": "selector"}
)

//nolint:funlen
func TestNewPolicyBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *PolicyBuilder {
				return NewPolicyBuilder(
					apiClient,
					name,
					nsname,
					defaultPolicyResName,
					defaultPolicyVFNum,
					defaultPolicyNICs,
					defaultPolicyNodeSelector,
				)
			},
			srIovV1.AddToScheme,
			policyGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		resourceName  string
		vfsNumber     int
		nicNames      []string
		nodeSelector  map[string]string
		expectedError string
	}{
		{
			name:         "valid policy",
			resourceName: defaultPolicyResName,
			vfsNumber:    defaultPolicyVFNum,
			nicNames:     defaultPolicyNICs,
			nodeSelector: defaultPolicyNodeSelector,
		},
		{
			name:          "empty resource name",
			resourceName:  "",
			vfsNumber:     defaultPolicyVFNum,
			nicNames:      defaultPolicyNICs,
			nodeSelector:  defaultPolicyNodeSelector,
			expectedError: "SriovNetworkNodePolicy 'resName' cannot be empty",
		},
		{
			name:          "empty NIC names",
			resourceName:  defaultPolicyResName,
			vfsNumber:     defaultPolicyVFNum,
			nicNames:      nil,
			nodeSelector:  defaultPolicyNodeSelector,
			expectedError: "SriovNetworkNodePolicy 'nicNames' cannot be empty list",
		},
		{
			name:          "empty node selector",
			resourceName:  defaultPolicyResName,
			vfsNumber:     defaultPolicyVFNum,
			nicNames:      defaultPolicyNICs,
			nodeSelector:  nil,
			expectedError: "SriovNetworkNodePolicy 'nodeSelector' cannot be empty map",
		},
		{
			name:          "zero VFs",
			resourceName:  defaultPolicyResName,
			vfsNumber:     0,
			nicNames:      defaultPolicyNICs,
			nodeSelector:  defaultPolicyNodeSelector,
			expectedError: "SriovNetworkNodePolicy 'vfsNumber' cannot be zero of negative",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewPolicyBuilder(
				buildTestClientWithDummyPolicyObject(),
				defaultPolicyName,
				defaultPolicyNsName,
				testCase.resourceName,
				testCase.vfsNumber,
				testCase.nicNames,
				testCase.nodeSelector,
			)

			require.NotNil(t, builder)

			if testCase.expectedError != "" {
				require.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.nodeSelector, builder.Definition.Spec.NodeSelector)
			assert.Equal(t, testCase.vfsNumber, builder.Definition.Spec.NumVfs)
			assert.Equal(t, testCase.resourceName, builder.Definition.Spec.ResourceName)
			assert.Equal(t, testCase.nicNames, builder.Definition.Spec.NicSelector.PfNames)
			assert.Equal(t, 1, builder.Definition.Spec.Priority)
		})
	}
}

func TestPullPolicy(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPolicy,
		srIovV1.AddToScheme,
		policyGVK,
	).ExecuteTests(t)
}

func TestPolicyBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := newPolicyCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestPolicyWithDevType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		devType        string
		invalidBuilder bool
		expectedError  string
	}{
		{name: "vfio-pci", devType: "vfio-pci"},
		{name: "netdevice", devType: "netdevice"},
		{name: "invalid device type", devType: "invalid", expectedError: errInvalidDeviceType},
		{name: "invalid builder", devType: "vfio-pci", invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			if testCase.invalidBuilder {
				builder = buildInvalidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			}

			result := builder.WithDevType(testCase.devType)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.devType, result.Definition.Spec.DeviceType)
		})
	}
}

func TestPolicyWithVFRange(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		firstVF        int
		lastVF         int
		invalidBuilder bool
		expectedError  string
		expectedPFs    []string
	}{
		{name: "full range", firstVF: 0, lastVF: 63, expectedPFs: []string{"eth1#0-63"}},
		{name: "partial range", firstVF: 1, lastVF: 7, expectedPFs: []string{"eth1#1-7"}},
		{
			name:          "negative first VF",
			firstVF:       -1,
			lastVF:        1,
			expectedError: "firstPF or lastVF can not be less than 0",
			expectedPFs:   defaultPolicyNICs,
		},
		{
			name:          "first VF greater than last VF",
			firstVF:       5,
			lastVF:        2,
			expectedError: "firstPF argument can not be greater than lastPF",
			expectedPFs:   defaultPolicyNICs,
		},
		{
			name:          "last VF too large",
			firstVF:       0,
			lastVF:        64,
			expectedError: "lastVF can not be greater than 63",
			expectedPFs:   defaultPolicyNICs,
		},
		{name: "invalid builder", firstVF: 0, lastVF: 1, invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			if testCase.invalidBuilder {
				builder = buildInvalidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			}

			result := builder.WithVFRange(testCase.firstVF, testCase.lastVF)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)
			} else {
				require.NoError(t, result.GetError())
			}

			assert.Equal(t, testCase.expectedPFs, result.Definition.Spec.NicSelector.PfNames)
		})
	}
}

func TestPolicyWithMTU(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		mtu            int
		invalidBuilder bool
		expectedError  string
	}{
		{name: "minimum MTU", mtu: 1},
		{name: "typical MTU", mtu: 1500},
		{name: "maximum MTU", mtu: 9192},
		{name: "zero MTU", mtu: 0, expectedError: "invalid mtu size 0 allowed mtu should be in range 1...9192"},
		{name: "large MTU", mtu: 9193, expectedError: "invalid mtu size 9193 allowed mtu should be in range 1...9192"},
		{name: "invalid builder", mtu: 1500, invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			if testCase.invalidBuilder {
				builder = buildInvalidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			}

			result := builder.WithMTU(testCase.mtu)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.mtu, result.Definition.Spec.Mtu)
		})
	}
}

func TestPolicyWithRDMA(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		rdma           bool
		invalidBuilder bool
	}{
		{name: "enable RDMA", rdma: true},
		{name: "disable RDMA", rdma: false},
		{name: "invalid builder", rdma: true, invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			if testCase.invalidBuilder {
				builder = buildInvalidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			}

			result := builder.WithRDMA(testCase.rdma)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.rdma, result.Definition.Spec.IsRdma)
		})
	}
}

func TestPolicyWithVhostNet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		vhostNet       bool
		invalidBuilder bool
	}{
		{name: "enable vhost-net", vhostNet: true},
		{name: "disable vhost-net", vhostNet: false},
		{name: "invalid builder", vhostNet: true, invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			if testCase.invalidBuilder {
				builder = buildInvalidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			}

			result := builder.WithVhostNet(testCase.vhostNet)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.vhostNet, result.Definition.Spec.NeedVhostNet)
		})
	}
}

func TestPolicyWithExternallyManaged(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		externallyManaged bool
		invalidBuilder    bool
	}{
		{name: "enable externally managed", externallyManaged: true},
		{name: "disable externally managed", externallyManaged: false},
		{name: "invalid builder", externallyManaged: true, invalidBuilder: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			if testCase.invalidBuilder {
				builder = buildInvalidPolicyTestBuilder(buildTestClientWithDummyPolicyObject())
			}

			result := builder.WithExternallyManaged(testCase.externallyManaged)
			require.Same(t, builder, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.externallyManaged, result.Definition.Spec.ExternallyManaged)
		})
	}
}

func TestPolicyWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newPolicyCommonTestConfig()).ExecuteTests(t)
}

func newPolicyCommonTestConfig() testhelper.CommonTestConfig[
	srIovV1.SriovNetworkNodePolicy,
	PolicyBuilder,
	*srIovV1.SriovNetworkNodePolicy,
	*PolicyBuilder,
] {
	return testhelper.NewCommonTestConfig[srIovV1.SriovNetworkNodePolicy, PolicyBuilder](
		srIovV1.AddToScheme,
		policyGVK,
		testhelper.ResourceScopeNamespaced,
	)
}

// buildValidPolicyTestBuilder returns a valid PolicyBuilder for testing purposes.
func buildValidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(
		apiClient,
		defaultPolicyName,
		defaultPolicyNsName,
		defaultPolicyResName,
		defaultPolicyVFNum,
		defaultPolicyNICs,
		defaultPolicyNodeSelector)
}

// buildInvalidPolicyTestBuilder returns a PolicyBuilder with an invalid resource name for testing purposes.
func buildInvalidPolicyTestBuilder(apiClient *clients.Settings) *PolicyBuilder {
	return NewPolicyBuilder(
		apiClient,
		defaultPolicyName,
		defaultPolicyNsName,
		"",
		defaultPolicyVFNum,
		defaultPolicyNICs,
		defaultPolicyNodeSelector)
}

// buildTestClientWithDummyPolicyObject returns a test client containing a dummy SriovNetworkNodePolicy.
func buildTestClientWithDummyPolicyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummySrIovPolicyObject(),
		SchemeAttachers: testSchemes,
	})
}

// buildDummySrIovPolicyObject returns the runtime objects for a default SriovNetworkNodePolicy.
func buildDummySrIovPolicyObject() []runtime.Object {
	return append([]runtime.Object{}, buildDummySrIovPolicy(defaultPolicyName, defaultPolicyNsName))
}

// buildDummySrIovPolicy returns a SriovNetworkNodePolicy with the supplied name and namespace.
func buildDummySrIovPolicy(name, namespace string) *srIovV1.SriovNetworkNodePolicy {
	return &srIovV1.SriovNetworkNodePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: srIovV1.SriovNetworkNodePolicySpec{
			ResourceName: defaultPolicyResName,
			NodeSelector: defaultPolicyNodeSelector,
			NumVfs:       defaultPolicyVFNum,
			NicSelector: srIovV1.SriovNetworkNicSelector{
				PfNames: defaultPolicyNICs,
			},
			Priority: 1,
		},
	}
}
