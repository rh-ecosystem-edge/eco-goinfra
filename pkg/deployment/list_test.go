package deployment

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const testDeploymentLabelKey = "app"

var (
	defaultDeploymentName      = "test-deployment"
	defaultDeploymentNamespace = "test-namespace"
	otherDeploymentName        = "other-deployment"
	otherDeploymentNamespace   = "other-namespace"

	matchingLabelSelector = metav1.ListOptions{LabelSelector: testDeploymentLabelKey + "=match"}

	testDeployments = []runtime.Object{
		buildDummyDeploymentWithLabel(defaultDeploymentName, defaultDeploymentNamespace, "match"),
		buildDummyDeploymentWithLabel(otherDeploymentName, otherDeploymentNamespace, "other"),
	}
)

func TestList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*Builder, error) {
			return List(apiClient, nsname)
		},
		appsv1.AddToScheme,
		deploymentGVK,
	).ExecuteTests(t)
}

func TestListInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(ListInAllNamespaces, appsv1.AddToScheme, deploymentGVK).ExecuteTests(t)
}

func TestListOptionsAndScoping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		listAll       bool
		nsname        string
		listOptions   []metav1.ListOptions
		expectedNames []types.NamespacedName
	}{
		{
			name:          "only deployments in requested namespace",
			nsname:        otherDeploymentNamespace,
			expectedNames: []types.NamespacedName{{Name: otherDeploymentName, Namespace: otherDeploymentNamespace}},
		},
		{
			name:          "label selector excludes non-matching deployment",
			nsname:        defaultDeploymentNamespace,
			listOptions:   []metav1.ListOptions{matchingLabelSelector},
			expectedNames: []types.NamespacedName{{Name: defaultDeploymentName, Namespace: defaultDeploymentNamespace}},
		},
		{
			name:    "deployments from all namespaces",
			listAll: true,
			expectedNames: []types.NamespacedName{
				{Name: defaultDeploymentName, Namespace: defaultDeploymentNamespace},
				{Name: otherDeploymentName, Namespace: otherDeploymentNamespace},
			},
		},
		{
			name:          "label selector filters all namespaces",
			listAll:       true,
			listOptions:   []metav1.ListOptions{matchingLabelSelector},
			expectedNames: []types.NamespacedName{{Name: defaultDeploymentName, Namespace: defaultDeploymentNamespace}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testClient := newDeploymentTestClient(testDeployments...)

			var (
				result []*Builder
				err    error
			)
			if testCase.listAll {
				result, err = ListInAllNamespaces(testClient, testCase.listOptions...)
			} else {
				result, err = List(testClient, testCase.nsname, testCase.listOptions...)
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, testCase.expectedNames, deploymentKeys(result))
		})
	}
}

func TestListRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	testClient := newDeploymentTestClient()
	options := []metav1.ListOptions{{}, {}}

	_, err := List(testClient, defaultDeploymentNamespace, options...)
	require.EqualError(t, err, "error: more than one ListOptions was passed")

	_, err = ListInAllNamespaces(testClient, options...)
	require.EqualError(t, err, "error: more than one ListOptions was passed")
}

func buildDummyDeploymentWithLabel(name, namespace, labelValue string) *appsv1.Deployment {
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    map[string]string{testDeploymentLabelKey: labelValue},
	}}
}

func deploymentKeys(builders []*Builder) []types.NamespacedName {
	keys := make([]types.NamespacedName, 0, len(builders))

	for _, builder := range builders {
		keys = append(keys, types.NamespacedName{Name: builder.GetObject().Name, Namespace: builder.GetObject().Namespace})
	}

	return keys
}
