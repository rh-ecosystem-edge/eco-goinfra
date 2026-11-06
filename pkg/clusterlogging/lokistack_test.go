package clusterlogging

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	lokiv1 "github.com/grafana/loki/operator/apis/loki/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultLokiStackName      = "lokistack-test"
	defaultLokiStackNamespace = "lokistack-space"
	lokiStackComponents       = &lokiv1.LokiComponentSpec{
		NodeSelector: map[string]string{
			"node-role.kubernetes.io/infra": "",
		},
		Tolerations: []corev1.Toleration{{
			Key:      "node-role.kubernetes.io/infra",
			Operator: "Exists",
		}},
	}
	lokiv1TestSchemes = []clients.SchemeAttacher{
		lokiv1.AddToScheme,
	}
)

func TestNewLokiStackBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewLokiStackBuilder,
		lokiv1.AddToScheme,
		lokiStackGVK,
	).ExecuteTests(t)
}

func TestPullLokiStack(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullLokiStack,
		lokiv1.AddToScheme,
		lokiStackGVK,
	).ExecuteTests(t)
}

func TestLokiStackBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[lokiv1.LokiStack, LokiStackBuilder](
		lokiv1.AddToScheme,
		lokiStackGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestLokiStackWithSize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		testSize       lokiv1.LokiStackSizeType
		expectedError  string
		invalidBuilder bool
	}{
		{
			name:     "sets one-times-demo size",
			testSize: lokiv1.SizeOneXDemo,
		},
		{
			name:     "sets one-times-small size",
			testSize: lokiv1.SizeOneXSmall,
		},
		{
			name:     "sets one-times-medium size",
			testSize: lokiv1.SizeOneXMedium,
		},
		{
			name:     "sets one-times-extra-small size",
			testSize: lokiv1.SizeOneXExtraSmall,
		},
		{
			name:          "empty size returns error",
			expectedError: "'size' argument cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			testSize:       lokiv1.SizeOneXDemo,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildLokiStackTestBuilder(buildLokiStackClientWithDummyObject(), testCase.invalidBuilder)
			result := testBuilder.WithSize(testCase.testSize)

			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				requireInvalidLokiStackBuilderError(t, result.GetError())

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.testSize, result.Definition.Spec.Size)
		})
	}
}

func TestLokiStackWithStorage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		testStorage    lokiv1.ObjectStorageSpec
		invalidBuilder bool
	}{
		{
			name: "sets storage",
			testStorage: lokiv1.ObjectStorageSpec{
				Secret: lokiv1.ObjectStorageSecretSpec{
					Type: "s3",
					Name: "test",
				},
			},
		},
		{
			name: "invalid builder short circuits",
			testStorage: lokiv1.ObjectStorageSpec{
				Secret: lokiv1.ObjectStorageSecretSpec{
					Type: "s3",
					Name: "test",
				},
			},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildLokiStackTestBuilder(buildLokiStackClientWithDummyObject(), testCase.invalidBuilder)
			result := testBuilder.WithStorage(testCase.testStorage)

			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				requireInvalidLokiStackBuilderError(t, result.GetError())

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.testStorage, result.Definition.Spec.Storage)
		})
	}
}

func TestLokiStackWithStorageClassName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		testStorageClass string
		expectedError    string
		invalidBuilder   bool
	}{
		{
			name:             "sets storage class name",
			testStorageClass: "gp2",
		},
		{
			name:          "empty storage class name returns error",
			expectedError: "'storageClassName' argument cannot be empty",
		},
		{
			name:             "invalid builder short circuits",
			testStorageClass: "gp2",
			invalidBuilder:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildLokiStackTestBuilder(buildLokiStackClientWithDummyObject(), testCase.invalidBuilder)
			result := testBuilder.WithStorageClassName(testCase.testStorageClass)

			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				requireInvalidLokiStackBuilderError(t, result.GetError())

				return
			}

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.testStorageClass, result.Definition.Spec.StorageClassName)
		})
	}
}

func TestLokiStackWithTenants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		testTenants    lokiv1.TenantsSpec
		invalidBuilder bool
	}{
		{
			name: "sets tenants",
			testTenants: lokiv1.TenantsSpec{
				Mode: "openshift-logging",
			},
		},
		{
			name: "invalid builder short circuits",
			testTenants: lokiv1.TenantsSpec{
				Mode: "openshift-logging",
			},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildLokiStackTestBuilder(buildLokiStackClientWithDummyObject(), testCase.invalidBuilder)
			result := testBuilder.WithTenants(testCase.testTenants)

			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				requireInvalidLokiStackBuilderError(t, result.GetError())

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Tenants)
			assert.Equal(t, testCase.testTenants, *result.Definition.Spec.Tenants)
		})
	}
}

func TestLokiStackWithRules(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		testRules      lokiv1.RulesSpec
		invalidBuilder bool
	}{
		{
			name: "sets enabled rules",
			testRules: lokiv1.RulesSpec{
				Enabled: true,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"openshift.io/cluster-monitoring": "true"},
				},
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"openshift.io/cluster-monitoring": "true"},
				},
			},
		},
		{
			name: "sets disabled rules",
			testRules: lokiv1.RulesSpec{
				Enabled: false,
			},
		},
		{
			name: "invalid builder short circuits",
			testRules: lokiv1.RulesSpec{
				Enabled: true,
			},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildLokiStackTestBuilder(buildLokiStackClientWithDummyObject(), testCase.invalidBuilder)
			result := testBuilder.WithRules(testCase.testRules)

			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				requireInvalidLokiStackBuilderError(t, result.GetError())

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Rules)
			assert.Equal(t, testCase.testRules, *result.Definition.Spec.Rules)
		})
	}
}

func TestLokiStackWithManagementState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		testManagementState lokiv1.ManagementStateType
		invalidBuilder      bool
	}{
		{
			name:                "sets managed state",
			testManagementState: lokiv1.ManagementStateManaged,
		},
		{
			name:                "sets unmanaged state",
			testManagementState: lokiv1.ManagementStateUnmanaged,
		},
		{
			name:                "invalid builder short circuits",
			testManagementState: lokiv1.ManagementStateManaged,
			invalidBuilder:      true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildLokiStackTestBuilder(buildLokiStackClientWithDummyObject(), testCase.invalidBuilder)
			result := testBuilder.WithManagementState(testCase.testManagementState)

			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				requireInvalidLokiStackBuilderError(t, result.GetError())

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.testManagementState, result.Definition.Spec.ManagementState)
		})
	}
}

func TestLokiStackWithLimits(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		testLimits     lokiv1.LimitsSpec
		invalidBuilder bool
	}{
		{
			name: "sets limits",
			testLimits: lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					Retention: &lokiv1.RetentionLimitSpec{
						Days: 7,
					},
				},
			},
		},
		{
			name:       "sets empty limits",
			testLimits: lokiv1.LimitsSpec{},
		},
		{
			name: "invalid builder short circuits",
			testLimits: lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{},
			},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildLokiStackTestBuilder(buildLokiStackClientWithDummyObject(), testCase.invalidBuilder)
			result := testBuilder.WithLimits(testCase.testLimits)

			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				requireInvalidLokiStackBuilderError(t, result.GetError())

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Limits)
			assert.Equal(t, testCase.testLimits, *result.Definition.Spec.Limits)
		})
	}
}

func TestLokiStackWithTemplate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		testTemplate   lokiv1.LokiTemplateSpec
		invalidBuilder bool
	}{
		{
			name: "sets component template",
			testTemplate: lokiv1.LokiTemplateSpec{
				Compactor:     lokiStackComponents,
				Distributor:   lokiStackComponents,
				Ingester:      lokiStackComponents,
				Querier:       lokiStackComponents,
				QueryFrontend: lokiStackComponents,
				Gateway:       lokiStackComponents,
				IndexGateway:  lokiStackComponents,
				Ruler:         lokiStackComponents,
			},
		},
		{
			name:         "sets empty template",
			testTemplate: lokiv1.LokiTemplateSpec{},
		},
		{
			name: "invalid builder short circuits",
			testTemplate: lokiv1.LokiTemplateSpec{
				Compactor: lokiStackComponents,
			},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildLokiStackTestBuilder(buildLokiStackClientWithDummyObject(), testCase.invalidBuilder)
			result := testBuilder.WithTemplate(testCase.testTemplate)

			require.Same(t, testBuilder, result)

			if testCase.invalidBuilder {
				requireInvalidLokiStackBuilderError(t, result.GetError())

				return
			}

			require.NoError(t, result.GetError())
			require.NotNil(t, result.Definition.Spec.Template)
			assert.Equal(t, testCase.testTemplate, *result.Definition.Spec.Template)
		})
	}
}

func TestLokiStackIsReady(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		conditionType  string
		invalidBuilder bool
		expectedResult bool
	}{
		{
			name:           "ready condition",
			conditionType:  "Ready",
			expectedResult: true,
		},
		{
			name:          "non-ready condition",
			conditionType: "NotReady",
		},
		{
			name:           "invalid builder returns false",
			conditionType:  "Ready",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := buildLokiStackClientWithDummyObject()
			testBuilder := buildLokiStackTestBuilder(apiClient, testCase.invalidBuilder)
			testBuilder.Definition.Status.Conditions = []metav1.Condition{{
				Type:   testCase.conditionType,
				Status: metav1.ConditionTrue,
			}}

			assert.Equal(t, testCase.expectedResult, testBuilder.IsReady(2*time.Second))
		})
	}

	t.Run("missing resource returns false", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildValidLokiStackBuilder(clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: lokiv1TestSchemes,
		}))
		testBuilder.Definition.Status.Conditions = []metav1.Condition{{
			Type:   "Ready",
			Status: metav1.ConditionTrue,
		}}

		assert.False(t, testBuilder.IsReady(2*time.Second))
	})
}

func buildLokiStackTestBuilder(apiClient *clients.Settings, invalid bool) *LokiStackBuilder {
	if invalid {
		return buildInValidLokiStackBuilder(apiClient)
	}

	return buildValidLokiStackBuilder(apiClient)
}

func requireInvalidLokiStackBuilderError(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)
	assert.True(t, commonerrors.IsBuilderNameEmpty(err), "expected an empty-name builder error, got: %v", err)
}

func buildValidLokiStackBuilder(apiClient *clients.Settings) *LokiStackBuilder {
	return NewLokiStackBuilder(apiClient, defaultLokiStackName, defaultLokiStackNamespace)
}

func buildInValidLokiStackBuilder(apiClient *clients.Settings) *LokiStackBuilder {
	return NewLokiStackBuilder(apiClient, "", defaultLokiStackNamespace)
}

func buildLokiStackClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyLokiStack(),
		SchemeAttachers: lokiv1TestSchemes,
	})
}

func buildDummyLokiStack() []runtime.Object {
	return append([]runtime.Object{}, &lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultLokiStackName,
			Namespace: defaultLokiStackNamespace,
		},
	})
}
