package oran

import (
	"context"
	"fmt"
	"testing"
	"time"

	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultPRName            = "9c5372f3-ea1d-4a96-8157-b3b874a55cf9"
	defaultPRTemplateName    = "test-cluster-template"
	defaultPRTemplateVersion = "v4-18-0-1"
)

var (
	defaultPRCondition = metav1.Condition{
		Type:   string(provisioningv1alpha1.PRconditionTypes.Validated),
		Reason: string(provisioningv1alpha1.CRconditionReasons.Completed),
		Status: metav1.ConditionTrue,
	}

	provisioningTestSchemes = []clients.SchemeAttacher{
		provisioningv1alpha1.AddToScheme,
	}

	provisioningRequestGVK = provisioningv1alpha1.GroupVersion.WithKind("ProvisioningRequest")
)

func TestNewPRBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName        string
		name            string
		templateName    string
		templateVersion string
		client          bool
		assertError     func(error) bool
	}{
		{
			testName:        "valid builder",
			name:            defaultPRName,
			templateName:    defaultPRTemplateName,
			templateVersion: defaultPRTemplateVersion,
			client:          true,
			assertError:     func(err error) bool { return err == nil },
		},
		{
			testName:        "empty name",
			name:            "",
			templateName:    defaultPRTemplateName,
			templateVersion: defaultPRTemplateVersion,
			client:          true,
			assertError:     commonerrors.IsBuilderNameEmpty,
		},
		{
			testName:        "invalid uuid",
			name:            "not-a-uuid",
			templateName:    defaultPRTemplateName,
			templateVersion: defaultPRTemplateVersion,
			client:          true,
			assertError: func(err error) bool {
				return err != nil && err.Error() == "provisioningRequest 'name' must be a valid UUID"
			},
		},
		{
			testName:        "empty template name",
			name:            defaultPRName,
			templateName:    "",
			templateVersion: defaultPRTemplateVersion,
			client:          true,
			assertError: func(err error) bool {
				return err != nil && err.Error() == errEmptyTemplateName
			},
		},
		{
			testName:        "empty template version",
			name:            defaultPRName,
			templateName:    defaultPRTemplateName,
			templateVersion: "",
			client:          true,
			assertError: func(err error) bool {
				return err != nil && err.Error() == "provisioningRequest 'templateVersion' cannot be empty"
			},
		},
		{
			testName:        "nil client",
			name:            defaultPRName,
			templateName:    defaultPRTemplateName,
			templateVersion: defaultPRTemplateVersion,
			client:          false,
			assertError:     commonerrors.IsAPIClientNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings
			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{})
			}

			builder := NewPRBuilder(
				testSettings, testCase.name, testCase.templateName, testCase.templateVersion)

			require.NotNil(t, builder)
			require.Truef(t, testCase.assertError(builder.GetError()), "unexpected error, got: %v", builder.GetError())

			if builder.GetError() == nil {
				assert.Equal(t, testCase.name, builder.Definition.Name)
				assert.Equal(t, testCase.name, builder.Definition.Spec.Name)
				assert.Equal(t, testCase.templateName, builder.Definition.Spec.TemplateName)
				assert.Equal(t, testCase.templateVersion, builder.Definition.Spec.TemplateVersion)
				assert.Equal(t, provisioningRequestGVK, builder.GetGVK())
			}
		})
	}
}

func TestPRWithTemplateParameter(t *testing.T) {
	t.Parallel()

	t.Run("sets template parameter", func(t *testing.T) {
		t.Parallel()

		builder := buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		builder = builder.WithTemplateParameter("nodeClusterName", "cluster-1")

		require.NoError(t, builder.GetError())

		templateParameters, err := builder.GetTemplateParameters()
		require.NoError(t, err)
		assert.Equal(t, "cluster-1", templateParameters["nodeClusterName"])
	})

	t.Run("rejects empty key", func(t *testing.T) {
		t.Parallel()

		builder := buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		builder = builder.WithTemplateParameter("", nil)

		assert.EqualError(t, builder.GetError(), "provisioningRequest TemplateParameter 'key' cannot be empty")
	})

	t.Run("invalid builder short-circuits", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		builder = builder.WithTemplateParameter("nodeClusterName", "cluster-1")

		assert.EqualError(t, builder.GetError(), errEmptyTemplateName)
	})
}

func TestPRGetTemplateParameters(t *testing.T) {
	t.Parallel()

	t.Run("empty template parameters", func(t *testing.T) {
		t.Parallel()

		builder := buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		templateParameters, err := builder.GetTemplateParameters()

		require.NoError(t, err)
		assert.Empty(t, templateParameters)
	})

	t.Run("unmarshals template parameters", func(t *testing.T) {
		t.Parallel()

		builder := buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		builder.Definition.Spec.TemplateParameters.Raw = []byte(`{"nodeClusterName":"cluster-1"}`)

		templateParameters, err := builder.GetTemplateParameters()

		require.NoError(t, err)
		assert.Equal(t, "cluster-1", templateParameters["nodeClusterName"])
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		templateParameters, err := builder.GetTemplateParameters()

		assert.EqualError(t, err, "failed to validate: "+errEmptyTemplateName)
		assert.Nil(t, templateParameters)
	})
}

func TestPRWithTemplateParameters(t *testing.T) {
	t.Parallel()

	t.Run("nil map becomes empty object", func(t *testing.T) {
		t.Parallel()

		builder := buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		builder = builder.WithTemplateParameters(nil)

		require.NoError(t, builder.GetError())
		assert.JSONEq(t, `{}`, string(builder.Definition.Spec.TemplateParameters.Raw))
	})

	t.Run("marshals map", func(t *testing.T) {
		t.Parallel()

		builder := buildValidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		builder = builder.WithTemplateParameters(map[string]any{"nodeClusterName": "cluster-1"})

		require.NoError(t, builder.GetError())
		assert.JSONEq(t, `{"nodeClusterName":"cluster-1"}`, string(builder.Definition.Spec.TemplateParameters.Raw))
	})

	t.Run("invalid builder short-circuits", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidPRTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		builder = builder.WithTemplateParameters(nil)

		assert.EqualError(t, builder.GetError(), errEmptyTemplateName)
		assert.Empty(t, builder.Definition.Spec.TemplateParameters.Raw)
	})
}

func TestPullPR(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		pullPRForTest,
		provisioningv1alpha1.AddToScheme,
		provisioningRequestGVK,
	).ExecuteTests(t)
}

func TestProvisioningRequestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[
		provisioningv1alpha1.ProvisioningRequest, ProvisioningRequestBuilder](
		provisioningv1alpha1.AddToScheme,
		provisioningRequestGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestPRDeleteAndWait(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		objectExists  bool
		expectedError string
	}{
		{
			name:         "deletes existing resource",
			valid:        true,
			objectExists: true,
		},
		{
			name:         "missing resource is a no-op",
			valid:        true,
			objectExists: false,
		},
		{
			name:          "invalid builder returns error",
			valid:         false,
			objectExists:  true,
			expectedError: "failed to validate: " + errEmptyTemplateName,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.objectExists {
				objects = append(objects, buildDummyPR(defaultPRName))
			}

			settings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  objects,
				SchemeAttachers: provisioningTestSchemes,
			})

			var builder *ProvisioningRequestBuilder
			if testCase.valid {
				builder = buildValidPRTestBuilder(settings)
			} else {
				builder = buildInvalidPRTestBuilder(settings)
			}

			err := builder.DeleteAndWait(time.Second)

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
			} else {
				require.NoError(t, err)
				assert.Nil(t, builder.Object)
			}
		})
	}
}

func TestPRWaitForCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		conditionMet  bool
		exists        bool
		valid         bool
		expectedError string
	}{
		{
			conditionMet: true,
			exists:       true,
			valid:        true,
		},
		{
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded.Error(),
		},
		{
			conditionMet:  true,
			valid:         true,
			expectedError: "cannot wait for non-existent ProvisioningRequest",
		},
		{
			conditionMet:  true,
			exists:        true,
			expectedError: "failed to validate: " + errEmptyTemplateName,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("conditionMet=%t/exists=%t/valid=%t", testCase.conditionMet, testCase.exists, testCase.valid), func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object

			if testCase.exists {
				provisioningRequest := buildDummyPR(defaultPRName)
				if testCase.conditionMet {
					provisioningRequest.Status.Conditions = append(
						provisioningRequest.Status.Conditions, defaultPRCondition)
				}

				objects = append(objects, provisioningRequest)
			}

			settings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  objects,
				SchemeAttachers: provisioningTestSchemes,
			})

			var builder *ProvisioningRequestBuilder
			if testCase.valid {
				builder = buildValidPRTestBuilder(settings)
			} else {
				builder = buildInvalidPRTestBuilder(settings)
			}

			_, err := builder.WaitForCondition(defaultPRCondition, time.Second)

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPRWaitUntilFulfilled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		fulfilled     bool
		exists        bool
		valid         bool
		expectedError string
	}{
		{
			fulfilled: true,
			exists:    true,
			valid:     true,
		},
		{
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded.Error(),
		},
		{
			fulfilled:     true,
			valid:         true,
			expectedError: "cannot wait for non-existent ProvisioningRequest",
		},
		{
			fulfilled:     true,
			exists:        true,
			expectedError: "failed to validate: " + errEmptyTemplateName,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("fulfilled=%t/exists=%t/valid=%t", testCase.fulfilled, testCase.exists, testCase.valid), func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object

			if testCase.exists {
				provisioningRequest := buildDummyPR(defaultPRName)
				if testCase.fulfilled {
					provisioningRequest.Status.ProvisioningStatus.ProvisioningPhase = provisioningv1alpha1.StateFulfilled
				}

				objects = append(objects, provisioningRequest)
			}

			settings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  objects,
				SchemeAttachers: provisioningTestSchemes,
			})

			var builder *ProvisioningRequestBuilder
			if testCase.valid {
				builder = buildValidPRTestBuilder(settings)
			} else {
				builder = buildInvalidPRTestBuilder(settings)
			}

			_, err := builder.WaitUntilFulfilled(time.Second)

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPRWaitForPhaseAfter covers the non-zero update time branch. The remaining WaitForPhaseAfter behavior is covered by
// TestPRWaitUntilFulfilled.
func TestPRWaitForPhaseAfter(t *testing.T) {
	t.Parallel()

	testDummyPR := buildDummyPR(defaultPRName)
	testDummyPR.Status.ProvisioningStatus.ProvisioningPhase = provisioningv1alpha1.StateFulfilled
	testDummyPR.Status.ProvisioningStatus.UpdateTime = metav1.Now()
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{testDummyPR},
		SchemeAttachers: provisioningTestSchemes,
	})
	testBuilder := buildValidPRTestBuilder(testSettings)

	go func() {
		t.Helper()

		// Delay the update long enough for the initial poll to observe the old timestamp.
		select {
		case <-time.After(time.Second):
			testDummyPR.Status.ProvisioningStatus.UpdateTime = metav1.Now()
			err := testSettings.Update(t.Context(), testDummyPR)
			assert.NoError(t, err, "failed to update ProvisioningRequest update time")
		case <-t.Context().Done():
		}
	}()

	err := testBuilder.WaitForPhaseAfter(provisioningv1alpha1.StateFulfilled, time.Now(), 4*time.Second)
	require.NoError(t, err)
}

// pullPRForTest adapts PullPR's runtimeclient.Client parameter to the testhelper's Settings-based pull signature.
func pullPRForTest(apiClient *clients.Settings, name string) (*ProvisioningRequestBuilder, error) {
	return PullPR(apiClient, name)
}

//nolint:unparam
func buildDummyPR(name string) *provisioningv1alpha1.ProvisioningRequest {
	return &provisioningv1alpha1.ProvisioningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: provisioningv1alpha1.ProvisioningRequestSpec{
			TemplateName:    defaultPRTemplateName,
			TemplateVersion: defaultPRTemplateVersion,
		},
	}
}

func buildValidPRTestBuilder(apiClient *clients.Settings) *ProvisioningRequestBuilder {
	return NewPRBuilder(apiClient, defaultPRName, defaultPRTemplateName, defaultPRTemplateVersion)
}

func buildInvalidPRTestBuilder(apiClient *clients.Settings) *ProvisioningRequestBuilder {
	return NewPRBuilder(apiClient, defaultPRName, "", defaultPRTemplateVersion)
}
