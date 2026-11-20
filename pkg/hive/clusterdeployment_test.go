package hive

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	hiveextV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/hiveextension/v1beta1"
	hivev1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/hive/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/hive/api/v1/agent"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/hive/api/v1/none"
	ibiv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/imagebasedinstall/api/hiveextensions/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		hivev1.AddToScheme,
	}

	defaultClusterDeploymentName      = "clusterdeploymentname"
	defaultClusterDeploymentNamespace = "clusterdeploymentnamespace"
	defaultClusterName                = "clustername"
	defaultBaseDomain                 = "clusterbasedomain"
	defaultClusterInstallRef          = "clusterInstallRef"
	defaultAgentSelector              = metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}}
)

//nolint:funlen // Table-driven constructor validation.
func TestNewABMClusterDeploymentBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		description       string
		name              string
		nsName            string
		clusterName       string
		baseDomain        string
		clusterInstallRef string
		agentSelector     metav1.LabelSelector
		assertError       func(error) bool
	}{
		{
			description:       "valid builder",
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			assertError:       func(err error) bool { return err == nil },
		},
		{
			description:       "empty name",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			assertError:       commonerrors.IsBuilderNameEmpty,
		},
		{
			description:       "empty namespace",
			name:              "testdeployment",
			nsName:            "",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			assertError:       commonerrors.IsBuilderNamespaceEmpty,
		},
		{
			description:       "empty cluster name",
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			assertError: func(err error) bool {
				return err != nil && err.Error() == errEmptyClusterName
			},
		},
		{
			description:       "empty base domain",
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			assertError: func(err error) bool {
				return err != nil && err.Error() == errEmptyBaseDomain
			},
		},
		{
			description:       "empty cluster install reference",
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			assertError: func(err error) bool {
				return err != nil && err.Error() == errEmptyClusterInstallRefName
			},
		},
		{
			description:       "empty agent selector is allowed",
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{},
			assertError:       func(err error) bool { return err == nil },
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			testSettings := buildTestClientWithClusterDeploymentScheme()
			testClusterDeployment := NewABMClusterDeploymentBuilder(
				testSettings, testCase.name, testCase.nsName, testCase.clusterName,
				testCase.baseDomain, testCase.clusterInstallRef, testCase.agentSelector)

			require.NotNil(t, testClusterDeployment)
			assert.Truef(t, testCase.assertError(testClusterDeployment.GetError()),
				"unexpected error, got: %v", testClusterDeployment.GetError())
			require.NotNil(t, testClusterDeployment.Definition)

			if testClusterDeployment.GetError() == nil {
				assert.Equal(t, testCase.name, testClusterDeployment.Definition.Name)
				assert.Equal(t, testCase.clusterName, testClusterDeployment.Definition.Spec.ClusterName)
				assert.Equal(t, testCase.baseDomain, testClusterDeployment.Definition.Spec.BaseDomain)
				require.NotNil(t, testClusterDeployment.Definition.Spec.ClusterInstallRef)
				assert.Equal(t, kindAgentClusterInstall, testClusterDeployment.Definition.Spec.ClusterInstallRef.Kind)
				require.NotNil(t, testClusterDeployment.Definition.Spec.Platform.AgentBareMetal)
				assert.Equal(t, testCase.agentSelector,
					testClusterDeployment.Definition.Spec.Platform.AgentBareMetal.AgentSelector)
			}
		})
	}
}

//nolint:funlen // Table-driven constructor validation.
func TestNewClusterDeploymentByInstallRefBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *ClusterDeploymentBuilder {
				return NewClusterDeploymentByInstallRefBuilder(
					apiClient,
					name,
					nsname,
					defaultClusterName,
					defaultBaseDomain,
					buildTestClusterInstallRef(defaultClusterInstallRef),
					buildTestAgentPlatform(defaultAgentSelector),
				)
			},
			hivev1.AddToScheme,
			clusterDeploymentGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name              string
		clusterName       string
		baseDomain        string
		clusterInstallRef hivev1.ClusterInstallLocalReference
		platform          hivev1.Platform
		expectedError     string
	}{
		{
			name:        "valid image cluster install",
			clusterName: "clustertest",
			baseDomain:  "domaintest",
			clusterInstallRef: hivev1.ClusterInstallLocalReference{
				Group:   ibiv1alpha1.Group,
				Version: ibiv1alpha1.Version,
				Kind:    "ImageClusterInstall",
				Name:    "test-imageclusterinstall",
			},
			platform: hivev1.Platform{
				None: &none.Platform{},
			},
		},
		{
			name:              "empty cluster name",
			baseDomain:        "domaintest",
			clusterInstallRef: buildTestClusterInstallRef("test-agentclusterinstall"),
			platform:          buildTestAgentPlatform(defaultAgentSelector),
			expectedError:     errEmptyClusterName,
		},
		{
			name:              "empty base domain",
			clusterName:       "clustertest",
			clusterInstallRef: buildTestClusterInstallRef("test-agentclusterinstall"),
			platform:          buildTestAgentPlatform(defaultAgentSelector),
			expectedError:     errEmptyBaseDomain,
		},
		{
			name:              "empty cluster install reference name",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: buildTestClusterInstallRef(""),
			platform:          buildTestAgentPlatform(defaultAgentSelector),
			expectedError:     errEmptyClusterInstallRefName,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testClusterDeployment := NewClusterDeploymentByInstallRefBuilder(
				buildTestClientWithClusterDeploymentScheme(),
				"testdeployment",
				"test-namespace",
				testCase.clusterName,
				testCase.baseDomain,
				testCase.clusterInstallRef,
				testCase.platform,
			)

			if testCase.expectedError != "" {
				require.Error(t, testClusterDeployment.GetError())
				assert.Equal(t, testCase.expectedError, testClusterDeployment.GetError().Error())

				return
			}

			require.NoError(t, testClusterDeployment.GetError())
			assert.Equal(t, testCase.clusterName, testClusterDeployment.Definition.Spec.ClusterName)
			assert.Equal(t, testCase.baseDomain, testClusterDeployment.Definition.Spec.BaseDomain)
			assert.Equal(t, testCase.clusterInstallRef, *testClusterDeployment.Definition.Spec.ClusterInstallRef)
			assert.Equal(t, testCase.platform, testClusterDeployment.Definition.Spec.Platform)
		})
	}
}

func TestPullClusterDeployment(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullClusterDeployment, hivev1.AddToScheme, clusterDeploymentGVK).ExecuteTests(t)
}

func TestClusterDeploymentBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[hivev1.ClusterDeployment, ClusterDeploymentBuilder](
		hivev1.AddToScheme,
		clusterDeploymentGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestClusterDeploymentWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[hivev1.ClusterDeployment, ClusterDeploymentBuilder](
			hivev1.AddToScheme,
			clusterDeploymentGVK,
			testhelper.ResourceScopeNamespaced,
		)).ExecuteTests(t)
}

func TestClusterDeploymentWithAdditionalAgentSelectorLabels(t *testing.T) {
	t.Parallel()

	validBuilder := func() *ClusterDeploymentBuilder {
		return buildValidClusterDeploymentBuilder(buildTestClientWithClusterDeploymentScheme())
	}
	invalidBuilder := func() *ClusterDeploymentBuilder {
		return buildInValidClusterDeploymentBuilder(buildTestClientWithClusterDeploymentScheme())
	}
	nonABMBuilder := func() *ClusterDeploymentBuilder {
		return buildValidNonABMClusterDeploymentBuilder(buildTestClientWithClusterDeploymentScheme())
	}

	testCases := []struct {
		name           string
		builder        func() *ClusterDeploymentBuilder
		agentSelector  map[string]string
		assertError    func(error) bool
		expectedLabels map[string]string
	}{
		{
			name:           "adds and merges labels",
			builder:        validBuilder,
			agentSelector:  map[string]string{"additional": "label"},
			assertError:    func(err error) bool { return err == nil },
			expectedLabels: map[string]string{"test": "test", "additional": "label"},
		},
		{
			name:          "empty selector",
			builder:       validBuilder,
			agentSelector: map[string]string{},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "agentSelector cannot be empty"
			},
		},
		{
			name:          "non agent bare metal platform",
			builder:       nonABMBuilder,
			agentSelector: map[string]string{"additional": "label"},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "clusterdeployment type must be AgentBareMetal to use agentSelector"
			},
		},
		{
			name:          "invalid builder",
			builder:       invalidBuilder,
			agentSelector: map[string]string{"additional": "label"},
			assertError:   commonerrors.IsBuilderNamespaceEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()

			result := builder.WithAdditionalAgentSelectorLabels(testCase.agentSelector)

			require.Same(t, builder, result)
			assert.Truef(t, testCase.assertError(result.GetError()), "unexpected error, got: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.expectedLabels,
					result.Definition.Spec.Platform.AgentBareMetal.AgentSelector.MatchLabels)
			}
		})
	}
}

func TestClusterDeploymentWithPullSecret(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		psName      string
		valid       bool
		assertError func(error) bool
	}{
		{
			name:        "sets pull secret",
			psName:      "test",
			valid:       true,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "empty pull secret name is allowed",
			psName:      "",
			valid:       true,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "invalid builder",
			psName:      "test",
			assertError: commonerrors.IsBuilderNamespaceEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var builder *ClusterDeploymentBuilder
			if testCase.valid {
				builder = buildValidClusterDeploymentBuilder(buildTestClientWithClusterDeploymentScheme())
			} else {
				builder = buildInValidClusterDeploymentBuilder(buildTestClientWithClusterDeploymentScheme())
			}

			result := builder.WithPullSecret(testCase.psName)

			require.Same(t, builder, result)
			assert.Truef(t, testCase.assertError(result.GetError()), "unexpected error, got: %v", result.GetError())

			if result.GetError() == nil {
				require.NotNil(t, result.Definition.Spec.PullSecretRef)
				assert.Equal(t, testCase.psName, result.Definition.Spec.PullSecretRef.Name)
			}
		})
	}
}

func buildValidClusterDeploymentBuilder(apiClient *clients.Settings) *ClusterDeploymentBuilder {
	return NewABMClusterDeploymentBuilder(
		apiClient,
		defaultClusterDeploymentName,
		defaultClusterDeploymentNamespace,
		defaultClusterName,
		defaultBaseDomain,
		defaultClusterInstallRef,
		*defaultAgentSelector.DeepCopy())
}

func buildInValidClusterDeploymentBuilder(apiClient *clients.Settings) *ClusterDeploymentBuilder {
	return NewABMClusterDeploymentBuilder(
		apiClient,
		defaultClusterDeploymentName,
		"",
		defaultClusterName,
		defaultBaseDomain,
		defaultClusterInstallRef,
		*defaultAgentSelector.DeepCopy())
}

func buildValidNonABMClusterDeploymentBuilder(apiClient *clients.Settings) *ClusterDeploymentBuilder {
	return NewClusterDeploymentByInstallRefBuilder(
		apiClient,
		defaultClusterDeploymentName,
		defaultClusterDeploymentNamespace,
		defaultClusterName,
		defaultBaseDomain,
		hivev1.ClusterInstallLocalReference{
			Group:   ibiv1alpha1.Group,
			Version: ibiv1alpha1.Version,
			Kind:    "ImageClusterInstall",
			Name:    defaultClusterInstallRef,
		},
		hivev1.Platform{None: &none.Platform{}},
	)
}

func buildTestClusterInstallRef(name string) hivev1.ClusterInstallLocalReference {
	return hivev1.ClusterInstallLocalReference{
		Group:   hiveextV1Beta1.Group,
		Version: hiveextV1Beta1.Version,
		Kind:    kindAgentClusterInstall,
		Name:    name,
	}
}

func buildTestAgentPlatform(agentSelector metav1.LabelSelector) hivev1.Platform {
	return hivev1.Platform{
		AgentBareMetal: &agent.BareMetalPlatform{
			AgentSelector: agentSelector,
		},
	}
}

func buildClusterDeploymentClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterDeployment(),
		SchemeAttachers: testSchemes,
	})
}

func buildTestClientWithClusterDeploymentScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

func buildDummyClusterDeployment() []runtime.Object {
	return append([]runtime.Object{}, &hivev1.ClusterDeployment{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "999",
			Name:            defaultClusterDeploymentName,
			Namespace:       defaultClusterDeploymentNamespace,
		},
		Spec: hivev1.ClusterDeploymentSpec{},
	})
}
