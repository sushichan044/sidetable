package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sushichan044/sidetable/internal/xdg"
)

var (
	ErrConfigNotFound       = errors.New("config file not found")
	ErrConfigBothExist      = errors.New("config.yml and config.yaml both exist")
	ErrCommandNotFound      = errors.New("command not found")
	ErrCommandsMissing      = errors.New("commands is required")
	ErrDirectoryMissing     = errors.New("directory is required")
	ErrDirectoryAbsolute    = errors.New("directory must be relative")
	ErrCommandMissing       = errors.New("command is required")
	ErrCommandHasSpace      = errors.New("command must not contain spaces")
	ErrAliasDuplicate       = errors.New("alias is duplicated")
	ErrAliasCommandConflict = errors.New("alias conflicts with command name")
)

// Config represents configuration file structure.
type Config struct {
	Directory string             `yaml:"directory"`
	Commands  map[string]Command `yaml:"commands"`
}

// Command represents a delegated command configuration.
type Command struct {
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args"`
	Env         map[string]string `yaml:"env"`
	Description string            `yaml:"description"`
	Alias       string            `yaml:"alias"`
}

// ResolvePath returns the config path resolved from XDG_CONFIG_HOME.
func ResolvePath() (string, error) {
	cfgHome, err := xdg.ConfigHome()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(cfgHome, "sidetable")
	yamlPath := filepath.Join(dir, "config.yaml")
	ymlPath := filepath.Join(dir, "config.yml")

	yamlExists := fileExists(yamlPath)
	ymlExists := fileExists(ymlPath)

	switch {
	case yamlExists && ymlExists:
		return "", ErrConfigBothExist
	case yamlExists:
		return yamlPath, nil
	case ymlExists:
		return ymlPath, nil
	default:
		return "", ErrConfigNotFound
	}
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

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate ensures config follows the specification.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Directory) == "" {
		return ErrDirectoryMissing
	}
	if filepath.IsAbs(c.Directory) {
		return ErrDirectoryAbsolute
	}
	if len(c.Commands) == 0 {
		return ErrCommandsMissing
	}

	aliasSeen := map[string]struct{}{}
	for name, cmd := range c.Commands {
		if strings.TrimSpace(cmd.Command) == "" {
			return fmt.Errorf("command %q: %w", name, ErrCommandMissing)
		}
		if strings.ContainsAny(cmd.Command, " \t\n\r") {
			return fmt.Errorf("command %q: %w", name, ErrCommandHasSpace)
		}
		if cmd.Alias == "" {
			continue
		}
		if _, exists := c.Commands[cmd.Alias]; exists {
			return fmt.Errorf("command %q: %w", name, ErrAliasCommandConflict)
		}
		if _, exists := aliasSeen[cmd.Alias]; exists {
			return fmt.Errorf("command %q: %w", name, ErrAliasDuplicate)
		}
		aliasSeen[cmd.Alias] = struct{}{}
	}

	return nil
}

// ResolveCommand resolves command by name or alias.
func (c *Config) ResolveCommand(name string) (string, Command, error) {
	if cmd, ok := c.Commands[name]; ok {
		return name, cmd, nil
	}
	for cmdName, cmd := range c.Commands {
		if cmd.Alias == name {
			return cmdName, cmd, nil
		}
	}
	return "", Command{}, ErrCommandNotFound
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
