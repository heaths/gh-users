package options

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveRepository_ParsesExplicitRepo(t *testing.T) {
	repo, err := ResolveRepository("heaths/gh-users")
	require.NoError(t, err)
	require.Equal(t, "heaths", repo.Owner())
	require.Equal(t, "gh-users", repo.Name())
}

func TestResolveRepository_RejectsInvalidRepo(t *testing.T) {
	_, err := ResolveRepository("heaths")
	require.ErrorContains(t, err, `expected the "[HOST/]OWNER/REPO" format`)
}
