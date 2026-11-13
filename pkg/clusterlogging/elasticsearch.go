package clusterlogging

import (
	"context"
	"fmt"

	eskv1 "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var elasticsearchGVK = eskv1.GroupVersion.WithKind("Elasticsearch")

// ElasticsearchBuilder provides struct for the elasticsearch object.
type ElasticsearchBuilder struct {
	common.EmbeddableBuilder[eskv1.Elasticsearch, *eskv1.Elasticsearch]
	common.EmbeddableCreator[eskv1.Elasticsearch, ElasticsearchBuilder,
		*eskv1.Elasticsearch, *ElasticsearchBuilder]
	common.EmbeddableDeleter[eskv1.Elasticsearch, *eskv1.Elasticsearch]
	common.EmbeddableUpdater[eskv1.Elasticsearch, ElasticsearchBuilder,
		*eskv1.Elasticsearch, *ElasticsearchBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *ElasticsearchBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the Elasticsearch GVK for this builder.
func (builder *ElasticsearchBuilder) GetGVK() schema.GroupVersionKind {
	return elasticsearchGVK
}

// NewElasticsearchBuilder method creates new instance of builder.
func NewElasticsearchBuilder(apiClient *clients.Settings, name, nsname string) *ElasticsearchBuilder {
	return common.NewNamespacedBuilder[eskv1.Elasticsearch, ElasticsearchBuilder](
		apiClient, eskv1.AddToScheme, name, nsname)
}

// PullElasticsearch retrieves an existing elasticsearch object from the cluster.
func PullElasticsearch(apiClient *clients.Settings, name, nsname string) (*ElasticsearchBuilder, error) {
	return common.PullNamespacedBuilder[eskv1.Elasticsearch, ElasticsearchBuilder](
		context.TODO(), apiClient, eskv1.AddToScheme, name, nsname)
}

// WithManagementState sets the elasticsearch operator's management state.
func (builder *ElasticsearchBuilder) WithManagementState(
	expectedManagementState eskv1.ManagementState) *ElasticsearchBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof(
		"Setting elasticsearch %s in namespace %s with the ManagementState: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedManagementState)

	builder.Definition.Spec.ManagementState = expectedManagementState

	return builder
}

// GetManagementState fetches elasticsearch ManagementState.
func (builder *ElasticsearchBuilder) GetManagementState() (*eskv1.ManagementState, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Info("Getting elasticsearch ManagementState configuration")

	if !builder.Exists() {
		return nil, fmt.Errorf("elasticsearch object does not exist")
	}

	return &builder.Object.Spec.ManagementState, nil
}
