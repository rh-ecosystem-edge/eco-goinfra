package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/filter"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/artifacts"
	"github.com/stretchr/testify/assert"
)

var (
	// dummyManagedInfrastructureTemplate is a test template for use in tests.
	dummyManagedInfrastructureTemplate = artifacts.ManagedInfrastructureTemplate{
		ArtifactResourceId: uuid.New(),
		Name:               "test-template",
		Description:        "Test template description",
		Version:            "v1.0.0",
		ParameterSchema:    map[string]any{"param1": "string"},
		Extensions:         &map[string]string{"key1": "value1"},
	}

	// dummyManagedInfrastructureTemplateDefaults is test defaults for use in tests.
	dummyManagedInfrastructureTemplateDefaults = artifacts.ManagedInfrastructureTemplateDefaults{
		ClusterInstanceDefaults: &map[string]any{"cluster": "default"},
		PolicyTemplateDefaults:  &map[string]any{"policy": "default"},
	}

	defaultTemplateID = "test-template.v1-0-0"
)

func TestListManagedInfrastructureTemplates(t *testing.T) {
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
			handler:        jsonResponseHandler([]artifacts.ManagedInfrastructureTemplate{dummyManagedInfrastructureTemplate}),
			expectedFilter: "",
		},
		{
			name:           "success with filters",
			filter:         []filter.Filter{filter.Equals("name", "test-template"), filter.Equals("version", "v1.0")},
			handler:        jsonResponseHandler([]artifacts.ManagedInfrastructureTemplate{dummyManagedInfrastructureTemplate}),
			expectedFilter: "(eq,name,test-template)",
		},
		{
			name:           "server error 500",
			filter:         nil,
			handler:        jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:  "failed to list ManagedInfrastructureTemplates: received error from api:",
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

			client, err := artifacts.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			artifactsClient := &ArtifactsClient{ClientWithResponsesInterface: client}

			result, err := artifactsClient.ListManagedInfrastructureTemplates(testCase.filter...)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, dummyManagedInfrastructureTemplate.Name, result[0].Name)

			queryParams := make(map[string]string)
			if testCase.expectedFilter != "" {
				queryParams["filter"] = testCase.expectedFilter
			}

			validateHTTPRequest(
				t, capturedRequest, "GET", "/o2ims-infrastructureArtifacts/v1/managedInfrastructureTemplates", queryParams)
		})
	}
}

func TestGetManagedInfrastructureTemplate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		templateID    string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:       "success",
			templateID: defaultTemplateID,
			handler:    jsonResponseHandler(dummyManagedInfrastructureTemplate, http.StatusOK),
		},
		{
			name:          "server error 500",
			templateID:    defaultTemplateID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get ManagedInfrastructureTemplate: received error from api:",
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

			client, err := artifacts.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			artifactsClient := &ArtifactsClient{ClientWithResponsesInterface: client}

			result, err := artifactsClient.GetManagedInfrastructureTemplate(testCase.templateID)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, dummyManagedInfrastructureTemplate.Name, result.Name)
			assert.Equal(t, dummyManagedInfrastructureTemplate.Description, result.Description)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureArtifacts/v1/managedInfrastructureTemplates/%s", testCase.templateID)
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestGetManagedInfrastructureTemplateDefaults(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		templateID    string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:       "success",
			templateID: defaultTemplateID,
			handler:    jsonResponseHandler(dummyManagedInfrastructureTemplateDefaults),
		},
		{
			name:          "server error 500",
			templateID:    defaultTemplateID,
			handler:       jsonResponseHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get ManagedInfrastructureTemplateDefaults: received error from api:",
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

			client, err := artifacts.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			artifactsClient := &ArtifactsClient{ClientWithResponsesInterface: client}

			result, err := artifactsClient.GetManagedInfrastructureTemplateDefaults(testCase.templateID)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			assert.NotNil(t, result.ClusterInstanceDefaults)
			assert.NotNil(t, result.PolicyTemplateDefaults)

			assert.Equal(t, dummyManagedInfrastructureTemplateDefaults.ClusterInstanceDefaults, result.ClusterInstanceDefaults)
			assert.Equal(t, dummyManagedInfrastructureTemplateDefaults.PolicyTemplateDefaults, result.PolicyTemplateDefaults)

			expectedPath := fmt.Sprintf(
				"/o2ims-infrastructureArtifacts/v1/managedInfrastructureTemplates/%s/defaults", testCase.templateID)
			validateHTTPRequest(t, capturedRequest, "GET", expectedPath, nil)
		})
	}
}

func TestArtifactsNetworkError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		testFunc func(client *ArtifactsClient) error
	}{
		{
			name: "ListManagedInfrastructureTemplates network error",
			testFunc: func(client *ArtifactsClient) error {
				_, err := client.ListManagedInfrastructureTemplates()

				return err
			},
		},
		{
			name: "GetManagedInfrastructureTemplate network error",
			testFunc: func(client *ArtifactsClient) error {
				_, err := client.GetManagedInfrastructureTemplate("test-id")

				return err
			},
		},
		{
			name: "GetManagedInfrastructureTemplateDefaults network error",
			testFunc: func(client *ArtifactsClient) error {
				_, err := client.GetManagedInfrastructureTemplateDefaults("test-id")

				return err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// 192.0.2.0 is a reserved test address so we never accidentally use a valid IP. Still, we set a
			// timeout to ensure that we do not timeout the test.
			client, err := artifacts.NewClientWithResponses("http://192.0.2.0:8080",
				artifacts.WithHTTPClient(&http.Client{Timeout: time.Second * 1}))
			assert.NoError(t, err)

			artifactsClient := &ArtifactsClient{ClientWithResponsesInterface: client}
			err = testCase.testFunc(artifactsClient)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error contacting api")
		})
	}
}
