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

### JSON, jq, and template output

Use `--json` to select fields from `login,name,email,status`, `--jq` to filter JSON output, and `--template` to format it with a Go template.

```bash
gh users --json login,name,email,status
gh users --json login,status --jq '.[] | select(.status != null) | .login'
gh users --template '{{range .}}{{printf "%s\t%s\n" .login .email}}{{end}}'
```

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
