package storage

import (
	"testing"

	odfoperatorv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultSystemODFName      = "ocs-cluster-system"
	defaultSystemODFNamespace = "openshift-odf"
)

func TestNewSystemODFBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewSystemODFBuilder, odfoperatorv1alpha1.AddToScheme, systemODFGVK).ExecuteTests(t)
}

func TestPullSystemODF(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullSystemODF, odfoperatorv1alpha1.AddToScheme, systemODFGVK).ExecuteTests(t)
}

func TestSystemODFBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[odfoperatorv1alpha1.StorageSystem, SystemODFBuilder](
		odfoperatorv1alpha1.AddToScheme,
		systemODFGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestSystemODFWithSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		kind           odfoperatorv1alpha1.StorageKind
		specName       string
		specNamespace  string
		invalidBuilder bool
		expectedError  string
	}{
		{
			name:          "valid spec",
			kind:          odfoperatorv1alpha1.StorageKind("storagecluster.ocs.openshift.io/v1"),
			specName:      "ocs-cluster",
			specNamespace: defaultSystemODFNamespace,
		},
		{
			name:          "empty kind",
			specName:      "ocs-cluster",
			specNamespace: defaultSystemODFNamespace,
			expectedError: "SystemODF spec 'kind' cannot be empty",
		},
		{
			name:          "empty spec name",
			kind:          odfoperatorv1alpha1.StorageKind("cluster.ocs.openshift.io/v1"),
			specNamespace: defaultSystemODFNamespace,
			expectedError: "SystemODF spec 'name' cannot be empty",
		},
		{
			name:          "empty spec namespace",
			kind:          odfoperatorv1alpha1.StorageKind("storagecluster.ocs.openshift.io/v1"),
			specName:      "ocs-cluster",
			expectedError: "SystemODF spec 'nsname' cannot be empty",
		},
		{
			name:           "invalid builder",
			kind:           odfoperatorv1alpha1.StorageKind("storagecluster.ocs.openshift.io/v1"),
			specName:       "ocs-cluster",
			specNamespace:  defaultSystemODFNamespace,
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: []clients.SchemeAttacher{odfoperatorv1alpha1.AddToScheme},
			})

			var builder *SystemODFBuilder
			if testCase.invalidBuilder {
				builder = NewSystemODFBuilder(apiClient, defaultSystemODFName, "")
			} else {
				builder = NewSystemODFBuilder(apiClient, defaultSystemODFName, defaultSystemODFNamespace)
			}

			result := builder.WithSpec(testCase.kind, testCase.specName, testCase.specNamespace)
			require.NotNil(t, result)

			if testCase.invalidBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(result.GetError()))
				assert.Equal(t, odfoperatorv1alpha1.StorageSystemSpec{}, result.Definition.Spec)

				return
			}

			if testCase.expectedError != "" {
				require.Error(t, result.GetError())
				assert.Equal(t, testCase.expectedError, result.GetError().Error())
				assert.Equal(t, odfoperatorv1alpha1.StorageSystemSpec{}, result.Definition.Spec)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.kind, result.Definition.Spec.Kind)
			assert.Equal(t, testCase.specName, result.Definition.Spec.Name)
			assert.Equal(t, testCase.specNamespace, result.Definition.Spec.Namespace)
		})
	}
}
