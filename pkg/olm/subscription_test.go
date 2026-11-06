package olm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	oplmV1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNewSubscriptionBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		func(apiClient *clients.Settings, name, nsname string) *SubscriptionBuilder {
			return NewSubscriptionBuilder(apiClient, name, nsname, "test-source", "test-namespace", "package-test")
		},
		oplmV1alpha1.AddToScheme,
		subscriptionGVK,
	).ExecuteTests(t)

	t.Run("initializes Subscription fields", func(t *testing.T) {
		t.Parallel()

		builder := NewSubscriptionBuilder(
			buildTestSubscriptionClientWithDummyObject(),
			"subscription",
			"test-namespace",
			"test-source",
			"catalog-namespace",
			"package-test",
		)

		require.NoError(t, builder.GetError())
		require.NotNil(t, builder.Definition.Spec)
		assert.Equal(t, "test-source", builder.Definition.Spec.CatalogSource)
		assert.Equal(t, "catalog-namespace", builder.Definition.Spec.CatalogSourceNamespace)
		assert.Equal(t, "package-test", builder.Definition.Spec.Package)
	})

	testCases := []struct {
		name                   string
		catalogSource          string
		catalogSourceNamespace string
		packageName            string
		expectedError          string
	}{
		{
			name:                   "empty catalog source",
			catalogSource:          "",
			catalogSourceNamespace: "catalog-namespace",
			packageName:            "package-test",
			expectedError:          "subscription 'catalogSource' cannot be empty",
		},
		{
			name:                   "empty catalog source namespace",
			catalogSource:          "test-source",
			catalogSourceNamespace: "",
			packageName:            "package-test",
			expectedError:          "subscription 'catalogSourceNamespace' cannot be empty",
		},
		{
			name:                   "empty package name",
			catalogSource:          "test-source",
			catalogSourceNamespace: "catalog-namespace",
			packageName:            "",
			expectedError:          "subscription 'packageName' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewSubscriptionBuilder(
				buildTestSubscriptionClientWithDummyObject(),
				"subscription",
				"test-namespace",
				testCase.catalogSource,
				testCase.catalogSourceNamespace,
				testCase.packageName,
			)

			assert.EqualError(t, builder.GetError(), testCase.expectedError)
		})
	}
}

func TestPullSubscription(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullSubscription,
		oplmV1alpha1.AddToScheme,
		subscriptionGVK,
	).ExecuteTests(t)
}

func TestSubscriptionBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[oplmV1alpha1.Subscription, SubscriptionBuilder](
		oplmV1alpha1.AddToScheme,
		subscriptionGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestSubscriptionWithChannel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		channel        string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:    "sets channel",
			channel: "test",
		},
		{
			name:          "rejects empty channel",
			expectedError: "can not redefine subscription with empty channel",
		},
		{
			name:           "invalid builder short-circuits",
			channel:        "test",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject())

			if testCase.invalidBuilder {
				builder = buildInValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject())
			}

			builder = builder.WithChannel(testCase.channel)

			if testCase.invalidBuilder {
				assert.Error(t, builder.GetError())

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.channel, builder.Definition.Spec.Channel)
		})
	}
}

func TestSubscriptionWithStartingCSV(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		startingCSV    string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:        "sets starting CSV",
			startingCSV: "test",
		},
		{
			name:          "rejects empty starting CSV",
			expectedError: "can not redefine subscription with empty startingCSV",
		},
		{
			name:           "invalid builder short-circuits",
			startingCSV:    "test",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject())

			if testCase.invalidBuilder {
				builder = buildInValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject())
			}

			builder = builder.WithStartingCSV(testCase.startingCSV)

			if testCase.invalidBuilder {
				assert.Error(t, builder.GetError())

				return
			}

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.startingCSV, builder.Definition.Spec.StartingCSV)
		})
	}
}

func TestSubscriptionWithInstallPlanApproval(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		installPlan       oplmV1alpha1.Approval
		invalidBuilder    bool
		expectedErrorText string
	}{
		{
			name:        "sets automatic approval",
			installPlan: oplmV1alpha1.ApprovalAutomatic,
		},
		{
			name:              "rejects invalid approval",
			expectedErrorText: "Subscription 'installPlanApproval' must be either \"Automatic\" or \"Manual\"",
		},
		{
			name:           "invalid builder short-circuits",
			installPlan:    oplmV1alpha1.ApprovalAutomatic,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject())

			if testCase.invalidBuilder {
				builder = buildInValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject())
			}

			builder = builder.WithInstallPlanApproval(testCase.installPlan)

			if testCase.invalidBuilder {
				assert.Error(t, builder.GetError())

				return
			}

			if testCase.expectedErrorText != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedErrorText)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.installPlan, builder.Definition.Spec.InstallPlanApproval)
		})
	}
}

func buildInValidSubscriptionBuilder(apiClient *clients.Settings) *SubscriptionBuilder {
	return NewSubscriptionBuilder(
		apiClient,
		"subscription",
		"",
		"test-source",
		"test-namespace",
		"package-test")
}

func buildValidSubscriptionBuilder(apiClient *clients.Settings) *SubscriptionBuilder {
	return NewSubscriptionBuilder(
		apiClient,
		"subscription",
		"test-namespace",
		"test-source",
		"test-namespace",
		"package-test")
}

func buildTestSubscriptionClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummySubscription(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummySubscription() []runtime.Object {
	return append([]runtime.Object{}, &oplmV1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subscription",
			Namespace: "test-namespace",
		},
		Spec: &oplmV1alpha1.SubscriptionSpec{
			Package: "test",
		},
	})
}
