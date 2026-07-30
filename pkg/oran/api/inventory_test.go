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
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/inventory"
	"github.com/stretchr/testify/assert"
)

var (
	dummyDeploymentManager = DeploymentManager{
		DeploymentManagerId: uuid.New(),
		Name:                "test-deployment-manager",
		Description:         "test deployment manager",
		OCloudId:            uuid.New(),
		ServiceUri:          "https://my.cluster:6443",
		SupportedLocations:  []string{"location-east-1"},
		Capabilities:        map[string]interface{}{},
		Capacity:            map[string]interface{}{},
	}

	dummyInventorySubscription = InventorySubscription{
		SubscriptionId:         new(uuid.New()),
		Callback:               "https://smo.example.com/smo/v1/ocloud_inventory_observer",
		ConsumerSubscriptionId: new(uuid.New()),
	}

	dummyOCloudInfo = OCloudInfo{
		OCloudId:      uuid.New(),
		GlobalCloudId: uuid.New(),
		Name:          "my-cloud",
		Description:   "My cloud",
		ServiceUri:    "http://localhost:8000",
		Extensions:    map[string]interface{}{},
	}

	dummyResourceType = ResourceType{
		ResourceTypeId: uuid.New(),
		Name:           "medium",
		Description:    "test resource type",
		Vendor:         "Red Hat",
		Model:          "X-1000",
		Version:        "v2024.11",
		ResourceKind:   inventory.ResourceTypeResourceKindPHYSICAL,
		ResourceClass:  inventory.ResourceTypeResourceClassCOMPUTE,
		Extensions:     map[string]interface{}{},
	}

	dummyResourcePool = ResourcePool{
		ResourcePoolId: uuid.New(),
		Name:           "my-cluster",
		OCloudSiteId:   uuid.New(),
		Description:    "test resource pool",
	}

	dummyResource = Resource{
		ResourceId:     uuid.New(),
		ResourcePoolId: uuid.New(),
		ResourceTypeId: uuid.New(),
		GlobalAssetId:  "asset-123",
		Description:    "my-node",
		Elements:       []Resource{},
		Tags:           []string{},
		Groups:         []string{},
		Extensions:     map[string]interface{}{},
	}

	dummyLocationInfo = LocationInfo{
		GlobalLocationId: "location-east-1",
		Name:             "test location",
		Description:      "test location description",
		OCloudSiteIds:    []uuid.UUID{uuid.New()},
		Extensions:       map[string]interface{}{},
	}

	dummyOCloudSiteInfo = OCloudSiteInfo{
		OCloudSiteId:     uuid.New(),
		GlobalLocationId: "location-east-1",
		Name:             "test ocloud site",
		Description:      "test ocloud site description",
		ResourcePools:    []uuid.UUID{uuid.New()},
		Extensions:       map[string]interface{}{},
	}

	dummyInventoryAlarmDictionary = AlarmDictionary{
		AlarmDictionaryId:            uuid.New(),
		AlarmDictionarySchemaVersion: "1.0",
		AlarmDictionaryVersion:       "1.0",
		EntityType:                   "ResourceType",
		ManagementInterfaceId:        []common.AlarmDictionaryManagementInterfaceId{common.AlarmDictionaryManagementInterfaceIdO2IMS},
		PkNotificationField:          []string{"alarmDefinitionId"},
		Vendor:                       "Red Hat",
		AlarmDefinition:              []common.AlarmDefinition{},
	}

	dummyInventoryAllAPIVersions = APIVersions{
		UriPrefix: new("/o2ims-infrastructureInventory"),
	}

	dummyInventoryMinorAPIVersions = APIVersions{
		UriPrefix: new("/o2ims-infrastructureInventory/v2"),
	}

	defaultDeploymentManagerID     = dummyDeploymentManager.DeploymentManagerId
	defaultInventorySubscriptionID = *dummyInventorySubscription.SubscriptionId
	defaultResourceTypeID          = dummyResourceType.ResourceTypeId
	defaultResourcePoolID          = dummyResourcePool.ResourcePoolId
	defaultResourceID              = dummyResource.ResourceId
	defaultInventoryAlarmDictID    = dummyInventoryAlarmDictionary.AlarmDictionaryId
	defaultGlobalLocationID        = dummyLocationInfo.GlobalLocationId
	defaultOCloudSiteID            = dummyOCloudSiteInfo.OCloudSiteId
)

func TestInventoryGetAllVersions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			handler: jsonResponseHandler(dummyInventoryAllAPIVersions),
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetAllVersions()

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyInventoryAllAPIVersions.UriPrefix, result.UriPrefix)
			validateHTTPRequest(t, capturedRequest, "GET", "/o2ims-infrastructureInventory/api_versions", nil)
		})
	}
}

func TestInventoryGetCloudInfo(t *testing.T) {
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
			handler:             jsonResponseHandler(dummyOCloudInfo),
			expectedQueryParams: nil,
		},
		{
			name: "success with fields",
			opts: []ListOption{
				WithFields(fields.Include("name", "oCloudId")),
			},
			handler: jsonResponseHandler(dummyOCloudInfo),
			expectedQueryParams: map[string]string{
				"fields": "name,oCloudId",
			},
		},
		{
			name:          "server error 500",
			opts:          nil,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get O-Cloud info: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetCloudInfo(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyOCloudInfo.Name, result.Name)
			validateHTTPRequest(t, capturedRequest, "GET", "/o2ims-infrastructureInventory/v2", testCase.expectedQueryParams)
		})
	}
}

func TestInventoryGetMinorVersions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			handler: jsonResponseHandler(dummyInventoryMinorAPIVersions),
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetMinorVersions()

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyInventoryMinorAPIVersions.UriPrefix, result.UriPrefix)
			validateHTTPRequest(t, capturedRequest, "GET", "/o2ims-infrastructureInventory/v2/api_versions", nil)
		})
	}
}

func resourceTypesListTestCases() []struct {
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
			handler:             jsonResponseHandler([]inventory.ResourceType{dummyResourceType}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "medium")),
			},
			handler: jsonResponseHandler([]inventory.ResourceType{dummyResourceType}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,medium)",
			},
		},
		{
			name: "success with fields",
			opts: []ListOption{
				WithFields(fields.Include("name", "resourceTypeId")),
			},
			handler: jsonResponseHandler([]inventory.ResourceType{dummyResourceType}),
			expectedQueryParams: map[string]string{
				"fields": "name,resourceTypeId",
			},
		},
		{
			name: "success with exclude fields",
			opts: []ListOption{
				WithFields(fields.Exclude(fields.Path("extensions", "country"))),
			},
			handler: jsonResponseHandler([]inventory.ResourceType{dummyResourceType}),
			expectedQueryParams: map[string]string{
				"exclude_fields": "extensions/country",
			},
		},
		{
			name: "success with all fields",
			opts: []ListOption{
				WithFields(fields.All()),
			},
			handler: jsonResponseHandler([]inventory.ResourceType{dummyResourceType}),
			expectedQueryParams: map[string]string{
				"all_fields": "",
			},
		},
		{
			name: "success with filter and fields",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "medium")),
				WithFields(fields.Include("name")),
			},
			handler: jsonResponseHandler([]inventory.ResourceType{dummyResourceType}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,medium)",
				"fields": "name",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list resource types: received error from api:",
			expectedQueryParams: nil,
		},
	}
}

func TestInventoryListDeploymentManagers(t *testing.T) {
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
			handler:             jsonResponseHandler([]inventory.DeploymentManager{dummyDeploymentManager}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "test-deployment-manager")),
			},
			handler: jsonResponseHandler([]inventory.DeploymentManager{dummyDeploymentManager}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,test-deployment-manager)",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list deployment managers: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.ListDeploymentManagers(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyDeploymentManager.DeploymentManagerId, result[0].DeploymentManagerId)
			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureInventory/v2/deploymentManagers", testCase.expectedQueryParams)
		})
	}
}

func TestInventoryGetDeploymentManager(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		managerID     uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:      "success",
			managerID: defaultDeploymentManagerID,
			handler:   jsonResponseHandler(dummyDeploymentManager, http.StatusOK),
		},
		{
			name:          "server error 500",
			managerID:     defaultDeploymentManagerID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get deployment manager: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetDeploymentManager(testCase.managerID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyDeploymentManager.DeploymentManagerId, result.DeploymentManagerId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/deploymentManagers/%s", testCase.managerID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestInventoryListResourceTypes(t *testing.T) {
	t.Parallel()

	for _, testCase := range resourceTypesListTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
			defer server.Close()

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.ListResourceTypes(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyResourceType.ResourceTypeId, result[0].ResourceTypeId)
			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureInventory/v2/resourceTypes", testCase.expectedQueryParams)
		})
	}
}

func TestInventoryGetResourceType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		typeID        uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			typeID:  defaultResourceTypeID,
			handler: jsonResponseHandler(dummyResourceType, http.StatusOK),
		},
		{
			name:          "server error 500",
			typeID:        defaultResourceTypeID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get resource type: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetResourceType(testCase.typeID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyResourceType.ResourceTypeId, result.ResourceTypeId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/resourceTypes/%s", testCase.typeID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestInventoryGetResourceTypeAlarmDictionary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		typeID        uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			typeID:  defaultResourceTypeID,
			handler: jsonResponseHandler(dummyInventoryAlarmDictionary, http.StatusOK),
		},
		{
			name:          "server error 500",
			typeID:        defaultResourceTypeID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get resource type alarm dictionary: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetResourceTypeAlarmDictionary(testCase.typeID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyInventoryAlarmDictionary.AlarmDictionaryId, result.AlarmDictionaryId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/resourceTypes/%s/alarmDictionary", testCase.typeID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestInventoryListResourcePools(t *testing.T) {
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
			handler:             jsonResponseHandler([]inventory.ResourcePool{dummyResourcePool}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "my-cluster")),
			},
			handler: jsonResponseHandler([]inventory.ResourcePool{dummyResourcePool}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,my-cluster)",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list resource pools: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.ListResourcePools(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyResourcePool.ResourcePoolId, result[0].ResourcePoolId)
			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureInventory/v2/resourcePools", testCase.expectedQueryParams)
		})
	}
}

func TestInventoryGetResourcePool(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		poolID        uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			poolID:  defaultResourcePoolID,
			handler: jsonResponseHandler(dummyResourcePool, http.StatusOK),
		},
		{
			name:          "server error 500",
			poolID:        defaultResourcePoolID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get resource pool: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetResourcePool(testCase.poolID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyResourcePool.ResourcePoolId, result.ResourcePoolId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/resourcePools/%s", testCase.poolID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestInventoryListResources(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		poolID              uuid.UUID
		opts                []ListOption
		handler             http.HandlerFunc
		expectedError       string
		expectedQueryParams map[string]string
	}{
		{
			name:                "success without options",
			poolID:              defaultResourcePoolID,
			opts:                nil,
			handler:             jsonResponseHandler([]inventory.Resource{dummyResource}),
			expectedQueryParams: nil,
		},
		{
			name:   "success with filter",
			poolID: defaultResourcePoolID,
			opts: []ListOption{
				WithFilter(filter.Equals("description", "my-node")),
			},
			handler: jsonResponseHandler([]inventory.Resource{dummyResource}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,description,my-node)",
			},
		},
		{
			name:                "server error 500",
			poolID:              defaultResourcePoolID,
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list resources: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.ListResources(testCase.poolID, testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyResource.ResourceId, result[0].ResourceId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/resourcePools/%s/resources", testCase.poolID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, testCase.expectedQueryParams)
		})
	}
}

func TestInventoryGetResource(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		poolID        uuid.UUID
		resourceID    uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:       "success",
			poolID:     defaultResourcePoolID,
			resourceID: defaultResourceID,
			handler:    jsonResponseHandler(dummyResource, http.StatusOK),
		},
		{
			name:          "server error 500",
			poolID:        defaultResourcePoolID,
			resourceID:    defaultResourceID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get resource: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetResource(testCase.poolID, testCase.resourceID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyResource.ResourceId, result.ResourceId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/resourcePools/%s/resources/%s",
				testCase.poolID.String(), testCase.resourceID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestInventoryListInventorySubscriptions(t *testing.T) {
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
			handler:             jsonResponseHandler([]inventory.Subscription{dummyInventorySubscription}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("callback", dummyInventorySubscription.Callback)),
			},
			handler: jsonResponseHandler([]inventory.Subscription{dummyInventorySubscription}),
			expectedQueryParams: map[string]string{
				"filter": fmt.Sprintf("(eq,callback,%s)", dummyInventorySubscription.Callback),
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list inventory subscriptions: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.ListInventorySubscriptions(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyInventorySubscription.SubscriptionId, result[0].SubscriptionId)
			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureInventory/v2/subscriptions", testCase.expectedQueryParams)
		})
	}
}

func TestInventoryCreateInventorySubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		subscription  InventorySubscription
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:         "success",
			subscription: dummyInventorySubscription,
			handler:      jsonResponseHandler(dummyInventorySubscription, http.StatusCreated),
		},
		{
			name:          "server error 500",
			subscription:  dummyInventorySubscription,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to create inventory subscription: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.CreateInventorySubscription(testCase.subscription)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyInventorySubscription.SubscriptionId, result.SubscriptionId)
			assert.Equal(t, dummyInventorySubscription.Callback, result.Callback)
			validateHTTPRequest(
				t, capturedRequest, "POST", "/o2ims-infrastructureInventory/v2/subscriptions", nil, "application/json")
		})
	}
}

func TestInventoryGetInventorySubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID uuid.UUID
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: defaultInventorySubscriptionID,
			handler:        jsonResponseHandler(dummyInventorySubscription, http.StatusOK),
		},
		{
			name:           "server error 500",
			subscriptionID: defaultInventorySubscriptionID,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to get inventory subscription: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetInventorySubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyInventorySubscription.SubscriptionId, result.SubscriptionId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/subscriptions/%s", testCase.subscriptionID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestInventoryDeleteInventorySubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID uuid.UUID
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: defaultInventorySubscriptionID,
			handler:        jsonResponseHandler(nil, http.StatusOK),
		},
		{
			name:           "server error 500",
			subscriptionID: defaultInventorySubscriptionID,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to delete inventory subscription: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			err = inventoryClient.DeleteInventorySubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/subscriptions/%s", testCase.subscriptionID.String())
			validateHTTPRequest(t, capturedRequest, "DELETE", expectedPath, nil)
		})
	}
}

func TestInventoryListInventoryAlarmDictionaries(t *testing.T) {
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
			handler:             jsonResponseHandler([]common.AlarmDictionary{dummyInventoryAlarmDictionary}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("entityType", "ResourceType")),
			},
			handler: jsonResponseHandler([]common.AlarmDictionary{dummyInventoryAlarmDictionary}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,entityType,ResourceType)",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list inventory alarm dictionaries: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.ListInventoryAlarmDictionaries(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyInventoryAlarmDictionary.AlarmDictionaryId, result[0].AlarmDictionaryId)
			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureInventory/v2/alarmDictionaries", testCase.expectedQueryParams)
		})
	}
}

func TestInventoryGetInventoryAlarmDictionary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		dictionaryID  uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:         "success",
			dictionaryID: defaultInventoryAlarmDictID,
			handler:      jsonResponseHandler(dummyInventoryAlarmDictionary, http.StatusOK),
		},
		{
			name:          "server error 500",
			dictionaryID:  defaultInventoryAlarmDictID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get inventory alarm dictionary: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetInventoryAlarmDictionary(testCase.dictionaryID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyInventoryAlarmDictionary.AlarmDictionaryId, result.AlarmDictionaryId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/alarmDictionaries/%s", testCase.dictionaryID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestInventoryListLocations(t *testing.T) {
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
			handler:             jsonResponseHandler([]inventory.LocationInfo{dummyLocationInfo}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "test location")),
			},
			handler: jsonResponseHandler([]inventory.LocationInfo{dummyLocationInfo}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,'test location')",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list locations: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.ListLocations(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyLocationInfo.GlobalLocationId, result[0].GlobalLocationId)
			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureInventory/v2/locations", testCase.expectedQueryParams)
		})
	}
}

func TestInventoryGetLocation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		globalLocationID string
		handler          http.HandlerFunc
		expectedError    string
	}{
		{
			name:             "success",
			globalLocationID: defaultGlobalLocationID,
			handler:          jsonResponseHandler(dummyLocationInfo, http.StatusOK),
		},
		{
			name:             "server error 500",
			globalLocationID: defaultGlobalLocationID,
			handler:          jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:    "failed to get location: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetLocation(testCase.globalLocationID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyLocationInfo.GlobalLocationId, result.GlobalLocationId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/locations/%s", testCase.globalLocationID)
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestInventoryListOCloudSites(t *testing.T) {
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
			handler:             jsonResponseHandler([]inventory.OCloudSiteInfo{dummyOCloudSiteInfo}),
			expectedQueryParams: nil,
		},
		{
			name: "success with filter",
			opts: []ListOption{
				WithFilter(filter.Equals("name", "test ocloud site")),
			},
			handler: jsonResponseHandler([]inventory.OCloudSiteInfo{dummyOCloudSiteInfo}),
			expectedQueryParams: map[string]string{
				"filter": "(eq,name,'test ocloud site')",
			},
		},
		{
			name:                "server error 500",
			opts:                nil,
			handler:             jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:       "failed to list O-Cloud sites: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.ListOCloudSites(testCase.opts...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyOCloudSiteInfo.OCloudSiteId, result[0].OCloudSiteId)
			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureInventory/v2/oCloudSites", testCase.expectedQueryParams)
		})
	}
}

func TestInventoryGetOCloudSite(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		siteID        uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			siteID:  defaultOCloudSiteID,
			handler: jsonResponseHandler(dummyOCloudSiteInfo, http.StatusOK),
		},
		{
			name:          "server error 500",
			siteID:        defaultOCloudSiteID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get O-Cloud site: received error from api:",
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

			client, err := inventory.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			result, err := inventoryClient.GetOCloudSite(testCase.siteID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyOCloudSiteInfo.OCloudSiteId, result.OCloudSiteId)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureInventory/v2/oCloudSites/%s", testCase.siteID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestBuildInventory(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(jsonResponseHandler([]inventory.ResourceType{dummyResourceType}))
	defer server.Close()

	inventoryClient, err := NewClientBuilder(server.URL).BuildInventory()

	assert.NoError(t, err)
	assert.NotNil(t, inventoryClient)

	result, err := inventoryClient.ListResourceTypes()
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

//nolint:funlen // Since this is only long because of the number of functions, we can ignore the length.
func TestInventoryNetworkError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		testFunc func(client *InventoryClient) error
	}{
		{
			name: "GetAllVersions network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetAllVersions()

				return err
			},
		},
		{
			name: "GetCloudInfo network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetCloudInfo()

				return err
			},
		},
		{
			name: "GetMinorVersions network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetMinorVersions()

				return err
			},
		},
		{
			name: "ListDeploymentManagers network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.ListDeploymentManagers()

				return err
			},
		},
		{
			name: "GetDeploymentManager network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetDeploymentManager(defaultDeploymentManagerID)

				return err
			},
		},
		{
			name: "ListInventorySubscriptions network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.ListInventorySubscriptions()

				return err
			},
		},
		{
			name: "CreateInventorySubscription network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.CreateInventorySubscription(dummyInventorySubscription)

				return err
			},
		},
		{
			name: "GetInventorySubscription network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetInventorySubscription(defaultInventorySubscriptionID)

				return err
			},
		},
		{
			name: "DeleteInventorySubscription network error",
			testFunc: func(client *InventoryClient) error {
				return client.DeleteInventorySubscription(defaultInventorySubscriptionID)
			},
		},
		{
			name: "ListResourceTypes network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.ListResourceTypes()

				return err
			},
		},
		{
			name: "GetResourceType network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetResourceType(defaultResourceTypeID)

				return err
			},
		},
		{
			name: "GetResourceTypeAlarmDictionary network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetResourceTypeAlarmDictionary(defaultResourceTypeID)

				return err
			},
		},
		{
			name: "ListResourcePools network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.ListResourcePools()

				return err
			},
		},
		{
			name: "GetResourcePool network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetResourcePool(defaultResourcePoolID)

				return err
			},
		},
		{
			name: "ListResources network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.ListResources(defaultResourcePoolID)

				return err
			},
		},
		{
			name: "GetResource network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetResource(defaultResourcePoolID, defaultResourceID)

				return err
			},
		},
		{
			name: "ListInventoryAlarmDictionaries network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.ListInventoryAlarmDictionaries()

				return err
			},
		},
		{
			name: "GetInventoryAlarmDictionary network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetInventoryAlarmDictionary(defaultInventoryAlarmDictID)

				return err
			},
		},
		{
			name: "ListLocations network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.ListLocations()

				return err
			},
		},
		{
			name: "GetLocation network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetLocation(defaultGlobalLocationID)

				return err
			},
		},
		{
			name: "ListOCloudSites network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.ListOCloudSites()

				return err
			},
		},
		{
			name: "GetOCloudSite network error",
			testFunc: func(client *InventoryClient) error {
				_, err := client.GetOCloudSite(defaultOCloudSiteID)

				return err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client, err := inventory.NewClientWithResponses("http://192.0.2.0:8080",
				inventory.WithHTTPClient(&http.Client{Timeout: time.Second * 1}))
			assert.NoError(t, err)

			inventoryClient := &InventoryClient{ClientWithResponsesInterface: client}
			err = testCase.testFunc(inventoryClient)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error contacting api")
		})
	}
}
