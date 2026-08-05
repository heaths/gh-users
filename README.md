# gh users

[GitHub CLI] extension to list all users for a repository, or only those matching a specified substring.

```bash
# Show all members' status:
gh users

# Show status for any member matching the given substring:
gh users heath
```

## Query users

Generate a compact GraphQL query for users matching one or more partial logins:

```bash
scripts/query-users.sh heath octo
```

The script searches public GitHub users for each partial login, deduplicates the
matching logins, and writes a query using a reusable `User` fragment.

## Install

Make sure you have version 2.0 or [newer] of the [GitHub CLI] installed.

```bash
gh extension install heaths/gh-users
```

### Upgrade

The `gh extension list` command shows if updates are available for extensions. To upgrade, you can use the `gh extension upgrade` command:

```bash
gh extension upgrade heaths/gh-users

# Or upgrade all extensions:
gh extension upgrade --all
```

## License

Licensed under the [MIT](LICENSE.txt) license.

[GitHub CLI]: https://github.com/cli/cli
[newer]: https://github.com/cli/cli/releases/latest
