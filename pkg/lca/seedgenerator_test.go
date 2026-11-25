package lca

import (
	"context"
	"fmt"
	"testing"
	"time"

	lcasgv1 "github.com/openshift-kni/lifecycle-agent/api/seedgenerator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var seedGeneratorGVK = lcasgv1.GroupVersion.WithKind("SeedGenerator")

func TestNewSeedGeneratorBuilder(t *testing.T) {
	t.Parallel()

	t.Run("valid seed image name", func(t *testing.T) {
		t.Parallel()

		builder := buildValidSeedGeneratorBuilder(newSeedGeneratorTestClient(nil))

		require.NoError(t, builder.GetError())
		assert.Equal(t, seedImageName, builder.Definition.Name)
		assert.Equal(t, seedGeneratorGVK, builder.GetGVK())
	})

	t.Run("nil client returns error-bearing builder", func(t *testing.T) {
		t.Parallel()

		builder := NewSeedGeneratorBuilder(nil, seedImageName)

		require.NotNil(t, builder)
		assert.True(t, commonerrors.IsAPIClientNil(builder.GetError()))
	})

	testCases := []struct {
		name        string
		expectedErr string
	}{
		{
			name:        "",
			expectedErr: "SeedGenerator name must be seedimage",
		},
		{
			name:        "seedgen2",
			expectedErr: "SeedGenerator name must be seedimage",
		},
	}

	for _, testCase := range testCases {
		t.Run("invalid name "+testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewSeedGeneratorBuilder(newSeedGeneratorTestClient(nil), testCase.name)

			require.NotNil(t, builder)
			assert.EqualError(t, builder.GetError(), testCase.expectedErr)
			assert.Equal(t, testCase.name, builder.Definition.Name)
		})
	}
}

func TestPullSeedGenerator(t *testing.T) {
	t.Parallel()

	testhelper.NewClusterScopedPullTestConfig(
		PullSeedGenerator,
		lcasgv1.AddToScheme,
		seedGeneratorGVK,
	).ExecuteTests(t)
}

func TestSeedGeneratorBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[lcasgv1.SeedGenerator, SeedGeneratorBuilder](
		lcasgv1.AddToScheme,
		seedGeneratorGVK,
		testhelper.ResourceScopeClusterScoped,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		Run(t)
}

func TestSeedGeneratorWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(
		testhelper.NewCommonTestConfig[lcasgv1.SeedGenerator, SeedGeneratorBuilder](
			lcasgv1.AddToScheme,
			seedGeneratorGVK,
			testhelper.ResourceScopeClusterScoped,
		)).ExecuteTests(t)
}

func TestSeedGeneratorWithSeedImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		seedImage     string
		expectedError string
		expectedValue string
	}{
		{
			name:          "sets seed image",
			seedImage:     "testSeedImage",
			expectedValue: "testSeedImage",
		},
		{
			name:          "rejects empty seed image",
			expectedError: "seedImage 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidSeedGeneratorBuilder(newSeedGeneratorTestClient(nil))
			result := builder.WithSeedImage(testCase.seedImage)

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
			} else {
				require.NoError(t, result.GetError())
			}

			assert.Equal(t, testCase.expectedValue, result.Definition.Spec.SeedImage)

			invalidBuilder := buildValidSeedGeneratorBuilder(newSeedGeneratorTestClient(nil))
			invalidBuilder.SetError(fmt.Errorf("invalid builder"))
			invalidResult := invalidBuilder.WithSeedImage("should-not-be-set")

			require.Error(t, invalidResult.GetError())
			assert.Empty(t, invalidResult.Definition.Spec.SeedImage)
		})
	}
}

func TestSeedGeneratorWithRecertImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		recertImage   string
		expectedError string
		expectedValue string
	}{
		{
			name:          "sets recert image",
			recertImage:   "testRecertImage",
			expectedValue: "testRecertImage",
		},
		{
			name:          "rejects empty recert image",
			expectedError: "recertImage 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidSeedGeneratorBuilder(newSeedGeneratorTestClient(nil))
			result := builder.WithRecertImage(testCase.recertImage)

			if testCase.expectedError != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedError)
			} else {
				require.NoError(t, result.GetError())
			}

			assert.Equal(t, testCase.expectedValue, result.Definition.Spec.RecertImage)

			invalidBuilder := buildValidSeedGeneratorBuilder(newSeedGeneratorTestClient(nil))
			invalidBuilder.SetError(fmt.Errorf("invalid builder"))
			invalidResult := invalidBuilder.WithRecertImage("should-not-be-set")

			require.Error(t, invalidResult.GetError())
			assert.Empty(t, invalidResult.Definition.Spec.RecertImage)
		})
	}
}

func TestSeedGeneratorWaitUntilComplete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		object        *lcasgv1.SeedGenerator
		expectedError error
	}{
		{
			name: "waits until completed condition is true",
			object: &lcasgv1.SeedGenerator{
				ObjectMeta: metav1.ObjectMeta{Name: seedImageName},
				Status: lcasgv1.SeedGeneratorStatus{
					Conditions: []metav1.Condition{{
						Status: isFalse,
						Type:   conditionTypeSeedGenCompleted,
						Reason: isComplete,
					}},
				},
			},
			expectedError: context.DeadlineExceeded,
		},
		{
			name: "completed condition succeeds",
			object: &lcasgv1.SeedGenerator{
				ObjectMeta: metav1.ObjectMeta{Name: seedImageName},
				Status: lcasgv1.SeedGeneratorStatus{
					Conditions: []metav1.Condition{{
						Status: isTrue,
						Type:   conditionTypeSeedGenCompleted,
						Reason: isComplete,
					}},
				},
			},
		},
		{
			name:          "missing resource returns error",
			expectedError: fmt.Errorf("seedgenerator object %s does not exist", seedImageName),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object
			if testCase.object != nil {
				objects = append(objects, testCase.object)
			}

			builder := buildValidSeedGeneratorBuilder(newSeedGeneratorTestClient(objects))
			result, err := builder.WaitUntilComplete(100 * time.Millisecond)

			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				require.NotNil(t, result)
			}
		})
	}
}

func buildValidSeedGeneratorBuilder(apiClient *clients.Settings) *SeedGeneratorBuilder {
	return NewSeedGeneratorBuilder(apiClient, seedImageName)
}

func newSeedGeneratorTestClient(objects []runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: []clients.SchemeAttacher{lcasgv1.AddToScheme},
	})
}
