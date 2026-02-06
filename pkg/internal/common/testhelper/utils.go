// Package testhelper provides reusable test utilities for the common builder patterns. These helpers enforce consistent
// test coverage for CRUD operations across all builder types by providing table-driven test generators that verify both
// success and failure scenarios.
package testhelper

import (
	"context"
	"errors"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Test fixture constants used across all helper tests. These provide deterministic values that can be asserted against
// in test expectations without worrying about test pollution between test cases.
const (
	testResourceName      = "test-resource-name"
	testResourceNamespace = "test-resource-namespace"
	testAnnotationKey     = "test-annotation-key"
	testAnnotationValue   = "test-annotation-value"
)

// Sentinel errors for simulating API failures in tests. Each corresponds to a specific Kubernetes client operation and
// is returned by the matching testFailing* interceptor. Tests use the corresponding isAPICallFailedWith* predicates to
// verify error handling.
var (
	errCreateFailure = errors.New("simulated create failure")
	errGetFailure    = errors.New("simulated get failure")
	errListFailure   = errors.New("simulated list failure")
	errUpdateFailure = errors.New("simulated update failure")
	errDeleteFailure = errors.New("simulated delete failure")

	// errInvalidBuilder is injected into builder.errorMsg to test validation logic. Unlike the API errors above,
	// this simulates a builder-level validation failure rather than a Kubernetes API failure.
	errInvalidBuilder = errors.New("invalid builder error")

	// errSchemeAttachment simulates a failure when registering types with the runtime scheme. Used to verify that
	// builders properly handle scheme registration errors during client setup.
	errSchemeAttachment = errors.New("scheme attachment failed")
)

// testFailingCreate is an interceptor function that always returns errCreateFailure. Used with fake client interceptors
// to simulate Kubernetes API create failures.
func testFailingCreate(
	ctx context.Context,
	client runtimeclient.WithWatch,
	obj runtimeclient.Object,
	opts ...runtimeclient.CreateOption,
) error {
	return errCreateFailure
}

// testFailingDelete is an interceptor function that always returns errDeleteFailure. Used with fake client interceptors
// to simulate Kubernetes API delete failures.
func testFailingDelete(
	ctx context.Context,
	client runtimeclient.WithWatch,
	obj runtimeclient.Object,
	opts ...runtimeclient.DeleteOption,
) error {
	return errDeleteFailure
}

// testFailingUpdate is an interceptor function that always returns errUpdateFailure. Used with fake client interceptors
// to simulate Kubernetes API update failures.
func testFailingUpdate(
	ctx context.Context,
	client runtimeclient.WithWatch,
	obj runtimeclient.Object,
	opts ...runtimeclient.UpdateOption,
) error {
	return errUpdateFailure
}

// testFailingList is an interceptor function that always returns errListFailure. Used with fake client interceptors to
// simulate Kubernetes API list failures.
func testFailingList(
	ctx context.Context,
	client runtimeclient.WithWatch,
	list runtimeclient.ObjectList,
	opts ...runtimeclient.ListOption,
) error {
	return errListFailure
}

// testFailingGet is an interceptor function that always returns errGetFailure. Used with fake client interceptors to
// simulate Kubernetes API get failures.
func testFailingGet(
	ctx context.Context,
	client runtimeclient.WithWatch,
	key runtimeclient.ObjectKey,
	obj runtimeclient.Object,
	opts ...runtimeclient.GetOption,
) error {
	return errGetFailure
}

// Error predicate functions for use in table-driven tests. These are passed to test case assertError fields to verify
// the correct error type is returned. Using predicates rather than direct error comparison allows tests to check error
// categories without coupling to specific error messages or wrapped error chains.

func isErrorNil(err error) bool {
	return err == nil
}

func isAPICallFailedWithCreate(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "create")
}

func isAPICallFailedWithGet(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "get")
}

func isAPICallFailedWithList(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "list")
}

func isAPICallFailedWithUpdate(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "update")
}

func isAPICallFailedWithDelete(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "delete")
}

func isInvalidBuilder(err error) bool {
	return errors.Is(err, errInvalidBuilder)
}

// buildDummyObject creates a minimal Kubernetes object with only name and namespace set. The namespace is always set
// even for cluster-scoped resources since the Kubernetes API simply ignores it for those types. This avoids needing
// separate constructors for namespaced vs cluster-scoped test objects.
func buildDummyObject[O any, SO common.ObjectPointer[O]](name, namespace string) SO {
	var dummyObject SO = new(O)

	dummyObject.SetName(name)
	dummyObject.SetNamespace(namespace)

	return dummyObject
}

// testFailingSchemeAttacher always returns errSchemeAttachment. Used to verify that builder constructors properly
// propagate scheme registration failures rather than silently continuing with an incomplete scheme.
func testFailingSchemeAttacher(scheme *runtime.Scheme) error {
	return errSchemeAttachment
}

// createSchemeAttacherGVKTest generates a test function that verifies a scheme attacher correctly registers the
// expected GroupVersionKind. This catches misconfigurations where a builder's scheme attacher registers the wrong type
// or version.
//
// The returned function is designed to be passed directly to t.Run() as a subtest.
func createSchemeAttacherGVKTest[O any, SO common.ObjectPointer[O]](
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		t.Parallel()

		scheme := runtime.NewScheme()
		err := schemeAttacher(scheme)
		require.NoError(t, err, "schemeAttacher failed when attaching to a fresh scheme")

		var obj SO = new(O)

		kinds, _, err := scheme.ObjectKinds(obj)
		assert.NoError(t, err, "scheme.ObjectKinds failed for test object; scheme attacher may be wrong")
		assert.Contains(t, kinds, expectedGVK, "scheme attacher did not register the expected GVK")
	}
}
