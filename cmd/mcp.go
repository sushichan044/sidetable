package cmd

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	sidetable "github.com/sushichan044/sidetable"
	internalmcp "github.com/sushichan044/sidetable/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:          "mcp",
	Short:        "Start a stdio MCP server exposing sidetable tools",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		workspace, err := sidetable.Open(cwd)
		if err != nil {
			return err
		}

		catalog, err := workspace.Catalog()
		if err != nil {
			return err
		}

		tools := make([]internalmcp.ToolDef, 0, len(catalog.Entries))
		for _, e := range catalog.Entries {
			if e.Kind != sidetable.EntryKindTool {
				continue
			}
			desc := e.Instructions
			if desc == "" {
				desc = e.Description
			}
			tools = append(tools, internalmcp.ToolDef{Name: e.Name, Description: desc})
		}

		executor := func(ctx context.Context, name string, args []string) (string, string, error) {
			var stdoutBuf, stderrBuf bytes.Buffer
			runErr := workspace.Run(ctx, name, args, sidetable.InvokeOptions{
				Stdin:  strings.NewReader(""),
				Stdout: &stdoutBuf,
				Stderr: &stderrBuf,
			})
			return stdoutBuf.String(), stderrBuf.String(), runErr
		}

		server := internalmcp.BuildServer(tools, executor)
		return server.Run(cmd.Context(), &sdkmcp.StdioTransport{})
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
