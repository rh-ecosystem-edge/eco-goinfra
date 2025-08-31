package ovn

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/msg"
	ovnv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ovn/routeadvertisement"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// RouteAdvertisementBuilder provides struct for the RouteAdvertisement object containing connection to
// the cluster and the RouteAdvertisement definitions.
type RouteAdvertisementBuilder struct {
	Definition *ovnv1.RouteAdvertisement
	Object     *ovnv1.RouteAdvertisement
	apiClient  runtimeClient.Client
	errorMsg   string
}

// RouteAdvertisementAdditionalOptions additional options for RouteAdvertisement object.
type RouteAdvertisementAdditionalOptions func(builder *RouteAdvertisementBuilder) (*RouteAdvertisementBuilder, error)

// NewRouteAdvertisementBuilder creates a new instance of RouteAdvertisementBuilder.
func NewRouteAdvertisementBuilder(
	apiClient *clients.Settings, name, nsname string, advertisements []ovnv1.AdvertisementType) *RouteAdvertisementBuilder {
	glog.V(100).Infof(
		"Initializing new RouteAdvertisement structure with the following params: %s, %s, %v",
		name, nsname, advertisements)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient cannot be nil")

		return nil
	}

	err := apiClient.AttachScheme(ovnv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add ovn scheme to client schemes")

		return nil
	}

	builder := &RouteAdvertisementBuilder{
		apiClient: apiClient.Client,
		Definition: &ovnv1.RouteAdvertisement{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			Spec: ovnv1.RouteAdvertisementSpec{
				Advertisements: advertisements,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the RouteAdvertisement is empty")

		builder.errorMsg = "RouteAdvertisement 'name' cannot be empty"

		return builder
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the RouteAdvertisement is empty")

		builder.errorMsg = "RouteAdvertisement 'namespace' cannot be empty"

		return builder
	}

	if len(advertisements) == 0 {
		glog.V(100).Infof("The advertisements list of the RouteAdvertisement is empty")

		builder.errorMsg = "RouteAdvertisement 'advertisements' cannot be empty"

		return builder
	}

	if len(advertisements) > 2 {
		glog.V(100).Infof("The advertisements list of the RouteAdvertisement has more than 2 items")

		builder.errorMsg = "RouteAdvertisement 'advertisements' cannot have more than 2 items"

		return builder
	}

	// Validate that advertisements are unique
	seen := make(map[ovnv1.AdvertisementType]bool)
	for _, ad := range advertisements {
		if seen[ad] {
			glog.V(100).Infof("Duplicate advertisement type found: %s", ad)

			builder.errorMsg = fmt.Sprintf("RouteAdvertisement 'advertisements' cannot contain duplicates: %s", ad)

			return builder
		}
		seen[ad] = true
	}

	return builder
}

// Get returns RouteAdvertisement object if found.
func (builder *RouteAdvertisementBuilder) Get() (*ovnv1.RouteAdvertisement, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Collecting RouteAdvertisement object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	routeAdvertisement := &ovnv1.RouteAdvertisement{}
	err := builder.apiClient.Get(context.TODO(),
		runtimeClient.ObjectKey{Name: builder.Definition.Name, Namespace: builder.Definition.Namespace},
		routeAdvertisement)

	if err != nil {
		glog.V(100).Infof(
			"RouteAdvertisement object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return routeAdvertisement, nil
}

// Exists checks whether the given RouteAdvertisement exists.
func (builder *RouteAdvertisementBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if RouteAdvertisement %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Pull loads an existing RouteAdvertisement into the Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*RouteAdvertisementBuilder, error) {
	glog.V(100).Infof("Pulling existing RouteAdvertisement name: %s namespace: %s", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is nil")

		return nil, fmt.Errorf("apiClient cannot be nil")
	}

	err := apiClient.AttachScheme(ovnv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add ovn scheme to client schemes")

		return nil, err
	}

	builder := &RouteAdvertisementBuilder{
		apiClient: apiClient.Client,
		Definition: &ovnv1.RouteAdvertisement{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the RouteAdvertisement is empty")

		return nil, fmt.Errorf("RouteAdvertisement 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the RouteAdvertisement is empty")

		return nil, fmt.Errorf("RouteAdvertisement 'namespace' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("RouteAdvertisement object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// Create makes a RouteAdvertisement in the cluster and stores the created object in struct.
func (builder *RouteAdvertisementBuilder) Create() (*RouteAdvertisementBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the RouteAdvertisement %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		err := builder.apiClient.Create(context.TODO(), builder.Definition)
		if err != nil {
			glog.V(100).Infof("Failed to create RouteAdvertisement")

			return nil, err
		}
	}

	builder.Object = builder.Definition

	return builder, nil
}

// Delete removes RouteAdvertisement object from a cluster.
func (builder *RouteAdvertisementBuilder) Delete() (*RouteAdvertisementBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the RouteAdvertisement object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof("RouteAdvertisement %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.Object = nil

		return builder, nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("can not delete RouteAdvertisement: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Update renovates the existing RouteAdvertisement object with the RouteAdvertisement definition in builder.
func (builder *RouteAdvertisementBuilder) Update(force bool) (*RouteAdvertisementBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the RouteAdvertisement object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil, fmt.Errorf("failed to update RouteAdvertisement, object does not exist on cluster")
	}

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				msg.FailToUpdateNotification("RouteAdvertisement", builder.Definition.Name, builder.Definition.Namespace))

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					msg.FailToUpdateError("RouteAdvertisement", builder.Definition.Name, builder.Definition.Namespace))

				return nil, err
			}

			return builder.Create()
		}
	}

	return builder, err
}

// WithFRRConfigurationSelector sets the FRRConfigurationSelector for the RouteAdvertisement.
func (builder *RouteAdvertisementBuilder) WithFRRConfigurationSelector(selector *metav1.LabelSelector) *RouteAdvertisementBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Setting RouteAdvertisement %s in namespace %s with FRRConfigurationSelector: %v",
		builder.Definition.Name, builder.Definition.Namespace, selector)

	builder.Definition.Spec.FRRConfigurationSelector = selector

	return builder
}

// WithOptions creates RouteAdvertisement with generic mutation options.
func (builder *RouteAdvertisementBuilder) WithOptions(options ...RouteAdvertisementAdditionalOptions) *RouteAdvertisementBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting RouteAdvertisement additional options")

	if builder.Definition == nil {
		glog.V(100).Infof("The RouteAdvertisement is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("RouteAdvertisement")

		return builder
	}

	for _, option := range options {
		if option != nil {
			builder, err := option(builder)

			if err != nil {
				glog.V(100).Infof("Error occurred in mutation function")

				builder.errorMsg = err.Error()

				return builder
			}
		}
	}

	return builder
}

// GetRouteAdvertisementGVR returns RouteAdvertisement's GroupVersionResource, which could be used for Clean function.
func GetRouteAdvertisementGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: "k8s.ovn.org", Version: "v1", Resource: "routeadvertisements",
	}
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *RouteAdvertisementBuilder) validate() (bool, error) {
	resourceCRD := "RouteAdvertisement"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf("%s", msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf("%s", builder.errorMsg)
	}

	return true, nil
}
