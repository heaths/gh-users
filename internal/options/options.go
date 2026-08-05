package options

import (
	"fmt"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/repository"
)

func ResolveRepository(repoFlag string) (repository.Repository, error) {
	if repoFlag != "" {
		repo, err := repository.Parse(repoFlag)
		if err != nil {
			return nil, fmt.Errorf("invalid repository: %w", err)
		}

		return repo, nil
	}

	repo, err := gh.CurrentRepository()
	if err != nil {
		return nil, fmt.Errorf("could not determine repository; pass --repo OWNER/REPO: %w", err)
	}

	return repo, nil
}
