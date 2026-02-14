package action_test

import (
	"errors"
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

func TestBuildArgsWithAliasPrepend(t *testing.T) {
	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"ghq": {
				Command: "ghq",
			},
		},
		Aliases: map[string]config.Alias{
			"gg": {
				Command: "ghq",
				Args: config.Args{
					Prepend: []string{"get"},
				},
			},
		},
	}

	spec, err := action.Build(cfg, "gg", []string{"https://github.com/example/repo"}, t.TempDir())
	require.NoError(t, err)
	require.Equal(t, []string{"get", "https://github.com/example/repo"}, spec.Args)
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

func TestExitError(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		err := &action.ExecError{Code: 1}
		require.Equal(t, "command exited with code 1", err.Error())
	})

	t.Run("Unwrap", func(t *testing.T) {
		innerErr := errors.New("inner error")
		err := &action.ExecError{Code: 1, Err: innerErr}
		require.Equal(t, innerErr, err.Unwrap())
	})
}

func TestExecute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		spec := &action.Action{
			Command: "echo",
			Args:    []string{"hello"},
		}
		err := action.Execute(spec)
		require.NoError(t, err)
	})

	t.Run("exit_code", func(t *testing.T) {
		spec := &action.Action{
			Command: "sh",
			Args:    []string{"-c", "exit 42"},
		}
		err := action.Execute(spec)
		require.Error(t, err)
		var exitErr *action.ExecError
		require.ErrorAs(t, err, &exitErr)
		require.Equal(t, 42, exitErr.Code)
	})

	t.Run("command_not_found", func(t *testing.T) {
		spec := &action.Action{
			Command: "nonexistent_command_that_does_not_exist_12345",
			Args:    []string{},
		}
		err := action.Execute(spec)
		require.Error(t, err)
	})
}

func TestBuildWithTemplateErrors(t *testing.T) {
	projectDir := t.TempDir()

	t.Run("empty_command", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"tool": {Command: "  "},
			},
		}
		_, err := action.Build(cfg, "tool", []string{}, projectDir)
		require.ErrorIs(t, err, action.ErrCommandTemplateEmpty)
	})

	t.Run("invalid_template_in_command", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"tool": {Command: "{{.Invalid}}"},
			},
		}
		_, err := action.Build(cfg, "tool", []string{}, projectDir)
		require.Error(t, err)
	})

	t.Run("invalid_template_in_args", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"tool": {
					Command: "tool",
					Args:    config.Args{Prepend: []string{"{{.Invalid}}"}},
				},
			},
		}
		_, err := action.Build(cfg, "tool", []string{}, projectDir)
		require.Error(t, err)
	})

	t.Run("invalid_template_in_env", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"tool": {
					Command: "tool",
					Env:     map[string]string{"KEY": "{{.Invalid}}"},
				},
			},
		}
		_, err := action.Build(cfg, "tool", []string{}, projectDir)
		require.Error(t, err)
	})
}

func TestBuildEnvMerge(t *testing.T) {
	projectDir := t.TempDir()

	t.Run("override_existing", func(t *testing.T) {
		t.Setenv("TEST_VAR", "original")

		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"tool": {
					Command: "tool",
					Env:     map[string]string{"TEST_VAR": "overridden"},
				},
			},
		}

		spec, err := action.Build(cfg, "tool", []string{}, projectDir)
		require.NoError(t, err)
		require.Contains(t, spec.Env, "TEST_VAR=overridden")
	})

	t.Run("add_new_variable", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"tool": {
					Command: "tool",
					Env:     map[string]string{"NEW_VAR": "value"},
				},
			},
		}

		spec, err := action.Build(cfg, "tool", []string{}, projectDir)
		require.NoError(t, err)
		require.Contains(t, spec.Env, "NEW_VAR=value")
	})

	t.Run("no_env_specified", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"tool": {Command: "tool"},
			},
		}

		spec, err := action.Build(cfg, "tool", []string{}, projectDir)
		require.NoError(t, err)
		require.NotEmpty(t, spec.Env)
	})

	t.Run("alias_env_overrides_command_env", func(t *testing.T) {
		cfg := &config.Config{
			Directory: ".private",
			Commands: map[string]config.Command{
				"tool": {
					Command: "tool",
					Env:     map[string]string{"ROOT": "command"},
				},
			},
			Aliases: map[string]config.Alias{
				"tt": {
					Command: "tool",
					Env: map[string]string{
						"ROOT": "alias",
						"NEW":  "x",
					},
				},
			},
		}

		spec, err := action.Build(cfg, "tt", []string{}, projectDir)
		require.NoError(t, err)
		require.Contains(t, spec.Env, "ROOT=alias")
		require.Contains(t, spec.Env, "NEW=x")
	})
}
