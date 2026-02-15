//nolint:testpackage // Need access to package-private resolution helpers.
package sidetable

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/config"
)

func TestResolveInvocationArgsPrependAppend(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Tools: map[string]config.Tool{
			"tool": {
				Run: "tool",
				Args: config.Args{
					Prepend: []string{"-a"},
					Append:  []string{"-b"},
				},
			},
		},
	}

	inv, err := resolveInvocation(cfg, "tool", []string{"x", "y"}, t.TempDir(), []string{})
	require.NoError(t, err)
	require.Equal(t, []string{"-a", "x", "y", "-b"}, inv.Args)
}

func TestResolveInvocationArgsWithAliasPrepend(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Tools: map[string]config.Tool{
			"ghq": {Run: "ghq"},
		},
		Aliases: map[string]config.Alias{
			"gg": {
				Tool: "ghq",
				Args: config.Args{Prepend: []string{"get"}},
			},
		},
	}

	inv, err := resolveInvocation(cfg, "gg", []string{"https://github.com/example/repo"}, t.TempDir(), []string{})
	require.NoError(t, err)
	require.Equal(t, []string{"get", "https://github.com/example/repo"}, inv.Args)
}

func TestResolveInvocationTemplateEvaluation(t *testing.T) {
	workspaceRoot := t.TempDir()
	configDir := t.TempDir()
	cfg := &config.Config{
		Directory: ".private",
		FilePath:  filepath.Join(configDir, "config.yml"),
		Tools: map[string]config.Tool{
			"tool": {
				Run: "{{.ToolDir}}",
				Env: map[string]string{
					"ROOT":   "{{.WorkspaceRoot}}",
					"CONFIG": "{{.ConfigDir}}",
				},
			},
		},
	}

	inv, err := resolveInvocation(cfg, "tool", []string{}, workspaceRoot, []string{})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(workspaceRoot, ".private", "tool"), inv.Program)
	require.Contains(t, inv.Env, "CONFIG="+configDir)
	require.Contains(t, inv.Env, "ROOT="+workspaceRoot)
}

func TestResolveInvocationTemplateErrors(t *testing.T) {
	projectDir := t.TempDir()

	t.Run("empty_run", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"tool": {Run: "  "},
			},
		}

		_, err := resolveInvocation(cfg, "tool", []string{}, projectDir, []string{})
		require.ErrorIs(t, err, errRunTemplateEmpty)
	})

	t.Run("invalid_template_in_run", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"tool": {Run: "{{.Invalid}}"},
			},
		}

		_, err := resolveInvocation(cfg, "tool", []string{}, projectDir, []string{})
		require.Error(t, err)
	})

	t.Run("invalid_template_in_args", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"tool": {
					Run:  "tool",
					Args: config.Args{Prepend: []string{"{{.Invalid}}"}},
				},
			},
		}

		_, err := resolveInvocation(cfg, "tool", []string{}, projectDir, []string{})
		require.Error(t, err)
	})

	t.Run("invalid_template_in_env", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Tools: map[string]config.Tool{
				"tool": {
					Run: "tool",
					Env: map[string]string{"KEY": "{{.Invalid}}"},
				},
			},
		}

		_, err := resolveInvocation(cfg, "tool", []string{}, projectDir, []string{})
		require.Error(t, err)
	})
}
