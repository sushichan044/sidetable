package sidetable_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable"
	"github.com/sushichan044/sidetable/internal/config"
)

func setupTestWorkspace(
	t *testing.T,
	tools map[string]config.Tool,
	aliases map[string]config.Alias,
) *sidetable.Workspace {
	t.Helper()

	// Create a temporary directory for the test workspace.
	projectDir := t.TempDir()

	// Write the config file.
	cfg := &config.Config{
		Directory: ".sidetable",
		Tools:     tools,
		Aliases:   aliases,
	}
	configContent, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	configPath := filepath.Join(projectDir, ".sidetable.yml")
	require.NoError(t, os.WriteFile(configPath, configContent, 0o600))

	// Open the workspace.
	workspace, err := sidetable.Open(projectDir, sidetable.WithConfigPath(configPath))
	require.NoError(t, err)
	return workspace
}

func TestWorkspaceRun(t *testing.T) {
	ws := setupTestWorkspace(
		t,
		map[string]config.Tool{
			"hello": {
				Run:         "echo",
				Args:        config.Args{Prepend: []string{"hello"}},
				Env:         map[string]string{},
				Description: "echoes hello",
			},
			"fail": {
				Run:         "sh",
				Args:        config.Args{Prepend: []string{"-c", "exit 42"}},
				Env:         map[string]string{},
				Description: "exits with code 42",
			},
		},
		map[string]config.Alias{},
	)

	t.Run("success", func(t *testing.T) {
		err := ws.Run(context.Background(), "hello", []string{}, sidetable.InvokeOptions{})
		require.NoError(t, err)
	})

	t.Run("preserve exit code", func(t *testing.T) {
		err := ws.Run(context.Background(), "fail", []string{}, sidetable.InvokeOptions{})
		require.Error(t, err)

		var invErr *sidetable.InvocationError
		require.ErrorAs(t, err, &invErr)
		require.Equal(t, 42, invErr.Code)
	})

	t.Run("error if specified non existent command", func(t *testing.T) {
		err := ws.Run(
			context.Background(),
			"nonexistent_command_that_does_not_exist_12345",
			[]string{},
			sidetable.InvokeOptions{},
		)
		require.Error(t, err)
	})
}
