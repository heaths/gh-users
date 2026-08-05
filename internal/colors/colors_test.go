package colors

// cspell:ignore aths mcat mheaths mocto moctocat

import (
	"testing"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestHighlight(t *testing.T) {
	tests := map[string]struct {
		colorScheme *iostreams.ColorScheme
		input       string
		patterns    []string
		want        string
	}{
		"disabled": {
			input:    "Octocat",
			patterns: []string{"octo"},
			want:     "Octocat",
		},
		"no patterns": {
			colorScheme: enabledScheme(iostreams.DarkTheme),
			input:       "Octocat",
			want:        "Octocat",
		},
		"empty patterns": {
			colorScheme: enabledScheme(iostreams.DarkTheme),
			input:       "Octocat",
			patterns:    []string{"", ""},
			want:        "Octocat",
		},
		"dark theme 16 color highlight": {
			colorScheme: enabledScheme(iostreams.DarkTheme),
			input:       "Octocat OCTO",
			patterns:    []string{"octo"},
			want:        "\x1b[30;43mOcto\x1b[39;49mcat \x1b[30;43mOCTO\x1b[39;49m",
		},
		"light theme 16 color plain and highlight": {
			colorScheme: enabledScheme(iostreams.LightTheme),
			input:       "Octocat",
			patterns:    []string{"octo"},
			want:        "\x1b[30;43mOcto\x1b[39;49mcat",
		},
		"unknown theme uses dark name plain": {
			colorScheme: enabledScheme(iostreams.NoTheme),
			input:       "Octocat",
			patterns:    []string{"octo"},
			want:        "\x1b[30;43mOcto\x1b[39;49mcat",
		},
		"overlapping patterns": {
			colorScheme: enabledScheme(iostreams.DarkTheme),
			input:       "heaths",
			patterns:    []string{"heat", "aths"},
			want:        "\x1b[30;43mheaths\x1b[39;49m",
		},
		"overlapping occurrences": {
			colorScheme: enabledScheme(iostreams.DarkTheme),
			input:       "banana",
			patterns:    []string{"ana"},
			want:        "b\x1b[30;43manana\x1b[39;49m",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.want, Highlight(tt.colorScheme, tt.input, tt.patterns...))
		})
	}
}

func TestHighlightLogin(t *testing.T) {
	t.Run("applies green plain color when there are no matches", func(t *testing.T) {
		colorScheme := enabledScheme(iostreams.DarkTheme)
		require.Equal(t, "\x1b[32moctocat\x1b[0m", HighlightLogin(colorScheme, "octocat"))
	})

	t.Run("uses the shared dark-on-yellow highlight spans", func(t *testing.T) {
		colorScheme := enabledScheme(iostreams.DarkTheme)
		want := "\x1b[30;43mocto\x1b[39;49m\x1b[32mcat\x1b[0m"
		require.Equal(t, want, HighlightLogin(colorScheme, "octocat", "octo"))
	})

	t.Run("does not style text when color is disabled", func(t *testing.T) {
		colorScheme := &iostreams.ColorScheme{}
		require.Equal(t, "octocat", HighlightLogin(colorScheme, "octocat", "octo"))
	})
}

func enabledScheme(theme string) *iostreams.ColorScheme {
	return &iostreams.ColorScheme{
		Enabled: true,
		Theme:   theme,
	}
}
