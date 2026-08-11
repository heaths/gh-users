package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	ghjq "github.com/cli/go-gh/v2/pkg/jq"
	"github.com/cli/go-gh/v2/pkg/jsonpretty"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	ghtemplate "github.com/cli/go-gh/v2/pkg/template"
	"github.com/cli/go-gh/v2/pkg/term"
	ghclient "github.com/heaths/gh-users/internal/github"
	"github.com/heaths/gh-users/internal/options"
	appterm "github.com/heaths/gh-users/internal/terminal"
	"github.com/spf13/cobra"
)

func main() {
	os.Exit(run(os.Args[1:], term.FromEnv()))
}

type userService interface {
	QueryUsers(owner, repo string, partials []string) (*ghclient.QueryEnvelope, error)
}

type rootOptions struct {
	term   term.Term
	client userService
	repo   string

	jsonFields   string
	jqExpression string
	tmpl         string
}

func run(args []string, terminal term.Term) int {
	root := newRootCmd(terminal, &rootOptions{term: terminal})
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(terminal.ErrOut(), err)
		return 1
	}

	return 0
}

func newRootCmd(terminal term.Term, opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gh users [partial-username...]",
		Short: "List repository users, optionally filtered by partial username",
		Long:  "List assignable repository users, optionally filtered by one or more partial usernames.",
		Example: "  gh users\n" +
			"  gh users heath octo\n" +
			"  gh users --json login,name,email,status\n" +
			"  gh users --jq '.[].login'\n" +
			"  gh users --template '{{range .}}{{printf \"%s\\t%s\\n\" .login .email}}{{end}}'",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUsers(opts, args)
		},
	}
	cmd.SetOut(terminal.Out())
	cmd.SetErr(terminal.ErrOut())
	cmd.PersistentFlags().StringVarP(&opts.repo, "repo", "R", "", "Select another repository using the [HOST/]OWNER/REPO format")
	cmd.Flags().StringVar(&opts.jsonFields, "json", "", fmt.Sprintf("Output JSON with the specified fields (%s)", strings.Join(ghclient.UserFields(), ",")))
	cmd.Flags().StringVar(&opts.jqExpression, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().StringVar(&opts.tmpl, "template", "", "Format JSON output using a Go template")
	cmd.MarkFlagsMutuallyExclusive("jq", "template")

	return cmd
}

func runUsers(opts *rootOptions, partials []string) error {
	repo, err := options.ResolveRepository(opts.repo)
	if err != nil {
		return err
	}

	client, err := ensureClient(opts)
	if err != nil {
		return err
	}
	opts.client = client

	response, err := client.QueryUsers(repo.Owner, repo.Name, partials)
	if err != nil {
		return err
	}

	users, err := processUsers(response)
	if err != nil {
		return err
	}

	if opts.jsonFields != "" || opts.jqExpression != "" || opts.tmpl != "" {
		return writeUserOutput(opts, users)
	}

	return printUsers(opts.term, users, partials)
}

func ensureClient(opts *rootOptions) (userService, error) {
	if opts.client != nil {
		return opts.client, nil
	}

	return ghclient.New(nil, opts.term.ErrOut())
}

func processUsers(response *ghclient.QueryEnvelope) ([]ghclient.User, error) {
	data, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	var filtered bytes.Buffer
	if err := ghjq.Evaluate(bytes.NewReader(data), &filtered, ghclient.UsersJQExpression()); err != nil {
		return nil, err
	}

	var users []ghclient.User
	if err := json.Unmarshal(filtered.Bytes(), &users); err != nil {
		return nil, err
	}

	return users, nil
}

func printUsers(terminal term.Term, users []ghclient.User, patterns []string) error {
	table := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), terminalWidth(terminal))

	for _, user := range users {
		table.AddField(appterm.HighlightLogin(terminal, user.Login, patterns...), tableprinter.WithTruncate(nil))
		table.AddField(appterm.Highlight(terminal, user.Name, patterns...))
		table.AddField(user.Email, tableprinter.WithColor(appterm.Muted(terminal)))
		table.AddField(user.Status, tableprinter.WithColor(appterm.Yellow(terminal)))
		table.EndRow()
	}

	return table.Render()
}

func writeUserOutput(opts *rootOptions, users []ghclient.User) error {
	data, err := marshalUserOutput(opts, users)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(data)
	switch {
	case opts.tmpl != "":
		tmpl := ghtemplate.New(opts.term.Out(), terminalWidth(opts.term), opts.term.IsColorEnabled())
		if err := tmpl.Parse(opts.tmpl); err != nil {
			return err
		}
		if err := tmpl.Execute(reader); err != nil {
			return err
		}
		return tmpl.Flush()
	case opts.jqExpression != "":
		return ghjq.Evaluate(reader, opts.term.Out(), opts.jqExpression)
	default:
		return writeJSONOutput(opts.term, reader)
	}
}

func marshalUserOutput(opts *rootOptions, users []ghclient.User) ([]byte, error) {
	if opts.jsonFields == "" {
		return marshalJSON(users)
	}

	fields, err := ghclient.ParseUserFields(opts.jsonFields)
	if err != nil {
		return nil, err
	}

	data, err := ghclient.ExportUsers(users, fields)
	if err != nil {
		return nil, err
	}

	return marshalJSON(data)
}

func marshalJSON(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeJSONOutput(terminal term.Term, input io.Reader) error {
	if err := prettyPrintJSONOutput(terminal, input); err == nil {
		return nil
	}

	_, err := io.Copy(terminal.Out(), input)
	return err
}

func prettyPrintJSONOutput(terminal term.Term, input io.Reader) error {
	if !terminal.IsTerminalOutput() {
		return ioCopyUnsupported{}
	}

	return jsonpretty.Format(terminal.Out(), input, "  ", terminal.IsColorEnabled())
}

type ioCopyUnsupported struct{}

func (ioCopyUnsupported) Error() string { return "pretty JSON output is not supported" }

func terminalWidth(terminal term.Term) int {
	width, _, err := terminal.Size()
	if err == nil && width > 0 {
		return width
	}
	return 80
}
