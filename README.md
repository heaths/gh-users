# gh users

[![releases](https://img.shields.io/github/v/release/heaths/gh-users.svg?logo=github)](https://github.com/heaths/gh-users/releases/latest)
[![ci](https://github.com/heaths/gh-users/actions/workflows/ci.yml/badge.svg?event=push)](https://github.com/heaths/gh-users/actions/workflows/ci.yml)

GitHub CLI extension to list assignable users for a repository, optionally filtered by one or more partial usernames.

## Install

Make sure you have version 2.0 or [newer](https://github.com/cli/cli/releases/latest) of the GitHub CLI installed.

```bash
gh extension install heaths/gh-users
```

Release builds are published for macOS, Linux, and Windows.

## Usage

```bash
# Show all assignable users for the current repository:
gh users

# Show users matching a single partial username:
gh users heath

# Show users matching any of multiple partial usernames:
gh users heath octo
```

Use `-R` / `--repo` to target another repository in `[HOST/]OWNER/REPO` format.

The extension builds one GraphQL alias per partial username, reuses a shared `UserFragment`, then runs the GraphQL response through `cli/go-gh/pkg/jq` to deduplicate and sort matches before printing the final table with `cli/go-gh/pkg/tableprinter`.

## Development

```bash
go test ./...
./tests/test.sh
```

The CI workflow runs on `ubuntu-latest`, `macos-latest`, and `windows-latest`. The smoke test queries this repository for `heath` and asserts that the single returned login is `heaths`.

## Upgrade

```bash
gh extension upgrade heaths/gh-users

# Or upgrade all extensions:
gh extension upgrade --all
```

## License

Licensed under the [MIT](LICENSE.txt) license.
