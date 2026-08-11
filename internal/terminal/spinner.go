package terminal

import (
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cli/go-gh/v2/pkg/term"
)

const spinnerDisabledEnv = "GH_SPINNER_DISABLED"

var (
	spinnerDefaultOutput = func() *os.File { return os.Stderr }
	spinnerIsTerminal    = term.IsTerminal
	spinnerColorEnabled  = func() bool { return term.IsColorForced() || !term.IsColorDisabled() }
)

type Spinner struct {
	spinner *spinner.Spinner
	color   string
	output  *os.File
}

func NewSpinner(output io.Writer) *Spinner {
	file := spinnerDefaultOutput()
	if output != nil {
		var ok bool
		file, ok = output.(*os.File)
		if !ok {
			return nil
		}
	}

	if IsSet(spinnerDisabledEnv) || !spinnerIsTerminal(file) {
		return nil
	}

	options := []spinner.Option{spinner.WithWriterFile(file)}
	color := ""
	if spinnerColorEnabled() {
		color = "fgCyan"
		options = append(options, spinner.WithColor(color))
	}

	return &Spinner{
		spinner: spinner.New(spinner.CharSets[11], 120*time.Millisecond, options...),
		color:   color,
		output:  file,
	}
}

func (s *Spinner) Start() {
	if s == nil {
		return
	}
	s.spinner.Start()
}

func (s *Spinner) Stop() {
	if s == nil {
		return
	}
	s.spinner.Stop()
}
