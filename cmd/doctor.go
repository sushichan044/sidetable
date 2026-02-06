package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sushichan044/sidetable"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check configuration for issues",
	Long: `Check sidetable configuration for common problems in the current project.

Doctor validates command names and aliases and reports conflicts with built-in commands such as "list", "doctor", and "completion".`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		project, err := sidetable.NewProject(cwd)
		if err != nil {
			return err
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
			for _, e := range errs {
				fmt.Fprintln(cmd.OutOrStdout(), e.Error())
			}
			return fmt.Errorf("doctor found %d issue(s)", len(errs))
		}

		fmt.Fprintln(cmd.OutOrStdout(), "✅  no issues found")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
