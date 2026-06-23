package namespace

import (
	"context"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List returns namespace inventory.
func List(apiClient *clients.Settings, options ...metav1.ListOptions) ([]*Builder, error) {
	convertedOptions, err := common.ConvertMetaListOptionsToListOptions(options)
	if err != nil {
		return nil, err
	}

	return common.List[corev1.Namespace, corev1.NamespaceList, Builder](
		context.TODO(), apiClient, corev1.AddToScheme, convertedOptions...)
}
