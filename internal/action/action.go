package action

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sushichan044/sidetable/internal/config"
)

var (
	ErrCommandTemplateEmpty    = errors.New("command template resolved to empty")
	ErrCommandTemplateHasSpace = errors.New("command template contains spaces")
)

// ExecError represents a command that exited with non-zero status.
type ExecError struct {
	Code int
	Err  error
}

func (e *ExecError) Error() string {
	return fmt.Sprintf("command exited with code %d", e.Code)
}

func (e *ExecError) Unwrap() error {
	return e.Err
}

func GetExecError(err error) *ExecError {
	if err == nil {
		return nil
	}

	// NOTE: use error.AsType after Go 1.26 released
	var execErr *ExecError
	if errors.As(err, &execErr) {
		return execErr
	}
	return nil
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

// Build resolves command/args/env based on config.
func Build(cfg *config.Config, name string, userArgs []string, projectDir string) (*Action, error) {
	resolved, err := cfg.ResolveCommand(name)
	if err != nil {
		return nil, err
	}

	cmd := resolved.Command
	privateDir := filepath.Join(projectDir, cfg.Directory)
	commandDir := filepath.Join(privateDir, resolved.Name)

	ctx := templateContext{
		ProjectDir: projectDir,
		PrivateDir: privateDir,
		CommandDir: commandDir,
		ConfigDir:  cfg.ConfigDir,
	}

	resolvedCmd, err := evalTemplate(cmd.Command, ctx)
	if err != nil {
		return nil, fmt.Errorf("command: %w", err)
	}
	if strings.TrimSpace(resolvedCmd) == "" {
		return nil, ErrCommandTemplateEmpty
	}
	if strings.ContainsAny(resolvedCmd, " \t\n\r") {
		return nil, ErrCommandTemplateHasSpace
	}

	resolvedArgs, err := buildArgsWithAlias(cmd.Args, resolved.AliasArgs, userArgs, ctx)
	if err != nil {
		return nil, err
	}

	env, err := buildEnvWithAlias(cmd.Env, resolved.AliasEnv, ctx)
	if err != nil {
		return nil, err
	}

	return &Action{
		Command:    resolvedCmd,
		Args:       resolvedArgs,
		Env:        env,
		ProjectDir: projectDir,
		PrivateDir: privateDir,
		CommandDir: commandDir,
	}, nil
}

// Execute runs the delegated command and returns its exit code.
func Execute(spec *Action) error {
	// #nosec G204 -- command/args are from user-owned config; explicit delegation is intended.
	cmd := exec.CommandContext(context.Background(), spec.Command, spec.Args...)
	cmd.Env = spec.Env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr := new(exec.ExitError); errors.As(err, &exitErr) {
			return &ExecError{
				Code: exitErr.ExitCode(),
				Err:  err,
			}
		}
		return err
	}
	return nil
}

type templateContext struct {
	ProjectDir string
	PrivateDir string
	CommandDir string
	ConfigDir  string
}

func evalTemplate(raw string, ctx templateContext) (string, error) {
	tpl, err := template.New("value").Option("missingkey=error").Parse(raw)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if execErr := tpl.Execute(&b, ctx); execErr != nil {
		return "", execErr
	}
	return b.String(), nil
}

func buildArgsWithAlias(
	commandArgs config.Args,
	aliasArgs *config.Args,
	userArgs []string,
	ctx templateContext,
) ([]string, error) {
	var aliasPrepend, aliasAppend []string
	var err error

	if aliasArgs != nil {
		aliasPrepend, err = buildArgList(aliasArgs.Prepend, ctx)
		if err != nil {
			return nil, fmt.Errorf("alias prepend: %w", err)
		}
		aliasAppend, err = buildArgList(aliasArgs.Append, ctx)
		if err != nil {
			return nil, fmt.Errorf("alias append: %w", err)
		}
	}

	commandPrepend, err := buildArgList(commandArgs.Prepend, ctx)
	if err != nil {
		return nil, fmt.Errorf("command prepend: %w", err)
	}
	commandAppend, err := buildArgList(commandArgs.Append, ctx)
	if err != nil {
		return nil, fmt.Errorf("command append: %w", err)
	}

	totalLen := len(aliasPrepend) + len(commandPrepend) + len(userArgs) + len(commandAppend) + len(aliasAppend)
	result := make([]string, 0, totalLen)
	result = append(result, aliasPrepend...)
	result = append(result, commandPrepend...)
	result = append(result, userArgs...)
	result = append(result, commandAppend...)
	result = append(result, aliasAppend...)

	return result, nil
}

func buildArgList(args []string, ctx templateContext) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
	}
	result := make([]string, 0, len(args))
	for _, raw := range args {
		resolved, err := evalTemplate(raw, ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, resolved)
	}
	return result, nil
}

func buildEnvWithAlias(
	commandEnv map[string]string,
	aliasEnv map[string]string,
	ctx templateContext,
) ([]string, error) {
	base := os.Environ()
	if len(commandEnv) == 0 && len(aliasEnv) == 0 {
		return base, nil
	}

	overrides := make(map[string]string, len(commandEnv)+len(aliasEnv))
	for key, raw := range commandEnv {
		value, err := evalTemplate(raw, ctx)
		if err != nil {
			return nil, err
		}
		overrides[key] = value
	}
	for key, raw := range aliasEnv {
		value, err := evalTemplate(raw, ctx)
		if err != nil {
			return nil, err
		}
		overrides[key] = value
	}

	return mergeEnv(base, overrides), nil
}

func mergeEnv(base []string, overrides map[string]string) []string {
	result := make([]string, 0, len(base)+len(overrides))
	seen := make(map[string]struct{}, len(overrides))
	for _, entry := range base {
		key, _, ok := strings.Cut(entry, "=")
		if ok {
			if value, exists := overrides[key]; exists {
				result = append(result, key+"="+value)
				seen[key] = struct{}{}
				continue
			}
		}
		result = append(result, entry)
	}
	for key, value := range overrides {
		if _, exists := seen[key]; exists {
			continue
		}
		result = append(result, key+"="+value)
	}
	return result
}
