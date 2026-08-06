package capi

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	clusterv1beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/capi/cluster/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var clusterGVK = clusterv1beta1.GroupVersion.WithKind("Cluster")

func TestNewClusterBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[clusterv1beta1.Cluster, ClusterBuilder](
		NewClusterBuilder, clusterv1beta1.AddToScheme, clusterGVK).
		ExecuteTests(t)
}

func TestPullCluster(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[clusterv1beta1.Cluster, ClusterBuilder](
		PullCluster, clusterv1beta1.AddToScheme, clusterGVK).
		ExecuteTests(t)
}

func TestClusterBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newClusterCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestClusterWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newClusterCommonTestConfig()).ExecuteTests(t)
}

func TestWithControlPlaneEndpoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *ClusterBuilder
		host        string
		port        int32
		assertError func(error) bool
		expectSet   bool
	}{
		{
			name: "valid host and port",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			host:        "example.com",
			port:        6443,
			assertError: func(err error) bool { return err == nil },
			expectSet:   true,
		},
		{
			name: "empty host sets builder error",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			host:        "",
			port:        6443,
			assertError: func(err error) bool { return err != nil && err.Error() == "'host' cannot be empty" },
		},
		{
			name: "port zero sets builder error",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			host:        "example.com",
			port:        0,
			assertError: func(err error) bool { return err != nil && err.Error() == "'port' must be between 1 and 65535, got 0" },
		},
		{
			name: "port exceeds 65535 sets builder error",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			host: "example.com",
			port: 65536,
			assertError: func(err error) bool {
				return err != nil && err.Error() == "'port' must be between 1 and 65535, got 65536"
			},
		},
		{
			name: "invalid builder short circuits",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "", "test-ns")
			},
			host:        "example.com",
			port:        6443,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithControlPlaneEndpoint(testCase.host, testCase.port)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if testCase.expectSet {
				assert.Equal(t, testCase.host, result.Definition.Spec.ControlPlaneEndpoint.Host)
				assert.Equal(t, testCase.port, result.Definition.Spec.ControlPlaneEndpoint.Port)
			}
		})
	}
}

func TestWithControlPlaneRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *ClusterBuilder
		apiVersion  string
		kind        string
		refName     string
		assertError func(error) bool
		expectSet   bool
	}{
		{
			name: "valid reference",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			apiVersion:  "controlplane.cluster.x-k8s.io/v1alpha3",
			kind:        "OpenshiftAssistedControlPlane",
			refName:     "my-cp",
			assertError: func(err error) bool { return err == nil },
			expectSet:   true,
		},
		{
			name: "empty apiVersion sets builder error",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			apiVersion:  "",
			kind:        "OpenshiftAssistedControlPlane",
			refName:     "my-cp",
			assertError: func(err error) bool { return err != nil && err.Error() == "'apiVersion' cannot be empty" },
		},
		{
			name: "empty kind sets builder error",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			apiVersion:  "controlplane.cluster.x-k8s.io/v1alpha3",
			kind:        "",
			refName:     "my-cp",
			assertError: func(err error) bool { return err != nil && err.Error() == "'kind' cannot be empty" },
		},
		{
			name: "empty name sets builder error",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			apiVersion:  "controlplane.cluster.x-k8s.io/v1alpha3",
			kind:        "OpenshiftAssistedControlPlane",
			refName:     "",
			assertError: func(err error) bool { return err != nil && err.Error() == "'name' cannot be empty" },
		},
		{
			name: "invalid builder short circuits",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "", "test-ns")
			},
			apiVersion:  "controlplane.cluster.x-k8s.io/v1alpha3",
			kind:        "OpenshiftAssistedControlPlane",
			refName:     "my-cp",
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithControlPlaneRef(testCase.apiVersion, testCase.kind, testCase.refName)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if testCase.expectSet {
				require.NotNil(t, result.Definition.Spec.ControlPlaneRef)
				assert.Equal(t, testCase.apiVersion, result.Definition.Spec.ControlPlaneRef.APIVersion)
				assert.Equal(t, testCase.kind, result.Definition.Spec.ControlPlaneRef.Kind)
				assert.Equal(t, testCase.refName, result.Definition.Spec.ControlPlaneRef.Name)
			}
		})
	}
}

func TestWithInfrastructureRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *ClusterBuilder
		apiVersion  string
		kind        string
		refName     string
		assertError func(error) bool
		expectSet   bool
	}{
		{
			name: "valid reference",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			apiVersion:  "infrastructure.cluster.x-k8s.io/v1beta1",
			kind:        "Metal3Cluster",
			refName:     "my-infra",
			assertError: func(err error) bool { return err == nil },
			expectSet:   true,
		},
		{
			name: "empty apiVersion sets builder error",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			apiVersion:  "",
			kind:        "Metal3Cluster",
			refName:     "my-infra",
			assertError: func(err error) bool { return err != nil && err.Error() == "'apiVersion' cannot be empty" },
		},
		{
			name: "empty kind sets builder error",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			apiVersion:  "infrastructure.cluster.x-k8s.io/v1beta1",
			kind:        "",
			refName:     "my-infra",
			assertError: func(err error) bool { return err != nil && err.Error() == "'kind' cannot be empty" },
		},
		{
			name: "empty name sets builder error",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			apiVersion:  "infrastructure.cluster.x-k8s.io/v1beta1",
			kind:        "Metal3Cluster",
			refName:     "",
			assertError: func(err error) bool { return err != nil && err.Error() == "'name' cannot be empty" },
		},
		{
			name: "invalid builder short circuits",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "", "test-ns")
			},
			apiVersion:  "infrastructure.cluster.x-k8s.io/v1beta1",
			kind:        "Metal3Cluster",
			refName:     "my-infra",
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithInfrastructureRef(testCase.apiVersion, testCase.kind, testCase.refName)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if testCase.expectSet {
				require.NotNil(t, result.Definition.Spec.InfrastructureRef)
				assert.Equal(t, testCase.apiVersion, result.Definition.Spec.InfrastructureRef.APIVersion)
				assert.Equal(t, testCase.kind, result.Definition.Spec.InfrastructureRef.Kind)
				assert.Equal(t, testCase.refName, result.Definition.Spec.InfrastructureRef.Name)
			}
		})
	}
}

func TestWithPaused(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     func() *ClusterBuilder
		paused      bool
		assertError func(error) bool
	}{
		{
			name: "sets paused true",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			paused:      true,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name: "sets paused false",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "test", "test-ns")
			},
			paused:      false,
			assertError: func(err error) bool { return err == nil },
		},
		{
			name: "invalid builder short circuits",
			builder: func() *ClusterBuilder {
				return NewClusterBuilder(newClusterTestClient(), "", "test-ns")
			},
			paused:      true,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithPaused(testCase.paused)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if result.GetError() == nil {
				assert.Equal(t, testCase.paused, result.Definition.Spec.Paused)
			}
		})
	}
}

func newClusterCommonTestConfig() testhelper.CommonTestConfig[
	clusterv1beta1.Cluster, ClusterBuilder, *clusterv1beta1.Cluster, *ClusterBuilder,
] {
	return testhelper.NewCommonTestConfig[clusterv1beta1.Cluster, ClusterBuilder](
		clusterv1beta1.AddToScheme, clusterGVK, testhelper.ResourceScopeNamespaced)
}

func newClusterTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{clusterv1beta1.AddToScheme},
	})
}
