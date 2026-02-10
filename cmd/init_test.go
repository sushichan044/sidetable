//nolint:testpackage // Need package-level access to unexported helpers.
package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/config"
)

func TestInitCommandCreatesConfig(t *testing.T) {
	base := t.TempDir()
	t.Setenv("SIDETABLE_CONFIG_DIR", base)

	var buf bytes.Buffer
	initCmd.SetOut(&buf)
	initCmd.SetErr(&buf)

	err := initCmd.RunE(initCmd, []string{})
	require.NoError(t, err)

	path := filepath.Join(base, "config.yml")
	data, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	require.Equal(t, config.DefaultConfigYAML, data)
}

func TestInitCommandFailsWhenConfigExists(t *testing.T) {
	base := t.TempDir()
	t.Setenv("SIDETABLE_CONFIG_DIR", base)

	path := filepath.Join(base, "config.yml")
	require.NoError(t, os.WriteFile(path, []byte("directory: .sidetable\ncommands: {x: {command: echo}}\n"), 0o644))

	var buf bytes.Buffer
	initCmd.SetOut(&buf)
	initCmd.SetErr(&buf)

	err := initCmd.RunE(initCmd, []string{})
	require.Error(t, err)

	data, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	require.Equal(t, "directory: .sidetable\ncommands: {x: {command: echo}}\n", string(data))
}
