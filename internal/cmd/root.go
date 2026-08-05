package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cli/cli/v2/pkg/iostreams"
	ghjq "github.com/cli/go-gh/pkg/jq"
	"github.com/cli/go-gh/pkg/tableprinter"
	ghclient "github.com/heaths/gh-users/internal/github"
	"github.com/heaths/gh-users/internal/options"
	"github.com/spf13/cobra"
)

type userService interface {
	QueryUsers(owner, repo string, partials []string) (*ghclient.QueryEnvelope, error)
}

type rootOptions struct {
	io     *iostreams.IOStreams
	client userService
	repo   string
}

var executableName = func() string {
	if len(os.Args) == 0 {
		return "gh-users"
	}

	return filepath.Base(os.Args[0])
}

func New() *cobra.Command {
	return NewWithIO(iostreams.System())
}

func NewWithIO(streams *iostreams.IOStreams) *cobra.Command {
	opts := &rootOptions{io: streams}
	displayName := commandDisplayName()

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [partial-username...]", displayName),
		Short: "List repository users, optionally filtered by partial username",
		Long:  "List assignable repository users, optionally filtered by one or more partial usernames.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(opts, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.ErrOut)
	cmd.PersistentFlags().StringVarP(&opts.repo, "repo", "R", "", "Select another repository using the [HOST/]OWNER/REPO format")

	return cmd
}

func commandDisplayName() string {
	if _, ok := os.LookupEnv("GH_EXTENSION"); ok {
		return "gh users"
	}

	return executableName()
}

func run(opts *rootOptions, partials []string) error {
	repo, err := options.ResolveRepository(opts.repo)
	if err != nil {
		return err
	}

	client, err := ensureClient(opts.client)
	if err != nil {
		return err
	}
	opts.client = client

	response, err := client.QueryUsers(repo.Owner(), repo.Name(), partials)
	if err != nil {
		return err
	}

	users, err := processUsers(response)
	if err != nil {
		return err
	}

	return printUsers(opts.io, users)
}

func ensureClient(client userService) (userService, error) {
	if client != nil {
		return client, nil
	}

	return ghclient.New(nil)
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

func printUsers(streams *iostreams.IOStreams, users []ghclient.User) error {
	colors := streams.ColorScheme()
	table := tableprinter.New(streams.Out, streams.IsStdoutTTY(), streams.TerminalWidth())

	for _, user := range users {
		status := ""
		if user.Status != nil {
			status = user.Status.Message
		}

		table.AddField(user.Login, tableprinter.WithTruncate(nil), tableprinter.WithColor(colors.Green))
		table.AddField(user.Name)
		table.AddField(user.Email, tableprinter.WithColor(colors.Muted))
		table.AddField(status, tableprinter.WithColor(colors.Yellow))
		table.EndRow()
	}

	return table.Render()
}
