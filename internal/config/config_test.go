package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
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
		path, err := config.FindConfigPath()
		require.NoError(t, err)
		require.Equal(t, ymlPath, path)
		require.NoError(t, os.Remove(ymlPath))
	})

	t.Run("missing", func(t *testing.T) {
		_, err := config.FindConfigPath()
		require.ErrorIs(t, err, config.ErrConfigMissing)
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

	path, err := config.FindConfigPath()
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

	path, err := config.FindConfigPath()
	require.NoError(t, err)
	require.Equal(t, ymlPath, path)
}

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"ghq": {Command: "ghq"},
			},
			Aliases: map[string]config.Alias{
				"gg": {
					Command: "ghq",
					Args: config.Args{
						Append: []string{"get"},
					},
				},
			},
		}
		require.NoError(t, cfg.Validate())
	})

	t.Run("missing directory", func(t *testing.T) {
		cfg := &config.Config{Commands: map[string]config.Command{"a": {Command: "a"}}}
		require.ErrorIs(t, cfg.Validate(), config.ErrDirectoryRequired)
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
		require.ErrorIs(t, cfg.Validate(), config.ErrDirectoryMustBeRelative)
	})

	t.Run("missing commands key", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrCommandsMissing)
	})

	t.Run("empty command", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands:  map[string]config.Command{"a": {Command: ""}},
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrCommandRequired)
	})

	t.Run("command with spaces", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands:  map[string]config.Command{"a": {Command: "bad cmd"}},
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrCommandMustNotContainSpaces)
	})

	t.Run("command collides with builtin", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"list": {Command: "ghq"},
			},
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrCommandConflictsWithBuiltin)
	})

	t.Run("legacy alias is removed", func(t *testing.T) {
		legacyAlias := "x"
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"a": {Command: "a", LegacyAlias: &legacyAlias},
			},
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrLegacyCommandAliasRemoved)
	})

	t.Run("alias command required", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"a": {Command: "a"},
			},
			Aliases: map[string]config.Alias{
				"x": {Command: ""},
			},
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrAliasCommandRequired)
	})

	t.Run("alias target unknown", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"a": {Command: "a"},
			},
			Aliases: map[string]config.Alias{
				"x": {Command: "missing"},
			},
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrAliasTargetUnknown)
	})

	t.Run("alias with spaces", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"a": {Command: "a"},
			},
			Aliases: map[string]config.Alias{
				"bad alias": {Command: "a"},
			},
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrAliasMustNotContainSpaces)
	})

	t.Run("alias collides with command", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"a": {Command: "a"},
			},
			Aliases: map[string]config.Alias{
				"a": {Command: "a"},
			},
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrAliasConflictsWithCommand)
	})

	t.Run("alias collides with builtin", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"a": {Command: "a"},
			},
			Aliases: map[string]config.Alias{
				"list": {Command: "a"},
			},
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrAliasConflictsWithBuiltin)
	})
}

func TestResolveCommandName(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"ghq": {Command: "ghq"},
			"b":   {Command: "b"},
		},
		Aliases: map[string]config.Alias{
			"gg": {
				Command: "ghq",
				Args: config.Args{
					Append: []string{"get"},
				},
				Env: map[string]string{"X": "1"},
			},
		},
	}

	resolved, err := cfg.ResolveCommand("ghq")
	require.NoError(t, err)
	require.Equal(t, "ghq", resolved.Name)
	require.Equal(t, "ghq", resolved.Command.Command)
	require.Empty(t, resolved.AliasName)
	require.Nil(t, resolved.AliasArgs)
	require.Nil(t, resolved.AliasEnv)

	resolved, err = cfg.ResolveCommand("gg")
	require.NoError(t, err)
	require.Equal(t, "ghq", resolved.Name)
	require.Equal(t, "ghq", resolved.Command.Command)
	require.Equal(t, "gg", resolved.AliasName)
	require.Equal(t, []string{"get"}, resolved.AliasArgs.Append)
	require.Equal(t, map[string]string{"X": "1"}, resolved.AliasEnv)

	_, err = cfg.ResolveCommand("missing")
	require.ErrorIs(t, err, config.ErrCommandUnknown)
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
    env:
      A: a
      B: b
    args:
      prepend: ["-l"]
      append:
        - "-v"
aliases:
  gg:
    command: ghq
    description: "ghq get"
    args:
      append: ["get"]
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)

	require.Equal(t, ".private", cfg.Directory)
	require.Equal(t, filepath.Dir(path), cfg.ConfigDir)

	cmd, ok := cfg.Commands["ghq"]
	require.True(t, ok)
	require.Equal(t, "ghq", cmd.Command)
	require.Equal(t, "ghq wrapper", cmd.Description)
	require.Equal(t, map[string]string{"A": "a", "B": "b"}, cmd.Env)
	require.ElementsMatch(t, []string{"-l"}, cmd.Args.Prepend)
	require.ElementsMatch(t, []string{"-v"}, cmd.Args.Append)

	alias, ok := cfg.Aliases["gg"]
	require.True(t, ok)
	require.Equal(t, "ghq", alias.Command)
	require.ElementsMatch(t, []string{"get"}, alias.Args.Append)
}

func TestLoad_InvalidYAML(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "config.yml")

	bad := `
directory: .private
commands:
  - name: bad
`
	require.NoError(t, os.WriteFile(path, []byte(bad), 0o644))

	_, err := config.Load(path)
	require.Error(t, err)
}

func TestResolveCommandWithAliasInfo(t *testing.T) {
	cfg := &config.Config{
		Directory: ".test",
		Commands: map[string]config.Command{
			"ghq": {
				Command: "ghq",
			},
			"foo": {
				Command: "foo-bin",
			},
		},
		Aliases: map[string]config.Alias{
			"gg": {
				Command: "ghq",
				Args: config.Args{
					Append: []string{"get"},
				},
			},
		},
	}

	tests := []struct {
		name          string
		inputName     string
		wantName      string
		wantAlias     string
		wantCommand   string
		wantAliasArgs *config.Args
		wantErr       error
	}{
		{
			name:          "resolve by direct command name",
			inputName:     "ghq",
			wantName:      "ghq",
			wantAlias:     "",
			wantCommand:   "ghq",
			wantAliasArgs: nil,
			wantErr:       nil,
		},
		{
			name:        "resolve by alias name",
			inputName:   "gg",
			wantName:    "ghq",
			wantAlias:   "gg",
			wantCommand: "ghq",
			wantAliasArgs: &config.Args{
				Append: []string{"get"},
			},
			wantErr: nil,
		},
		{
			name:          "command without alias",
			inputName:     "foo",
			wantName:      "foo",
			wantAlias:     "",
			wantCommand:   "foo-bin",
			wantAliasArgs: nil,
			wantErr:       nil,
		},
		{
			name:      "command not found",
			inputName: "unknown",
			wantErr:   config.ErrCommandUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := cfg.ResolveCommand(tt.inputName)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, resolved)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resolved)
			assert.Equal(t, tt.wantName, resolved.Name)
			assert.Equal(t, tt.wantAlias, resolved.AliasName)
			assert.Equal(t, tt.wantCommand, resolved.Command.Command)
			assert.Equal(t, tt.wantAliasArgs, resolved.AliasArgs)
		})
	}
}
