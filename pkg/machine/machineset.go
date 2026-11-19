package machine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var machineSetGVK = machinev1beta1.GroupVersion.WithKind("MachineSet")

// SetBuilder provides a struct for MachineSet object from the cluster and a MachineSet definition.
type SetBuilder struct {
	common.EmbeddableBuilder[machinev1beta1.MachineSet, *machinev1beta1.MachineSet]
	common.EmbeddableCreator[machinev1beta1.MachineSet, SetBuilder, *machinev1beta1.MachineSet, *SetBuilder]
	common.EmbeddableDeleter[machinev1beta1.MachineSet, *machinev1beta1.MachineSet]
	// string to store the public cloud
	publicCloud string
}

// AttachMixins wires the embedded CRUD mixins to this builder instance.
func (builder *SetBuilder) AttachMixins() {
	builder.EmbeddableCreator.SetBase(builder)
	builder.EmbeddableDeleter.SetBase(builder)
}

// GetGVK returns the MachineSet GVK for this builder.
func (builder *SetBuilder) GetGVK() schema.GroupVersionKind {
	return machineSetGVK
}

const (
	// AwsCloud const definition.
	AwsCloud = "aws"
	// GcpCloud const definition.
	GcpCloud = "gcp"
	// AzureCloud const definition.
	AzureCloud = "azure"
)

// NewSetBuilderFromCopy returns an SetBuilder struct from a copied MachineSet.
func NewSetBuilderFromCopy(
	apiClient *clients.Settings,
	nsName string,
	instanceType string,
	workerLabel string,
	replicas int32) *SetBuilder {
	klog.V(100).Infof("Initializing new SetBuilder structure from copied MachineSet with the following"+
		" params: namespace: %s, instanceType: %s, workerLabel: %s, and replicas: %v", nsName, instanceType,
		workerLabel, replicas)

	builder := newSetBuilderWithDefinition(apiClient, nil)

	newSetBuilder, err := createNewWorkerMachineSetFromCopy(apiClient, nsName, instanceType, workerLabel, replicas)
	if err != nil {
		klog.V(100).Infof("Error initializing MachineSet from copy: %s", err.Error())

		builder.SetError(fmt.Errorf("error initializing MachineSet from copy: %s", err.Error()))

		return builder
	}

	builder.SetDefinition(newSetBuilder.Definition)

	err = builder.getPublicCloudKind()
	if err != nil {
		builder.SetError(fmt.Errorf("error getting the public cloud kind: %v", err.Error()))

		return builder
	}

	klog.V(100).Infof("Updating copied MachineSet provider instanceType to: %s", instanceType)

	err = builder.ChangeCloudProviderInstanceType(instanceType)
	if err != nil {
		builder.SetError(fmt.Errorf("error changing the instanceType: %v", err.Error()))

		return builder
	}

	if nsName == "" {
		klog.V(100).Info("The Namespace of the MachineSet is empty")

		builder.SetError(fmt.Errorf("MachineSet 'nsName' cannot be empty"))

		return builder
	}

	if instanceType == "" {
		klog.V(100).Info("The instanceType of the MachineSet is empty")

		builder.SetError(fmt.Errorf("MachineSet 'instanceType' cannot be empty"))

		return builder
	}

	if replicas == 0 {
		klog.V(100).Info("The replicas of the MachineSet is zero")

		builder.SetError(fmt.Errorf("MachineSet 'replicas' cannot be zero"))

		return builder
	}

	if workerLabel == "" {
		klog.V(100).Info("The workerLabel of the MachineSet is empty")

		builder.SetError(fmt.Errorf("MachineSet 'workerLabel' cannot be empty"))

		return builder
	}

	if builder.Definition == nil {
		klog.V(100).Info("The MachineSet object definition is nil")

		builder.SetError(fmt.Errorf("MachineSet 'Object.Definition' is nil"))

		return builder
	}

	return builder
}

// PullSet loads an existing MachineSet into Builder struct.
func PullSet(apiClient *clients.Settings, name, namespace string) (*SetBuilder, error) {
	klog.V(100).Infof("Pulling existing machineSet name %s in namespace %s", name, namespace)

	return common.PullNamespacedBuilder[machinev1beta1.MachineSet, SetBuilder](
		context.TODO(), apiClient, machinev1beta1.AddToScheme, name, namespace)
}

// WaitForMachineSetReady waits until MachineSet first replica is Ready.
func WaitForMachineSetReady(
	apiClient *clients.Settings,
	namespace,
	machineSetName string,
	timeout time.Duration) error {
	return wait.PollUntilContextTimeout(
		context.TODO(), 30*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			machineSetPulled, err := PullSet(apiClient, namespace, machineSetName)
			if err != nil {
				klog.V(100).Infof("MachineSet pull from cluster error: %v\n", err)

				return false, err
			}

			if machineSetPulled.Object.Status.ReadyReplicas > 0 &&
				machineSetPulled.Object.Status.Replicas == machineSetPulled.Object.Status.ReadyReplicas {
				klog.V(100).Infof("MachineSet %s has %v replicas in Ready state",
					machineSetPulled.Object.Name, machineSetPulled.Object.Status.ReadyReplicas)

				// this exits out of the wait.PollUntilContextTimeout()
				return true, nil
			}

			klog.V(100).Infof("MachineSet %s has %v replicas in Ready state",
				machineSetPulled.Object.Name, machineSetPulled.Object.Status.ReadyReplicas)

			return false, err
		})
}

// ChangeCloudProviderInstanceType calls the cloud-specific function to change the ProviderSpec instance type param.
func (builder *SetBuilder) ChangeCloudProviderInstanceType(instanceType string) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Info("Updating the cloud provider instance type field")

	switch builder.publicCloud {
	case AwsCloud:
		klog.V(100).Info("Updating ProviderSpec InstanceType param for AWS public cloud")

		err := builder.AWSChangeProviderInstanceType(instanceType)
		if err != nil {
			return fmt.Errorf("error from func AWSChangeProviderInstanceType(instanceType): %w", err)
		}

	case GcpCloud:
		klog.V(100).Infof("Updating ProviderSpec MachineType and OnHostTerminate params for " +
			"GCP public cloud")

		err := builder.GCPChangeProviderMachineType(instanceType)
		if err != nil {
			return fmt.Errorf("error from func GCPChangeProviderMachineType(instanceType): %w", err)
		}

	case AzureCloud:
		klog.V(100).Info("Updating ProviderSpec VMSize param for Azure public cloud")

		err := builder.AzureChangeProviderVMSize(instanceType)
		if err != nil {
			return fmt.Errorf("error from func AzureChangeProviderVMSize(instanceType): %w", err)
		}

	default:
		klog.V(100).Infof("Public cloud '%s' is not supported, must be 'aws', 'gcp' or 'azure'", builder.publicCloud)

		return fmt.Errorf("could not find supported public cloud")
	}

	return nil
}

// AWSChangeProviderInstanceType changes the ProviderSpec InstanceType param for AWS public cloud.
func (builder *SetBuilder) AWSChangeProviderInstanceType(instanceType string) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	if instanceType == "" {
		return fmt.Errorf("instanceType parameter cannot be empty")
	}

	byteArray, err := json.Marshal(builder.Definition.Spec.Template.Spec.ProviderSpec.Value)
	if err != nil {
		return fmt.Errorf("error marshalling machineSet providerSpec.Value into byte array: %w", err)
	}

	klog.V(100).Infof("Updating ProviderSpec InstanceType param '%s' for AWS public cloud",
		instanceType)

	var AWSProviderSpecObject *machinev1beta1.AWSMachineProviderConfig

	err = json.Unmarshal(byteArray, &AWSProviderSpecObject)
	if err != nil {
		klog.V(100).Infof("error unmarshalling byte array into AWSMachineProviderConfig object: %v", err)

		return fmt.Errorf("could not update InstanceType param: %w", err)
	}

	klog.V(100).Infof("Setting AWSMachineProviderConfig.InstanceType param value to: %s", instanceType)

	AWSProviderSpecObject.InstanceType = instanceType

	byteArrayAWS, err := json.Marshal(AWSProviderSpecObject)
	if err != nil {
		klog.V(100).Infof("error marshalling AWSMachineProviderConfig object into byte array: %v", err)

		return fmt.Errorf("could not update InstanceType param: %w", err)
	}

	err = json.Unmarshal(byteArrayAWS, builder.Definition.Spec.Template.Spec.ProviderSpec.Value)
	if err != nil {
		klog.V(100).Infof("error unmarshalling AWSMachineProviderConfig byte array into "+
			"ProviderSpec.Value object: %v", err)

		return fmt.Errorf("could not update InstanceType param: %w", err)
	}

	return nil
}

// GCPChangeProviderMachineType changes the ProviderSpec MachineType param for GCP public cloud.
func (builder *SetBuilder) GCPChangeProviderMachineType(machineType string) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	if machineType == "" {
		return fmt.Errorf("machineType parameter cannot be empty")
	}

	byteArray, err := json.Marshal(builder.Definition.Spec.Template.Spec.ProviderSpec.Value)
	if err != nil {
		return fmt.Errorf("error marshalling machineSet providerSpec.Value into byte array: %w", err)
	}

	klog.V(100).Infof("Updating ProviderSpec MachineType param '%s' for GCP public cloud", machineType)

	var GCPProviderSpecObject *machinev1beta1.GCPMachineProviderSpec

	err = json.Unmarshal(byteArray, &GCPProviderSpecObject)
	if err != nil {
		klog.V(100).Infof("error unmarshalling byte array into GCPMachineProviderSpec object: %v", err)

		return fmt.Errorf("could not update MachineType param: %w", err)
	}

	klog.V(100).Infof("Setting GCPMachineProviderConfig.MachineType param value to: %s", machineType)

	GCPProviderSpecObject.MachineType = machineType

	byteArrayGCP, err := json.Marshal(GCPProviderSpecObject)
	if err != nil {
		klog.V(100).Infof("error marshalling GCPMachineProviderSpec object into byte array: %v", err)

		return fmt.Errorf("could not update MachineType param: %w", err)
	}

	err = json.Unmarshal(byteArrayGCP, builder.Definition.Spec.Template.Spec.ProviderSpec.Value)
	if err != nil {
		klog.V(100).Infof("error unmarshalling ProviderSpec byte array into ProviderSpec.Value: %v", err)

		return fmt.Errorf("could not update MachineType param: %w", err)
	}

	return nil
}

// AzureChangeProviderVMSize changes the ProviderSpec VMSize param for Azure public cloud.
func (builder *SetBuilder) AzureChangeProviderVMSize(vmSize string) error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	if vmSize == "" {
		return fmt.Errorf("vmSize parameter cannot be empty")
	}

	byteArray, err := json.Marshal(builder.Definition.Spec.Template.Spec.ProviderSpec.Value)
	if err != nil {
		return fmt.Errorf("error marshalling machineSet providerSpec.Value into byte array: %w", err)
	}

	klog.V(100).Infof("Updating ProviderSpec Value VMSize param '%s' for Azure public cloud",
		vmSize)

	var AzureProviderSpecObject *machinev1beta1.AzureMachineProviderSpec

	err = json.Unmarshal(byteArray, &AzureProviderSpecObject)
	if err != nil {
		klog.V(100).Infof("error unmarshalling byte array into AzureMachineProviderSpec object: %v", err)

		return fmt.Errorf("could not update VMSize param: %w", err)
	}

	klog.V(100).Infof("Setting AzureMachineProviderSpec.VMSize param value to: %s", vmSize)

	AzureProviderSpecObject.VMSize = vmSize

	byteArrayAzure, err := json.Marshal(AzureProviderSpecObject)
	if err != nil {
		klog.V(100).Infof("error marshalling AzureMachineProviderSpec object into byte array: %v", err)

		return fmt.Errorf("could not update VMSize param: %w", err)
	}

	err = json.Unmarshal(byteArrayAzure, builder.Definition.Spec.Template.Spec.ProviderSpec.Value)
	if err != nil {
		klog.V(100).Infof("error unmarshalling AzureMachineProviderSpec byte array into "+
			"ProviderSpec.Value object: %v", err)

		return fmt.Errorf("could not update VMSize param: %w", err)
	}

	return nil
}

// createNewWorkerMachineSetFromCopy returns a SetBuilder object which was copied from an exiting one on cluster.
func createNewWorkerMachineSetFromCopy(
	apiClient *clients.Settings,
	namespace string,
	instanceType string,
	workerLabel string,
	replicas int32) (*SetBuilder, error) {
	workerSetBuilders, err := ListWorkerMachineSets(apiClient, namespace, workerLabel)
	if err != nil {
		return nil, fmt.Errorf("could not list worker MachineSets: %w", err)
	}

	if len(workerSetBuilders) == 0 {
		klog.V(100).Info("The array of worker MachineSets is empty")

		return nil, fmt.Errorf("no worker MachineSets were found")
	}

	// picking the first worker SetBuilder in array
	baseSetBuilder := workerSetBuilders[0]

	klog.V(100).Infof("Creating new SetBuilder copy of first existing worker MachineSet: %s",
		baseSetBuilder.Definition.Name)

	copiedSetBuilder := newSetBuilderWithDefinition(apiClient, &machinev1beta1.MachineSet{
		ObjectMeta: *baseSetBuilder.Definition.ObjectMeta.DeepCopy(),
		Spec:       *baseSetBuilder.Definition.Spec.DeepCopy(),
	})

	klog.V(100).Infof("Renaming copied SetBuilder to: %s",
		copiedSetBuilder.Definition.Name)

	// replace dots in name with dashes.  Cannot have dots or underscores in machineSet name, must also be lower case
	copiedSetBuilder.Definition.Name = fmt.Sprintf("%v-%v",
		copiedSetBuilder.Definition.Name,
		strings.ToLower(regexp.MustCompile(`[\.|\_]`).ReplaceAllString(instanceType, "-")))

	klog.V(100).Info("Updating copied MachineSet name in metadata, selector and template parameters")

	copiedSetBuilder.Definition.UID = ""
	copiedSetBuilder.Definition.ResourceVersion = ""

	// change spec labels
	copiedSetBuilder.Definition.Spec.Selector.MatchLabels["machine.openshift.io/cluster-api-machineset"] =
		copiedSetBuilder.Definition.Name
	copiedSetBuilder.Definition.Spec.Template.Labels["machine.openshift.io/cluster-api-machineset"] =
		copiedSetBuilder.Definition.Name

	klog.V(100).Infof("Updating copied MachineSet replicas value to: %v", replicas)
	copiedSetBuilder.Definition.Spec.Replicas = &replicas

	return copiedSetBuilder, nil
}

// getPublicCloudKind determines the public cloud kind and stores it in the builder struct.
func (builder *SetBuilder) getPublicCloudKind() error {
	if err := common.Validate(builder); err != nil {
		return err
	}

	klog.V(100).Info("Determining the public cloud kind")

	providerSpecMap := make(map[string]interface{})

	// Value field is of type *runtime.RawExtension
	byteArray, err := json.Marshal(builder.Definition.Spec.Template.Spec.ProviderSpec.Value)
	if err != nil {
		builder.SetError(fmt.Errorf("error determining public cloud kind: %w", err))

		return fmt.Errorf("error marshalling the providerSpec Value element into a byte array")
	}

	err = json.Unmarshal(byteArray, &providerSpecMap)
	if err != nil {
		builder.SetError(fmt.Errorf("error determining public cloud kind: %w", err))

		return fmt.Errorf("error unmarshalling the byte array into a providerSpec map")
	}

	// publicCloudKind is of type: map[string]interface{}
	publicCloudKind := providerSpecMap["kind"]

	publicCloud, ok := publicCloudKind.(string)

	if !ok {
		return fmt.Errorf("failed to detect public cloud kind")
	}

	klog.V(100).Infof("ProviderSpec kind param is '%s'", publicCloud)

	switch publicCloud {
	case "AWSMachineProviderConfig":
		builder.publicCloud = AwsCloud
	case "GCPMachineProviderSpec":
		builder.publicCloud = GcpCloud
	case "AzureMachineProviderSpec":
		builder.publicCloud = AzureCloud
	default:
		builder.SetError(fmt.Errorf("unsupported cloud platform. Supported public cloud are AWS, GCP, and Azure"))

		return fmt.Errorf("unsupported cloud platform. Supported public cloud are AWS, GCP, and Azure")
	}

	return nil
}

func newSetBuilderWithDefinition(apiClient *clients.Settings, definition *machinev1beta1.MachineSet) *SetBuilder {
	builder := &SetBuilder{}
	builder.AttachMixins()
	builder.SetGVK(builder.GetGVK())
	builder.SetClient(apiClient)
	builder.SetDefinition(definition)

	return builder
}
