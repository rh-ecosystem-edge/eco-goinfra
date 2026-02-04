package lca

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	lcaipcv1 "github.com/openshift-kni/lifecycle-agent/api/ipconfig/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/logging"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ipConfigName = "ipconfig"
)

// IPConfigBuilder provides struct for the ipconfig object containing connection to
// the cluster and the ipconfig definitions.
type IPConfigBuilder struct {
	// IPConfig definition. Used to store the ipconfig object.
	Definition *lcaipcv1.IPConfig
	// Created ipconfig object.
	Object *lcaipcv1.IPConfig
	// Used in functions that define or mutate the ipconfig definition.
	// errorMsg is processed before the ipconfig object is created
	errorMsg  string
	apiClient goclient.Client
}

// IPConfigAdditionalOptions additional options for ipconfig object.
type IPConfigAdditionalOptions func(builder *IPConfigBuilder) (*IPConfigBuilder, error)

// NewIPConfigBuilder creates a new instance of IPConfig.
func NewIPConfigBuilder(
	apiClient *clients.Settings,
	name string,
) *IPConfigBuilder {
	if apiClient == nil {
		klog.V(100).Info("The apiClient cannot be nil")

		return nil
	}

	err := apiClient.AttachScheme(lcaipcv1.AddToScheme)
	if err != nil {
		klog.V(100).Info("Failed to add lcaipc v1 scheme to client schemes")

		return nil
	}

	builder := &IPConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &lcaipcv1.IPConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name != ipConfigName {
		klog.V(100).Infof("The name of the ipconfig must be " + ipConfigName)

		builder.errorMsg = "IPConfig name must be " + ipConfigName

		return builder
	}

	return builder
}

// WithOptions creates ipconfig with generic mutation options.
func (builder *IPConfigBuilder) WithOptions(options ...IPConfigAdditionalOptions) *IPConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	klog.V(100).Info("Setting ipconfig additional options")

	for _, option := range options {
		if option != nil {
			builder, err := option(builder)
			if err != nil {
				klog.V(100).Info("Error occurred in mutation function")

				builder.errorMsg = err.Error()

				return builder
			}
		}
	}

	return builder
}

// Create makes an ipconfig in the cluster and stores the created object in struct.
func (builder *IPConfigBuilder) Create() (*IPConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	klog.V(100).Infof("Creating the ipconfig %s",
		builder.Definition.Name)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(logging.DiscardContext(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// PullIPConfig pulls existing ipconfig from cluster.
func PullIPConfig(apiClient *clients.Settings, name string) (*IPConfigBuilder, error) {
	klog.V(100).Infof("Pulling existing ipconfig name %s from cluster", name)

	if apiClient == nil {
		klog.V(100).Info("The apiClient cannot be nil")

		return nil, fmt.Errorf("the apiClient is nil")
	}

	err := apiClient.AttachScheme(lcaipcv1.AddToScheme)
	if err != nil {
		klog.V(100).Info("Failed to add lcaipc v1 scheme to client schemes")

		return nil, err
	}

	builder := &IPConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &lcaipcv1.IPConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		klog.V(100).Info("The name of the ipconfig is empty")

		return nil, fmt.Errorf("ipconfig 'name' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("ipconfig object %s does not exist", name)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// Delete removes the existing ipconfig from a cluster.
func (builder *IPConfigBuilder) Delete() (*IPConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	klog.V(100).Infof("Deleting the ipconfig %s",
		builder.Definition.Name)

	if !builder.Exists() {
		builder.Object = nil

		return builder, nil
	}

	err := builder.apiClient.Delete(logging.DiscardContext(), builder.Definition)
	if err != nil {
		return builder, fmt.Errorf("can not delete ipconfig: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Get returns ipconfig object if found.
func (builder *IPConfigBuilder) Get() (*lcaipcv1.IPConfig, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	klog.V(100).Infof("Getting ipconfig %s",
		builder.Definition.Name)

	ipconfig := &lcaipcv1.IPConfig{}

	err := builder.apiClient.Get(logging.DiscardContext(), goclient.ObjectKey{
		Name: builder.Definition.Name,
	}, ipconfig)
	if err != nil {
		return nil, err
	}

	return ipconfig, nil
}

// Exists checks whether the given ipconfig exists.
func (builder *IPConfigBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	klog.V(100).Infof("Checking if ipconfig %s exists",
		builder.Definition.Name)

	var err error

	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// WaitUntilComplete waits the specified timeout for the ipconfig to complete
// actions.
func (builder *IPConfigBuilder) WaitUntilComplete(timeout time.Duration) (*IPConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	klog.V(100).Infof("Waiting for ipconfig %s to complete actions",
		builder.Definition.Name)

	if !builder.Exists() {
		klog.V(100).Info("The ipconfig does not exist on the cluster")

		return builder, fmt.Errorf("%s", builder.errorMsg)
	}

	// Polls periodically to determine if ipconfig is in desired state.
	var err error

	err = wait.PollUntilContextTimeout(
		context.TODO(), time.Second*3, timeout, true, func(ctx context.Context) (bool, error) {
			builder.Object, err = builder.Get()
			if err != nil {
				return false, nil
			}

			for _, condition := range builder.Object.Status.Conditions {
				if condition.Status == "True" && condition.Type == "ConfigCompleted" &&
					condition.Reason == "Completed" {
					return true, nil
				}
			}

			return false, nil
		})
	if err == nil {
		return builder, nil
	}

	return nil, err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *IPConfigBuilder) validate() (bool, error) {
	resourceCRD := "IPConfig"

	if builder == nil {
		klog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		klog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf("%s", msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		klog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		klog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf("%s", builder.errorMsg)
	}

	return true, nil
}

// WithIPv4Address sets the IPv4 address used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv4Address(
	ipv4Address string) *IPConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	ipAddress := net.ParseIP(ipv4Address)

	if ipAddress == nil {
		klog.V(100).Infof("Invalid IP %s", ipv4Address)

		builder.errorMsg = fmt.Sprintf("invalid IP argument %s", ipv4Address)

		return builder
	}

	klog.V(100).Infof("Setting IPv4 %s in ipconfig", ipv4Address)

	builder.Definition.Spec.IPv4.Address = ipv4Address

	return builder
}

// WithIPv6Address sets the IPv6 address used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv6Address(
	ipv6Address string) *IPConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	ipAddress := net.ParseIP(ipv6Address)

	if ipAddress == nil {
		klog.V(100).Infof("Invalid IP %s", ipv6Address)

		builder.errorMsg = fmt.Sprintf("invalid IP argument %s", ipv6Address)

		return builder
	}

	klog.V(100).Infof("Setting IPv6 %s in ipconfig", ipv6Address)

	builder.Definition.Spec.IPv6.Address = ipv6Address

	return builder
}

// WithIPv6Gateway sets the IPv6 gateway address used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv6Gateway(
	ipv6Address string) *IPConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	gwIPAddress := net.ParseIP(ipv6Address)

	if gwIPAddress == nil {
		klog.V(100).Infof("Invalid IP %s", ipv6Address)

		builder.errorMsg = fmt.Sprintf("invalid IP argument %s", ipv6Address)

		return builder
	}

	klog.V(100).Infof("Setting IPv6 gateway %s in ipconfig", ipv6Address)

	builder.Definition.Spec.IPv6.Gateway = ipv6Address

	return builder
}

// WithIPv4Gateway sets the IPv4 gateway address used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv4Gateway(
	ipv4Address string) *IPConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	gwIPAddress := net.ParseIP(ipv4Address)

	if gwIPAddress == nil {
		klog.V(100).Infof("Invalid IP %s", ipv4Address)

		builder.errorMsg = fmt.Sprintf("invalid IP argument %s", ipv4Address)

		return builder
	}

	klog.V(100).Infof("Setting IPv4 gateway %s in ipconfig", ipv4Address)

	builder.Definition.Spec.IPv4.Gateway = ipv4Address

	return builder
}

// WithIPv4MachineNetwork sets the IPv4 machine network used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv4MachineNetwork(
	ipv4MachineNetwork string) *IPConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	_, _, err := net.ParseCIDR(ipv4MachineNetwork)
	if err != nil {
		klog.V(100).Infof("Invalid CIDR %s", ipv4MachineNetwork)

		builder.errorMsg = fmt.Sprintf("invalid CIDR argument %s", ipv4MachineNetwork)

		return builder
	}

	klog.V(100).Infof("Setting IPv4 machine network %s in ipconfig", ipv4MachineNetwork)

	builder.Definition.Spec.IPv4.MachineNetwork = ipv4MachineNetwork

	return builder
}

// WithIPv6MachineNetwork sets the IPv6 machine network used by the ipconfig.
func (builder *IPConfigBuilder) WithIPv6MachineNetwork(
	ipv6MachineNetwork string) *IPConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	_, _, err := net.ParseCIDR(ipv6MachineNetwork)
	if err != nil {
		klog.V(100).Infof("Invalid CIDR %s", ipv6MachineNetwork)

		builder.errorMsg = fmt.Sprintf("invalid CIDR argument %s", ipv6MachineNetwork)

		return builder
	}

	klog.V(100).Infof("Setting IPv6 machine network %s in ipconfig", ipv6MachineNetwork)

	builder.Definition.Spec.IPv6.MachineNetwork = ipv6MachineNetwork

	return builder
}

// WithStage sets the stage used by the ipconfig.
func (builder *IPConfigBuilder) WithStage(
	stage string) *IPConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	klog.V(100).Infof("Setting stage %s in ipconfig", stage)
	builder.Definition.Spec.Stage = lcaipcv1.IPConfigStage(stage)

	return builder
}

// WithDNS sets the list of DNS servers used by the ipconfig.
func (builder *IPConfigBuilder) WithDNS(
	dnsServers []string) *IPConfigBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	klog.V(100).Infof("Setting DNS servers %s in ipconfig", dnsServers)

	if len(dnsServers) == 0 {
		klog.V(100).Infof("The DNS servers list is empty")

		builder.errorMsg = "The DNS servers list cannot be empty"

		return builder
	}

	for _, dnsServer := range dnsServers {
		dnsServerTrimmed := strings.TrimSpace(dnsServer)

		if dnsServerTrimmed == "" {
			klog.V(100).Infof("DNS server entry is empty")

			builder.errorMsg = "DNS server cannot be empty"

			return builder
		}

		builder.Definition.Spec.DNSServers =
			append(builder.Definition.Spec.DNSServers, lcaipcv1.IPAddress(dnsServerTrimmed))
	}

	return builder
}
