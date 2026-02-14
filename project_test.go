package sidetable_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable"
	"github.com/sushichan044/sidetable/internal/action"
	"github.com/sushichan044/sidetable/internal/config"
)

// setupProject writes a config file and tell sidetable to use it.
func setupProject(t *testing.T, cfg *config.Config) error {
	dir := t.TempDir()
	t.Setenv("SIDETABLE_CONFIG_DIR", dir)

	configPath := filepath.Join(dir, "config.yml")
	content, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if writeErr := os.WriteFile(configPath, content, 0o600); writeErr != nil {
		return writeErr
	}

	return nil
}

func minimalConfigForTest() *config.Config {
	return &config.Config{
		Directory: ".sidetable",
		Commands: map[string]config.Command{
			"hello": {
				Command:     "echo",
				Args:        config.Args{},
				Env:         map[string]string{},
				Description: "hi",
			},
		},
	}
}

func TestNewProjectInputValidation(t *testing.T) {
	t.Run("emptyPath", func(t *testing.T) {
		_, err := sidetable.NewProject("")
		require.EqualError(t, err, "projectDir must not be empty")
	})

	t.Run("missingDirectory", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing")
		_, err := sidetable.NewProject(missing)
		require.ErrorContains(t, err, "projectDir does not exist")
	})

	t.Run("notADirectory", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "file.txt")
		require.NoError(t, os.WriteFile(file, []byte("content"), 0o600))

		_, err := sidetable.NewProject(file)
		require.ErrorContains(t, err, "projectDir is not a directory")
	})
}

func TestNewProjectLoadsConfig(t *testing.T) {
	projectDir := t.TempDir()
	err := setupProject(t, minimalConfigForTest())
	require.NoError(t, err)

	project, err := sidetable.NewProject(projectDir)
	require.NoError(t, err)
	require.NotNil(t, project)
	require.Equal(t, projectDir, project.ProjectDir())
}

func TestProjectListCommandsValidation(t *testing.T) {
	t.Run("nilProject", func(t *testing.T) {
		var project *sidetable.Project
		_, err := project.ListCommands()
		require.EqualError(t, err, "project is not initialized")
	})

	t.Run("nilConfig", func(t *testing.T) {
		project := &sidetable.Project{}
		_, err := project.ListCommands()
		require.EqualError(t, err, "project is not initialized")
	})
}

func TestProjectListCommandsSuccess(t *testing.T) {
	projectDir := filepath.Join(t.TempDir(), "proj")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"zeta": {
				Command:     "cmd-zeta",
				Description: "Command: {{.CommandDir}}",
			},
			"alpha": {
				Command:     "cmd-alpha",
				Description: "Project: {{.ProjectDir}}",
			},
		},
		Aliases: map[string]config.Alias{
			"a": {
				Command:     "alpha",
				Description: "alpha alias",
			},
		},
	}

	err := setupProject(t, cfg)
	require.NoError(t, err)

	project, err := sidetable.NewProject(projectDir)
	require.NoError(t, err)

	commands, err := project.ListCommands()
	require.NoError(t, err)
	require.Len(t, commands, 3)

	require.Equal(t, "alpha", commands[0].Name)
	require.Equal(t, "command", commands[0].Kind)
	require.Equal(t, "alpha", commands[0].Target)
	require.Equal(t, "Project: "+projectDir, commands[0].Description)

	require.Equal(t, "zeta", commands[1].Name)
	require.Equal(t, "command", commands[1].Kind)
	require.Equal(t, "zeta", commands[1].Target)
	expectedCommandDir := filepath.Join(projectDir, ".private", "zeta")
	require.Equal(t, "Command: "+expectedCommandDir, commands[1].Description)

	require.Equal(t, "a", commands[2].Name)
	require.Equal(t, "alias", commands[2].Kind)
	require.Equal(t, "alpha", commands[2].Target)
	require.Equal(t, "alpha alias", commands[2].Description)
}

func TestProjectListCommandsTemplateError(t *testing.T) {
	projectDir := t.TempDir()

	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"oops": {
				Command:     "cmd-oops",
				Description: "{{.Missing}}",
			},
		},
	}

	err := setupProject(t, cfg)
	require.NoError(t, err)

	project, err := sidetable.NewProject(projectDir)
	require.NoError(t, err)

	_, err = project.ListCommands()
	require.Error(t, err)
}

func TestProjectBuildActionValidation(t *testing.T) {
	projectDir := t.TempDir()
	err := setupProject(t, minimalConfigForTest())
	require.NoError(t, err)

	t.Run("nilProject", func(t *testing.T) {
		var project *sidetable.Project
		_, buildErr := project.BuildAction("anything", nil)
		require.EqualError(t, buildErr, "project is not initialized")
	})

	t.Run("nilConfig", func(t *testing.T) {
		project, projErr := sidetable.NewProject(projectDir)
		require.NoError(t, projErr)
		require.NotNil(t, project)
	})
}

func TestProjectBuildActionCommandNotFound(t *testing.T) {
	projectDir := t.TempDir()

	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"hello": {
				Command: "echo",
				Args:    config.Args{},
				Env:     map[string]string{},
			},
		},
	}

	err := setupProject(t, cfg)
	require.NoError(t, err)

	project, err := sidetable.NewProject(projectDir)
	require.NoError(t, err)

	_, err = project.BuildAction("missing", nil)
	require.ErrorIs(t, err, config.ErrCommandUnknown)
}

func TestProjectBuildActionSuccess(t *testing.T) {
	projectDir := t.TempDir()

	cfg := &config.Config{
		Directory: ".private",
		Commands: map[string]config.Command{
			"hello": {
				Command: "echo",
				Args: config.Args{
					Prepend: []string{"hi"},
				},
			},
		},
	}

	err := setupProject(t, cfg)
	require.NoError(t, err)

	project, err := sidetable.NewProject(projectDir)
	require.NoError(t, err)

	actionSpec, err := project.BuildAction("hello", []string{"user"})
	require.NoError(t, err)
	require.NotNil(t, actionSpec)
	require.Equal(t, "echo", actionSpec.Command)
	require.Equal(t, []string{"hi", "user"}, actionSpec.Args)
	require.Equal(t, filepath.Join(projectDir, ".private"), actionSpec.PrivateDir)
	require.Equal(t, filepath.Join(projectDir, ".private", "hello"), actionSpec.CommandDir)
}

func TestProjectExecuteValidation(t *testing.T) {
	projectDir := t.TempDir()
	err := setupProject(t, minimalConfigForTest())
	require.NoError(t, err)

	t.Run("nilProject", func(t *testing.T) {
		var project *sidetable.Project
		execErr := project.Execute(&action.Action{})
		require.EqualError(t, execErr, "project is not initialized")
	})

	t.Run("nilAction", func(t *testing.T) {
		project, projErr := sidetable.NewProject(projectDir)
		require.NoError(t, projErr)
		execErr := project.Execute(nil)
		require.EqualError(t, execErr, "action is nil")
	})
}

func TestProjectExecuteSuccess(t *testing.T) {
	projectDir := t.TempDir()
	err := setupProject(t, minimalConfigForTest())
	require.NoError(t, err)

	project, err := sidetable.NewProject(projectDir)
	require.NoError(t, err)

	act := &action.Action{Command: "true"}
	execErr := project.Execute(act)
	require.NoError(t, execErr)
}
