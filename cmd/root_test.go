//nolint:testpackage // Need package-level access to unexported helpers.
package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/action"
	"github.com/sushichan044/sidetable/pkg/sidetable"
)

func TestDetermineExitCode(t *testing.T) {
	t.Run("nil error returns zero", func(t *testing.T) {
		require.Equal(t, 0, determineExitCode(nil))
	})

	t.Run("delegated exit code is preserved", func(t *testing.T) {
		err := &action.ExecError{
			Code: 42,
			Err:  errors.New("boom"),
		}
		require.Equal(t, 42, determineExitCode(err))
	})

	t.Run("non-delegated errors return one", func(t *testing.T) {
		require.Equal(t, 1, determineExitCode(errors.New("unexpected")))
	})
}

func TestBuildProjectCommands(t *testing.T) {
	base := t.TempDir()
	projectDir := filepath.Join(base, "project")
	configDir := filepath.Join(base, "config")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	configPath := filepath.Join(configDir, "config.yml")
	content := `
directory: .private
commands:
  hello:
    command: echo
    description: hello command
    alias: h
  list:
    command: echo
    description: conflicts with builtin and must be ignored
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o644))
	t.Setenv("SIDETABLE_CONFIG_DIR", configDir)

	project, err := sidetable.NewProject(projectDir)
	require.NoError(t, err)

	root := &cobra.Command{Use: "sidetable"}
	for _, sub := range buildProjectCommands(project) {
		root.AddCommand(sub)
	}

	found, _, err := root.Find([]string{"hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", found.Name())

	found, _, err = root.Find([]string{"h"})
	require.NoError(t, err)
	require.Equal(t, "hello", found.Name())

	_, _, err = root.Find([]string{"list"})
	require.Error(t, err)
}
