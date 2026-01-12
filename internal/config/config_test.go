package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePath(t *testing.T) {
	base := t.TempDir()
	require.NoError(t, os.Setenv("XDG_CONFIG_HOME", base))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))
	})

	sidetableDir := filepath.Join(base, "sidetable")
	require.NoError(t, os.MkdirAll(sidetableDir, 0o755))

	yamlPath := filepath.Join(sidetableDir, "config.yaml")
	ymlPath := filepath.Join(sidetableDir, "config.yml")

	t.Run("yaml only", func(t *testing.T) {
		require.NoError(t, os.WriteFile(yamlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
		path, err := ResolvePath()
		require.NoError(t, err)
		require.YAMLEq(t, yamlPath, path)
		require.NoError(t, os.Remove(yamlPath))
	})

	t.Run("yml only", func(t *testing.T) {
		require.NoError(t, os.WriteFile(ymlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
		path, err := ResolvePath()
		require.NoError(t, err)
		require.YAMLEq(t, ymlPath, path)
		require.NoError(t, os.Remove(ymlPath))
	})

	t.Run("both exist", func(t *testing.T) {
		require.NoError(t, os.WriteFile(yamlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
		require.NoError(t, os.WriteFile(ymlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
		_, err := ResolvePath()
		require.Error(t, err)
		require.NoError(t, os.Remove(yamlPath))
		require.NoError(t, os.Remove(ymlPath))
	})

	t.Run("missing", func(t *testing.T) {
		_, err := ResolvePath()
		require.Error(t, err)
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg := &Config{
			Directory: ".private",
			Commands: map[string]Command{
				"ghq": {Command: "ghq"},
			},
		}
		require.NoError(t, cfg.Validate())
	})

	t.Run("missing directory", func(t *testing.T) {
		cfg := &Config{Commands: map[string]Command{"a": {Command: "a"}}}
		require.Error(t, cfg.Validate())
	})

	t.Run("absolute directory", func(t *testing.T) {
		abs := "/abs"
		if runtime.GOOS == "windows" {
			abs = "C:\\abs"
		}
		cfg := &Config{
			Directory: abs,
			Commands:  map[string]Command{"a": {Command: "a"}},
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("missing commands", func(t *testing.T) {
		cfg := &Config{Directory: ".private"}
		require.Error(t, cfg.Validate())
	})

	t.Run("empty command", func(t *testing.T) {
		cfg := &Config{
			Directory: ".private",
			Commands:  map[string]Command{"a": {Command: ""}},
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("command with spaces", func(t *testing.T) {
		cfg := &Config{
			Directory: ".private",
			Commands:  map[string]Command{"a": {Command: "bad cmd"}},
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("alias duplicate", func(t *testing.T) {
		cfg := &Config{
			Directory: ".private",
			Commands: map[string]Command{
				"a": {Command: "a", Alias: "x"},
				"b": {Command: "b", Alias: "x"},
			},
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("alias collides with command", func(t *testing.T) {
		cfg := &Config{
			Directory: ".private",
			Commands: map[string]Command{
				"a": {Command: "a", Alias: "b"},
				"b": {Command: "b"},
			},
		}
		require.Error(t, cfg.Validate())
	})
}

func TestResolveCommandName(t *testing.T) {
	cfg := &Config{
		Directory: ".private",
		Commands: map[string]Command{
			"a": {Command: "a", Alias: "x"},
			"b": {Command: "b"},
		},
	}

	name, cmd, err := cfg.ResolveCommand("a")
	require.NoError(t, err)
	require.Equal(t, "a", name)
	require.Equal(t, "a", cmd.Command)

	name, cmd, err = cfg.ResolveCommand("x")
	require.NoError(t, err)
	require.Equal(t, "a", name)
	require.Equal(t, "a", cmd.Command)

	_, _, err = cfg.ResolveCommand("missing")
	require.Error(t, err)
}
