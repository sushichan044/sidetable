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

	ymlPath := filepath.Join(envDir, "config.yml")

	t.Run("yml only", func(t *testing.T) {
		require.NoError(t, os.WriteFile(ymlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
		path, err := config.ResolvePath()
		require.NoError(t, err)
		require.Equal(t, ymlPath, path) //nolint:testifylint // Comparing file paths, not YAML content
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

	envPath := filepath.Join(envDir, "config.yml")
	xdgPath := filepath.Join(xdgDir, "config.yml")

	require.NoError(t, os.WriteFile(envPath, []byte("directory: .private\ncommands: {}\n"), 0o644))
	require.NoError(t, os.WriteFile(xdgPath, []byte("directory: .private\ncommands: {}\n"), 0o644))

	path, err := config.ResolvePath()
	require.NoError(t, err)
	require.Equal(t, envPath, path)
}

func TestResolvePathFallbackXDG(t *testing.T) {
	base := t.TempDir()
	xdgHome := filepath.Join(base, "xdg")
	xdgDir := filepath.Join(xdgHome, "sidetable")
	require.NoError(t, os.MkdirAll(xdgDir, 0o755))

	t.Setenv("SIDETABLE_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", xdgHome)

	ymlPath := filepath.Join(xdgDir, "config.yml")
	require.NoError(t, os.WriteFile(ymlPath, []byte("directory: .private\ncommands: {}\n"), 0o644))

	path, err := config.ResolvePath()
	require.NoError(t, err)
	require.Equal(t, ymlPath, path) //nolint:testifylint // Comparing file paths, not YAML content
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

func TestLoad_ParsesYAML(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "config.yml")

	content := `
directory: .private
commands:
  ghq:
    command: ghq
    description: "ghq wrapper"
    alias: q
    env:
      A: a
      B: b
    args:
      prepend: ["-l"]
      append:
        - "-v"
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)

	require.Equal(t, ".private", cfg.Directory)
	require.Equal(t, filepath.Dir(path), cfg.ConfigDir)

	cmd, ok := cfg.Commands["ghq"]
	require.True(t, ok)
	require.Equal(t, "ghq", cmd.Command)
	require.Equal(t, "q", cmd.Alias)
	require.Equal(t, "ghq wrapper", cmd.Description)
	require.Equal(t, map[string]string{"A": "a", "B": "b"}, cmd.Env)
	require.ElementsMatch(t, []string{"-l"}, cmd.Args.Prepend)
	require.ElementsMatch(t, []string{"-v"}, cmd.Args.Append)
}

func TestLoad_InvalidYAML(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "config.yml")

	// commands should be a mapping; provide a list to force unmarshal error
	bad := `
directory: .private
commands:
  - name: bad
`
	require.NoError(t, os.WriteFile(path, []byte(bad), 0o644))

	_, err := config.Load(path)
	require.Error(t, err)
}
