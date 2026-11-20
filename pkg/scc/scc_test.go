package scc

import (
	"testing"

	securityV1 "github.com/openshift/api/security/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

var sccGVK = securityV1.SchemeGroupVersion.WithKind("SecurityContextConstraints")

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	newBuilder := func(apiClient *clients.Settings, name string) *Builder {
		return NewBuilder(apiClient, name, "test-user", "default")
	}

	testhelper.NewClusterScopedBuilderTestConfig(newBuilder, securityV1.Install, sccGVK).ExecuteTests(t)
}

func TestNewBuilderRequiredFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		runAsUser      string
		selinuxContext string
		expectedError  string
	}{
		{
			name:           "valid strategy fields",
			runAsUser:      "RunAsAny",
			selinuxContext: "MustRunAs",
		},
		{
			name:           "empty runAsUser",
			runAsUser:      "",
			selinuxContext: "MustRunAs",
			expectedError:  errEmptyRunAsUser,
		},
		{
			name:           "empty selinuxContext",
			runAsUser:      "RunAsAny",
			selinuxContext: "",
			expectedError:  errEmptySELinuxContext,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewBuilder(
				clients.GetTestClients(clients.TestClientParams{}),
				"test-scc", testCase.runAsUser, testCase.selinuxContext)
			require.NotNil(t, builder)

			if testCase.expectedError != "" {
				require.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, securityV1.RunAsUserStrategyType(testCase.runAsUser), builder.Definition.RunAsUser.Type)
			assert.Equal(t, securityV1.SELinuxContextStrategyType(testCase.selinuxContext), builder.Definition.SELinuxContext.Type)
		})
	}
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(Pull, securityV1.Install, sccGVK).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newSCCCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestWithPrivilegedContainer(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithPrivilegedContainer, []sccModifierTest[bool]{
		{
			name:  "allow privileged containers",
			input: true,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowPrivilegedContainer)
			},
		},
		{
			name:  "disallow privileged containers",
			input: false,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowPrivilegedContainer)
			},
		},
		invalidSCCModifierTest(false),
	})
}

func TestWithPrivilegedEscalation(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithPrivilegedEscalation, []sccModifierTest[bool]{
		{
			name:  "allow privileged escalation",
			input: true,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				require.NotNil(t, builder.Definition.DefaultAllowPrivilegeEscalation)
				assert.Equal(t, input, *builder.Definition.DefaultAllowPrivilegeEscalation)
			},
		},
		{
			name:  "disallow privileged escalation",
			input: false,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				require.NotNil(t, builder.Definition.DefaultAllowPrivilegeEscalation)
				assert.Equal(t, input, *builder.Definition.DefaultAllowPrivilegeEscalation)
			},
		},
		invalidSCCModifierTest(false),
	})
}

func TestWithHostDirVolumePlugin(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithHostDirVolumePlugin, []sccModifierTest[bool]{
		{
			name:  "allow hostDir volume plugin",
			input: true,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostDirVolumePlugin)
			},
		},
		{
			name:  "disallow hostDir volume plugin",
			input: false,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostDirVolumePlugin)
			},
		},
		invalidSCCModifierTest(false),
	})
}

func TestWithHostIPC(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithHostIPC, []sccModifierTest[bool]{
		{
			name:  "allow host IPC",
			input: true,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostIPC)
			},
		},
		{
			name:  "disallow host IPC",
			input: false,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostIPC)
			},
		},
		invalidSCCModifierTest(false),
	})
}

func TestWithHostNetwork(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithHostNetwork, []sccModifierTest[bool]{
		{
			name:  "allow host network",
			input: true,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostNetwork)
			},
		},
		{
			name:  "disallow host network",
			input: false,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostNetwork)
			},
		},
		invalidSCCModifierTest(false),
	})
}

func TestWithHostPID(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithHostPID, []sccModifierTest[bool]{
		{
			name:  "allow host PID",
			input: true,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostPID)
			},
		},
		{
			name:  "disallow host PID",
			input: false,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostPID)
			},
		},
		invalidSCCModifierTest(false),
	})
}

func TestWithHostPorts(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithHostPorts, []sccModifierTest[bool]{
		{
			name:  "allow host ports",
			input: true,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostPorts)
			},
		},
		{
			name:  "disallow host ports",
			input: false,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowHostPorts)
			},
		},
		invalidSCCModifierTest(false),
	})
}

func TestWithReadOnlyRootFilesystem(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithReadOnlyRootFilesystem, []sccModifierTest[bool]{
		{
			name:  "use read-only root filesystem",
			input: true,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.ReadOnlyRootFilesystem)
			},
		},
		{
			name:  "allow writable root filesystem",
			input: false,
			assertValue: func(t *testing.T, builder *Builder, input bool) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.ReadOnlyRootFilesystem)
			},
		},
		invalidSCCModifierTest(false),
	})
}

func TestWithDropCapabilities(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithDropCapabilities, []sccModifierTest[[]corev1.Capability]{
		{
			name:  "valid capabilities",
			input: []corev1.Capability{"NET_RAW", "NET_ADMIN"},
			assertValue: func(t *testing.T, builder *Builder, input []corev1.Capability) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.RequiredDropCapabilities)
			},
		},
		{
			name:          "empty capabilities",
			input:         []corev1.Capability{},
			expectedError: "securityContextConstraints 'requiredDropCapabilities' cannot be empty list",
		},
		invalidSCCModifierTest([]corev1.Capability{"NET_RAW"}),
	})
}

func TestWithAllowCapabilities(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithAllowCapabilities, []sccModifierTest[[]corev1.Capability]{
		{
			name:  "valid capabilities",
			input: []corev1.Capability{"NET_RAW", "NET_ADMIN"},
			assertValue: func(t *testing.T, builder *Builder, input []corev1.Capability) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.AllowedCapabilities)
			},
		},
		{
			name:          "empty capabilities",
			input:         []corev1.Capability{},
			expectedError: "securityContextConstraints 'allowCapabilities' cannot be empty list",
		},
		invalidSCCModifierTest([]corev1.Capability{"NET_RAW"}),
	})
}

func TestWithDefaultAddCapabilities(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithDefaultAddCapabilities, []sccModifierTest[[]corev1.Capability]{
		{
			name:  "valid capabilities",
			input: []corev1.Capability{"NET_RAW", "NET_ADMIN"},
			assertValue: func(t *testing.T, builder *Builder, input []corev1.Capability) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.DefaultAddCapabilities)
			},
		},
		{
			name:          "empty capabilities",
			input:         []corev1.Capability{},
			expectedError: "securityContextConstraints 'defaultAddCapabilities' cannot be empty list",
		},
		invalidSCCModifierTest([]corev1.Capability{"NET_RAW"}),
	})
}

func TestWithPriority(t *testing.T) {
	t.Parallel()

	priority := int32(100)

	runSCCModifierTests(t, (*Builder).WithPriority, []sccModifierTest[*int32]{
		{
			name:  "valid priority",
			input: &priority,
			assertValue: func(t *testing.T, builder *Builder, input *int32) {
				t.Helper()
				require.NotNil(t, builder.Definition.Priority)
				assert.Equal(t, *input, *builder.Definition.Priority)
			},
		},
		invalidSCCModifierTest(&priority),
	})
}

func TestWithFSGroup(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithFSGroup, []sccModifierTest[string]{
		{
			name:  "valid fsGroup",
			input: "RunAsAny",
			assertValue: func(t *testing.T, builder *Builder, input string) {
				t.Helper()
				assert.Equal(t, securityV1.FSGroupStrategyType(input), builder.Definition.FSGroup.Type)
			},
		},
		{
			name:          "empty fsGroup",
			input:         "",
			expectedError: "securityContextConstraints 'fsGroup' cannot be empty string",
		},
		invalidSCCModifierTest("RunAsAny"),
	})
}

func TestWithFSGroupRange(t *testing.T) {
	t.Parallel()

	type rangeInput struct {
		min int64
		max int64
	}

	runSCCModifierTests(t, func(builder *Builder, input rangeInput) *Builder {
		return builder.WithFSGroupRange(input.min, input.max)
	}, []sccModifierTest[rangeInput]{
		{
			name:  "valid range",
			input: rangeInput{min: 1001, max: 1005},
			assertValue: func(t *testing.T, builder *Builder, input rangeInput) {
				t.Helper()
				assert.Equal(t, []securityV1.IDRange{{Min: input.min, Max: input.max}}, builder.Definition.FSGroup.Ranges)
			},
		},
		{
			name:          "minimum greater than maximum",
			input:         rangeInput{min: 1005, max: 1001},
			expectedError: "securityContextConstraints 'fsGroupMin' argument can not be greater than fsGroupMax",
		},
		invalidSCCModifierTest(rangeInput{min: 1001, max: 1005}),
	})
}

func TestWithGroups(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithGroups, []sccModifierTest[[]string]{
		{
			name:  "valid groups",
			input: []string{"1001", "1002"},
			assertValue: func(t *testing.T, builder *Builder, input []string) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.Groups)
			},
		},
		{
			name:          "empty groups",
			input:         []string{},
			expectedError: "securityContextConstraints 'fsGroupType' cannot be empty string",
		},
		invalidSCCModifierTest([]string{"1001"}),
	})
}

func TestWithSeccompProfiles(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithSeccompProfiles, []sccModifierTest[[]string]{
		{
			name:  "valid seccomp profiles",
			input: []string{"test1", "test2"},
			assertValue: func(t *testing.T, builder *Builder, input []string) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.SeccompProfiles)
			},
		},
		{
			name:          "empty seccomp profiles",
			input:         []string{},
			expectedError: "securityContextConstraints 'seccompProfiles' cannot be empty list",
		},
		invalidSCCModifierTest([]string{"test1"}),
	})
}

func TestWithSupplementalGroups(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithSupplementalGroups, []sccModifierTest[string]{
		{
			name:  "valid supplemental groups type",
			input: "RunAsAny",
			assertValue: func(t *testing.T, builder *Builder, input string) {
				t.Helper()
				assert.Equal(t, securityV1.SupplementalGroupsStrategyType(input), builder.Definition.SupplementalGroups.Type)
			},
		},
		{
			name:          "empty supplemental groups type",
			input:         "",
			expectedError: "securityContextConstraints 'SupplementalGroups' cannot be empty string",
		},
		invalidSCCModifierTest("RunAsAny"),
	})
}

func TestWithUsers(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithUsers, []sccModifierTest[[]string]{
		{
			name:  "valid users",
			input: []string{"1001", "1002"},
			assertValue: func(t *testing.T, builder *Builder, input []string) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.Users)
			},
		},
		{
			name:          "empty users",
			input:         []string{},
			expectedError: "securityContextConstraints 'users' cannot be empty list",
		},
		invalidSCCModifierTest([]string{"1001"}),
	})
}

func TestWithVolumes(t *testing.T) {
	t.Parallel()

	runSCCModifierTests(t, (*Builder).WithVolumes, []sccModifierTest[[]securityV1.FSType]{
		{
			name:  "valid volumes",
			input: []securityV1.FSType{"test1", "test2"},
			assertValue: func(t *testing.T, builder *Builder, input []securityV1.FSType) {
				t.Helper()
				assert.Equal(t, input, builder.Definition.Volumes)
			},
		},
		{
			name:          "empty volumes",
			input:         []securityV1.FSType{},
			expectedError: "securityContextConstraints 'volumes' cannot be empty list",
		},
		invalidSCCModifierTest([]securityV1.FSType{"test1"}),
	})
}

type sccModifierTest[T any] struct {
	name          string
	input         T
	expectedError string
	invalid       bool
	assertValue   func(*testing.T, *Builder, T)
}

func invalidSCCModifierTest[T any](input T) sccModifierTest[T] {
	return sccModifierTest[T]{
		name:          "invalid builder short circuits",
		input:         input,
		expectedError: errEmptySELinuxContext,
		invalid:       true,
	}
}

func runSCCModifierTests[T any](
	t *testing.T,
	modifier func(*Builder, T) *Builder,
	testCases []sccModifierTest[T],
) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: []clients.SchemeAttacher{securityV1.Install},
			})

			var builder *Builder
			if testCase.invalid {
				builder = buildInvalidSCCTestBuilder(apiClient)
			} else {
				builder = buildValidSCCTestBuilder(apiClient)
			}

			result := modifier(builder, testCase.input)
			require.Same(t, builder, result)

			if testCase.expectedError != "" {
				require.EqualError(t, result.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, result.GetError())
			testCase.assertValue(t, result, testCase.input)
		})
	}
}

func newSCCCommonTestConfig() testhelper.CommonTestConfig[
	securityV1.SecurityContextConstraints,
	Builder,
	*securityV1.SecurityContextConstraints,
	*Builder,
] {
	return testhelper.NewCommonTestConfig[securityV1.SecurityContextConstraints, Builder](
		securityV1.Install, sccGVK, testhelper.ResourceScopeClusterScoped)
}

func buildValidSCCTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "test-scc", "RunAsAny", "MustRunAs")
}

func buildInvalidSCCTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "test-scc", "RunAsAny", "")
}
