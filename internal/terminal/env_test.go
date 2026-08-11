package terminal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsSet(t *testing.T) {
	tests := map[string]struct {
		value string
		set   bool
		want  bool
	}{
		"unset":           {want: false},
		"one":             {value: "1", set: true, want: true},
		"true":            {value: "true", set: true, want: true},
		"mixed case true": {value: "TrUe", set: true, want: true},
		"zero":            {value: "0", set: true, want: false},
		"false":           {value: "false", set: true, want: false},
		"empty":           {value: "", set: true, want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			const env = "GH_USERS_TEST_ENV"
			original, ok := os.LookupEnv(env)
			t.Cleanup(func() {
				if ok {
					_ = os.Setenv(env, original)
					return
				}
				_ = os.Unsetenv(env)
			})

			if tt.set {
				require.NoError(t, os.Setenv(env, tt.value))
			} else {
				require.NoError(t, os.Unsetenv(env))
			}

			require.Equal(t, tt.want, IsSet(env))
		})
	}
}
