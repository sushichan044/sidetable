package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sushichan044/sidetable"
	"github.com/sushichan044/sidetable/version"
)

var rootCmd = &cobra.Command{
	Use:   "sidetable",
	Short: "Personal directory manager per project",
	Long: `sidetable manages a project-local private area and runs commands defined in your project configuration.

Define commands in config.yml, then execute them as "sidetable <command> [args...]".
Use "sidetable list" to inspect available commands and "sidetable doctor" to validate configuration.`,
	SilenceUsage: true,
	Version:      version.Get(),
}

// Execute executes the root command and returns the exit code.
func Execute() int {
	if err := injectUserDefinedCommands(); err != nil {
		fmt.Fprintln(
			os.Stderr,
			"Error occurred while loading user-defined commands. Run `sidetable doctor` to diagnose the problem.",
		)
	}

	err := rootCmd.Execute()
	return determineExitCode(err)
}

func determineExitCode(err error) int {
	if err == nil {
		return 0
	}

	execErr := sidetable.GetExecError(err)
	if execErr != nil {
		return execErr.Code
	}

	return 1
}

func buildProjectCommands(project *sidetable.Project) []*cobra.Command {
	commands, err := project.ListCommands()
	if err != nil {
		return nil
	}

	cmds := make([]*cobra.Command, 0, len(commands))
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

func injectUserDefinedCommands() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	project, err := sidetable.NewProject(cwd)
	if err != nil {
		return err
	}

	subCommands := buildProjectCommands(project)
	rootCmd.AddCommand(subCommands...)

	return nil
}
