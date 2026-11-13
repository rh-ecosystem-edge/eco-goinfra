package storage

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	defaultLabelSelector = "ceph.rook.io/DeviceSet=ocs-deviceset-0"
	defaultLabel         = map[string]string{"ceph.rook.io/DeviceSet": "ocs-deviceset-0"}
	defaultNamespace     = "persistentvolumeclaim-namespace"
)

func TestListPV(t *testing.T) {
	testCases := []struct {
		pv            []*PVBuilder
		listOptions   []metav1.ListOptions
		expectedError error
		client        bool
	}{
		{
			pv:            []*PVBuilder{buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume())},
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			pv:            []*PVBuilder{buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume())},
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}},
			expectedError: nil,
			client:        true,
		},
		{
			pv:            []*PVBuilder{buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume())},
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}, {LabelSelector: defaultLabelSelector}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			pv:            []*PVBuilder{buildValidPersistentVolumeTestBuilder(buildTestClientWithDummyPersistentVolume())},
			listOptions:   []metav1.ListOptions{{LabelSelector: defaultLabelSelector}},
			expectedError: fmt.Errorf("failed to list persistentVolume, 'apiClient' parameter is empty"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummyPersistentVolume()
		}

		pvBuilders, err := ListPV(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.pv), len(pvBuilders))
		}
	}
}

func TestListPVC(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*PVCBuilder, error) {
			return ListPVC(apiClient, nsname)
		},
		corev1.AddToScheme,
		persistentVolumeClaimGVK,
	).ExecuteTests(t)
}

func TestListPVCWithOptions(t *testing.T) {
	t.Parallel()

	apiClient := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPVCObject("pvctest1", defaultNamespace),
			buildDummyPVCObject("pvctest2", defaultNamespace),
			buildDummyPVCObject("pvctest3", defaultNamespace),
		},
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})

	builders, err := ListPVC(apiClient, defaultNamespace, metav1.ListOptions{LabelSelector: defaultLabelSelector})
	require.NoError(t, err)
	require.Len(t, builders, 3)

	for _, builder := range builders {
		assert.Equal(t, defaultNamespace, builder.Definition.Namespace)
	}
}

func TestListPVCMultipleOptionsReturnsError(t *testing.T) {
	t.Parallel()

	apiClient := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})

	_, err := ListPVC(apiClient, defaultNamespace, metav1.ListOptions{}, metav1.ListOptions{})
	require.EqualError(t, err, "error: more than one ListOptions was passed")
}

func TestListPVCInvalidLabelSelectorReturnsError(t *testing.T) {
	t.Parallel()

	apiClient := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})

	_, err := ListPVC(apiClient, defaultNamespace, metav1.ListOptions{LabelSelector: "demo in (("})
	require.Error(t, err)
}

func buildDummyPVCObject(name, namespace string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    defaultLabel,
		},
	}
}
