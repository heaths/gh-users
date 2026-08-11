package main

// cspell:ignore mcat mheath mocto

import (
	"io"
	"os"
	"testing"

	"github.com/cli/go-gh/v2/pkg/term"
	ghclient "github.com/heaths/gh-users/internal/github"
	appterm "github.com/heaths/gh-users/internal/terminal"
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
	terminal, stdout, stderr := testTerminal(t, false)

	code := run([]string{"--bogus"}, terminal)

	require.Equal(t, 1, code)
	require.Empty(t, fileString(t, stdout))
	require.Contains(t, fileString(t, stderr), "unknown flag: --bogus")
}

func TestNewRootCmd_UsesExtensionCommandName(t *testing.T) {
	terminal, _, _ := testTerminal(t, false)
	cmd := newRootCmd(terminal, &rootOptions{})

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
	terminal, stdout, _ := testTerminal(t, false)

	mock := &mockService{
		response: &ghclient.QueryEnvelope{
			Data: ghclient.QueryResponse{
				Repository: map[string]ghclient.UserConnection{
					"alias_0": {
						Nodes: []ghclient.User{
							{
								Login:  "heaths",
								Name:   "Heath Stewart",
								Email:  "heath@example.com",
								Status: "available",
							},
						},
					},
				},
			},
		},
	}

	err := runUsers(&rootOptions{
		term:   terminal,
		client: mock,
		repo:   "heaths/gh-users",
	}, []string{"heath"})
	require.NoError(t, err)
	require.Equal(t, "heaths\tHeath Stewart\theath@example.com\tavailable\n", fileString(t, stdout))
	require.Equal(t, "heaths", mock.owner)
	require.Equal(t, "gh-users", mock.repo)
	require.Equal(t, []string{"heath"}, mock.partials)
}

func TestRunUsers_PrintsJSONOutput(t *testing.T) {
	terminal, stdout, _ := testTerminal(t, false)
	mock := &mockService{
		response: &ghclient.QueryEnvelope{
			Data: ghclient.QueryResponse{
				Repository: map[string]ghclient.UserConnection{
					"alias_0": {
						Nodes: []ghclient.User{
							{
								Login:  "heaths",
								Name:   "Heath Stewart",
								Email:  "heath@example.com",
								Status: "available",
							},
						},
					},
				},
			},
		},
	}

	err := runUsers(&rootOptions{
		term:       terminal,
		client:     mock,
		repo:       "heaths/gh-users",
		jsonFields: "login,status",
	}, []string{"heath"})
	require.NoError(t, err)
	require.Equal(t, "[{\"login\":\"heaths\",\"status\":\"available\"}]\n", fileString(t, stdout))
}

func TestRunUsers_FiltersJSONOutputWithJQ(t *testing.T) {
	terminal, stdout, _ := testTerminal(t, false)
	mock := &mockService{
		response: &ghclient.QueryEnvelope{
			Data: ghclient.QueryResponse{
				Repository: map[string]ghclient.UserConnection{
					"alias_0": {
						Nodes: []ghclient.User{
							{Login: "heaths", Name: "Heath Stewart"},
						},
					},
				},
			},
		},
	}

	err := runUsers(&rootOptions{
		term:         terminal,
		client:       mock,
		repo:         "heaths/gh-users",
		jqExpression: ".[].login",
	}, []string{"heath"})
	require.NoError(t, err)
	require.Equal(t, "heaths\n", fileString(t, stdout))
}

func TestRunUsers_FormatsJSONOutputWithTemplate(t *testing.T) {
	terminal, stdout, _ := testTerminal(t, false)
	mock := &mockService{
		response: &ghclient.QueryEnvelope{
			Data: ghclient.QueryResponse{
				Repository: map[string]ghclient.UserConnection{
					"alias_0": {
						Nodes: []ghclient.User{
							{Login: "heaths", Email: "heath@example.com"},
						},
					},
				},
			},
		},
	}

	err := runUsers(&rootOptions{
		term:   terminal,
		client: mock,
		repo:   "heaths/gh-users",
		tmpl:   "{{range .}}{{printf \"%s\\t%s\\n\" .login .email}}{{end}}",
	}, []string{"heath"})
	require.NoError(t, err)
	require.Equal(t, "heaths\theath@example.com\n", fileString(t, stdout))
}

func TestRunUsers_HighlightsPatternsInDefaultOutput(t *testing.T) {
	terminal, stdout, _ := testTerminal(t, true)
	mock := &mockService{
		response: &ghclient.QueryEnvelope{
			Data: ghclient.QueryResponse{
				Repository: map[string]ghclient.UserConnection{
					"alias_0": {
						Nodes: []ghclient.User{
							{Login: "octocat", Name: "The Octo Cat"},
						},
					},
				},
			},
		},
	}

	err := runUsers(&rootOptions{
		term:   terminal,
		client: mock,
		repo:   "heaths/gh-users",
	}, []string{"OCTO"})
	require.NoError(t, err)
	output := fileString(t, stdout)
	require.Contains(t, output, appterm.HighlightLogin(terminal, "octocat", "OCTO"))
	require.Contains(t, output, appterm.Highlight(terminal, "The Octo Cat", "OCTO"))
}

func TestRunUsers_PrintsEmptyOutputWhenNoMatches(t *testing.T) {
	terminal, stdout, _ := testTerminal(t, false)
	mock := &mockService{
		response: &ghclient.QueryEnvelope{
			Data: ghclient.QueryResponse{
				Repository: map[string]ghclient.UserConnection{},
			},
		},
	}

	err := runUsers(&rootOptions{
		term:   terminal,
		client: mock,
		repo:   "heaths/gh-users",
	}, nil)
	require.NoError(t, err)
	require.Empty(t, fileString(t, stdout))
}

func testTerminal(t *testing.T, tty bool) (term.Term, *os.File, *os.File) {
	t.Helper()

	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	require.NoError(t, err)
	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	require.NoError(t, err)

	originalStdout, originalStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdout, stderr
	t.Cleanup(func() {
		os.Stdout, os.Stderr = originalStdout, originalStderr
		if err := stdout.Close(); err != nil {
			t.Errorf("closing stdout temp file: %v", err)
		}
		if err := stderr.Close(); err != nil {
			t.Errorf("closing stderr temp file: %v", err)
		}
	})

	t.Setenv("GH_FORCE_TTY", "")
	t.Setenv("CLICOLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
	t.Setenv("NO_COLOR", "1")
	if tty {
		t.Setenv("GH_FORCE_TTY", "80")
		t.Setenv("NO_COLOR", "")
	}

	return term.FromEnv(), stdout, stderr
}

func fileString(t *testing.T, file *os.File) string {
	t.Helper()
	_, err := file.Seek(0, io.SeekStart)
	require.NoError(t, err)
	data, err := io.ReadAll(file)
	require.NoError(t, err)
	return string(data)
}
