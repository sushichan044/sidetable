package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/sushichan044/sidetable/internal/xdg"
)

var (
	ErrConfigMissing               = errors.New("config.yml file not found")
	ErrCommandsMissing             = errors.New("commands are required")
	ErrCommandUnknown              = errors.New("command not found")
	ErrDirectoryRequired           = errors.New("directory is required")
	ErrDirectoryMustBeRelative     = errors.New("directory must be relative")
	ErrCommandRequired             = errors.New("command is required")
	ErrCommandMustNotContainSpaces = errors.New("command must not contain spaces")
	ErrCommandConflictsWithBuiltin = errors.New("command conflicts with builtin command")
	ErrAliasDuplicate              = errors.New("alias is duplicated")
	ErrAliasConflictsWithCommand   = errors.New("alias conflicts with command name")
	ErrAliasConflictsWithBuiltin   = errors.New("alias conflicts with builtin command")
	ErrAliasNameRequired           = errors.New("alias name is required")
	ErrAliasMustNotContainSpaces   = errors.New("alias must not contain spaces")
	ErrAliasCommandRequired        = errors.New("alias command is required")
	ErrAliasTargetUnknown          = errors.New("alias command not found")
	ErrLegacyCommandAliasRemoved   = errors.New("commands.<name>.alias has been removed; use top-level aliases")
)

// Config represents configuration file structure.
type Config struct {
	Directory string             `yaml:"directory"` // User's private directory per project
	Commands  map[string]Command `yaml:"commands"`
	Aliases   map[string]Alias   `yaml:"aliases"`
	ConfigDir string             `yaml:"-"` // INTERNAL: Directory of the loaded config file
}

// Command represents a delegated command configuration.
type Command struct {
	Command     string            `yaml:"command"`
	Args        Args              `yaml:"args"`
	Env         map[string]string `yaml:"env"`
	Description string            `yaml:"description"`
	LegacyAlias *string           `yaml:"alias"`
}

// Alias represents a command alias configuration.
type Alias struct {
	Command     string            `yaml:"command"`
	Args        Args              `yaml:"args"`
	Description string            `yaml:"description"`
}

// Args represents user-arg injection configuration.
type Args struct {
	Prepend []string `yaml:"prepend"`
	Append  []string `yaml:"append"`
}

// ResolvedCommand represents a fully resolved command with optional alias information.
type ResolvedCommand struct {
	Name    string
	Command Command
	// The name of the alias used to invoke this command, if any.
	// Empty if invoked by the original command name.
	AliasName string
	AliasArgs *Args
	// DisplayName is the resolved CLI entrypoint name (command or alias).
	DisplayName string
}

const configDirEnv = "SIDETABLE_CONFIG_DIR"

// FindConfigPath returns the config path, erroring if it does not exist.
// This is used for commands that require an existing config.
func FindConfigPath() (string, error) {
	path, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	if _, statErr := os.Stat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			return "", ErrConfigMissing
		}
		return "", statErr
	}

	return path, nil
}

// GetConfigPath returns the config path from SIDETABLE_CONFIG_DIR or XDG_CONFIG_HOME.
func GetConfigPath() (string, error) {
	if dir := os.Getenv(configDirEnv); dir != "" {
		return configPathFromDir(dir), nil
	}
	cfgHome, err := xdg.ConfigHome()
	if err != nil {
		return "", err
	}

	return configPathFromDir(filepath.Join(cfgHome, "sidetable")), nil
}

func configPathFromDir(dir string) string {
	cleanDir := filepath.Clean(dir)
	return filepath.Join(cleanDir, "config.yml")
}

// Load reads and validates config from path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.ConfigDir = filepath.Dir(filepath.Clean(path))

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate ensures config follows the specification.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Directory) == "" {
		return ErrDirectoryRequired
	}
	if filepath.IsAbs(c.Directory) {
		return ErrDirectoryMustBeRelative
	}
	if len(c.Commands) == 0 {
		return ErrCommandsMissing
	}

	if err := c.validateCommands(); err != nil {
		return err
	}
	if err := c.validateAliases(); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateCommands() error {
	for name, cmd := range c.Commands {
		if strings.TrimSpace(cmd.Command) == "" {
			return fmt.Errorf("command %q: %w", name, ErrCommandRequired)
		}
		if strings.ContainsAny(cmd.Command, " \t\n\r") {
			return fmt.Errorf("command %q: %w", name, ErrCommandMustNotContainSpaces)
		}
		if cmd.LegacyAlias != nil {
			return fmt.Errorf("command %q: %w", name, ErrLegacyCommandAliasRemoved)
		}
	}

	return nil
}

func (c *Config) validateAliases() error {
	for aliasName, alias := range c.Aliases {
		if strings.TrimSpace(aliasName) == "" {
			return fmt.Errorf("alias %q: %w", aliasName, ErrAliasNameRequired)
		}
		if strings.ContainsAny(aliasName, " \t\n\r") {
			return fmt.Errorf("alias %q: %w", aliasName, ErrAliasMustNotContainSpaces)
		}
		if strings.TrimSpace(alias.Command) == "" {
			return fmt.Errorf("alias %q: %w", aliasName, ErrAliasCommandRequired)
		}
		if _, exists := c.Commands[aliasName]; exists {
			return fmt.Errorf("alias %q: %w", aliasName, ErrAliasConflictsWithCommand)
		}
		if _, exists := c.Commands[alias.Command]; !exists {
			return fmt.Errorf("alias %q: %w", aliasName, ErrAliasTargetUnknown)
		}
	}

	return nil
}

// ResolveCommand resolves command by name or alias.
func (c *Config) ResolveCommand(name string) (*ResolvedCommand, error) {
	if cmd, ok := c.Commands[name]; ok {
		return &ResolvedCommand{
			Name:        name,
			Command:     cmd,
			AliasName:   "",
			AliasArgs:   nil,
			DisplayName: name,
		}, nil
	}
	alias, ok := c.Aliases[name]
	if !ok {
		return nil, ErrCommandUnknown
	}
	cmd, ok := c.Commands[alias.Command]
	if !ok {
		return nil, ErrCommandUnknown
	}

	return &ResolvedCommand{
		Name:        alias.Command,
		Command:     cmd,
		AliasName:   name,
		AliasArgs:   &alias.Args,
		DisplayName: name,
	}, nil
}

// CommandNames returns sorted command names.
func (c *Config) CommandNames() []string {
	names := make([]string, 0, len(c.Commands))
	for name := range c.Commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
