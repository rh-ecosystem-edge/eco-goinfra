package bmh

import (
	"context"
	"fmt"
	"testing"
	"time"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultBmHostName       = "metallbio"
	defaultBmHostNsName     = "test-namespace"
	defaultBmHostAddress    = "1.1.1.1"
	defaultBmHostSecretName = "testsecret"
	defaultBmHostMacAddress = "AA:BB:CC:11:22:33"
	defaultBmHostBootMode   = bootModeUEFISecureBoot
	defaultBmHostAnnotation = "annotation.openshift.io/test-annotation"
	testSchemes             = []clients.SchemeAttacher{
		bmhv1alpha1.AddToScheme,
	}
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *BmhBuilder {
				return NewBuilder(
					apiClient, name, nsname, defaultBmHostAddress, defaultBmHostSecretName,
					defaultBmHostMacAddress, defaultBmHostBootMode)
			},
			bmhv1alpha1.AddToScheme,
			bareMetalHostGVK,
		).ExecuteTests(t)
	})

	t.Run("sets default baremetalhost spec", func(t *testing.T) {
		t.Parallel()

		builder := buildValidBmHostBuilder(getBmhTestClient())

		require.NoError(t, builder.GetError())
		assert.Equal(t, defaultBmHostAddress, builder.Definition.Spec.BMC.Address)
		assert.Equal(t, defaultBmHostSecretName, builder.Definition.Spec.BMC.CredentialsName)
		assert.True(t, builder.Definition.Spec.BMC.DisableCertificateVerification)
		assert.Equal(t, bmhv1alpha1.BootMode(defaultBmHostBootMode), builder.Definition.Spec.BootMode)
		assert.Equal(t, defaultBmHostMacAddress, builder.Definition.Spec.BootMACAddress)
		assert.True(t, builder.Definition.Spec.Online)
		assert.False(t, builder.Definition.Spec.ExternallyProvisioned)
	})
}

func TestNewBuilderRequiredFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		bmcAddress       string
		bmcSecretName    string
		bootMacAddress   string
		bootMode         string
		expectedError    string
		expectedBootMode bmhv1alpha1.BootMode
	}{
		{
			name:             "UEFI boot mode",
			bmcAddress:       defaultBmHostAddress,
			bmcSecretName:    defaultBmHostSecretName,
			bootMacAddress:   defaultBmHostMacAddress,
			bootMode:         "UEFI",
			expectedBootMode: bmhv1alpha1.BootMode("UEFI"),
		},
		{
			name:             "legacy boot mode",
			bmcAddress:       defaultBmHostAddress,
			bmcSecretName:    defaultBmHostSecretName,
			bootMacAddress:   defaultBmHostMacAddress,
			bootMode:         "legacy",
			expectedBootMode: bmhv1alpha1.BootMode("legacy"),
		},
		{
			name:           "empty BMC address",
			bmcSecretName:  defaultBmHostSecretName,
			bootMacAddress: defaultBmHostMacAddress,
			bootMode:       defaultBmHostBootMode,
			expectedError:  "BMH 'bmcAddress' cannot be empty",
		},
		{
			name:           "empty BMC secret name",
			bmcAddress:     defaultBmHostAddress,
			bootMacAddress: defaultBmHostMacAddress,
			bootMode:       defaultBmHostBootMode,
			expectedError:  "BMH 'bmcSecretName' cannot be empty",
		},
		{
			name:           "invalid boot mode",
			bmcAddress:     defaultBmHostAddress,
			bmcSecretName:  defaultBmHostSecretName,
			bootMacAddress: defaultBmHostMacAddress,
			bootMode:       "invalid",
			expectedError:  errInvalidBootMode,
		},
		{
			name:          "empty boot MAC address",
			bmcAddress:    defaultBmHostAddress,
			bmcSecretName: defaultBmHostSecretName,
			bootMode:      defaultBmHostBootMode,
			expectedError: "BMH 'bootMacAddress' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewBuilder(
				getBmhTestClient(), defaultBmHostName, defaultBmHostNsName, testCase.bmcAddress,
				testCase.bmcSecretName, testCase.bootMacAddress, testCase.bootMode)

			require.NotNil(t, builder)

			if testCase.expectedError != "" {
				assert.EqualError(t, builder.GetError(), testCase.expectedError)

				return
			}

			require.NoError(t, builder.GetError())
			assert.Equal(t, testCase.expectedBootMode, builder.Definition.Spec.BootMode)
		})
	}
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull, bmhv1alpha1.AddToScheme, bareMetalHostGVK).ExecuteTests(t)
}

func TestBmhBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[bmhv1alpha1.BareMetalHost, BmhBuilder](
		bmhv1alpha1.AddToScheme,
		bareMetalHostGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		Run(t)
}

func TestBmhWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[bmhv1alpha1.BareMetalHost, BmhBuilder](
			bmhv1alpha1.AddToScheme,
			bareMetalHostGVK,
			testhelper.ResourceScopeNamespaced,
		)).ExecuteTests(t)
}

type rootDeviceModifierTestCase struct {
	name          string
	apply         func(builder *BmhBuilder) *BmhBuilder
	applyInvalid  func(builder *BmhBuilder) *BmhBuilder
	expectedError string
	expectedField func(hints *bmhv1alpha1.RootDeviceHints) any
	expectedValue any
}

//nolint:funlen // Keeping the table together makes modifier coverage easy to compare.
func bmhRootDeviceModifierTestCases() []rootDeviceModifierTestCase {
	return []rootDeviceModifierTestCase{
		{
			name: "device name",
			apply: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceDeviceName("device-name")
			},
			applyInvalid: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceDeviceName("")
			},
			expectedError: "the baremetalhost rootDeviceHint deviceName cannot be empty",
			expectedField: func(hints *bmhv1alpha1.RootDeviceHints) any { return hints.DeviceName },
			expectedValue: "device-name",
		},
		{
			name: "HCTL",
			apply: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceHTCL("0:0:0:0")
			},
			applyInvalid: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceHTCL("")
			},
			expectedError: "the baremetalhost rootDeviceHint hctl cannot be empty",
			expectedField: func(hints *bmhv1alpha1.RootDeviceHints) any { return hints.HCTL },
			expectedValue: "0:0:0:0",
		},
		{
			name: "model",
			apply: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceModel("model")
			},
			applyInvalid: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceModel("")
			},
			expectedError: "the baremetalhost rootDeviceHint model cannot be empty",
			expectedField: func(hints *bmhv1alpha1.RootDeviceHints) any { return hints.Model },
			expectedValue: "model",
		},
		{
			name: "vendor",
			apply: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceVendor("vendor")
			},
			applyInvalid: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceVendor("")
			},
			expectedError: "the baremetalhost rootDeviceHint vendor cannot be empty",
			expectedField: func(hints *bmhv1alpha1.RootDeviceHints) any { return hints.Model },
			expectedValue: "vendor",
		},
		{
			name: "serial number",
			apply: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceSerialNumber("serial-number")
			},
			applyInvalid: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceSerialNumber("")
			},
			expectedError: "the baremetalhost rootDeviceHint serialNumber cannot be empty",
			expectedField: func(hints *bmhv1alpha1.RootDeviceHints) any { return hints.SerialNumber },
			expectedValue: "serial-number",
		},
		{
			name: "minimum size",
			apply: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceMinSizeGigabytes(12)
			},
			applyInvalid: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceMinSizeGigabytes(-1)
			},
			expectedError: "the baremetalhost rootDeviceHint size cannot be less than 0",
			expectedField: func(hints *bmhv1alpha1.RootDeviceHints) any { return hints.MinSizeGigabytes },
			expectedValue: 12,
		},
		{
			name: "WWN",
			apply: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceWWN("wwn")
			},
			applyInvalid: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceWWN("")
			},
			expectedError: "the baremetalhost rootDeviceHint wwn cannot be empty",
			expectedField: func(hints *bmhv1alpha1.RootDeviceHints) any { return hints.WWN },
			expectedValue: "wwn",
		},
		{
			name: "WWN with extension",
			apply: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceWWNWithExtension("wwn-with-extension")
			},
			applyInvalid: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceWWNWithExtension("")
			},
			expectedError: "the baremetalhost rootDeviceHint wwnWithExtension cannot be empty",
			expectedField: func(hints *bmhv1alpha1.RootDeviceHints) any { return hints.WWNWithExtension },
			expectedValue: "wwn-with-extension",
		},
		{
			name: "WWN vendor extension",
			apply: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceWWNVendorExtension("wwn-vendor-extension")
			},
			applyInvalid: func(builder *BmhBuilder) *BmhBuilder {
				return builder.WithRootDeviceWWNVendorExtension("")
			},
			expectedError: "the baremetalhost rootDeviceHint wwnVendorExtension cannot be empty",
			expectedField: func(hints *bmhv1alpha1.RootDeviceHints) any { return hints.WWNVendorExtension },
			expectedValue: "wwn-vendor-extension",
		},
	}
}

func TestBmhRootDeviceModifiers(t *testing.T) {
	t.Parallel()

	for _, testCase := range bmhRootDeviceModifierTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			validBuilder := testCase.apply(buildValidBmHostBuilder(getBmhTestClient()))
			require.NoError(t, validBuilder.GetError())
			require.NotNil(t, validBuilder.Definition.Spec.RootDeviceHints)
			assert.Equal(t, testCase.expectedValue, testCase.expectedField(validBuilder.Definition.Spec.RootDeviceHints))

			invalidValueBuilder := testCase.applyInvalid(buildValidBmHostBuilder(getBmhTestClient()))
			assert.EqualError(t, invalidValueBuilder.GetError(), testCase.expectedError)

			invalidBuilder := testCase.apply(buildInvalidBmHostBuilder(getBmhTestClient()))
			assert.EqualError(t, invalidBuilder.GetError(), errInvalidBootMode)
		})
	}
}

func TestBmhWithRootDeviceRotationalDisk(t *testing.T) {
	t.Parallel()

	t.Run("sets rotational disk", func(t *testing.T) {
		t.Parallel()

		validBuilder := buildValidBmHostBuilder(getBmhTestClient()).WithRootDeviceRotationalDisk(false)
		require.NoError(t, validBuilder.GetError())
		require.NotNil(t, validBuilder.Definition.Spec.RootDeviceHints)
		require.NotNil(t, validBuilder.Definition.Spec.RootDeviceHints.Rotational)
		assert.False(t, *validBuilder.Definition.Spec.RootDeviceHints.Rotational)
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		invalidBuilder := buildInvalidBmHostBuilder(getBmhTestClient()).WithRootDeviceRotationalDisk(true)
		assert.EqualError(t, invalidBuilder.GetError(), errInvalidBootMode)
		assert.Nil(t, invalidBuilder.Definition.Spec.RootDeviceHints)
	})
}

func TestBmhStatusMethods(t *testing.T) {
	t.Parallel()

	dummyBmh := buildDummyBmHost(bmhv1alpha1.StateProvisioned, bmhv1alpha1.OperationalStatusDelayed)
	dummyBmh.Status.PoweredOn = false

	testCases := []struct {
		name                     string
		builder                  *BmhBuilder
		expectedOperationalState bmhv1alpha1.OperationalStatus
		expectedPoweredOn        bool
	}{
		{
			name:                     "existing baremetalhost",
			builder:                  buildValidBmHostBuilder(buildBmhTestClientWithObjects(dummyBmh)),
			expectedOperationalState: bmhv1alpha1.OperationalStatusDelayed,
			expectedPoweredOn:        false,
		},
		{
			name:    "missing baremetalhost",
			builder: buildValidBmHostBuilder(getBmhTestClient()),
		},
		{
			name:    "invalid builder",
			builder: buildInvalidBmHostBuilder(getBmhTestClient()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expectedOperationalState, testCase.builder.GetBmhOperationalState())
			assert.Equal(t, testCase.expectedPoweredOn, testCase.builder.GetBmhPowerOnStatus())
		})
	}
}

type bmhWaitMethod struct {
	name  string
	wait  func(builder *BmhBuilder, timeout time.Duration) error
	state bmhv1alpha1.ProvisioningState
}

func bmhWaitUntilStatusMethods() []bmhWaitMethod {
	return []bmhWaitMethod{
		{
			name: "provisioned",
			wait: func(builder *BmhBuilder, timeout time.Duration) error {
				return builder.WaitUntilProvisioned(timeout)
			},
			state: bmhv1alpha1.StateProvisioned,
		},
		{
			name: "provisioning",
			wait: func(builder *BmhBuilder, timeout time.Duration) error {
				return builder.WaitUntilProvisioning(timeout)
			},
			state: bmhv1alpha1.StateProvisioning,
		},
		{
			name: "ready",
			wait: func(builder *BmhBuilder, timeout time.Duration) error {
				return builder.WaitUntilReady(timeout)
			},
			state: bmhv1alpha1.StateReady,
		},
		{
			name: "available",
			wait: func(builder *BmhBuilder, timeout time.Duration) error {
				return builder.WaitUntilAvailable(timeout)
			},
			state: bmhv1alpha1.StateAvailable,
		},
		{
			name: "specific status",
			wait: func(builder *BmhBuilder, timeout time.Duration) error {
				return builder.WaitUntilInStatus(bmhv1alpha1.StateAvailable, timeout)
			},
			state: bmhv1alpha1.StateAvailable,
		},
	}
}

func TestBmhWaitUntilStatus(t *testing.T) {
	t.Parallel()

	for _, waitMethod := range bmhWaitUntilStatusMethods() {
		t.Run(waitMethod.name, func(t *testing.T) {
			t.Parallel()

			testCases := []struct {
				name          string
				builder       *BmhBuilder
				expectedError error
			}{
				{
					name:    "reaches status",
					builder: buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(waitMethod.state)),
				},
				{
					name:          "wrong status times out",
					builder:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateNone)),
					expectedError: context.DeadlineExceeded,
				},
				{
					name:          "missing resource times out",
					builder:       buildValidBmHostBuilder(getBmhTestClient()),
					expectedError: context.DeadlineExceeded,
				},
			}

			for _, testCase := range testCases {
				t.Run(testCase.name, func(t *testing.T) {
					t.Parallel()

					err := waitMethod.wait(testCase.builder, 10*time.Millisecond)
					if testCase.expectedError == nil {
						assert.NoError(t, err)

						return
					}

					assert.ErrorIs(t, err, testCase.expectedError)
				})
			}

			t.Run("invalid builder", func(t *testing.T) {
				t.Parallel()

				err := waitMethod.wait(buildInvalidBmHostBuilder(getBmhTestClient()), time.Second)
				assert.ErrorContains(t, err, errInvalidBootMode)
			})
		})
	}
}

func TestBmhCreateAndWaitUntilProvisioned(t *testing.T) {
	t.Parallel()

	t.Run("existing provisioned baremetalhost", func(t *testing.T) {
		t.Parallel()

		builder, err := buildValidBmHostBuilder(
			buildBareMetalHostTestClientWithDummyObject()).CreateAndWaitUntilProvisioned(10 * time.Millisecond)
		require.NoError(t, err)
		require.NotNil(t, builder)
		require.NotNil(t, builder.Object)
		assert.Equal(t, defaultBmHostName, builder.Object.Name)
		assert.Equal(t, defaultBmHostNsName, builder.Object.Namespace)
	})

	t.Run("new baremetalhost does not reach provisioned", func(t *testing.T) {
		t.Parallel()

		_, err := buildValidBmHostBuilder(getBmhTestClient()).CreateAndWaitUntilProvisioned(10 * time.Millisecond)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		_, err := buildInvalidBmHostBuilder(getBmhTestClient()).CreateAndWaitUntilProvisioned(time.Second)
		assert.ErrorContains(t, err, errInvalidBootMode)
	})
}

func TestBmhDeleteAndWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     *BmhBuilder
		expectError bool
	}{
		{
			name:    "existing baremetalhost",
			builder: buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
		},
		{
			name:    "missing baremetalhost",
			builder: buildValidBmHostBuilder(getBmhTestClient()),
		},
		{
			name:        "invalid builder",
			builder:     buildInvalidBmHostBuilder(getBmhTestClient()),
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder, err := testCase.builder.DeleteAndWaitUntilDeleted(2 * time.Second)
			if testCase.expectError {
				assert.ErrorContains(t, err, errInvalidBootMode)

				return
			}

			assert.NoError(t, err)
			assert.Nil(t, builder)
			assert.Nil(t, testCase.builder.Object)
		})
	}
}

func TestBmhWaitUntilDeleted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       *BmhBuilder
		expectedError error
	}{
		{
			name:    "missing baremetalhost",
			builder: buildValidBmHostBuilder(getBmhTestClient()),
		},
		{
			name:          "existing baremetalhost times out",
			builder:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.builder.WaitUntilDeleted(1500 * time.Millisecond)
			if testCase.expectedError == nil {
				assert.NoError(t, err)

				return
			}

			assert.ErrorIs(t, err, testCase.expectedError)
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		err := buildInvalidBmHostBuilder(getBmhTestClient()).WaitUntilDeleted(time.Second)
		assert.ErrorContains(t, err, errInvalidBootMode)
	})
}

func TestBmhWaitUntilAnnotationExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		annotation    string
		exists        bool
		annotate      bool
		valid         bool
		deadline      bool
		expectedError error
	}{
		{
			name:       "annotation exists",
			annotation: defaultBmHostAnnotation,
			exists:     true,
			annotate:   true,
			valid:      true,
		},
		{
			name:          "empty annotation",
			exists:        true,
			annotate:      true,
			valid:         true,
			expectedError: fmt.Errorf("bmh annotation key cannot be empty"),
		},
		{
			name:       "baremetalhost missing",
			annotation: defaultBmHostAnnotation,
			valid:      true,
			expectedError: fmt.Errorf(
				"baremetalhost object %s does not exist in namespace %s", defaultBmHostName, defaultBmHostNsName),
		},
		{
			name:       "annotation times out",
			annotation: defaultBmHostAnnotation,
			exists:     true,
			valid:      true,
			deadline:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object

			if testCase.exists {
				dummyBmh := buildDummyBmHost(bmhv1alpha1.StateProvisioned, bmhv1alpha1.OperationalStatusOK)

				if testCase.annotate {
					dummyBmh.Annotations = map[string]string{defaultBmHostAnnotation: ""}
				}

				objects = append(objects, dummyBmh)
			}

			testSettings := buildBmhTestClientWithObjects(objects...)

			var builder *BmhBuilder
			if testCase.valid {
				builder = buildValidBmHostBuilder(testSettings)
			} else {
				builder = buildInvalidBmHostBuilder(testSettings)
			}

			_, err := builder.WaitUntilAnnotationExists(testCase.annotation, time.Second)
			if testCase.deadline {
				assert.ErrorIs(t, err, context.DeadlineExceeded)

				return
			}

			if testCase.expectedError == nil {
				assert.NoError(t, err)

				return
			}

			assert.Equal(t, testCase.expectedError, err)
		})
	}

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		_, err := buildInvalidBmHostBuilder(getBmhTestClient()).
			WaitUntilAnnotationExists(defaultBmHostAnnotation, time.Second)
		assert.ErrorContains(t, err, errInvalidBootMode)
	})
}

func buildValidBmHostBuilder(apiClient *clients.Settings) *BmhBuilder {
	return NewBuilder(
		apiClient,
		defaultBmHostName,
		defaultBmHostNsName,
		defaultBmHostAddress,
		defaultBmHostSecretName,
		defaultBmHostMacAddress,
		defaultBmHostBootMode,
	)
}

func buildInvalidBmHostBuilder(apiClient *clients.Settings) *BmhBuilder {
	return NewBuilder(
		apiClient,
		defaultBmHostName,
		defaultBmHostNsName,
		defaultBmHostAddress,
		defaultBmHostSecretName,
		defaultBmHostMacAddress,
		"invalid",
	)
}

func getBmhTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

func buildBmhTestClientWithObjects(objects ...runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: testSchemes,
	})
}

func buildBareMetalHostTestClientWithDummyObject(state ...bmhv1alpha1.ProvisioningState) *clients.Settings {
	provisionState := bmhv1alpha1.StateProvisioned
	if len(state) > 0 {
		provisionState = state[0]
	}

	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyBmHostObject(provisionState),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyBmHost(
	state bmhv1alpha1.ProvisioningState, operState bmhv1alpha1.OperationalStatus) *bmhv1alpha1.BareMetalHost {
	return &bmhv1alpha1.BareMetalHost{
		Spec: bmhv1alpha1.BareMetalHostSpec{
			BMC: bmhv1alpha1.BMCDetails{
				Address:                        defaultBmHostAddress,
				CredentialsName:                defaultBmHostSecretName,
				DisableCertificateVerification: true,
			},
			BootMode:              bmhv1alpha1.BootMode(defaultBmHostBootMode),
			BootMACAddress:        defaultBmHostMacAddress,
			Online:                true,
			ExternallyProvisioned: false,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultBmHostName,
			Namespace: defaultBmHostNsName,
		},
		Status: bmhv1alpha1.BareMetalHostStatus{
			OperationalStatus: operState,
			PoweredOn:         true,
			Provisioning: bmhv1alpha1.ProvisionStatus{
				State: state,
			},
		},
	}
}

func buildDummyBmHostObject(
	state bmhv1alpha1.ProvisioningState, operationalStatus ...bmhv1alpha1.OperationalStatus) []runtime.Object {
	operState := bmhv1alpha1.OperationalStatusOK
	if len(operationalStatus) > 0 {
		operState = operationalStatus[0]
	}

	return append([]runtime.Object{}, buildDummyBmHost(state, operState))
}
