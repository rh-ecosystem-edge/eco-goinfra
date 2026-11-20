package oran

import (
	"context"
	"errors"
	"testing"
	"time"

	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultClusterTemplateName      = "test-cluster-template"
	defaultClusterTemplateNamespace = "test-namespace"
)

var (
	clusterTemplateGVK = provisioningv1alpha1.GroupVersion.WithKind("ClusterTemplate")

	defaultClusterTemplateCondition = metav1.Condition{
		Type:   string(provisioningv1alpha1.CTconditionTypes.Validated),
		Reason: string(provisioningv1alpha1.CTconditionReasons.Completed),
		Status: metav1.ConditionTrue,
	}
)

func TestPullClusterTemplate(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullClusterTemplate,
		provisioningv1alpha1.AddToScheme,
		clusterTemplateGVK,
	).ExecuteTests(t)
}

func TestClusterTemplateMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[provisioningv1alpha1.ClusterTemplate, ClusterTemplateBuilder](
		provisioningv1alpha1.AddToScheme,
		clusterTemplateGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}

func TestClusterTemplateWaitForCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		conditionMet bool
		exists       bool
		valid        bool
		assertError  func(error) bool
	}{
		{
			name:         "condition met",
			conditionMet: true,
			exists:       true,
			valid:        true,
			assertError:  func(err error) bool { return err == nil },
		},
		{
			name:         "condition not met",
			conditionMet: false,
			exists:       true,
			valid:        true,
			assertError:  func(err error) bool { return errors.Is(err, context.DeadlineExceeded) },
		},
		{
			name:         "cluster template does not exist",
			conditionMet: true,
			exists:       false,
			valid:        true,
			assertError: func(err error) bool {
				return err != nil && err.Error() == "cannot wait for non-existent ClusterTemplate"
			},
		},
		{
			name:        "invalid builder",
			exists:      true,
			valid:       false,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				clusterTemplate := buildDummyClusterTemplate(defaultClusterTemplateName, defaultClusterTemplateNamespace)
				if testCase.conditionMet {
					clusterTemplate.Status.Conditions = append(clusterTemplate.Status.Conditions, defaultClusterTemplateCondition)
				}

				runtimeObjects = append(runtimeObjects, clusterTemplate)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: provisioningTestSchemes,
			})

			name := defaultClusterTemplateName
			if !testCase.valid {
				name = ""
			}

			testBuilder := common.NewNamespacedBuilder[
				provisioningv1alpha1.ClusterTemplate,
				ClusterTemplateBuilder,
			](testSettings, provisioningv1alpha1.AddToScheme, name, defaultClusterTemplateNamespace)

			_, err := testBuilder.WaitForCondition(defaultClusterTemplateCondition, time.Second)
			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)
		})
	}
}

// buildDummyClusterTemplate returns a ClusterTemplate with the provided name and namespace.
func buildDummyClusterTemplate(name, nsname string) *provisioningv1alpha1.ClusterTemplate {
	return &provisioningv1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}
