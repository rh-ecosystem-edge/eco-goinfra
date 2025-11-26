package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/filter"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/alarms"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

var (
	// dummyAlarmEventRecord is a test alarm for use in tests.
	dummyAlarmEventRecord = AlarmEventRecord{
		AlarmEventRecordId: uuid.New(),
		AlarmRaisedTime:    time.Now(),
		AlarmDefinitionID:  uuid.New(),
		ProbableCauseID:    uuid.New(),
		ResourceID:         uuid.New(),
		ResourceTypeID:     uuid.New(),
		PerceivedSeverity:  alarms.CRITICAL,
		AlarmAcknowledged:  false,
		Extensions:         map[string]string{"key1": "value1"},
	}

	// dummyAlarmEventRecordModifications is test modifications for use in tests.
	dummyAlarmEventRecordModifications = AlarmEventRecordModifications{
		AlarmAcknowledged: ptr.To(true),
		PerceivedSeverity: ptr.To(alarms.MINOR),
	}

	// dummyAlarmServiceConfiguration is test configuration for use in tests.
	dummyAlarmServiceConfiguration = AlarmServiceConfiguration{
		RetentionPeriod: 30,
		Extensions:      map[string]string{"config1": "value1"},
	}

	// dummyAlarmServiceConfigurationPatch is test patch for use in tests.
	dummyAlarmServiceConfigurationPatch = AlarmServiceConfigurationPatch{
		RetentionPeriod: ptr.To(30),
	}

	// dummyAlarmSubscriptionInfo is test subscription for use in tests.
	dummyAlarmSubscriptionInfo = AlarmSubscriptionInfo{
		AlarmSubscriptionId:    ptr.To(uuid.New()),
		Callback:               "http://callback.example.com",
		Filter:                 ptr.To(AlarmSubscriptionFilterACKNOWLEDGE),
		ConsumerSubscriptionId: ptr.To(uuid.New()),
	}

	defaultAlarmID        = dummyAlarmEventRecord.AlarmEventRecordId
	defaultSubscriptionID = *dummyAlarmSubscriptionInfo.AlarmSubscriptionId
)

func TestListAlarms(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		filter         []filter.Filter
		handler        http.HandlerFunc
		expectedError  string
		expectedFilter string
	}{
		{
			name:           "success without filter",
			filter:         nil,
			handler:        jsonResponseHandler([]alarms.AlarmEventRecord{dummyAlarmEventRecord}),
			expectedFilter: "",
		},
		{
			name: "success with filters",
			filter: []filter.Filter{
				filter.Equals("perceivedSeverity", "CRITICAL"),
				filter.Equals("alarmAcknowledged", "false"),
			},
			handler:        jsonResponseHandler([]alarms.AlarmEventRecord{dummyAlarmEventRecord}),
			expectedFilter: "(eq,perceivedSeverity,CRITICAL)",
		},
		{
			name:           "server error 500",
			filter:         nil,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to list alarms: received error from api:",
			expectedFilter: "",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.ListAlarms(testCase.filter...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmEventRecord.AlarmEventRecordId, result[0].AlarmEventRecordId)

			queryParams := make(map[string]string)
			if testCase.expectedFilter != "" {
				queryParams["filter"] = testCase.expectedFilter
			}

			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureMonitoring/v1/alarms", queryParams)
		})
	}
}

func TestGetAlarm(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		alarmID       uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			alarmID: defaultAlarmID,
			handler: jsonResponseHandler(dummyAlarmEventRecord, http.StatusOK),
		},
		{
			name:          "server error 500",
			alarmID:       defaultAlarmID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get alarm: received error from api:",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.GetAlarm(testCase.alarmID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmEventRecord.AlarmEventRecordId, result.AlarmEventRecordId)
			assert.Equal(t, dummyAlarmEventRecord.ProbableCauseID, result.ProbableCauseID)
			assert.Equal(t, dummyAlarmEventRecord.PerceivedSeverity, result.PerceivedSeverity)

			expectedPath := fmt.Sprintf("/o2ims-infrastructureMonitoring/v1/alarms/%s", testCase.alarmID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestPatchAlarm(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		alarmID       uuid.UUID
		patch         alarms.AlarmEventRecordModifications
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			alarmID: defaultAlarmID,
			patch:   dummyAlarmEventRecordModifications,
			handler: jsonResponseHandler(dummyAlarmEventRecordModifications, http.StatusOK),
		},
		{
			name:          "server error 500",
			alarmID:       defaultAlarmID,
			patch:         dummyAlarmEventRecordModifications,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to patch alarm: received error from api:",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}

			result, err := alarmsClient.PatchAlarm(testCase.alarmID, testCase.patch)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmEventRecordModifications.AlarmAcknowledged, result.AlarmAcknowledged)
			assert.Equal(t, dummyAlarmEventRecordModifications.PerceivedSeverity, result.PerceivedSeverity)

			expectedPath := fmt.Sprintf("/o2ims-infrastructureMonitoring/v1/alarms/%s", testCase.alarmID.String())
			validateHTTPRequest(t, capturedRequest, "PATCH", expectedPath, nil, "application/merge-patch+json")
		})
	}
}

func TestGetServiceConfiguration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			handler: jsonResponseHandler(dummyAlarmServiceConfiguration),
		},
		{
			name: "server error 500",
			handler: jsonResponseHandler(
				dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get service configuration: received error from api:",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}

			result, err := alarmsClient.GetServiceConfiguration()
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmServiceConfiguration.RetentionPeriod, result.RetentionPeriod)
			assert.Equal(t, dummyAlarmServiceConfiguration.Extensions, result.Extensions)

			validateHTTPRequest(t, capturedRequest, "GET", "/o2ims-infrastructureMonitoring/v1/alarmServiceConfiguration", nil)
		})
	}
}

func TestUpdateAlarmServiceConfiguration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		config        alarms.AlarmServiceConfiguration
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			config:  dummyAlarmServiceConfiguration,
			handler: jsonResponseHandler(dummyAlarmServiceConfiguration),
		},
		{
			name:          "server error 500",
			config:        dummyAlarmServiceConfiguration,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to update service configuration: received error from api:",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}

			result, err := alarmsClient.UpdateAlarmServiceConfiguration(testCase.config)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmServiceConfiguration.RetentionPeriod, result.RetentionPeriod)
			assert.Equal(t, dummyAlarmServiceConfiguration.Extensions, result.Extensions)

			validateHTTPRequest(t, capturedRequest, "PUT", "/o2ims-infrastructureMonitoring/v1/alarmServiceConfiguration", nil)
		})
	}
}

func TestPatchAlarmServiceConfiguration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		config        alarms.AlarmServiceConfigurationPatch
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			config:  dummyAlarmServiceConfigurationPatch,
			handler: jsonResponseHandler(dummyAlarmServiceConfiguration),
		},
		{
			name:          "server error 500",
			config:        dummyAlarmServiceConfigurationPatch,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to patch service configuration: received error from api:",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}

			result, err := alarmsClient.PatchAlarmServiceConfiguration(testCase.config)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmServiceConfiguration.RetentionPeriod, result.RetentionPeriod)
			assert.Equal(t, dummyAlarmServiceConfiguration.Extensions, result.Extensions)

			expectedPath := "/o2ims-infrastructureMonitoring/v1/alarmServiceConfiguration"
			validateHTTPRequest(t, capturedRequest, "PATCH", expectedPath, nil, "application/merge-patch+json")
		})
	}
}

func TestListSubscriptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		filter         []filter.Filter
		handler        http.HandlerFunc
		expectedError  string
		expectedFilter string
	}{
		{
			name:           "success without filter",
			filter:         nil,
			handler:        jsonResponseHandler([]alarms.AlarmSubscriptionInfo{dummyAlarmSubscriptionInfo}),
			expectedFilter: "",
		},
		{
			name:           "success with filters",
			filter:         []filter.Filter{filter.Equals("callback", "http://example.com"), filter.Equals("filter", "test")},
			handler:        jsonResponseHandler([]alarms.AlarmSubscriptionInfo{dummyAlarmSubscriptionInfo}),
			expectedFilter: "(eq,callback,http://example.com)",
		},
		{
			name:           "server error 500",
			filter:         nil,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to list subscriptions: received error from api:",
			expectedFilter: "",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}

			result, err := alarmsClient.ListSubscriptions(testCase.filter...)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmSubscriptionInfo.AlarmSubscriptionId, result[0].AlarmSubscriptionId)

			queryParams := make(map[string]string)
			if testCase.expectedFilter != "" {
				queryParams["filter"] = testCase.expectedFilter
			}

			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureMonitoring/v1/alarmSubscriptions", queryParams)
		})
	}
}

func TestCreateSubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		subscription  alarms.AlarmSubscriptionInfo
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:         "success",
			subscription: dummyAlarmSubscriptionInfo,
			handler:      jsonResponseHandler(dummyAlarmSubscriptionInfo, http.StatusCreated),
		},
		{
			name:          "server error 500",
			subscription:  dummyAlarmSubscriptionInfo,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to create subscription: received error from api:",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.CreateSubscription(testCase.subscription)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmSubscriptionInfo.AlarmSubscriptionId, result.AlarmSubscriptionId)
			assert.Equal(t, dummyAlarmSubscriptionInfo.Callback, result.Callback)
			assert.Equal(t, dummyAlarmSubscriptionInfo.Filter, result.Filter)

			validateHTTPRequest(t, capturedRequest, "POST", "/o2ims-infrastructureMonitoring/v1/alarmSubscriptions", nil)
		})
	}
}

func TestGetSubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID uuid.UUID
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: defaultSubscriptionID,
			handler:        jsonResponseHandler(dummyAlarmSubscriptionInfo, http.StatusOK),
		},
		{
			name:           "server error 500",
			subscriptionID: defaultSubscriptionID,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to get subscription: received error from api:",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.GetSubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmSubscriptionInfo.AlarmSubscriptionId, result.AlarmSubscriptionId)
			assert.Equal(t, dummyAlarmSubscriptionInfo.Callback, result.Callback)
			assert.Equal(t, dummyAlarmSubscriptionInfo.Filter, result.Filter)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureMonitoring/v1/alarmSubscriptions/%s", testCase.subscriptionID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID uuid.UUID
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: defaultSubscriptionID,
			handler:        jsonResponseHandler(nil, http.StatusOK),
		},
		{
			name:           "server error 500",
			subscriptionID: defaultSubscriptionID,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to delete subscription: received error from api:",
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

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			err = alarmsClient.DeleteSubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureMonitoring/v1/alarmSubscriptions/%s", testCase.subscriptionID.String())
			validateHTTPRequest(t, capturedRequest, "DELETE", expectedPath, nil)
		})
	}
}

//nolint:funlen // Since this is only long because of the number of functions, we can ignore the length.
func TestAlarmsNetworkError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		testFunc func(client *AlarmsClient) error
	}{
		{
			name: "ListAlarms network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.ListAlarms()

				return err
			},
		},
		{
			name: "GetAlarm network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.GetAlarm(defaultAlarmID)

				return err
			},
		},
		{
			name: "PatchAlarm network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.PatchAlarm(defaultAlarmID, dummyAlarmEventRecordModifications)

				return err
			},
		},
		{
			name: "GetServiceConfiguration network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.GetServiceConfiguration()

				return err
			},
		},
		{
			name: "UpdateAlarmServiceConfiguration network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.UpdateAlarmServiceConfiguration(dummyAlarmServiceConfiguration)

				return err
			},
		},
		{
			name: "PatchAlarmServiceConfiguration network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.PatchAlarmServiceConfiguration(dummyAlarmServiceConfigurationPatch)

				return err
			},
		},
		{
			name: "ListSubscriptions network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.ListSubscriptions()

				return err
			},
		},
		{
			name: "CreateSubscription network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.CreateSubscription(dummyAlarmSubscriptionInfo)

				return err
			},
		},
		{
			name: "GetSubscription network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.GetSubscription(defaultSubscriptionID)

				return err
			},
		},
		{
			name: "DeleteSubscription network error",
			testFunc: func(client *AlarmsClient) error {
				return client.DeleteSubscription(defaultSubscriptionID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// 192.0.2.0 is a reserved test address so we never accidentally use a valid IP. Still, we set a
			// timeout to ensure that we do not timeout the test.
			client, err := alarms.NewClientWithResponses("http://192.0.2.0:8080",
				alarms.WithHTTPClient(&http.Client{Timeout: time.Second * 1}))
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			err = testCase.testFunc(alarmsClient)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error contacting api")
		})
	}
}
