package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/sushichan044/sidetable/internal/action"
	"github.com/sushichan044/sidetable/pkg/sidetable"
	"github.com/sushichan044/sidetable/version"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)

		var exitErr *action.ExitError
		// NOTE: use error.AsType after Go 1.26 released
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	root := newRootCommand(stdout, stderr, cwd)
	root.SetArgs(args)
	return root.Execute()
}

func newRootCommand(stdout, stderr io.Writer, cwd string) *cobra.Command {
	root := &cobra.Command{
		Use:          "sidetable",
		Short:        "Personal directory manager per project",
		SilenceUsage: true,
		Version:      version.Get(),
	}
	root.SetOut(stdout)
	root.SetErr(stderr)

	root.AddCommand(newVersionCommand(stdout))

	// Try to load project and add dynamic commands
	// If project initialization fails, only built-in commands are available (graceful fallback)
	project, err := sidetable.NewProject(cwd)
	if err == nil {
		root.AddCommand(newListCommand(project))
		root.AddCommand(buildProjectCommands(project)...)
	}

	return root
}

func buildProjectCommands(project *sidetable.Project) []*cobra.Command {
	commands, err := project.ListCommands()
	if err != nil {
		return nil
	}

	var cmds []*cobra.Command
	for _, info := range commands {
		if isBuiltinCommand(info.Name) {
			// Skip built-in commands to avoid conflict
			continue
		}

		subCmd := &cobra.Command{
			Use:                info.Name,
			Short:              info.Description,
			DisableFlagParsing: true,
			SilenceErrors:      true,
			SilenceUsage:       true,
			RunE: func(_ *cobra.Command, args []string) error {
				act, buildErr := project.BuildAction(info.Name, args)
				if buildErr != nil {
					return buildErr
				}
				return project.Execute(act)
			},
		}
		if info.Alias != "" {
			subCmd.Aliases = []string{info.Alias}
		}
		cmds = append(cmds, subCmd)
	}

	return cmds
}

func newVersionCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Fprintln(stdout, version.Get())
		},
	}
}

func newListCommand(project *sidetable.Project) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available commands",
		RunE: func(cmd *cobra.Command, _ []string) error {
			commands, err := project.ListCommands()
			if err != nil {
				return err
			}
			for _, entry := range commands {
				if entry.Alias != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\t%s\n", entry.Name, entry.Alias, entry.Description)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", entry.Name, entry.Description)
			}
			return nil
		},
	}
}
