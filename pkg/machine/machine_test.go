package machine

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultMachineSetName = "worker"
	machineSetLabel       = "machine.openshift.io/cluster-api-machineset"
)

func TestPullSet(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullSet, machinev1beta1.AddToScheme, machineSetGVK).ExecuteTests(t)
}

func TestSetBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[machinev1beta1.MachineSet, SetBuilder](
		machinev1beta1.AddToScheme,
		machineSetGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestNewSetBuilderFromCopy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		providerKind     string
		providerField    string
		instanceType     string
		expectedInstance string
	}{
		{
			name:             "aws provider",
			providerKind:     "AWSMachineProviderConfig",
			providerField:    "instanceType",
			instanceType:     "m5.xlarge",
			expectedInstance: "m5.xlarge",
		},
		{
			name:             "gcp provider",
			providerKind:     "GCPMachineProviderSpec",
			providerField:    "machineType",
			instanceType:     "n2_standard_4",
			expectedInstance: "n2_standard_4",
		},
		{
			name:             "azure provider",
			providerKind:     "AzureMachineProviderSpec",
			providerField:    "vmSize",
			instanceType:     "Standard_D4s_v3",
			expectedInstance: "Standard_D4s_v3",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			workerSet := buildTestMachineSet(t, defaultMachineSetName, testCase.providerKind)
			testClient := buildTestClientWithMachineSets(workerSet)

			builder := NewSetBuilderFromCopy(
				testClient,
				defaultMachineSetNamespace,
				testCase.instanceType,
				defaultMachineSetWorkerRole,
				3,
			)

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition)

			expectedName := fmt.Sprintf(
				"%s-%s", defaultMachineSetName, strings.ToLower(strings.NewReplacer(".", "-", "_", "-").Replace(
					testCase.instanceType)))
			assert.Equal(t, expectedName, builder.Definition.Name)
			assert.Empty(t, builder.Definition.UID)
			assert.Empty(t, builder.Definition.ResourceVersion)
			assert.Equal(t, expectedName, builder.Definition.Spec.Selector.MatchLabels[machineSetLabel])
			assert.Equal(t, expectedName, builder.Definition.Spec.Template.Labels[machineSetLabel])
			require.NotNil(t, builder.Definition.Spec.Replicas)
			assert.Equal(t, int32(3), *builder.Definition.Spec.Replicas)

			providerSpec := getProviderSpecMap(t, builder.Definition)
			assert.Equal(t, testCase.expectedInstance, providerSpec[testCase.providerField])
		})
	}
}

//nolint:funlen // Table-driven error coverage for the package-specific copy constructor.
func TestNewSetBuilderFromCopyErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		namespace       string
		instanceType    string
		workerLabel     string
		replicas        int32
		object          *machinev1beta1.MachineSet
		expectedMessage string
	}{
		{
			name:            "empty namespace",
			namespace:       "",
			instanceType:    "m5.xlarge",
			workerLabel:     defaultMachineSetWorkerRole,
			replicas:        1,
			object:          buildTestMachineSet(t, defaultMachineSetName, "AWSMachineProviderConfig"),
			expectedMessage: "failed to list MachineSets, 'namespace' parameter is empty",
		},
		{
			name:            "empty worker label",
			namespace:       defaultMachineSetNamespace,
			instanceType:    "m5.xlarge",
			workerLabel:     "",
			replicas:        1,
			object:          buildTestMachineSet(t, defaultMachineSetName, "AWSMachineProviderConfig"),
			expectedMessage: "failed to list MachineSets, 'workerLabel' parameter is empty",
		},
		{
			name:            "no worker machine sets",
			namespace:       defaultMachineSetNamespace,
			instanceType:    "m5.xlarge",
			workerLabel:     defaultMachineSetWorkerRole,
			replicas:        1,
			expectedMessage: "no worker MachineSets were found",
		},
		{
			name:            "zero replicas",
			namespace:       defaultMachineSetNamespace,
			instanceType:    "m5.xlarge",
			workerLabel:     defaultMachineSetWorkerRole,
			replicas:        0,
			object:          buildTestMachineSet(t, defaultMachineSetName, "AWSMachineProviderConfig"),
			expectedMessage: "MachineSet 'replicas' cannot be zero",
		},
		{
			name:            "empty instance type",
			namespace:       defaultMachineSetNamespace,
			instanceType:    "",
			workerLabel:     defaultMachineSetWorkerRole,
			replicas:        1,
			object:          buildTestMachineSet(t, defaultMachineSetName, "AWSMachineProviderConfig"),
			expectedMessage: "instanceType parameter cannot be empty",
		},
		{
			name:            "unsupported provider",
			namespace:       defaultMachineSetNamespace,
			instanceType:    "m5.xlarge",
			workerLabel:     defaultMachineSetWorkerRole,
			replicas:        1,
			object:          buildTestMachineSet(t, defaultMachineSetName, "UnsupportedProvider"),
			expectedMessage: "unsupported cloud platform",
		},
		{
			name:            "invalid provider spec",
			namespace:       defaultMachineSetNamespace,
			instanceType:    "m5.xlarge",
			workerLabel:     defaultMachineSetWorkerRole,
			replicas:        1,
			object:          buildTestMachineSetWithRawProvider([]byte("123")),
			expectedMessage: "error unmarshalling the byte array into a providerSpec map",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.object != nil {
				objects = append(objects, testCase.object)
			}

			builder := NewSetBuilderFromCopy(
				buildTestClientWithMachineSets(objects...),
				testCase.namespace,
				testCase.instanceType,
				testCase.workerLabel,
				testCase.replicas,
			)

			require.Error(t, builder.GetError())
			assert.Contains(t, builder.GetError().Error(), testCase.expectedMessage)
		})
	}
}

func TestSetBuilderGetPublicCloudKindErrors(t *testing.T) {
	t.Parallel()

	builder := newSetBuilderWithDefinition(
		buildTestClientWithMachineSets(),
		buildTestMachineSetWithRawProvider([]byte("123")),
	)

	err := builder.getPublicCloudKind()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshalling the byte array into a providerSpec map")
	require.Error(t, builder.GetError())
	assert.Contains(t, builder.GetError().Error(), "error determining public cloud kind")
}

func TestSetBuilderChangeCloudProviderInstanceType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		providerKind  string
		providerField string
		instanceType  string
	}{
		{
			name:          "aws provider",
			providerKind:  "AWSMachineProviderConfig",
			providerField: "instanceType",
			instanceType:  "m6i.large",
		},
		{
			name:          "gcp provider",
			providerKind:  "GCPMachineProviderSpec",
			providerField: "machineType",
			instanceType:  "n2-standard-8",
		},
		{
			name:          "azure provider",
			providerKind:  "AzureMachineProviderSpec",
			providerField: "vmSize",
			instanceType:  "Standard_D8s_v3",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := newSetBuilderWithDefinition(
				buildTestClientWithMachineSets(),
				buildTestMachineSet(t, defaultMachineSetName, testCase.providerKind),
			)

			require.NoError(t, builder.getPublicCloudKind())
			require.NoError(t, builder.ChangeCloudProviderInstanceType(testCase.instanceType))
			assert.Equal(t, testCase.instanceType, getProviderSpecMap(t, builder.Definition)[testCase.providerField])
		})
	}
}

func TestSetBuilderCloudProviderMutationErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       *SetBuilder
		mutate        func(*SetBuilder) error
		assertError   func(error) bool
		expectedError string
	}{
		{
			name:    "unsupported public cloud",
			builder: buildSetBuilderForProvider(t, "AWSMachineProviderConfig"),
			mutate: func(builder *SetBuilder) error {
				builder.publicCloud = "unsupported"

				return builder.ChangeCloudProviderInstanceType("value")
			},
			expectedError: "could not find supported public cloud",
		},
		{
			name:    "aws empty instance type",
			builder: buildSetBuilderForProvider(t, "AWSMachineProviderConfig"),
			mutate: func(builder *SetBuilder) error {
				return builder.AWSChangeProviderInstanceType("")
			},
			expectedError: "instanceType parameter cannot be empty",
		},
		{
			name:    "gcp empty machine type",
			builder: buildSetBuilderForProvider(t, "GCPMachineProviderSpec"),
			mutate: func(builder *SetBuilder) error {
				return builder.GCPChangeProviderMachineType("")
			},
			expectedError: "machineType parameter cannot be empty",
		},
		{
			name:    "azure empty vm size",
			builder: buildSetBuilderForProvider(t, "AzureMachineProviderSpec"),
			mutate: func(builder *SetBuilder) error {
				return builder.AzureChangeProviderVMSize("")
			},
			expectedError: "vmSize parameter cannot be empty",
		},
		{
			name:    "nil builder",
			builder: nil,
			mutate: func(builder *SetBuilder) error {
				return builder.ChangeCloudProviderInstanceType("value")
			},
			assertError: commonerrors.IsBuilderNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.mutate(testCase.builder)

			if testCase.assertError != nil {
				assert.True(t, testCase.assertError(err))

				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.expectedError)
		})
	}
}

func buildTestClientWithMachineSets(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{machinev1beta1.AddToScheme},
	})
}

func buildSetBuilderForProvider(t *testing.T, providerKind string) *SetBuilder {
	t.Helper()

	builder := newSetBuilderWithDefinition(
		buildTestClientWithMachineSets(),
		buildTestMachineSet(t, defaultMachineSetName, providerKind),
	)
	require.NoError(t, builder.getPublicCloudKind())

	return builder
}

func buildTestMachineSet(t *testing.T, name, providerKind string) *machinev1beta1.MachineSet {
	t.Helper()

	return &machinev1beta1.MachineSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: machinev1beta1.GroupVersion.String(),
			Kind:       "MachineSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       defaultMachineSetNamespace,
			UID:             "test-uid",
			ResourceVersion: "1",
		},
		Spec: machinev1beta1.MachineSetSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{machineSetLabel: name},
			},
			Template: machinev1beta1.MachineTemplateSpec{
				ObjectMeta: machinev1beta1.ObjectMeta{
					Labels: map[string]string{
						defaultMachineSetWorkerRole: "worker",
						machineSetLabel:             name,
					},
				},
				Spec: machinev1beta1.MachineSpec{
					ProviderSpec: machinev1beta1.ProviderSpec{
						Value: buildProviderSpec(t, providerKind, nil),
					},
				},
			},
		},
	}
}

func buildTestMachineSetWithRawProvider(rawProvider []byte) *machinev1beta1.MachineSet {
	machineSet := &machinev1beta1.MachineSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultMachineSetName,
			Namespace: defaultMachineSetNamespace,
		},
		Spec: machinev1beta1.MachineSetSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{machineSetLabel: defaultMachineSetName},
			},
			Template: machinev1beta1.MachineTemplateSpec{
				ObjectMeta: machinev1beta1.ObjectMeta{
					Labels: map[string]string{
						defaultMachineSetWorkerRole: "worker",
						machineSetLabel:             defaultMachineSetName,
					},
				},
				Spec: machinev1beta1.MachineSpec{
					ProviderSpec: machinev1beta1.ProviderSpec{
						Value: &runtime.RawExtension{Raw: rawProvider},
					},
				},
			},
		},
	}

	return machineSet
}

func buildProviderSpec(t *testing.T, providerKind string, extraFields map[string]any) *runtime.RawExtension {
	t.Helper()

	providerSpec := map[string]any{
		"apiVersion": machinev1beta1.GroupVersion.String(),
		"kind":       providerKind,
	}

	for field, value := range extraFields {
		providerSpec[field] = value
	}

	rawProvider, err := json.Marshal(providerSpec)
	require.NoError(t, err)

	return &runtime.RawExtension{Raw: rawProvider}
}

func getProviderSpecMap(t *testing.T, machineSet *machinev1beta1.MachineSet) map[string]any {
	t.Helper()

	require.NotNil(t, machineSet.Spec.Template.Spec.ProviderSpec.Value)

	providerSpec := map[string]any{}
	require.NoError(t, json.Unmarshal(machineSet.Spec.Template.Spec.ProviderSpec.Value.Raw, &providerSpec))

	return providerSpec
}
