package olm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	oplmV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPullClusterServiceVersion(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullClusterServiceVersion,
		oplmV1alpha1.AddToScheme,
		clusterServiceVersionGVK,
	).ExecuteTests(t)
}

func TestClusterServiceVersionBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[oplmV1alpha1.ClusterServiceVersion,
		ClusterServiceVersionBuilder](oplmV1alpha1.AddToScheme, clusterServiceVersionGVK,
		testhelper.ResourceScopeNamespaced)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestClusterServiceGetAlmExamples(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		clusterService *ClusterServiceVersionBuilder
		expectedValue  string
		expectedError  string
	}{
		{
			name: "alm-examples annotation present",
			clusterService: buildValidClusterServiceBuilder(
				buildTestClientWithDummyClusterServiceObjectWithAlm(map[string]string{
					annotationAlmExamples: "test",
				})),
			expectedValue: "test",
		},
		{
			name: "resource does not exist",
			clusterService: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: "alm-examples not found in given clusterserviceversion named clusterservice",
		},
		{
			name:           "alm-examples annotation absent",
			clusterService: buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject()),
			expectedError:  "alm-examples not found in given clusterserviceversion named clusterservice",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			almExamples, err := testCase.clusterService.GetAlmExamples()

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedValue, almExamples)
		})
	}
}

func TestClusterServiceGetPhase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		clusterService *ClusterServiceVersionBuilder
		expectedPhase  oplmV1alpha1.ClusterServiceVersionPhase
		expectedError  string
	}{
		{
			name: "phase present",
			clusterService: buildValidClusterServiceBuilder(
				buildTestClientWithDummyClusterServiceObjectWithPhase(oplmV1alpha1.CSVPhaseSucceeded)),
			expectedPhase: oplmV1alpha1.CSVPhaseSucceeded,
		},
		{
			name: "resource does not exist",
			clusterService: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: "clusterservice clusterserviceversion not found in test-namespace namespace",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			phase, err := testCase.clusterService.GetPhase()

			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedPhase, phase)
		})
	}
}

func TestClusterServiceIsSuccessful(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		clusterService *ClusterServiceVersionBuilder
		successful     bool
		expectedError  string
	}{
		{
			name: "succeeded phase",
			clusterService: buildValidClusterServiceBuilder(
				buildTestClientWithDummyClusterServiceObjectWithPhase(oplmV1alpha1.CSVPhaseSucceeded)),
			successful: true,
		},
		{
			name: "failed phase",
			clusterService: buildValidClusterServiceBuilder(
				buildTestClientWithDummyClusterServiceObjectWithPhase(oplmV1alpha1.CSVPhaseFailed)),
			successful: false,
		},
		{
			name: "resource does not exist",
			clusterService: buildValidClusterServiceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: "failed to get phase value for clusterservice clusterserviceversion in " +
				"test-namespace namespace",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			successful, err := testCase.clusterService.IsSuccessful()

			if testCase.expectedError != "" {
				assert.ErrorContains(t, err, testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.successful, successful)
		})
	}
}

func buildValidClusterServiceBuilder(apiClient *clients.Settings) *ClusterServiceVersionBuilder {
	return common.NewNamespacedBuilder[oplmV1alpha1.ClusterServiceVersion, ClusterServiceVersionBuilder](
		apiClient, oplmV1alpha1.AddToScheme, "clusterservice", "test-namespace")
}

func buildTestClientWithDummyClusterServiceObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterService(nil, ""),
		SchemeAttachers: testSchemes,
	})
}

func buildTestClientWithDummyClusterServiceObjectWithAlm(almExamples map[string]string) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterServiceWithAlm(almExamples),
		SchemeAttachers: testSchemes,
	})
}

func buildTestClientWithDummyClusterServiceObjectWithPhase(
	phase oplmV1alpha1.ClusterServiceVersionPhase) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterServiceWithPhase(phase),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyClusterServiceWithAlm(almExamples map[string]string) []runtime.Object {
	return buildDummyClusterService(almExamples, "")
}

func buildDummyClusterServiceWithPhase(phase oplmV1alpha1.ClusterServiceVersionPhase) []runtime.Object {
	return buildDummyClusterService(nil, phase)
}

func buildDummyClusterService(
	almExamples map[string]string, phase oplmV1alpha1.ClusterServiceVersionPhase) []runtime.Object {
	csv := oplmV1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "clusterservice",
			Namespace: "test-namespace",
		},
		Spec: oplmV1alpha1.ClusterServiceVersionSpec{
			DisplayName: "test",
		},
		Status: oplmV1alpha1.ClusterServiceVersionStatus{
			Phase: oplmV1alpha1.CSVPhaseSucceeded,
		},
	}

	if len(almExamples) > 0 {
		csv.Annotations = almExamples
	}

	if phase != "" {
		csv.Status.Phase = phase
	}

	return append([]runtime.Object{}, &csv)
}
