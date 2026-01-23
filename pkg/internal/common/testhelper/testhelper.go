package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceScope is a type that represents the scope of a resource, i.e. cluster-scoped or namespaced.
type ResourceScope int

const (
	// ResourceScopeClusterScoped represents a cluster-scoped resource.
	ResourceScopeClusterScoped ResourceScope = iota
	// ResourceScopeNamespaced represents a namespaced resource.
	ResourceScopeNamespaced
)

// IsClusterScoped returns true if the resource scope is cluster-scoped.
func (rs ResourceScope) IsClusterScoped() bool {
	return rs == ResourceScopeClusterScoped
}

// IsNamespaced returns true if the resource scope is namespaced.
func (rs ResourceScope) IsNamespaced() bool {
	return rs == ResourceScopeNamespaced
}

// CommonTestConfig is the shared configuration used by the common test helper configs. It avoids repeating the scheme
// attacher, expected GVK, and resource scope across per-method test configs.
type CommonTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	SchemeAttacher clients.SchemeAttacher
	ExpectedGVK    schema.GroupVersionKind
	ResourceScope  ResourceScope
}

// NewCommonTestConfig creates a new CommonTestConfig with the given parameters. By using a constructor instead of a
// struct literal, some of the type parameters can be inferred.
func NewCommonTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
	resourceScope ResourceScope,
) CommonTestConfig[O, B, SO, SB] {
	return CommonTestConfig[O, B, SO, SB]{
		SchemeAttacher: schemeAttacher,
		ExpectedGVK:    expectedGVK,
		ResourceScope:  resourceScope,
	}
}

// TestExecutor is implemented by test helper configs that can execute their test suite.
type TestExecutor interface {
	ExecuteTests(t *testing.T)
	Name() string
}

// TestSuite represents a collection of test executors that are run in parallel subtests.
type TestSuite struct {
	executors []TestExecutor
}

// NewTestSuite creates a new TestSuite with the given test executors. Passing the test executors as variadic arguments
// is equivalent to using the With method to add each executor individually.
func NewTestSuite(executors ...TestExecutor) *TestSuite {
	return &TestSuite{
		executors: executors,
	}
}

// With adds a test executor to the test suite.
func (suite *TestSuite) With(executor TestExecutor) *TestSuite {
	suite.executors = append(suite.executors, executor)

	return suite
}

// Run executes the test suite. Each executor is run in a parallel subtest using the executor's Name method as the
// subtest name.
func (suite *TestSuite) Run(t *testing.T) {
	t.Helper()

	for _, executor := range suite.executors {
		t.Run(executor.Name(), func(t *testing.T) {
			t.Parallel()

			executor.ExecuteTests(t)
		})
	}
}
