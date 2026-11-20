package oran

import (
	"context"
	"fmt"
	"strings"
	"time"

	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

// ClusterTemplateBuilder provides a struct to inferface with ClusterTemplate resources on a specific cluster.
type ClusterTemplateBuilder struct {
	common.EmbeddableBuilder[provisioningv1alpha1.ClusterTemplate, *provisioningv1alpha1.ClusterTemplate]
}

// GetGVK returns the ClusterTemplate GVK for this builder.
func (builder *ClusterTemplateBuilder) GetGVK() schema.GroupVersionKind {
	return provisioningv1alpha1.GroupVersion.WithKind("ClusterTemplate")
}

// PullClusterTemplate pulls an existing ClusterTemplate into a ClusterTemplateBuilder struct.
func PullClusterTemplate(apiClient *clients.Settings, name, nsname string) (*ClusterTemplateBuilder, error) {
	return common.PullNamespacedBuilder[provisioningv1alpha1.ClusterTemplate, ClusterTemplateBuilder](
		context.TODO(), apiClient, provisioningv1alpha1.AddToScheme, name, nsname)
}

// WaitForCondition waits up to the provided timeout for a condition matching expected. It checks only the Type, Status,
// Reason, and Message fields. For the message, it matches if the message contains the expected. Zero fields in the
// expected condition are ignored.
func (builder *ClusterTemplateBuilder) WaitForCondition(
	expected metav1.Condition, timeout time.Duration) (*ClusterTemplateBuilder, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Waiting up to %s until ClusterTemplate %s in namespace %s has condition %v",
		timeout, builder.Definition.Name, builder.Definition.Namespace, expected)

	object, err := builder.Get()
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(100).Infof("ClusterTemplate %s does not exist in namespace %s",
				builder.Definition.Name, builder.Definition.Namespace)

			return nil, fmt.Errorf("cannot wait for non-existent ClusterTemplate")
		}

		return nil, err
	}

	builder.Object = object

	err = wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			var err error

			builder.Object, err = builder.Get()
			if err != nil {
				klog.V(100).Infof("Failed to get ClusterTemplate %s in namespace %s: %v",
					builder.Definition.Name, builder.Definition.Namespace, err)

				return false, nil
			}

			builder.Definition = builder.Object

			for _, condition := range builder.Object.Status.Conditions {
				if expected.Type != "" && condition.Type != expected.Type {
					continue
				}

				if expected.Status != "" && condition.Status != expected.Status {
					continue
				}

				if expected.Reason != "" && condition.Reason != expected.Reason {
					continue
				}

				if expected.Message != "" && !strings.Contains(condition.Message, expected.Message) {
					continue
				}

				return true, nil
			}

			return false, nil
		})
	if err != nil {
		return nil, err
	}

	return builder, nil
}
