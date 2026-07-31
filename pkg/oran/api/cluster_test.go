package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/fields"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/filter"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/cluster"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/common"
	"github.com/stretchr/testify/assert"
)

var (
	// dummyNodeClusterType is a test node cluster type for use in tests.
	dummyNodeClusterType = NodeClusterType{
		NodeClusterTypeId: uuid.New(),
		Name:              "test-node-cluster-type",
		Description:       "test node cluster type",
	}

	// dummyNodeCluster is a test node cluster for use in tests.
	dummyNodeCluster = NodeCluster{
		NodeClusterId:                  uuid.New(),
		ClientNodeClusterId:            uuid.New(),
		Name:                           "test-cluster",
		Description:                    "test cluster",
		ClusterDistributionDescription: "distributed",
		NodeClusterTypeId:              uuid.New(),
		ArtifactResourceId:             uuid.New(),
		ClusterResourceIds:             []uuid.UUID{uuid.New()},
	}

	// dummyClusterResourceType is a test cluster resource type for use in tests.
	dummyClusterResourceType = ClusterResourceType{
		ClusterResourceTypeId: uuid.New(),
		Name:                  "test-cluster-resource-type",
		Description:           "test cluster resource type",
	}

	// dummyClusterResource is a test cluster resource for use in tests.
	dummyClusterResource = ClusterResource{
		ClusterResourceId:     uuid.New(),
		Name:                  "test-cluster-resource",
		Description:           "test cluster resource",
		ClusterResourceTypeId: uuid.New(),
		ResourceId:            uuid.New(),
	}

	// dummyClusterSubscription is a test cluster subscription for use in tests.
	dummyClusterSubscription = ClusterSubscription{
		SubscriptionId:         new(uuid.New()),
		Callback:               "https://smo.example.com/smo/v1/ocloud_inventory_observer",
		ConsumerSubscriptionId: new(uuid.New()),
	}

	// dummyAlarmDictionary is a test alarm dictionary for use in tests.
	dummyAlarmDictionary = AlarmDictionary{
		AlarmDictionaryId:            uuid.New(),
		AlarmDictionarySchemaVersion: "1.0",
		AlarmDictionaryVersion:       "1.0",
		EntityType:                   "NodeClusterType",
		ManagementInterfaceId:        []common.AlarmDictionaryManagementInterfaceId{common.AlarmDictionaryManagementInterfaceIdO2IMS},
		PkNotificationField:          []string{"alarmDefinitionId"},
		Vendor:                       "Red Hat",
		AlarmDefinition:              []common.AlarmDefinition{},
	}

	// dummyAPIVersions is a test API versions response for use in tests.
	dummyAPIVersions = APIVersions{
		UriPrefix: new("/o2ims-infrastructureCluster"),
	}

	// defaultNodeClusterTypeID is the ID from dummyNodeClusterType for use in tests.
	defaultNodeClusterTypeID = dummyNodeClusterType.NodeClusterTypeId
	// defaultNodeClusterID is the ID from dummyNodeCluster for use in tests.
	defaultNodeClusterID = dummyNodeCluster.NodeClusterId
	// defaultClusterResourceTypeID is the ID from dummyClusterResourceType for use in tests.
	defaultClusterResourceTypeID = dummyClusterResourceType.ClusterResourceTypeId
	// defaultClusterResourceID is the ID from dummyClusterResource for use in tests.
	defaultClusterResourceID = dummyClusterResource.ClusterResourceId
	// defaultClusterSubscriptionID is the ID from dummyClusterSubscription for use in tests.
	defaultClusterSubscriptionID = *dummyClusterSubscription.SubscriptionId
	// defaultAlarmDictionaryID is the ID from dummyAlarmDictionary for use in tests.
	defaultAlarmDictionaryID = dummyAlarmDictionary.AlarmDictionaryId
)

func TestGetAllVersions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			handler: jsonResponseHandler(dummyAPIVersions),
		},
		{
			name:          "server error 500",
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get all API versions: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.GetAllVersions()

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAPIVersions.UriPrefix, result.UriPrefix)
			validateHTTPRequest(t, capturedRequest, "GET", "/o2ims-infrastructureCluster/api_versions", nil)
		})
	}
}

func TestGetMinorVersions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			handler: jsonResponseHandler(dummyAPIVersions),
		},
		{
			name:          "server error 500",
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get minor API versions: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.GetMinorVersions()

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAPIVersions.UriPrefix, result.UriPrefix)
			validateHTTPRequest(t, capturedRequest, "GET", "/o2ims-infrastructureCluster/v1/api_versions", nil)
		})
	}
}

func TestListNodeClusterTypes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		opts                []ListOption
		handler             http.HandlerFunc
		expectedError       string
		expectedQueryParams map[string]string
	}{
		{
			name:                "success without options",
			opts:                nil,
			handler:             jsonResponseHandler([]cluster.NodeClusterType{dummyNodeClusterType}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "test-node-cluster-type")),
			},
			handler: jsonResponseHandler([]cluster.NodeClusterType{dummyNodeClusterType}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,test-node-cluster-type)",
			},
		},
		{
			name: "success with fields",
			opts: []ListOption{
				WithFields(fields.Include("name", "nodeClusterTypeId")),
			},
			handler: jsonResponseHandler([]cluster.NodeClusterType{dummyNodeClusterType}),
			expectedQueryParams: map[string]string{
				"fields": "name,nodeClusterTypeId",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list node cluster types: received error from api:",
			expectedQueryParams: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.ListNodeClusterTypes(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyNodeClusterType.NodeClusterTypeId, result[0].NodeClusterTypeId)

			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureCluster/v1/nodeClusterTypes", testCase.expectedQueryParams)
		})
	}
}

func TestGetNodeClusterType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		typeID        uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			typeID:  defaultNodeClusterTypeID,
			handler: jsonResponseHandler(dummyNodeClusterType, http.StatusOK),
		},
		{
			name:          "server error 500",
			typeID:        defaultNodeClusterTypeID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get node cluster type: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.GetNodeClusterType(testCase.typeID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyNodeClusterType.NodeClusterTypeId, result.NodeClusterTypeId)
			assert.Equal(t, dummyNodeClusterType.Name, result.Name)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureCluster/v1/nodeClusterTypes/%s", testCase.typeID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestGetNodeClusterTypeAlarmDictionary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		typeID        uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			typeID:  defaultNodeClusterTypeID,
			handler: jsonResponseHandler(dummyAlarmDictionary, http.StatusOK),
		},
		{
			name:          "server error 500",
			typeID:        defaultNodeClusterTypeID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get node cluster type alarm dictionary: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.GetNodeClusterTypeAlarmDictionary(testCase.typeID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmDictionary.AlarmDictionaryId, result.AlarmDictionaryId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureCluster/v1/nodeClusterTypes/%s/alarmDictionary", testCase.typeID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func nodeClustersListTestCases() []struct {
	name                string
	opts                []ListOption
	handler             http.HandlerFunc
	expectedError       string
	expectedQueryParams map[string]string
} {
	return []struct {
		name                string
		opts                []ListOption
		handler             http.HandlerFunc
		expectedError       string
		expectedQueryParams map[string]string
	}{
		{
			name:                "success without options",
			opts:                nil,
			handler:             jsonResponseHandler([]cluster.NodeCluster{dummyNodeCluster}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "test-cluster")),
			},
			handler: jsonResponseHandler([]cluster.NodeCluster{dummyNodeCluster}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,test-cluster)",
			},
		},
		{
			name: "success with fields",
			opts: []ListOption{
				WithFields(fields.Include("name", "nodeClusterId")),
			},
			handler: jsonResponseHandler([]cluster.NodeCluster{dummyNodeCluster}),
			expectedQueryParams: map[string]string{
				"fields": "name,nodeClusterId",
			},
		},
		{
			name: "success with exclude fields",
			opts: []ListOption{
				WithFields(fields.Exclude(fields.Path("extensions", "country"))),
			},
			handler: jsonResponseHandler([]cluster.NodeCluster{dummyNodeCluster}),
			expectedQueryParams: map[string]string{
				"exclude_fields": "extensions/country",
			},
		},
		{
			name: "success with all fields",
			opts: []ListOption{
				WithFields(fields.All()),
			},
			handler: jsonResponseHandler([]cluster.NodeCluster{dummyNodeCluster}),
			expectedQueryParams: map[string]string{
				"all_fields": "",
			},
		},
		{
			name: "success with filter and fields",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "test-cluster")),
				WithFields(fields.Include("name")),
			},
			handler: jsonResponseHandler([]cluster.NodeCluster{dummyNodeCluster}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,test-cluster)",
				"fields": "name",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list node clusters: received error from api:",
			expectedQueryParams: nil,
		},
	}
}

func TestListNodeClusters(t *testing.T) {
	t.Parallel()

	for _, testCase := range nodeClustersListTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.ListNodeClusters(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyNodeCluster.NodeClusterId, result[0].NodeClusterId)

			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureCluster/v1/nodeClusters", testCase.expectedQueryParams)
		})
	}
}

func TestGetNodeCluster(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		clusterID     uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:      "success",
			clusterID: defaultNodeClusterID,
			handler:   jsonResponseHandler(dummyNodeCluster, http.StatusOK),
		},
		{
			name:          "server error 500",
			clusterID:     defaultNodeClusterID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get node cluster: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.GetNodeCluster(testCase.clusterID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyNodeCluster.NodeClusterId, result.NodeClusterId)
			assert.Equal(t, dummyNodeCluster.Name, result.Name)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureCluster/v1/nodeClusters/%s", testCase.clusterID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestListClusterResourceTypes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		opts                []ListOption
		handler             http.HandlerFunc
		expectedError       string
		expectedQueryParams map[string]string
	}{
		{
			name:                "success without options",
			opts:                nil,
			handler:             jsonResponseHandler([]cluster.ClusterResourceType{dummyClusterResourceType}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "test-cluster-resource-type")),
			},
			handler: jsonResponseHandler([]cluster.ClusterResourceType{dummyClusterResourceType}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,test-cluster-resource-type)",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list cluster resource types: received error from api:",
			expectedQueryParams: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.ListClusterResourceTypes(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyClusterResourceType.ClusterResourceTypeId, result[0].ClusterResourceTypeId)

			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureCluster/v1/clusterResourceTypes", testCase.expectedQueryParams)
		})
	}
}

func TestGetClusterResourceType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		typeID        uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			typeID:  defaultClusterResourceTypeID,
			handler: jsonResponseHandler(dummyClusterResourceType, http.StatusOK),
		},
		{
			name:          "server error 500",
			typeID:        defaultClusterResourceTypeID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get cluster resource type: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.GetClusterResourceType(testCase.typeID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyClusterResourceType.ClusterResourceTypeId, result.ClusterResourceTypeId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureCluster/v1/clusterResourceTypes/%s", testCase.typeID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestListClusterResources(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		opts                []ListOption
		handler             http.HandlerFunc
		expectedError       string
		expectedQueryParams map[string]string
	}{
		{
			name:                "success without options",
			opts:                nil,
			handler:             jsonResponseHandler([]cluster.ClusterResource{dummyClusterResource}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "test-cluster-resource")),
			},
			handler: jsonResponseHandler([]cluster.ClusterResource{dummyClusterResource}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,test-cluster-resource)",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list cluster resources: received error from api:",
			expectedQueryParams: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.ListClusterResources(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyClusterResource.ClusterResourceId, result[0].ClusterResourceId)

			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureCluster/v1/clusterResources", testCase.expectedQueryParams)
		})
	}
}

func TestGetClusterResource(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		resourceID    uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:       "success",
			resourceID: defaultClusterResourceID,
			handler:    jsonResponseHandler(dummyClusterResource, http.StatusOK),
		},
		{
			name:          "server error 500",
			resourceID:    defaultClusterResourceID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get cluster resource: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.GetClusterResource(testCase.resourceID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyClusterResource.ClusterResourceId, result.ClusterResourceId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureCluster/v1/clusterResources/%s", testCase.resourceID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestListClusterSubscriptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		opts                []ListOption
		handler             http.HandlerFunc
		expectedError       string
		expectedQueryParams map[string]string
	}{
		{
			name:                "success without options",
			opts:                nil,
			handler:             jsonResponseHandler([]cluster.Subscription{dummyClusterSubscription}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("callback", dummyClusterSubscription.Callback)),
			},
			handler: jsonResponseHandler([]cluster.Subscription{dummyClusterSubscription}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,callback,https://smo.example.com/smo/v1/ocloud_inventory_observer)",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list cluster subscriptions: received error from api:",
			expectedQueryParams: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.ListClusterSubscriptions(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyClusterSubscription.SubscriptionId, result[0].SubscriptionId)

			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureCluster/v1/subscriptions", testCase.expectedQueryParams)
		})
	}
}

func TestCreateClusterSubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		subscription  ClusterSubscription
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:         "success",
			subscription: dummyClusterSubscription,
			handler:      jsonResponseHandler(dummyClusterSubscription, http.StatusCreated),
		},
		{
			name:          "server error 500",
			subscription:  dummyClusterSubscription,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to create cluster subscription: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.CreateClusterSubscription(testCase.subscription)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyClusterSubscription.SubscriptionId, result.SubscriptionId)
			assert.Equal(t, dummyClusterSubscription.Callback, result.Callback)

			validateHTTPRequest(
				t, capturedRequest, "POST", "/o2ims-infrastructureCluster/v1/subscriptions", nil, "application/json")
		})
	}
}

func TestGetClusterSubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID uuid.UUID
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: defaultClusterSubscriptionID,
			handler:        jsonResponseHandler(dummyClusterSubscription, http.StatusOK),
		},
		{
			name:           "server error 500",
			subscriptionID: defaultClusterSubscriptionID,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to get cluster subscription: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.GetClusterSubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyClusterSubscription.SubscriptionId, result.SubscriptionId)
			assert.Equal(t, dummyClusterSubscription.Callback, result.Callback)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureCluster/v1/subscriptions/%s", testCase.subscriptionID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestDeleteClusterSubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID uuid.UUID
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: defaultClusterSubscriptionID,
			handler:        jsonResponseHandler(nil, http.StatusOK),
		},
		{
			name:           "server error 500",
			subscriptionID: defaultClusterSubscriptionID,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to delete cluster subscription: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			err = clusterClient.DeleteClusterSubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureCluster/v1/subscriptions/%s", testCase.subscriptionID.String())
			validateHTTPRequest(t, capturedRequest, "DELETE", expectedPath, nil)
		})
	}
}

func TestListAlarmDictionaries(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		opts                []ListOption
		handler             http.HandlerFunc
		expectedError       string
		expectedQueryParams map[string]string
	}{
		{
			name:                "success without options",
			opts:                nil,
			handler:             jsonResponseHandler([]common.AlarmDictionary{dummyAlarmDictionary}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("entityType", "NodeClusterType")),
			},
			handler: jsonResponseHandler([]common.AlarmDictionary{dummyAlarmDictionary}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,entityType,NodeClusterType)",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list alarm dictionaries: received error from api:",
			expectedQueryParams: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.ListAlarmDictionaries(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmDictionary.AlarmDictionaryId, result[0].AlarmDictionaryId)

			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureCluster/v1/alarmDictionaries", testCase.expectedQueryParams)
		})
	}
}

func TestGetAlarmDictionary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		dictionaryID  uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:         "success",
			dictionaryID: defaultAlarmDictionaryID,
			handler:      jsonResponseHandler(dummyAlarmDictionary, http.StatusOK),
		},
		{
			name:          "server error 500",
			dictionaryID:  defaultAlarmDictionaryID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get alarm dictionary: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := cluster.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			result, err := clusterClient.GetAlarmDictionary(testCase.dictionaryID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmDictionary.AlarmDictionaryId, result.AlarmDictionaryId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureCluster/v1/alarmDictionaries/%s", testCase.dictionaryID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestBuildCluster(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(jsonResponseHandler([]cluster.NodeCluster{dummyNodeCluster}))
	defer server.Close()

	clusterClient, err := NewClientBuilder(server.URL).BuildCluster()

	assert.NoError(t, err)
	assert.NotNil(t, clusterClient)

	result, err := clusterClient.ListNodeClusters()
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

//nolint:funlen // Since this is only long because of the number of functions, we can ignore the length.
func TestClusterNetworkError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		testFunc func(client *ClusterClient) error
	}{
		{
			name: "GetAllVersions network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.GetAllVersions()

				return err
			},
		},
		{
			name: "GetMinorVersions network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.GetMinorVersions()

				return err
			},
		},
		{
			name: "ListNodeClusterTypes network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.ListNodeClusterTypes()

				return err
			},
		},
		{
			name: "GetNodeClusterType network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.GetNodeClusterType(defaultNodeClusterTypeID)

				return err
			},
		},
		{
			name: "GetNodeClusterTypeAlarmDictionary network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.GetNodeClusterTypeAlarmDictionary(defaultNodeClusterTypeID)

				return err
			},
		},
		{
			name: "ListNodeClusters network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.ListNodeClusters()

				return err
			},
		},
		{
			name: "GetNodeCluster network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.GetNodeCluster(defaultNodeClusterID)

				return err
			},
		},
		{
			name: "ListClusterResourceTypes network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.ListClusterResourceTypes()

				return err
			},
		},
		{
			name: "GetClusterResourceType network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.GetClusterResourceType(defaultClusterResourceTypeID)

				return err
			},
		},
		{
			name: "ListClusterResources network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.ListClusterResources()

				return err
			},
		},
		{
			name: "GetClusterResource network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.GetClusterResource(defaultClusterResourceID)

				return err
			},
		},
		{
			name: "ListClusterSubscriptions network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.ListClusterSubscriptions()

				return err
			},
		},
		{
			name: "CreateClusterSubscription network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.CreateClusterSubscription(dummyClusterSubscription)

				return err
			},
		},
		{
			name: "GetClusterSubscription network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.GetClusterSubscription(defaultClusterSubscriptionID)

				return err
			},
		},
		{
			name: "DeleteClusterSubscription network error",
			testFunc: func(client *ClusterClient) error {
				return client.DeleteClusterSubscription(defaultClusterSubscriptionID)
			},
		},
		{
			name: "ListAlarmDictionaries network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.ListAlarmDictionaries()

				return err
			},
		},
		{
			name: "GetAlarmDictionary network error",
			testFunc: func(client *ClusterClient) error {
				_, err := client.GetAlarmDictionary(defaultAlarmDictionaryID)

				return err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// 192.0.2.0 is a reserved test address so we never accidentally use a valid IP. Still, we set a
			// timeout to ensure that we do not timeout the test.
			client, err := cluster.NewClientWithResponses("http://192.0.2.0:8080",
				cluster.WithHTTPClient(&http.Client{Timeout: time.Second * 1}))
			assert.NoError(t, err)

			clusterClient := &ClusterClient{ClientWithResponsesInterface: client}
			err = testCase.testFunc(clusterClient)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error contacting api")
		})
	}
}
