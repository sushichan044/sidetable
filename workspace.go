package sidetable

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sushichan044/sidetable/internal/config"
)

// Workspace provides API access to sidetable.
type Workspace struct {
	config  *config.Config
	rootDir string
}

// Option configures Open.
type Option func(*workspaceOptions)

type workspaceOptions struct {
	configPath string
}

// WithConfigPath overrides config path resolution.
func WithConfigPath(path string) Option {
	return func(o *workspaceOptions) {
		o.configPath = path
	}
}

// Open loads config and prepares workspace context.
func Open(root string, opts ...Option) (*Workspace, error) {
	if root == "" {
		return nil, errors.New("root must not be empty")
	}
	root = filepath.Clean(root)

	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("root does not exist: %s", root)
		}
		return nil, fmt.Errorf("failed to stat root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root is not a directory: %s", root)
	}

	openOpts := workspaceOptions{}
	for _, opt := range opts {
		opt(&openOpts)
	}

	path := openOpts.configPath
	if path == "" {
		path, err = config.GetConfigPath()
		if err != nil {
			return nil, err
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	return &Workspace{config: cfg, rootDir: root}, nil
}

// Root returns the root directory for this workspace.
func (w *Workspace) Root() string {
	if w == nil {
		return ""
	}
	return w.rootDir
}

// Run resolves then executes a tool or alias.
func (w *Workspace) Run(ctx context.Context, name string, userArgs []string, opts InvokeOptions) error {
	if w == nil || w.config == nil {
		return errors.New("workspace is not initialized")
	}

	inv, err := resolveInvocation(w.config, name, userArgs, w.rootDir)
	if err != nil {
		return err
	}

	return w.execute(ctx, inv, opts)
}
