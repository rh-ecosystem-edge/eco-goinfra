package common

import (
	"context"
	"errors"
	"reflect"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// objectPointer is a type constraint that requires a type be a pointer to O and implement the runtimeclient.Object
// interface. The type parameter O is meant to be a K8s resource, such as corev1.Namespace. In that case,
// *corev1.Namespace would satisfy the constraint objectPointer[corev1.Namespace].
type objectPointer[O any] interface {
	*O
	runtimeclient.Object
}

// Builder represents the set of methods that must be present to use the common versions of CRUD and other methods.
// Since each builder struct is a different type, this interface allows common functions to update fields on the
// builder. Generally, consumers of eco-goinfra should not call these methods.
//
// The type parameter O (short for object) is expected to be the struct that represents a K8s resource, such as
// corev1.Namespace. SO (short for star O) is the pointer to O, with the additional constraint of that pointer
// implementing runtimeclient.Object. To continue the example, this would be *corev1.Namespace.
//
// Although only SO appears in the interface definition, it is important to have access to the derefenced form of the
// type so we may do new(O) and get a runtimeclient.Object.
type Builder[O any, SO objectPointer[O]] interface {
	// GetDefinition allows for getting the desired form of a K8s resource from the builder.
	GetDefinition() SO
	// SetDefinition allows for updating the desired form of the K8s resource.
	SetDefinition(SO)

	// GetObject allows for getting the form of a K8s resource, as last pulled from the cluster.
	GetObject() SO
	// SetObject allows for updating what the K8s resource last was on the cluster.
	SetObject(SO)

	// GetErrorMessage returns the error stored in the builder from methods that do not return errors.
	GetErrorMessage() string
	// SetErrorMessage allows for updating the error message stored in the builder.
	SetErrorMessage(string)

	// GetClient returns the client used for connecting with the K8s cluster.
	GetClient() runtimeclient.Client
	// SetClient allows for updating the client used to connect to the K8s cluster. Since this is a simple setter,
	// it will not handle updating the scheme of the client and should generally be avoided outside of creating the
	// builder.
	SetClient(runtimeclient.Client)

	// GetGVK returns the GVK for the resource the builder represents, even if the builder is zero-valued.
	GetGVK() schema.GroupVersionKind
	// SetGVK allows for updating the GVK for the resource the builder represents. This method solves the problem
	// where New*Builder and Pull*Builder will take the GVK from the method on the embedding builder, but then calls
	// to CRUD methods will use the GVK from the EmbeddableBuilder. Since these initialization functions will call
	// SetGVK, this is a somewhat nonobvious way to pass information from the embedding builder to the
	// EmbeddableBuilder.
	SetGVK(schema.GroupVersionKind)
}

// MixinAttacher is an interface for types which require a step during initialization of a builder to ensure mixins are
// all attached. In practice, this is meant to be implemented by the resource-specific builders so that any mixins they
// embed can be connected to the EmbeddableBuilder.
//
// This interface is defined separately from the Builder interface since while the resource-specific builders must
// implement the Builder interface, they do not need to implement the MixinAttacher interface. Similarly,
// EmbeddableBuilder should not implement this interface even though it is a Builder.
type MixinAttacher interface {
	// AttachMixins ensures that all mixins are attached to their base builders. This method will be called on the
	// zero-value of builder pointers right after allocation, provided the builder also implements MixinAttacher.
	AttachMixins()
}

// builderPointer is similar to objectPointer and is a constraint that is satisfied by a Builder that is a pointer. It
// exists for the same reason as objectPointer: needing access to the dereferenced form of builders to construct new
// ones.
type builderPointer[B, O any, SO objectPointer[O]] interface {
	*B
	Builder[O, SO]
}

// NewClusterScopedBuilder creates a new builder for a cluster-scoped resource. It is generic over the actual builder
// type and uses the methods from the Builder interface to create the actual builder. Generic parameters are ordered so
// that SO and SB can be elided and only O and B must be provided.
func NewClusterScopedBuilder[O, B any, SO objectPointer[O], SB builderPointer[B, O, SO]](
	apiClient runtimeclient.Client, schemeAttacher clients.SchemeAttacher, name string) SB {
	var builder SB = new(B)

	if mixinAttacher, ok := any(builder).(MixinAttacher); ok {
		mixinAttacher.AttachMixins()
	}

	logNewClusterScopedBuilderInitializing(builder.GetGVK().Kind, name)

	if apiClient == nil || reflect.ValueOf(apiClient).IsNil() {
		logAPIClientNil(builder.GetGVK().Kind)

		return nil
	}

	err := schemeAttacher(apiClient.Scheme())
	if err != nil {
		logSchemedFailedToAttach(builder.GetGVK().Kind, err)

		return nil
	}

	builder.SetGVK(builder.GetGVK())
	builder.SetClient(apiClient)
	builder.SetDefinition(new(O))
	builder.GetDefinition().SetName(name)

	if name == "" {
		logBuilderNameEmpty(builder.GetGVK().Kind)

		builder.SetErrorMessage(getBuilderNameEmptyError(builder.GetGVK().Kind).Error())

		return builder
	}

	return builder
}

// NewNamespacedBuilder creates a new builder for a namespaced resource. It is generic over the actual builder type and
// uses the methods from the Builder interface to create the actual builder. Generic parameters are ordered so that SO
// and SB can be elided and only O and B must be provided.
func NewNamespacedBuilder[O, B any, SO objectPointer[O], SB builderPointer[B, O, SO]](
	apiClient runtimeclient.Client, schemeAttacher clients.SchemeAttacher, name, nsname string) SB {
	var builder SB = new(B)

	if mixinAttacher, ok := any(builder).(MixinAttacher); ok {
		mixinAttacher.AttachMixins()
	}

	logNewNamespacedBuilderInitializing(builder.GetGVK().Kind, name, nsname)

	if apiClient == nil || reflect.ValueOf(apiClient).IsNil() {
		logAPIClientNil(builder.GetGVK().Kind)

		return nil
	}

	err := schemeAttacher(apiClient.Scheme())
	if err != nil {
		logSchemedFailedToAttach(builder.GetGVK().Kind, err)

		return nil
	}

	// To clarify the comment on the Builder interface, this line is intended to get the GVK from the embedding
	// builder (e.g. NamespaceBuilder) and then set it on the EmbeddableBuilder. This is arguably an anti-pattern,
	// but it is necessary to have functions which create builders that have CRUD methods coming from an embedded
	// struct.
	builder.SetGVK(builder.GetGVK())
	builder.SetClient(apiClient)
	builder.SetDefinition(new(O))
	builder.GetDefinition().SetName(name)
	builder.GetDefinition().SetNamespace(nsname)

	if name == "" {
		logBuilderNameEmpty(builder.GetGVK().Kind)

		builder.SetErrorMessage(getBuilderNameEmptyError(builder.GetGVK().Kind).Error())

		return builder
	}

	if nsname == "" {
		logBuilderNamespaceEmpty(builder.GetGVK().Kind)

		builder.SetErrorMessage(getBuilderNamespaceEmptyError(builder.GetGVK().Kind).Error())

		return builder
	}

	return builder
}

// PullClusterScopedBuilder creates a new Builder for a cluster-scoped resource, pulling the resource from the cluster.
// It is generic over the actual builder type and uses the methods from the Builder interface to create the actual
// builder. Generic parameters are ordered so that SO and SB can be elided and only O and B must be provided.
func PullClusterScopedBuilder[O, B any, SO objectPointer[O], SB builderPointer[B, O, SO]](
	apiClient runtimeclient.Client, schemeAttacher clients.SchemeAttacher, name string) (SB, error) {
	var builder SB = new(B)

	if mixinAttacher, ok := any(builder).(MixinAttacher); ok {
		mixinAttacher.AttachMixins()
	}

	logPullClusterScopedBuilderPulling(builder.GetGVK().Kind, name)

	if apiClient == nil || reflect.ValueOf(apiClient).IsNil() {
		logAPIClientNil(builder.GetGVK().Kind)

		return nil, getAPIClientNilError(builder.GetGVK().Kind)
	}

	err := schemeAttacher(apiClient.Scheme())
	if err != nil {
		logSchemedFailedToAttach(builder.GetGVK().Kind, err)

		return nil, wrapSchemeAttacherError(builder.GetGVK().Kind, err)
	}

	builder.SetGVK(builder.GetGVK())
	builder.SetClient(apiClient)
	builder.SetDefinition(new(O))
	builder.GetDefinition().SetName(name)

	if name == "" {
		logBuilderNameEmpty(builder.GetGVK().Kind)

		return nil, getBuilderNameEmptyError(builder.GetGVK().Kind)
	}

	if !Exists(builder) {
		logBuilderNotFound(builder.GetGVK().Kind)

		return nil, getBuilderNotFoundError(builder.GetGVK().Kind)
	}

	builder.SetDefinition(builder.GetObject())

	return builder, nil
}

// PullNamespacedBuilder creates a new Builder for a namespaced resource, pulling the resource from the cluster.
// It is generic over the actual builder type and uses the methods from the Builder interface to create the actual
// builder. Generic parameters are ordered so that SO and SB can be elided and only O and B must be provided.
func PullNamespacedBuilder[O, B any, SO objectPointer[O], SB builderPointer[B, O, SO]](
	apiClient runtimeclient.Client, schemeAttacher clients.SchemeAttacher, name, nsname string) (SB, error) {
	var builder SB = new(B)

	if mixinAttacher, ok := any(builder).(MixinAttacher); ok {
		mixinAttacher.AttachMixins()
	}

	logPullNamespacedBuilderPulling(builder.GetGVK().Kind, name, nsname)

	if apiClient == nil || reflect.ValueOf(apiClient).IsNil() {
		logAPIClientNil(builder.GetGVK().Kind)

		return nil, getAPIClientNilError(builder.GetGVK().Kind)
	}

	err := schemeAttacher(apiClient.Scheme())
	if err != nil {
		logSchemedFailedToAttach(builder.GetGVK().Kind, err)

		return nil, wrapSchemeAttacherError(builder.GetGVK().Kind, err)
	}

	builder.SetGVK(builder.GetGVK())
	builder.SetClient(apiClient)
	builder.SetDefinition(new(O))
	builder.GetDefinition().SetName(name)
	builder.GetDefinition().SetNamespace(nsname)

	if name == "" {
		logBuilderNameEmpty(builder.GetGVK().Kind)

		return nil, getBuilderNameEmptyError(builder.GetGVK().Kind)
	}

	if nsname == "" {
		logBuilderNamespaceEmpty(builder.GetGVK().Kind)

		return nil, getBuilderNamespaceEmptyError(builder.GetGVK().Kind)
	}

	if !Exists(builder) {
		logBuilderNotFound(builder.GetGVK().Kind)

		return nil, getBuilderNotFoundError(builder.GetGVK().Kind)
	}

	builder.SetDefinition(builder.GetObject())

	return builder, nil
}

// Get pulls the resource from the cluster and returns it. It does not modify the builder.
func Get[O any, SO objectPointer[O]](builder Builder[O, SO]) (SO, error) {
	if err := Validate(builder); err != nil {
		return nil, err
	}

	logBuilderGet(builder)

	var object SO = new(O)

	err := builder.GetClient().Get(context.TODO(), runtimeclient.ObjectKeyFromObject(builder.GetDefinition()), object)
	if err != nil {
		return nil, wrapGetError(builder, err)
	}

	return object, nil
}

// Exists checks if the resource exists in the cluster. If the resource does exist, the builder's object is updated with
// the resource and this returns true. If the resource does not exist or an error was encountered getting the resource,
// this returns false without modifying the builder.
func Exists[O any, SO objectPointer[O]](builder Builder[O, SO]) bool {
	if err := Validate(builder); err != nil {
		return false
	}

	logBuilderExists(builder)

	object, err := Get(builder)
	if err != nil {
		logBuilderNotFoundWithError(builder.GetGVK().Kind, err)

		return false
	}

	builder.SetObject(object)

	return true
}

// Delete deletes the resource from the cluster. It immediately tries to delete the resource and if successful, or the
// resource did not exist, the builder's object is set to nil. Otherwise, the error is wrapped and returned without
// modifying the builder.
func Delete[O any, SO objectPointer[O]](builder Builder[O, SO]) error {
	if err := Validate(builder); err != nil {
		return err
	}

	logBuilderDelete(builder)

	err := builder.GetClient().Delete(context.TODO(), builder.GetDefinition())
	if err == nil || k8serrors.IsNotFound(err) {
		builder.SetObject(nil)

		return nil
	}

	return wrapDeleteError(builder, err)
}

// Update updates the resource on the cluster using the builder's definition. It immediately tries to update the
// resource and if successful, will update the builder's object to be the definition. Otherwise, it checks to see if the
// error is because the resource did not exist, returning with an error if so. If the error is for any other reason, the
// behavior depends on the force flag.
//
// If force is true, the resource will be deleted and recreated. Otherwise, the error is wrapped and returned without
// modifying the builder.
func Update[O any, SO objectPointer[O]](builder Builder[O, SO], force bool) error {
	if err := Validate(builder); err != nil {
		return err
	}

	logBuilderUpdate(builder, force)

	if !Exists(builder) {
		logBuilderNotFound(builder.GetGVK().Kind)

		return getBuilderNotFoundError(builder.GetGVK().Kind)
	}

	// Object is set by Exists so we do not need to do a nil check here.
	builder.GetDefinition().SetResourceVersion(builder.GetObject().GetResourceVersion())

	err := builder.GetClient().Update(context.TODO(), builder.GetDefinition())
	if err == nil {
		builder.SetObject(builder.GetDefinition())

		return nil
	}

	if !force {
		return wrapUpdateError(builder, err)
	}

	err = Delete(builder)
	if err != nil {
		return wrapForceUpdateDeleteError(builder, err)
	}

	builder.GetDefinition().SetResourceVersion("")

	err = Create(builder)
	if err != nil {
		return wrapForceUpdateCreateError(builder, err)
	}

	return nil
}

// Create creates the definition on the cluster. If the resource already exists, this is a no-op.
func Create[O any, SO objectPointer[O]](builder Builder[O, SO]) error {
	if err := Validate(builder); err != nil {
		return err
	}

	logBuilderCreate(builder)

	if Exists(builder) {
		logBuilderAlreadyExists(builder)

		return nil
	}

	err := builder.GetClient().Create(context.TODO(), builder.GetDefinition())
	if err != nil {
		return wrapCreateError(builder, err)
	}

	builder.SetObject(builder.GetDefinition())

	return nil
}

// Validate checks that the builder is valid, that is, it is non-nil, has a non-nil definition, has a non-nil client,
// and has no error message. Additional checks are performed on any interface so that we know it is not nil and its
// concrete type is not nil.
func Validate[O any, SO objectPointer[O]](builder Builder[O, SO]) error {
	if builder == nil || reflect.ValueOf(builder).IsNil() {
		logBuilderUninitialized()

		return getBuilderUninitializedError()
	}

	resourceCRD := builder.GetGVK().Kind

	if builder.GetDefinition() == nil {
		logBuilderUndefined(resourceCRD)

		return getBuilderDefinitionNilError(resourceCRD)
	}

	client := builder.GetClient()
	if client == nil || reflect.ValueOf(client).IsNil() {
		logBuilderAPIClientNil(resourceCRD)

		return getBuilderAPIClientNilError(resourceCRD)
	}

	if builder.GetErrorMessage() != "" {
		logBuilderErrorMessage(resourceCRD, builder.GetErrorMessage())

		return errors.New(builder.GetErrorMessage())
	}

	return nil
}
