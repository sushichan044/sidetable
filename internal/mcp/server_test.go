package mcp_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/sushichan044/sidetable/internal/config"
	internalmcp "github.com/sushichan044/sidetable/internal/mcp"
)

func TestToolDescription(t *testing.T) {
	t.Run("prefers instructions over description", func(t *testing.T) {
		tool := config.Tool{Description: "short", Instructions: "long AI instructions"}
		require.Equal(t, "long AI instructions", internalmcp.ToolDescription(tool))
	})
	t.Run("falls back to description when instructions empty", func(t *testing.T) {
		tool := config.Tool{Description: "short", Instructions: ""}
		require.Equal(t, "short", internalmcp.ToolDescription(tool))
	})
	t.Run("returns empty string when both empty", func(t *testing.T) {
		tool := config.Tool{}
		require.Empty(t, internalmcp.ToolDescription(tool))
	})
}

func testTools() map[string]config.Tool {
	return map[string]config.Tool{
		"my-tool": {Run: "echo", Description: "A tool"},
	}
}

func TestBuildServer_RegistersOnlyTools(t *testing.T) {
	executor := func(_ context.Context, _ string, _ []string) (string, string, error) {
		return "output", "", nil
	}
	server := internalmcp.BuildServer(testTools(), executor)

	ctx := context.Background()
	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	result, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)

	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	require.Contains(t, names, "my-tool")
}

func TestBuildServer_ToolHandler_ReturnsOutput(t *testing.T) {
	executor := func(_ context.Context, _ string, _ []string) (string, string, error) {
		return "hello world\n", "", nil
	}
	server := internalmcp.BuildServer(testTools(), executor)

	ctx := context.Background()
	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	callResult, err := clientSession.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "my-tool",
		Arguments: map[string]any{"args": []any{"hello"}},
	})
	require.NoError(t, err)
	require.False(t, callResult.IsError)
	require.Len(t, callResult.Content, 1)
	textContent, ok := callResult.Content[0].(*sdkmcp.TextContent)
	require.True(t, ok)
	require.Equal(t, "hello world\n", textContent.Text)
}
