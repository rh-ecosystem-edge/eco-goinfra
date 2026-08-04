package capoa

import (
	"errors"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	hiveext "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/hiveextension/v1beta1"
	v1alpha3 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/capoa/controlplane/v1alpha3"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	oacpTestName      = "oacp-test"
	oacpTestNamespace = "oacp-test-ns"
	oacpTestDomain    = "example.com"
	oacpTestVersion   = "4.18.0"
)

var oacpTestSchemes = []clients.SchemeAttacher{
	v1alpha3.AddToScheme,
}

func TestNewOpenshiftAssistedControlPlaneBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		baseDomain    string
		distVersion   string
		replicas      int32
		client        bool
		expectedError string
	}{
		{
			name:          oacpTestName,
			namespace:     oacpTestNamespace,
			baseDomain:    oacpTestDomain,
			distVersion:   oacpTestVersion,
			replicas:      3,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     oacpTestNamespace,
			baseDomain:    oacpTestDomain,
			distVersion:   oacpTestVersion,
			replicas:      3,
			client:        true,
			expectedError: "OpenshiftAssistedControlPlane 'name' cannot be empty",
		},
		{
			name:          oacpTestName,
			namespace:     "",
			baseDomain:    oacpTestDomain,
			distVersion:   oacpTestVersion,
			replicas:      3,
			client:        true,
			expectedError: "OpenshiftAssistedControlPlane 'namespace' cannot be empty",
		},
		{
			name:          oacpTestName,
			namespace:     oacpTestNamespace,
			baseDomain:    "",
			distVersion:   oacpTestVersion,
			replicas:      3,
			client:        true,
			expectedError: "OpenshiftAssistedControlPlane 'baseDomain' cannot be empty",
		},
		{
			name:          oacpTestName,
			namespace:     oacpTestNamespace,
			baseDomain:    oacpTestDomain,
			distVersion:   "",
			replicas:      3,
			client:        true,
			expectedError: "OpenshiftAssistedControlPlane 'distributionVersion' cannot be empty",
		},
		{
			name:        oacpTestName,
			namespace:   oacpTestNamespace,
			baseDomain:  oacpTestDomain,
			distVersion: oacpTestVersion,
			replicas:    3,
			client:      false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		builder := NewOpenshiftAssistedControlPlaneBuilder(
			testSettings, testCase.name, testCase.namespace, testCase.baseDomain, testCase.distVersion, testCase.replicas)

		if testCase.client {
			assert.NotNil(t, builder)
			assert.Equal(t, testCase.expectedError, builder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, builder.Definition.Name)
				assert.Equal(t, testCase.namespace, builder.Definition.Namespace)
				assert.Equal(t, testCase.baseDomain, builder.Definition.Spec.Config.BaseDomain)
				assert.Equal(t, testCase.distVersion, builder.Definition.Spec.DistributionVersion)
				assert.Equal(t, testCase.replicas, builder.Definition.Spec.Replicas)
			}
		} else {
			assert.Nil(t, builder)
		}
	}
}

func TestPullOpenshiftAssistedControlPlane(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		client        bool
		exists        bool
		expectedError error
	}{
		{
			name:          oacpTestName,
			namespace:     oacpTestNamespace,
			client:        true,
			exists:        true,
			expectedError: nil,
		},
		{
			name:          "",
			namespace:     oacpTestNamespace,
			client:        true,
			exists:        true,
			expectedError: errors.New("OpenshiftAssistedControlPlane 'name' cannot be empty"),
		},
		{
			name:          oacpTestName,
			namespace:     "",
			client:        true,
			exists:        true,
			expectedError: errors.New("OpenshiftAssistedControlPlane 'namespace' cannot be empty"),
		},
		{
			name:          oacpTestName,
			namespace:     oacpTestNamespace,
			client:        false,
			exists:        true,
			expectedError: errors.New("the apiClient is nil"),
		},
		{
			name:      oacpTestName,
			namespace: oacpTestNamespace,
			client:    true,
			exists:    false,
			expectedError: errors.New(
				"OpenshiftAssistedControlPlane object oacp-test does not exist in namespace oacp-test-ns"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateOACP())
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects, SchemeAttachers: oacpTestSchemes})
		}

		builder, err := PullOpenshiftAssistedControlPlane(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Nil(t, builder)
		} else {
			assert.Equal(t, testCase.name, builder.Object.Name)
			assert.Equal(t, testCase.name, builder.Definition.Name)
			assert.Equal(t, testCase.namespace, builder.Object.Namespace)
			assert.Equal(t, testCase.namespace, builder.Definition.Namespace)
		}
	}
}

func TestOACPCreate(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{exists: true},
		{exists: false},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateOACP())
		}

		builder := buildOACPTestBuilderWithFakeObjects(runtimeObjects)
		result, err := builder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, oacpTestName, result.Definition.Name)
		assert.Equal(t, oacpTestNamespace, result.Definition.Namespace)

		persisted, getErr := result.Get()
		assert.Nil(t, getErr)
		assert.Equal(t, oacpTestName, persisted.Name)
		assert.Equal(t, oacpTestNamespace, persisted.Namespace)
	}
}

func TestOACPDelete(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{exists: true},
		{exists: false},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateOACP())
		}

		builder := buildOACPTestBuilderWithFakeObjects(runtimeObjects)
		err := builder.Delete()
		assert.Nil(t, err)
		assert.Nil(t, builder.Object)
	}
}

func TestOACPDeleteAndWait(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{exists: true},
		{exists: false},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateOACP())
		}

		builder := buildOACPTestBuilderWithFakeObjects(runtimeObjects)
		err := builder.DeleteAndWait(time.Second * 1)
		assert.Nil(t, err)
		assert.Nil(t, builder.Object)
	}
}

func TestOACPGet(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{exists: true},
		{exists: false},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateOACP())
		}

		builder := buildOACPTestBuilderWithFakeObjects(runtimeObjects)
		oacp, err := builder.Get()

		if testCase.exists {
			assert.Nil(t, err)
			assert.NotNil(t, oacp)
		} else {
			assert.NotNil(t, err)
			assert.Nil(t, oacp)
		}
	}
}

func TestOACPExists(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{exists: true},
		{exists: false},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateOACP())
		}

		builder := buildOACPTestBuilderWithFakeObjects(runtimeObjects)
		assert.Equal(t, testCase.exists, builder.Exists())
	}
}

func TestOACPUpdate(t *testing.T) {
	testCases := []struct {
		exists        bool
		force         bool
		staleVersion  bool
		expectedError error
	}{
		{
			exists:        true,
			force:         true,
			expectedError: nil,
		},
		{
			exists:        false,
			force:         true,
			expectedError: errors.New("cannot update non-existent OpenshiftAssistedControlPlane"),
		},
		{
			exists:        true,
			force:         true,
			staleVersion:  true,
			expectedError: nil,
		},
		{
			exists:       true,
			force:        false,
			staleVersion: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateOACP())
		}

		builder := buildOACPTestBuilderWithFakeObjects(runtimeObjects)
		assert.NotNil(t, builder)

		if testCase.exists {
			assert.True(t, builder.Exists())
			builder.Definition.Spec.Config.NetworkType = "Calico"

			if testCase.staleVersion {
				builder.Definition.ResourceVersion = "stale"
			}
		}

		result, err := builder.Update(testCase.force)

		if testCase.staleVersion && !testCase.force {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "object was modified")

			return
		}

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, result)
			assert.Equal(t, "Calico", result.Definition.Spec.Config.NetworkType)
		} else {
			assert.Nil(t, result)
		}
	}
}

func TestOACPWithNetworkType(t *testing.T) {
	testCases := []struct {
		networkType string
	}{
		{networkType: "OpenShiftSDN"},
		{networkType: "OVNKubernetes"},
		{networkType: "Calico"},
		{networkType: "Cilium"},
		{networkType: "CiscoACI"},
		{networkType: "None"},
		{networkType: ""},
	}

	for _, testCase := range testCases {
		builder := generateOACPTestBuilder()
		builder.WithNetworkType(testCase.networkType)
		assert.Equal(t, testCase.networkType, builder.Definition.Spec.Config.NetworkType)
	}
}

func TestOACPGetNetworkType(t *testing.T) {
	builder := generateOACPTestBuilder()
	builder.Definition.Spec.Config.NetworkType = "Calico"
	assert.Equal(t, "Calico", builder.GetNetworkType())
}

func TestOACPWithAPIVIPs(t *testing.T) {
	builder := generateOACPTestBuilder()
	vips := []string{"192.168.1.100", "fd00::100"}
	builder.WithAPIVIPs(vips)
	assert.Equal(t, vips, builder.Definition.Spec.Config.APIVIPs)
}

func TestOACPWithIngressVIPs(t *testing.T) {
	builder := generateOACPTestBuilder()
	vips := []string{"192.168.1.101"}
	builder.WithIngressVIPs(vips)
	assert.Equal(t, vips, builder.Definition.Spec.Config.IngressVIPs)
}

func TestOACPWithPullSecretRef(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "my-pull-secret",
			expectedError: "",
		},
		{
			name:          "",
			expectedError: "OpenshiftAssistedControlPlane pullSecretRef name cannot be empty",
		},
	}

	for _, testCase := range testCases {
		builder := generateOACPTestBuilder()
		builder.WithPullSecretRef(testCase.name)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, builder.Definition.Spec.Config.PullSecretRef.Name)
		} else {
			assert.Equal(t, testCase.expectedError, builder.errorMsg)
		}
	}
}

func TestOACPWithManifestsConfigMapRefs(t *testing.T) {
	builder := generateOACPTestBuilder()
	refs := []hiveext.ManifestsConfigMapReference{
		{Name: "calico-manifests"},
		{Name: "extra-manifests"},
	}

	builder.WithManifestsConfigMapRefs(refs)
	assert.Equal(t, refs, builder.Definition.Spec.Config.ManifestsConfigMapRefs)
}

func TestOACPWithMastersSchedulable(t *testing.T) {
	builder := generateOACPTestBuilder()
	builder.WithMastersSchedulable(true)
	assert.True(t, builder.Definition.Spec.Config.MastersSchedulable)
}

func TestOACPWithSSHAuthorizedKey(t *testing.T) {
	builder := generateOACPTestBuilder()
	key := "ssh-rsa AAAA..."
	builder.WithSSHAuthorizedKey(key)
	assert.Equal(t, key, builder.Definition.Spec.Config.SSHAuthorizedKey)
}

func TestOACPWithReplicas(t *testing.T) {
	builder := generateOACPTestBuilder()
	builder.WithReplicas(1)
	assert.Equal(t, int32(1), builder.Definition.Spec.Replicas)
}

func TestOACPWithClusterName(t *testing.T) {
	builder := generateOACPTestBuilder()
	builder.WithClusterName("my-cluster")
	assert.Equal(t, "my-cluster", builder.Definition.Spec.Config.ClusterName)
}

func TestOACPWithMachineTemplate(t *testing.T) {
	builder := generateOACPTestBuilder()
	tmpl := v1alpha3.OpenshiftAssistedControlPlaneMachineTemplate{
		InfrastructureRef: v1alpha3.ContractVersionedObjectReference{
			Kind:     "Metal3MachineTemplate",
			Name:     "my-template",
			APIGroup: "infrastructure.cluster.x-k8s.io",
		},
	}

	builder.WithMachineTemplate(tmpl)
	assert.Equal(t, tmpl, builder.Definition.Spec.MachineTemplate)
}

func TestOACPValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		errorMsg      string
		expectedError string
	}{
		{
			builderNil:    true,
			expectedError: "error: received nil OpenshiftAssistedControlPlane builder",
		},
		{
			definitionNil: true,
			expectedError: "can not redefine the undefined OpenshiftAssistedControlPlane",
		},
		{
			apiClientNil:  true,
			expectedError: "OpenshiftAssistedControlPlane builder cannot have nil apiClient",
		},
		{
			errorMsg:      "test error",
			expectedError: "test error",
		},
		{
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var builder *OpenshiftAssistedControlPlaneBuilder

		if !testCase.builderNil {
			builder = generateOACPTestBuilder()
		}

		if testCase.definitionNil && builder != nil {
			builder.Definition = nil
		}

		if testCase.apiClientNil && builder != nil {
			builder.apiClient = nil
		}

		if testCase.errorMsg != "" && builder != nil {
			builder.errorMsg = testCase.errorMsg
		}

		valid, err := builder.validate()

		if testCase.expectedError != "" {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

func buildOACPTestBuilderWithFakeObjects(objects []runtime.Object) *OpenshiftAssistedControlPlaneBuilder {
	fakeClientScheme := runtime.NewScheme()

	err := clients.SetScheme(fakeClientScheme)
	if err != nil {
		return nil
	}

	apiClient := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: oacpTestSchemes,
	})

	return NewOpenshiftAssistedControlPlaneBuilder(
		apiClient, oacpTestName, oacpTestNamespace, oacpTestDomain, oacpTestVersion, 3)
}

func generateOACPTestBuilder() *OpenshiftAssistedControlPlaneBuilder {
	return &OpenshiftAssistedControlPlaneBuilder{
		apiClient:  clients.GetTestClients(clients.TestClientParams{}).Client,
		Definition: generateOACP(),
	}
}

func generateOACP() *v1alpha3.OpenshiftAssistedControlPlane {
	return &v1alpha3.OpenshiftAssistedControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      oacpTestName,
			Namespace: oacpTestNamespace,
		},
		Spec: v1alpha3.OpenshiftAssistedControlPlaneSpec{
			Config: v1alpha3.OpenshiftAssistedControlPlaneConfigSpec{
				BaseDomain: oacpTestDomain,
			},
			DistributionVersion: oacpTestVersion,
			Replicas:            3,
		},
	}
}
