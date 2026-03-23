package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/sushichan044/sidetable/version"
)

// ToolExecutor runs a named tool with arguments and returns captured output.
// Returns (stdout, stderr, error).
type ToolExecutor func(ctx context.Context, name string, args []string) (string, string, error)

// ToolDef is the minimal information needed to register a sidetable tool as an MCP tool.
type ToolDef struct {
	Name        string
	Description string
}

// ToolInput is the MCP input schema for sidetable tools.
type ToolInput struct {
	Args []string `json:"args" jsonschema:"arguments to pass to the tool"`
}

// BuildServer constructs an MCP server populated with the given tools.
// The executor is called for each tool invocation.
func BuildServer(tools []ToolDef, executor ToolExecutor) *sdkmcp.Server {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "sidetable",
		Version: version.Get(),
	}, nil)

	for _, tool := range tools {
		toolName := tool.Name
		desc := tool.Description

		sdkmcp.AddTool(server, &sdkmcp.Tool{
			Name:        toolName,
			Description: desc,
		}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, input ToolInput) (*sdkmcp.CallToolResult, any, error) {
			return runWithExecutor(ctx, executor, toolName, input.Args)
		})
	}

	return server
}

func runWithExecutor(
	ctx context.Context,
	executor ToolExecutor,
	name string,
	args []string,
) (*sdkmcp.CallToolResult, any, error) {
	stdout, stderr, err := executor(ctx, name, args)

	if err != nil {
		var result sdkmcp.CallToolResult
		result.SetError(err)
		if stderr != "" {
			result.Content = append(result.Content, &sdkmcp.TextContent{Text: "stderr:\n" + stderr})
		}
		return &result, nil, nil
	}

	var contents []sdkmcp.Content
	if stdout != "" {
		contents = append(contents, &sdkmcp.TextContent{Text: stdout})
	}
	if stderr != "" {
		contents = append(contents, &sdkmcp.TextContent{Text: "stderr:\n" + stderr})
	}
	if len(contents) == 0 {
		contents = append(contents, &sdkmcp.TextContent{Text: "(no output)"})
	}

	return &sdkmcp.CallToolResult{Content: contents}, nil, nil
}
