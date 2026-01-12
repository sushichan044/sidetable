package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/sushichan044/sidetable/pkg/sidetable"
	"github.com/sushichan044/sidetable/utils/spacing"
	"github.com/sushichan044/sidetable/version"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	root := newRootCommand(os.Stdout, os.Stderr, cwd)
	root.SetArgs(os.Args[1:])
	if cliErr := root.Execute(); cliErr != nil {
		if exitErr := sidetable.ExtractExitError(cliErr); exitErr != nil {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
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

	root.AddCommand(newDoctorCommand(cwd))

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

func newListCommand(project *sidetable.Project) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available commands",
		RunE: func(cmd *cobra.Command, _ []string) error {
			commands, listErr := project.ListCommands()
			if listErr != nil {
				return listErr
			}

			formatter := spacing.NewFormatter(
				spacing.Column(), // Command Name
				//nolint:mnd // Justification: fixed spacing value for better readability
				spacing.MinSpacing(2),
				spacing.Column(), // Alias
				//nolint:mnd // Justification: fixed spacing value for better readability
				spacing.MinSpacing(4),
				spacing.Column(), // Description
			)

			rows := make([][]string, 0, len(commands))
			for _, info := range commands {
				alias := info.Alias
				if alias != "" {
					alias = fmt.Sprintf("(%s)", alias)
				}
				rows = append(rows, []string{info.Name, alias, info.Description})
			}

			if err := formatter.AddRows(rows...); err != nil {
				return err
			}
			return formatter.Format(cmd.OutOrStdout())
		},
	}
}

func newDoctorCommand(cwd string) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check configuration for issues",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Running sidetable doctor...")

			project, projectErr := sidetable.NewProject(cwd)
			if projectErr != nil {
				return fmt.Errorf("⚠️  cannot initialize project: %w", projectErr)
			}

			var errs []error
			cmds, listErr := project.ListCommands()
			if listErr != nil {
				errs = append(errs, fmt.Errorf("⚠️  cannot list commands: %w", listErr))
			}

			for _, info := range cmds {
				if isBuiltinCommand(info.Name) {
					errs = append(errs, fmt.Errorf("⚠️  command %q conflicts with builtin command", info.Name))
				}
				if info.Alias != "" && isBuiltinCommand(info.Alias) {
					errs = append(errs, fmt.Errorf("⚠️  command alias %q conflicts with builtin command", info.Alias))
				}
			}

			if len(errs) > 0 {
				return errors.Join(errs...)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "✅  no issues found")
			return nil
		},
	}
}
