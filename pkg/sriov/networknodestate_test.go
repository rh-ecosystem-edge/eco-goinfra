package sriov

import (
	"fmt"
	"testing"
	"time"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultNodeName   = "test1"
	defaultNodeNsName = "testnamespace"
)

func TestNewNetworkNodeStateBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewNetworkNodeStateBuilder,
		srIovV1.AddToScheme,
		networkNodeStateGVK,
	).ExecuteTests(t)
}

func TestNetworkNodeStateBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[srIovV1.SriovNetworkNodeState, NetworkNodeStateBuilder](
		srIovV1.AddToScheme,
		networkNodeStateGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}

func TestNetworkNodeStateDiscovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		nodeName      string
		nsName        string
		objectExists  bool
		expectedError func(error) bool
	}{
		{
			name:         "existing resource",
			nodeName:     defaultNodeName,
			nsName:       defaultNodeNsName,
			objectExists: true,
		},
		{
			name:          "missing resource",
			nodeName:      defaultNodeName,
			nsName:        defaultNodeNsName,
			expectedError: k8serrors.IsNotFound,
		},
		{
			name:          "empty node name",
			nodeName:      "",
			nsName:        defaultNodeNsName,
			expectedError: commonerrors.IsBuilderNameEmpty,
		},
		{
			name:          "empty namespace",
			nodeName:      defaultNodeName,
			nsName:        "",
			expectedError: commonerrors.IsBuilderNamespaceEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.objectExists {
				objects = []runtime.Object{buildNodeNetworkState(testCase.nodeName, testCase.nsName)}
			}

			builder := NewNetworkNodeStateBuilder(getNetworkNodeStateTestClient(objects...), testCase.nodeName, testCase.nsName)
			err := builder.Discover()

			if testCase.expectedError == nil {
				require.NoError(t, err)
				require.NotNil(t, builder.Objects)
				assert.Same(t, builder.Object, builder.Objects)
				assert.Equal(t, testCase.nodeName, builder.Objects.Name)
				assert.Equal(t, testCase.nsName, builder.Objects.Namespace)
			} else {
				require.Error(t, err)
				assert.Truef(t, testCase.expectedError(err), "unexpected error, got: %v", err)
				assert.Nil(t, builder.Objects)
			}
		})
	}
}

func TestNetworkNodeStateGetNICs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		interfaces   srIovV1.InterfaceExts
		nodeName     string
		expectError  bool
		errorMatches func(error) bool
	}{
		{
			name:       "one interface",
			interfaces: srIovV1.InterfaceExts{{Name: "eth1"}},
			nodeName:   defaultNodeName,
		},
		{
			name:       "multiple interfaces",
			interfaces: srIovV1.InterfaceExts{{Name: "eth1"}, {Name: "eth2"}},
			nodeName:   defaultNodeName,
		},
		{
			name:       "empty interface list",
			interfaces: srIovV1.InterfaceExts{},
			nodeName:   defaultNodeName,
		},
		{
			name:         "invalid builder",
			interfaces:   srIovV1.InterfaceExts{{Name: "eth1"}},
			nodeName:     "",
			expectError:  true,
			errorMatches: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			networkNodeState := buildNodeNetworkStateWithNics(testCase.interfaces)
			builder := NewNetworkNodeStateBuilder(
				getNetworkNodeStateTestClient(networkNodeState), testCase.nodeName, defaultNodeNsName)

			nics, err := builder.GetNICs()
			if testCase.expectError {
				require.Error(t, err)
				assert.Truef(t, testCase.errorMatches(err), "unexpected error, got: %v", err)
				assert.Nil(t, nics)

				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, testCase.interfaces, nics)
			assert.Same(t, builder.Object, builder.Objects)
		})
	}
}

func TestNetworkNodeStateGetUpNICs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		interfaces   srIovV1.InterfaceExts
		upInterfaces srIovV1.InterfaceExts
		nodeName     string
		expectError  bool
		errorMatches func(error) bool
	}{
		{
			name: "all interfaces up",
			interfaces: srIovV1.InterfaceExts{
				{Name: "eth1", LinkSpeed: "1000Mb"},
				{Name: "eth2", LinkSpeed: "1000Mb"},
			},
			upInterfaces: srIovV1.InterfaceExts{
				{Name: "eth1", LinkSpeed: "1000Mb"},
				{Name: "eth2", LinkSpeed: "1000Mb"},
			},
			nodeName: defaultNodeName,
		},
		{
			name: "filters interfaces without link speed",
			interfaces: srIovV1.InterfaceExts{
				{Name: "eth1", LinkSpeed: "1000Mb"},
				{Name: "eth2"},
			},
			upInterfaces: srIovV1.InterfaceExts{{Name: "eth1", LinkSpeed: "1000Mb"}},
			nodeName:     defaultNodeName,
		},
		{
			name: "filters down interfaces",
			interfaces: srIovV1.InterfaceExts{
				{Name: "eth1", LinkSpeed: "-1 Mb/s"},
				{Name: "eth2", LinkSpeed: "1000Mb"},
			},
			upInterfaces: srIovV1.InterfaceExts{{Name: "eth2", LinkSpeed: "1000Mb"}},
			nodeName:     defaultNodeName,
		},
		{
			name:         "invalid builder",
			interfaces:   srIovV1.InterfaceExts{{Name: "eth1", LinkSpeed: "1000Mb"}},
			nodeName:     "",
			expectError:  true,
			errorMatches: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			networkNodeState := buildNodeNetworkStateWithNics(testCase.interfaces)
			builder := NewNetworkNodeStateBuilder(
				getNetworkNodeStateTestClient(networkNodeState), testCase.nodeName, defaultNodeNsName)

			nics, err := builder.GetUpNICs()
			if testCase.expectError {
				require.Error(t, err)
				assert.Truef(t, testCase.errorMatches(err), "unexpected error, got: %v", err)
				assert.Nil(t, nics)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.upInterfaces, nics)
		})
	}
}

func TestNetworkNodeStateWaitUntilSyncStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		syncStatus    string
		nodeName      string
		expectedError error
		errorMatches  func(error) bool
	}{
		{
			name:       "matching status",
			syncStatus: "Succeeded",
			nodeName:   defaultNodeName,
		},
		{
			name:          "empty status",
			syncStatus:    "",
			nodeName:      defaultNodeName,
			expectedError: fmt.Errorf("syncStatus cannot be empty"),
		},
		{
			name:         "invalid builder",
			syncStatus:   "Succeeded",
			nodeName:     "",
			errorMatches: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			networkNodeState := buildNodeNetworkStateSyncStatus(
				defaultNodeName, defaultNodeNsName, "Succeeded")
			builder := NewNetworkNodeStateBuilder(
				getNetworkNodeStateTestClient(networkNodeState), testCase.nodeName, defaultNodeNsName)

			err := builder.Discover()
			if testCase.nodeName == "" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			err = builder.WaitUntilSyncStatus(testCase.syncStatus, 30*time.Second)
			if testCase.errorMatches != nil {
				require.Error(t, err)
				assert.Truef(t, testCase.errorMatches(err), "unexpected error, got: %v", err)

				return
			}

			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

//nolint:funlen // Table-driven coverage for every interface accessor.
func TestNetworkNodeStateInterfaceAccessors(t *testing.T) {
	t.Parallel()

	interfaces := srIovV1.InterfaceExts{
		{Name: "eth1", NumVfs: 5, TotalVfs: 8, Driver: "ice", PciAddress: "0000:01:00.0"},
		{Name: "eth2", NumVfs: 10, TotalVfs: 16, Driver: "mlx5_core", PciAddress: "0000:02:00.0"},
	}

	accessors := []struct {
		name     string
		getValue func(*NetworkNodeStateBuilder, string) (any, error)
		expected any
	}{
		{
			name:     "num VFs",
			getValue: func(builder *NetworkNodeStateBuilder, name string) (any, error) { return builder.GetNumVFs(name) },
			expected: 10,
		},
		{
			name:     "total VFs",
			getValue: func(builder *NetworkNodeStateBuilder, name string) (any, error) { return builder.GetTotalVFs(name) },
			expected: 16,
		},
		{
			name: "driver name",
			getValue: func(builder *NetworkNodeStateBuilder, name string) (any, error) {
				return builder.GetDriverName(name)
			},
			expected: "mlx5_core",
		},
		{
			name: "PCI address",
			getValue: func(builder *NetworkNodeStateBuilder, name string) (any, error) {
				return builder.GetPciAddress(name)
			},
			expected: "0000:02:00.0",
		},
	}

	for _, accessor := range accessors {
		t.Run(accessor.name, func(t *testing.T) {
			t.Parallel()

			testCases := []struct {
				name          string
				interfaceName string
				nodeName      string
				expectedError error
				errorMatches  func(error) bool
			}{
				{name: "found", interfaceName: "eth2", nodeName: defaultNodeName},
				{
					name:          "interface not found",
					interfaceName: "eth3",
					nodeName:      defaultNodeName,
					expectedError: fmt.Errorf("interface eth3 was not found"),
				},
				{
					name:          "empty interface name",
					interfaceName: "",
					nodeName:      defaultNodeName,
					expectedError: fmt.Errorf("sriovInterface can not be empty string"),
				},
				{
					name:          "invalid builder",
					interfaceName: "eth2",
					nodeName:      "",
					errorMatches:  commonerrors.IsBuilderNameEmpty,
				},
			}

			for _, testCase := range testCases {
				t.Run(testCase.name, func(t *testing.T) {
					t.Parallel()

					networkNodeState := buildNodeNetworkStateWithNics(interfaces)
					builder := NewNetworkNodeStateBuilder(
						getNetworkNodeStateTestClient(networkNodeState), testCase.nodeName, defaultNodeNsName)

					value, err := accessor.getValue(builder, testCase.interfaceName)
					if testCase.errorMatches != nil {
						require.Error(t, err)
						assert.Truef(t, testCase.errorMatches(err), "unexpected error, got: %v", err)

						return
					}

					assert.Equal(t, testCase.expectedError, err)

					if testCase.expectedError == nil {
						assert.Equal(t, accessor.expected, value)
					}
				})
			}
		})
	}
}

func getNetworkNodeStateTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: testSchemes,
	})
}

func buildNodeNetworkState(name, nsName string) *srIovV1.SriovNetworkNodeState {
	return &srIovV1.SriovNetworkNodeState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsName,
		},
		Spec:   srIovV1.SriovNetworkNodeStateSpec{},
		Status: srIovV1.SriovNetworkNodeStateStatus{},
	}
}

func buildNodeNetworkStateSyncStatus(name, nsName, syncStatus string) *srIovV1.SriovNetworkNodeState {
	nodeNetworkState := buildNodeNetworkState(name, nsName)
	nodeNetworkState.Status.SyncStatus = syncStatus

	return nodeNetworkState
}

func buildNodeNetworkStateWithNics(sriovInterfaces srIovV1.InterfaceExts) *srIovV1.SriovNetworkNodeState {
	nodeNetworkState := buildNodeNetworkState(defaultNodeName, defaultNodeNsName)
	nodeNetworkState.Status.Interfaces = sriovInterfaces

	return nodeNetworkState
}
