package resource

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var resourceSliceGVK = resourcev1.SchemeGroupVersion.WithKind("ResourceSlice")

func TestListResourceSlices(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListResourceSlices,
		resourcev1.AddToScheme,
		resourceSliceGVK,
	).ExecuteTests(t)
}

func TestResourceSliceBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[resourcev1.ResourceSlice, ResourceSliceBuilder](
		resourcev1.AddToScheme,
		resourceSliceGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}

func TestListResourceSlicesByDriver(t *testing.T) {
	testCases := []struct {
		name          string
		driverName    string
		addSlices     bool
		client        bool
		expectedCount int
		expectedError bool
	}{
		{
			name:          "matching driver",
			driverName:    "test-driver",
			addSlices:     true,
			client:        true,
			expectedCount: 1,
		},
		{
			name:          "non-matching driver",
			driverName:    "other-driver",
			addSlices:     true,
			client:        true,
			expectedCount: 0,
		},
		{
			name:          "empty driverName",
			driverName:    "",
			client:        true,
			expectedError: true,
		},
		{
			name:          "nil apiClient",
			driverName:    "test-driver",
			client:        false,
			expectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				runtimeObjects []runtime.Object
				testSettings   *clients.Settings
			)

			if testCase.addSlices {
				runtimeObjects = append(runtimeObjects,
					generateResourceSlice("test-slice", "test-driver"))
			}

			if testCase.client {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  runtimeObjects,
					SchemeAttachers: testResourceSchemes,
				})
			}

			builders, err := ListResourceSlicesByDriver(testSettings, testCase.driverName)

			if testCase.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, builders, testCase.expectedCount)
			}
		})
	}
}

func TestResourceSliceBuilderGetGVK(t *testing.T) {
	builder := &ResourceSliceBuilder{}
	assert.Equal(t, resourceSliceGVK, builder.GetGVK())
}

func generateResourceSlice(name, driverName string) *resourcev1.ResourceSlice {
	return &resourcev1.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: resourcev1.ResourceSliceSpec{
			Driver: driverName,
		},
	}
}
