package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sushichan044/sidetable/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the sidetable configuration",
	RunE: func(cmd *cobra.Command, _ []string) error {
		path, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		if _, statErr := os.Stat(path); statErr == nil {
			return fmt.Errorf("config already exists: %s", path)
		} else if !os.IsNotExist(statErr) {
			return statErr
		}

		dir := filepath.Dir(path)
		err = os.MkdirAll(dir, 0o700)
		if err != nil {
			return err
		}

		err = os.WriteFile(path, config.DefaultConfigYAML, 0o600)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
