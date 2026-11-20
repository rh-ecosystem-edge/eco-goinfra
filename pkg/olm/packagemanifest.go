package olm

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	operatorv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/olm/package-server/operators/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

var packageManifestGVK = operatorv1.SchemeGroupVersion.WithKind(operatorv1.PackageManifestKind)

// PackageManifestBuilder provides a struct for PackageManifest object from the cluster
// and a PackageManifest definition.
type PackageManifestBuilder struct {
	common.EmbeddableBuilder[operatorv1.PackageManifest, *operatorv1.PackageManifest]
	common.EmbeddableDeleter[operatorv1.PackageManifest, *operatorv1.PackageManifest]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *PackageManifestBuilder) AttachMixins() {
	builder.SetBase(builder)
}

// GetGVK returns the PackageManifest GVK for this builder.
func (builder *PackageManifestBuilder) GetGVK() schema.GroupVersionKind {
	return packageManifestGVK
}

// PullPackageManifest loads an existing PackageManifest into Builder struct.
func PullPackageManifest(apiClient *clients.Settings, name, nsname string) (*PackageManifestBuilder, error) {
	return common.PullNamespacedBuilder[operatorv1.PackageManifest, PackageManifestBuilder](
		context.TODO(), apiClient, operatorv1.AddToScheme, name, nsname)
}

// PullPackageManifestByCatalog loads an existing PackageManifest from specified catalog into Builder struct.
func PullPackageManifestByCatalog(apiClient *clients.Settings, name, nsname,
	catalog string) (*PackageManifestBuilder, error) {
	klog.V(100).Infof("Pulling existing PackageManifest name %s in namespace %s and from catalog %s",
		name, nsname, catalog)

	if name == "" {
		klog.V(100).Info("The Name of the PackageManifest is empty")

		return nil, fmt.Errorf("packageManifest 'name' cannot be empty")
	}

	fieldSelector, err := fields.ParseSelector(fmt.Sprintf("metadata.name=%s", name))
	if err != nil {
		klog.V(100).Infof("Failed to parse invalid packageManifest name %s", name)

		return nil, err
	}

	if catalog == "" {
		klog.V(100).Info("The Catalog of the PackageManifest is empty")

		return nil, fmt.Errorf("packageManifest 'catalog' cannot be empty")
	}

	labelSelector, err := labels.Parse(fmt.Sprintf("catalog=%s", catalog))
	if err != nil {
		klog.V(100).Infof("Failed to parse invalid catalog name %s", catalog)

		return nil, err
	}

	packageManifests, err := ListPackageManifest(apiClient, nsname, runtimeClient.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		klog.V(100).Infof("Failed to list PackageManifests with name %s in namespace %s from catalog"+
			" %s due to %s", name, nsname, catalog, err.Error())

		return nil, err
	}

	if len(packageManifests) == 0 {
		klog.V(100).Info("The list of matching PackageManifests is empty")

		return nil, fmt.Errorf("no matching PackageManifests were found")
	}

	if len(packageManifests) > 1 {
		klog.V(100).Info("More than one matching PackageManifests were found")

		return nil, fmt.Errorf("more than one matching PackageManifests were found")
	}

	return packageManifests[0], nil
}
