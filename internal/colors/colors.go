package colors

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/cli/cli/v2/pkg/iostreams"
)

const (
	greenForeground16 = "\x1b[32m"
	// Use 30;103 for a bright-yellow background instead of the current dim yellow.
	blackOnYellow16 = "\x1b[30;43m"

	resetDefault = "\x1b[39;49m"
	resetAll     = "\x1b[0m"
)

type match struct {
	start int
	end   int
}

type themeColors struct {
	namePlain     string
	highlightSpan string
}

type fieldStyle struct {
	plainPrefix     string
	plainSuffix     string
	highlightPrefix string
	highlightSuffix string
}

// Highlight applies theme-aware highlighting for name-like fields.
func Highlight(colorScheme *iostreams.ColorScheme, input string, patterns ...string) string {
	colors := colorsFor(colorScheme)
	return highlight(fieldStyle{
		plainPrefix:     colors.namePlain,
		plainSuffix:     "",
		highlightPrefix: colors.highlightSpan,
		highlightSuffix: resetDefault,
	}, colorScheme, input, patterns...)
}

// HighlightLogin applies theme-aware highlighting for login fields.
func HighlightLogin(colorScheme *iostreams.ColorScheme, input string, patterns ...string) string {
	colors := colorsFor(colorScheme)
	return highlight(fieldStyle{
		plainPrefix:     loginPlain(colorScheme),
		plainSuffix:     resetAll,
		highlightPrefix: colors.highlightSpan,
		highlightSuffix: resetDefault,
	}, colorScheme, input, patterns...)
}

func highlight(style fieldStyle, colorScheme *iostreams.ColorScheme, input string, patterns ...string) string {
	if input == "" {
		return input
	}
	if colorScheme == nil || !colorScheme.Enabled {
		return input
	}

	matches := findMatches(input, patterns)
	if len(matches) == 0 {
		return apply(style.plainPrefix, style.plainSuffix, input)
	}

	var output strings.Builder
	output.Grow(len(input) + len(matches)*(len(style.highlightPrefix)+len(style.highlightSuffix)))
	offset := 0
	for _, current := range matches {
		output.WriteString(apply(style.plainPrefix, style.plainSuffix, input[offset:current.start]))
		output.WriteString(style.highlightPrefix)
		output.WriteString(input[current.start:current.end])
		output.WriteString(style.highlightSuffix)
		offset = current.end
	}
	output.WriteString(apply(style.plainPrefix, style.plainSuffix, input[offset:]))

	return output.String()
}

func colorsFor(colorScheme *iostreams.ColorScheme) themeColors {
	return themeColors{
		namePlain:     namePlain(colorScheme),
		highlightSpan: blackOnYellow16,
	}
}

func namePlain(_ *iostreams.ColorScheme) string {
	return ""
}

func loginPlain(colorScheme *iostreams.ColorScheme) string {
	return greenForeground16
}

func findMatches(input string, patterns []string) []match {
	lowerInput := strings.ToLower(input)
	var matches []match
	for _, pattern := range patterns {
		lowerPattern := strings.ToLower(pattern)
		if lowerPattern == "" {
			continue
		}

		for offset := 0; offset < len(lowerInput); {
			index := strings.Index(lowerInput[offset:], lowerPattern)
			if index < 0 {
				break
			}

			start := offset + index
			end := start + len(lowerPattern)
			matches = append(matches, match{start: start, end: end})
			_, size := utf8.DecodeRuneInString(input[start:])
			offset = start + size
		}
	}

	if len(matches) == 0 {
		return nil
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].start == matches[j].start {
			return matches[i].end < matches[j].end
		}
		return matches[i].start < matches[j].start
	})

	merged := matches[:1]
	for _, current := range matches[1:] {
		last := &merged[len(merged)-1]
		if current.start <= last.end {
			if current.end > last.end {
				last.end = current.end
			}
			continue
		}
		merged = append(merged, current)
	}

	return merged
}

func apply(prefix, suffix, input string) string {
	if input == "" || prefix == "" {
		return input
	}
	return prefix + input + suffix
}
