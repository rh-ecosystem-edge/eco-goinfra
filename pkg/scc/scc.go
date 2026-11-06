package scc

import (
	"context"
	"fmt"

	securityV1 "github.com/openshift/api/security/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	errEmptyRunAsUser      = "securityContextConstraints 'runAsUser' cannot be empty"
	errEmptySELinuxContext = "securityContextConstraints 'selinuxContext' cannot be empty"
)

const redefiningMsg = "Redefining SecurityContextConstraints"

// Builder provides struct for SecurityContextConstraints object containing connection
// to the cluster SecurityContextConstraints definition.
type Builder struct {
	common.EmbeddableBuilder[securityV1.SecurityContextConstraints, *securityV1.SecurityContextConstraints]
	common.EmbeddableCreator[securityV1.SecurityContextConstraints, Builder, *securityV1.SecurityContextConstraints, *Builder]
	common.EmbeddableDeleter[securityV1.SecurityContextConstraints, *securityV1.SecurityContextConstraints]
	common.EmbeddableUpdater[securityV1.SecurityContextConstraints, Builder, *securityV1.SecurityContextConstraints, *Builder]
}

// SecurityContextConstraintsAdditionalOptions additional options for SecurityContextConstraints object.
type SecurityContextConstraintsAdditionalOptions func(builder *Builder) (*Builder, error)

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *Builder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
	builder.EmbeddableUpdater.SetBase(builder)
}

// GetGVK returns the SecurityContextConstraints GVK for this builder.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
	return securityV1.SchemeGroupVersion.WithKind("SecurityContextConstraints")
}

// NewBuilder creates new instance of Builder.
func NewBuilder(apiClient *clients.Settings, name, runAsUser, selinuxContext string) *Builder {
	klog.V(100).Infof(
		"Initializing new SecurityContextConstraints structure with the following params: "+
			"name: %s, runAsUser type: %s, selinuxContext type: %s", name, runAsUser, selinuxContext)

	builder := common.NewClusterScopedBuilder[securityV1.SecurityContextConstraints, Builder](
		apiClient, securityV1.Install, name)
	if builder.GetError() != nil {
		return builder
	}

	builder.Definition.RunAsUser = securityV1.RunAsUserStrategyOptions{
		Type: securityV1.RunAsUserStrategyType(runAsUser),
	}
	builder.Definition.SELinuxContext = securityV1.SELinuxContextStrategyOptions{
		Type: securityV1.SELinuxContextStrategyType(selinuxContext),
	}

	if runAsUser == "" {
		klog.V(100).Info("The runAsUser of the SecurityContextConstraints is empty")

		builder.SetError(fmt.Errorf(errEmptyRunAsUser))

		return builder
	}

	if selinuxContext == "" {
		klog.V(100).Info("The selinuxContext of the SecurityContextConstraints is empty")

		builder.SetError(fmt.Errorf(errEmptySELinuxContext))

		return builder
	}

	return builder
}

// Pull pulls existing SecurityContextConstraints from cluster.
func Pull(apiClient *clients.Settings, name string) (*Builder, error) {
	return common.PullClusterScopedBuilder[securityV1.SecurityContextConstraints, Builder](
		context.TODO(), apiClient, securityV1.Install, name)
}

// WithPrivilegedContainer adds bool flag to the allowPrivilegedContainer of SecurityContextConstraints.
func (builder *Builder) WithPrivilegedContainer(allowPrivileged bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with AllowPrivilegedContainer: %t flag",
		redefiningMsg, builder.Definition.Name, allowPrivileged)

	builder.Definition.AllowPrivilegedContainer = allowPrivileged

	return builder
}

// WithPrivilegedEscalation adds bool flag to the allowPrivilegeEscalation of SecurityContextConstraints.
func (builder *Builder) WithPrivilegedEscalation(allowPrivilegedEscalation bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with allowPrivilegedEscalation: %t flag",
		redefiningMsg, builder.Definition.Name, allowPrivilegedEscalation)

	builder.Definition.DefaultAllowPrivilegeEscalation = &allowPrivilegedEscalation

	return builder
}

// WithHostDirVolumePlugin adds bool flag to the allowHostDirVolumePlugin of SecurityContextConstraints.
func (builder *Builder) WithHostDirVolumePlugin(allowPlugin bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with allowHostDirVolumePlugin: %t flag",
		redefiningMsg, builder.Definition.Name, allowPlugin)

	builder.Definition.AllowHostDirVolumePlugin = allowPlugin

	return builder
}

// WithHostIPC adds bool flag to the allowHostIPC of SecurityContextConstraints.
func (builder *Builder) WithHostIPC(allowHostIPC bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with allowHostIPC: %t flag",
		redefiningMsg, builder.Definition.Name, allowHostIPC)

	builder.Definition.AllowHostIPC = allowHostIPC

	return builder
}

// WithHostNetwork adds bool flag to the allowHostNetwork of SecurityContextConstraints.
func (builder *Builder) WithHostNetwork(allowHostNetwork bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with allowHostNetwork: %t flag",
		redefiningMsg, builder.Definition.Name, allowHostNetwork)

	builder.Definition.AllowHostNetwork = allowHostNetwork

	return builder
}

// WithHostPID adds bool flag to the allowHostPID of SecurityContextConstraints.
func (builder *Builder) WithHostPID(allowHostPID bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with allowHostPID: %t flag", redefiningMsg, builder.Definition.Name, allowHostPID)

	builder.Definition.AllowHostPID = allowHostPID

	return builder
}

// WithHostPorts adds bool flag to the allowHostPorts of SecurityContextConstraints.
func (builder *Builder) WithHostPorts(allowHostPorts bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with allowHostPorts: %t flag",
		redefiningMsg, builder.Definition.Name, allowHostPorts)

	builder.Definition.AllowHostPorts = allowHostPorts

	return builder
}

// WithReadOnlyRootFilesystem adds bool flag to the readOnlyRootFilesystem of SecurityContextConstraints.
func (builder *Builder) WithReadOnlyRootFilesystem(readOnlyRootFilesystem bool) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with readOnlyRootFilesystem: %t flag",
		redefiningMsg, builder.Definition.Name, readOnlyRootFilesystem)

	builder.Definition.ReadOnlyRootFilesystem = readOnlyRootFilesystem

	return builder
}

// WithDropCapabilities adds list of drop capabilities to SecurityContextConstraints.
func (builder *Builder) WithDropCapabilities(requiredDropCapabilities []corev1.Capability) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with requiredDropCapabilities: %v",
		redefiningMsg, builder.Definition.Name, requiredDropCapabilities)

	if len(requiredDropCapabilities) == 0 {
		klog.V(100).Info("SecurityContextConstraints 'requiredDropCapabilities' argument cannot be empty")

		builder.SetError(fmt.Errorf("securityContextConstraints 'requiredDropCapabilities' cannot be empty list"))

		return builder
	}

	builder.Definition.RequiredDropCapabilities = append(
		builder.Definition.RequiredDropCapabilities, requiredDropCapabilities...)

	return builder
}

// WithAllowCapabilities adds list of allow capabilities to SecurityContextConstraints.
func (builder *Builder) WithAllowCapabilities(allowCapabilities []corev1.Capability) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with allowCapabilities: %v",
		redefiningMsg, builder.Definition.Name, allowCapabilities)

	if len(allowCapabilities) == 0 {
		klog.V(100).Info("SecurityContextConstraints 'allowCapabilities' argument cannot be empty")

		builder.SetError(fmt.Errorf("securityContextConstraints 'allowCapabilities' cannot be empty list"))

		return builder
	}

	builder.Definition.AllowedCapabilities = append(builder.Definition.AllowedCapabilities, allowCapabilities...)

	return builder
}

// WithDefaultAddCapabilities adds list of defaultAddCapabilities to SecurityContextConstraints.
func (builder *Builder) WithDefaultAddCapabilities(defaultAddCapabilities []corev1.Capability) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with defaultAddCapabilities: %v",
		redefiningMsg, builder.Definition.Name, defaultAddCapabilities)

	if len(defaultAddCapabilities) == 0 {
		klog.V(100).Info("SecurityContextConstraints 'defaultAddCapabilities' argument cannot be empty")

		builder.SetError(fmt.Errorf("securityContextConstraints 'defaultAddCapabilities' cannot be empty list"))

		return builder
	}

	builder.Definition.DefaultAddCapabilities = append(builder.Definition.DefaultAddCapabilities,
		defaultAddCapabilities...)

	return builder
}

// WithPriority adds priority to SecurityContextConstraints.
func (builder *Builder) WithPriority(priority *int32) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with priority: %v", redefiningMsg, builder.Definition.Name, priority)

	builder.Definition.Priority = priority

	return builder
}

// WithFSGroup adds fsGroup to SecurityContextConstraints.
func (builder *Builder) WithFSGroup(fsGroup string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with fsGroup: %s", redefiningMsg, builder.Definition.Name, fsGroup)

	if fsGroup == "" {
		klog.V(100).Info("SecurityContextConstraints 'fsGroup' argument cannot be empty")

		builder.SetError(fmt.Errorf("securityContextConstraints 'fsGroup' cannot be empty string"))

		return builder
	}

	builder.Definition.FSGroup.Type = securityV1.FSGroupStrategyType(fsGroup)

	return builder
}

// WithFSGroupRange adds fsGroupRange to SecurityContextConstraints.
func (builder *Builder) WithFSGroupRange(fsGroupMin, fsGroupMax int64) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with fsGroupRange: fsGroupMin: %d, fsGroupMax: %d",
		redefiningMsg, builder.Definition.Name, fsGroupMin, fsGroupMax)

	if fsGroupMin > fsGroupMax {
		klog.V(100).Info("SecurityContextConstraints 'fsGroupMin' argument can not be greater than fsGroupMax")

		builder.SetError(fmt.Errorf(
			"securityContextConstraints 'fsGroupMin' argument can not be greater than fsGroupMax"))

		return builder
	}

	builder.Definition.FSGroup.Ranges = append(
		builder.Definition.FSGroup.Ranges, securityV1.IDRange{Min: fsGroupMin, Max: fsGroupMax})

	return builder
}

// WithGroups adds groups to SecurityContextConstraints.
func (builder *Builder) WithGroups(groups []string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with groups: %v", redefiningMsg, builder.Definition.Name, groups)

	if len(groups) == 0 {
		klog.V(100).Info("SecurityContextConstraints 'groups' argument cannot be empty")

		builder.SetError(fmt.Errorf("securityContextConstraints 'fsGroupType' cannot be empty string"))

		return builder
	}

	builder.Definition.Groups = append(builder.Definition.Groups, groups...)

	return builder
}

// WithSeccompProfiles adds list of seccompProfiles to SecurityContextConstraints.
func (builder *Builder) WithSeccompProfiles(seccompProfiles []string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with SeccompProfiles: %v",
		redefiningMsg, builder.Definition.Name, seccompProfiles)

	if len(seccompProfiles) == 0 {
		klog.V(100).Info("SecurityContextConstraints 'seccompProfiles' argument cannot be empty")

		builder.SetError(fmt.Errorf("securityContextConstraints 'seccompProfiles' cannot be empty list"))

		return builder
	}

	builder.Definition.SeccompProfiles = append(builder.Definition.SeccompProfiles, seccompProfiles...)

	return builder
}

// WithSupplementalGroups adds SupplementalGroups to SecurityContextConstraints.
func (builder *Builder) WithSupplementalGroups(supplementalGroupsType string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with supplementalGroupsType: %s",
		redefiningMsg, builder.Definition.Name, supplementalGroupsType)

	if supplementalGroupsType == "" {
		klog.V(100).Info("SecurityContextConstraints 'SupplementalGroups' argument cannot be empty")

		builder.SetError(fmt.Errorf("securityContextConstraints 'SupplementalGroups' cannot be empty string"))

		return builder
	}

	builder.Definition.SupplementalGroups.Type = securityV1.SupplementalGroupsStrategyType(supplementalGroupsType)

	return builder
}

// WithUsers adds users to SecurityContextConstraints.
func (builder *Builder) WithUsers(users []string) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with users: %v", redefiningMsg, builder.Definition.Name, users)

	if len(users) == 0 {
		klog.V(100).Info("SecurityContextConstraints 'users' argument cannot be empty")

		builder.SetError(fmt.Errorf("securityContextConstraints 'users' cannot be empty list"))

		return builder
	}

	builder.Definition.Users = append(builder.Definition.Users, users...)

	return builder
}

// WithVolumes adds list of volumes to SecurityContextConstraints.
func (builder *Builder) WithVolumes(volumes []securityV1.FSType) *Builder {
	if err := common.Validate(builder); err != nil {
		return builder
	}

	klog.V(100).Infof("%s %s with volumes: %v", redefiningMsg, builder.Definition.Name, volumes)

	if len(volumes) == 0 {
		klog.V(100).Info("SecurityContextConstraints 'volumes' argument cannot be empty")

		builder.SetError(fmt.Errorf("securityContextConstraints 'volumes' cannot be empty list"))

		return builder
	}

	builder.Definition.Volumes = append(builder.Definition.Volumes, volumes...)

	return builder
}
