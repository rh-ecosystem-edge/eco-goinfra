package kmm

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	mcmV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm-hub/v1beta1"
	kmmV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultManagedClusterModuleName      = "managedclustermoduletest"
	defaultManagedClusterModuleNamespace = "managedclustermoduletestns"
)

func TestNewManagedClusterModuleBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewManagedClusterModuleBuilder,
		mcmV1Beta1.AddToScheme,
		managedClusterModuleGVK,
	).ExecuteTests(t)
}

func TestPullManagedClusterModule(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullManagedClusterModule,
		mcmV1Beta1.AddToScheme,
		managedClusterModuleGVK,
	).ExecuteTests(t)
}

func TestManagedClusterModuleBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[mcmV1Beta1.ManagedClusterModule, ManagedClusterModuleBuilder](
		mcmV1Beta1.AddToScheme,
		managedClusterModuleGVK,
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

func TestManagedClusterModuleWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[mcmV1Beta1.ManagedClusterModule, ManagedClusterModuleBuilder](
			mcmV1Beta1.AddToScheme,
			managedClusterModuleGVK,
			testhelper.ResourceScopeNamespaced,
		),
	).ExecuteTests(t)
}

func TestManagedClusterModuleWithModuleSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		moduleSpec kmmV1Beta1.ModuleSpec
		valid      bool
	}{
		{
			name: "set module spec",
			moduleSpec: kmmV1Beta1.ModuleSpec{
				Selector: map[string]string{"test": "test"},
			},
			valid: true,
		},
		{
			name: "invalid builder",
			moduleSpec: kmmV1Beta1.ModuleSpec{
				Selector: map[string]string{"test": "test"},
			},
			valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildManagedClusterModuleTestBuilder(testCase.valid)
			testBuilder.WithModuleSpec(testCase.moduleSpec)

			if testCase.valid {
				require.NoError(t, testBuilder.GetError())
				assert.Equal(t, testCase.moduleSpec, testBuilder.Definition.Spec.ModuleSpec)
			} else {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(testBuilder.GetError()))
			}
		})
	}
}

func TestManagedClusterModuleWithSpokeNamespace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		spokeNamespace string
		expectedError  string
		valid          bool
	}{
		{
			name:           "set spoke namespace",
			spokeNamespace: "test",
			valid:          true,
		},
		{
			name:           "empty spoke namespace",
			spokeNamespace: "",
			expectedError:  "invalid 'spokeNamespace' argument cannot be nil",
			valid:          true,
		},
		{
			name:           "invalid builder",
			spokeNamespace: "test",
			valid:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildManagedClusterModuleTestBuilder(testCase.valid)
			testBuilder.WithSpokeNamespace(testCase.spokeNamespace)

			switch {
			case !testCase.valid:
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(testBuilder.GetError()))
			case testCase.expectedError != "":
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
			default:
				require.NoError(t, testBuilder.GetError())
				assert.Equal(t, testCase.spokeNamespace, testBuilder.Definition.Spec.SpokeNamespace)
			}
		})
	}
}

func TestManagedClusterModuleWithSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		selector      map[string]string
		expectedError string
		valid         bool
	}{
		{
			name:     "set selector",
			selector: map[string]string{"test": "test"},
			valid:    true,
		},
		{
			name:          "empty selector",
			selector:      nil,
			expectedError: "invalid 'selector' argument cannot be empty",
			valid:         true,
		},
		{
			name:     "invalid builder",
			selector: map[string]string{"test": "test"},
			valid:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildManagedClusterModuleTestBuilder(testCase.valid)
			testBuilder.WithSelector(testCase.selector)

			switch {
			case !testCase.valid:
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(testBuilder.GetError()))
			case testCase.expectedError != "":
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
			default:
				require.NoError(t, testBuilder.GetError())
				assert.Equal(t, testCase.selector, testBuilder.Definition.Spec.Selector)
			}
		})
	}
}

func TestManagedClusterModuleWithImageRebuildTriggerGeneration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		generation int
		valid      bool
	}{
		{
			name:       "set generation one",
			generation: 1,
			valid:      true,
		},
		{
			name:       "set generation zero",
			generation: 0,
			valid:      true,
		},
		{
			name:       "set generation forty two",
			generation: 42,
			valid:      true,
		},
		{
			name:       "invalid builder",
			generation: 1,
			valid:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildManagedClusterModuleTestBuilder(testCase.valid)
			testBuilder.WithImageRebuildTriggerGeneration(testCase.generation)

			if testCase.valid {
				require.NoError(t, testBuilder.GetError())
				require.NotNil(t, testBuilder.Definition.Spec.ModuleSpec.ImageRebuildTriggerGeneration)
				assert.Equal(
					t,
					testCase.generation,
					*testBuilder.Definition.Spec.ModuleSpec.ImageRebuildTriggerGeneration,
				)
			} else {
				require.Error(t, testBuilder.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.ModuleSpec.ImageRebuildTriggerGeneration)
			}
		})
	}
}

// buildManagedClusterModuleTestBuilder returns a valid or invalid ManagedClusterModuleBuilder for testing.
func buildManagedClusterModuleTestBuilder(valid bool) *ManagedClusterModuleBuilder {
	testSettings := buildManagedClusterModuleTestClientWithDummyObject()
	if valid {
		return NewManagedClusterModuleBuilder(
			testSettings,
			defaultManagedClusterModuleName,
			defaultManagedClusterModuleNamespace,
		)
	}

	return NewManagedClusterModuleBuilder(testSettings, "", defaultManagedClusterModuleNamespace)
}

// buildManagedClusterModuleTestClientWithDummyObject returns a client containing a dummy ManagedClusterModule.
func buildManagedClusterModuleTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyManagedClusterModule(),
		SchemeAttachers: []clients.SchemeAttacher{mcmV1Beta1.AddToScheme},
	})
}

// buildDummyManagedClusterModule returns a dummy ManagedClusterModule for testing.
func buildDummyManagedClusterModule() []runtime.Object {
	return []runtime.Object{
		generateManagedClusterModule(defaultManagedClusterModuleName, defaultManagedClusterModuleNamespace),
	}
}

// generateManagedClusterModule returns a ManagedClusterModule with the given name and namespace.
func generateManagedClusterModule(name, nsname string) *mcmV1Beta1.ManagedClusterModule {
	return &mcmV1Beta1.ManagedClusterModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}
