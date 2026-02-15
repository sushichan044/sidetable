package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/sushichan044/sidetable"
	"github.com/sushichan044/sidetable/internal/spacing"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available tools and aliases",
	Long: `List available tools and aliases defined in the sidetable configuration for the current project.

The output shows entry name, kind, target, and description for each configured entry.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		workspace, err := sidetable.Open(cwd)
		if err != nil {
			return err
		}

		catalog, catalogErr := workspace.Catalog()
		if catalogErr != nil {
			return catalogErr
		}

		formatter := spacing.NewFormatter(
			spacing.Column(), // Entry name
			//nolint:mnd // fixed spacing value for readability
			spacing.MinSpacing(2),
			spacing.Column(), // Kind
			//nolint:mnd // fixed spacing value for readability
			spacing.MinSpacing(4),
			spacing.Column(), // Target
			//nolint:mnd // fixed spacing value for readability
			spacing.MinSpacing(4),
			spacing.Column(), // Description
		)

		rows := make([][]string, 0, len(catalog.Entries)+1)
		rows = append(rows, []string{"NAME", "KIND", "TARGET", "DESCRIPTION"})
		for _, entry := range catalog.Entries {
			target := "-"
			if entry.Kind == sidetable.EntryKindAlias {
				target = entry.Target
			}
			rows = append(rows, []string{entry.Name, string(entry.Kind), target, entry.Description})
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
