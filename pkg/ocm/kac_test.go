package ocm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ocm/kacv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultKACName      = "klusterletaddonconfig-test"
	defaultKACNamespace = "test-ns"
)

var kacGVK = kacv1.SchemeGroupVersion.WithKind("KlusterletAddonConfig")

func TestNewKACBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(NewKACBuilder, kacv1.SchemeBuilder.AddToScheme, kacGVK).ExecuteTests(t)
}

func TestPullKAC(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(PullKAC, kacv1.SchemeBuilder.AddToScheme, kacGVK).ExecuteTests(t)
}

func TestKACBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[kacv1.KlusterletAddonConfig, KACBuilder](
		kacv1.SchemeBuilder.AddToScheme,
		kacGVK,
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

func TestKACWaitUntilSearchCollectorEnabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		exists        bool
		valid         bool
		enabled       bool
		expectedError error
	}{
		{
			name:    "enabled collector",
			exists:  true,
			valid:   true,
			enabled: true,
		},
		{
			name:    "not exists",
			exists:  false,
			valid:   true,
			enabled: true,
			expectedError: fmt.Errorf(
				"klusterletAddonConfig object %s does not exist in namespace %s",
				defaultKACName, defaultKACNamespace),
		},
		{
			name:    "invalid builder",
			exists:  true,
			valid:   false,
			enabled: true,
		},
		{
			name:          "timeout waiting for enabled",
			exists:        true,
			valid:         true,
			enabled:       false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				kac := buildDummyKAC(defaultKACName, defaultKACNamespace)

				if testCase.enabled {
					kac.Spec.SearchCollectorConfig.Enabled = true
				}

				runtimeObjects = append(runtimeObjects, kac)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{
					kacv1.SchemeBuilder.AddToScheme,
				},
			})

			var kacBuilder *KACBuilder
			if testCase.valid {
				kacBuilder = buildValidKACTestBuilder(testSettings)
			} else {
				kacBuilder = buildInvalidKACTestBuilder(testSettings)
			}

			_, err := kacBuilder.WaitUntilSearchCollectorEnabled(time.Second)

			switch {
			case testCase.name == "invalid builder":
				require.Error(t, err)
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(err))
			case testCase.expectedError != nil:
				assert.Equal(t, testCase.expectedError, err)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

// buildDummyKAC returns a KlusterletAddonConfig with the provided name and namespace.
func buildDummyKAC(name, namespace string) *kacv1.KlusterletAddonConfig {
	return &kacv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// buildValidKACTestBuilder returns a valid Builder for testing.
func buildValidKACTestBuilder(apiClient *clients.Settings) *KACBuilder {
	return NewKACBuilder(apiClient, defaultKACName, defaultKACNamespace)
}

// buildInvalidKACTestBuilder returns an invalid Builder for testing.
func buildInvalidKACTestBuilder(apiClient *clients.Settings) *KACBuilder {
	return NewKACBuilder(apiClient, defaultKACName, "")
}
