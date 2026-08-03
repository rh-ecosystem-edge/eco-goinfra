package bmh

import (
	"context"
	"fmt"
	"testing"
	"time"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBareMetalHostList(t *testing.T) {
	testCases := []struct {
		BareMetalHosts []*BmhBuilder
		nsName         string
		listOptions    []goclient.ListOptions
		expectedError  error
		client         bool
	}{
		{

			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "test-namespace",
			expectedError:  nil,
			client:         true,
		},
		{
			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "",
			expectedError:  fmt.Errorf("failed to list bareMetalHosts, 'nsname' parameter is empty"),
			client:         true,
		},
		{
			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "test-namespace",
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:         true,
			expectedError:  nil,
		},
		{
			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "test-namespace",
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}, {LabelSelector: labels.NewSelector()}},
			expectedError:  fmt.Errorf("error: more than one ListOptions was passed"),
			client:         true,
		},
		{
			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "test-namespace",
			expectedError:  fmt.Errorf("failed to list bareMetalHosts, 'apiClient' parameter is empty"),
			client:         false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyBmHostObject(bmhv1alpha1.StateProvisioned),
				SchemeAttachers: testSchemes,
			})
		}

		bmhBuilders, err := List(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.BareMetalHosts), len(bmhBuilders))
		}
	}
}

func TestBareMetalHostListInAllNamespaces(t *testing.T) {
	testCases := []struct {
		bareMetalHosts []*BmhBuilder
		listOptions    []goclient.ListOptions
		expectedError  error
		client         bool
	}{
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    nil,
			expectedError:  nil,
			client:         true,
		},
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError:  nil,
			client:         true,
		},
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}, {LabelSelector: labels.NewSelector()}},
			expectedError:  fmt.Errorf("error: more than one ListOptions was passed"),
			client:         true,
		},
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError:  fmt.Errorf("failed to list bareMetalHosts, 'apiClient' parameter is empty"),
			client:         false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildBareMetalHostTestClientWithDummyObject()
		}

		bmhBuilders, err := ListInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.bareMetalHosts), len(bmhBuilders))
		}
	}
}

func TestBareMetalWaitForAllBareMetalHostsInGoodOperationalState(t *testing.T) {
	testCases := []struct {
		BareMetalHosts   []*BmhBuilder
		nsName           string
		listOptions      []goclient.ListOptions
		operationalState bmhv1alpha1.OperationalStatus
		expectedError    error
		client           bool
		expectedStatus   bool
	}{
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			listOptions:      nil,
			expectedError:    nil,
			expectedStatus:   true,
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusDelayed,
			expectedError:    context.DeadlineExceeded,
			listOptions:      nil,
			expectedStatus:   false,
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    fmt.Errorf("failed to list bareMetalHosts, 'nsname' parameter is empty"),
			expectedStatus:   false,
			listOptions:      nil,
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    fmt.Errorf("failed to list bareMetalHosts, 'apiClient' parameter is empty"),
			expectedStatus:   false,
			listOptions:      nil,
			client:           false,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    nil,
			expectedStatus:   true,
			listOptions:      []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    fmt.Errorf("error: more than one ListOptions was passed"),
			expectedStatus:   false,
			listOptions: []goclient.ListOptions{
				{LabelSelector: labels.NewSelector()}, {LabelSelector: labels.NewSelector()}},
			client: true,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyBmHostObject(bmhv1alpha1.StateProvisioned, testCase.operationalState),
				SchemeAttachers: testSchemes,
			})
		}

		status, err := WaitForAllBareMetalHostsInGoodOperationalState(
			testSettings, testCase.nsName, 1*time.Second, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedStatus, status)
	}
}

func TestListInventoryEligibleBMH(t *testing.T) {
	testCases := []struct {
		name           string
		bareMetalHosts []runtime.Object
		listOptions    []goclient.ListOptions
		client         bool
		expectedError  error
		expectedNames  []string
	}{
		{
			name: "returns only inventory-eligible baremetalhosts",
			bareMetalHosts: []runtime.Object{
				buildInventoryEligibleBMH("eligible-bmh", defaultBmHostNsName, bmhv1alpha1.StateProvisioned),
				buildDummyBmHost(bmhv1alpha1.StateProvisioned, bmhv1alpha1.OperationalStatusOK),
				buildInventoryEligibleBMH("eligible-available", defaultBmHostNsName, bmhv1alpha1.StateAvailable),
				buildInventoryEligibleBMH("ineligible-state", defaultBmHostNsName, bmhv1alpha1.StateInspecting),
			},
			expectedNames: []string{"eligible-bmh", "eligible-available"},
			client:        true,
		},
		{
			name: "includes resource selector label without resource pool name",
			bareMetalHosts: []runtime.Object{
				buildInventoryEligibleBMHWithLabels("selector-bmh", defaultBmHostNsName, bmhv1alpha1.StateProvisioned, map[string]string{
					labelPrefixResourceSelector + "zone": "east",
				}),
			},
			expectedNames: []string{"selector-bmh"},
			client:        true,
		},
		{
			name: "supports list options",
			bareMetalHosts: []runtime.Object{
				buildInventoryEligibleBMH("eligible-bmh", defaultBmHostNsName, bmhv1alpha1.StateProvisioned),
			},
			listOptions:   []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedNames: []string{"eligible-bmh"},
			client:        true,
		},
		{
			name:          "rejects multiple list options",
			listOptions:   []goclient.ListOptions{{LabelSelector: labels.NewSelector()}, {LabelSelector: labels.NewSelector()}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			name:          "requires api client",
			expectedError: fmt.Errorf("failed to list bareMetalHosts, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var testSettings *clients.Settings

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  testCase.bareMetalHosts,
					SchemeAttachers: testSchemes,
				})
			}

			bmhBuilders, err := ListInventoryEligibleBMH(testSettings, testCase.listOptions...)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError != nil {
				return
			}

			assert.Len(t, bmhBuilders, len(testCase.expectedNames))

			names := make([]string, 0, len(bmhBuilders))
			for _, builder := range bmhBuilders {
				names = append(names, builder.Definition.Name)
			}

			assert.ElementsMatch(t, testCase.expectedNames, names)
		})
	}
}

func TestIsInventoryEligibleBMH(t *testing.T) {
	testCases := []struct {
		name     string
		bmh      *bmhv1alpha1.BareMetalHost
		expected bool
	}{
		{
			name:     "nil baremetalhost",
			bmh:      nil,
			expected: false,
		},
		{
			name:     "missing labels",
			bmh:      buildDummyBmHost(bmhv1alpha1.StateProvisioned, bmhv1alpha1.OperationalStatusOK),
			expected: false,
		},
		{
			name: "resource pool name with eligible state",
			bmh: buildInventoryEligibleBMHWithLabels("eligible-bmh", defaultBmHostNsName,
				bmhv1alpha1.StateProvisioned, map[string]string{
					labelResourcePoolName: "pool123",
				}),
			expected: true,
		},
		{
			name: "resource selector label with eligible state",
			bmh: buildInventoryEligibleBMHWithLabels("selector-bmh", defaultBmHostNsName,
				bmhv1alpha1.StateAvailable, map[string]string{
					labelPrefixResourceSelector + "zone": "east",
				}),
			expected: true,
		},
		{
			name: "resource pool name with ineligible state",
			bmh: buildInventoryEligibleBMHWithLabels("ineligible-bmh", defaultBmHostNsName,
				bmhv1alpha1.StateInspecting, map[string]string{
					labelResourcePoolName: "pool123",
				}),
			expected: false,
		},
		{
			name: "eligible deprovisioning state",
			bmh: buildInventoryEligibleBMHWithLabels("deprovisioning-bmh", defaultBmHostNsName,
				bmhv1alpha1.StateDeprovisioning, map[string]string{
					labelResourcePoolName: "pool123",
				}),
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, IsInventoryEligibleBMH(testCase.bmh))
		})
	}
}

// buildInventoryEligibleBMH returns a BareMetalHost with the resource pool label required for inventory eligibility.
func buildInventoryEligibleBMH(name, namespace string, state bmhv1alpha1.ProvisioningState) *bmhv1alpha1.BareMetalHost {
	return buildInventoryEligibleBMHWithLabels(name, namespace, state, map[string]string{
		labelResourcePoolName: "pool123",
	})
}

// buildInventoryEligibleBMHWithLabels returns a BareMetalHost with the provided inventory eligibility labels.
func buildInventoryEligibleBMHWithLabels(
	name, namespace string,
	state bmhv1alpha1.ProvisioningState,
	labels map[string]string,
) *bmhv1alpha1.BareMetalHost {
	bmh := buildDummyBmHost(state, bmhv1alpha1.OperationalStatusOK)
	bmh.Name = name
	bmh.Namespace = namespace
	bmh.Labels = labels

	return bmh
}
