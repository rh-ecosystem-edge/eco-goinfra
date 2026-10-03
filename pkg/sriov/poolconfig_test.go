package sriov

import (
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	defaultPoolConfigName   = "poolconfig"
	defaultPoolConfigNsName = "testnamespace"
)

func TestNewPoolConfigBuilder(t *testing.T) {
	testCases := []struct {
		poolConfigName      string
		poolConfigNamespace string
		expectedErrorText   string
		client              bool
	}{
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: defaultPoolConfigNsName,
			client:              true,
		},
		{
			poolConfigName:      "",
			poolConfigNamespace: defaultPoolConfigNsName,
			expectedErrorText:   errEmptyName,
			client:              true,
		},
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: "",
			expectedErrorText:   "SriovNetworkPoolConfig 'nsname' cannot be empty",
			client:              true,
		},
	}
	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testPoolconfigStructure := NewPoolConfigBuilder(
			testSettings, testCase.poolConfigName, testCase.poolConfigNamespace)
		assert.NotNil(t, testPoolconfigStructure)

		if len(testCase.expectedErrorText) > 0 {
			assert.Equal(t, testCase.expectedErrorText, testPoolconfigStructure.errorMsg)
		}
	}
}

func TestPooConfigCreate(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedError  error
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testPoolConfig: buildInvalidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  fmt.Errorf("SriovNetworkPoolConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testPoolConfig.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition, testBuilder.Object)
		}
	}
}

func TestPooConfigDelete(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedError  error
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testPoolConfig: buildInvalidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  fmt.Errorf("SriovNetworkPoolConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testPoolConfig.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testPoolConfig.Object)
		}
	}
}

func TestPooConfigExist(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedStatus bool
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testPoolConfig: buildInvalidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		_, _ = testCase.testPoolConfig.Create()
		assert.Equal(t, testCase.expectedStatus, testCase.testPoolConfig.Exists())
	}
}

func TestTestPooConfigGet(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedError  error
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testPoolConfig: buildInvalidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  fmt.Errorf("SriovNetworkPoolConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, _ = testCase.testPoolConfig.Create()
		testBuilder, err := testCase.testPoolConfig.Get()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, testBuilder)
		}
	}
}

func TestPoolConfigUpdate(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedError  error
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  nil,
		},
	}
	for _, testCase := range testCases {
		poolConfigBuilder, err := testCase.testPoolConfig.WithMaxUnavailable(intstr.FromInt32(2)).Create()
		assert.Nil(t, err)
		assert.Equal(t, int32(2), poolConfigBuilder.Definition.Spec.MaxUnavailable.IntVal)
		testCase.testPoolConfig.WithMaxUnavailable(intstr.FromString("100%"))

		poolConfigBuilder.Definition.ResourceVersion = "999"
		poolConfigBuilder, err = poolConfigBuilder.Update()
		assert.Nil(t, err)
		assert.Equal(t, "100%", poolConfigBuilder.Object.Spec.MaxUnavailable.StrVal)
		assert.Equal(t, poolConfigBuilder.Definition, poolConfigBuilder.Object)
	}
}

func TestWithNodeSelector(t *testing.T) {
	testCases := []struct {
		nodeSelector      map[string]string
		expectedErrorText string
	}{
		{
			nodeSelector:      map[string]string{"test": "test"},
			expectedErrorText: "",
		},
		{
			nodeSelector:      map[string]string{},
			expectedErrorText: "SriovNetworkPoolConfig 'nodeSelector' cannot be empty map",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestPoolConfigClientWithDummyObject()
		testBuilder := buildValidPoolConfigTestBuilder(testSettings).WithNodeSelector(testCase.nodeSelector)
		assert.Equal(t, testCase.expectedErrorText, testBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testBuilder.Definition.Spec.NodeSelector.MatchLabels, testCase.nodeSelector)
		}
	}
}

func TestWithMaxUnavailable(t *testing.T) {
	testCases := []struct {
		maxUnavailable    intstr.IntOrString
		expectedErrorText string
	}{
		{
			maxUnavailable:    intstr.FromInt32(3),
			expectedErrorText: "",
		},
		{
			maxUnavailable:    intstr.FromString("wrongValue"),
			expectedErrorText: "invalid type: strings needs to be a percentage: {1 0 wrongValue}",
		},
		{
			maxUnavailable:    intstr.FromString("99s%"),
			expectedErrorText: "invalid value \"99s%\": strconv.Atoi: parsing \"99s\": invalid syntax",
		},
		{
			maxUnavailable:    intstr.FromString("101%"),
			expectedErrorText: "invalid value: percentage needs to be between 1 and 100: {1 0 101%}",
		},
		{
			maxUnavailable:    intstr.FromInt32(-2),
			expectedErrorText: "negative number is not allowed: {0 -2 }",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestPoolConfigClientWithDummyObject()
		testBuilder := buildValidPoolConfigTestBuilder(testSettings).WithMaxUnavailable(testCase.maxUnavailable)
		assert.Equal(t, testCase.expectedErrorText, testBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testBuilder.Definition.Spec.MaxUnavailable.IntVal, testCase.maxUnavailable.IntVal)
		}
	}
}

func TestWithRDMAMode(t *testing.T) {
	testCases := []struct {
		RdmaMode      string
		ExpectedError string
	}{
		{
			RdmaMode:      "shared",
			ExpectedError: "",
		},
		{
			RdmaMode:      "exclusive",
			ExpectedError: "",
		},
		{
			RdmaMode:      "",
			ExpectedError: "rdmaMode cannot be empty",
		},
		{
			RdmaMode:      "none",
			ExpectedError: "invalid value for rdmaMode. It should be 'shared' or 'exclusive'",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestPoolConfigClientWithDummyObject()
		testBuilder := buildValidPoolConfigTestBuilder(testSettings).WithRDMAMode(testCase.RdmaMode)
		assert.Equal(t, testCase.ExpectedError, testBuilder.errorMsg)

		if testCase.ExpectedError == "" {
			assert.Equal(t, testBuilder.Definition.Spec.RdmaMode, testCase.RdmaMode)
		}
	}
}

func TestPullPoolConfig(t *testing.T) {
	testCases := []struct {
		poolConfigName      string
		poolConfigNamespace string
		expectedErrorText   string
		expectedError       bool
		client              bool
	}{
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: defaultPoolConfigNsName,
			expectedError:       false,
			client:              true,
		},
		{
			poolConfigName:      "",
			poolConfigNamespace: defaultPoolConfigNsName,
			expectedErrorText:   errEmptyName,
			expectedError:       true,
			client:              true,
		},
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: "",
			expectedErrorText:   "SriovNetworkPoolConfig 'namespace' cannot be empty",
			expectedError:       true,
			client:              true,
		},
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: defaultPoolConfigNsName,
			expectedErrorText:   "SriovNetworkPoolConfig 'apiClient' cannot be empty",
			expectedError:       true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyPoolConfigObject(),
				SchemeAttachers: testSchemes,
			})
			_, _ = buildValidPoolConfigTestBuilder(testSettings).Create()
		}

		testPoolConfig, err := PullPoolConfig(testSettings, testCase.poolConfigName, testCase.poolConfigNamespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.poolConfigName, testPoolConfig.Object.Name)
			assert.Equal(t, testCase.poolConfigNamespace, testPoolConfig.Object.Namespace)
		}
	}
}

func TestPoolConfigWaitUntilStable(t *testing.T) {
	testCases := []struct {
		name         string
		pool         *srIovV1.SriovNetworkPoolConfig
		expectStable bool
	}{
		{
			name:         "stable Ready True without Progressing or Degraded",
			pool:         buildStablePoolConfig(),
			expectStable: true,
		},
		{
			name:         "missing Ready does not pass",
			pool:         buildDummyPoolConfig(defaultPoolConfigName, defaultPoolConfigNsName),
			expectStable: false,
		},
		{
			name: "stale observedGeneration does not pass",
			pool: buildPoolConfigWithConditions(
				2, metav1.ConditionTrue, nil, nil),
			expectStable: false,
		},
		{
			name: "Progressing True is not stable",
			pool: buildPoolConfigWithConditions(
				1, metav1.ConditionTrue, condStatusPtr(metav1.ConditionTrue), nil),
			expectStable: false,
		},
		{
			name: "Degraded True is not stable",
			pool: buildPoolConfigWithConditions(
				1, metav1.ConditionTrue, nil, condStatusPtr(metav1.ConditionTrue)),
			expectStable: false,
		},
		{
			name: "Ready False is not stable",
			pool: buildPoolConfigWithConditions(
				1, metav1.ConditionFalse, nil, nil),
			expectStable: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{testCase.pool},
				SchemeAttachers: testSchemes,
			})
			builder := NewPoolConfigBuilder(testSettings, defaultPoolConfigName, defaultPoolConfigNsName)
			timeout := 2 * time.Second
			err := builder.WaitUntilStable(timeout)

			if testCase.expectStable {
				assert.NoError(t, err)
				assert.Equal(t, metav1.ConditionTrue, builder.GetCondition(srIovV1.ConditionReady).Status)

				return
			}

			assert.Error(t, err)
		})
	}
}

func TestPoolConfigWaitForCondition(t *testing.T) {
	pool := buildStablePoolConfig()
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{pool},
		SchemeAttachers: testSchemes,
	})
	builder := NewPoolConfigBuilder(testSettings, defaultPoolConfigName, defaultPoolConfigNsName)

	err := builder.WaitForCondition(metav1.Condition{
		Type:   srIovV1.ConditionReady,
		Status: metav1.ConditionTrue,
		Reason: srIovV1.ReasonPoolConfigReady,
	}, 2*time.Second)
	assert.NoError(t, err)

	err = builder.WaitForCondition(metav1.Condition{
		Type:   srIovV1.ConditionReady,
		Status: metav1.ConditionFalse,
	}, 2*time.Second)
	assert.Error(t, err)

	err = builder.WaitForCondition(metav1.Condition{
		Type:               srIovV1.ConditionReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: 2,
	}, 2*time.Second)
	assert.Error(t, err)
}

func TestPoolConfigWaitInvalidBuilder(t *testing.T) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
	timeout := time.Second
	emptyNameErr := errEmptyName
	nilBuilderErr := "error: received nil SriovNetworkPoolConfig builder"

	t.Run("empty name WaitUntilStable", func(t *testing.T) {
		builder := NewPoolConfigBuilder(testSettings, "", defaultPoolConfigNsName)

		err := builder.WaitUntilStable(timeout)
		assert.EqualError(t, err, emptyNameErr)
	})

	t.Run("empty name WaitForCondition", func(t *testing.T) {
		builder := NewPoolConfigBuilder(testSettings, "", defaultPoolConfigNsName)

		err := builder.WaitForCondition(metav1.Condition{Type: srIovV1.ConditionReady}, timeout)
		assert.EqualError(t, err, emptyNameErr)
	})

	t.Run("nil builder WaitUntilStable", func(t *testing.T) {
		err := (*PoolConfigBuilder)(nil).WaitUntilStable(timeout)
		assert.EqualError(t, err, nilBuilderErr)
	})

	t.Run("nil builder WaitForCondition", func(t *testing.T) {
		err := (*PoolConfigBuilder)(nil).WaitForCondition(
			metav1.Condition{Type: srIovV1.ConditionReady}, timeout)
		assert.EqualError(t, err, nilBuilderErr)
	})
}

func TestPoolConfigGetConditionAndCounts(t *testing.T) {
	t.Run("nil builder", func(t *testing.T) {
		assert.Nil(t, (*PoolConfigBuilder)(nil).GetCondition(srIovV1.ConditionReady))
		assert.Nil(t, (*PoolConfigBuilder)(nil).GetMatchedNodeCount())
		assert.Nil(t, (*PoolConfigBuilder)(nil).GetReadyNodeCount())
	})

	t.Run("returns matching condition and counts", func(t *testing.T) {
		pool := buildStablePoolConfig()
		builder := NewPoolConfigBuilder(
			clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{pool},
				SchemeAttachers: testSchemes,
			}),
			defaultPoolConfigName, defaultPoolConfigNsName)
		assert.True(t, builder.Exists())

		cond := builder.GetCondition(srIovV1.ConditionReady)
		assert.NotNil(t, cond)
		assert.Equal(t, metav1.ConditionTrue, cond.Status)
		assert.Equal(t, int32(2), *builder.GetMatchedNodeCount())
		assert.Equal(t, int32(2), *builder.GetReadyNodeCount())
	})
}

// buildValidPoolConfigTestBuilder returns a valid Builder for testing purposes.
func buildValidPoolConfigTestBuilder(apiClient *clients.Settings) *PoolConfigBuilder {
	return NewPoolConfigBuilder(apiClient, defaultPoolConfigName, defaultPoolConfigNsName)
}

// buildInvalidPoolConfigTestBuilder returns an invalid Builder for testing purposes.
func buildInvalidPoolConfigTestBuilder(apiClient *clients.Settings) *PoolConfigBuilder {
	return NewPoolConfigBuilder(apiClient, defaultPoolConfigName, "")
}

func buildTestPoolConfigClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPoolConfigObject(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyPoolConfigObject() []runtime.Object {
	return append([]runtime.Object{}, buildDummyPoolConfig(defaultPoolConfigName, defaultPoolConfigNsName))
}

func buildDummyPoolConfig(name, nsName string) *srIovV1.SriovNetworkPoolConfig {
	return &srIovV1.SriovNetworkPoolConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsName,
		},
		Spec: srIovV1.SriovNetworkPoolConfigSpec{},
	}
}

func buildStablePoolConfig() *srIovV1.SriovNetworkPoolConfig {
	return buildPoolConfigWithConditions(1, metav1.ConditionTrue, nil, nil)
}

func condStatusPtr(status metav1.ConditionStatus) *metav1.ConditionStatus {
	return &status
}

func buildPoolConfigWithConditions(
	generation int64,
	ready metav1.ConditionStatus,
	progressing, degraded *metav1.ConditionStatus,
) *srIovV1.SriovNetworkPoolConfig {
	const observedGeneration int64 = 1

	pool := buildDummyPoolConfig(defaultPoolConfigName, defaultPoolConfigNsName)
	pool.Generation = generation
	matched := int32(2)
	readyCount := int32(2)
	pool.Status.MatchedNodeCount = &matched
	pool.Status.ReadyNodeCount = &readyCount
	pool.Status.Conditions = []metav1.Condition{
		{
			Type:               srIovV1.ConditionReady,
			Status:             ready,
			Reason:             srIovV1.ReasonPoolConfigReady,
			ObservedGeneration: observedGeneration,
		},
	}

	if progressing != nil {
		pool.Status.Conditions = append(pool.Status.Conditions, metav1.Condition{
			Type:               srIovV1.ConditionProgressing,
			Status:             *progressing,
			ObservedGeneration: observedGeneration,
		})
	}

	if degraded != nil {
		pool.Status.Conditions = append(pool.Status.Conditions, metav1.Condition{
			Type:               conditionDegraded,
			Status:             *degraded,
			ObservedGeneration: observedGeneration,
		})
	}

	return pool
}
