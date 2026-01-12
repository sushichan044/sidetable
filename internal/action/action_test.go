package action_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/action"
	"github.com/sushichan044/sidetable/internal/config"
)

func TestBuildArgsPrependAppend(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {
				Command: "tool",
				Args: config.Args{
					Prepend: []string{"-a"},
					Append:  []string{"-b"},
				},
			},
		},
	}

	spec, err := action.Build(cfg, "tool", []string{"x", "y"}, t.TempDir())
	require.NoError(t, err)
	require.Equal(t, []string{"-a", "x", "y", "-b"}, spec.Args)
}

func TestBuildArgsPrependOnly(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {
				Command: "tool",
				Args: config.Args{
					Prepend: []string{"-a"},
				},
			},
		},
	}

	spec, err := action.Build(cfg, "tool", []string{"x"}, t.TempDir())
	require.NoError(t, err)
	require.Equal(t, []string{"-a", "x"}, spec.Args)
}

func TestBuildArgsNoArgsSpecified(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {Command: "tool"},
		},
	}

	spec, err := action.Build(cfg, "tool", []string{"x"}, t.TempDir())
	require.NoError(t, err)
	require.Equal(t, []string{"x"}, spec.Args)
}

func TestBuildArgsAppendOnly(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {
				Command: "tool",
				Args: config.Args{
					Append: []string{"-b"},
				},
			},
		},
	}

	spec, err := action.Build(cfg, "tool", []string{"x"}, t.TempDir())
	require.NoError(t, err)
	require.Equal(t, []string{"x", "-b"}, spec.Args)
}

func TestTemplateEvaluation(t *testing.T) {
	projectDir := t.TempDir()
	configDir := t.TempDir()
	cfg := &config.Config{
		Directory: ".private",
		ConfigDir: configDir,
		Commands: map[string]config.Command{
			"tool": {
				Command:     "{{.CommandDir}}",
				Args:        config.Args{Append: []string{"--root={{.PrivateDir}}"}},
				Env:         map[string]string{"ROOT": "{{.ProjectDir}}", "CONFIG": "{{.ConfigDir}}"},
				Description: "{{.CommandDir}}",
			},
		},
	}

	spec, err := action.Build(cfg, "tool", []string{}, projectDir)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(projectDir, ".private", "tool"), spec.Command)
	require.Equal(t, []string{"--root=" + filepath.Join(projectDir, ".private")}, spec.Args)
	require.Contains(t, spec.Env, "ROOT="+projectDir)
	require.Contains(t, spec.Env, "CONFIG="+configDir)
}

func TestCommandValidationAfterTemplate(t *testing.T) {
	projectDir := t.TempDir()
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {
				Command: "bad cmd",
			},
		},
	}

	_, err := action.Build(cfg, "tool", []string{}, projectDir)
	require.Error(t, err)
}
