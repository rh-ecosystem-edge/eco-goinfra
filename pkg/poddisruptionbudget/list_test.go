package poddisruptionbudget

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testPDBLabelKey       = "app"
	otherPDBName          = "other-poddisruptionbudget"
	otherPDBNamespace     = "other-namespace"
	matchingPDBLabelValue = "match"
)

var matchingPDBListOptions = metav1.ListOptions{LabelSelector: testPDBLabelKey + "=" + matchingPDBLabelValue}

func TestList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*Builder, error) {
			return List(apiClient, nsname)
		},
		policyv1.AddToScheme,
		pdbGVK,
	).ExecuteTests(t)
}

func TestListInAllNamespaces(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(ListInAllNamespaces, policyv1.AddToScheme, pdbGVK).ExecuteTests(t)
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
			name:          "only pod disruption budgets in requested namespace",
			nsname:        otherPDBNamespace,
			expectedNames: []types.NamespacedName{{Name: otherPDBName, Namespace: otherPDBNamespace}},
		},
		{
			name:          "label selector excludes non-matching pod disruption budget",
			nsname:        defaultPDBNamespace,
			listOptions:   []metav1.ListOptions{matchingPDBListOptions},
			expectedNames: []types.NamespacedName{{Name: defaultPDBName, Namespace: defaultPDBNamespace}},
		},
		{
			name:    "pod disruption budgets from all namespaces",
			listAll: true,
			expectedNames: []types.NamespacedName{
				{Name: defaultPDBName, Namespace: defaultPDBNamespace},
				{Name: otherPDBName, Namespace: otherPDBNamespace},
			},
		},
		{
			name:          "label selector filters all namespaces",
			listAll:       true,
			listOptions:   []metav1.ListOptions{matchingPDBListOptions},
			expectedNames: []types.NamespacedName{{Name: defaultPDBName, Namespace: defaultPDBNamespace}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testClient := getPDBTestClient(buildPDBTestObjects()...)

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
			assert.ElementsMatch(t, testCase.expectedNames, pdbKeys(result))
		})
	}
}

func TestListRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	testClient := getPDBTestClient()
	options := []metav1.ListOptions{{}, {}}

	_, err := List(testClient, defaultPDBNamespace, options...)
	require.EqualError(t, err, "error: more than one ListOptions was passed")

	_, err = ListInAllNamespaces(testClient, options...)
	require.EqualError(t, err, "error: more than one ListOptions was passed")
}

func TestListRejectsInvalidSelectors(t *testing.T) {
	t.Parallel()

	testClient := getPDBTestClient()
	invalidOptions := []metav1.ListOptions{{LabelSelector: "invalid in (selector"}}

	_, err := List(testClient, defaultPDBNamespace, invalidOptions...)
	require.Error(t, err)

	_, err = ListInAllNamespaces(testClient, invalidOptions...)
	require.Error(t, err)
}

// getPDBTestClient returns a test client with the PodDisruptionBudget scheme attached.
func getPDBTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{policyv1.AddToScheme},
	})
}

// buildPDBTestObjects returns PodDisruptionBudgets used to verify list scoping and selectors.
func buildPDBTestObjects() []runtime.Object {
	return []runtime.Object{
		buildDummyPDBWithLabel(defaultPDBName, defaultPDBNamespace, matchingPDBLabelValue),
		buildDummyPDBWithLabel(otherPDBName, otherPDBNamespace, "other"),
	}
}

// buildDummyPDBWithLabel returns a PodDisruptionBudget with the provided metadata and label value.
func buildDummyPDBWithLabel(name, namespace, labelValue string) *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{ObjectMeta: metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    map[string]string{testPDBLabelKey: labelValue},
	}}
}

// pdbKeys returns the namespaced names of the listed PodDisruptionBudget builders.
func pdbKeys(builders []*Builder) []types.NamespacedName {
	keys := make([]types.NamespacedName, 0, len(builders))

	for _, builder := range builders {
		keys = append(keys, types.NamespacedName{
			Name:      builder.GetObject().Name,
			Namespace: builder.GetObject().Namespace,
		})
	}

	return keys
}
