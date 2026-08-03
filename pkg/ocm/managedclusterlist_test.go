package ocm

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ocm/clusterv1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testSpokeClusterName = "spoke-cluster"
	testHubClusterName   = "local-cluster"
)

func TestListManagedClusters(t *testing.T) {
	testCases := []struct {
		managedClusters []*ManagedClusterBuilder
		listOptions     []runtimeclient.ListOption
		client          bool
		expectedError   error
	}{
		{
			managedClusters: []*ManagedClusterBuilder{
				buildValidManagedClusterTestBuilder(buildTestClientWithDummyManagedCluster()),
			},
			listOptions:   nil,
			client:        true,
			expectedError: nil,
		},
		{
			managedClusters: []*ManagedClusterBuilder{
				buildValidManagedClusterTestBuilder(buildTestClientWithDummyManagedCluster()),
			},
			listOptions:   []runtimeclient.ListOption{&runtimeclient.ListOptions{LabelSelector: labels.NewSelector()}},
			client:        true,
			expectedError: nil,
		},
		{
			managedClusters: []*ManagedClusterBuilder{
				buildValidManagedClusterTestBuilder(buildTestClientWithDummyManagedCluster()),
			},
			listOptions:   nil,
			client:        false,
			expectedError: fmt.Errorf("failed to list managedClusters, 'apiClient' parameter is nil"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyManagedCluster()
		}

		builders, err := ListManagedClusters(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.managedClusters), len(builders))
		}
	}
}

func TestListNodeClusterEligibleManagedClusters(t *testing.T) {
	clusterID := uuid.New().String()
	templateID := uuid.New().String()

	testCases := []struct {
		name          string
		clusters      []*clusterv1.ManagedCluster
		expectedNames []string
	}{
		{
			name: "returns only node-cluster-eligible clusters",
			clusters: []*clusterv1.ManagedCluster{
				buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID),
				buildDummyManagedCluster("not-yet-available"),
			},
			expectedNames: []string{testSpokeClusterName},
		},
		{
			name: "includes hub cluster without template id label",
			clusters: []*clusterv1.ManagedCluster{
				buildNodeClusterEligibleHubManagedCluster(clusterID),
			},
			expectedNames: []string{testHubClusterName},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runtimeObjects := make([]runtime.Object, 0, len(testCase.clusters))
			for _, cluster := range testCase.clusters {
				runtimeObjects = append(runtimeObjects, cluster)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: clusterTestSchemes,
			})

			builders, err := ListNodeClusterEligibleManagedClusters(testSettings)
			assert.NoError(t, err)
			assert.Len(t, builders, len(testCase.expectedNames))

			names := make([]string, 0, len(builders))
			for _, builder := range builders {
				names = append(names, builder.Definition.Name)
			}

			assert.ElementsMatch(t, testCase.expectedNames, names)
		})
	}
}

func TestListDeploymentManagerEligibleManagedClusters(t *testing.T) {
	clusterID := uuid.New().String()
	templateID := uuid.New().String()

	testCases := []struct {
		name          string
		clusters      []*clusterv1.ManagedCluster
		expectedNames []string
	}{
		{
			name: "returns only deployment-manager-eligible clusters",
			clusters: []*clusterv1.ManagedCluster{
				buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID),
				buildDummyManagedCluster("not-yet-available"),
				func() *clusterv1.ManagedCluster {
					cluster := buildDeploymentManagerEligibleHubManagedCluster(clusterID, templateID)
					delete(cluster.Labels, clusterTemplateArtifactsLabel)

					return cluster
				}(),
			},
			expectedNames: []string{testSpokeClusterName},
		},
		{
			name: "includes cluster with non-uuid cluster id",
			clusters: []*clusterv1.ManagedCluster{
				buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, "not-a-uuid", templateID),
			},
			expectedNames: []string{testSpokeClusterName},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runtimeObjects := make([]runtime.Object, 0, len(testCase.clusters))
			for _, cluster := range testCase.clusters {
				runtimeObjects = append(runtimeObjects, cluster)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: clusterTestSchemes,
			})

			builders, err := ListDeploymentManagerEligibleManagedClusters(testSettings)
			assert.NoError(t, err)
			assert.Len(t, builders, len(testCase.expectedNames))

			names := make([]string, 0, len(builders))
			for _, builder := range builders {
				names = append(names, builder.Definition.Name)
			}

			assert.ElementsMatch(t, testCase.expectedNames, names)
		})
	}
}

//nolint:funlen // table-driven test with multiple eligibility scenarios.
func TestIsNodeClusterEligibleManagedCluster(t *testing.T) {
	clusterID := uuid.New().String()
	templateID := uuid.New().String()

	testCases := []struct {
		name     string
		cluster  *clusterv1.ManagedCluster
		eligible bool
	}{
		{
			name:     "eligible spoke cluster",
			cluster:  buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID),
			eligible: true,
		},
		{
			name:     "eligible hub cluster",
			cluster:  buildNodeClusterEligibleHubManagedCluster(clusterID),
			eligible: true,
		},
		{
			name: "non-openshift vendor without openshift version",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Labels[clusterVendorLabel] = "OtherVendor"
				delete(cluster.Labels, openshiftVersionLabel)

				return cluster
			}(),
			eligible: true,
		},
		{
			name: "available condition unknown",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Status.Conditions = []metav1.Condition{{
					Type:   clusterv1.ManagedClusterConditionAvailable,
					Status: metav1.ConditionUnknown,
				}}

				return cluster
			}(),
			eligible: true,
		},
		{
			name:     "invalid cluster id",
			cluster:  buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, "not-a-uuid", templateID),
			eligible: false,
		},
		{
			name: "missing cluster id label",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				delete(cluster.Labels, clusterIDLabel)

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "available condition false",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Status.Conditions = []metav1.Condition{{
					Type:   clusterv1.ManagedClusterConditionAvailable,
					Status: metav1.ConditionFalse,
				}}

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "spoke without template id",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				delete(cluster.Labels, clusterTemplateArtifactsLabel)

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "missing available condition",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Status.Conditions = nil

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "missing vendor label",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				delete(cluster.Labels, clusterVendorLabel)

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "openshift vendor without openshift version",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildNodeClusterEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				delete(cluster.Labels, openshiftVersionLabel)

				return cluster
			}(),
			eligible: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.eligible, IsNodeClusterEligibleManagedCluster(testCase.cluster))
		})
	}
}

//nolint:funlen // table-driven test with multiple eligibility scenarios.
func TestIsDeploymentManagerEligibleManagedCluster(t *testing.T) {
	clusterID := uuid.New().String()
	templateID := uuid.New().String()

	testCases := []struct {
		name     string
		cluster  *clusterv1.ManagedCluster
		eligible bool
	}{
		{
			name:     "eligible spoke cluster",
			cluster:  buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID),
			eligible: true,
		},
		{
			name:     "eligible hub cluster",
			cluster:  buildDeploymentManagerEligibleHubManagedCluster(clusterID, templateID),
			eligible: true,
		},
		{
			name:     "non-uuid cluster id",
			cluster:  buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, "not-a-uuid", templateID),
			eligible: true,
		},
		{
			name: "cluster without vendor label",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				delete(cluster.Labels, clusterVendorLabel)
				delete(cluster.Labels, openshiftVersionLabel)

				return cluster
			}(),
			eligible: true,
		},
		{
			name: "available condition unknown",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Status.Conditions = []metav1.Condition{{
					Type:   clusterv1.ManagedClusterConditionAvailable,
					Status: metav1.ConditionUnknown,
				}}

				return cluster
			}(),
			eligible: true,
		},
		{
			name: "hub cluster without template id",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildDeploymentManagerEligibleHubManagedCluster(clusterID, templateID)
				delete(cluster.Labels, clusterTemplateArtifactsLabel)

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "spoke without template id",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				delete(cluster.Labels, clusterTemplateArtifactsLabel)

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "missing client url",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Spec.ManagedClusterClientConfigs = nil

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "empty client url",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Spec.ManagedClusterClientConfigs = []clusterv1.ClientConfig{{URL: ""}}

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "missing cluster id label",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				delete(cluster.Labels, clusterIDLabel)

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "available condition false",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Status.Conditions = []metav1.Condition{{
					Type:   clusterv1.ManagedClusterConditionAvailable,
					Status: metav1.ConditionFalse,
				}}

				return cluster
			}(),
			eligible: false,
		},
		{
			name: "missing available condition",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildDeploymentManagerEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Status.Conditions = nil

				return cluster
			}(),
			eligible: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.eligible, IsDeploymentManagerEligibleManagedCluster(testCase.cluster))
		})
	}
}

// buildNodeClusterEligibleSpokeManagedCluster returns a spoke ManagedCluster that passes all NodeCluster eligibility
// checks.
func buildNodeClusterEligibleSpokeManagedCluster(name, clusterID, templateID string) *clusterv1.ManagedCluster {
	cluster := buildDummyManagedCluster(name)
	cluster.Labels = map[string]string{
		clusterVendorLabel:            openshiftVendor,
		clusterIDLabel:                clusterID,
		openshiftVersionLabel:         "4.20.0",
		clusterTemplateArtifactsLabel: templateID,
	}
	cluster.Status.Conditions = []metav1.Condition{{
		Type:   clusterv1.ManagedClusterConditionAvailable,
		Status: metav1.ConditionTrue,
	}}

	return cluster
}

// buildNodeClusterEligibleHubManagedCluster returns a hub ManagedCluster that passes all NodeCluster eligibility
// checks.
func buildNodeClusterEligibleHubManagedCluster(clusterID string) *clusterv1.ManagedCluster {
	cluster := buildNodeClusterEligibleSpokeManagedCluster(testHubClusterName, clusterID, uuid.New().String())
	cluster.Labels[localClusterLabel] = "true"
	delete(cluster.Labels, clusterTemplateArtifactsLabel)

	return cluster
}

// buildDeploymentManagerEligibleSpokeManagedCluster returns a spoke ManagedCluster that passes all DeploymentManager
// eligibility checks.
func buildDeploymentManagerEligibleSpokeManagedCluster(name, clusterID, templateID string) *clusterv1.ManagedCluster {
	cluster := buildNodeClusterEligibleSpokeManagedCluster(name, clusterID, templateID)
	cluster.Spec.ManagedClusterClientConfigs = []clusterv1.ClientConfig{{
		URL: "https://api.example.com:6443",
	}}

	return cluster
}

// buildDeploymentManagerEligibleHubManagedCluster returns a hub ManagedCluster that passes all DeploymentManager
// eligibility checks.
func buildDeploymentManagerEligibleHubManagedCluster(clusterID, templateID string) *clusterv1.ManagedCluster {
	cluster := buildDeploymentManagerEligibleSpokeManagedCluster(testHubClusterName, clusterID, templateID)
	cluster.Labels[localClusterLabel] = "true"

	return cluster
}
