package secret

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

const (
	testDataKey             = "key"
	testDataValue           = "value"
	invalidBuilderShortCirc = "invalid builder short circuits"
)

var (
	secretGVK              = corev1.SchemeGroupVersion.WithKind("Secret")
	defaultSecretName      = "test-name"
	defaultSecretNamespace = "test-namespace"
	defaultSecretType      = corev1.SecretType("test-secretType")
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		namespace   string
		secretType  corev1.SecretType
		nilClient   bool
		assertError func(error) bool
	}{
		{
			name:        defaultSecretName,
			namespace:   defaultSecretNamespace,
			secretType:  defaultSecretType,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name:        "",
			namespace:   defaultSecretNamespace,
			secretType:  defaultSecretType,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
		{
			name:        defaultSecretName,
			namespace:   "",
			secretType:  defaultSecretType,
			assertError: commonerrors.IsBuilderNamespaceEmpty,
		},
		{
			name:       defaultSecretName,
			namespace:  defaultSecretNamespace,
			secretType: "",
			assertError: func(err error) bool {
				return err != nil && err.Error() == "secret 'secretType' cannot be empty"
			},
		},
		{
			name:        defaultSecretName,
			namespace:   defaultSecretNamespace,
			secretType:  defaultSecretType,
			nilClient:   true,
			assertError: commonerrors.IsAPIClientNil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var apiClient *clients.Settings
			if !testCase.nilClient {
				apiClient = newSecretTestClient()
			}

			builder := NewBuilder(apiClient, testCase.name, testCase.namespace, testCase.secretType)
			require.NotNil(t, builder)
			require.Truef(t, testCase.assertError(builder.GetError()),
				"unexpected error: %v", builder.GetError())

			if builder.GetError() == nil {
				assert.Equal(t, testCase.name, builder.Definition.Name)
				assert.Equal(t, testCase.namespace, builder.Definition.Namespace)
				assert.Equal(t, testCase.secretType, builder.Definition.Type)
			}
		})
	}
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[corev1.Secret, Builder](Pull, corev1.AddToScheme, secretGVK).
		ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newSecretCommonTestConfig()

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

	builder := NewBuilder(newSecretTestClient(), defaultSecretName, defaultSecretNamespace, defaultSecretType)
	require.NotNil(t, builder)

	result := builder.WithOptions(func(b *Builder) (*Builder, error) {
		return b, nil
	})
	assert.Nil(t, result.GetError())

	result = builder.WithOptions(func(b *Builder) (*Builder, error) {
		return b, fmt.Errorf("option error")
	})
	assert.EqualError(t, result.GetError(), "option error")
}

func TestWithData(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       func() *Builder
		data          map[string][]byte
		assertError   func(error) bool
		expectedData  map[string][]byte
		expectDataNil bool
	}{
		{
			name: "sets data on valid builder",
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), defaultSecretName, defaultSecretNamespace, defaultSecretType)
			},
			data:         map[string][]byte{testDataKey: []byte(testDataValue)},
			assertError:  func(err error) bool { return err == nil },
			expectedData: map[string][]byte{testDataKey: []byte(testDataValue)},
		},
		{
			name: "empty data sets builder error",
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), defaultSecretName, defaultSecretNamespace, defaultSecretType)
			},
			data:          map[string][]byte{},
			assertError:   func(err error) bool { return err != nil && err.Error() == "'data' cannot be empty" },
			expectDataNil: true,
		},
		{
			name: invalidBuilderShortCirc,
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), "", defaultSecretNamespace, defaultSecretType)
			},
			data:          map[string][]byte{testDataKey: []byte(testDataValue)},
			assertError:   commonerrors.IsBuilderNameEmpty,
			expectDataNil: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithData(testCase.data)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if testCase.expectDataNil {
				assert.Nil(t, result.Definition.Data)
			} else {
				assert.Equal(t, testCase.expectedData, result.Definition.Data)
			}
		})
	}
}

func TestWithStringData(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		builder            func() *Builder
		data               map[string]string
		assertError        func(error) bool
		expectedStringData map[string]string
		expectDataNil      bool
	}{
		{
			name: "sets stringData on valid builder",
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), defaultSecretName, defaultSecretNamespace, defaultSecretType)
			},
			data:               map[string]string{testDataKey: testDataValue},
			assertError:        func(err error) bool { return err == nil },
			expectedStringData: map[string]string{testDataKey: testDataValue},
		},
		{
			name: "empty stringData sets builder error",
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), defaultSecretName, defaultSecretNamespace, defaultSecretType)
			},
			data:          map[string]string{},
			assertError:   func(err error) bool { return err != nil && err.Error() == "'stringData' cannot be empty" },
			expectDataNil: true,
		},
		{
			name: invalidBuilderShortCirc,
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), "", defaultSecretNamespace, defaultSecretType)
			},
			data:          map[string]string{testDataKey: testDataValue},
			assertError:   commonerrors.IsBuilderNameEmpty,
			expectDataNil: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithStringData(testCase.data)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if testCase.expectDataNil {
				assert.Nil(t, result.Definition.StringData)
			} else {
				assert.Equal(t, testCase.expectedStringData, result.Definition.StringData)
			}
		})
	}
}

func TestWithAnnotations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		builder             func() *Builder
		annotations         map[string]string
		assertError         func(error) bool
		expectedAnnotations map[string]string
	}{
		{
			name: "sets annotations on valid builder",
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), defaultSecretName, defaultSecretNamespace, defaultSecretType)
			},
			annotations:         map[string]string{"openshift.io/internal-registry-auth-token.binding": "bound"},
			assertError:         func(err error) bool { return err == nil },
			expectedAnnotations: map[string]string{"openshift.io/internal-registry-auth-token.binding": "bound"},
		},
		{
			name: "empty annotations sets builder error",
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), defaultSecretName, defaultSecretNamespace, defaultSecretType)
			},
			annotations: map[string]string{},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "'annotations' argument cannot be empty"
			},
		},
		{
			name: "empty annotation key sets builder error",
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), defaultSecretName, defaultSecretNamespace, defaultSecretType)
			},
			annotations: map[string]string{"": testDataValue},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "can not apply an annotations with an empty key"
			},
		},
		{
			name: invalidBuilderShortCirc,
			builder: func() *Builder {
				return NewBuilder(newSecretTestClient(), "", defaultSecretNamespace, defaultSecretType)
			},
			annotations: map[string]string{testDataKey: testDataValue},
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithAnnotations(testCase.annotations)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if testCase.expectedAnnotations != nil {
				assert.Equal(t, testCase.expectedAnnotations, result.Definition.Annotations)
			}
		})
	}
}

func newSecretCommonTestConfig() testhelper.CommonTestConfig[corev1.Secret, Builder, *corev1.Secret, *Builder] {
	return testhelper.NewCommonTestConfig[corev1.Secret, Builder](
		corev1.AddToScheme, secretGVK, testhelper.ResourceScopeNamespaced)
}

func newSecretTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
}
