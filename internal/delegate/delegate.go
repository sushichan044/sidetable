package delegate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/sushichan044/sidetable/internal/config"
)

var (
	ErrArgsPlaceholderMissing  = errors.New("args placeholder missing")
	ErrArgsPlaceholderMultiple = errors.New("args placeholder appears multiple times")
	ErrArgsPlaceholderWithText = errors.New("args placeholder must be standalone")
	ErrCommandTemplateEmpty    = errors.New("command template resolved to empty")
	ErrCommandTemplateHasSpace = errors.New("command template contains spaces")
)

var argsPlaceholder = regexp.MustCompile(`^{{\s*\.Args\s*}}$`)
var argsPlaceholderAny = regexp.MustCompile(`{{\s*\.Args\s*}}`)

// Spec is a resolved delegated command.
type Spec struct {
	Command    string
	Args       []string
	Env        []string
	ProjectDir string
	PrivateDir string
	CommandDir string
}

// Build resolves command/args/env based on config.
func Build(cfg *config.Config, name string, userArgs []string, projectDir string) (*Spec, error) {
	cmdName, cmd, err := cfg.ResolveCommand(name)
	if err != nil {
		return nil, err
	}

	privateDir := filepath.Join(projectDir, cfg.Directory)
	commandDir := filepath.Join(privateDir, cmdName)

	ctx := templateContext{
		ProjectDir: projectDir,
		PrivateDir: privateDir,
		CommandDir: commandDir,
		Args:       userArgs,
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

	resolvedArgs, err := buildArgs(cmd.Args, userArgs, ctx)
	if err != nil {
		return nil, err
	}

	env, err := buildEnv(cmd.Env, ctx)
	if err != nil {
		return nil, err
	}

	return &Spec{
		Command:    resolvedCmd,
		Args:       resolvedArgs,
		Env:        env,
		ProjectDir: projectDir,
		PrivateDir: privateDir,
		CommandDir: commandDir,
	}, nil
}

// Execute runs the delegated command and returns its exit code.
func Execute(spec *Spec) error {
	// #nosec G204 -- command/args are from user-owned config; explicit delegation is intended.
	cmd := exec.CommandContext(context.Background(), spec.Command, spec.Args...)
	cmd.Env = spec.Env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr := new(exec.ExitError); errors.As(err, &exitErr) {
			return fmt.Errorf("command exited with code %d", exitErr.ExitCode())
		}
		return err
	}
	return nil
}

type templateContext struct {
	ProjectDir string
	PrivateDir string
	CommandDir string
	Args       []string
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

func buildArgs(configArgs []string, userArgs []string, ctx templateContext) ([]string, error) {
	if configArgs == nil {
		return userArgs, nil
	}

	placeholderCount := 0
	for _, raw := range configArgs {
		if argsPlaceholderAny.MatchString(raw) {
			if !argsPlaceholder.MatchString(raw) {
				return nil, ErrArgsPlaceholderWithText
			}
			placeholderCount++
		}
	}

	switch {
	case placeholderCount > 1:
		return nil, ErrArgsPlaceholderMultiple
	case placeholderCount == 0 && len(userArgs) > 0:
		return nil, ErrArgsPlaceholderMissing
	}

	result := make([]string, 0, len(configArgs)+len(userArgs))
	for _, raw := range configArgs {
		if argsPlaceholder.MatchString(raw) {
			result = append(result, userArgs...)
			continue
		}
		resolved, err := evalTemplate(raw, ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, resolved)
	}
	return result, nil
}

func buildEnv(env map[string]string, ctx templateContext) ([]string, error) {
	base := os.Environ()
	if len(env) == 0 {
		return base, nil
	}

	overrides := make(map[string]string, len(env))
	for key, raw := range env {
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
