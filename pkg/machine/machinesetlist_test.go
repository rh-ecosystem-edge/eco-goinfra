package machine

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultMachineSetNamespace  = "openshift-machine-api"
	defaultMachineSetWorkerRole = "machine.openshift.io/cluster-api-machine-role"
)

//nolint:funlen // Table-driven list coverage for filtering, options, and validation.
func TestListWorkerMachineSets(t *testing.T) {
	t.Parallel()

	workerMachineSet := buildTestMachineSet(t, "worker-a", "AWSMachineProviderConfig")
	masterMachineSet := buildTestMachineSet(t, "master-a", "AWSMachineProviderConfig")
	masterMachineSet.Spec.Template.Labels[defaultMachineSetWorkerRole] = "master"

	poolAMachineSet := buildTestMachineSet(t, "worker-pool-a", "AWSMachineProviderConfig")
	poolAMachineSet.Spec.Template.Labels["pool"] = "a"
	poolAMachineSet.Labels = map[string]string{"pool": "a"}
	poolBMachineSet := buildTestMachineSet(t, "worker-pool-b", "AWSMachineProviderConfig")
	poolBMachineSet.Spec.Template.Labels["pool"] = "b"
	poolBMachineSet.Labels = map[string]string{"pool": "b"}

	testCases := []struct {
		name          string
		namespace     string
		workerLabel   string
		listOptions   []metav1.ListOptions
		objects       []runtime.Object
		nilClient     bool
		expectedCount int
		expectedError string
		assertError   func(error) bool
	}{
		{
			name:          "filters non-worker machine sets",
			namespace:     defaultMachineSetNamespace,
			workerLabel:   defaultMachineSetWorkerRole,
			objects:       []runtime.Object{workerMachineSet, masterMachineSet},
			expectedCount: 1,
		},
		{
			name:          "returns empty list",
			namespace:     defaultMachineSetNamespace,
			workerLabel:   defaultMachineSetWorkerRole,
			expectedCount: 0,
		},
		{
			name:          "applies label selector",
			namespace:     defaultMachineSetNamespace,
			workerLabel:   defaultMachineSetWorkerRole,
			listOptions:   []metav1.ListOptions{{LabelSelector: "pool=a"}},
			objects:       []runtime.Object{poolAMachineSet, poolBMachineSet},
			expectedCount: 1,
		},
		{
			name:          "empty namespace",
			namespace:     "",
			workerLabel:   defaultMachineSetWorkerRole,
			expectedError: "failed to list MachineSets, 'namespace' parameter is empty",
		},
		{
			name:          "empty workerLabel",
			namespace:     defaultMachineSetNamespace,
			workerLabel:   "",
			expectedError: "failed to list MachineSets, 'workerLabel' parameter is empty",
		},
		{
			name:          "too many list options",
			namespace:     defaultMachineSetNamespace,
			workerLabel:   defaultMachineSetWorkerRole,
			listOptions:   []metav1.ListOptions{{}, {}},
			expectedError: "error: more than one ListOptions was passed",
		},
		{
			name:          "invalid label selector",
			namespace:     defaultMachineSetNamespace,
			workerLabel:   defaultMachineSetWorkerRole,
			listOptions:   []metav1.ListOptions{{LabelSelector: "pool in ("}},
			expectedError: "failed to parse label selector",
		},
		{
			name:        "nil client",
			namespace:   defaultMachineSetNamespace,
			workerLabel: defaultMachineSetWorkerRole,
			nilClient:   true,
			assertError: commonerrors.IsAPIClientNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testClient *clients.Settings
			if !testCase.nilClient {
				testClient = buildTestClientWithMachineSets(testCase.objects...)
			}

			builders, err := ListWorkerMachineSets(testClient, testCase.namespace, testCase.workerLabel,
				testCase.listOptions...)

			if testCase.assertError != nil {
				assert.True(t, testCase.assertError(err))

				return
			}

			if testCase.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			require.NoError(t, err)
			require.Len(t, builders, testCase.expectedCount)

			for _, builder := range builders {
				require.NotNil(t, builder.Definition)
				require.NotNil(t, builder.Object)
				require.NotNil(t, builder.GetClient())
				assert.Equal(t, machineSetGVK, builder.GetGVK())
			}
		})
	}
}
