package storage

import (
	"testing"

	noobaav1alpha1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultObjectBucketClaimName      = "obc-test"
	defaultObjectBucketClaimNamespace = "obc-space"
)

func TestNewObjectBucketClaimBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewObjectBucketClaimBuilder,
		noobaav1alpha1.AddToScheme,
		objectBucketClaimGVK,
	).ExecuteTests(t)
}

func TestPullObjectBucketClaim(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullObjectBucketClaim,
		noobaav1alpha1.AddToScheme,
		objectBucketClaimGVK,
	).ExecuteTests(t)
}

func TestObjectBucketClaimBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[noobaav1alpha1.ObjectBucketClaim, ObjectBucketClaimBuilder](
		noobaav1alpha1.AddToScheme,
		objectBucketClaimGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestObjectBucketClaimUpdateStorageClassName(t *testing.T) {
	t.Parallel()

	builder := buildValidObjectBucketClaimTestBuilder(buildObjectBucketClaimClientWithDummyObject())
	builder.WithStorageClassName("gp2")

	result, err := builder.Update()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Object)
	assert.Equal(t, "gp2", result.Object.Spec.StorageClassName)
}

func TestObjectBucketClaimWithStorageClassName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		storageClassName string
		invalidBuilder   bool
		assertError      func(error) bool
	}{
		{
			name:             "sets storage class name",
			storageClassName: "gp2",
			assertError:      func(err error) bool { return err == nil },
		},
		{
			name:             "empty storage class name returns error",
			storageClassName: "",
			assertError: func(err error) bool {
				return err != nil && err.Error() == "'storageClassName' argument cannot be empty"
			},
		},
		{
			name:             "invalid builder short circuits",
			storageClassName: "gp2",
			invalidBuilder:   true,
			assertError:      commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidObjectBucketClaimTestBuilder(buildObjectBucketClaimClientWithDummyObject())
			if testCase.invalidBuilder {
				builder = buildInvalidObjectBucketClaimTestBuilder(buildObjectBucketClaimClientWithDummyObject())
			}

			result := builder.WithStorageClassName(testCase.storageClassName)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.storageClassName, result.Definition.Spec.StorageClassName)
			}
		})
	}
}

func TestObjectBucketClaimWithGenerateBucketName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		generateBucketName string
		invalidBuilder     bool
		assertError        func(error) bool
	}{
		{
			name:               "sets generated bucket name",
			generateBucketName: "test-bucket-odf",
			assertError:        func(err error) bool { return err == nil },
		},
		{
			name:               "empty generated bucket name returns error",
			generateBucketName: "",
			assertError: func(err error) bool {
				return err != nil && err.Error() == "'generateBucketName' argument cannot be empty"
			},
		},
		{
			name:               "invalid builder short circuits",
			generateBucketName: "test-bucket-odf",
			invalidBuilder:     true,
			assertError:        commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := buildValidObjectBucketClaimTestBuilder(buildObjectBucketClaimClientWithDummyObject())
			if testCase.invalidBuilder {
				builder = buildInvalidObjectBucketClaimTestBuilder(buildObjectBucketClaimClientWithDummyObject())
			}

			result := builder.WithGenerateBucketName(testCase.generateBucketName)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.generateBucketName, result.Definition.Spec.GenerateBucketName)
			}
		})
	}
}

func buildValidObjectBucketClaimTestBuilder(apiClient *clients.Settings) *ObjectBucketClaimBuilder {
	return NewObjectBucketClaimBuilder(
		apiClient, defaultObjectBucketClaimName, defaultObjectBucketClaimNamespace)
}

func buildInvalidObjectBucketClaimTestBuilder(apiClient *clients.Settings) *ObjectBucketClaimBuilder {
	return NewObjectBucketClaimBuilder(apiClient, "", defaultObjectBucketClaimNamespace)
}

func buildObjectBucketClaimClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyObjectBucketClaim(),
		SchemeAttachers: []clients.SchemeAttacher{noobaav1alpha1.AddToScheme},
	})
}

func buildDummyObjectBucketClaim() []runtime.Object {
	return []runtime.Object{
		&noobaav1alpha1.ObjectBucketClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultObjectBucketClaimName,
				Namespace: defaultObjectBucketClaimNamespace,
			},
		},
	}
}
