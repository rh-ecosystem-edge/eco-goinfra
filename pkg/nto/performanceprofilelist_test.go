package nto //nolint:misspell

import (
	"context"
	"errors"
	"fmt"
	"testing"

	performanceprofilev2 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/performanceprofile/v2"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	testProfileLabelKey  = "role"
	testProfileMatchName = "matching-profile"
)

var (
	errSimulatedDelete = errors.New("simulated delete failure")

	matchingProfileSelector = goclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{testProfileLabelKey: "match"}),
	}
)

func TestListProfiles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		apiClient     *clients.Settings
		listOptions   []goclient.ListOptions
		expectedError error
		expectedNames []string
	}{
		{
			name:          "nil apiClient",
			apiClient:     nil,
			expectedError: fmt.Errorf("the apiClient cannot be nil"),
		},
		{
			name:      "no profiles",
			apiClient: newPerformanceProfileTestClient(),
		},
		{
			name:          "one profile",
			apiClient:     buildPerformanceProfileWithDummyObject(),
			expectedNames: []string{defaultPerformanceProfileName},
		},
		{
			name:          "label selector filters profiles",
			apiClient:     newPerformanceProfileTestClient(buildLabeledTestPerformanceProfiles()...),
			listOptions:   []goclient.ListOptions{matchingProfileSelector},
			expectedNames: []string{testProfileMatchName},
		},
		{
			name:          "too many list options",
			apiClient:     buildPerformanceProfileWithDummyObject(),
			listOptions:   []goclient.ListOptions{{}, {}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := ListProfiles(testCase.apiClient, testCase.listOptions...)
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				var names []string
				for _, builder := range result {
					names = append(names, builder.Object.Name)
				}

				assert.ElementsMatch(t, testCase.expectedNames, names)
			}
		})
	}
}

func TestCleanAllPerformanceProfiles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		apiClient      *clients.Settings
		listOptions    []goclient.ListOptions
		expectedError  error
		remainingNames []string
	}{
		{
			name:          "nil apiClient",
			apiClient:     nil,
			expectedError: fmt.Errorf("the apiClient cannot be nil"),
		},
		{
			name:      "no profiles",
			apiClient: newPerformanceProfileTestClient(),
		},
		{
			name:      "one profile deleted",
			apiClient: buildPerformanceProfileWithDummyObject(),
		},
		{
			name:           "label selector deletes only matching profiles",
			apiClient:      newPerformanceProfileTestClient(buildLabeledTestPerformanceProfiles()...),
			listOptions:    []goclient.ListOptions{matchingProfileSelector},
			remainingNames: []string{defaultPerformanceProfileName},
		},
		{
			name:          "too many list options",
			apiClient:     buildPerformanceProfileWithDummyObject(),
			listOptions:   []goclient.ListOptions{{}, {}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			name: "delete failure",
			apiClient: clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyPerformanceProfile(),
				SchemeAttachers: paoTestSchemes,
				InterceptorFuncs: interceptor.Funcs{Delete: func(
					context.Context, goclient.WithWatch, goclient.Object, ...goclient.DeleteOption) error {
					return errSimulatedDelete
				}},
			}),
			expectedError: errSimulatedDelete,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := CleanAllPerformanceProfiles(testCase.apiClient, testCase.listOptions...)

			if errors.Is(testCase.expectedError, errSimulatedDelete) {
				assert.ErrorIs(t, err, testCase.expectedError)
				assert.True(t, commonerrors.IsAPICallFailedWithVerb(err, "delete"))
			} else {
				assert.Equal(t, testCase.expectedError, err)
			}

			if testCase.expectedError == nil {
				var remaining performanceprofilev2.PerformanceProfileList

				require.NoError(t, testCase.apiClient.List(context.TODO(), &remaining))

				var names []string
				for _, profile := range remaining.Items {
					names = append(names, profile.Name)
				}

				assert.ElementsMatch(t, testCase.remainingNames, names)
			}
		})
	}
}

// buildLabeledTestPerformanceProfiles returns two profiles of which only testProfileMatchName matches
// matchingProfileSelector.
func buildLabeledTestPerformanceProfiles() []runtime.Object {
	return []runtime.Object{
		&performanceprofilev2.PerformanceProfile{ObjectMeta: metav1.ObjectMeta{
			Name: defaultPerformanceProfileName, Labels: map[string]string{testProfileLabelKey: "other"}}},
		&performanceprofilev2.PerformanceProfile{ObjectMeta: metav1.ObjectMeta{
			Name: testProfileMatchName, Labels: map[string]string{testProfileLabelKey: "match"}}},
	}
}

func newPerformanceProfileTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: paoTestSchemes,
	})
}
