package sidetable

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/sushichan044/sidetable/internal/config"
)

// Invocation is a fully resolved process invocation.
type Invocation struct {
	Program string
	Args    []string
	Env     []string
}

// InvokeOptions configures process execution.
type InvokeOptions struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// InvocationError represents a process that exited with non-zero status.
type InvocationError struct {
	Code int
	Err  error
}

func (e *InvocationError) Error() string {
	return fmt.Sprintf("invocation failed with exit code %d: %v", e.Code, e.Err)
}

func (e *InvocationError) Unwrap() error {
	return e.Err
}

// AsInvocationError extracts InvocationError from err.
//
//	if invErr, ok := sidetable.AsInvocationError(err); ok {
//	    fmt.Printf("Exit code: %d\n", invErr.Code)
//	}
func AsInvocationError(err error) (*InvocationError, bool) {
	if err == nil {
		return nil, false
	}
	if invErr := new(InvocationError); errors.As(err, &invErr) {
		return invErr, true
	}

	return nil, false
}

var (
	errRunTemplateEmpty    = errors.New("run template resolved to empty")
	errRunTemplateHasSpace = errors.New("run template contains spaces")
)

func resolveInvocation(
	cfg *config.Config,
	entryName string,
	userArgs []string,
	workspaceRoot string,
	baseEnv []string,
) (Invocation, error) {
	resolved, err := cfg.ResolveEntry(entryName)
	if err != nil {
		return Invocation{}, err
	}

	ctx := templateContext{
		WorkspaceRoot: workspaceRoot,
		ToolDir:       filepath.Join(workspaceRoot, cfg.Directory, resolved.ToolName),
		ConfigDir:     filepath.Dir(cfg.FilePath),
	}

	program, err := evalTemplate(resolved.Tool.Run, ctx)
	if err != nil {
		return Invocation{}, fmt.Errorf("run: %w", err)
	}
	if strings.TrimSpace(program) == "" {
		return Invocation{}, errRunTemplateEmpty
	}
	if strings.ContainsAny(program, " \t\n\r") {
		return Invocation{}, errRunTemplateHasSpace
	}

	resolvedArgs, err := buildArgsWithAlias(resolved.Tool.Args, resolved.AliasArgs, userArgs, ctx)
	if err != nil {
		return Invocation{}, err
	}

	envMap, err := buildEnvMap(baseEnv, resolved.Tool.Env, ctx)
	if err != nil {
		return Invocation{}, err
	}
	env := envSliceFromMap(envMap)

	return Invocation{
		Program: program,
		Args:    resolvedArgs,
		Env:     env,
	}, nil
}

func buildArgsWithAlias(
	toolArgs config.Args,
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

	toolPrepend, err := buildArgList(toolArgs.Prepend, ctx)
	if err != nil {
		return nil, fmt.Errorf("tool prepend: %w", err)
	}
	toolAppend, err := buildArgList(toolArgs.Append, ctx)
	if err != nil {
		return nil, fmt.Errorf("tool append: %w", err)
	}

	totalLen := len(aliasPrepend) + len(toolPrepend) + len(userArgs) + len(toolAppend) + len(aliasAppend)
	result := make([]string, 0, totalLen)
	result = append(result, aliasPrepend...)
	result = append(result, toolPrepend...)
	result = append(result, userArgs...)
	result = append(result, toolAppend...)
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

// buildEnvMap evaluates tool environment variables and merges them with the base environment.
//
// baseEnv is typically `os.Environ()` or the parent process environment.
// baseEnv is not handled as template.
func buildEnvMap(baseEnv []string, toolEnv map[string]string, ctx templateContext) (map[string]string, error) {
	baseEnvMap := envMapFromSlice(baseEnv)

	evaluatedToolEnv := make(map[string]string, len(toolEnv))
	for key, raw := range toolEnv {
		value, err := evalTemplate(raw, ctx)
		if err != nil {
			return nil, err
		}
		evaluatedToolEnv[key] = value
	}

	merged := mergeMap(baseEnvMap, evaluatedToolEnv)
	return merged, nil
}
