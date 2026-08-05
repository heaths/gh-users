# AGENTS

## Project

- `gh users` is a GitHub CLI extension that lists assignable users for a
  repo, optionally filtered by partial usernames.
- `main.go` owns CLI wiring and output.
- `internal/github` owns GraphQL query construction, paging, and response types.
- `internal/options` resolves the target repository.

## Modify

- Keep behavior CLI-first: clear stdout output, real errors on stderr, no
  hidden fallback behavior.
- Prefer small changes in existing files before adding new packages or layers.
- Reuse stdlib and existing dependencies first.
- For GitHub CLI integration, use as much of `github.com/cli/go-gh/v2` as
  available before adding custom code or other dependencies.
- Avoid new third-party modules unless they are explicitly requested or
  clearly necessary for the change.
- Prefer Go APIs over shelling out.
- Keep user-visible command behavior documented in `README.md`.
- Add or update focused Go tests near the changed code.

## Validate

- `go test ./...`
- `tests/test.sh`
