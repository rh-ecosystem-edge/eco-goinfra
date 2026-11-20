package metallb

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	defaultBFDProfileName   = "default-bfd"
	defaultBFDProfileNsName = "test-namespace"
)

func TestNewBFDBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewBFDBuilder,
		mlbtypes.AddToScheme,
		bfdProfileGVK,
	).ExecuteTests(t)
}

func TestPullBFDProfile(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullBFDProfile,
		mlbtypes.AddToScheme,
		bfdProfileGVK,
	).ExecuteTests(t)
}

func TestBFDProfileMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[mlbtypes.BFDProfile, BFDBuilder](
		mlbtypes.AddToScheme,
		bfdProfileGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestBFDProfileWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[mlbtypes.BFDProfile, BFDBuilder](
			mlbtypes.AddToScheme,
			bfdProfileGVK,
			testhelper.ResourceScopeNamespaced,
		),
	).ExecuteTests(t)
}

func TestBFDProfileWithRcvInterval(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		interval uint32
		valid    bool
	}{
		{name: "receive interval 10", interval: 10, valid: true},
		{name: "receive interval 20", interval: 20, valid: true},
		{name: "invalid builder", interval: 10, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildBFDProfileBuilderForTest(testCase.valid)
			builder = builder.WithRcvInterval(testCase.interval)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
				assert.Nil(t, builder.Definition.Spec.ReceiveInterval)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.ReceiveInterval)
			assert.Equal(t, testCase.interval, *builder.Definition.Spec.ReceiveInterval)
		})
	}
}

func TestBFDProfileWithTransmitInterval(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		interval uint32
		valid    bool
	}{
		{name: "transmit interval 10", interval: 10, valid: true},
		{name: "transmit interval 20", interval: 20, valid: true},
		{name: "invalid builder", interval: 10, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildBFDProfileBuilderForTest(testCase.valid)
			builder = builder.WithTransmitInterval(testCase.interval)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
				assert.Nil(t, builder.Definition.Spec.TransmitInterval)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.TransmitInterval)
			assert.Equal(t, testCase.interval, *builder.Definition.Spec.TransmitInterval)
		})
	}
}

func TestBFDProfileWithEchoInterval(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		interval uint32
		valid    bool
	}{
		{name: "echo interval 10", interval: 10, valid: true},
		{name: "echo interval 20", interval: 20, valid: true},
		{name: "invalid builder", interval: 10, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildBFDProfileBuilderForTest(testCase.valid)
			builder = builder.WithEchoInterval(testCase.interval)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
				assert.Nil(t, builder.Definition.Spec.EchoInterval)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.EchoInterval)
			assert.Equal(t, testCase.interval, *builder.Definition.Spec.EchoInterval)
		})
	}
}

func TestBFDProfileWithMultiplier(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		multiplier uint32
		valid      bool
	}{
		{name: "multiplier 10", multiplier: 10, valid: true},
		{name: "multiplier 20", multiplier: 20, valid: true},
		{name: "invalid builder", multiplier: 10, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildBFDProfileBuilderForTest(testCase.valid)
			builder = builder.WithMultiplier(testCase.multiplier)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
				assert.Nil(t, builder.Definition.Spec.DetectMultiplier)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.DetectMultiplier)
			assert.Equal(t, testCase.multiplier, *builder.Definition.Spec.DetectMultiplier)
		})
	}
}

func TestBFDProfileWithEchoMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		echoMode bool
		valid    bool
	}{
		{name: "echo mode true", echoMode: true, valid: true},
		{name: "echo mode false", echoMode: false, valid: true},
		{name: "invalid builder", echoMode: true, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildBFDProfileBuilderForTest(testCase.valid)
			builder = builder.WithEchoMode(testCase.echoMode)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
				assert.Nil(t, builder.Definition.Spec.EchoMode)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.EchoMode)
			assert.Equal(t, testCase.echoMode, *builder.Definition.Spec.EchoMode)
		})
	}
}

func TestBFDProfileWithPassiveMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		passiveMode bool
		valid       bool
	}{
		{name: "passive mode true", passiveMode: true, valid: true},
		{name: "passive mode false", passiveMode: false, valid: true},
		{name: "invalid builder", passiveMode: true, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildBFDProfileBuilderForTest(testCase.valid)
			builder = builder.WithPassiveMode(testCase.passiveMode)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
				assert.Nil(t, builder.Definition.Spec.PassiveMode)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.PassiveMode)
			assert.Equal(t, testCase.passiveMode, *builder.Definition.Spec.PassiveMode)
		})
	}
}

func TestBFDProfileWithMinimumTTL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		minimumTTL uint32
		valid      bool
	}{
		{name: "minimum TTL 10", minimumTTL: 10, valid: true},
		{name: "minimum TTL 20", minimumTTL: 20, valid: true},
		{name: "invalid builder", minimumTTL: 10, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildBFDProfileBuilderForTest(testCase.valid)
			builder = builder.WithMinimumTTL(testCase.minimumTTL)

			if !testCase.valid {
				require.Error(t, builder.GetError())
				assert.True(t, commonerrors.IsBuilderNamespaceEmpty(builder.GetError()))
				assert.Nil(t, builder.Definition.Spec.MinimumTTL)

				return
			}

			require.NoError(t, builder.GetError())
			require.NotNil(t, builder.Definition.Spec.MinimumTTL)
			assert.Equal(t, testCase.minimumTTL, *builder.Definition.Spec.MinimumTTL)
		})
	}
}

func TestBFDProfileGVR(t *testing.T) {
	t.Parallel()

	assert.Equal(t, GetBFDProfileGVR(),
		schema.GroupVersionResource{
			Group: APIGroup, Version: APIVersion, Resource: "bfdprofiles",
		})
}

// buildBFDProfileBuilderForTest returns a valid or invalid BFDProfile builder.
func buildBFDProfileBuilderForTest(valid bool) *BFDBuilder {
	apiClient := clients.GetTestClients(clients.TestClientParams{})

	if valid {
		return buildValidBFDProfileBuilder(apiClient)
	}

	return buildInvalidBFDProfileBuilder(apiClient)
}

// buildValidBFDProfileBuilder returns a valid BFDProfile builder for testing.
func buildValidBFDProfileBuilder(apiClient *clients.Settings) *BFDBuilder {
	return NewBFDBuilder(
		apiClient, defaultBFDProfileName, defaultBFDProfileNsName)
}

// buildInvalidBFDProfileBuilder returns an invalid BFDProfile builder for testing.
func buildInvalidBFDProfileBuilder(apiClient *clients.Settings) *BFDBuilder {
	return NewBFDBuilder(
		apiClient, defaultBFDProfileName, "")
}

// buildDummyBFDProfile returns a BFDProfile runtime object for shared package tests.
func buildDummyBFDProfile() []runtime.Object {
	return append([]runtime.Object{}, &mlbtypes.BFDProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultBFDProfileName,
			Namespace: defaultBFDProfileNsName,
		},
		Spec: mlbtypes.BFDProfileSpec{},
	})
}
