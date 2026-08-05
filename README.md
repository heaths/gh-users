# gh users

GitHub CLI extension to list all users for a repository, or only those
matching one or more specified substrings.

## Install

Make sure you have version 2.0 or [newer] of the [GitHub CLI] installed.

```bash
gh extension install heaths/gh-users
```

## Usage

```bash
# Show all members' status:
gh users

# Show status for any member matching the given substring:
gh users heath

# Show status for any member matching any of the given substrings:
gh users heath octo
```

Use `-R` / `--repo` to target another repository in `[HOST/]OWNER/REPO` format.
When color is enabled, matching text in login and name fields is highlighted.

### JSON and jq output

Use `--json` to select fields from `login,name,email,status` and `--jq` to
filter JSON output.

```bash
gh users --json login,name,email,status
gh users --json login,status --jq '.[] | select(.status != null) | .login'
```

### Template

Use `--template` to format output with a Go template. For the built-in
formatting and template helpers available through GitHub CLI, run:

```bash
gh help formatting
```

This command also documents the built-in jq and template formatting helpers.

### Upgrade

```bash
gh extension upgrade heaths/gh-users

# Or upgrade all extensions:
gh extension upgrade --all
```

## License

Licensed under the [MIT](LICENSE.txt) license.

[GitHub CLI]: https://github.com/cli/cli
[newer]: https://github.com/cli/cli/releases/latest
