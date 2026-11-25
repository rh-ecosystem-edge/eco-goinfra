package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	mlbtypes "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlboperator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	mlbTestSchemes = []clients.SchemeAttacher{
		mlbtypes.AddToScheme,
	}
	metalLbGVK                 = mlbtypes.GroupVersion.WithKind("MetalLB")
	defaultMetalLbName         = "metallbio"
	defaultMetalLbNsName       = "test-namespace"
	defaultMetalLbNodeSelector = map[string]string{"test": "test"}
)

func TestMetalLbNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *Builder {
				return NewBuilder(apiClient, name, nsname, defaultMetalLbNodeSelector)
			},
			mlbtypes.AddToScheme,
			metalLbGVK,
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		builderName   string
		nodeSelector  map[string]string
		assertBuilder func(t *testing.T, builder *Builder)
	}{
		{
			name:         "valid node selector",
			builderName:  defaultMetalLbName,
			nodeSelector: defaultMetalLbNodeSelector,
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				require.NoError(t, builder.GetError())
				assert.Equal(t, defaultMetalLbNodeSelector, builder.Definition.Spec.SpeakerNodeSelector)
			},
		},
		{
			name:         "empty node selector returns error",
			builderName:  defaultMetalLbName,
			nodeSelector: map[string]string{},
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.EqualError(t, builder.GetError(), "metallb 'nodeSelector' cannot be empty")
			},
		},
		{
			name:         "nil node selector returns error",
			builderName:  defaultMetalLbName,
			nodeSelector: nil,
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.EqualError(t, builder.GetError(), "metallb 'nodeSelector' cannot be empty")
			},
		},
		{
			name:         "common validation happens before node selector validation",
			builderName:  "",
			nodeSelector: nil,
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: mlbTestSchemes,
			})
			testBuilder := NewBuilder(
				testSettings, testCase.builderName, defaultMetalLbNsName, testCase.nodeSelector)

			require.NotNil(t, testBuilder)
			testCase.assertBuilder(t, testBuilder)
		})
	}
}

func TestMetalLbPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull, mlbtypes.AddToScheme, metalLbGVK,
	).ExecuteTests(t)
}

func TestMetalLbBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[mlbtypes.MetalLB, Builder](
		mlbtypes.AddToScheme,
		metalLbGVK,
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

func TestMetalLbWithOptions(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[mlbtypes.MetalLB, Builder](
		mlbtypes.AddToScheme,
		metalLbGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewWithOptionsTestConfig(commonTestConfig).ExecuteTests(t)
}

func TestMetalLbRemoveLabel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		key           string
		assertBuilder func(t *testing.T, builder *Builder)
	}{
		{
			name:  "removes existing key",
			valid: true,
			key:   "test",
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				require.NoError(t, builder.GetError())
				assert.Empty(t, builder.Definition.Spec.SpeakerNodeSelector)
			},
		},
		{
			name:  "empty key returns error",
			valid: true,
			key:   "",
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.EqualError(t, builder.GetError(), "error to remove empty key from metalLbIo")
			},
		},
		{
			name:  "invalid builder short circuits",
			valid: false,
			key:   "test",
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := buildMetalLbTestClientWithDummyObject()
			testBuilder := buildBuilderForModifierTest(testSettings, testCase.valid)
			testBuilder.RemoveLabel(testCase.key)

			testCase.assertBuilder(t, testBuilder)
		})
	}
}

func TestMetalLbWithSpeakerNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		nodeSelector  map[string]string
		assertBuilder func(t *testing.T, builder *Builder)
	}{
		{
			name:         "sets speaker node selector",
			valid:        true,
			nodeSelector: map[string]string{"node": "nodes"},
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				require.NoError(t, builder.GetError())
				assert.Equal(t, map[string]string{"node": "nodes"}, builder.Definition.Spec.SpeakerNodeSelector)
			},
		},
		{
			name:         "empty selector returns error",
			valid:        true,
			nodeSelector: map[string]string{},
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.EqualError(
					t, builder.GetError(), "can not accept empty label and redefine metallb NodeSelector")
			},
		},
		{
			name:         "invalid builder short circuits",
			valid:        false,
			nodeSelector: map[string]string{"node": "nodes"},
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := buildMetalLbTestClientWithDummyObject()
			testBuilder := buildBuilderForModifierTest(testSettings, testCase.valid)
			testBuilder.WithSpeakerNodeSelector(testCase.nodeSelector)

			testCase.assertBuilder(t, testBuilder)
		})
	}
}

func TestMetalLbWithFRRConfigAlwaysBlock(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		prefixes      []string
		assertBuilder func(t *testing.T, builder *Builder)
	}{
		{
			name:     "sets always block prefixes",
			valid:    true,
			prefixes: []string{"192.168.1.0/24"},
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				require.NoError(t, builder.GetError())
				require.NotNil(t, builder.Definition.Spec.FRRK8SConfig)
				assert.Equal(t, []string{"192.168.1.0/24"}, builder.Definition.Spec.FRRK8SConfig.AlwaysBlock)
			},
		},
		{
			name:     "empty prefixes returns error",
			valid:    true,
			prefixes: nil,
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.EqualError(
					t, builder.GetError(), "can not accept empty prefix list for the metallb alwaysBlock mode")
			},
		},
		{
			name:     "invalid CIDR returns error",
			valid:    true,
			prefixes: []string{"192.168.1."},
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.EqualError(t, builder.GetError(), "the prefix 192.168.1. is not a valid CIDR")
			},
		},
		{
			name:     "invalid builder short circuits",
			valid:    false,
			prefixes: []string{"192.168.1.0/24"},
			assertBuilder: func(t *testing.T, builder *Builder) {
				t.Helper()
				assert.True(t, commonerrors.IsBuilderNameEmpty(builder.GetError()))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := buildMetalLbTestClientWithDummyObject()
			testBuilder := buildBuilderForModifierTest(testSettings, testCase.valid)
			testBuilder.WithFRRConfigAlwaysBlock(testCase.prefixes)

			testCase.assertBuilder(t, testBuilder)
		})
	}
}

func TestGetMetalLbIoGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, GetMetalLbIoGVR(), schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "metallbs",
	})
}

func buildBuilderForModifierTest(apiClient *clients.Settings, valid bool) *Builder {
	if valid {
		return buildValidMetalLbBuilder(apiClient)
	}

	return buildInValidMetalLbBuilder(apiClient)
}

func buildValidMetalLbBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(
		apiClient, defaultMetalLbName, defaultMetalLbNsName, map[string]string{"test": "test"})
}

func buildInValidMetalLbBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "", defaultMetalLbNsName, defaultMetalLbNodeSelector)
}

func buildMetalLbTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyMetalLb(),
		SchemeAttachers: mlbTestSchemes,
	})
}

func buildDummyMetalLb() []runtime.Object {
	return append([]runtime.Object{}, &mlbtypes.MetalLB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultMetalLbName,
			Namespace: defaultMetalLbNsName,
		},
		Spec: mlbtypes.MetalLBSpec{
			SpeakerNodeSelector: defaultMetalLbNodeSelector,
		},
	})
}
