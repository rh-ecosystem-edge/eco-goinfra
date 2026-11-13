package siteconfig

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	aiv1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	siteconfigv1alpha1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/siteconfig/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	// defaultClusterInstanceCondition depicts the default ClusterInstance condition value.
	defaultClusterInstanceCondition = metav1.Condition{Type: "", Status: metav1.ConditionTrue}

	// defaultClusterInstanceReinstallCondition depicts the default ClusterInstance reinstall condition value.
	defaultClusterInstanceReinstallCondition = metav1.Condition{Type: "", Status: metav1.ConditionTrue}
)

const testClusterInstance = "test-cluster-instance"

var testSchemes = []clients.SchemeAttacher{siteconfigv1alpha1.AddToScheme}

func TestNewClusterInstanceBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[siteconfigv1alpha1.ClusterInstance, CIBuilder](
		NewCIBuilder,
		siteconfigv1alpha1.AddToScheme,
		clusterInstanceGVK,
	).ExecuteTests(t)
}

func TestClusterInstancePull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[siteconfigv1alpha1.ClusterInstance, CIBuilder](
		PullClusterInstance,
		siteconfigv1alpha1.AddToScheme,
		clusterInstanceGVK,
	).ExecuteTests(t)
}

func TestClusterInstanceBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[siteconfigv1alpha1.ClusterInstance, CIBuilder](
		siteconfigv1alpha1.AddToScheme,
		clusterInstanceGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestClusterInstanceWithExtraManifests(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		extraManifest    string
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name:          "sets extra manifest",
			extraManifest: "ci-extra-manifest",
		},
		{
			name:             "empty extra manifest",
			expectedErrorMsg: "clusterinstance extramanifest cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			extraManifest:  "ci-extra-manifest",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithExtraManifests(testCase.extraManifest)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.ExtraManifestsRefs)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.Len(t, testBuilder.Definition.Spec.ExtraManifestsRefs, 1)
			assert.Equal(t, testCase.extraManifest, testBuilder.Definition.Spec.ExtraManifestsRefs[0].Name)
		})
	}
}

func TestClusterInstanceWithCABundle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		caBundle         string
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name:     "sets CA bundle",
			caBundle: "ibi-ca-bundle",
		},
		{
			name:             "empty CA bundle",
			expectedErrorMsg: "clusterinstance cabundle cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			caBundle:       "ibi-ca-bundle",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithCABundle(testCase.caBundle)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.CaBundleRef)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.NotNil(t, testBuilder.Definition.Spec.CaBundleRef)
			assert.Equal(t, testCase.caBundle, testBuilder.Definition.Spec.CaBundleRef.Name)
		})
	}
}

func TestClusterInstanceWithPullSecretRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		secretRef        string
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name:      "sets pull secret reference",
			secretRef: "test-secret",
		},
		{
			name:             "empty pull secret reference",
			expectedErrorMsg: "clusterinstance secretRef cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			secretRef:      "test-secret",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithPullSecretRef(testCase.secretRef)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.PullSecretRef.Name)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.secretRef, testBuilder.Definition.Spec.PullSecretRef.Name)
		})
	}
}

func TestClusterInstanceWithClusterTemplateRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		templateName      string
		templateNamespace string
		expectedErrorMsg  string
		invalidBuilder    bool
	}{
		{
			name:              "sets cluster template reference",
			templateName:      "test-template",
			templateNamespace: "test-namespace",
		},
		{
			name:              "empty template name",
			templateNamespace: "test-namespace",
			expectedErrorMsg:  "clusterinstance clusterTemplateName cannot be empty",
		},
		{
			name:             "empty template namespace",
			templateName:     "test-template",
			expectedErrorMsg: "clusterinstance clusterTemplateNamespace cannot be empty",
		},
		{
			name:              "invalid builder short circuits",
			templateName:      "test-template",
			templateNamespace: "test-namespace",
			invalidBuilder:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithClusterTemplateRef(testCase.templateName, testCase.templateNamespace)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.TemplateRefs)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.Len(t, testBuilder.Definition.Spec.TemplateRefs, 1)
			assert.Equal(t, testCase.templateName, testBuilder.Definition.Spec.TemplateRefs[0].Name)
			assert.Equal(t, testCase.templateNamespace, testBuilder.Definition.Spec.TemplateRefs[0].Namespace)
		})
	}
}

func TestClusterInstanceWithBaseDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		baseDomain       string
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name:       "sets base domain",
			baseDomain: "example.domain.com",
		},
		{
			name:             "empty base domain",
			expectedErrorMsg: "clusterinstance baseDomain cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			baseDomain:     "example.domain.com",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithBaseDomain(testCase.baseDomain)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.BaseDomain)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.baseDomain, testBuilder.Definition.Spec.BaseDomain)
		})
	}
}

func TestClusterInstanceWithClusterImageSetRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		imageSet         string
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name:     "sets cluster image set reference",
			imageSet: "quay.io/example/img:latest",
		},
		{
			name:             "empty cluster image set reference",
			expectedErrorMsg: "clusterinstance imageSet cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			imageSet:       "quay.io/example/img:latest",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithClusterImageSetRef(testCase.imageSet)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.ClusterImageSetNameRef)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.imageSet, testBuilder.Definition.Spec.ClusterImageSetNameRef)
		})
	}
}

func TestClusterInstanceWithClusterName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		clusterName      string
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name:        "sets cluster name",
			clusterName: "test-cluster",
		},
		{
			name:             "empty cluster name",
			expectedErrorMsg: "clusterinstance clusterName cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			clusterName:    "test-cluster",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithClusterName(testCase.clusterName)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.ClusterName)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.clusterName, testBuilder.Definition.Spec.ClusterName)
		})
	}
}

func TestClusterInstanceWithSSHPubKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		sshPubKey        string
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name:      "sets SSH public key",
			sshPubKey: "mykey",
		},
		{
			name:             "empty SSH public key",
			expectedErrorMsg: "clusterinstance sshPubKey cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			sshPubKey:      "mykey",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithSSHPubKey(testCase.sshPubKey)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.SSHPublicKey)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.sshPubKey, testBuilder.Definition.Spec.SSHPublicKey)
		})
	}
}

func TestClusterInstanceWithMachineNetwork(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		network          string
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name:    "sets IPv4 machine network",
			network: "192.168.122.0/24",
		},
		{
			name:    "sets IPv6 machine network",
			network: "2001:db8:abcd:1::/64",
		},
		{
			name:             "invalid CIDR",
			network:          "192.168.122.0",
			expectedErrorMsg: "clusterinstance contains invalid machineNetwork cidr",
		},
		{
			name:           "invalid builder short circuits",
			network:        "192.168.122.0/24",
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithMachineNetwork(testCase.network)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.MachineNetwork)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.Len(t, testBuilder.Definition.Spec.MachineNetwork, 1)
			assert.Equal(t, testCase.network, testBuilder.Definition.Spec.MachineNetwork[0].CIDR)
		})
	}
}

func TestClusterInstanceWithAdditionalNTPSources(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		ntpSources       []string
		expectedErrorMsg string
		expectedSources  []string
		invalidBuilder   bool
	}{
		{
			name:            "sets NTP sources",
			ntpSources:      []string{"0.pool.ntp.org", "1.pool.ntp.org"},
			expectedSources: []string{"0.pool.ntp.org", "1.pool.ntp.org"},
		},
		{
			name:            "trims NTP sources",
			ntpSources:      []string{" 2.pool.ntp.org "},
			expectedSources: []string{"2.pool.ntp.org"},
		},
		{
			name:             "empty NTP source list",
			ntpSources:       []string{},
			expectedErrorMsg: "clusterinstance additional NTP sources cannot be empty",
		},
		{
			name:             "empty NTP source",
			ntpSources:       []string{" "},
			expectedErrorMsg: "clusterinstance additional NTP source cannot be empty",
		},
		{
			name:           "invalid builder short circuits",
			ntpSources:     []string{"0.pool.ntp.org"},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithAdditionalNTPSources(testCase.ntpSources)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.AdditionalNTPSources)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.expectedSources, testBuilder.Definition.Spec.AdditionalNTPSources)
		})
	}
}

func TestClusterInstanceWithProxy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		proxy            *aiv1beta1.Proxy
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name: "sets proxy",
			proxy: &aiv1beta1.Proxy{
				HTTPProxy:  "http://myproxy.com:3128",
				HTTPSProxy: "http://myproxy.com:3128",
				NoProxy:    ".dont.proxy.this.com",
			},
		},
		{
			name:             "nil proxy",
			expectedErrorMsg: "clusterinstance proxy cannot be nil",
		},
		{
			name: "invalid builder short circuits",
			proxy: &aiv1beta1.Proxy{
				HTTPProxy: "http://myproxy.com:3128",
			},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithProxy(testCase.proxy)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Nil(t, testBuilder.Definition.Spec.Proxy)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.proxy, testBuilder.Definition.Spec.Proxy)
		})
	}
}

func TestClusterInstanceWithNode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		node             *siteconfigv1alpha1.NodeSpec
		expectedErrorMsg string
		invalidBuilder   bool
	}{
		{
			name: "sets node",
			node: generateNodeSpec(),
		},
		{
			name:             "nil node",
			expectedErrorMsg: "clusterinstance node cannot be nil",
		},
		{
			name:           "invalid builder short circuits",
			node:           generateNodeSpec(),
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithNode(testCase.node)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.Nodes)

				return
			}

			if testCase.expectedErrorMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrorMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.Len(t, testBuilder.Definition.Spec.Nodes, 1)
			assert.Equal(t, *testCase.node, testBuilder.Definition.Spec.Nodes[0])
		})
	}
}

func TestClusterInstanceWithExtraLabels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		testKey        string
		testLabels     map[string]string
		expectedErrMsg string
		invalidBuilder bool
	}{
		{
			name:       "sets extra labels",
			testKey:    "test-key",
			testLabels: map[string]string{"test-label-key": "test-label-value"},
		},
		{
			name:           "empty kind key",
			testLabels:     map[string]string{"test-label-key": "test-label-value"},
			expectedErrMsg: "can not apply empty key",
		},
		{
			name:           "empty label key",
			testKey:        "test-key",
			testLabels:     map[string]string{"": "test-label-value"},
			expectedErrMsg: "can not apply a labels with an empty key",
		},
		{
			name:           "empty labels",
			testKey:        "test-key",
			testLabels:     map[string]string{},
			expectedErrMsg: "labels can not be empty",
		},
		{
			name:           "invalid builder short circuits",
			testKey:        "test-key",
			testLabels:     map[string]string{"test-label-key": "test-label-value"},
			invalidBuilder: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			if testCase.invalidBuilder {
				testBuilder = buildInvalidClusterInstanceTestBuilder(getTestClusterInstanceAPIClient())
			}

			testBuilder.WithExtraLabels(testCase.testKey, testCase.testLabels)

			if testCase.invalidBuilder {
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
				assert.Empty(t, testBuilder.Definition.Spec.ExtraLabels)

				return
			}

			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedErrMsg)

				return
			}

			require.NoError(t, testBuilder.GetError())
			assert.Equal(t, testCase.testLabels, testBuilder.Definition.Spec.ExtraLabels[testCase.testKey])
		})
	}
}

func TestClusterInstanceWaitForCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		condition     metav1.Condition
		exists        bool
		conditionMet  bool
		valid         bool
		expectedError error
	}{
		{
			name:         "condition found",
			condition:    defaultClusterInstanceCondition,
			exists:       true,
			conditionMet: true,
			valid:        true,
		},
		{
			name:         "resource does not exist",
			condition:    defaultClusterInstanceCondition,
			conditionMet: true,
			valid:        true,
			expectedError: fmt.Errorf("clusterinstance object %s does not exist in namespace %s",
				testClusterInstance, testClusterInstance),
		},
		{
			name:          "condition times out",
			condition:     defaultClusterInstanceCondition,
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:         "invalid builder",
			condition:    defaultClusterInstanceCondition,
			exists:       true,
			conditionMet: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				clusterInstance := generateClusterInstance()

				if testCase.conditionMet {
					clusterInstance.Status.Conditions = append(clusterInstance.Status.Conditions, testCase.condition)
				}

				runtimeObjects = append(runtimeObjects, clusterInstance)
			}

			testSettings := buildClusterInstanceTestClient(runtimeObjects...)

			var clusterInstanceBuilder *CIBuilder
			if testCase.valid {
				clusterInstanceBuilder = buildValidClusterInstanceTestBuilder(testSettings)
			} else {
				clusterInstanceBuilder = buildInvalidClusterInstanceTestBuilder(testSettings)
			}

			_, err := clusterInstanceBuilder.WaitForCondition(testCase.condition, time.Second)

			if !testCase.valid {
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))

				return
			}

			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestClusterInstanceWaitForReinstallCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		condition     metav1.Condition
		exists        bool
		conditionMet  bool
		valid         bool
		expectedError error
	}{
		{
			name:         "reinstall condition found",
			condition:    defaultClusterInstanceReinstallCondition,
			exists:       true,
			conditionMet: true,
			valid:        true,
		},
		{
			name:         "resource does not exist",
			condition:    defaultClusterInstanceReinstallCondition,
			conditionMet: true,
			valid:        true,
			expectedError: fmt.Errorf("clusterinstance object %s does not exist in namespace %s",
				testClusterInstance, testClusterInstance),
		},
		{
			name:          "reinstall condition times out",
			condition:     defaultClusterInstanceReinstallCondition,
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:         "invalid builder",
			condition:    defaultClusterInstanceReinstallCondition,
			exists:       true,
			conditionMet: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				clusterInstance := generateClusterInstance()

				if testCase.conditionMet {
					clusterInstance.Status.Reinstall = &siteconfigv1alpha1.ReinstallStatus{
						Conditions: []metav1.Condition{testCase.condition},
					}
				}

				runtimeObjects = append(runtimeObjects, clusterInstance)
			}

			testSettings := buildClusterInstanceTestClient(runtimeObjects...)

			var clusterInstanceBuilder *CIBuilder
			if testCase.valid {
				clusterInstanceBuilder = buildValidClusterInstanceTestBuilder(testSettings)
			} else {
				clusterInstanceBuilder = buildInvalidClusterInstanceTestBuilder(testSettings)
			}

			_, err := clusterInstanceBuilder.WaitForReinstallCondition(testCase.condition, time.Second)

			if !testCase.valid {
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))

				return
			}

			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestClusterInstanceWaitForExtraLabel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		valid         bool
		hasLabel      bool
		expectedError error
	}{
		{
			name:     "extra label found",
			exists:   true,
			valid:    true,
			hasLabel: true,
		},
		{
			name:     "resource does not exist",
			valid:    true,
			hasLabel: true,
			expectedError: fmt.Errorf("clusterinstance object %s does not exist in namespace %s",
				testClusterInstance, testClusterInstance),
		},
		{
			name:     "invalid builder",
			exists:   true,
			hasLabel: true,
		},
		{
			name:          "extra label times out",
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				clusterInstance := generateClusterInstance()

				if testCase.hasLabel {
					clusterInstance.Spec.ExtraLabels = map[string]map[string]string{"test": {"test": ""}}
				}

				runtimeObjects = append(runtimeObjects, clusterInstance)
			}

			testSettings := buildClusterInstanceTestClient(runtimeObjects...)

			var clusterInstanceBuilder *CIBuilder
			if testCase.valid {
				clusterInstanceBuilder = buildValidClusterInstanceTestBuilder(testSettings)
			} else {
				clusterInstanceBuilder = buildInvalidClusterInstanceTestBuilder(testSettings)
			}

			_, err := clusterInstanceBuilder.WaitForExtraLabel("test", "test", time.Second)

			if !testCase.valid {
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))

				return
			}

			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func generateClusterInstance() *siteconfigv1alpha1.ClusterInstance {
	return &siteconfigv1alpha1.ClusterInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testClusterInstance,
			Namespace: testClusterInstance,
		},
	}
}

func generateNodeSpec() *siteconfigv1alpha1.NodeSpec {
	return &siteconfigv1alpha1.NodeSpec{
		HostName:       "mynode",
		BmcAddress:     "https+redfish://mybmcaddress",
		BootMACAddress: "00:00:00:00:00:00",
		BmcCredentialsName: siteconfigv1alpha1.BmcCredentialsName{
			Name: "mycreds",
		},
		TemplateRefs: []siteconfigv1alpha1.TemplateRef{
			{
				Name:      "test-template",
				Namespace: "test-namespace",
			},
		},
	}
}

func buildClusterInstanceTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: testSchemes,
	})
}

// buildValidClusterInstanceTestBuilder returns a valid ClusterInstanceBuilder for testing purposes.
func buildValidClusterInstanceTestBuilder(apiClient *clients.Settings) *CIBuilder {
	return NewCIBuilder(apiClient, testClusterInstance, testClusterInstance)
}

// buildInvalidClusterInstanceTestBuilder returns an invalid ClusterInstanceBuilder for testing purposes.
func buildInvalidClusterInstanceTestBuilder(apiClient *clients.Settings) *CIBuilder {
	return NewCIBuilder(apiClient, testClusterInstance, "")
}

func getTestClusterInstanceAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}
