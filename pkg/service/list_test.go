package service

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestList(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedListTestConfig(
		func(apiClient *clients.Settings, nsname string, _ ...runtimeclient.ListOptions) ([]*Builder, error) {
			return List(apiClient, nsname)
		},
		corev1.AddToScheme,
		serviceGVK,
	).ExecuteTests(t)
}

func TestListWithOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		options       []metav1.ListOptions
		assertError   func(error) bool
		expectedNames []string
	}{
		{
			name:          "label selector filters services",
			options:       []metav1.ListOptions{{LabelSelector: "demo=first"}},
			assertError:   func(err error) bool { return err == nil },
			expectedNames: []string{"first"},
		},
		{
			name:          "non-matching label selector returns no services",
			options:       []metav1.ListOptions{{LabelSelector: "demo=missing"}},
			assertError:   func(err error) bool { return err == nil },
			expectedNames: nil,
		},
		{
			name:        "invalid label selector returns error",
			options:     []metav1.ListOptions{{LabelSelector: "demo in (("}},
			assertError: func(err error) bool { return err != nil },
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			apiClient := newServiceTestClient(
				buildLabeledService("first", defaultServiceNamespace),
				buildLabeledService("second", defaultServiceNamespace),
				buildLabeledService("first", "other-namespace"),
			)

			builders, err := List(apiClient, defaultServiceNamespace, testCase.options...)
			require.Truef(t, testCase.assertError(err), "unexpected error: %v", err)

			var names []string

			for _, builder := range builders {
				assert.Equal(t, defaultServiceNamespace, builder.Definition.Namespace)

				names = append(names, builder.Definition.Name)
			}

			assert.Equal(t, testCase.expectedNames, names)
		})
	}
}

func buildLabeledService(name, nsname string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
			Labels:    map[string]string{"demo": name},
		},
	}
}
