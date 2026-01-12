package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/config"
)

func TestResolvePath(t *testing.T) {
	base := t.TempDir()
	envDir := filepath.Join(base, "env")
	require.NoError(t, os.MkdirAll(envDir, 0o755))
	t.Setenv("SIDETABLE_CONFIG_DIR", envDir)

	yamlPath := filepath.Join(envDir, "config.yaml")
	ymlPath := filepath.Join(envDir, "config.yml")

	t.Run("yaml only", func(t *testing.T) {
		require.NoError(t, os.WriteFile(yamlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
		path, err := config.ResolvePath()
		require.NoError(t, err)
		require.YAMLEq(t, yamlPath, path)
		require.NoError(t, os.Remove(yamlPath))
	})

	t.Run("yml only", func(t *testing.T) {
		require.NoError(t, os.WriteFile(ymlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
		path, err := config.ResolvePath()
		require.NoError(t, err)
		require.YAMLEq(t, ymlPath, path)
		require.NoError(t, os.Remove(ymlPath))
	})

	t.Run("both exist", func(t *testing.T) {
		require.NoError(t, os.WriteFile(yamlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
		require.NoError(t, os.WriteFile(ymlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
		_, err := config.ResolvePath()
		require.Error(t, err)
		require.NoError(t, os.Remove(yamlPath))
		require.NoError(t, os.Remove(ymlPath))
	})

	t.Run("missing", func(t *testing.T) {
		_, err := config.ResolvePath()
		require.Error(t, err)
	})
}

func TestResolvePathPrefersEnvDir(t *testing.T) {
	base := t.TempDir()
	envDir := filepath.Join(base, "env")
	xdgHome := filepath.Join(base, "xdg")
	xdgDir := filepath.Join(xdgHome, "sidetable")

	require.NoError(t, os.MkdirAll(envDir, 0o755))
	require.NoError(t, os.MkdirAll(xdgDir, 0o755))

	t.Setenv("SIDETABLE_CONFIG_DIR", envDir)
	t.Setenv("XDG_CONFIG_HOME", xdgHome)

	envPath := filepath.Join(envDir, "config.yaml")
	xdgPath := filepath.Join(xdgDir, "config.yaml")

	require.NoError(t, os.WriteFile(envPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
	require.NoError(t, os.WriteFile(xdgPath, []byte("directory: .private\ncommands: {}\n"), 0o644))

	path, err := config.ResolvePath()
	require.NoError(t, err)
	require.YAMLEq(t, envPath, path)
}

func TestResolvePathFallbackXDG(t *testing.T) {
	base := t.TempDir()
	xdgHome := filepath.Join(base, "xdg")
	xdgDir := filepath.Join(xdgHome, "sidetable")
	require.NoError(t, os.MkdirAll(xdgDir, 0o755))

	t.Setenv("SIDETABLE_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", xdgHome)

	yamlPath := filepath.Join(xdgDir, "config.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))

	path, err := config.ResolvePath()
	require.NoError(t, err)
	require.YAMLEq(t, yamlPath, path)
}

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"ghq": {Command: "ghq"},
			},
		}
		require.NoError(t, cfg.Validate())
	})

	t.Run("missing directory", func(t *testing.T) {
		cfg := &config.Config{Commands: map[string]config.Command{"a": {Command: "a"}}}
		require.Error(t, cfg.Validate())
	})

	t.Run("absolute directory", func(t *testing.T) {
		abs := "/abs"
		if runtime.GOOS == "windows" {
			abs = "C:\\abs"
		}
		cfg := &config.Config{
			Directory: abs,
			Commands:  map[string]config.Command{"a": {Command: "a"}},
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("missing commands", func(t *testing.T) {
		cfg := &config.Config{Directory: ".private"}
		require.Error(t, cfg.Validate())
	})

	t.Run("empty command", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands:  map[string]config.Command{"a": {Command: ""}},
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("command with spaces", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands:  map[string]config.Command{"a": {Command: "bad cmd"}},
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("alias duplicate", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"a": {Command: "a", Alias: "x"},
				"b": {Command: "b", Alias: "x"},
			},
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("alias collides with command", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"a": {Command: "a", Alias: "b"},
				"b": {Command: "b"},
			},
		}
		require.Error(t, cfg.Validate())
	})
}

func TestResolveCommandName(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
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
