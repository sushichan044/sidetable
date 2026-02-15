package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sushichan044/sidetable"
	"github.com/sushichan044/sidetable/internal/errutils"
	"github.com/sushichan044/sidetable/version"
)

var rootCmd = &cobra.Command{
	Use:   "sidetable",
	Short: "Project-local tool workspace manager",
	Long: `sidetable manages a project-local tool area and runs tools defined in your configuration.

Define tools in config file, then execute them as "sidetable <tool-or-alias> [args...]".
Use "sidetable list" to inspect available entries.
Use "sidetable init" to scaffold a config file.`,
	SilenceUsage: true,
	Version:      version.Get(),
}

var injectedUserCommands []*cobra.Command

// Execute executes the root command and returns the exit code.
func Execute() int {
	if err := injectUserDefinedCommands(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Error occurred while loading config:"))

		// If config loading fails, print error details and continue.
		// This allows users to use built-in commands like "help" or "init" anytime.
		errs, _ := errutils.UnwrapJoinError(err)
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, color.RedString("- %v", e))
		}
		fmt.Fprintln(os.Stderr)
	}

	if err := rootCmd.Execute(); err != nil {
		return determineExitCode(err)
	}
	return 0
}

func determineExitCode(err error) int {
	if err == nil {
		return 0
	}

	invErr, ok := sidetable.AsInvocationError(err)
	if ok {
		return invErr.Code
	}

	return 1
}

func injectUserDefinedCommands() error {
	clearInjectedUserCommands()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	workspace, err := sidetable.Open(cwd)
	if err != nil {
		return err
	}

	subCommands, err := buildWorkspaceCommands(workspace)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(subCommands...)
	injectedUserCommands = subCommands

	return nil
}

func clearInjectedUserCommands() {
	if len(injectedUserCommands) == 0 {
		return
	}
	rootCmd.RemoveCommand(injectedUserCommands...)
	injectedUserCommands = nil
}

func buildWorkspaceCommands(workspace *sidetable.Workspace) ([]*cobra.Command, error) {
	catalog, err := workspace.Catalog()
	if err != nil {
		return nil, err
	}

	cmds := make([]*cobra.Command, 0, len(catalog.Entries))
	for _, entry := range catalog.Entries {
		name := entry.Name
		description := entry.Description
		subCmd := &cobra.Command{
			Use:                name,
			Short:              description,
			DisableFlagParsing: true,
			SilenceUsage:       true,
			RunE: func(_ *cobra.Command, args []string) error {
				return workspace.Run(context.Background(), name, args, sidetable.InvokeOptions{})
			},
		}
		cmds = append(cmds, subCmd)
	}

	return cmds, nil
}
