package colors

// cspell:ignore aths mcat mheaths mocto moctocat

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/stretchr/testify/require"
)

func TestHighlight(t *testing.T) {
	tests := map[string]struct {
		colorEnabled bool
		input        string
		patterns     []string
		want         string
	}{
		"disabled": {
			input:    "Octocat",
			patterns: []string{"octo"},
			want:     "Octocat",
		},
		"no patterns": {
			colorEnabled: true,
			input:        "Octocat",
			want:         "Octocat",
		},
		"empty patterns": {
			colorEnabled: true,
			input:        "Octocat",
			patterns:     []string{"", ""},
			want:         "Octocat",
		},
		"16 color highlight": {
			colorEnabled: true,
			input:        "Octocat OCTO",
			patterns:     []string{"octo"},
			want:         "\x1b[30;43mOcto\x1b[39;49mcat \x1b[30;43mOCTO\x1b[39;49m",
		},
		"overlapping patterns": {
			colorEnabled: true,
			input:        "heaths",
			patterns:     []string{"heat", "aths"},
			want:         "\x1b[30;43mheaths\x1b[39;49m",
		},
		"overlapping occurrences": {
			colorEnabled: true,
			input:        "banana",
			patterns:     []string{"ana"},
			want:         "b\x1b[30;43manana\x1b[39;49m",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			terminal := testTerm(t, tt.colorEnabled)
			require.Equal(t, tt.want, Highlight(terminal, tt.input, tt.patterns...))
		})
	}
}

func TestHighlightLogin(t *testing.T) {
	t.Run("applies green plain color when there are no matches", func(t *testing.T) {
		terminal := testTerm(t, true)
		require.Equal(t, "\x1b[32moctocat\x1b[0m", HighlightLogin(terminal, "octocat"))
	})

	t.Run("uses the shared dark-on-yellow highlight spans", func(t *testing.T) {
		terminal := testTerm(t, true)
		want := "\x1b[30;43mocto\x1b[39;49m\x1b[32mcat\x1b[0m"
		require.Equal(t, want, HighlightLogin(terminal, "octocat", "octo"))
	})

	t.Run("does not style text when color is disabled", func(t *testing.T) {
		terminal := testTerm(t, false)
		require.Equal(t, "octocat", HighlightLogin(terminal, "octocat", "octo"))
	})
}

func TestFieldColors(t *testing.T) {
	defaultTheme := termTheme
	t.Cleanup(func() {
		termTheme = defaultTheme
	})

	t.Run("uses bright black on light themes", func(t *testing.T) {
		terminal := testTerm(t, true)
		termTheme = func(term.Term) string { return lightTheme }
		require.Equal(t, "\x1b[0;90memail\x1b[0m", Muted(terminal)("email"))
		require.Equal(t, "\x1b[0;33mstatus\x1b[0m", Yellow(terminal)("status"))
	})

	t.Run("uses dim white on dark themes", func(t *testing.T) {
		terminal := testTerm(t, true)
		termTheme = func(term.Term) string { return darkTheme }
		require.Equal(t, "\x1b[0;2;37memail\x1b[0m", Muted(terminal)("email"))
	})

	t.Run("does not style muted fields when there is no theme", func(t *testing.T) {
		terminal := testTerm(t, true)
		termTheme = func(term.Term) string { return "none" }
		require.Equal(t, "email", Muted(terminal)("email"))
		require.Equal(t, "\x1b[0;33mstatus\x1b[0m", Yellow(terminal)("status"))
	})

	t.Run("does not style fields when color is disabled", func(t *testing.T) {
		terminal := testTerm(t, false)
		termTheme = func(term.Term) string { return lightTheme }
		require.Equal(t, "email", Muted(terminal)("email"))
		require.Equal(t, "status", Yellow(terminal)("status"))
	})
}

func testTerm(t *testing.T, colorEnabled bool) term.Term {
	t.Helper()
	t.Setenv("GH_FORCE_TTY", "")
	t.Setenv("CLICOLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "")
	t.Setenv("COLORTERM", "")
	if colorEnabled {
		t.Setenv("CLICOLOR_FORCE", "1")
	} else {
		t.Setenv("NO_COLOR", "1")
	}
	return term.FromEnv()
}
