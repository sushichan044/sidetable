package sidetable

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/sushichan044/sidetable/internal/action"
	"github.com/sushichan044/sidetable/internal/builtin"
	"github.com/sushichan044/sidetable/internal/config"
)

// CommandInfo represents a resolved command entry.
type CommandInfo struct {
	Name        string
	Description string
}

// AliasInfo represents a resolved alias entry.
type AliasInfo struct {
	Name        string
	Target      string
	Description string
}

// InvalidCommandInfo represents a command or alias disabled by validation rules.
type InvalidCommandInfo struct {
	Name        string
	Kind        string
	Target      string
	Description string
	Reason      string
}

// CommandList contains both valid and invalid command entries.
type CommandList struct {
	Commands []CommandInfo
	Aliases  []AliasInfo
	Invalid  []InvalidCommandInfo
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

	path, err := FindConfigPath()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	return &Project{config: cfg, projectDir: projectDir}, nil
}

// FindConfigPath returns the config path, erroring if it does not exist.
func FindConfigPath() (string, error) {
	return config.FindConfigPath()
}

// GetExecError extracts ExecError if err caused by user-defined command execution.
func GetExecError(err error) *action.ExecError {
	return action.GetExecError(err)
}

// ListCommands returns valid command/alias entries and invalid entries with reasons.
func (p *Project) ListCommands() (CommandList, error) {
	if p == nil || p.config == nil {
		return CommandList{}, errors.New("project is not initialized")
	}
	result := CommandList{
		Commands: make([]CommandInfo, 0, len(p.config.Commands)),
		Aliases:  make([]AliasInfo, 0, len(p.config.Aliases)),
		Invalid:  make([]InvalidCommandInfo, 0),
	}

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
			return CommandList{}, err
		}

		if builtin.IsReservedCommand(name) {
			result.Invalid = append(result.Invalid, InvalidCommandInfo{
				Name:        name,
				Kind:        "command",
				Target:      name,
				Description: desc,
				Reason:      "conflicts with builtin command",
			})
			continue
		}

		result.Commands = append(result.Commands, CommandInfo{
			Name:        name,
			Description: desc,
		})
	}

	aliasNames := make([]string, 0, len(p.config.Aliases))
	for aliasName := range p.config.Aliases {
		aliasNames = append(aliasNames, aliasName)
	}
	sort.Strings(aliasNames)

	for _, aliasName := range aliasNames {
		aliasCfg := p.config.Aliases[aliasName]
		desc, err := renderDescription(
			aliasCfg.Description,
			p.projectDir,
			p.config.Directory,
			aliasCfg.Command,
			p.config.ConfigDir,
		)
		if err != nil {
			return CommandList{}, err
		}

		if builtin.IsReservedCommand(aliasName) {
			result.Invalid = append(result.Invalid, InvalidCommandInfo{
				Name:        aliasName,
				Kind:        "alias",
				Target:      aliasCfg.Command,
				Description: desc,
				Reason:      "conflicts with builtin command",
			})
			continue
		}

		result.Aliases = append(result.Aliases, AliasInfo{
			Name:        aliasName,
			Target:      aliasCfg.Command,
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
