# gh users

[GitHub CLI] extension to list all users for a repository, or only those matching one or more specified substrings.

```bash
# Show all members' status:
gh users

# Show status for any member matching the given substring:
gh users heath

# Show status for any member matching any of the given substrings:
gh users heath octo
```

Each substring is queried using its own alias in a single GraphQL request. The aliases share a
`UserFields` fragment to minimize what is sent to the server, and the results from every alias are
deduplicated and sorted by login. This is equivalent to:

```bash
gh api graphql -F owner=":owner" -F repo=":repo" -f user_0="heath" -f query='...' \
  --jq '[.data.repository[].nodes[]] | unique_by(.login) | sort_by(.login) | .[].login'
```

The [GitHub CLI] does not support `--jq` and `--template` together, so `--jq` formats the output.

## Tests

Tests run on `windows-latest`, `macos-latest`, and `ubuntu-latest` for every pull request. They use
the `GITHUB_TOKEN` provided to the workflow and require no additional secrets. You can also run them
locally:

```bash
./tests/test.sh
```

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
