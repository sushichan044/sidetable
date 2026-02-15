package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	z "github.com/Oudwins/zog"

	"github.com/sushichan044/sidetable/internal/config"
)

func TestResolvePath(t *testing.T) {
	base := t.TempDir()
	envDir := filepath.Join(base, "env")
	require.NoError(t, os.MkdirAll(envDir, 0o755))
	t.Setenv("SIDETABLE_CONFIG_DIR", envDir)

	configPath := filepath.Join(envDir, "config.yml")

	t.Run("yml only", func(t *testing.T) {
		require.NoError(t, os.WriteFile(configPath, []byte("directory: .private\ntools: {}\n"), 0o644))
		path, err := config.FindConfigPath()
		require.NoError(t, err)
		require.Equal(t, configPath, path)
		require.NoError(t, os.Remove(configPath))
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

	require.NoError(t, os.WriteFile(envPath, []byte("directory: .private\ntools: {}\n"), 0o644))
	require.NoError(t, os.WriteFile(xdgPath, []byte("directory: .private\ntools: {}\n"), 0o644))

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

	configPath := filepath.Join(xdgDir, "config.yml")
	require.NoError(t, os.WriteFile(configPath, []byte("directory: .private\ntools: {}\n"), 0o644))

	path, err := config.FindConfigPath()
	require.NoError(t, err)
	require.Equal(t, configPath, path)
}

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"ghq": {Run: "ghq"},
			},
			Aliases: map[string]config.Alias{
				"gg": {
					Tool: "ghq",
					Args: config.Args{Append: []string{"get"}},
				},
			},
		}
		require.NoError(t, cfg.Validate())
	})

	t.Run("missing directory", func(t *testing.T) {
		cfg := &config.Config{Tools: map[string]config.Tool{"a": {Run: "a"}}}
		requireHasIssue(t, cfg.Validate(), "directory", "directory is required")
	})

	t.Run("absolute directory", func(t *testing.T) {
		abs := "/abs"
		if runtime.GOOS == "windows" {
			abs = "C:\\abs"
		}
		cfg := &config.Config{
			Directory: abs,
			Tools:     map[string]config.Tool{"a": {Run: "a"}},
		}
		requireHasIssue(t, cfg.Validate(), "directory", "directory must be relative")
	})

	t.Run("empty run", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools:     map[string]config.Tool{"a": {Run: ""}},
		}
		requireHasIssue(t, cfg.Validate(), `tools["a"].run`, "tool run is required")
	})

	t.Run("run with spaces", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools:     map[string]config.Tool{"a": {Run: "bad run"}},
		}
		requireHasIssue(t, cfg.Validate(), `tools["a"].run`, "tool run must not contain spaces")
	})

	t.Run("tool collides with builtin", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"list": {Run: "ghq"},
			},
		}
		requireHasIssue(t, cfg.Validate(), `tools["list"]`, "tool conflicts with builtin command")
	})

	t.Run("alias tool required", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"a": {Run: "a"},
			},
			Aliases: map[string]config.Alias{
				"x": {Tool: ""},
			},
		}
		err := cfg.Validate()
		requireHasIssue(t, err, `aliases["x"].tool`, "alias tool is required")
		requireNoIssueMessage(t, err, "alias tool not found")
	})

	t.Run("alias target unknown", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"a": {Run: "a"},
			},
			Aliases: map[string]config.Alias{
				"x": {Tool: "missing"},
			},
		}
		requireHasIssue(t, cfg.Validate(), `aliases["x"].tool`, "alias tool not found")
	})

	t.Run("alias with spaces", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"a": {Run: "a"},
			},
			Aliases: map[string]config.Alias{
				"bad alias": {Tool: "a"},
			},
		}
		requireHasIssue(t, cfg.Validate(), `aliases["bad alias"]`, "alias must not contain spaces")
	})

	t.Run("alias collides with tool", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"a": {Run: "a"},
			},
			Aliases: map[string]config.Alias{
				"a": {Tool: "a"},
			},
		}
		requireHasIssue(t, cfg.Validate(), `aliases["a"]`, "alias conflicts with tool name")
	})

	t.Run("alias collides with builtin", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"a": {Run: "a"},
			},
			Aliases: map[string]config.Alias{
				"list": {Tool: "a"},
			},
		}
		requireHasIssue(t, cfg.Validate(), `aliases["list"]`, "alias conflicts with builtin command")
	})

	t.Run("collects multiple validation errors", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"list": {
					Run: "bad run",
				},
			},
			Aliases: map[string]config.Alias{
				"help": {
					Tool: "missing",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		requireHasIssue(t, err, `tools["list"].run`, "tool run must not contain spaces")
		requireHasIssue(t, err, `tools["list"]`, "tool conflicts with builtin command")
		requireHasIssue(t, err, `aliases["help"]`, "alias conflicts with builtin command")
		requireHasIssue(t, err, `aliases["help"].tool`, "alias tool not found")
	})
}

func requireHasIssue(t *testing.T, err error, path string, message string) {
	t.Helper()
	require.Error(t, err)

	issues := collectIssues(err)
	for _, issue := range issues {
		if issue.PathString() == path && issue.Message == message {
			return
		}
	}

	t.Fatalf("issue not found path=%q message=%q issues=%v", path, message, issues)
}

func requireNoIssueMessage(t *testing.T, err error, message string) {
	t.Helper()
	require.Error(t, err)

	issues := collectIssues(err)
	for _, issue := range issues {
		if issue.Message == message {
			t.Fatalf("unexpected issue found message=%q issues=%v", message, issues)
		}
	}
}

func collectIssues(err error) z.ZogIssueList {
	issues := make(z.ZogIssueList, 0)

	var walk func(error)
	walk = func(e error) {
		if e == nil {
			return
		}

		var joinErr interface{ Unwrap() []error }
		if errors.As(e, &joinErr) {
			for _, unwrapped := range joinErr.Unwrap() {
				walk(unwrapped)
			}
			return
		}

		var issue *z.ZogIssue
		if errors.As(e, &issue) {
			issues = append(issues, issue)
		}
	}

	walk(err)
	return issues
}

func TestResolveEntryName(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Tools: map[string]config.Tool{
			"ghq": {Run: "ghq"},
			"b":   {Run: "b"},
		},
		Aliases: map[string]config.Alias{
			"gg": {
				Tool: "ghq",
				Args: config.Args{Append: []string{"get"}},
			},
		},
	}

	resolved, err := cfg.ResolveEntry("ghq")
	require.NoError(t, err)
	require.Equal(t, "ghq", resolved.ToolName)
	require.Equal(t, "ghq", resolved.Tool.Run)
	require.Empty(t, resolved.AliasName)
	require.Nil(t, resolved.AliasArgs)

	resolved, err = cfg.ResolveEntry("gg")
	require.NoError(t, err)
	require.Equal(t, "ghq", resolved.ToolName)
	require.Equal(t, "ghq", resolved.Tool.Run)
	require.Equal(t, "gg", resolved.AliasName)
	require.Equal(t, []string{"get"}, resolved.AliasArgs.Append)

	_, err = cfg.ResolveEntry("missing")
	require.ErrorIs(t, err, config.ErrEntryUnknown)
}

func TestLoad_ParsesYAML(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "config.yml")

	content := `
directory: .private
tools:
  ghq:
    run: ghq
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
    tool: ghq
    description: "ghq get"
    args:
      append: ["get"]
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)

	require.Equal(t, ".private", cfg.Directory)
	require.Equal(t, path, cfg.FilePath)

	tool, ok := cfg.Tools["ghq"]
	require.True(t, ok)
	require.Equal(t, "ghq", tool.Run)
	require.Equal(t, "ghq wrapper", tool.Description)
	require.Equal(t, map[string]string{"A": "a", "B": "b"}, tool.Env)
	require.ElementsMatch(t, []string{"-l"}, tool.Args.Prepend)
	require.ElementsMatch(t, []string{"-v"}, tool.Args.Append)

	alias, ok := cfg.Aliases["gg"]
	require.True(t, ok)
	require.Equal(t, "ghq", alias.Tool)
	require.ElementsMatch(t, []string{"get"}, alias.Args.Append)
}

func TestLoad_InvalidYAML(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "config.yml")

	bad := `
directory: .private
tools:
  - name: bad
`
	require.NoError(t, os.WriteFile(path, []byte(bad), 0o644))

	_, err := config.Load(path)
	require.Error(t, err)
}

func TestResolveEntryWithAliasInfo(t *testing.T) {
	cfg := &config.Config{
		Directory: ".test",
		Tools: map[string]config.Tool{
			"ghq": {
				Run: "ghq",
			},
			"foo": {
				Run: "foo-bin",
			},
		},
		Aliases: map[string]config.Alias{
			"gg": {
				Tool: "ghq",
				Args: config.Args{Append: []string{"get"}},
			},
		},
	}

	tests := []struct {
		name          string
		inputName     string
		wantToolName  string
		wantAlias     string
		wantRun       string
		wantAliasArgs *config.Args
		wantErr       error
	}{
		{
			name:          "resolve by direct tool name",
			inputName:     "ghq",
			wantToolName:  "ghq",
			wantAlias:     "",
			wantRun:       "ghq",
			wantAliasArgs: nil,
			wantErr:       nil,
		},
		{
			name:         "resolve by alias name",
			inputName:    "gg",
			wantToolName: "ghq",
			wantAlias:    "gg",
			wantRun:      "ghq",
			wantAliasArgs: &config.Args{
				Append: []string{"get"},
			},
			wantErr: nil,
		},
		{
			name:          "tool without alias",
			inputName:     "foo",
			wantToolName:  "foo",
			wantAlias:     "",
			wantRun:       "foo-bin",
			wantAliasArgs: nil,
			wantErr:       nil,
		},
		{
			name:      "entry not found",
			inputName: "unknown",
			wantErr:   config.ErrEntryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := cfg.ResolveEntry(tt.inputName)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, resolved)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resolved)
			assert.Equal(t, tt.wantToolName, resolved.ToolName)
			assert.Equal(t, tt.wantAlias, resolved.AliasName)
			assert.Equal(t, tt.wantRun, resolved.Tool.Run)
			assert.Equal(t, tt.wantAliasArgs, resolved.AliasArgs)
		})
	}
}
