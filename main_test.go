package main

import (
	"testing"

	"github.com/cli/cli/v2/pkg/iostreams"
	ghclient "github.com/heaths/gh-users/internal/github"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	response *ghclient.QueryEnvelope
	err      error
	owner    string
	repo     string
	partials []string
}

func (m *mockService) QueryUsers(owner, repo string, partials []string) (*ghclient.QueryEnvelope, error) {
	m.owner = owner
	m.repo = repo
	m.partials = append([]string(nil), partials...)
	return m.response, m.err
}

func TestRun_PrintsCobraError(t *testing.T) {
	streams, _, stdout, stderr := iostreams.Test()

	code := run([]string{"--bogus"}, streams)

	require.Equal(t, 1, code)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "unknown flag: --bogus")
}

func TestNewRootCmd_UsesExtensionCommandName(t *testing.T) {
	cmd := newRootCmd(iostreams.System(), &rootOptions{})

	require.Equal(t, "gh users [partial-username...]", cmd.Use)
}

func TestProcessUsers_DedupesAndSortsViaJQ(t *testing.T) {
	users, err := processUsers(&ghclient.QueryEnvelope{
		Data: ghclient.QueryResponse{
			Repository: map[string]ghclient.UserConnection{
				"alias_0": {
					Nodes: []ghclient.User{
						{Login: "heaths", Name: "Heath"},
						{Login: "octocat", Name: "Octo"},
					},
				},
				"alias_1": {
					Nodes: []ghclient.User{
						{Login: "heaths", Name: "Heath"},
						{Login: "alpha", Name: "Alpha"},
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []ghclient.User{
		{Login: "alpha", Name: "Alpha"},
		{Login: "heaths", Name: "Heath"},
		{Login: "octocat", Name: "Octo"},
	}, users)
}

func TestRunUsers_PrintsTSVOutput(t *testing.T) {
	streams, _, stdout, _ := iostreams.Test()
	streams.SetStdoutTTY(false)

	mock := &mockService{
		response: &ghclient.QueryEnvelope{
			Data: ghclient.QueryResponse{
				Repository: map[string]ghclient.UserConnection{
					"alias_0": {
						Nodes: []ghclient.User{
							{
								Login: "heaths",
								Name:  "Heath Stewart",
								Email: "heath@example.com",
								Status: &ghclient.UserStatus{
									Message: "available",
								},
							},
						},
					},
				},
			},
		},
	}

	err := runUsers(&rootOptions{
		io:     streams,
		client: mock,
		repo:   "heaths/gh-users",
	}, []string{"heath"})
	require.NoError(t, err)
	require.Equal(t, "heaths\tHeath Stewart\theath@example.com\tavailable\n", stdout.String())
	require.Equal(t, "heaths", mock.owner)
	require.Equal(t, "gh-users", mock.repo)
	require.Equal(t, []string{"heath"}, mock.partials)
}

func TestRunUsers_PrintsEmptyOutputWhenNoMatches(t *testing.T) {
	streams, _, stdout, _ := iostreams.Test()
	mock := &mockService{
		response: &ghclient.QueryEnvelope{
			Data: ghclient.QueryResponse{
				Repository: map[string]ghclient.UserConnection{},
			},
		},
	}

	err := runUsers(&rootOptions{
		io:     streams,
		client: mock,
		repo:   "heaths/gh-users",
	}, nil)
	require.NoError(t, err)
	require.Empty(t, stdout.String())
}
