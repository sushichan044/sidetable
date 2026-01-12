package action_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/action"
	"github.com/sushichan044/sidetable/internal/config"
)

func TestBuildArgsExpansion(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {
				Command: "tool",
				Args:    []string{"-a", "{{.Args}}", "-b"},
			},
		},
	}

	spec, err := action.Build(cfg, "tool", []string{"x", "y"}, t.TempDir())
	require.NoError(t, err)
	require.Equal(t, []string{"-a", "x", "y", "-b"}, spec.Args)
}

func TestBuildArgsMissingPlaceholder(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {
				Command: "tool",
				Args:    []string{"-a"},
			},
		},
	}

	_, err := action.Build(cfg, "tool", []string{"x"}, t.TempDir())
	require.Error(t, err)
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

func TestBuildArgsPlaceholderTwice(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {
				Command: "tool",
				Args:    []string{"{{.Args}}", "{{.Args}}"},
			},
		},
	}

	_, err := action.Build(cfg, "tool", []string{"x"}, t.TempDir())
	require.Error(t, err)
}

func TestBuildArgsPlaceholderWithText(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {
				Command: "tool",
				Args:    []string{"prefix-{{.Args}}"},
			},
		},
	}

	_, err := action.Build(cfg, "tool", []string{"x"}, t.TempDir())
	require.Error(t, err)
}

func TestTemplateEvaluation(t *testing.T) {
	projectDir := t.TempDir()
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"tool": {
				Command:     "{{.CommandDir}}",
				Args:        []string{"--root={{.PrivateDir}}"},
				Env:         map[string]string{"ROOT": "{{.ProjectDir}}"},
				Description: "{{.CommandDir}}",
			},
		},
	}

	spec, err := action.Build(cfg, "tool", []string{}, projectDir)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(projectDir, ".private", "tool"), spec.Command)
	require.Equal(t, []string{"--root=" + filepath.Join(projectDir, ".private")}, spec.Args)
	require.Contains(t, spec.Env, "ROOT="+projectDir)
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
