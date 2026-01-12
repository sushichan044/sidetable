package sidetable

import (
	"errors"
	"os"

	"github.com/sushichan044/sidetable/internal/config"
	"github.com/sushichan044/sidetable/internal/delegate"
)

// CommandInfo represents a resolved command entry.
type CommandInfo struct {
	Name        string
	Alias       string
	Description string
}

// Action is a resolved delegated command.
type Action struct {
	Command    string
	Args       []string
	Env        []string
	ProjectDir string
	PrivateDir string
	CommandDir string
}

// Project provides API access to sidetable core logic.
type Project struct {
	config     *config.Config
	projectDir string
}

// NewProject loads config and prepares a project context.
func NewProject(configPath string, projectDir string) (*Project, error) {
	path, err := ResolveConfigPath(configPath)
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
func ResolveConfigPath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}
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
		desc, err := renderDescription(commandCfg.Description, p.projectDir, p.config.Directory, name)
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
func (p *Project) BuildAction(name string, userArgs []string) (*Action, error) {
	if p == nil || p.config == nil {
		return nil, errors.New("project is not initialized")
	}
	spec, err := delegate.Build(p.config, name, userArgs, p.projectDir)
	if err != nil {
		return nil, err
	}
	return fromDelegateAction(spec), nil
}

// Execute runs the delegated command.
func (p *Project) Execute(action *Action) error {
	if p == nil {
		return errors.New("project is not initialized")
	}
	if action == nil {
		return errors.New("spec is nil")
	}
	return delegate.Execute(toDelegateAction(action))
}

func fromDelegateAction(action *delegate.Action) *Action {
	if action == nil {
		return nil
	}
	return &Action{
		Command:    action.Command,
		Args:       action.Args,
		Env:        action.Env,
		ProjectDir: action.ProjectDir,
		PrivateDir: action.PrivateDir,
		CommandDir: action.CommandDir,
	}
}

func toDelegateAction(action *Action) *delegate.Action {
	if action == nil {
		return nil
	}
	return &delegate.Action{
		Command:    action.Command,
		Args:       action.Args,
		Env:        action.Env,
		ProjectDir: action.ProjectDir,
		PrivateDir: action.PrivateDir,
		CommandDir: action.CommandDir,
	}
}

// ProjectDir returns the working project directory for this Project.
func (p *Project) ProjectDir() string {
	if p == nil {
		return ""
	}
	return p.projectDir
}

// SetProjectDir sets the project directory for this Project.
func (p *Project) SetProjectDir(projectDir string) {
	if p == nil {
		return
	}
	p.projectDir = projectDir
}

// Env returns the current environment, useful for API callers.
func Env() []string {
	return os.Environ()
}
