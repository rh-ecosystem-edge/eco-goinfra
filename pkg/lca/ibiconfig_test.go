package lca

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	assistedv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func minimalInstallationConfigInput() InstallationConfigInput {
	return InstallationConfigInput{
		SeedImage:        "registry.example.com/ocp/release:4.16.0",
		SeedVersion:      "4.16.0",
		PullSecret:       `{"auths":{}}`,
		InstallationDisk: "/dev/disk/by-id/wwn-123",
	}
}

func TestNormalizeIgnitionConfigOverrideForIBI(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		out, err := NormalizeIgnitionConfigOverrideForIBI("  ")
		require.NoError(t, err)
		assert.Empty(t, out)
	})

	t.Run("valid json compacted", func(t *testing.T) {
		t.Parallel()

		out, err := NormalizeIgnitionConfigOverrideForIBI(`{"ignition":{"version":"3.2.0"}}`)
		require.NoError(t, err)
		assert.Equal(t, `{"ignition":{"version":"3.2.0"}}`, out)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		_, err := NormalizeIgnitionConfigOverrideForIBI("{not-json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ignitionConfigOverride is not valid JSON")
	})
}

func TestSeedVersionFromSeedImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		seedImage string
		want      string
	}{
		{name: "empty", seedImage: "", want: ""},
		{name: "tag", seedImage: "registry.example.com/ocp/release:4.16.0", want: "4.16.0"},
		{name: "digest pinned", seedImage: "registry.example.com/ocp/release@sha256:abc123", want: ""},
		{name: "host port not tag", seedImage: "registry.example.com:5000/ocp/release", want: ""},
		{name: "tag after digest stripped", seedImage: "quay.io/foo/bar:4.17.1@sha256:deadbeef", want: "4.17.1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, SeedVersionFromSeedImage(tc.seedImage))
		})
	}
}

func TestWriteInstallationConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		err := WriteInstallationConfig(InstallationConfigInput{
			SeedImage:           "registry.example.com/ocp/release:4.16.0",
			SeedVersion:         "4.16.0",
			PullSecret:          `{"auths":{}}`,
			InstallationDisk:    "/dev/disk/by-id/wwn-123",
			SSHKey:              "ssh-rsa AAA",
			Architecture:        "amd64",
			ExtraPartitionLabel: "var-lib-containers",
			NetworkConfig:       &assistedv1.NetConfig{Raw: []byte("interfaces: []\n")},
			ImageDigestSources: []ImageDigestSource{
				{Source: "quay.io", Mirrors: []string{"mirror.example.com"}},
			},
		}, dir)
		require.NoError(t, err)

		raw, err := os.ReadFile(filepath.Join(dir, ibiConfigFileName))
		require.NoError(t, err)

		content := string(raw)
		assert.Contains(t, content, "apiVersion: v1beta1")
		assert.Contains(t, content, "kind: ImageBasedInstallationConfig")
		assert.Contains(t, content, "seedVersion: 4.16.0")
		assert.Contains(t, content, "installationDisk: /dev/disk/by-id/wwn-123")
		assert.Contains(t, content, "interfaces: []")
	})

	t.Run("includes normalized ignition override", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		input := minimalInstallationConfigInput()
		input.IgnitionConfigOverride = `  {"ignition":{"version":"3.2.0"}}  `

		err := WriteInstallationConfig(input, dir)
		require.NoError(t, err)

		raw, err := os.ReadFile(filepath.Join(dir, ibiConfigFileName))
		require.NoError(t, err)

		content := string(raw)
		assert.Contains(t, content, "ignitionConfigOverride:")
		assert.Contains(t, content, `{"ignition":{"version":"3.2.0"}}`)
	})
}

func TestWriteInstallationConfigErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid ignition override", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		input := minimalInstallationConfigInput()
		input.IgnitionConfigOverride = "{not-json"

		err := WriteInstallationConfig(input, dir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ignitionConfigOverride:")
		assert.Contains(t, err.Error(), "ignitionConfigOverride is not valid JSON")

		_, statErr := os.Stat(filepath.Join(dir, ibiConfigFileName))
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("missing destination directory", func(t *testing.T) {
		t.Parallel()

		destDir := filepath.Join(t.TempDir(), "missing", "nested")

		err := WriteInstallationConfig(minimalInstallationConfigInput(), destDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write config file")
	})

	t.Run("unwritable destination directory", func(t *testing.T) {
		t.Parallel()

		if runtime.GOOS == "windows" {
			t.Skip("directory permission modes are not portable on Windows")
		}

		dir := t.TempDir()
		require.NoError(t, os.Chmod(dir, 0o500))
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

		err := WriteInstallationConfig(minimalInstallationConfigInput(), dir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write config file")
	})
}
