package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/sushichan044/sidetable"
	"github.com/sushichan044/sidetable/internal/spacing"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available commands",
	Long: `List available commands defined in the sidetable configuration for the current project.

The output shows command/alias name, kind, target, and description for each configured command.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		project, err := sidetable.NewProject(cwd)
		if err != nil {
			return err
		}

		cmds, listErr := project.ListCommands()
		if listErr != nil {
			return listErr
		}

		formatter := spacing.NewFormatter(
			spacing.Column(), // Command Name
			//nolint:mnd // Justification: fixed spacing value for better readability
			spacing.MinSpacing(2),
			spacing.Column(), // Kind
			//nolint:mnd // Justification: fixed spacing value for better readability
			spacing.MinSpacing(4),
			spacing.Column(), // Target
			//nolint:mnd // Justification: fixed spacing value for better readability
			spacing.MinSpacing(4),
			spacing.Column(), // Description
		)

		rows := make([][]string, 0, len(cmds.Commands)+len(cmds.Aliases)+len(cmds.Invalid))
		for _, info := range cmds.Commands {
			rows = append(rows, []string{info.Name, "command", "-", info.Description})
		}
		for _, info := range cmds.Aliases {
			rows = append(rows, []string{info.Name, "alias", info.Target, info.Description})
		}
		for _, info := range cmds.Invalid {
			rows = append(rows, []string{
				info.Name,
				"invalid " + info.Kind,
				info.Target,
				info.Description + " (" + info.Reason + ")",
			})
		}
		if fmtErr := formatter.AddRows(rows...); fmtErr != nil {
			return fmtErr
		}

		return formatter.Println(cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
