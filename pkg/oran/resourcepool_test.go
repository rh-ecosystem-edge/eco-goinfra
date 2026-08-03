package oran

import (
	"context"
	"fmt"
	"testing"
	"time"

	inventoryv1alpha1 "github.com/openshift-kni/oran-o2ims/api/inventory/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testResourcePoolName      = "test-resourcepool"
	testResourcePoolNamespace = "test-namespace"
	testOCloudSiteName        = "test-ocloudsite"
)

var resourcePoolGVK = inventoryv1alpha1.GroupVersion.WithKind("ResourcePool")

var (
	defaultResourcePoolCondition = metav1.Condition{
		Type:   inventoryv1alpha1.ConditionTypeReady,
		Status: metav1.ConditionTrue,
		Reason: inventoryv1alpha1.ReasonReady,
	}

	inventoryTestSchemes = []clients.SchemeAttacher{
		inventoryv1alpha1.AddToScheme,
	}

	errResourcePoolNameEmpty = commonerrors.NewBuilderFieldEmpty(
		key.NewResourceKey("ResourcePool", "", testResourcePoolNamespace),
		commonerrors.BuilderFieldName,
	)
)

func TestNewResourcePoolBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewResourcePoolBuilder,
		inventoryv1alpha1.AddToScheme,
		resourcePoolGVK,
	).ExecuteTests(t)
}

func TestPullResourcePool(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullResourcePool,
		inventoryv1alpha1.AddToScheme,
		resourcePoolGVK,
	).ExecuteTests(t)
}

func TestListResourcePools(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListResourcePools,
		inventoryv1alpha1.AddToScheme,
		resourcePoolGVK,
	).ExecuteTests(t)
}

func TestListReadyResourcePools(t *testing.T) {
	t.Parallel()

	const (
		readyResourcePoolName    = "ready-resourcepool"
		notReadyResourcePoolName = "not-ready-resourcepool"
	)

	readyResourcePool := buildDummyResourcePool(readyResourcePoolName, testResourcePoolNamespace)
	readyResourcePool.Status.Conditions = append(readyResourcePool.Status.Conditions, defaultResourcePoolCondition)

	notReadyResourcePool := buildDummyResourcePool(notReadyResourcePoolName, testResourcePoolNamespace)

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			readyResourcePool,
			notReadyResourcePool,
		},
		SchemeAttachers: inventoryTestSchemes,
	})

	resourcePools, err := ListReadyResourcePools(testSettings)
	require.NoError(t, err)
	require.Len(t, resourcePools, 1)
	assert.Equal(t, readyResourcePoolName, resourcePools[0].Definition.Name)
	assert.Equal(t, testResourcePoolNamespace, resourcePools[0].Definition.Namespace)
}

func TestResourcePoolMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[inventoryv1alpha1.ResourcePool, ResourcePoolBuilder](
		inventoryv1alpha1.AddToScheme,
		resourcePoolGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestResourcePoolWithOCloudSiteName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *ResourcePoolBuilder
		oCloudSiteName string
		expectedError  error
		expectedValue  string
	}{
		{
			name:           "sets oCloudSiteName on valid builder",
			builder:        newValidResourcePoolBuilder(newResourcePoolTestClient()),
			oCloudSiteName: testOCloudSiteName,
			expectedValue:  testOCloudSiteName,
		},
		{
			name:           "empty oCloudSiteName sets builder error",
			builder:        newValidResourcePoolBuilder(newResourcePoolTestClient()),
			oCloudSiteName: "",
			expectedError:  fmt.Errorf("resourcePool 'oCloudSiteName' cannot be empty"),
		},
		{
			name:           "invalid builder short circuits",
			builder:        newInvalidResourcePoolBuilder(),
			oCloudSiteName: testOCloudSiteName,
			expectedError:  errResourcePoolNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, testCase.builder)

			result := testCase.builder.WithOCloudSiteName(testCase.oCloudSiteName)
			require.Same(t, testCase.builder, result)
			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.expectedValue, result.Definition.Spec.OCloudSiteName)
			}
		})
	}
}

func TestResourcePoolWithDescription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       *ResourcePoolBuilder
		description   string
		expectedError error
		expectedValue string
	}{
		{
			name:          "sets description on valid builder",
			builder:       newValidResourcePoolBuilder(newResourcePoolTestClient()),
			description:   "test description",
			expectedValue: "test description",
		},
		{
			name:          "invalid builder short circuits",
			builder:       newInvalidResourcePoolBuilder(),
			description:   "test description",
			expectedError: errResourcePoolNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, testCase.builder)

			result := testCase.builder.WithDescription(testCase.description)
			require.Same(t, testCase.builder, result)
			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.expectedValue, result.Definition.Spec.Description)
			}
		})
	}
}

func TestResourcePoolWaitForCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		conditionMet  bool
		exists        bool
		expectedError error
	}{
		{
			name:          "condition met",
			conditionMet:  true,
			exists:        true,
			expectedError: nil,
		},
		{
			name:          "condition not met",
			conditionMet:  false,
			exists:        true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "resourcepool does not exist",
			conditionMet:  true,
			exists:        false,
			expectedError: fmt.Errorf("cannot wait for non-existent ResourcePool"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				resourcePool := buildDummyResourcePool(testResourcePoolName, testResourcePoolNamespace)
				if testCase.conditionMet {
					resourcePool.Status.Conditions = append(resourcePool.Status.Conditions, defaultResourcePoolCondition)
				}

				runtimeObjects = append(runtimeObjects, resourcePool)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: inventoryTestSchemes,
			})
			testBuilder := newValidResourcePoolBuilder(testSettings)

			_, err := testBuilder.WaitForCondition(defaultResourcePoolCondition, time.Second)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func buildDummyResourcePool(name, nsname string) *inventoryv1alpha1.ResourcePool {
	return &inventoryv1alpha1.ResourcePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

func newResourcePoolTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: inventoryTestSchemes,
	})
}

func newValidResourcePoolBuilder(apiClient *clients.Settings) *ResourcePoolBuilder {
	return NewResourcePoolBuilder(apiClient, testResourcePoolName, testResourcePoolNamespace)
}

func newInvalidResourcePoolBuilder() *ResourcePoolBuilder {
	return NewResourcePoolBuilder(newResourcePoolTestClient(), "", testResourcePoolNamespace)
}
