package pfstatus

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/pfstatus/pfstatustypes"
	"github.com/stretchr/testify/assert"
)

const (
	defaultPfStatusConfigurationName   = "pfstatusconfiguration"
	defaultPfStatusConfigurationNsName = "test-namespace"
)

var pfStatusGVK = pfstatustypes.GroupVersion.WithKind("PFLACPMonitor")

func TestNewPfStatusConfigurationBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[pfstatustypes.PFLACPMonitor, PfStatusConfigurationBuilder](
		NewPfStatusConfigurationBuilder, pfstatustypes.AddToScheme, pfStatusGVK).ExecuteTests(t)
}

func TestPullPfStatusConfiguration(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[pfstatustypes.PFLACPMonitor, PfStatusConfigurationBuilder](
		PullPfStatusConfiguration, pfstatustypes.AddToScheme, pfStatusGVK).ExecuteTests(t)
}

func TestPfStatusConfigurationBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[pfstatustypes.PFLACPMonitor, PfStatusConfigurationBuilder](
		pfstatustypes.AddToScheme,
		pfStatusGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestPfStatusConfigurationWithNodeSelector(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		nodeSelector  map[string]string
		expectedError string
	}{
		{
			name:         "valid node selector",
			valid:        true,
			nodeSelector: map[string]string{"test": "test1"},
		},
		{
			name:          "empty node selector",
			valid:         true,
			nodeSelector:  map[string]string{},
			expectedError: "pfStatusConfiguration 'nodeSelector' cannot be empty map",
		},
		{
			name:         "invalid builder",
			valid:        false,
			nodeSelector: map[string]string{"test": "test1"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildPfStatusConfigurationTestBuilder(testCase.valid)
			testBuilder = testBuilder.WithNodeSelector(testCase.nodeSelector)

			switch {
			case !testCase.valid:
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
			case testCase.expectedError != "":
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
			default:
				assert.NoError(t, testBuilder.GetError())
				assert.Equal(t, testCase.nodeSelector, testBuilder.Definition.Spec.NodeSelector)
			}
		})
	}
}

func TestPfStatusConfigurationWithInterface(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		valid         bool
		interfaceName string
		expectedError string
	}{
		{
			name:          "valid interface",
			valid:         true,
			interfaceName: "ens3f0np0",
		},
		{
			name:          "empty interface",
			valid:         true,
			interfaceName: "",
			expectedError: "interface can not be empty string",
		},
		{
			name:          "invalid builder",
			valid:         false,
			interfaceName: "ens3f0np0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildPfStatusConfigurationTestBuilder(testCase.valid)
			testBuilder = testBuilder.WithInterface(testCase.interfaceName)

			switch {
			case !testCase.valid:
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
			case testCase.expectedError != "":
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
			default:
				assert.NoError(t, testBuilder.GetError())
				assert.Equal(t, []string{testCase.interfaceName}, testBuilder.Definition.Spec.Interfaces)
			}
		})
	}
}

func TestPfStatusConfigurationWithPollingInterval(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		valid           bool
		pollingInterval int
		expectedError   string
	}{
		{
			name:            "valid polling interval",
			valid:           true,
			pollingInterval: 1000,
		},
		{
			name:            "polling interval below minimum",
			valid:           true,
			pollingInterval: 10,
			expectedError:   pfstatusconfigurationPollingintervalValueIsNotValid,
		},
		{
			name:            "polling interval above maximum",
			valid:           true,
			pollingInterval: 65555,
			expectedError:   pfstatusconfigurationPollingintervalValueIsNotValid,
		},
		{
			name:            "invalid builder",
			valid:           false,
			pollingInterval: 1000,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildPfStatusConfigurationTestBuilder(testCase.valid)
			testBuilder = testBuilder.WithPollingInterval(testCase.pollingInterval)

			switch {
			case !testCase.valid:
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(testBuilder.GetError()))
			case testCase.expectedError != "":
				assert.EqualError(t, testBuilder.GetError(), testCase.expectedError)
			default:
				assert.NoError(t, testBuilder.GetError())
				assert.Equal(t, testCase.pollingInterval, testBuilder.Definition.Spec.PollingInterval)
			}
		})
	}
}

func TestGetPfStatusConfigurationGVR(t *testing.T) {
	t.Parallel()

	gvr := GetPfStatusConfigurationGVR()

	assert.Equal(t, "pfstatusrelay.openshift.io", gvr.Group)
	assert.Equal(t, "v1alpha1", gvr.Version)
	assert.Equal(t, "pflacpmonitors", gvr.Resource)
}

func buildPfStatusConfigurationTestBuilder(valid bool) *PfStatusConfigurationBuilder {
	if valid {
		return NewPfStatusConfigurationBuilder(
			getTestPfStatusAPIClient(), defaultPfStatusConfigurationName, defaultPfStatusConfigurationNsName)
	}

	return NewPfStatusConfigurationBuilder(getTestPfStatusAPIClient(), defaultPfStatusConfigurationName, "")
}

func getTestPfStatusAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}
