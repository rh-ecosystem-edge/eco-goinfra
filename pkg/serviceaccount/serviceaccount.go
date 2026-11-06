package serviceaccount

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	corev1Typed "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

var serviceAccountGVK = corev1.SchemeGroupVersion.WithKind("ServiceAccount")

// Builder provides struct for serviceaccount object containing connection to the cluster and the
// serviceaccount definitions.
type Builder struct {
	common.EmbeddableBuilder[corev1.ServiceAccount, *corev1.ServiceAccount]
	common.EmbeddableWithOptions[corev1.ServiceAccount, Builder, *corev1.ServiceAccount, *Builder, AdditionalOptions]
	common.EmbeddableCreator[corev1.ServiceAccount, Builder, *corev1.ServiceAccount, *Builder]
	common.EmbeddableDeleter[corev1.ServiceAccount, *corev1.ServiceAccount]

	// serviceAccountClient is retained for the token subresource, which is not exposed by the common runtime client.
	serviceAccountClient corev1Typed.ServiceAccountInterface
}

// AdditionalOptions additional options for ServiceAccount object.
type AdditionalOptions func(builder *Builder) (*Builder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableWithOptions.SetBase(builder)
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the ServiceAccount GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return serviceAccountGVK
}

// NewBuilder creates a new instance of Builder.
func NewBuilder(apiClient *clients.Settings, name, nsname string) *Builder {
	klog.V(100).Infof("Initializing new serviceaccount structure with the following params: %s, %s", name, nsname)

	builder := common.NewNamespacedBuilder[corev1.ServiceAccount, Builder](
		apiClient, corev1.AddToScheme, name, nsname)
	if apiClient != nil {
		builder.serviceAccountClient = apiClient.ServiceAccounts(nsname)
	}

	return builder
}

// Pull loads an existing serviceaccount into Builder struct.
func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
	klog.V(100).Infof("Pulling existing serviceaccount name: %s under namespace: %s", name, nsname)

	builder, err := common.PullNamespacedBuilder[corev1.ServiceAccount, Builder](
		context.TODO(), apiClient, corev1.AddToScheme, name, nsname)
	if err != nil {
		return nil, err
	}

	builder.serviceAccountClient = apiClient.ServiceAccounts(nsname)

	return builder, err
}

// CreateToken creates a new token for the builder's service account using the provided duration and audiences. The zero
// values of duration and audiences are both allowed. Note that the duration of the token returned is not guaranteed to
// match the requested duration. Its expiration will be logged, however.
func (builder *Builder) CreateToken(duration time.Duration, audiences ...string) (string, error) {
	if err := common.Validate(builder); err != nil {
		return "", err
	}

	klog.V(100).Infof("Creating token for serviceaccount %s in namespace %s with duration %s",
		builder.Definition.Name, builder.Definition.Namespace, duration)

	if !builder.Exists() {
		klog.V(100).Infof("Cannot create a token for serviceaccount %s in namespace %s because it does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		return "", fmt.Errorf("serviceaccount %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	var durationSeconds *int64
	if floatDuration := duration.Round(time.Second).Seconds(); floatDuration > 0 {
		durationSeconds = ptr.To(int64(floatDuration))
	}

	tokenRequest := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences:         audiences,
			ExpirationSeconds: durationSeconds,
		},
	}

	if builder.serviceAccountClient == nil {
		return "", fmt.Errorf("serviceaccount typed client is nil")
	}

	tokenRequest, err := builder.serviceAccountClient.CreateToken(
		logging.DiscardContext(), builder.Definition.Name, tokenRequest, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}

	klog.V(100).Infof("Successfully created token for serviceaccount %s in namespace %s with expiration %s",
		builder.Definition.Name, builder.Definition.Namespace, tokenRequest.Status.ExpirationTimestamp.Format(time.RFC3339))

	return tokenRequest.Status.Token, nil
}

// GetGVR returns service's GroupVersionResource which could be used for Clean function.
func GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccounts"}
}
