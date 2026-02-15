package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/sushichan044/sidetable/internal/builtin"
	"github.com/sushichan044/sidetable/internal/xdg"
)

var (
	ErrConfigMissing              = errors.New("config.yml file not found")
	ErrToolsMissing               = errors.New("tools are required")
	ErrEntryUnknown               = errors.New("entry not found")
	ErrDirectoryRequired          = errors.New("directory is required")
	ErrDirectoryMustBeRelative    = errors.New("directory must be relative")
	ErrToolRunRequired            = errors.New("tool run is required")
	ErrToolRunMustNotContainSpace = errors.New("tool run must not contain spaces")
	ErrToolConflictsWithBuiltin   = errors.New("tool conflicts with builtin command")
	ErrAliasConflictsWithTool     = errors.New("alias conflicts with tool name")
	ErrAliasConflictsWithBuiltin  = errors.New("alias conflicts with builtin command")
	ErrAliasNameRequired          = errors.New("alias name is required")
	ErrAliasMustNotContainSpaces  = errors.New("alias must not contain spaces")
	ErrAliasToolRequired          = errors.New("alias tool is required")
	ErrAliasTargetUnknown         = errors.New("alias tool not found")
	ErrLegacyCommandsRemoved      = errors.New("top-level commands has been removed; use tools")
	ErrLegacyToolCommandRemoved   = errors.New("tools.<name>.command has been removed; use run")
	ErrLegacyAliasCommandRemoved  = errors.New("aliases.<name>.command has been removed; use tool")
)

// Config represents configuration file structure.
type Config struct {
	Directory      string           `yaml:"directory"`
	Tools          map[string]Tool  `yaml:"tools"`
	Aliases        map[string]Alias `yaml:"aliases"`
	LegacyCommands map[string]any   `yaml:"commands"`
	ConfigDir      string           `yaml:"-"`
}

// Tool represents a tool definition.
type Tool struct {
	Run           string            `yaml:"run"`
	Args          Args              `yaml:"args"`
	Env           map[string]string `yaml:"env"`
	Description   string            `yaml:"description"`
	LegacyCommand *string           `yaml:"command"`
}

// Alias represents an alias definition.
type Alias struct {
	Tool          string  `yaml:"tool"`
	Args          Args    `yaml:"args"`
	Description   string  `yaml:"description"`
	LegacyCommand *string `yaml:"command"`
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
	errs := make([]error, 0)

	if strings.TrimSpace(c.Directory) == "" {
		errs = append(errs, ErrDirectoryRequired)
	}
	if filepath.IsAbs(c.Directory) {
		errs = append(errs, ErrDirectoryMustBeRelative)
	}
	if len(c.LegacyCommands) > 0 {
		errs = append(errs, ErrLegacyCommandsRemoved)
	}
	if len(c.Tools) == 0 {
		errs = append(errs, ErrToolsMissing)
	}

	errs = append(errs, c.validateTools()...)
	errs = append(errs, c.validateAliases()...)

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func (c *Config) validateTools() []error {
	if len(c.Tools) == 0 {
		return nil
	}

	toolNames := c.ToolNames()
	errs := make([]error, 0)
	for _, name := range toolNames {
		tool := c.Tools[name]
		if tool.LegacyCommand != nil {
			errs = append(errs, fmt.Errorf("tool %q: %w", name, ErrLegacyToolCommandRemoved))
		}
		if strings.TrimSpace(tool.Run) == "" {
			errs = append(errs, fmt.Errorf("tool %q: %w", name, ErrToolRunRequired))
		}
		if strings.ContainsAny(tool.Run, " \t\n\r") {
			errs = append(errs, fmt.Errorf("tool %q: %w", name, ErrToolRunMustNotContainSpace))
		}
		if builtin.IsReservedName(name) {
			errs = append(errs, fmt.Errorf("tool %q: %w", name, ErrToolConflictsWithBuiltin))
		}
	}

	return errs
}

func (c *Config) validateAliases() []error {
	if len(c.Aliases) == 0 {
		return nil
	}

	aliasNames := make([]string, 0, len(c.Aliases))
	for aliasName := range c.Aliases {
		aliasNames = append(aliasNames, aliasName)
	}
	sort.Strings(aliasNames)

	errs := make([]error, 0)
	for _, aliasName := range aliasNames {
		alias := c.Aliases[aliasName]
		if strings.TrimSpace(aliasName) == "" {
			errs = append(errs, fmt.Errorf("alias %q: %w", aliasName, ErrAliasNameRequired))
		}
		if strings.ContainsAny(aliasName, " \t\n\r") {
			errs = append(errs, fmt.Errorf("alias %q: %w", aliasName, ErrAliasMustNotContainSpaces))
		}
		if alias.LegacyCommand != nil {
			errs = append(errs, fmt.Errorf("alias %q: %w", aliasName, ErrLegacyAliasCommandRemoved))
		}
		aliasTool := strings.TrimSpace(alias.Tool)
		if aliasTool == "" {
			errs = append(errs, fmt.Errorf("alias %q: %w", aliasName, ErrAliasToolRequired))
		}
		if _, exists := c.Tools[aliasName]; exists {
			errs = append(errs, fmt.Errorf("alias %q: %w", aliasName, ErrAliasConflictsWithTool))
		}
		if builtin.IsReservedName(aliasName) {
			errs = append(errs, fmt.Errorf("alias %q: %w", aliasName, ErrAliasConflictsWithBuiltin))
		}
		if aliasTool != "" {
			if _, exists := c.Tools[aliasTool]; !exists {
				errs = append(errs, fmt.Errorf("alias %q: %w", aliasName, ErrAliasTargetUnknown))
			}
		}
	}

	return errs
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
