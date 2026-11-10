package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/filter"
	alarmsv2 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/alarms/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

var (
	// dummyAlarmEventRecordV2 is a test alarm for use in v2 tests.
	dummyAlarmEventRecordV2 = AlarmEventRecordV2{
		AlarmEventRecordId: uuid.New(),
		AlarmRaisedTime:    time.Now(),
		AlarmDefinitionId:  uuid.New(),
		ProbableCauseId:    uuid.New(),
		ObjectId:           uuid.New(),
		ObjectTypeId:       uuid.New(),
		ObjectClass:        "oran.o2ims.cluster.NodeCluster",
		PerceivedSeverity:  PerceivedSeverityV2CRITICAL,
		AlarmAcknowledged:  false,
		Extensions:         map[string]any{"key1": "value1"},
	}

	// dummyAlarmEventRecordModificationsV2 is test modifications for use in v2 tests.
	dummyAlarmEventRecordModificationsV2 = AlarmEventRecordModificationsV2{
		AlarmAcknowledged: ptr.To(true),
		PerceivedSeverity: ptr.To(PerceivedSeverityV2MINOR),
	}

	// dummyAlarmServiceConfigurationV2 is test configuration for use in v2 tests.
	dummyAlarmServiceConfigurationV2 = AlarmServiceConfigurationV2{
		RetentionPeriod: 30,
		Extensions:      ptr.To(map[string]any{"config1": "value1"}),
	}

	// dummyAlarmServiceConfigurationPatchV2 is test patch for use in v2 tests.
	dummyAlarmServiceConfigurationPatchV2 = AlarmServiceConfigurationPatchV2{
		RetentionPeriod: ptr.To(30),
	}

	// dummyAlarmSubscriptionInfoV2 is test subscription for use in v2 tests.
	dummyAlarmSubscriptionInfoV2 = AlarmSubscriptionInfoV2{
		AlarmSubscriptionId:    ptr.To(uuid.New()),
		Callback:               "http://callback.example.com",
		EventFilter:            nullable.NewNullableWithValue(AlarmSubscriptionFilterV2ACKNOWLEDGE),
		ConsumerSubscriptionId: ptr.To("consumer-1"),
	}

	// dummyAlarmSubscriptionUpdateV2 is test subscription update for use in v2 tests.
	dummyAlarmSubscriptionUpdateV2 = AlarmSubscriptionUpdateV2{
		EventFilter: nullable.NewNullableWithValue(AlarmSubscriptionUpdateFilterV2CLEAR),
	}

	defaultAlarmIDV2        = dummyAlarmEventRecordV2.AlarmEventRecordId
	defaultSubscriptionIDV2 = *dummyAlarmSubscriptionInfoV2.AlarmSubscriptionId
)

func TestListAlarmsV2(t *testing.T) {
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
			handler:        jsonResponseHandler([]alarmsv2.AlarmEventRecord{dummyAlarmEventRecordV2}),
			expectedFilter: "",
		},
		{
			name: "success with filters",
			filter: []filter.Filter{
				filter.Equals("perceivedSeverity", "CRITICAL"),
				filter.Equals("alarmAcknowledged", "false"),
			},
			handler:        jsonResponseHandler([]alarmsv2.AlarmEventRecord{dummyAlarmEventRecordV2}),
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}
			result, err := alarmsClient.ListAlarms(testCase.filter...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmEventRecordV2.AlarmEventRecordId, result[0].AlarmEventRecordId)

			queryParams := make(map[string]string)
			if testCase.expectedFilter != "" {
				queryParams["filter"] = testCase.expectedFilter
			}

			validateHTTPRequest(
				t, capturedRequest, "GET", "/O2ims_infrastructureMonitoring/v2/alarms", queryParams)
		})
	}
}

func TestGetAlarmV2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		alarmID       uuid.UUID
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			alarmID: defaultAlarmIDV2,
			handler: jsonResponseHandler(dummyAlarmEventRecordV2, http.StatusOK),
		},
		{
			name:          "server error 500",
			alarmID:       defaultAlarmIDV2,
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}
			result, err := alarmsClient.GetAlarm(testCase.alarmID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmEventRecordV2.AlarmEventRecordId, result.AlarmEventRecordId)
			assert.Equal(t, dummyAlarmEventRecordV2.ProbableCauseId, result.ProbableCauseId)
			assert.Equal(t, dummyAlarmEventRecordV2.PerceivedSeverity, result.PerceivedSeverity)
			assert.Equal(t, dummyAlarmEventRecordV2.ObjectClass, result.ObjectClass)

			expectedPath := fmt.Sprintf("/O2ims_infrastructureMonitoring/v2/alarms/%s", testCase.alarmID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestPatchAlarmV2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		alarmID       uuid.UUID
		patch         alarmsv2.AlarmEventRecordModifications
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			alarmID: defaultAlarmIDV2,
			patch:   dummyAlarmEventRecordModificationsV2,
			handler: jsonResponseHandler(dummyAlarmEventRecordModificationsV2, http.StatusOK),
		},
		{
			name:          "server error 500",
			alarmID:       defaultAlarmIDV2,
			patch:         dummyAlarmEventRecordModificationsV2,
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}

			result, err := alarmsClient.PatchAlarm(testCase.alarmID, testCase.patch)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmEventRecordModificationsV2.AlarmAcknowledged, result.AlarmAcknowledged)
			assert.Equal(t, dummyAlarmEventRecordModificationsV2.PerceivedSeverity, result.PerceivedSeverity)

			expectedPath := fmt.Sprintf("/O2ims_infrastructureMonitoring/v2/alarms/%s", testCase.alarmID.String())
			validateHTTPRequest(t, capturedRequest, "PATCH", expectedPath, nil, "application/merge-patch+json")
		})
	}
}

func TestGetServiceConfigurationV2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			handler: jsonResponseHandler(dummyAlarmServiceConfigurationV2),
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}

			result, err := alarmsClient.GetServiceConfiguration()
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmServiceConfigurationV2.RetentionPeriod, result.RetentionPeriod)
			assert.Equal(t, dummyAlarmServiceConfigurationV2.Extensions, result.Extensions)

			validateHTTPRequest(
				t, capturedRequest, "GET", "/O2ims_infrastructureMonitoring/v2/alarmServiceConfiguration", nil)
		})
	}
}

func TestUpdateAlarmServiceConfigurationV2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		config        alarmsv2.AlarmServiceConfiguration
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			config:  dummyAlarmServiceConfigurationV2,
			handler: jsonResponseHandler(dummyAlarmServiceConfigurationV2),
		},
		{
			name:          "server error 500",
			config:        dummyAlarmServiceConfigurationV2,
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}

			result, err := alarmsClient.UpdateAlarmServiceConfiguration(testCase.config)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmServiceConfigurationV2.RetentionPeriod, result.RetentionPeriod)
			assert.Equal(t, dummyAlarmServiceConfigurationV2.Extensions, result.Extensions)

			validateHTTPRequest(
				t, capturedRequest, "PUT", "/O2ims_infrastructureMonitoring/v2/alarmServiceConfiguration", nil)
		})
	}
}

func TestPatchAlarmServiceConfigurationV2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		config        alarmsv2.AlarmServiceConfigurationPatch
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			config:  dummyAlarmServiceConfigurationPatchV2,
			handler: jsonResponseHandler(dummyAlarmServiceConfigurationV2),
		},
		{
			name:          "server error 500",
			config:        dummyAlarmServiceConfigurationPatchV2,
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}

			result, err := alarmsClient.PatchAlarmServiceConfiguration(testCase.config)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmServiceConfigurationV2.RetentionPeriod, result.RetentionPeriod)
			assert.Equal(t, dummyAlarmServiceConfigurationV2.Extensions, result.Extensions)

			expectedPath := "/O2ims_infrastructureMonitoring/v2/alarmServiceConfiguration"
			validateHTTPRequest(t, capturedRequest, "PATCH", expectedPath, nil, "application/merge-patch+json")
		})
	}
}

func TestListSubscriptionsV2(t *testing.T) {
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
			handler:        jsonResponseHandler([]alarmsv2.AlarmSubscriptionInfo{dummyAlarmSubscriptionInfoV2}),
			expectedFilter: "",
		},
		{
			name: "success with filters",
			filter: []filter.Filter{
				filter.Equals("callback", "http://example.com"), filter.Equals("eventFilter", "test"),
			},
			handler:        jsonResponseHandler([]alarmsv2.AlarmSubscriptionInfo{dummyAlarmSubscriptionInfoV2}),
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}

			result, err := alarmsClient.ListSubscriptions(testCase.filter...)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.AlarmSubscriptionId, result[0].AlarmSubscriptionId)

			queryParams := make(map[string]string)
			if testCase.expectedFilter != "" {
				queryParams["filter"] = testCase.expectedFilter
			}

			validateHTTPRequest(
				t, capturedRequest, "GET", "/O2ims_infrastructureMonitoring/v2/alarmSubscriptions", queryParams)
		})
	}
}

func TestCreateSubscriptionV2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		subscription  alarmsv2.AlarmSubscriptionInfo
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:         "success",
			subscription: dummyAlarmSubscriptionInfoV2,
			handler:      jsonResponseHandler(dummyAlarmSubscriptionInfoV2, http.StatusCreated),
		},
		{
			name:          "server error 500",
			subscription:  dummyAlarmSubscriptionInfoV2,
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}
			result, err := alarmsClient.CreateSubscription(testCase.subscription)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.AlarmSubscriptionId, result.AlarmSubscriptionId)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.Callback, result.Callback)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.EventFilter, result.EventFilter)

			validateHTTPRequest(
				t, capturedRequest, "POST", "/O2ims_infrastructureMonitoring/v2/alarmSubscriptions", nil)
		})
	}
}

func TestGetSubscriptionV2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID uuid.UUID
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: defaultSubscriptionIDV2,
			handler:        jsonResponseHandler(dummyAlarmSubscriptionInfoV2, http.StatusOK),
		},
		{
			name:           "server error 500",
			subscriptionID: defaultSubscriptionIDV2,
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}
			result, err := alarmsClient.GetSubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.AlarmSubscriptionId, result.AlarmSubscriptionId)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.Callback, result.Callback)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.EventFilter, result.EventFilter)

			expectedPath := fmt.Sprintf(
				"/O2ims_infrastructureMonitoring/v2/alarmSubscriptions/%s", testCase.subscriptionID.String())
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestDeleteSubscriptionV2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID uuid.UUID
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: defaultSubscriptionIDV2,
			handler:        jsonResponseHandler(nil, http.StatusOK),
		},
		{
			name:           "server error 500",
			subscriptionID: defaultSubscriptionIDV2,
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}
			err = alarmsClient.DeleteSubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)

			expectedPath := fmt.Sprintf(
				"/O2ims_infrastructureMonitoring/v2/alarmSubscriptions/%s", testCase.subscriptionID.String())
			validateHTTPRequest(t, capturedRequest, "DELETE", expectedPath, nil)
		})
	}
}

func TestPatchSubscriptionV2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID uuid.UUID
		patch          alarmsv2.AlarmSubscriptionUpdate
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: defaultSubscriptionIDV2,
			patch:          dummyAlarmSubscriptionUpdateV2,
			handler:        jsonResponseHandler(dummyAlarmSubscriptionInfoV2, http.StatusOK),
		},
		{
			name:           "server error 500",
			subscriptionID: defaultSubscriptionIDV2,
			patch:          dummyAlarmSubscriptionUpdateV2,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to patch subscription: received error from api:",
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

			client, err := alarmsv2.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}
			result, err := alarmsClient.PatchSubscription(testCase.subscriptionID, testCase.patch)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.AlarmSubscriptionId, result.AlarmSubscriptionId)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.Callback, result.Callback)
			assert.Equal(t, dummyAlarmSubscriptionInfoV2.EventFilter, result.EventFilter)

			expectedPath := fmt.Sprintf(
				"/O2ims_infrastructureMonitoring/v2/alarmSubscriptions/%s", testCase.subscriptionID.String())
			validateHTTPRequest(t, capturedRequest, "PATCH", expectedPath, nil, "application/merge-patch+json")
		})
	}
}

//nolint:funlen // Since this is only long because of the number of functions, we can ignore the length.
func TestAlarmsV2NetworkError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		testFunc func(client *AlarmsV2Client) error
	}{
		{
			name: "ListAlarms network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.ListAlarms()

				return err
			},
		},
		{
			name: "GetAlarm network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.GetAlarm(defaultAlarmIDV2)

				return err
			},
		},
		{
			name: "PatchAlarm network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.PatchAlarm(defaultAlarmIDV2, dummyAlarmEventRecordModificationsV2)

				return err
			},
		},
		{
			name: "GetServiceConfiguration network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.GetServiceConfiguration()

				return err
			},
		},
		{
			name: "UpdateAlarmServiceConfiguration network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.UpdateAlarmServiceConfiguration(dummyAlarmServiceConfigurationV2)

				return err
			},
		},
		{
			name: "PatchAlarmServiceConfiguration network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.PatchAlarmServiceConfiguration(dummyAlarmServiceConfigurationPatchV2)

				return err
			},
		},
		{
			name: "ListSubscriptions network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.ListSubscriptions()

				return err
			},
		},
		{
			name: "CreateSubscription network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.CreateSubscription(dummyAlarmSubscriptionInfoV2)

				return err
			},
		},
		{
			name: "GetSubscription network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.GetSubscription(defaultSubscriptionIDV2)

				return err
			},
		},
		{
			name: "DeleteSubscription network error",
			testFunc: func(client *AlarmsV2Client) error {
				return client.DeleteSubscription(defaultSubscriptionIDV2)
			},
		},
		{
			name: "PatchSubscription network error",
			testFunc: func(client *AlarmsV2Client) error {
				_, err := client.PatchSubscription(defaultSubscriptionIDV2, dummyAlarmSubscriptionUpdateV2)

				return err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// 192.0.2.0 is a reserved test address so we never accidentally use a valid IP. Still, we set a
			// timeout to ensure that we do not timeout the test.
			client, err := alarmsv2.NewClientWithResponses("http://192.0.2.0:8080",
				alarmsv2.WithHTTPClient(&http.Client{Timeout: time.Second * 1}))
			assert.NoError(t, err)

			alarmsClient := &AlarmsV2Client{ClientWithResponsesInterface: client}
			err = testCase.testFunc(alarmsClient)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error contacting api")
		})
	}
}
