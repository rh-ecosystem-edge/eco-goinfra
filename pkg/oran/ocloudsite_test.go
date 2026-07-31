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
	testOCloudSiteName      = "test-ocloudsite"
	testOCloudSiteNamespace = "test-namespace"
	testGlobalLocationName  = "test-location"
)

var ocloudSiteGVK = inventoryv1alpha1.GroupVersion.WithKind("OCloudSite")

var (
	defaultOCloudSiteCondition = metav1.Condition{
		Type:   inventoryv1alpha1.ConditionTypeReady,
		Status: metav1.ConditionTrue,
		Reason: inventoryv1alpha1.ReasonReady,
	}

	inventoryTestSchemes = []clients.SchemeAttacher{
		inventoryv1alpha1.AddToScheme,
	}

	errOCloudSiteNameEmpty = commonerrors.NewBuilderFieldEmpty(
		key.NewResourceKey("OCloudSite", "", testOCloudSiteNamespace),
		commonerrors.BuilderFieldName,
	)
)

func TestNewOCloudSiteBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewOCloudSiteBuilder,
		inventoryv1alpha1.AddToScheme,
		ocloudSiteGVK,
	).ExecuteTests(t)
}

func TestPullOCloudSite(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullOCloudSite,
		inventoryv1alpha1.AddToScheme,
		ocloudSiteGVK,
	).ExecuteTests(t)
}

func TestListOCloudSites(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListOCloudSites,
		inventoryv1alpha1.AddToScheme,
		ocloudSiteGVK,
	).ExecuteTests(t)
}

func TestOCloudSiteMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[inventoryv1alpha1.OCloudSite, OCloudSiteBuilder](
		inventoryv1alpha1.AddToScheme,
		ocloudSiteGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestOCloudSiteWithGlobalLocationName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		builder            *OCloudSiteBuilder
		globalLocationName string
		expectedError      error
		expectedValue      string
	}{
		{
			name:               "sets globalLocationName on valid builder",
			builder:            newValidOCloudSiteBuilder(newOCloudSiteTestClient()),
			globalLocationName: testGlobalLocationName,
			expectedValue:      testGlobalLocationName,
		},
		{
			name:               "empty globalLocationName sets builder error",
			builder:            newValidOCloudSiteBuilder(newOCloudSiteTestClient()),
			globalLocationName: "",
			expectedError:      fmt.Errorf("oCloudSite 'globalLocationName' cannot be empty"),
		},
		{
			name:               "invalid builder short circuits",
			builder:            newInvalidOCloudSiteBuilder(),
			globalLocationName: testGlobalLocationName,
			expectedError:      errOCloudSiteNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, testCase.builder)

			result := testCase.builder.WithGlobalLocationName(testCase.globalLocationName)
			require.Same(t, testCase.builder, result)
			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.expectedValue, result.Definition.Spec.GlobalLocationName)
			}
		})
	}
}

func TestOCloudSiteWithDescription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       *OCloudSiteBuilder
		description   string
		expectedError error
		expectedValue string
	}{
		{
			name:          "sets description on valid builder",
			builder:       newValidOCloudSiteBuilder(newOCloudSiteTestClient()),
			description:   "test description",
			expectedValue: "test description",
		},
		{
			name:          "invalid builder short circuits",
			builder:       newInvalidOCloudSiteBuilder(),
			description:   "test description",
			expectedError: errOCloudSiteNameEmpty,
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

func TestOCloudSiteWaitForCondition(t *testing.T) {
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
			name:          "ocloudsite does not exist",
			conditionMet:  true,
			exists:        false,
			expectedError: fmt.Errorf("cannot wait for non-existent OCloudSite"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				ocloudSite := buildDummyOCloudSite(testOCloudSiteName, testOCloudSiteNamespace)
				if testCase.conditionMet {
					ocloudSite.Status.Conditions = append(ocloudSite.Status.Conditions, defaultOCloudSiteCondition)
				}

				runtimeObjects = append(runtimeObjects, ocloudSite)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: inventoryTestSchemes,
			})
			testBuilder := newValidOCloudSiteBuilder(testSettings)

			_, err := testBuilder.WaitForCondition(defaultOCloudSiteCondition, time.Second)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func buildDummyOCloudSite(name, nsname string) *inventoryv1alpha1.OCloudSite {
	return &inventoryv1alpha1.OCloudSite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

func newOCloudSiteTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: inventoryTestSchemes,
	})
}

func newValidOCloudSiteBuilder(apiClient *clients.Settings) *OCloudSiteBuilder {
	return NewOCloudSiteBuilder(apiClient, testOCloudSiteName, testOCloudSiteNamespace)
}

func newInvalidOCloudSiteBuilder() *OCloudSiteBuilder {
	return NewOCloudSiteBuilder(newOCloudSiteTestClient(), "", testOCloudSiteNamespace)
}
