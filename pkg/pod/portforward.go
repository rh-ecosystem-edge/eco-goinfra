package pod

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/klog/v2"
)

const defaultPortForwardReadyTimeout = 60 * time.Second

// PortForward establishes a port-forward to the pod and returns the local address (e.g. "localhost:8443")
// and a stop function. Callers must invoke the stop function to close the port-forward when done.
func (builder *Builder) PortForward(localPort, remotePort int) (string, func(), error) {
	if valid, err := builder.validate(); !valid {
		return "", nil, err
	}

	if !builder.Exists() {
		return "", nil, fmt.Errorf("pod object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	klog.V(100).Infof("Setting up port-forward %d:%d to pod %s in namespace %s",
		localPort, remotePort, builder.Object.Name, builder.Object.Namespace)

	restConfig := builder.apiClient.Config

	apiURL, err := url.Parse(restConfig.Host)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse API server URL: %w", err)
	}

	apiURL.Path = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		builder.Object.Namespace, builder.Object.Name)

	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create SPDY round-tripper: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiURL)

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	forwarder, err := portforward.New(dialer,
		[]string{fmt.Sprintf("%d:%d", localPort, remotePort)},
		stopChan, readyChan, nil, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create port-forwarder: %w", err)
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- forwarder.ForwardPorts()
	}()

	select {
	case <-readyChan:
	case err := <-errChan:
		return "", nil, fmt.Errorf("port-forward failed: %w", err)
	case <-time.After(defaultPortForwardReadyTimeout):
		close(stopChan)

		return "", nil, fmt.Errorf("port-forward to pod %s did not become ready in %s",
			builder.Object.Name, defaultPortForwardReadyTimeout)
	}

	stop := func() {
		close(stopChan)
	}

	return fmt.Sprintf("localhost:%d", localPort), stop, nil
}
