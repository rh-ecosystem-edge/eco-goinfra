package ptp

import (
	"fmt"
	"testing"

	ptpv1 "github.com/k8snetworkplumbingwg/ptp-operator/api/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPullPtpOperatorConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		addObject     bool
		nilClient     bool
		expectedError func(error) bool
	}{
		{
			name:      "existing singleton",
			addObject: true,
		},
		{
			name:          "missing singleton",
			expectedError: func(err error) bool { return err != nil },
		},
		{
			name:          "nil client",
			addObject:     true,
			nilClient:     true,
			expectedError: commonerrors.IsAPIClientNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var apiClient *clients.Settings

			if !testCase.nilClient {
				var objects []runtime.Object
				if testCase.addObject {
					objects = append(objects, buildDummyPtpOperatorConfig())
				}

				apiClient = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  objects,
					SchemeAttachers: []clients.SchemeAttacher{ptpv1.AddToScheme},
				})
			}

			builder, err := PullPtpOperatorConfig(apiClient)
			if testCase.expectedError != nil {
				require.Error(t, err)
				assert.True(t, testCase.expectedError(err))
				assert.Nil(t, builder)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, builder)
			assert.Equal(t, PtpOperatorConfigName, builder.Definition.Name)
			assert.Equal(t, PtpOperatorConfigNamespace, builder.Definition.Namespace)
		})
	}
}

func TestPtpOperatorConfigMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[ptpv1.PtpOperatorConfig, PtpOperatorConfigBuilder](
		ptpv1.AddToScheme,
		ptpv1.GroupVersion.WithKind("PtpOperatorConfig"),
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestPtpOperatorConfigWithEventConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		eventConfig    ptpv1.PtpEventConfig
		invalidBuilder bool
		expectedError  string
	}{
		{
			name: "valid version 1.0",
			eventConfig: ptpv1.PtpEventConfig{
				TransportHost: "http://example.com:8080",
				ApiVersion:    "1.0",
			},
		},
		{
			name: "valid version 2",
			eventConfig: ptpv1.PtpEventConfig{
				TransportHost: "http://example.com:8080",
				ApiVersion:    "2.0",
			},
		},
		{
			name:          "invalid transport host",
			eventConfig:   ptpv1.PtpEventConfig{TransportHost: "::::"},
			expectedError: "invalid TransportHost for PtpEventConfig: parse \"::::\": missing protocol scheme",
		},
		{
			name: "invalid API version",
			eventConfig: ptpv1.PtpEventConfig{
				TransportHost: "http://example.com:8080",
				ApiVersion:    "3.0",
			},
			expectedError: "invalid ApiVersion for PtpEventConfig: must be 1.0 or start with 2.",
		},
		{
			name:           "invalid builder",
			invalidBuilder: true,
			expectedError:  "test error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidPtpOperatorConfigBuilder(buildTestClientWithDummyPtpOperatorConfig())
			if testCase.invalidBuilder {
				testBuilder.SetError(fmt.Errorf("test error"))
			}

			testBuilder.WithEventConfig(testCase.eventConfig)

			if testCase.expectedError != "" {
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
				assert.Nil(t, testBuilder.Definition.Spec.EventConfig)

				return
			}

			require.NoError(t, testBuilder.GetError())
			require.NotNil(t, testBuilder.Definition.Spec.EventConfig)
			assert.Equal(t, testCase.eventConfig, *testBuilder.Definition.Spec.EventConfig)
		})
	}
}

func buildDummyPtpOperatorConfig() *ptpv1.PtpOperatorConfig {
	return &ptpv1.PtpOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PtpOperatorConfigName,
			Namespace: PtpOperatorConfigNamespace,
		},
		Spec: ptpv1.PtpOperatorConfigSpec{
			DaemonNodeSelector: make(map[string]string),
		},
	}
}

func buildTestClientWithDummyPtpOperatorConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyPtpOperatorConfig()},
		SchemeAttachers: []clients.SchemeAttacher{ptpv1.AddToScheme},
	})
}

func buildValidPtpOperatorConfigBuilder(apiClient *clients.Settings) *PtpOperatorConfigBuilder {
	return common.NewNamespacedBuilder[ptpv1.PtpOperatorConfig, PtpOperatorConfigBuilder](
		apiClient, ptpv1.AddToScheme, PtpOperatorConfigName, PtpOperatorConfigNamespace)
}
