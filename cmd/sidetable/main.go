package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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
	showVersion, showHelp, remaining, err := parseGlobalFlags(args)
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if showHelp {
		return executeBuiltin([]string{"--help"}, stdout, stderr, cwd)
	}

	if showVersion {
		return executeBuiltin([]string{"version"}, stdout, stderr, cwd)
	}

	// Handle built-in commands: no args, help, version, list
	// Pass only remaining args to avoid global flag interference with Cobra's flag parsing
	if len(remaining) == 0 || isBuiltIn(remaining[0]) {
		return executeBuiltin(remaining, stdout, stderr, cwd)
	}

	project, err := sidetable.NewProject(cwd)
	if err != nil {
		return err
	}

	action, err := project.BuildAction(remaining[0], remaining[1:])
	if err != nil {
		return err
	}

	return project.Execute(action)
}

func parseGlobalFlags(args []string) (bool, bool, []string, error) {
	fs := pflag.NewFlagSet("sidetable", pflag.ContinueOnError)
	fs.SetInterspersed(false)
	fs.SetOutput(io.Discard)
	fs.ParseErrorsAllowlist.UnknownFlags = true

	var showVersion bool
	var showHelp bool
	fs.BoolVarP(&showVersion, "version", "v", false, "show version")
	fs.BoolVarP(&showHelp, "help", "h", false, "show help")

	if err := fs.Parse(args); err != nil {
		return false, false, nil, err
	}

	return showVersion, showHelp, fs.Args(), nil
}

func executeBuiltin(args []string, stdout, stderr io.Writer, cwd string) error {
	root := newRootCommand(stdout, stderr, cwd)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		return err
	}
	return nil
}

func newRootCommand(stdout, stderr io.Writer, cwd string) *cobra.Command {
	root := &cobra.Command{
		Use:           "sidetable",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)

	root.AddCommand(newListCommand(cwd))
	root.AddCommand(newVersionCommand(stdout))

	return root
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

func newListCommand(projectDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available commands",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := sidetable.NewProject(projectDir)
			if err != nil {
				return err
			}
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

func isBuiltIn(name string) bool {
	switch name {
	case "help", "list", "version":
		return true
	default:
		return false
	}
}
