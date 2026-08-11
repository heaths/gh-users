package terminal

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSpinner(t *testing.T) {
	output, err := os.CreateTemp(t.TempDir(), "spinner-output")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := output.Close(); err != nil {
			t.Errorf("closing spinner temp file: %v", err)
		}
	})

	defaultOutput := spinnerDefaultOutput
	isTerminal := spinnerIsTerminal
	colorEnabled := spinnerColorEnabled
	t.Cleanup(func() {
		spinnerDefaultOutput = defaultOutput
		spinnerIsTerminal = isTerminal
		spinnerColorEnabled = colorEnabled
	})

	spinnerDefaultOutput = func() *os.File { return output }
	spinnerIsTerminal = func(file *os.File) bool { return file == output }

	t.Run("returns nil when spinner is disabled by env", func(t *testing.T) {
		t.Setenv(spinnerDisabledEnv, "true")
		require.Nil(t, NewSpinner(nil))
	})

	t.Run("returns nil when output is not a file", func(t *testing.T) {
		t.Setenv(spinnerDisabledEnv, "")
		require.Nil(t, NewSpinner(&bytes.Buffer{}))
	})

	t.Run("returns nil when output is not a tty", func(t *testing.T) {
		t.Setenv(spinnerDisabledEnv, "")
		spinnerIsTerminal = func(*os.File) bool { return false }
		require.Nil(t, NewSpinner(output))
	})

	t.Run("defaults to stderr and configures cyan when color is enabled", func(t *testing.T) {
		t.Setenv(spinnerDisabledEnv, "")
		spinnerIsTerminal = func(file *os.File) bool { return file == output }
		spinnerColorEnabled = func() bool { return true }

		sp := NewSpinner(nil)
		require.NotNil(t, sp)
		require.Equal(t, output, sp.output)
		require.Equal(t, "fgCyan", sp.color)
	})

	t.Run("omits color when color is disabled", func(t *testing.T) {
		t.Setenv(spinnerDisabledEnv, "")
		spinnerIsTerminal = func(file *os.File) bool { return file == output }
		spinnerColorEnabled = func() bool { return false }

		sp := NewSpinner(output)
		require.NotNil(t, sp)
		require.Equal(t, "", sp.color)
		require.Equal(t, output, sp.output)
	})
}
