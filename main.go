package main

import (
	"fmt"
	"os"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/heaths/gh-users/internal/cmd"
)

func main() {
	os.Exit(run(os.Args[1:], iostreams.System()))
}

func run(args []string, streams *iostreams.IOStreams) int {
	root := cmd.NewWithIO(streams)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(streams.ErrOut, err)
		return 1
	}

	return 0
}
