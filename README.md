# gh users

GitHub CLI extension to list all users for a repository, or only those
matching one or more specified substrings.

## Install

Make sure you have version 2.0 or [newer] of the [GitHub CLI] installed.

```sh
gh extension install heaths/gh-users
```

## Usage

```sh
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

```sh
gh users --json login,name,email,status
gh users --json login,status --jq '.[] | select(.status != null) | .login'
```

### Template

Use `--template` to format output with a Go template e.g., to display the
standard table of any login or name matching "wil" without highlighting:

```sh
gh users wil --template '{{range .}}{{tablerow (autocolor "green" .login) .name (autocolor "white+d" .email) (autocolor "yellow" .status)}}{{end}}'
```

For the built-in formatting and template helpers available through GitHub CLI, run:

```sh
gh help formatting
```

This command also documents the built-in jq and template formatting helpers.

### Upgrade

```sh
gh extension upgrade heaths/gh-users

# Or upgrade all extensions:
gh extension upgrade --all
```

### Environment variables

When fetching many records, `gh users` displays a spinner. Set
`GH_SPINNER_DISABLED=1` or `GH_SPINNER_DISABLED=true` to disable it.

Run `gh help environment` for more details about this and other environment
variables such as `GH_REPO` and `NO_COLOR`.

## License

Licensed under the [MIT](LICENSE.txt) license.

[GitHub CLI]: https://github.com/cli/cli
[newer]: https://github.com/cli/cli/releases/latest
