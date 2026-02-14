//nolint:testpackage // Need package-level access to unexported helpers.
package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/action"
)

func TestDetermineExitCode(t *testing.T) {
	t.Run("nil error returns zero", func(t *testing.T) {
		require.Equal(t, 0, determineExitCode(nil))
	})

	t.Run("delegated exit code is preserved", func(t *testing.T) {
		err := &action.ExecError{
			Code: 42,
			Err:  errors.New("boom"),
		}
		require.Equal(t, 42, determineExitCode(err))
	})

	t.Run("non-delegated errors return one", func(t *testing.T) {
		require.Equal(t, 1, determineExitCode(errors.New("unexpected")))
	})
}
