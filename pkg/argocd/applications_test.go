package argocd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	argocdtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/argocd/argocdtypes/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultApplicationName   = "application-name"
	defaultApplicationNsName = "application-ns-name"
)

var (
	defaultApplicationCondition = argocdtypes.ApplicationCondition{
		Type:    argocdtypes.ApplicationConditionSyncError,
		Message: "test-message",
	}
	applicationGVK = argocdtypes.ApplicationSchemaGroupVersionKind
)

func TestNewApplicationBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[argocdtypes.Application, ApplicationBuilder](
		NewApplicationBuilder, argocdtypes.AddToScheme, applicationGVK,
	).ExecuteTests(t)
}

func TestPullApplication(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[argocdtypes.Application, ApplicationBuilder](
		PullApplication, argocdtypes.AddToScheme, applicationGVK,
	).ExecuteTests(t)
}

func TestApplicationBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[argocdtypes.Application, ApplicationBuilder](
		argocdtypes.AddToScheme,
		applicationGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestApplicationWithGitDetails(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		testBuilder    *ApplicationBuilder
		gitRepo        string
		gitBranch      string
		gitPath        string
		expectedError  error
		invalidBuilder bool
	}{
		{
			name:          "valid-git-details",
			testBuilder:   buildValidApplicationBuilder(getApplicationTestClient()),
			gitRepo:       "http://test.git",
			gitBranch:     "main",
			gitPath:       "./dir/www/repo",
			expectedError: nil,
		},
		{
			name:          "empty-git-repo",
			testBuilder:   buildValidApplicationBuilder(getApplicationTestClient()),
			gitRepo:       "",
			gitBranch:     "main",
			gitPath:       "./dir/www/repo",
			expectedError: fmt.Errorf("'gitRepo' parameter is empty"),
		},
		{
			name:          "empty-git-branch",
			testBuilder:   buildValidApplicationBuilder(getApplicationTestClient()),
			gitRepo:       "http://test.git",
			gitBranch:     "",
			gitPath:       "./dir/www/repo",
			expectedError: fmt.Errorf("'gitBranch' parameter is empty"),
		},
		{
			name:          "empty-git-path",
			testBuilder:   buildValidApplicationBuilder(getApplicationTestClient()),
			gitRepo:       "http://test.git",
			gitBranch:     "main",
			gitPath:       "",
			expectedError: fmt.Errorf("'gitPath' parameter is empty"),
		},
		{
			name:           "invalid-builder",
			testBuilder:    buildInvalidApplicationBuilder(getApplicationTestClient()),
			gitRepo:        "http://test.git",
			gitBranch:      "main",
			gitPath:        "./dir/www/repo",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			applicationBuilder := testCase.testBuilder.WithGitDetails(
				testCase.gitRepo, testCase.gitBranch, testCase.gitPath)

			switch {
			case testCase.invalidBuilder:
				assert.True(t, commonerrors.IsBuilderNameEmpty(applicationBuilder.GetError()))
			case testCase.expectedError != nil:
				assert.Equal(t, testCase.expectedError, applicationBuilder.GetError())
			default:
				assert.NoError(t, applicationBuilder.GetError())
				assert.Equal(t, testCase.gitPath, applicationBuilder.Definition.Spec.Source.Path)
				assert.Equal(t, testCase.gitRepo, applicationBuilder.Definition.Spec.Source.RepoURL)
				assert.Equal(t, testCase.gitBranch, applicationBuilder.Definition.Spec.Source.TargetRevision)
			}
		})
	}
}

func TestApplicationWithGitPathAppended(t *testing.T) {
	t.Parallel()

	const testPath = "test/path"

	testCases := []struct {
		name           string
		testBuilder    *ApplicationBuilder
		hasSource      bool
		elements       []string
		expectedPath   string
		expectedError  error
		invalidBuilder bool
	}{
		{
			name:          "valid-builder-with-source",
			testBuilder:   buildValidApplicationBuilder(getApplicationTestClient()),
			hasSource:     true,
			elements:      []string{"element1", "element2"},
			expectedPath:  fmt.Sprintf("%s/%s/%s", testPath, "element1", "element2"),
			expectedError: nil,
		},
		{
			name:          "no-source",
			testBuilder:   buildValidApplicationBuilder(getApplicationTestClient()),
			hasSource:     false,
			elements:      []string{"element"},
			expectedPath:  "",
			expectedError: fmt.Errorf("cannot append to git path because the source is nil"),
		},
		{
			name:          "no-elements",
			testBuilder:   buildValidApplicationBuilder(getApplicationTestClient()),
			hasSource:     true,
			elements:      []string{},
			expectedPath:  testPath,
			expectedError: nil,
		},
		{
			name:           "invalid-builder",
			testBuilder:    buildInvalidApplicationBuilder(getApplicationTestClient()),
			hasSource:      true,
			elements:       []string{"element"},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.hasSource {
				testCase.testBuilder.Definition.Spec.Source = &argocdtypes.ApplicationSource{
					Path: testPath,
				}
			} else {
				testCase.testBuilder.Definition.Spec.Source = nil
			}

			applicationBuilder := testCase.testBuilder.WithGitPathAppended(testCase.elements...)

			switch {
			case testCase.invalidBuilder:
				assert.True(t, commonerrors.IsBuilderNameEmpty(applicationBuilder.GetError()))
			case testCase.expectedError != nil:
				assert.Equal(t, testCase.expectedError, applicationBuilder.GetError())
			default:
				assert.NoError(t, applicationBuilder.GetError())
				assert.Equal(t, testCase.expectedPath, applicationBuilder.Definition.Spec.Source.Path)
			}
		})
	}
}

func TestApplicationWaitForCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		conditionMet  bool
		expectedError error
	}{
		{
			name:          "condition-met",
			exists:        true,
			conditionMet:  true,
			expectedError: nil,
		},
		{
			name:         "application-does-not-exist",
			exists:       false,
			conditionMet: true,
			expectedError: fmt.Errorf(
				"application object %s in namespace %s does not exist", defaultApplicationName, defaultApplicationNsName),
		},
		{
			name:          "condition-not-met",
			exists:        true,
			conditionMet:  false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				application := buildDummyApplication(defaultApplicationName, defaultApplicationNsName)

				if testCase.conditionMet {
					application.Status.Conditions = append(application.Status.Conditions, defaultApplicationCondition)
				}

				runtimeObjects = append(runtimeObjects, application)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{argocdtypes.AddToScheme},
			})

			testBuilder := buildValidApplicationBuilder(testSettings)

			_, err := testBuilder.WaitForCondition(defaultApplicationCondition, time.Second)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestApplicationDoesGitPathExist(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		hasSource bool
		validURL  bool
		exists    bool
	}{
		{
			name:      "exists",
			hasSource: true,
			validURL:  true,
			exists:    true,
		},
		{
			name:      "no-source",
			hasSource: false,
			validURL:  true,
			exists:    false,
		},
		{
			name:      "invalid-url",
			hasSource: true,
			validURL:  false,
			exists:    false,
		},
		{
			name:      "does-not-exist",
			hasSource: true,
			validURL:  true,
			exists:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				requestedPath   string
				requestedMethod string
			)

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				requestedPath = request.URL.Path
				requestedMethod = request.Method

				if testCase.exists {
					writer.WriteHeader(http.StatusOK)

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			var serverURL string

			if testCase.validURL {
				serverURL = server.URL
			} else {
				serverURL = "invalid-url"
			}

			testBuilder := buildValidApplicationBuilder(getApplicationTestClient())

			if testCase.hasSource {
				testBuilder.Definition.Spec.Source = &argocdtypes.ApplicationSource{
					RepoURL:        serverURL,
					Path:           "some/path",
					TargetRevision: "main",
				}
			}

			exists := testBuilder.DoesGitPathExist("test")
			assert.Equal(t, testCase.exists, exists)

			if requestedMethod != "" {
				assert.Equal(t, http.MethodHead, requestedMethod)
			}

			if requestedPath != "" {
				assert.Equal(t, "/raw/main/some/path/test/kustomization.yaml", requestedPath)
			}
		})
	}
}

func TestApplicationWaitForSourceUpdate(t *testing.T) {
	t.Parallel()

	var expectedSource = argocdtypes.ApplicationSource{
		TargetRevision: "main",
	}

	testCases := []struct {
		name          string
		sourceExists  bool
		sourceUpdated bool
		synced        bool
		expectSynced  bool
		expectedError error
	}{
		{
			name:          "source-synced",
			sourceExists:  true,
			sourceUpdated: true,
			synced:        true,
			expectSynced:  true,
			expectedError: nil,
		},
		{
			name:          "source-not-updated",
			sourceExists:  true,
			sourceUpdated: false,
			synced:        true,
			expectSynced:  true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "source-updated-not-synced",
			sourceExists:  true,
			sourceUpdated: true,
			synced:        false,
			expectSynced:  true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "source-not-synced-expect-synced-false",
			sourceExists:  true,
			sourceUpdated: true,
			synced:        false,
			expectSynced:  false,
			expectedError: nil,
		},
		{
			name:          "source-does-not-exist",
			sourceExists:  false,
			sourceUpdated: true,
			synced:        true,
			expectSynced:  true,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testApp := buildDummyApplication(defaultApplicationName, defaultApplicationNsName)

			if !testCase.sourceExists {
				testApp.Spec.Source = nil
			} else {
				testApp.Spec.Source = &expectedSource
			}

			if testCase.sourceUpdated {
				testApp.Status.Sync.ComparedTo.Source = expectedSource
			}

			if testCase.synced {
				testApp.Status.Sync.Status = argocdtypes.SyncStatusCodeSynced
			}

			testBuilder := buildValidApplicationBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{testApp},
				SchemeAttachers: []clients.SchemeAttacher{argocdtypes.AddToScheme},
			}))

			err := testBuilder.WaitForSourceUpdate(testCase.expectSynced, time.Second)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func buildValidApplicationBuilder(apiClient *clients.Settings) *ApplicationBuilder {
	builder := NewApplicationBuilder(apiClient, defaultApplicationName, defaultApplicationNsName)
	builder.Definition.Spec.Source = &argocdtypes.ApplicationSource{}

	return builder
}

func buildInvalidApplicationBuilder(apiClient *clients.Settings) *ApplicationBuilder {
	return NewApplicationBuilder(apiClient, "", defaultApplicationNsName)
}

func getApplicationTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{argocdtypes.AddToScheme},
	})
}

func buildDummyApplication(name, namespace string) *argocdtypes.Application {
	return &argocdtypes.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},

		Spec: argocdtypes.ApplicationSpec{
			Source: &argocdtypes.ApplicationSource{},
		},
	}
}
