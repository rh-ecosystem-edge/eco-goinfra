package oauth

import (
	"context"

	oauthv1 "github.com/openshift/api/oauth/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var oauthClientGVK = oauthv1.GroupVersion.WithKind("OAuthClient")

// OAuthClientBuilder provides struct for the OAuthClient object containing connection to
// the cluster and the OAuthClient definitions.
type OAuthClientBuilder struct {
	common.EmbeddableBuilder[oauthv1.OAuthClient, *oauthv1.OAuthClient]
	common.EmbeddableCreator[
		oauthv1.OAuthClient, OAuthClientBuilder, *oauthv1.OAuthClient, *OAuthClientBuilder]
	common.EmbeddableDeleter[oauthv1.OAuthClient, *oauthv1.OAuthClient]
	common.EmbeddableUpdater[
		oauthv1.OAuthClient, OAuthClientBuilder, *oauthv1.OAuthClient, *OAuthClientBuilder]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *OAuthClientBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the OAuthClient GVK for this builder.
func (builder *OAuthClientBuilder) GetGVK() schema.GroupVersionKind {
	return oauthClientGVK
}

// PullOAuthClient loads an existing OAuthClient into Builder struct.
func PullOAuthClient(apiClient *clients.Settings, name string) (*OAuthClientBuilder, error) {
	return common.PullClusterScopedBuilder[oauthv1.OAuthClient, OAuthClientBuilder](
		context.TODO(), apiClient, oauthv1.AddToScheme, name)
}
