package lca

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	assistedv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	// ImageBasedInstallationConfigVersion is the apiVersion for image-based-installation-config.yaml.
	ImageBasedInstallationConfigVersion = "v1beta1"

	ibiConfigFileName = "image-based-installation-config.yaml"
)

// ImageDigestSource defines a source repository and optional mirrors for release-image content
// in image-based-installation-config.yaml (aligned with openshift/installer/pkg/types.ImageDigestSource).
type ImageDigestSource struct {
	Source  string   `json:"source"`
	Mirrors []string `json:"mirrors,omitempty"`
}

// InstallationConfig is the API for image-based-installation-config.yaml consumed by openshift-install.
type InstallationConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	AdditionalTrustBundle  string                `json:"additionalTrustBundle,omitempty"`
	Architecture           string                `json:"architecture,omitempty"`
	ExtraPartitionLabel    string                `json:"extraPartitionLabel,omitempty"`
	IgnitionConfigOverride string                `json:"ignitionConfigOverride,omitempty"`
	ImageDigestSources     []ImageDigestSource   `json:"imageDigestSources,omitempty"`
	InstallationDisk       string                `json:"installationDisk"`
	NetworkConfig          *assistedv1.NetConfig `json:"networkConfig,omitempty"`
	PullSecret             string                `json:"pullSecret"`
	SeedImage              string                `json:"seedImage"`
	SeedVersion            string                `json:"seedVersion"`
	SSHKey                 string                `json:"sshKey,omitempty"`
}

// InstallationConfigInput holds values used to build image-based-installation-config.yaml.
type InstallationConfigInput struct {
	Architecture           string
	SeedImage              string
	SeedVersion            string
	AdditionalTrustBundle  string
	ImageDigestSources     []ImageDigestSource
	PullSecret             string
	InstallationDisk       string
	SSHKey                 string
	NetworkConfig          *assistedv1.NetConfig
	IgnitionConfigOverride string
	ExtraPartitionLabel    string
}

// NormalizeIgnitionConfigOverrideForIBI validates ignition JSON and returns a single-line compact form,
// matching Ansible `ignition_config | to_json` for use in image-based-installation-config.yaml.
func NormalizeIgnitionConfigOverrideForIBI(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	var buf bytes.Buffer

	err := json.Compact(&buf, []byte(trimmed))
	if err != nil {
		return "", fmt.Errorf("ignitionConfigOverride is not valid JSON: %w", err)
	}

	return buf.String(), nil
}

// SeedVersionFromSeedImage derives seedVersion from seedImage for ImageBasedInstallationConfig.
// Digest-pinned refs (anything after "@") are ignored when extracting a tag: only the repository
// side of "@" is considered. The tag is the substring after the last ':' only when that ':'
// follows the last '/' (so digest hex after sha256 is not used as seedVersion, and host:port
// before the path is not mistaken for a tag).
func SeedVersionFromSeedImage(seedImage string) string {
	ref := strings.TrimSpace(seedImage)
	if ref == "" {
		return ""
	}

	if i := strings.Index(ref, "@"); i >= 0 {
		ref = strings.TrimSpace(ref[:i])
	}

	if ref == "" {
		return ""
	}

	lastSlash := strings.LastIndex(ref, "/")

	lastColon := strings.LastIndex(ref, ":")
	if lastColon <= lastSlash {
		return ""
	}

	return ref[lastColon+1:]
}

// WriteInstallationConfig writes image-based-installation-config.yaml to destDir.
func WriteInstallationConfig(data InstallationConfigInput, destDir string) error {
	klog.V(100).Infof("Generating %s in %s", ibiConfigFileName, destDir)

	ignition := strings.TrimSpace(data.IgnitionConfigOverride)
	if ignition != "" {
		normalized, err := NormalizeIgnitionConfigOverrideForIBI(ignition)
		if err != nil {
			return fmt.Errorf("ignitionConfigOverride: %w", err)
		}

		ignition = normalized
	}

	cfg := InstallationConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ImageBasedInstallationConfigVersion,
			Kind:       "ImageBasedInstallationConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "image-based-installation-config",
		},
		AdditionalTrustBundle: data.AdditionalTrustBundle,
		PullSecret:            data.PullSecret,
		InstallationDisk:      data.InstallationDisk,
		SSHKey:                data.SSHKey,
		SeedImage:             data.SeedImage,
		SeedVersion:           data.SeedVersion,
		NetworkConfig:         data.NetworkConfig,
		ImageDigestSources:    data.ImageDigestSources,
	}

	if data.Architecture != "" {
		cfg.Architecture = data.Architecture
	}

	if ignition != "" {
		cfg.IgnitionConfigOverride = ignition
	}

	if data.ExtraPartitionLabel != "" {
		cfg.ExtraPartitionLabel = data.ExtraPartitionLabel
	}

	out, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal InstallationConfig: %w", err)
	}

	destPath := filepath.Join(destDir, ibiConfigFileName)

	err = os.WriteFile(destPath, out, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	klog.V(100).Infof("Successfully generated %s", destPath)

	return nil
}
