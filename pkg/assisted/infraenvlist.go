package assisted

import (
	"context"
	"errors"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	agentInstallV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var errEmptyInfraEnvListNamespace = errors.New("namespace to list infraenvs cannot be empty")

// ListInfraEnvsInAllNamespaces returns a cluster-wide InfraEnv inventory.
func ListInfraEnvsInAllNamespaces(
	apiClient *clients.Settings, options ...goclient.ListOption) ([]*InfraEnvBuilder, error) {
	return common.List[agentInstallV1Beta1.InfraEnv, agentInstallV1Beta1.InfraEnvList, InfraEnvBuilder](
		context.TODO(), apiClient, agentInstallV1Beta1.AddToScheme, options...)
}

// ListInfraEnvs returns an InfraEnv inventory in a given namespace.
func ListInfraEnvs(apiClient *clients.Settings, namespace string) ([]*InfraEnvBuilder, error) {
	if namespace == "" {
		return nil, errEmptyInfraEnvListNamespace
	}

	return ListInfraEnvsInAllNamespaces(apiClient, &goclient.ListOptions{Namespace: namespace})
}
