package servicemesh

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	istiov1 "maistra.io/api/core/v1"
)

var memberRollGVK = istiov1.SchemeGroupVersion.WithKind("ServiceMeshMemberRoll")

// MemberRollBuilder provides a struct for serviceMeshMemberRoll object from the cluster and
// a serviceMeshMemberRoll definition.
type MemberRollBuilder struct {
	common.EmbeddableBuilder[istiov1.ServiceMeshMemberRoll, *istiov1.ServiceMeshMemberRoll]
	common.EmbeddableCreator[
		istiov1.ServiceMeshMemberRoll,
		MemberRollBuilder,
		*istiov1.ServiceMeshMemberRoll,
		*MemberRollBuilder,
	]
	common.EmbeddableDeleter[istiov1.ServiceMeshMemberRoll, *istiov1.ServiceMeshMemberRoll]
	common.EmbeddableForceUpdater[
		istiov1.ServiceMeshMemberRoll,
		MemberRollBuilder,
		*istiov1.ServiceMeshMemberRoll,
		*MemberRollBuilder,
	]
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *MemberRollBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableForceUpdater.SetBase(builder)
}

// GetGVK returns the ServiceMeshMemberRoll GVK for this builder.
func (builder *MemberRollBuilder) GetGVK() schema.GroupVersionKind {
	return memberRollGVK
}

// NewMemberRollBuilder method creates new instance of builder.
func NewMemberRollBuilder(apiClient *clients.Settings, name, nsname string) *MemberRollBuilder {
	klog.V(100).Infof("Initializing new serviceMeshMemberRollBuilder structure with the following "+
		"params: name: %s, namespace: %s", name, nsname)

	return common.NewNamespacedBuilder[istiov1.ServiceMeshMemberRoll, MemberRollBuilder](
		apiClient, istiov1.AddToScheme, name, nsname)
}

// PullMemberRoll retrieves an existing serviceMeshMemberRoll object from the cluster.
func PullMemberRoll(apiClient *clients.Settings, name, nsname string) (*MemberRollBuilder, error) {
	klog.V(100).Infof(
		"Pulling serviceMeshMemberRoll object name: %s in namespace: %s", name, nsname)

	return common.PullNamespacedBuilder[istiov1.ServiceMeshMemberRoll, MemberRollBuilder](
		context.TODO(), apiClient, istiov1.AddToScheme, name, nsname)
}

// WithMembersList adds member list section to the MemberRollBuilder.
func (builder *MemberRollBuilder) WithMembersList(membersList []string) *MemberRollBuilder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("Adding member list %v section to the MemberRollBuilder", membersList)

	if len(membersList) == 0 {
		klog.V(100).Info("Cannot add empty membersList to the memberRoll structure")

		builder.SetError(fmt.Errorf("can not modify memberRoll config with empty membersList"))

		return builder
	}

	builder.Definition.Spec.Members = append(builder.Definition.Spec.Members, membersList...)

	return builder
}

// GetMembersList fetches memberRoll's membersList.
func (builder *MemberRollBuilder) GetMembersList() (*[]string, error) {
	if err := common.Validate(builder); err != nil {
		return nil, err
	}

	klog.V(100).Infof("Getting memberRoll %s in namespace %s membersList configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil, fmt.Errorf("memberRoll object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return &builder.Object.Spec.Members, nil
}

// IsReady check if the ServiceMesh MemberRoll is Ready.
func (builder *MemberRollBuilder) IsReady(timeout time.Duration) (bool, error) {
	if err := common.Validate(builder); err != nil {
		return false, err
	}

	err := wait.PollUntilContextTimeout(
		context.TODO(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			if !builder.Exists() {
				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Type == istiov1.ConditionTypeMemberRollReady {
					if condition.Status == corev1.ConditionTrue {
						return true, nil
					}
				}
			}

			return false, nil
		})
	if err != nil {
		return false, fmt.Errorf("the Ready condition did not reached for the Service Mesh MemberRoll %s in "+
			"namespace %s during %v; %v", builder.Definition.Name, builder.Definition.Namespace, timeout, err)
	}

	return true, nil
}
