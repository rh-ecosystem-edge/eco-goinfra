package ocm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
)

const (
	defaultPlacementBindingName   = "placementbinding-test"
	defaultPlacementBindingNsName = "test-ns"
)

var (
	defaultPlacementBindingRef = policiesv1.PlacementSubject{
		Name:     defaultPlacementRuleName,
		APIGroup: appsOpenClusterManagementIo,
		Kind:     kindPlacementRule,
	}
	defaultPlacementBindingSubject = policiesv1.Subject{
		Name:     defaultPolicySetName,
		APIGroup: policyOpenClusterManagementIo,
		Kind:     kindPolicySet,
	}
)

func TestNewPlacementBindingBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *PlacementBindingBuilder {
				return NewPlacementBindingBuilder(
					apiClient, name, nsname, defaultPlacementBindingRef, defaultPlacementBindingSubject)
			},
			policiesv1.AddToScheme,
			placementBindingGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		placementRef  policiesv1.PlacementSubject
		subject       policiesv1.Subject
		expectedError error
	}{
		{
			name:          "valid refs",
			placementRef:  defaultPlacementBindingRef,
			subject:       defaultPlacementBindingSubject,
			expectedError: nil,
		},
		{
			name: "invalid placement ref",
			placementRef: policiesv1.PlacementSubject{
				Name:     defaultPlacementRuleName,
				APIGroup: "",
				Kind:     kindPlacementRule,
			},
			subject:       defaultPlacementBindingSubject,
			expectedError: fmt.Errorf("placementBinding's 'PlacementRef.APIGroup' must be a valid option"),
		},
		{
			name:         "invalid subject",
			placementRef: defaultPlacementBindingRef,
			subject: policiesv1.Subject{
				Name:     "",
				APIGroup: policyOpenClusterManagementIo,
				Kind:     kindPolicySet,
			},
			expectedError: fmt.Errorf(errEmptySubjectName),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := NewPlacementBindingBuilder(
				clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: []clients.SchemeAttacher{policiesv1.AddToScheme},
				}),
				defaultPlacementBindingName,
				defaultPlacementBindingNsName,
				testCase.placementRef,
				testCase.subject,
			)

			assert.Equal(t, testCase.expectedError, testBuilder.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.placementRef, testBuilder.Definition.PlacementRef)
				assert.Equal(t, []policiesv1.Subject{testCase.subject}, testBuilder.Definition.Subjects)
			}
		})
	}
}

func TestPullPlacementBinding(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullPlacementBinding, policiesv1.AddToScheme, placementBindingGVK).ExecuteTests(t)
}

func TestPlacementBindingBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[policiesv1.PlacementBinding, PlacementBindingBuilder](
		policiesv1.AddToScheme,
		placementBindingGVK,
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

func TestPlacementBindingWithAdditionalSubject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		subject       policiesv1.Subject
		valid         bool
		expectedError error
	}{
		{
			name:    "valid subject",
			subject: defaultPlacementBindingSubject,
			valid:   true,
		},
		{
			name: "empty subject name",
			subject: policiesv1.Subject{
				Name:     "",
				APIGroup: policyOpenClusterManagementIo,
				Kind:     kindPolicySet,
			},
			valid:         true,
			expectedError: fmt.Errorf(errEmptySubjectName),
		},
		{
			name:    "invalid builder",
			subject: defaultPlacementBindingSubject,
			valid:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var placementBindingBuilder *PlacementBindingBuilder
			if testCase.valid {
				placementBindingBuilder = buildValidPlacementBindingTestBuilder(buildTestClientWithPlacementBindingScheme())
			} else {
				placementBindingBuilder = buildInvalidPlacementBindingTestBuilder(buildTestClientWithPlacementBindingScheme())
			}

			placementBindingBuilder = placementBindingBuilder.WithAdditionalSubject(testCase.subject)

			switch testCase.name {
			case "invalid builder":
				require.Error(t, placementBindingBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(placementBindingBuilder.GetError()))
			default:
				assert.Equal(t, testCase.expectedError, placementBindingBuilder.GetError())

				if testCase.expectedError == nil {
					assert.Equal(
						t,
						[]policiesv1.Subject{defaultPlacementBindingSubject, testCase.subject},
						placementBindingBuilder.Definition.Subjects,
					)
				} else {
					assert.Equal(
						t,
						[]policiesv1.Subject{defaultPlacementBindingSubject},
						placementBindingBuilder.Definition.Subjects,
					)
				}
			}
		})
	}
}

func TestValidatePlacementRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		ref               policiesv1.PlacementSubject
		expectedErrorText string
	}{
		{
			ref:               defaultPlacementBindingRef,
			expectedErrorText: "",
		},
		{
			ref: policiesv1.PlacementSubject{
				Name:     "",
				APIGroup: appsOpenClusterManagementIo,
				Kind:     kindPlacementRule,
			},
			expectedErrorText: "placementBinding's 'PlacementRef.Name' cannot be empty",
		},
		{
			ref: policiesv1.PlacementSubject{
				Name:     defaultPlacementRuleName,
				APIGroup: "",
				Kind:     kindPlacementRule,
			},
			expectedErrorText: "placementBinding's 'PlacementRef.APIGroup' must be a valid option",
		},
		{
			ref: policiesv1.PlacementSubject{
				Name:     defaultPlacementRuleName,
				APIGroup: appsOpenClusterManagementIo,
				Kind:     "",
			},
			expectedErrorText: "placementBinding's 'PlacementRef.Kind' must be a valid option",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.expectedErrorText, func(t *testing.T) {
			t.Parallel()

			err := validatePlacementRef(testCase.ref)
			assert.Equal(t, testCase.expectedErrorText, err)
		})
	}
}

func TestValidateSubject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		subject           policiesv1.Subject
		expectedErrorText string
	}{
		{
			subject:           defaultPlacementBindingSubject,
			expectedErrorText: "",
		},
		{
			subject: policiesv1.Subject{
				Name:     "",
				APIGroup: policyOpenClusterManagementIo,
				Kind:     kindPolicySet,
			},
			expectedErrorText: errEmptySubjectName,
		},
		{
			subject: policiesv1.Subject{
				Name:     defaultPolicySetName,
				APIGroup: "",
				Kind:     kindPolicySet,
			},
			expectedErrorText: "placementBinding's 'Subject.APIGroup' must be 'policy.open-cluster-management.io'",
		},
		{
			subject: policiesv1.Subject{
				Name:     defaultPolicySetName,
				APIGroup: policyOpenClusterManagementIo,
				Kind:     "",
			},
			expectedErrorText: "placementBinding's 'Subject.Kind' must be a valid option",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.expectedErrorText, func(t *testing.T) {
			t.Parallel()

			err := validateSubject(testCase.subject)
			assert.Equal(t, testCase.expectedErrorText, err)
		})
	}
}

// buildDummyPlacementBinding returns a PlacementBinding with the provided name and namespace.
func buildDummyPlacementBinding(name, nsname string) *policiesv1.PlacementBinding {
	return &policiesv1.PlacementBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		PlacementRef: defaultPlacementBindingRef,
		Subjects:     []policiesv1.Subject{defaultPlacementBindingSubject},
	}
}

// buildTestClientWithDummyPlacementBinding returns a client with a mock dummy PlacementBinding.
func buildTestClientWithDummyPlacementBinding() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPlacementBinding(defaultPlacementBindingName, defaultPlacementBindingNsName),
		},
		SchemeAttachers: []clients.SchemeAttacher{
			policiesv1.AddToScheme,
		},
	})
}

// buildTestClientWithPlacementBindingScheme returns a client with no objects but the PlacementBinding scheme attached.
func buildTestClientWithPlacementBindingScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{
			policiesv1.AddToScheme,
		},
	})
}

// buildValidPlacementBindingTestBuilder returns a valid PlacementBindingBuilder for testing.
func buildValidPlacementBindingTestBuilder(apiClient *clients.Settings) *PlacementBindingBuilder {
	return NewPlacementBindingBuilder(
		apiClient,
		defaultPlacementBindingName,
		defaultPlacementBindingNsName,
		defaultPlacementBindingRef,
		defaultPlacementBindingSubject)
}

// buildInvalidPlacementBindingTestBuilder returns an invalid PlacementBindingBuilder for testing.
func buildInvalidPlacementBindingTestBuilder(apiClient *clients.Settings) *PlacementBindingBuilder {
	return NewPlacementBindingBuilder(
		apiClient, defaultPlacementBindingName, "", defaultPlacementBindingRef, defaultPlacementBindingSubject)
}
