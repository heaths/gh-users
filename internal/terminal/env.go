package terminal

import (
	"os"
	"strings"
)

// IsSet reports whether the named environment variable is set to 1 or true.
func IsSet(name string) bool {
	value, ok := os.LookupEnv(name)
	if !ok {
		return false
	}

	return value == "1" || strings.EqualFold(value, "true")
}
