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

// Spec is a resolved delegated command.
type Spec struct {
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

// BuildSpec resolves the delegated command spec for the given name and args.
func (p *Project) BuildSpec(name string, userArgs []string) (*Spec, error) {
	if p == nil || p.config == nil {
		return nil, errors.New("project is not initialized")
	}
	spec, err := delegate.Build(p.config, name, userArgs, p.projectDir)
	if err != nil {
		return nil, err
	}
	return fromDelegateSpec(spec), nil
}

// Execute runs the delegated command.
func (p *Project) Execute(spec *Spec) error {
	if p == nil {
		return errors.New("project is not initialized")
	}
	if spec == nil {
		return errors.New("spec is nil")
	}
	return delegate.Execute(toDelegateSpec(spec))
}

func fromDelegateSpec(spec *delegate.Spec) *Spec {
	if spec == nil {
		return nil
	}
	return &Spec{
		Command:    spec.Command,
		Args:       spec.Args,
		Env:        spec.Env,
		ProjectDir: spec.ProjectDir,
		PrivateDir: spec.PrivateDir,
		CommandDir: spec.CommandDir,
	}
}

func toDelegateSpec(spec *Spec) *delegate.Spec {
	if spec == nil {
		return nil
	}
	return &delegate.Spec{
		Command:    spec.Command,
		Args:       spec.Args,
		Env:        spec.Env,
		ProjectDir: spec.ProjectDir,
		PrivateDir: spec.PrivateDir,
		CommandDir: spec.CommandDir,
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
