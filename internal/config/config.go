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
	ErrConfigNotFound       = errors.New("config.yml file not found")
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
	Directory string             `yaml:"directory"` // User's private directory per project
	Commands  map[string]Command `yaml:"commands"`
	ConfigDir string             `yaml:"-"` // INTERNAL: Directory of the loaded config file
}

// Command represents a delegated command configuration.
type Command struct {
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args"`
	Env         map[string]string `yaml:"env"`
	Description string            `yaml:"description"`
	Alias       string            `yaml:"alias"`
}

const configDirEnv = "SIDETABLE_CONFIG_DIR"

// ResolvePath returns the config path resolved from SIDETABLE_CONFIG_DIR or XDG_CONFIG_HOME.
func ResolvePath() (string, error) {
	if dir := os.Getenv(configDirEnv); dir != "" {
		return resolvePathFromDir(dir)
	}
	cfgHome, err := xdg.ConfigHome()
	if err != nil {
		return "", err
	}
	return resolvePathFromDir(filepath.Join(cfgHome, "sidetable"))
}

func resolvePathFromDir(dir string) (string, error) {
	cleanDir := filepath.Clean(dir)
	ymlPath := filepath.Join(cleanDir, "config.yml")

	if ymlExists := fileExists(ymlPath); ymlExists {
		return ymlPath, nil
	}
	return "", fmt.Errorf("%w: looked for %q", ErrConfigNotFound, ymlPath)
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
