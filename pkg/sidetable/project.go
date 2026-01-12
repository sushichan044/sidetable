package sidetable

import (
	"errors"
	"fmt"
	"os"

	"github.com/sushichan044/sidetable/internal/action"
	"github.com/sushichan044/sidetable/internal/config"
)

// CommandInfo represents a resolved command entry.
type CommandInfo struct {
	Name        string
	Alias       string
	Description string
}

// Project provides API access to sidetable core logic.
type Project struct {
	config     *config.Config
	projectDir string
}

// NewProject loads config and prepares a project context.
func NewProject(projectDir string) (*Project, error) {
	if projectDir == "" {
		return nil, errors.New("projectDir must not be empty")
	}

	info, err := os.Stat(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("projectDir does not exist: %s", projectDir)
		}
		return nil, fmt.Errorf("failed to stat projectDir: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("projectDir is not a directory: %s", projectDir)
	}

	path, err := ResolveConfigPath()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	return &Project{config: cfg, projectDir: projectDir}, nil
}

// ResolveConfigPath resolves the config path with default search.
func ResolveConfigPath() (string, error) {
	return config.ResolvePath()
}

// ListCommands returns resolved command entries with descriptions.
func (p *Project) ListCommands() ([]CommandInfo, error) {
	if p == nil || p.config == nil {
		return nil, errors.New("project is not initialized")
	}
	result := make([]CommandInfo, 0, len(p.config.Commands))
	for _, name := range p.config.CommandNames() {
		commandCfg := p.config.Commands[name]
		desc, err := renderDescription(
			commandCfg.Description,
			p.projectDir,
			p.config.Directory,
			name,
			p.config.ConfigDir,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, CommandInfo{
			Name:        name,
			Alias:       commandCfg.Alias,
			Description: desc,
		})
	}
	return result, nil
}

// BuildAction resolves the delegated command spec for the given name and args.
func (p *Project) BuildAction(name string, userArgs []string) (*action.Action, error) {
	if p == nil || p.config == nil {
		return nil, errors.New("project is not initialized")
	}
	spec, err := action.Build(p.config, name, userArgs, p.projectDir)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

// Execute runs the delegated command.
func (p *Project) Execute(act *action.Action) error {
	if p == nil {
		return errors.New("project is not initialized")
	}
	if act == nil {
		return errors.New("action is nil")
	}
	return action.Execute(act)
}

// ProjectDir returns the working project directory for this Project.
func (p *Project) ProjectDir() string {
	if p == nil {
		return ""
	}
	return p.projectDir
}
