package builtin_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/builtin"
)

func TestIsReservedName(t *testing.T) {
	for _, name := range []string{"list", "completion", "init", "help", "mcp"} {
		require.True(t, builtin.IsReservedName(name), "expected %q to be reserved", name)
	}
	require.False(t, builtin.IsReservedName("ghq"))
}
