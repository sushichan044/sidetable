package config

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/goccy/go-yaml"

	"github.com/sushichan044/sidetable/internal/xdg"
)

var (
	ErrConfigMissing = errors.New("config.yml file not found")
	ErrEntryUnknown  = errors.New("entry not found")
)

// Config represents configuration file structure.
type Config struct {
	Directory string           `yaml:"directory"`
	Tools     map[string]Tool  `yaml:"tools"`
	Aliases   map[string]Alias `yaml:"aliases"`
	ConfigDir string           `yaml:"-"`
}

// Tool represents a tool definition.
type Tool struct {
	Run         string            `yaml:"run"`
	Args        Args              `yaml:"args"`
	Env         map[string]string `yaml:"env"`
	Description string            `yaml:"description"`
}

// Alias represents an alias definition.
type Alias struct {
	Tool        string `yaml:"tool"`
	Args        Args   `yaml:"args"`
	Description string `yaml:"description"`
}

// Args represents user-arg injection configuration.
type Args struct {
	Prepend []string `yaml:"prepend"`
	Append  []string `yaml:"append"`
}

// ResolvedEntry represents a resolved entry with optional alias information.
type ResolvedEntry struct {
	ToolName  string
	Tool      Tool
	AliasName string
	AliasArgs *Args
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
	issues := c.validateWithSchema()
	if len(issues) == 0 {
		return nil
	}

	errs := make([]error, 0, len(issues))
	for _, issue := range issues {
		errs = append(errs, newValidationIssueError(issue))
	}

	return errors.Join(errs...)
}

// ResolveEntry resolves a tool or alias name.
func (c *Config) ResolveEntry(name string) (*ResolvedEntry, error) {
	if tool, ok := c.Tools[name]; ok {
		return &ResolvedEntry{
			ToolName:  name,
			Tool:      tool,
			AliasName: "",
			AliasArgs: nil,
		}, nil
	}
	alias, ok := c.Aliases[name]
	if !ok {
		return nil, ErrEntryUnknown
	}
	tool, ok := c.Tools[alias.Tool]
	if !ok {
		return nil, ErrEntryUnknown
	}

	return &ResolvedEntry{
		ToolName:  alias.Tool,
		Tool:      tool,
		AliasName: name,
		AliasArgs: &alias.Args,
	}, nil
}

// ToolNames returns sorted tool names.
func (c *Config) ToolNames() []string {
	names := make([]string, 0, len(c.Tools))
	for name := range c.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
