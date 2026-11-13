package metallb

import (
	"context"
	"fmt"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// BGPSessionStateBuilder provides struct for BGPSessionState object which contains connection to cluster and
// BGPSessionState definitions.
type BGPSessionStateBuilder struct {
	common.EmbeddableBuilder[frrtypes.BGPSessionState, *frrtypes.BGPSessionState]
}

// GetGVK returns the BGPSessionState GVK for this builder.
func (builder *BGPSessionStateBuilder) GetGVK() schema.GroupVersionKind {
	return frrtypes.GroupVersion.WithKind("BGPSessionState")
}

// PullBGPSessionState retrieves an existing BGPSessionState object from the cluster.
func PullBGPSessionState(apiClient *clients.Settings, name string) (*BGPSessionStateBuilder, error) {
	klog.V(100).Infof("Pulling BGPSessionState object name:%s", name)

	return common.PullClusterScopedBuilder[frrtypes.BGPSessionState, BGPSessionStateBuilder](
		context.TODO(), apiClient, frrtypes.AddToScheme, name)
}

// PullBGPSessionStateByNodeAndPeer retrieves a BGPSessionState object by node name and peer IP.
// Since BGPSessionState names are auto-generated with random suffixes, this function
// lists all BGPSessionStates and filters by the node name and peer IP in the Status field.
func PullBGPSessionStateByNodeAndPeer(apiClient *clients.Settings, nodeName, peerIP string) (*BGPSessionStateBuilder, error) {
	klog.V(100).Infof("Pulling BGPSessionState object by node name:%s and peer IP:%s", nodeName, peerIP)

	if apiClient == nil {
		klog.V(100).Info("The apiClient cannot be nil")

		return nil, fmt.Errorf("the apiClient cannot be nil")
	}

	if nodeName == "" {
		klog.V(100).Info("The node name cannot be empty")

		return nil, fmt.Errorf("node name cannot be empty")
	}

	if peerIP == "" {
		klog.V(100).Info("The peer IP cannot be empty")

		return nil, fmt.Errorf("peer IP cannot be empty")
	}

	bgpSessionStates, err := ListBGPSessionState(apiClient)
	if err != nil {
		klog.V(100).Infof("Failed to list BGPSessionStates: %s", err.Error())

		return nil, fmt.Errorf("failed to list BGPSessionStates: %w", err)
	}

	for _, bgpSessionState := range bgpSessionStates {
		if bgpSessionState.Object != nil &&
			bgpSessionState.Object.Status.Node == nodeName &&
			bgpSessionState.Object.Status.Peer == peerIP {
			if err := common.Validate(bgpSessionState); err != nil {
				return nil, err
			}

			return bgpSessionState, nil
		}
	}

	return nil, fmt.Errorf("BGPSessionState for node %s and peer %s not found", nodeName, peerIP)
}
