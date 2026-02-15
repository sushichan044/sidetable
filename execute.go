package sidetable

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

// execute executes an already resolved invocation.
func (w *Workspace) execute(ctx context.Context, inv Invocation, opts InvokeOptions) error {
	if w == nil {
		return errors.New("workspace is not initialized")
	}
	if inv.Program == "" {
		return errors.New("invocation program is empty")
	}

	// Execute the invocation.
	if ctx == nil {
		ctx = context.Background()
	}

	// #nosec G204 -- command/args are from user-owned config; explicit delegation is intended.
	cmd := exec.CommandContext(ctx, inv.Program, inv.Args...)
	cmd.Env = inv.Env
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	} else {
		cmd.Stdin = os.Stdin
	}
	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		if exitErr := new(exec.ExitError); errors.As(err, &exitErr) {
			return &InvocationError{Code: exitErr.ExitCode(), Err: err}
		}
		return err
	}
	return nil
}
