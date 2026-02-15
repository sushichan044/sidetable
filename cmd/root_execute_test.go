//nolint:testpackage // Need package-level access to root command wiring.
package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecuteShowsTemplateEvaluationError(t *testing.T) {
	configYAML := `directory: .sidetable
tools:
  cmd_tmpl_missing_field:
    run: "{{.Invalid}}"
`

	exitCode, stderr := runExecuteWithTempConfig(t, configYAML, "cmd_tmpl_missing_field")

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr, "Error:")
	require.Contains(t, stderr, "Invalid")
}

func TestExecutePreservesInvocationExitCodeAndPrintsError(t *testing.T) {
	configYAML := `directory: .sidetable
tools:
  cmd_fail_exit_42:
    run: sh
    args:
      prepend: ["-c", "exit 42"]
`

	exitCode, stderr := runExecuteWithTempConfig(t, configYAML, "cmd_fail_exit_42")

	require.Equal(t, 42, exitCode)
	require.Contains(t, stderr, "Error:")
	require.Contains(t, stderr, "invocation failed with exit code 42")
}

func runExecuteWithTempConfig(t *testing.T, configYAML string, args ...string) (int, string) {
	t.Helper()

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yml")
	require.NoError(t, os.WriteFile(configPath, []byte(configYAML), 0o600))
	t.Setenv("SIDETABLE_CONFIG_DIR", configDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	origOut := rootCmd.OutOrStdout()
	origErr := rootCmd.ErrOrStderr()
	t.Cleanup(func() {
		rootCmd.SetOut(origOut)
		rootCmd.SetErr(origErr)
		rootCmd.SetArgs(nil)
	})

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)

	return Execute(), stderr.String()
}
