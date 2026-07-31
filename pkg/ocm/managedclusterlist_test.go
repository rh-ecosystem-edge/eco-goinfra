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

func TestListORANEligibleManagedClusters(t *testing.T) {
	clusterID := uuid.New().String()
	templateID := uuid.New().String()

	testCases := []struct {
		name          string
		clusters      []*clusterv1.ManagedCluster
		expectedNames []string
	}{
		{
			name: "returns only ORAN-eligible clusters",
			clusters: []*clusterv1.ManagedCluster{
				buildORANEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID),
				buildDummyManagedCluster("not-yet-available"),
			},
			expectedNames: []string{testSpokeClusterName},
		},
		{
			name: "includes hub cluster without template id label",
			clusters: []*clusterv1.ManagedCluster{
				buildORANEligibleHubManagedCluster(testHubClusterName, clusterID),
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

			builders, err := ListORANEligibleManagedClusters(testSettings)
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

func TestIsORANEligibleManagedCluster(t *testing.T) {
	clusterID := uuid.New().String()
	templateID := uuid.New().String()

	testCases := []struct {
		name    string
		cluster *clusterv1.ManagedCluster
	}{
		{
			name:    "eligible spoke cluster",
			cluster: buildORANEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID),
		},
		{
			name:    "eligible hub cluster",
			cluster: buildORANEligibleHubManagedCluster(testHubClusterName, clusterID),
		},
		{
			name: "non-openshift vendor without openshift version",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildORANEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Labels[clusterVendorLabel] = "OtherVendor"
				delete(cluster.Labels, openshiftVersionLabel)

				return cluster
			}(),
		},
		{
			name: "available condition unknown",
			cluster: func() *clusterv1.ManagedCluster {
				cluster := buildORANEligibleSpokeManagedCluster(testSpokeClusterName, clusterID, templateID)
				cluster.Status.Conditions = []metav1.Condition{{
					Type:   clusterv1.ManagedClusterConditionAvailable,
					Status: metav1.ConditionUnknown,
				}}

				return cluster
			}(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, IsORANEligibleManagedCluster(testCase.cluster))
		})
	}
}

// buildORANEligibleSpokeManagedCluster returns a spoke ManagedCluster that passes all ORAN eligibility checks.
func buildORANEligibleSpokeManagedCluster(name, clusterID, templateID string) *clusterv1.ManagedCluster {
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

// buildORANEligibleHubManagedCluster returns a hub ManagedCluster that passes all ORAN eligibility checks.
func buildORANEligibleHubManagedCluster(name, clusterID string) *clusterv1.ManagedCluster {
	cluster := buildORANEligibleSpokeManagedCluster(name, clusterID, uuid.New().String())
	cluster.Labels[localClusterLabel] = "true"
	delete(cluster.Labels, clusterTemplateArtifactsLabel)

	return cluster
}
