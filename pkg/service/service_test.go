package service

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const invalidBuilderShortCirc = "invalid builder short circuits"

var (
	serviceGVK              = corev1.SchemeGroupVersion.WithKind("Service")
	defaultServiceName      = "test-service-name"
	defaultServiceNamespace = "test-service-namespace"
	defaultServiceSelector  = map[string]string{"testLabel": "testLabelValue"}
	defaultServicePort      = corev1.ServicePort{
		Name:     "http",
		Protocol: corev1.ProtocolTCP,
		Port:     80,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: int32(8080),
		},
	}
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[corev1.Service, Builder](
		func(apiClient *clients.Settings, name, nsname string) *Builder {
			return NewBuilder(apiClient, name, nsname, defaultServiceSelector, defaultServicePort)
		},
		corev1.AddToScheme,
		serviceGVK,
	).ExecuteTests(t)
}

func TestNewBuilderSetsSpec(t *testing.T) {
	t.Parallel()

	builder := buildValidServiceBuilder()
	require.NoError(t, builder.GetError())
	assert.Equal(t, defaultServiceSelector, builder.Definition.Spec.Selector)
	assert.Equal(t, []corev1.ServicePort{defaultServicePort}, builder.Definition.Spec.Ports)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[corev1.Service, Builder](Pull, corev1.AddToScheme, serviceGVK).
		ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newServiceCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newServiceCommonTestConfig()).ExecuteTests(t)
}

func TestWithNodePort(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *Builder
		assertError func(error) bool
	}{
		{
			name:        "sets NodePort type and port",
			builder:     buildValidServiceBuilder,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name: "no ports sets builder error",
			builder: func() *Builder {
				builder := buildValidServiceBuilder()
				builder.Definition.Spec.Ports = nil

				return builder
			},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "service does not have the available ports"
			},
		},
		{
			name:        invalidBuilderShortCirc,
			builder:     buildInvalidServiceBuilder,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithNodePort()
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, corev1.ServiceTypeNodePort, result.Definition.Spec.Type)
				assert.Equal(t, defaultServicePort.Port, result.Definition.Spec.Ports[0].NodePort)
			}
		})
	}
}

//nolint:staticcheck // Removing deprecated types would break consumers of eco-goinfra
func TestWithExternalTrafficPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *Builder
		policyType  corev1.ServiceExternalTrafficPolicyType
		assertError func(error) bool
	}{
		{
			name:        "local policy",
			builder:     buildValidServiceBuilder,
			policyType:  corev1.ServiceExternalTrafficPolicyTypeLocal,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "cluster policy",
			builder:     buildValidServiceBuilder,
			policyType:  corev1.ServiceExternalTrafficPolicyTypeCluster,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:       "empty policy sets builder error",
			builder:    buildValidServiceBuilder,
			policyType: "",
			assertError: func(err error) bool {
				return err != nil && err.Error() == "ExternalTrafficPolicy can not be empty"
			},
		},
		{
			name:        invalidBuilderShortCirc,
			builder:     buildInvalidServiceBuilder,
			policyType:  corev1.ServiceExternalTrafficPolicyTypeLocal,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithExternalTrafficPolicy(testCase.policyType)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, corev1.ServiceTypeLoadBalancer, result.Definition.Spec.Type)
				assert.Equal(t, testCase.policyType, result.Definition.Spec.ExternalTrafficPolicy)
			} else {
				assert.Empty(t, result.Definition.Spec.ExternalTrafficPolicy)
			}
		})
	}
}

func TestWithSelector(t *testing.T) {
	t.Parallel()

	testMapSetter(t, "selector", "can not apply a selector with an empty key", "selector can not be empty",
		(*Builder).WithSelector,
		func(builder *Builder) map[string]string { return builder.Definition.Spec.Selector },
		defaultServiceSelector)
}

func TestWithLabels(t *testing.T) {
	t.Parallel()

	testMapSetter(t, "labels", "can not apply a labels with an empty key", "labels can not be empty",
		(*Builder).WithLabels,
		func(builder *Builder) map[string]string { return builder.Definition.Labels },
		nil)
}

func TestWithAnnotation(t *testing.T) {
	t.Parallel()

	testMapSetter(t, "annotation", "can not apply a annotation with an empty key", errEmptyAnnotation,
		(*Builder).WithAnnotation,
		func(builder *Builder) map[string]string { return builder.Definition.Annotations },
		nil)
}

//nolint:staticcheck // Removing deprecated types would break consumers of eco-goinfra
func TestWithIPFamily(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       func() *Builder
		ipFamily      []corev1.IPFamily
		ipStackPolicy corev1.IPFamilyPolicyType
		assertError   func(error) bool
	}{
		{
			name:          "single stack IPv4",
			builder:       buildValidServiceBuilder,
			ipFamily:      []corev1.IPFamily{corev1.IPv4Protocol},
			ipStackPolicy: corev1.IPFamilyPolicySingleStack,
			assertError:   func(err error) bool { return err == nil },
		},
		{
			name:          "require dual stack",
			builder:       buildValidServiceBuilder,
			ipFamily:      []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol},
			ipStackPolicy: corev1.IPFamilyPolicyRequireDualStack,
			assertError:   func(err error) bool { return err == nil },
		},
		{
			name:          "nil ipFamily sets builder error",
			builder:       buildValidServiceBuilder,
			ipFamily:      nil,
			ipStackPolicy: corev1.IPFamilyPolicySingleStack,
			assertError: func(err error) bool {
				return err != nil && err.Error() == "failed to set empty ipFamily"
			},
		},
		{
			name:          "empty ipStackPolicy sets builder error",
			builder:       buildValidServiceBuilder,
			ipFamily:      []corev1.IPFamily{corev1.IPv4Protocol},
			ipStackPolicy: "",
			assertError: func(err error) bool {
				return err != nil && err.Error() == failedToSetEmptyIpstackpolicy
			},
		},
		{
			name:          invalidBuilderShortCirc,
			builder:       buildInvalidServiceBuilder,
			ipFamily:      []corev1.IPFamily{corev1.IPv4Protocol},
			ipStackPolicy: corev1.IPFamilyPolicySingleStack,
			assertError:   commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := builder.WithIPFamily(testCase.ipFamily, testCase.ipStackPolicy)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.ipFamily, result.Definition.Spec.IPFamilies)
				require.NotNil(t, result.Definition.Spec.IPFamilyPolicy)
				assert.Equal(t, testCase.ipStackPolicy, *result.Definition.Spec.IPFamilyPolicy)
			} else {
				assert.Nil(t, result.Definition.Spec.IPFamilyPolicy)
			}
		})
	}
}

func TestDefineServicePort(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		port          int32
		targetPort    int32
		protocol      corev1.Protocol
		expectedError string
	}{
		{name: "TCP", port: 80, targetPort: 8080, protocol: corev1.ProtocolTCP},
		{name: "UDP", port: 80, targetPort: 8080, protocol: corev1.ProtocolUDP},
		{name: "empty protocol", port: 80, targetPort: 8080, protocol: ""},
		{name: "zero port", port: 0, targetPort: 8080, protocol: corev1.ProtocolSCTP,
			expectedError: "invalid port number"},
		{name: "port too large", port: 655350, targetPort: 8080, protocol: corev1.ProtocolSCTP,
			expectedError: "invalid port number"},
		{name: "zero target port", port: 80, targetPort: 0, protocol: corev1.ProtocolSCTP,
			expectedError: "invalid target port number"},
		{name: "target port too large", port: 80, targetPort: 655350, protocol: corev1.ProtocolSCTP,
			expectedError: "invalid target port number"},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			servicePort, err := DefineServicePort(testCase.port, testCase.targetPort, testCase.protocol)
			if testCase.expectedError != "" {
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, servicePort)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.port, servicePort.Port)
			assert.Equal(t, testCase.targetPort, servicePort.TargetPort.IntVal)
			assert.Equal(t, testCase.protocol, servicePort.Protocol)
		})
	}
}

func TestGetGVR(t *testing.T) {
	t.Parallel()

	testGVR := GetGVR()
	assert.Equal(t, "services", testGVR.Resource)
	assert.Equal(t, "v1", testGVR.Version)
	assert.Equal(t, "", testGVR.Group)
}

// testMapSetter runs the shared cases for setters that validate and assign a string map: WithSelector, WithLabels and
// WithAnnotation. initial is the value the getter returns before the setter is applied.
func testMapSetter(
	t *testing.T,
	field, emptyKeyErr, emptyMapErr string,
	setter func(*Builder, map[string]string) *Builder,
	getter func(*Builder) map[string]string,
	initial map[string]string,
) {
	t.Helper()

	testCases := []struct {
		name        string
		builder     func() *Builder
		value       map[string]string
		assertError func(error) bool
	}{
		{
			name:        "sets " + field,
			builder:     buildValidServiceBuilder,
			value:       map[string]string{"key1": "value1", "key2": "value2"},
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "empty value is allowed",
			builder:     buildValidServiceBuilder,
			value:       map[string]string{"key": ""},
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "empty key sets builder error",
			builder:     buildValidServiceBuilder,
			value:       map[string]string{"": "value"},
			assertError: func(err error) bool { return err != nil && err.Error() == emptyKeyErr },
		},
		{
			name:        "empty map sets builder error",
			builder:     buildValidServiceBuilder,
			value:       map[string]string{},
			assertError: func(err error) bool { return err != nil && err.Error() == emptyMapErr },
		},
		{
			name:        invalidBuilderShortCirc,
			builder:     buildInvalidServiceBuilder,
			value:       map[string]string{"key": "value"},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			result := setter(builder, testCase.value)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.value, getter(result))
			} else {
				assert.Equal(t, initial, getter(result))
			}
		})
	}
}

func buildValidServiceBuilder() *Builder {
	return NewBuilder(
		newServiceTestClient(), defaultServiceName, defaultServiceNamespace, defaultServiceSelector, defaultServicePort)
}

func buildInvalidServiceBuilder() *Builder {
	return NewBuilder(newServiceTestClient(), "", defaultServiceNamespace, defaultServiceSelector, defaultServicePort)
}

// newServiceCommonTestConfig returns the shared testhelper configuration for service builder tests.
func newServiceCommonTestConfig() testhelper.CommonTestConfig[corev1.Service, Builder, *corev1.Service, *Builder] {
	return testhelper.NewCommonTestConfig[corev1.Service, Builder](
		corev1.AddToScheme, serviceGVK, testhelper.ResourceScopeNamespaced)
}

// newServiceTestClient returns a fake client configured with the Service scheme.
func newServiceTestClient(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
}
