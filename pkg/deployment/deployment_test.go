package deployment

import (
	"context"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	multus "gopkg.in/k8snetworkplumbingwg/multus-cni.v4/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var deploymentGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[appsv1.Deployment, Builder](
		func(apiClient *clients.Settings, name, nsname string) *Builder {
			return NewBuilder(apiClient, name, nsname,
				map[string]string{testDeploymentLabelKey: "test"}, corev1.Container{Name: "test"})
		},
		appsv1.AddToScheme,
		deploymentGVK,
	).ExecuteTests(t)

	t.Run("sets deployment fields", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		require.NoError(t, builder.GetError())
		assert.Equal(t, map[string]string{testDeploymentLabelKey: "test"}, builder.Definition.Spec.Selector.MatchLabels)
		assert.Equal(t, map[string]string{testDeploymentLabelKey: "test"}, builder.Definition.Spec.Template.Labels)
		require.Len(t, builder.Definition.Spec.Template.Spec.Containers, 1)
		assert.Equal(t, "test-container", builder.Definition.Spec.Template.Spec.Containers[0].Name)
	})

	t.Run("empty labels returns error", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder(newDeploymentTestClient(), "test-name", "test-namespace", nil, corev1.Container{Name: "test"})
		require.EqualError(t, builder.GetError(), "deployment 'labels' cannot be empty")
	})
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[appsv1.Deployment, Builder](
		Pull, appsv1.AddToScheme, deploymentGVK).ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newDeploymentCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newDeploymentCommonTestConfig()).ExecuteTests(t)
}

func TestWithNodeSelector(t *testing.T) {
	t.Parallel()

	selector := map[string]string{"node": "worker"}

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithNodeSelector(selector)
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		assert.Equal(t, selector, builder.Definition.Spec.Template.Spec.NodeSelector)
	})
}

func TestWithReplicas(t *testing.T) {
	t.Parallel()

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithReplicas(3)
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		require.NotNil(t, builder.Definition.Spec.Replicas)
		assert.Equal(t, int32(3), *builder.Definition.Spec.Replicas)
	})
}

func TestWithAdditionalContainerSpecs(t *testing.T) {
	t.Parallel()

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithAdditionalContainerSpecs([]corev1.Container{{Name: "additional"}})
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		require.Len(t, builder.Definition.Spec.Template.Spec.Containers, 2)
		assert.Equal(t, "additional", builder.Definition.Spec.Template.Spec.Containers[1].Name)
	})

	t.Run("empty specs returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		builder.WithAdditionalContainerSpecs(nil)
		require.EqualError(t, builder.GetError(), "cannot accept empty list as container specs")
		assert.Len(t, builder.Definition.Spec.Template.Spec.Containers, 1)
	})
}

func TestWithSecondaryNetwork(t *testing.T) {
	t.Parallel()

	networks := []*multus.NetworkSelectionElement{{
		Name:      "test-secondary-network",
		Namespace: "test-secondary-network-namespace",
	}}

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithSecondaryNetwork(networks)
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		assert.Equal(t,
			"[{\"name\":\"test-secondary-network\",\"namespace\":\"test-secondary-network-namespace\",\"cni-args\":null}]",
			builder.Definition.Spec.Template.Annotations["k8s.v1.cni.cncf.io/networks"])
	})

	t.Run("empty networks returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		builder.WithSecondaryNetwork(nil)
		require.EqualError(t, builder.GetError(), "can not apply empty networks list")
		assert.Empty(t, builder.Definition.Spec.Template.Annotations)
	})
}

func TestWithHugePages(t *testing.T) {
	t.Parallel()

	testDeploymentModifier(t, (*Builder).WithHugePages, func(t *testing.T, builder *Builder) {
		t.Helper()

		require.Len(t, builder.Definition.Spec.Template.Spec.Volumes, 1)
		assert.Equal(t, "hugepages", builder.Definition.Spec.Template.Spec.Volumes[0].Name)
		require.NotNil(t, builder.Definition.Spec.Template.Spec.Volumes[0].EmptyDir)
		assert.Equal(t, corev1.StorageMedium("HugePages"), builder.Definition.Spec.Template.Spec.Volumes[0].EmptyDir.Medium)
		require.Len(t, builder.Definition.Spec.Template.Spec.Containers[0].VolumeMounts, 1)
		assert.Equal(t, "/mnt/huge", builder.Definition.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
	})
}

func TestWithSecurityContext(t *testing.T) {
	t.Parallel()

	runAsNonRoot := true
	securityContext := &corev1.PodSecurityContext{RunAsNonRoot: &runAsNonRoot}

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithSecurityContext(securityContext)
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		assert.Equal(t, securityContext, builder.Definition.Spec.Template.Spec.SecurityContext)
	})

	t.Run("nil security context returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		builder.WithSecurityContext(nil)
		require.EqualError(t, builder.GetError(), "'securityContext' parameter is empty")
		assert.Nil(t, builder.Definition.Spec.Template.Spec.SecurityContext)
	})
}

func TestWithLabel(t *testing.T) {
	t.Parallel()

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		builder.Definition.Spec.Template.Labels = nil

		return builder.WithLabel("test-label-key", "test-label-value")
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		assert.Equal(t, "test-label-value", builder.Definition.Spec.Template.Labels["test-label-key"])
	})

	t.Run("empty label key returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		builder.WithLabel("", "value")
		require.EqualError(t, builder.GetError(), "can not apply empty labelKey")
	})
}

func TestWithServiceAccountName(t *testing.T) {
	t.Parallel()

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithServiceAccountName("test-service-account")
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		assert.Equal(t, "test-service-account", builder.Definition.Spec.Template.Spec.ServiceAccountName)
	})

	t.Run("empty service account returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		builder.WithServiceAccountName("")
		require.EqualError(t, builder.GetError(), "can not apply empty serviceAccount")
	})
}

func TestWithVolume(t *testing.T) {
	t.Parallel()

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithVolume(corev1.Volume{Name: "test-volume"})
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		require.Len(t, builder.Definition.Spec.Template.Spec.Volumes, 1)
		assert.Equal(t, "test-volume", builder.Definition.Spec.Template.Spec.Volumes[0].Name)
	})

	t.Run("empty volume name returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		builder.WithVolume(corev1.Volume{})
		require.EqualError(t, builder.GetError(), "the volume's name cannot be empty")
		assert.Empty(t, builder.Definition.Spec.Template.Spec.Volumes)
	})
}

func TestWithSchedulerName(t *testing.T) {
	t.Parallel()

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithSchedulerName("test-scheduler")
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		assert.Equal(t, "test-scheduler", builder.Definition.Spec.Template.Spec.SchedulerName)
	})

	t.Run("empty scheduler name returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		builder.WithSchedulerName("")
		require.EqualError(t, builder.GetError(), "scheduler's name cannot be empty")
	})
}

func TestWithAffinity(t *testing.T) {
	t.Parallel()

	affinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "kubernetes.io/hostname",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"host1", "host2"},
					}},
				}},
			},
		},
	}

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithAffinity(affinity)
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		assert.Equal(t, affinity, builder.Definition.Spec.Template.Spec.Affinity)
	})

	t.Run("nil affinity returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		builder.WithAffinity(nil)
		require.EqualError(t, builder.GetError(), "affinity parameter is empty")
	})
}

func TestWithHostNetwork(t *testing.T) {
	t.Parallel()

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithHostNetwork(true)
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		assert.True(t, builder.Definition.Spec.Template.Spec.HostNetwork)
	})
}

func TestWithToleration(t *testing.T) {
	t.Parallel()

	toleration := corev1.Toleration{Key: "test", Operator: corev1.TolerationOpExists}

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithToleration(toleration)
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		require.Len(t, builder.Definition.Spec.Template.Spec.Tolerations, 1)
		assert.Equal(t, toleration, builder.Definition.Spec.Template.Spec.Tolerations[0])
	})

	t.Run("empty toleration returns error", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		builder.WithToleration(corev1.Toleration{})
		require.EqualError(t, builder.GetError(), "the toleration cannot be empty")
		assert.Empty(t, builder.Definition.Spec.Template.Spec.Tolerations)
	})
}

func TestWithTerminationGracePeriodSeconds(t *testing.T) {
	t.Parallel()

	testDeploymentModifier(t, func(builder *Builder) *Builder {
		return builder.WithTerminationGracePeriodSeconds(30)
	}, func(t *testing.T, builder *Builder) {
		t.Helper()

		require.NotNil(t, builder.Definition.Spec.Template.Spec.TerminationGracePeriodSeconds)
		assert.Equal(t, int64(30), *builder.Definition.Spec.Template.Spec.TerminationGracePeriodSeconds)
	})
}

func TestDeleteGraceful(t *testing.T) {
	t.Parallel()

	gracePeriod := int64(15)
	negativeGracePeriod := int64(-7)

	testCases := []struct {
		name        string
		gracePeriod *int64
		object      *appsv1.Deployment
		expectError string
	}{
		{name: "nil grace period", expectError: "gracePeriod cannot be nil"},
		{name: "negative grace period", gracePeriod: &negativeGracePeriod, expectError: "gracePeriod must be non-negative integer"},
		{name: "existing deployment", gracePeriod: &gracePeriod, object: buildDummyDeployment()},
		{name: "missing deployment", gracePeriod: &gracePeriod},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.object != nil {
				objects = append(objects, testCase.object)
			}

			builder := buildTestBuilderWithFakeObjects(objects)
			err := builder.DeleteGraceful(testCase.gracePeriod)

			if testCase.expectError != "" {
				require.EqualError(t, err, testCase.expectError)

				return
			}

			require.NoError(t, err)
			assert.Nil(t, builder.GetObject())
		})
	}
}

func TestCreateAndWaitUntilReady(t *testing.T) {
	t.Parallel()

	builder := buildTestBuilderWithFakeObjects([]runtime.Object{buildReadyDeployment()})

	result, err := builder.CreateAndWaitUntilReady(5 * time.Second)
	require.NoError(t, err)
	require.Same(t, builder, result)
}

func TestDeleteAndWait(t *testing.T) {
	t.Parallel()

	builder := buildTestBuilderWithFakeObjects([]runtime.Object{buildDummyDeployment()})

	require.NoError(t, builder.DeleteAndWait(5*time.Second))
	assert.Nil(t, builder.GetObject())
}

func TestWaitUntilCondition(t *testing.T) {
	t.Parallel()

	builder := buildTestBuilderWithFakeObjects([]runtime.Object{buildReadyDeployment()})

	require.NoError(t, builder.WaitUntilCondition(appsv1.DeploymentAvailable, 5*time.Second))
}

func TestWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		objects     []runtime.Object
		expectError error
	}{
		{name: "deployment absent"},
		{
			name:        "deployment present times out",
			objects:     []runtime.Object{buildDummyDeployment()},
			expectError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildTestBuilderWithFakeObjects(testCase.objects)
			assert.Equal(t, testCase.expectError, builder.WaitUntilDeleted(time.Second))
		})
	}
}

func TestGetGVR(t *testing.T) {
	t.Parallel()

	testGVR := GetGVR()
	assert.Equal(t, "apps", testGVR.Group)
	assert.Equal(t, "v1", testGVR.Version)
	assert.Equal(t, "deployments", testGVR.Resource)
}

func testDeploymentModifier(t *testing.T, modify func(*Builder) *Builder, verify func(*testing.T, *Builder)) {
	t.Helper()

	t.Run("applies to valid builder", func(t *testing.T) {
		t.Parallel()

		builder := buildValidTestBuilder()
		result := modify(builder)
		require.Same(t, builder, result)
		require.NoError(t, result.GetError())
		verify(t, result)
	})

	t.Run("invalid builder short circuits", func(t *testing.T) {
		t.Parallel()

		builder := buildInvalidTestBuilder()
		before := builder.Definition.DeepCopy()
		result := modify(builder)
		require.Same(t, builder, result)
		require.Error(t, result.GetError())
		assert.Equal(t, before, result.Definition)
	})
}

func newDeploymentCommonTestConfig() testhelper.CommonTestConfig[appsv1.Deployment, Builder, *appsv1.Deployment, *Builder] {
	return testhelper.NewCommonTestConfig[appsv1.Deployment, Builder](
		appsv1.AddToScheme, deploymentGVK, testhelper.ResourceScopeNamespaced)
}

func buildValidTestBuilder() *Builder {
	return NewBuilder(newDeploymentTestClient(), "test-name", "test-namespace",
		map[string]string{testDeploymentLabelKey: "test"}, corev1.Container{Name: "test-container", Image: "test-image"})
}

func buildInvalidTestBuilder() *Builder {
	return NewBuilder(newDeploymentTestClient(), "", "test-namespace",
		map[string]string{testDeploymentLabelKey: "test"}, corev1.Container{Name: "test-container", Image: "test-image"})
}

func buildTestBuilderWithFakeObjects(objects []runtime.Object) *Builder {
	return NewBuilder(newDeploymentTestClient(objects...), "test-name", "test-namespace",
		map[string]string{testDeploymentLabelKey: "test"}, corev1.Container{Name: "test-container", Image: "test-image"})
}

func newDeploymentTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{appsv1.AddToScheme},
	})
}

func buildDummyDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "test-name", Namespace: "test-namespace"}}
}

func buildReadyDeployment() *appsv1.Deployment {
	deployment := buildDummyDeployment()
	deployment.Status = appsv1.DeploymentStatus{
		Replicas:      1,
		ReadyReplicas: 1,
		Conditions: []appsv1.DeploymentCondition{{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionTrue,
		}},
	}

	return deployment
}
