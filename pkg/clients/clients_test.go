package clients

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	noopSchemeAttacher = func(s *runtime.Scheme) error { return nil }
	errSchemeAttacher  = func(s *runtime.Scheme) error { return fmt.Errorf("scheme error") }
)

func TestNew(t *testing.T) {
	result := New("/nonexistent/kubeconfig.yaml")
	assert.Nil(t, result)
}

func TestSetScheme(t *testing.T) {
	s := runtime.NewScheme()
	err := SetScheme(s)
	assert.Nil(t, err)
}

func TestGetAPIClient(t *testing.T) {
	testCases := []struct {
		name          string
		nilClient     bool
		expectedError error
	}{
		{
			name:          "nil settings",
			nilClient:     true,
			expectedError: fmt.Errorf("APIClient cannot be nil"),
		},
		{
			name: "valid settings",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var settings *Settings
			if !testCase.nilClient {
				settings = GetTestClients(TestClientParams{})
			}

			result, err := settings.GetAPIClient()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, settings, result)
			}
		})
	}
}

func TestAttachScheme(t *testing.T) {
	testCases := []struct {
		name          string
		nilClient     bool
		attacher      SchemeAttacher
		expectedError error
	}{
		{
			name:          "nil settings",
			nilClient:     true,
			attacher:      noopSchemeAttacher,
			expectedError: fmt.Errorf("cannot add scheme to nil client"),
		},
		{
			name:     "valid attacher",
			attacher: noopSchemeAttacher,
		},
		{
			name:          "attacher returns error",
			attacher:      errSchemeAttacher,
			expectedError: fmt.Errorf("scheme error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var settings *Settings
			if !testCase.nilClient {
				settings = GetTestClients(TestClientParams{})
			}

			err := settings.AttachScheme(testCase.attacher)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestGetTestClients(t *testing.T) {
	testCases := []struct {
		name   string
		params TestClientParams
	}{
		{
			name:   "empty params",
			params: TestClientParams{},
		},
		{
			name: "with scheme attacher",
			params: TestClientParams{
				SchemeAttachers: []SchemeAttacher{noopSchemeAttacher},
			},
		},
		{
			name: "with multiple scheme attachers",
			params: TestClientParams{
				SchemeAttachers: []SchemeAttacher{noopSchemeAttacher, noopSchemeAttacher},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := GetTestClients(testCase.params)
			assert.NotNil(t, result)
			assert.NotNil(t, result.Client)
			assert.NotNil(t, result.K8sClient)
		})
	}
}
