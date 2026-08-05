package main

import (
	"testing"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestRun_PrintsCobraError(t *testing.T) {
	streams, _, stdout, stderr := iostreams.Test()

	code := run([]string{"--bogus"}, streams)

	require.Equal(t, 1, code)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "unknown flag: --bogus")
}
